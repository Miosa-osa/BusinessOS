//! Gateway Mode CLI Tests — Chicago TDD (RED → GREEN → REFACTOR)
//!
//! Tests cover the --gateway / BOS_GATEWAY_URL code path in:
//!   T1: `bos pm4py discover` via gateway → success
//!   T2: `bos pm4py conform`  via gateway → success
//!   T3: `bos pm4py analyze`  via gateway → success
//!   T4: gateway returns 500  → non-zero exit + stderr error
//!   T5: local mode (no gateway) → reads file, no network call
//!   T6: connection refused   → non-zero exit + stderr error
//!
//! Mock strategy: raw std::net::TcpListener on 127.0.0.1:0 (OS-assigned port).
//! serve_one() handles exactly one connection in a background thread, which
//! matches the single HTTP request each CLI command makes.

use assert_cmd::Command;
use std::io::{Read, Write};
use std::net::TcpListener;

// ─── Mock HTTP server helpers ────────────────────────────────────────────────

fn bind_random_port() -> (TcpListener, u16) {
    let listener = TcpListener::bind("127.0.0.1:0").expect("failed to bind");
    let port = listener.local_addr().unwrap().port();
    (listener, port)
}

/// Spawn a background thread that accepts exactly one connection, reads the
/// request, and writes the supplied HTTP response.
fn serve_one<F>(listener: TcpListener, respond: F)
where
    F: Fn(&str) -> String + Send + 'static,
{
    std::thread::spawn(move || {
        if let Ok((mut stream, _)) = listener.accept() {
            let mut buf = vec![0u8; 8192];
            let n = stream.read(&mut buf).unwrap_or(0);
            let request = String::from_utf8_lossy(&buf[..n]).to_string();
            let response = respond(&request);
            let _ = stream.write_all(response.as_bytes());
        }
    });
}

fn http_200_json(body: &str) -> String {
    format!(
        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
        body.len(),
        body
    )
}

fn http_500(message: &str) -> String {
    let body = format!("{{\"error\":\"{}\"}}", message);
    format!(
        "HTTP/1.1 500 Internal Server Error\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n{}",
        body.len(),
        body
    )
}

/// Path to the test XES file used across CLI tests.
fn simple_xes_path() -> &'static str {
    "/Users/sac/chatmangpt/test_simple.xes"
}

// ─── T1: discover via gateway succeeds ───────────────────────────────────────

#[test]
fn test_gateway_mode_discover_succeeds() {
    let (listener, port) = bind_random_port();
    let url = format!("http://127.0.0.1:{}", port);

    let body = r#"{"algorithm":"alpha","places":4,"transitions":3,"arcs":6}"#;
    let response = http_200_json(body);
    serve_one(listener, move |_req| response.clone());

    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "discover", "--source", simple_xes_path()])
        .env("BOS_GATEWAY_URL", &url)
        .output()
        .expect("failed to run bos");

    let stdout = String::from_utf8_lossy(&output.stdout);
    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        output.status.success(),
        "exit non-zero. stdout: {}\nstderr: {}",
        stdout,
        stderr
    );

    let result: serde_json::Value =
        serde_json::from_str(&stdout).expect("stdout must be valid JSON");
    assert!(
        result["places"].as_u64().unwrap_or(0) > 0,
        "places must be > 0, got: {}",
        result
    );
    assert_eq!(result["algorithm"].as_str().unwrap_or(""), "alpha");
}

// ─── T2: conform via gateway succeeds ────────────────────────────────────────

#[test]
fn test_gateway_mode_conform_succeeds() {
    let (listener, port) = bind_random_port();
    let url = format!("http://127.0.0.1:{}", port);

    let body = r#"{"fitness":0.96,"traces_checked":5,"fitting_traces":4}"#;
    let response = http_200_json(body);
    serve_one(listener, move |_req| response.clone());

    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "conform", "--log", simple_xes_path()])
        .env("BOS_GATEWAY_URL", &url)
        .output()
        .expect("failed to run bos");

    let stdout = String::from_utf8_lossy(&output.stdout);
    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        output.status.success(),
        "exit non-zero. stdout: {}\nstderr: {}",
        stdout,
        stderr
    );

    let result: serde_json::Value =
        serde_json::from_str(&stdout).expect("stdout must be valid JSON");
    let fitness = result["fitness"].as_f64().unwrap_or(-1.0);
    assert!(
        (0.0..=1.0).contains(&fitness),
        "fitness must be in [0,1], got {}",
        fitness
    );
}

// ─── T3: analyze via gateway succeeds ────────────────────────────────────────

#[test]
fn test_gateway_mode_analyze_succeeds() {
    let (listener, port) = bind_random_port();
    let url = format!("http://127.0.0.1:{}", port);

    let body = r#"{
        "log_name":"test_simple.xes","num_traces":5,"num_events":15,
        "num_unique_activities":3,"num_variants":2,
        "avg_trace_length":3.0,"min_trace_length":3,"max_trace_length":3,
        "activity_frequency":[],"case_duration":{"min_seconds":0,"max_seconds":0,"avg_seconds":0.0,"median_seconds":0.0}
    }"#;
    let response = http_200_json(body);
    serve_one(listener, move |_req| response.clone());

    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "analyze", "--source", simple_xes_path()])
        .env("BOS_GATEWAY_URL", &url)
        .output()
        .expect("failed to run bos");

    let stdout = String::from_utf8_lossy(&output.stdout);
    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        output.status.success(),
        "exit non-zero. stdout: {}\nstderr: {}",
        stdout,
        stderr
    );

    let result: serde_json::Value =
        serde_json::from_str(&stdout).expect("stdout must be valid JSON");
    assert!(
        result["num_traces"].as_u64().unwrap_or(0) > 0,
        "num_traces must be > 0, got: {}",
        result
    );
}

// ─── T4: gateway returns 500 → non-zero exit ──────────────────────────────────

#[test]
fn test_cloud_mode_gateway_error_returns_nonzero() {
    let (listener, port) = bind_random_port();
    let url = format!("http://127.0.0.1:{}", port);

    let response = http_500("simulated backend error");
    serve_one(listener, move |_req| response.clone());

    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "discover", "--source", simple_xes_path()])
        .env("BOS_GATEWAY_URL", &url)
        .output()
        .expect("failed to run bos");

    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        !output.status.success(),
        "exit 0 unexpected — gateway 500 must cause non-zero exit. stderr: {}",
        stderr
    );

    let has_error_hint = stderr.contains("500")
        || stderr.contains("error")
        || stderr.contains("Error")
        || stderr.contains("failed")
        || stderr.contains("Gateway");
    assert!(
        has_error_hint,
        "stderr must mention the error, got: {}",
        stderr
    );
}

// ─── T5: local mode (no gateway env var) → reads file, no network ─────────────

#[test]
fn test_local_mode_does_not_call_network() {
    // No BOS_GATEWAY_URL → falls through to local pm4py engine
    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "discover", "--source", simple_xes_path()])
        .env_remove("BOS_GATEWAY_URL")
        .output()
        .expect("failed to run bos");

    let stdout = String::from_utf8_lossy(&output.stdout);
    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        output.status.success(),
        "local mode must exit 0. stdout: {}\nstderr: {}",
        stdout,
        stderr
    );

    let result: serde_json::Value =
        serde_json::from_str(&stdout).expect("stdout must be valid JSON");
    assert!(
        result["transitions"].as_u64().unwrap_or(0) > 0,
        "local engine must return transitions > 0, got: {}",
        result
    );
}

// ─── T6: connection refused → non-zero exit ──────────────────────────────────

#[test]
fn test_gateway_connection_refused_returns_error() {
    let (listener, port) = bind_random_port();
    let url = format!("http://127.0.0.1:{}", port);
    drop(listener); // close listener so connections are refused

    let output = Command::cargo_bin("bos")
        .unwrap()
        .args(["pm4py", "discover", "--source", simple_xes_path()])
        .env("BOS_GATEWAY_URL", &url)
        .output()
        .expect("failed to run bos");

    let stderr = String::from_utf8_lossy(&output.stderr);

    assert!(
        !output.status.success(),
        "exit 0 unexpected — connection refused must cause non-zero exit. stderr: {}",
        stderr
    );

    // reqwest reports connection failure; bos wraps it as "Gateway request failed:"
    let has_error_hint = stderr.contains("refused")
        || stderr.contains("error")
        || stderr.contains("Error")
        || stderr.contains("failed")
        || stderr.contains("Gateway");
    assert!(
        has_error_hint,
        "stderr must mention the connection error, got: {}",
        stderr
    );
}
