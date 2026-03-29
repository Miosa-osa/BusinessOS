//! bos-config — Configuration management for the BusinessOS data layer CLI.
//!
//! Reads/writes `~/.bos/config.toml` for persistent settings.

use anyhow::{Context, Result};
use serde::{Deserialize, Serialize};
use std::path::PathBuf;

/// Top-level bos configuration.
#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BosConfig {
    #[serde(default)]
    pub general: GeneralConfig,
    #[serde(default)]
    pub sources: SourcesConfig,
    #[serde(default)]
    pub export: ExportConfig,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GeneralConfig {
    /// Base URL for BusinessOS API (e.g., "http://localhost:8001")
    #[serde(default = "default_api_url")]
    pub api_url: String,
    /// Default output format for exports
    #[serde(default = "default_format")]
    pub default_format: String,
    /// Authentication token (read from env BOS_TOKEN if empty)
    pub auth_token: Option<String>,
}

impl Default for GeneralConfig {
    fn default() -> Self {
        Self {
            api_url: default_api_url(),
            default_format: default_format(),
            auth_token: None,
        }
    }
}

fn default_api_url() -> String {
    "http://localhost:8001".to_string()
}

fn default_format() -> String {
    "odc".to_string()
}

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct SourcesConfig {
    /// Default data directory for file-based operations
    pub data_dir: Option<String>,
    /// Named source configurations
    #[serde(default)]
    pub named: std::collections::HashMap<String, SourceConfig>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SourceConfig {
    pub kind: String,
    pub url: Option<String>,
    pub path: Option<String>,
    #[serde(default)]
    pub options: std::collections::HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExportConfig {
    /// Default output directory
    pub output_dir: Option<String>,
    /// Include metadata in exports
    #[serde(default = "default_true")]
    pub include_metadata: bool,
}

impl Default for ExportConfig {
    fn default() -> Self {
        Self {
            output_dir: None,
            include_metadata: true,
        }
    }
}

fn default_true() -> bool {
    true
}

/// Get the path to the bos config file.
pub fn config_path() -> Result<PathBuf> {
    let dir = dirs::config_dir()
        .context("Could not determine config directory")?;
    Ok(dir.join("bos").join("config.toml"))
}

/// Get the path to the bos data directory.
pub fn data_dir() -> Result<PathBuf> {
    let dir = dirs::data_dir()
        .context("Could not determine data directory")?;
    Ok(dir.join("bos"))
}

/// Load configuration from disk. Returns default config if file doesn't exist.
pub fn load() -> Result<BosConfig> {
    let path = config_path()?;
    if !path.exists() {
        return Ok(BosConfig::default());
    }
    let content = std::fs::read_to_string(&path)
        .with_context(|| format!("Failed to read config: {}", path.display()))?;
    let config: BosConfig = toml::from_str(&content)
        .with_context(|| format!("Failed to parse config: {}", path.display()))?;
    Ok(config)
}

/// Save configuration to disk.
pub fn save(config: &BosConfig) -> Result<()> {
    let path = config_path()?;
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    let content = toml::to_string_pretty(config)
        .context("Failed to serialize config")?;
    std::fs::write(&path, content)
        .with_context(|| format!("Failed to write config: {}", path.display()))?;
    Ok(())
}

/// Ensure the bos data directory exists.
pub fn ensure_data_dir() -> Result<PathBuf> {
    let dir = data_dir()?;
    std::fs::create_dir_all(&dir)?;
    Ok(dir)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = BosConfig::default();
        assert_eq!(config.general.api_url, "http://localhost:8001");
        assert_eq!(config.general.default_format, "odc");
        assert!(config.general.auth_token.is_none());
    }

    #[test]
    fn test_config_roundtrip() {
        let config = BosConfig::default();
        let toml_str = toml::to_string_pretty(&config).unwrap();
        let parsed: BosConfig = toml::from_str(&toml_str).unwrap();
        assert_eq!(parsed.general.api_url, config.general.api_url);
    }

    #[test]
    fn test_config_path() {
        let path = config_path().unwrap();
        assert!(path.to_string_lossy().contains("bos"));
        assert!(path.to_string_lossy().contains("config.toml"));
    }
}
