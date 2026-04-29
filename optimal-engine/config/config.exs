import Config

# Default paths point at the current directory; override in dev.exs / prod.exs
# or via OPTIMAL_ENGINE_ROOT env var at runtime.
config :optimal_engine,
  root_path: System.get_env("OPTIMAL_ENGINE_ROOT", File.cwd!()),
  db_path: System.get_env("OPTIMAL_ENGINE_DB", Path.join(File.cwd!(), ".optimal/index.db")),
  cache_path: System.get_env("OPTIMAL_ENGINE_CACHE", Path.join(File.cwd!(), ".optimal/cache")),
  topology_path:
    System.get_env("OPTIMAL_ENGINE_TOPOLOGY", Path.join(File.cwd!(), ".optimal/config.yaml")),
  topology_full_path:
    System.get_env("OPTIMAL_ENGINE_TOPOLOGY_FULL", Path.join(File.cwd!(), "topology.yaml"))

config :optimal_engine, :ollama,
  host: System.get_env("OLLAMA_HOST", "http://localhost:11434"),
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
