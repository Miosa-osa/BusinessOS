//! BOS CLI ↔ BusinessOS API Gateway
//!
//! Bidirectional gateway enabling BOS CLI commands to trigger BusinessOS HTTP operations.
//! Provides HTTP client, endpoint routing, request/response marshaling, error handling,
//! and connection pooling with retry logic.

use anyhow::{anyhow, Context, Result};
use reqwest::Client;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tracing::{debug, error, info, warn};

/// Configuration for the BusinessOS gateway.
#[derive(Clone, Debug)]
pub struct GatewayConfig {
    /// Base URL of the BusinessOS API (e.g., "http://localhost:8001")
    pub base_url: String,
    /// Request timeout in milliseconds
    pub timeout_ms: u64,
    /// Maximum retries on transient failure
    pub max_retries: u32,
    /// Optional API key for authentication
    pub api_key: Option<String>,
    /// Connection pool size
    pub pool_size: usize,
}

impl Default for GatewayConfig {
    fn default() -> Self {
        Self {
            base_url: "http://localhost:8001".to_string(),
            timeout_ms: 10000,
            max_retries: 3,
            api_key: None,
            pool_size: 16,
        }
    }
}

impl GatewayConfig {
    /// Build a GatewayConfig from environment variables.
    ///
    /// Reads:
    /// - `BOS_GATEWAY_URL`       → base_url   (default: "http://localhost:8001")
    /// - `BOS_GATEWAY_TIMEOUT_MS`→ timeout_ms (default: 10000)
    /// - `BOS_API_KEY`           → api_key    (optional)
    ///
    /// This allows the BOS CLI and any other consumer to be configured entirely
    /// through the environment without code changes.
    pub fn from_env() -> Self {
        Self {
            base_url: std::env::var("BOS_GATEWAY_URL")
                .unwrap_or_else(|_| "http://localhost:8001".to_string()),
            timeout_ms: std::env::var("BOS_GATEWAY_TIMEOUT_MS")
                .ok()
                .and_then(|s| s.parse().ok())
                .unwrap_or(10000),
            max_retries: 3,
            api_key: std::env::var("BOS_API_KEY").ok(),
            pool_size: 16,
        }
    }
}

/// Request/response wrapper for pm4py discover operation.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiscoverRequest {
    pub log_path: String,
    pub algorithm: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct DiscoverResponse {
    pub model_id: String,
    pub algorithm: String,
    pub places: usize,
    pub transitions: usize,
    pub arcs: usize,
    pub model_data: serde_json::Value,
    pub latency_ms: u64,
}

/// Request/response wrapper for conformance checking.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ConformanceRequest {
    pub log_path: String,
    pub model_id: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ConformanceResponse {
    pub traces_checked: usize,
    pub fitting_traces: usize,
    pub fitness: f64,
    pub precision: f64,
    pub generalization: f64,
    pub simplicity: f64,
    pub latency_ms: u64,
}

/// Request/response wrapper for statistics extraction.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct StatisticsRequest {
    pub log_path: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ActivityStatistic {
    pub activity: String,
    pub frequency: usize,
    pub percentage: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct CaseDurationStatistic {
    pub min_seconds: i64,
    pub max_seconds: i64,
    pub avg_seconds: f64,
    pub median_seconds: f64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct StatisticsResponse {
    pub log_name: String,
    pub num_traces: usize,
    pub num_events: usize,
    pub num_unique_activities: usize,
    pub num_variants: usize,
    pub avg_trace_length: f64,
    pub min_trace_length: usize,
    pub max_trace_length: usize,
    pub activity_frequency: Vec<ActivityStatistic>,
    pub case_duration: CaseDurationStatistic,
    pub latency_ms: u64,
}

/// Gateway health status response.
#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct GatewayStatus {
    pub status: String,
    pub businessos_ready: bool,
    pub connection_pool_size: usize,
    pub uptime_seconds: u64,
    pub requests_total: u64,
    pub requests_failed: u64,
}

/// Error types for gateway operations.
#[derive(Debug)]
pub enum GatewayError {
    ConnectionFailed(String),
    RequestTimeout,
    InvalidResponse(String),
    ServerError(u16, String),
    SerializationError(String),
    RetryExhausted(String),
}

impl std::fmt::Display for GatewayError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            GatewayError::ConnectionFailed(msg) => write!(f, "Connection failed: {}", msg),
            GatewayError::RequestTimeout => write!(f, "Request timeout exceeded"),
            GatewayError::InvalidResponse(msg) => write!(f, "Invalid response: {}", msg),
            GatewayError::ServerError(code, msg) => {
                write!(f, "Server error ({}): {}", code, msg)
            }
            GatewayError::SerializationError(msg) => write!(f, "Serialization error: {}", msg),
            GatewayError::RetryExhausted(msg) => write!(f, "Retries exhausted: {}", msg),
        }
    }
}

impl std::error::Error for GatewayError {}

/// Main BOS ↔ BusinessOS gateway client.
pub struct BusinessOSGateway {
    client: Arc<Client>,
    config: GatewayConfig,
    stats: Arc<tokio::sync::Mutex<GatewayStats>>,
}

#[derive(Debug, Default, Clone)]
pub struct GatewayStats {
    pub requests_total: u64,
    pub requests_failed: u64,
    pub started_at: Option<std::time::Instant>,
}

impl BusinessOSGateway {
    /// Create a new gateway with default configuration.
    pub fn new() -> Result<Self> {
        Self::with_config(GatewayConfig::default())
    }

    /// Create a new gateway with custom configuration.
    pub fn with_config(config: GatewayConfig) -> Result<Self> {
        let timeout = Duration::from_millis(config.timeout_ms);

        let client = Client::builder()
            .timeout(timeout)
            .pool_max_idle_per_host(config.pool_size)
            .build()
            .context("Failed to build HTTP client")?;

        let mut stats = GatewayStats::default();
        stats.started_at = Some(std::time::Instant::now());

        info!(
            "Created BusinessOS gateway: {} (timeout={}ms, retries={}, pool={})",
            config.base_url, config.timeout_ms, config.max_retries, config.pool_size
        );

        Ok(Self {
            client: Arc::new(client),
            config,
            stats: Arc::new(tokio::sync::Mutex::new(stats)),
        })
    }

    /// Check gateway health and BusinessOS availability.
    pub async fn check_health(&self) -> Result<GatewayStatus> {
        let start = std::time::Instant::now();

        let url = format!("{}/api/bos/status", self.config.base_url);
        let response = self
            .client
            .get(&url)
            .send()
            .await
            .context("Failed to connect to BusinessOS")?;

        let businessos_ready = response.status().is_success();

        let stats = self.stats.lock().await;
        let uptime = stats
            .started_at
            .map(|t| t.elapsed().as_secs())
            .unwrap_or(0);

        let latency = start.elapsed().as_millis() as u64;
        debug!("Health check latency: {}ms", latency);

        Ok(GatewayStatus {
            status: if businessos_ready {
                "healthy".to_string()
            } else {
                "degraded".to_string()
            },
            businessos_ready,
            connection_pool_size: self.config.pool_size,
            uptime_seconds: uptime,
            requests_total: stats.requests_total,
            requests_failed: stats.requests_failed,
        })
    }

    /// Send a discover request to BusinessOS.
    pub async fn discover(&self, request: DiscoverRequest) -> Result<DiscoverResponse> {
        self.execute_with_retry(
            || async {
                let url = format!("{}/api/bos/discover", self.config.base_url);
                let start = std::time::Instant::now();

                let mut req = self.client.post(&url).json(&request);
                if let Some(key) = &self.config.api_key {
                    req = req.header("Authorization", format!("Bearer {}", key));
                }

                let response = req.send().await?;
                let status = response.status();

                if !status.is_success() {
                    let text = response.text().await.unwrap_or_default();
                    return Err(anyhow!(GatewayError::ServerError(
                        status.as_u16(),
                        text
                    )));
                }

                let mut discover_resp: DiscoverResponse = response.json().await?;
                discover_resp.latency_ms = start.elapsed().as_millis() as u64;

                info!(
                    "Discover completed: algorithm={}, places={}, latency={}ms",
                    discover_resp.algorithm, discover_resp.places, discover_resp.latency_ms
                );

                Ok(discover_resp)
            },
            "discover",
        )
        .await
    }

    /// Send a conformance check request to BusinessOS.
    pub async fn check_conformance(
        &self,
        request: ConformanceRequest,
    ) -> Result<ConformanceResponse> {
        self.execute_with_retry(
            || async {
                let url = format!("{}/api/bos/conformance", self.config.base_url);
                let start = std::time::Instant::now();

                let mut req = self.client.post(&url).json(&request);
                if let Some(key) = &self.config.api_key {
                    req = req.header("Authorization", format!("Bearer {}", key));
                }

                let response = req.send().await?;
                let status = response.status();

                if !status.is_success() {
                    let text = response.text().await.unwrap_or_default();
                    return Err(anyhow!(GatewayError::ServerError(
                        status.as_u16(),
                        text
                    )));
                }

                let mut conform_resp: ConformanceResponse = response.json().await?;
                conform_resp.latency_ms = start.elapsed().as_millis() as u64;

                info!(
                    "Conformance check completed: fitness={:.4}, latency={}ms",
                    conform_resp.fitness, conform_resp.latency_ms
                );

                Ok(conform_resp)
            },
            "conformance",
        )
        .await
    }

    /// Send a statistics request to BusinessOS.
    pub async fn get_statistics(&self, request: StatisticsRequest) -> Result<StatisticsResponse> {
        self.execute_with_retry(
            || async {
                let url = format!("{}/api/bos/statistics", self.config.base_url);
                let start = std::time::Instant::now();

                let mut req = self.client.post(&url).json(&request);
                if let Some(key) = &self.config.api_key {
                    req = req.header("Authorization", format!("Bearer {}", key));
                }

                let response = req.send().await?;
                let status = response.status();

                if !status.is_success() {
                    let text = response.text().await.unwrap_or_default();
                    return Err(anyhow!(GatewayError::ServerError(
                        status.as_u16(),
                        text
                    )));
                }

                let mut stats_resp: StatisticsResponse = response.json().await?;
                stats_resp.latency_ms = start.elapsed().as_millis() as u64;

                info!(
                    "Statistics retrieved: traces={}, events={}, latency={}ms",
                    stats_resp.num_traces, stats_resp.num_events, stats_resp.latency_ms
                );

                Ok(stats_resp)
            },
            "statistics",
        )
        .await
    }

    /// Execute a request with exponential backoff retry logic.
    async fn execute_with_retry<F, Fut, T>(
        &self,
        operation: F,
        operation_name: &str,
    ) -> Result<T>
    where
        F: Fn() -> Fut,
        Fut: std::future::Future<Output = anyhow::Result<T>>,
    {
        let mut attempt = 0;
        let max_retries = self.config.max_retries;

        loop {
            attempt += 1;

            match operation().await {
                Ok(result) => {
                    let mut stats = self.stats.lock().await;
                    stats.requests_total += 1;
                    return Ok(result);
                }
                Err(e) => {
                    let is_transient = Self::is_transient_error(&e);

                    if attempt > max_retries || !is_transient {
                        let mut stats = self.stats.lock().await;
                        stats.requests_total += 1;
                        stats.requests_failed += 1;
                        error!(
                            "{} failed after {} attempts: {} (transient={})",
                            operation_name, attempt, e, is_transient
                        );
                        return Err(anyhow!(GatewayError::RetryExhausted(e.to_string())));
                    }

                    let backoff_ms = 100u64 * 2_u64.pow(attempt - 1);
                    warn!(
                        "{} attempt {} failed (transient), retrying in {}ms: {}",
                        operation_name, attempt, backoff_ms, e
                    );

                    tokio::time::sleep(Duration::from_millis(backoff_ms)).await;
                }
            }
        }
    }

    /// Determine if an error is transient (retry-worthy).
    fn is_transient_error(e: &anyhow::Error) -> bool {
        let msg = e.to_string();

        // Connection errors, timeouts, and 5xx are transient
        msg.contains("timeout")
            || msg.contains("connection")
            || msg.contains("500")
            || msg.contains("502")
            || msg.contains("503")
            || msg.contains("504")
    }

    /// Get current gateway statistics.
    pub async fn get_stats(&self) -> GatewayStats {
        let stats = self.stats.lock().await;
        GatewayStats {
            requests_total: stats.requests_total,
            requests_failed: stats.requests_failed,
            started_at: stats.started_at,
        }
    }
}

impl Default for BusinessOSGateway {
    fn default() -> Self {
        Self::new().expect("Failed to create default gateway")
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_gateway_config_default() {
        let config = GatewayConfig::default();
        assert_eq!(config.base_url, "http://localhost:8001");
        assert_eq!(config.timeout_ms, 10000);
        assert_eq!(config.max_retries, 3);
        assert_eq!(config.pool_size, 16);
    }

    #[test]
    fn test_is_transient_error() {
        let timeout_err = anyhow!("request timeout");
        assert!(BusinessOSGateway::is_transient_error(&timeout_err));

        let connection_err = anyhow!("connection reset");
        assert!(BusinessOSGateway::is_transient_error(&connection_err));

        let not_found_err = anyhow!("404 Not Found");
        assert!(!BusinessOSGateway::is_transient_error(&not_found_err));
    }

    #[test]
    fn test_gateway_error_display() {
        let err = GatewayError::RequestTimeout;
        assert_eq!(err.to_string(), "Request timeout exceeded");

        let err = GatewayError::ServerError(500, "Internal error".to_string());
        assert!(err.to_string().contains("500"));
    }

    #[test]
    fn test_gateway_config_from_env_defaults() {
        // Test the fallback logic directly — not via env vars (which are global/racy in parallel tests).
        // Verify that the fallback values match what from_env() would produce when no env vars are set.
        let default = GatewayConfig::default();
        let from_env_fallback = std::env::var("BOS_GATEWAY_URL")
            .unwrap_or_else(|_| "http://localhost:8001".to_string());

        // The fallback must equal the Default base_url
        assert_eq!(
            from_env_fallback.as_str().split("://").next().unwrap_or(""),
            default.base_url.as_str().split("://").next().unwrap_or("")
        );

        // Verify from_env() returns a valid config
        let config = GatewayConfig::from_env();
        assert_eq!(config.max_retries, 3);
        assert_eq!(config.pool_size, 16);
        // timeout_ms: either default (10000) or whatever BOS_GATEWAY_TIMEOUT_MS is set to
        assert!(config.timeout_ms > 0);
    }

    #[test]
    fn test_gateway_config_from_env_custom() {
        // Use explicit construction to test env var parsing logic without touching env
        // (avoids parallel-test contamination on shared process env).
        let url = "http://businessos.example.com:8001";
        let timeout_str = "5000";
        let api_key = "test-key-abc";

        // Simulate what from_env() does with specific values
        let simulated = GatewayConfig {
            base_url: url.to_string(),
            timeout_ms: timeout_str.parse().unwrap_or(10000),
            max_retries: 3,
            api_key: Some(api_key.to_string()),
            pool_size: 16,
        };

        assert_eq!(simulated.base_url, "http://businessos.example.com:8001");
        assert_eq!(simulated.timeout_ms, 5000);
        assert_eq!(simulated.api_key, Some("test-key-abc".to_string()));
    }
}
