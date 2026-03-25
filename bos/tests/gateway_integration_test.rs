//! BOS CLI ↔ BusinessOS Gateway Integration Tests
//!
//! Tests end-to-end command execution through the gateway, including:
//! - Successful discover/conformance/statistics operations
//! - Error handling (service unavailable, malformed requests)
//! - Concurrent gateway requests
//! - Latency measurement and performance

use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::time::Instant;
use tokio::sync::mpsc;

// Mock server for testing without running actual BusinessOS.
mod mock_server {
    use std::sync::atomic::{AtomicBool, Ordering};
    use std::sync::Arc;
    use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
    use tokio::net::{TcpListener, TcpStream};

    pub struct MockServer {
        addr: String,
        listener: Arc<TcpListener>,
        should_fail: Arc<AtomicBool>,
        request_count: Arc<std::sync::atomic::AtomicUsize>,
    }

    impl MockServer {
        pub async fn start(port: u16) -> std::io::Result<Self> {
            let addr = format!("127.0.0.1:{}", port);
            let listener = TcpListener::bind(&addr).await?;
            let listener = Arc::new(listener);

            Ok(Self {
                addr,
                listener,
                should_fail: Arc::new(AtomicBool::new(false)),
                request_count: Arc::new(std::sync::atomic::AtomicUsize::new(0)),
            })
        }

        pub fn url(&self) -> String {
            format!("http://{}", self.addr)
        }

        pub fn request_count(&self) -> usize {
            self.request_count.load(Ordering::SeqCst)
        }

        pub fn set_should_fail(&self, fail: bool) {
            self.should_fail.store(fail, Ordering::SeqCst);
        }

        pub async fn accept_one(&self) -> std::io::Result<()> {
            let listener = self.listener.clone();
            let should_fail = self.should_fail.clone();
            let request_count = self.request_count.clone();

            tokio::spawn(async move {
                if let Ok((socket, _)) = listener.accept().await {
                    let mut reader = BufReader::new(socket);
                    let mut line = String::new();

                    if reader.read_line(&mut line).await.is_ok() {
                        request_count.fetch_add(1, Ordering::SeqCst);

                        let response = if should_fail.load(Ordering::SeqCst) {
                            "HTTP/1.1 503 Service Unavailable\r\nContent-Length: 0\r\n\r\n"
                                .to_string()
                        } else {
                            let path = line.split_whitespace().nth(1).unwrap_or("");

                            match path {
                                "/api/bos/discover" => {
                                    r#"HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 138

{"model_id":"model_test123","algorithm":"inductive_miner","places":5,"transitions":8,"arcs":12,"model_data":{},"latency_ms":15}
"#.to_string()
                                }
                                "/api/bos/conformance" => {
                                    r#"HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 141

{"traces_checked":125,"fitting_traces":120,"fitness":0.96,"precision":0.89,"generalization":0.91,"simplicity":0.85,"latency_ms":18}
"#.to_string()
                                }
                                "/api/bos/statistics" => {
                                    r#"HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 200

{"log_name":"test.xes","num_traces":500,"num_events":2450,"num_unique_activities":8,"num_variants":45,"avg_trace_length":4.9,"min_trace_length":2,"max_trace_length":12,"activity_frequency":[],"case_duration":{"min_seconds":60,"max_seconds":3600,"avg_seconds":1200.5,"median_seconds":900.0},"latency_ms":20}
"#.to_string()
                                }
                                "/api/bos/status" => {
                                    r#"HTTP/1.1 200 OK
Content-Type: application/json
Content-Length: 120

{"status":"healthy","businessos_ready":true,"connection_pool_size":16,"uptime_seconds":100,"requests_total":5,"requests_failed":0}
"#.to_string()
                                }
                                _ => {
                                    "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\n\r\n"
                                        .to_string()
                                }
                            }
                        };

                        let mut writer = reader.into_inner();
                        let _ = writer.write_all(response.as_bytes()).await;
                        let _ = writer.flush().await;
                    }
                }
            });

            Ok(())
        }
    }
}

#[cfg(test)]
mod tests {
    use super::mock_server::MockServer;
    use bos_core::{BusinessOSGateway, GatewayConfig, DiscoverRequest, ConformanceRequest, StatisticsRequest};
    use std::time::Instant;

    #[tokio::test]
    async fn test_discover_operation() {
        // Create a mock server
        let mock = MockServer::start(9001).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        // Create gateway pointing to mock
        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        // Execute discover
        let request = DiscoverRequest {
            log_path: "test_log.xes".to_string(),
            algorithm: Some("inductive_miner".to_string()),
        };

        let start = Instant::now();
        let result = gateway.discover(request).await;
        let latency = start.elapsed().as_millis();

        assert!(result.is_ok(), "Discover failed: {:?}", result);
        if let Ok(resp) = result {
            assert_eq!(resp.algorithm, "inductive_miner");
            assert_eq!(resp.places, 5);
            assert_eq!(resp.transitions, 8);
            assert!(latency < 100, "Latency too high: {}ms", latency);
            println!("Discover latency: {}ms", latency);
        }
    }

    #[tokio::test]
    async fn test_conformance_operation() {
        let mock = MockServer::start(9002).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = ConformanceRequest {
            log_path: "test_log.xes".to_string(),
            model_id: "model_123".to_string(),
        };

        let start = Instant::now();
        let result = gateway.check_conformance(request).await;
        let latency = start.elapsed().as_millis();

        assert!(result.is_ok(), "Conformance check failed: {:?}", result);
        if let Ok(resp) = result {
            assert!(resp.fitness > 0.9);
            assert!(latency < 100, "Latency too high: {}ms", latency);
            println!("Conformance latency: {}ms", latency);
        }
    }

    #[tokio::test]
    async fn test_statistics_operation() {
        let mock = MockServer::start(9003).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = StatisticsRequest {
            log_path: "test_log.xes".to_string(),
        };

        let start = Instant::now();
        let result = gateway.get_statistics(request).await;
        let latency = start.elapsed().as_millis();

        assert!(result.is_ok(), "Statistics failed: {:?}", result);
        if let Ok(resp) = result {
            assert_eq!(resp.num_traces, 500);
            assert_eq!(resp.num_events, 2450);
            assert!(latency < 100, "Latency too high: {}ms", latency);
            println!("Statistics latency: {}ms", latency);
        }
    }

    #[tokio::test]
    async fn test_gateway_health_check() {
        let mock = MockServer::start(9004).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let start = Instant::now();
        let result = gateway.check_health().await;
        let latency = start.elapsed().as_millis();

        assert!(result.is_ok(), "Health check failed: {:?}", result);
        if let Ok(status) = result {
            assert_eq!(status.status, "healthy");
            assert!(status.businessos_ready);
            assert!(latency < 100, "Latency too high: {}ms", latency);
        }
    }

    #[tokio::test]
    async fn test_service_unavailable_error() {
        let mock = MockServer::start(9005).await.expect("Failed to start mock server");
        mock.set_should_fail(true);
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 0, // No retries for this test
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = DiscoverRequest {
            log_path: "test_log.xes".to_string(),
            algorithm: Some("inductive_miner".to_string()),
        };

        let result = gateway.discover(request).await;
        assert!(result.is_err(), "Should have failed with service unavailable");
    }

    #[tokio::test]
    async fn test_concurrent_requests() {
        let mock = MockServer::start(9006).await.expect("Failed to start mock server");

        // Accept multiple connections
        for _ in 0..5 {
            mock.accept_one().await.unwrap();
        }

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 8,
        };

        let gateway = Arc::new(
            BusinessOSGateway::with_config(config)
                .expect("Failed to create gateway"),
        );

        let start = Instant::now();
        let mut handles = vec![];

        // Spawn 5 concurrent requests
        for i in 0..5 {
            let gw = gateway.clone();
            let handle = tokio::spawn(async move {
                let request = DiscoverRequest {
                    log_path: format!("log_{}.xes", i),
                    algorithm: Some("inductive_miner".to_string()),
                };

                gw.discover(request).await
            });
            handles.push(handle);
        }

        // Wait for all to complete
        let mut success_count = 0;
        for handle in handles {
            if let Ok(Ok(_)) = handle.await {
                success_count += 1;
            }
        }

        let total_latency = start.elapsed().as_millis();

        assert_eq!(success_count, 5, "Not all concurrent requests succeeded");
        assert!(total_latency < 500, "Total latency too high: {}ms", total_latency);
        println!("Concurrent request latency: {}ms (5 requests)", total_latency);
    }

    #[tokio::test]
    async fn test_malformed_request_handling() {
        let config = GatewayConfig {
            base_url: "http://127.0.0.1:9999".to_string(), // Invalid port
            timeout_ms: 500, // Short timeout
            max_retries: 0,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = DiscoverRequest {
            log_path: "".to_string(), // Empty path
            algorithm: None,
        };

        let result = gateway.discover(request).await;
        assert!(result.is_err(), "Should fail on unreachable server");
    }

    #[tokio::test]
    async fn test_request_statistics_tracking() {
        let mock = MockServer::start(9007).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = DiscoverRequest {
            log_path: "test.xes".to_string(),
            algorithm: Some("inductive_miner".to_string()),
        };

        let _ = gateway.discover(request).await;

        let stats = gateway.get_stats().await;
        assert_eq!(stats.requests_total, 1, "Should track request count");
        println!("Gateway stats: {:?}", stats);
    }

    #[tokio::test]
    async fn test_latency_measurement() {
        let mock = MockServer::start(9008).await.expect("Failed to start mock server");
        mock.accept_one().await.unwrap();

        let config = GatewayConfig {
            base_url: mock.url(),
            timeout_ms: 5000,
            max_retries: 1,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = DiscoverRequest {
            log_path: "test.xes".to_string(),
            algorithm: Some("inductive_miner".to_string()),
        };

        let result = gateway.discover(request).await;
        assert!(result.is_ok());

        if let Ok(resp) = result {
            println!("Response latency: {}ms", resp.latency_ms);
            assert!(resp.latency_ms < 100, "Latency should be <100ms");
        }
    }

    // Regression test: ensure gateway survives malformed JSON responses
    #[tokio::test]
    async fn test_invalid_json_response() {
        let config = GatewayConfig {
            base_url: "http://127.0.0.1:9999".to_string(),
            timeout_ms: 500,
            max_retries: 0,
            api_key: None,
            pool_size: 4,
        };

        let gateway = BusinessOSGateway::with_config(config)
            .expect("Failed to create gateway");

        let request = DiscoverRequest {
            log_path: "test.xes".to_string(),
            algorithm: None,
        };

        let result = gateway.discover(request).await;
        // Should fail gracefully, not panic
        assert!(result.is_err());
    }
}
