defmodule OptimalEngine.MixProject do
  use Mix.Project

  @version "0.1.0"

  def project do
    [
      app: :optimal_engine,
      version: @version,
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      aliases: aliases(),
      elixirc_paths: elixirc_paths(Mix.env()),
      dialyzer: [plt_add_apps: [:mix]],
      description: description(),
      package: package(),
      escript: escript(),
      releases: releases()
    ]
  end

  # Single-file CLI for development and trusted environments that already have
  # Erlang installed.  For cross-platform standalone distribution (no Erlang on
  # the target machine) build the Burrito release below: `mix release optimal`.
  defp escript do
    [
      main_module: OptimalEngine.CLI,
      name: "optimal",
      app: nil
    ]
  end

  # `mix release optimal` produces a self-contained OTP release tarball in
  # _build/prod/rel/optimal/.  Wrap with Burrito for single-binary distribution
  # by adding the burrito config block here when we're ready.
  defp releases do
    [
      optimal: [
        include_executables_for: [:unix],
        applications: [optimal_engine: :permanent]
      ]
    ]
  end

  def application do
    [
      extra_applications: [:logger, :crypto, :inets, :ssl, :mnesia],
      mod: {OptimalEngine.Application, []}
    ]
  end

  defp elixirc_paths(:test), do: ["lib", "test/support"]
  defp elixirc_paths(_), do: ["lib"]

  defp deps do
    [
      # Storage & serialization
      {:exqlite, "~> 0.27"},
      {:jason, "~> 1.4"},
      {:yaml_elixir, "~> 2.11"},

      # Knowledge subsystem (RDF triples, OWL 2 RL reasoning)
      {:rdf, "~> 2.0"},

      # Signal subsystem (CloudEvents envelopes)
      {:nimble_options, "~> 1.1"},
      {:uuid, "~> 1.1"},

      # HTTP API (Plug.Router for the graph visualizer endpoint)
      {:plug, "~> 1.15"},
      {:plug_cowboy, "~> 2.7"},

      # Parser backends (Phase 2)
      {:nimble_csv, "~> 1.2"},
      {:floki, "~> 0.36"},

      # Observability
      {:telemetry, "~> 1.2"},

      # Auth
      {:bcrypt_elixir, "~> 3.1"},

      # Dev / test
      {:credo, "~> 1.7", only: [:dev, :test], runtime: false},
      {:dialyxir, "~> 1.4", only: [:dev], runtime: false}
    ]
  end

  defp aliases do
    [
      "optimal.index": ["run -e 'Mix.Tasks.Optimal.Index.run([])'"],
      "optimal.search": ["run -e 'Mix.Tasks.Optimal.Search.run(System.argv())'"],
      # Wipe the test SQLite store before every run so schema drift from
      # an older migration set can't leak into the current test process.
      test: [
        "cmd rm -f /tmp/optimal_engine_test_0.db /tmp/optimal_engine_test_0.db-wal /tmp/optimal_engine_test_0.db-shm",
        "test"
      ]
    ]
  end

  defp description do
    "Signal-native context storage engine for AI agents. SQLite + FTS5 + " <>
      "tiered L0/L1/L2 loading + OWL 2 RL reasoning."
  end

  defp package do
    [
      licenses: ["UNLICENSED"],
      links: %{},
      files: ~w(lib priv config mix.exs README.md)
    ]
  end
end
