defmodule OptimalEngine.MixProject do
  use Mix.Project

  def project do
    [
      app: :optimal_engine,
      version: "0.1.0",
      elixir: "~> 1.17",
      start_permanent: Mix.env() == :prod,
      deps: deps(),
      aliases: aliases(),
      elixirc_paths: elixirc_paths(Mix.env()),
      dialyzer: [plt_add_apps: [:mix]]
    ]
  end

  def application do
    [
      extra_applications: [:logger, :crypto, :inets, :ssl],
      mod: {OptimalEngine.Application, []}
    ]
  end

  defp elixirc_paths(:test), do: ["lib", "test/support"]
  defp elixirc_paths(_), do: ["lib"]

  defp deps do
    [
      {:exqlite, "~> 0.27"},
      {:yaml_elixir, "~> 2.11"},
      {:jason, "~> 1.4"},
      {:credo, "~> 1.7", only: [:dev, :test], runtime: false},
      {:dialyxir, "~> 1.4", only: [:dev], runtime: false},

      # MIOSA subsystems — composed into the Optimal Engine
      {:miosa_knowledge, path: "../../MIOSA/code/miosa_knowledge"},
      {:miosa_memory, path: "../../MIOSA/code/miosa_memory"},
      {:miosa_signal, path: "../../MIOSA/code/miosa_signal"}
    ]
  end

  defp aliases do
    [
      "optimal.index": ["run -e 'Mix.Tasks.Optimal.Index.run([])'"],
      "optimal.search": ["run -e 'Mix.Tasks.Optimal.Search.run(System.argv())'"]
    ]
  end
end
