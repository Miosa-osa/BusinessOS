import Config

# Use bcrypt cost 4 in tests — functionally identical, ~80x faster than cost 12.
config :optimal_engine, :auth,
  auth_required: false,
  bcrypt_cost: 4

config :optimal_engine,
  root_path: "/Users/rhl/Desktop/OptimalOS",
  db_path: "/tmp/optimal_engine_test_#{System.get_env("MIX_TEST_PARTITION", "0")}.db",
  cache_path: "/tmp/optimal_engine_test_cache",
  topology_path: "/Users/rhl/Desktop/OptimalOS/.system/config.yaml",
  topology_full_path: "/Users/rhl/Desktop/OptimalOS/topology.yaml"
