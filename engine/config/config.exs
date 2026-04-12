import Config

config :optimal_engine,
  root_path: "/Users/rhl/Desktop/OptimalOS",
  db_path: "/Users/rhl/Desktop/OptimalOS/.system/index.db",
  cache_path: "/Users/rhl/Desktop/OptimalOS/.system/cache",
  topology_path: "/Users/rhl/Desktop/OptimalOS/.system/config.yaml",
  topology_full_path: "/Users/rhl/Desktop/OptimalOS/topology.yaml"

config :optimal_engine, :ollama,
  host: "http://localhost:11434",
  embed_model: "nomic-embed-text",
  generate_model: "qwen3:8b",
  timeout_ms: 30_000

config :optimal_engine, :hybrid_search,
  alpha: 0.6,
  vector_enabled: true

config :logger, :console,
  format: "[$level] $message\n",
  metadata: [:module, :request_id]

import_config "#{config_env()}.exs"
