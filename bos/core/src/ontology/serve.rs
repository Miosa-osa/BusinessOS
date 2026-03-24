//! Self-Describing Workspace — serves an ODCS workspace's ontology as a
//! SPARQL-queryable HTTP endpoint.
//!
//! Allows any AI agent to discover a workspace's schema and data by sending
//! SPARQL queries to a local HTTP server.

use crate::ontology::mapping::MappingConfig;
use crate::rdf::store::TripleStore;
use anyhow::{Context, Result};
use oxigraph::io::RdfFormat;
use oxigraph::io::RdfParser;
use std::io::Cursor;
use std::net::TcpListener;
use std::path::Path;

/// Configuration for the semantic server.
#[derive(Debug, Clone)]
pub struct ServeConfig {
    /// Host to bind to [default: 127.0.0.1]
    pub host: String,
    /// Port to bind to [default: 7878]
    pub port: u16,
    /// Path to RDF data file to preload
    pub rdf_path: Option<String>,
}

impl Default for ServeConfig {
    fn default() -> Self {
        Self {
            host: "127.0.0.1".to_string(),
            port: 7878,
            rdf_path: None,
        }
    }
}

/// Serve the workspace ontology as a SPARQL endpoint.
///
/// Starts a minimal HTTP server that accepts:
/// - `POST /sparql` with SPARQL query in body → JSON results
/// - `GET /sparql` → returns available types
///
/// Blocks the calling thread.
pub fn serve(config: ServeConfig) -> Result<()> {
    let store = TripleStore::new();

    // Preload RDF data if provided
    if let Some(ref rdf_path) = config.rdf_path {
        let path = Path::new(rdf_path);
        let data = std::fs::read_to_string(path)
            .with_context(|| format!("Failed to read RDF file: {}", path.display()))?;
        load_rdf_into_store(&store, &data)?;
    }

    let addr = format!("{}:{}", config.host, config.port);
    let listener = TcpListener::bind(&addr)
        .with_context(|| format!("Failed to bind to {}", addr))?;

    eprintln!("Semantic server listening on http://{}/sparql", addr);
    eprintln!("Press Ctrl+C to stop.");

    // Simple single-threaded HTTP server loop
    for stream in listener.incoming() {
        match stream {
            Ok(stream) => {
                if let Err(e) = handle_connection(&store, stream) {
                    tracing::debug!("Connection error: {}", e);
                }
            }
            Err(e) => {
                tracing::debug!("Accept error: {}", e);
            }
        }
    }

    Ok(())
}

fn load_rdf_into_store(store: &TripleStore, data: &str) -> Result<()> {
    let trimmed = data.trim_start();
    let format = if trimmed.starts_with("@prefix") || trimmed.starts_with("PREFIX") {
        RdfFormat::Turtle
    } else {
        RdfFormat::NTriples
    };

    let cursor = Cursor::new(data.as_bytes());
    let parser = RdfParser::from_format(format).for_reader(cursor);
    for quad in parser {
        let quad = quad
            .map_err(|e| anyhow::anyhow!("Parse error: {e}"))?;
        let _ = store.insert(quad.subject, quad.predicate, quad.object);
    }
    Ok(())
}

fn handle_connection(store: &TripleStore, stream: std::net::TcpStream) -> Result<()> {
    use std::io::{BufRead, Write};

    let mut stream = stream;
    let _ = stream.set_read_timeout(Some(std::time::Duration::from_secs(5)));
    let _ = stream.set_write_timeout(Some(std::time::Duration::from_secs(5)));

    let mut reader = std::io::BufReader::new(&stream);
    let mut request_line = String::new();
    reader.read_line(&mut request_line)?;

    let parts: Vec<&str> = request_line.split_whitespace().collect();
    if parts.len() < 2 {
        let _ = write!(stream, "HTTP/1.1 400 Bad Request\r\n\r\n");
        return Ok(());
    }

    let method = parts[0];
    let path = parts[1];

    if path != "/sparql" {
        let _ = write!(stream, "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\n\r\nNot found. Use /sparql\r\n");
        return Ok(());
    }

    let response = match method {
        "GET" => {
            let types = store.predicates();
            let body = serde_json::to_string_pretty(&types).unwrap_or_else(|_| "[]".to_string());
            format!(
                "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\n\r\n{}",
                body
            )
        }
        "OPTIONS" => {
            "HTTP/1.1 204 No Content\r\nAccess-Control-Allow-Origin: *\r\nAccess-Control-Allow-Methods: POST, GET, OPTIONS\r\nAccess-Control-Allow-Headers: Content-Type\r\n\r\n".to_string()
        }
        "POST" => {
            let mut content_length = 0usize;
            loop {
                let mut line = String::new();
                reader.read_line(&mut line)?;
                if line == "\r\n" || line == "\n" {
                    break;
                }
                if line.to_lowercase().starts_with("content-length:") {
                    let val = line.split(':').nth(1).unwrap_or("").trim();
                    content_length = val.parse::<usize>().unwrap_or(0);
                }
            }

            let mut body = vec![0u8; content_length];
            if content_length > 0 {
                std::io::Read::read_exact(&mut reader, &mut body)?;
            }
            let query = String::from_utf8_lossy(&body);

            match store.query_sparql(&query) {
                Ok(rows) => {
                    let json = serde_json::to_string_pretty(&rows)
                        .unwrap_or_else(|_| "[]".to_string());
                    format!(
                        "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\nAccess-Control-Allow-Origin: *\r\n\r\n{}",
                        json
                    )
                }
                Err(e) => {
                    format!(
                        "HTTP/1.1 400 Bad Request\r\nContent-Type: text/plain\r\nAccess-Control-Allow-Origin: *\r\n\r\nSPARQL error: {}",
                        e
                    )
                }
            }
        }
        _ => "HTTP/1.1 405 Method Not Allowed\r\n\r\n".to_string(),
    };
    let _ = write!(stream, "{}", response);
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_serve_config_defaults() {
        let config = ServeConfig::default();
        assert_eq!(config.host, "127.0.0.1");
        assert_eq!(config.port, 7878);
        assert!(config.rdf_path.is_none());
    }

    #[test]
    fn test_serve_config_custom() {
        let config = ServeConfig {
            host: "0.0.0.0".to_string(),
            port: 9000,
            rdf_path: Some("/tmp/data.nt".to_string()),
        };
        assert_eq!(config.host, "0.0.0.0");
        assert_eq!(config.port, 9000);
    }
}
