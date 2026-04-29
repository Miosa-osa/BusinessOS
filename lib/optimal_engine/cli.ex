defmodule OptimalEngine.CLI do
  @moduledoc """
  Command-line entry point for the Optimal Engine.

  Exposes every `mix optimal.*` task as a subcommand of a single `optimal`
  binary, so any agent runtime (Python, Node, Go, shell scripts, Claude Agent
  SDK, etc.) can shell out without needing Elixir on its path.

  ## Usage

      optimal <subcommand> [args...]
      optimal --help
      optimal --version

  ## Subcommands

  The CLI proxies to the same `Mix.Tasks.Optimal.*` modules used by `mix`, so
  every task listed in `mix help | grep optimal` is available here too.
  """

  @version Mix.Project.config()[:version] || "0.1.0"

  @subcommands %{
    # Ingestion
    "ingest" => Mix.Tasks.Optimal.Ingest,
    "intake" => Mix.Tasks.Optimal.Intake,
    "index" => Mix.Tasks.Optimal.Index,
    # Retrieval
    "search" => Mix.Tasks.Optimal.Search,
    "read" => Mix.Tasks.Optimal.Read,
    "ls" => Mix.Tasks.Optimal.Ls,
    "l0" => Mix.Tasks.Optimal.L0,
    "assemble" => Mix.Tasks.Optimal.Assemble,
    "rag" => Mix.Tasks.Optimal.Rag,
    # Graph analysis
    "graph" => Mix.Tasks.Optimal.Graph,
    "reflect" => Mix.Tasks.Optimal.Reflect,
    "reweave" => Mix.Tasks.Optimal.Reweave,
    "simulate" => Mix.Tasks.Optimal.Simulate,
    "impact" => Mix.Tasks.Optimal.Impact,
    # Learning loop
    "remember" => Mix.Tasks.Optimal.Remember,
    "rethink" => Mix.Tasks.Optimal.Rethink,
    "knowledge" => Mix.Tasks.Optimal.Knowledge,
    # Health & verification
    "health" => Mix.Tasks.Optimal.Health,
    "verify" => Mix.Tasks.Optimal.Verify,
    "stats" => Mix.Tasks.Optimal.Stats,
    # HTTP API & visualizer
    "api" => Mix.Tasks.Optimal.Api,
    "graph-ui" => Mix.Tasks.Optimal.GraphUi,
    # Spec tooling
    "spec.drift" => Mix.Tasks.Optimal.Spec.Drift,
    "spec.init" => Mix.Tasks.Optimal.Spec.Init,
    "spec.report" => Mix.Tasks.Optimal.Spec.Report
  }

  @ordered_groups [
    {"Ingestion", ["ingest", "intake", "index"]},
    {"Retrieval", ["search", "read", "ls", "l0", "assemble", "rag"]},
    {"Graph analysis", ["graph", "reflect", "reweave", "simulate", "impact"]},
    {"Learning loop", ["remember", "rethink", "knowledge"]},
    {"Health & verification", ["health", "verify", "stats"]},
    {"HTTP API & visualizer", ["api", "graph-ui"]},
    {"Spec tooling", ["spec.drift", "spec.init", "spec.report"]}
  ]

  # Shortdocs are hardcoded here because escripts strip @shortdoc metadata.
  @shortdocs %{
    "ingest" => "Classify, route, write signal files, and index content",
    "intake" => "Interactive multi-line intake from stdin",
    "index" => "Full reindex of all markdown files",
    "search" => "Hybrid BM25 + temporal + graph-boosted search",
    "read" => "Read a context by optimal:// URI at a given tier",
    "ls" => "List contexts under an optimal:// URI",
    "l0" => "Print the always-loaded L0 context (~100 tokens per node)",
    "assemble" => "Build a tiered (L0/L1/L2) context bundle for a topic",
    "rag" => "Return LLM-ready retrieved chunks for a query",
    "graph" => "Knowledge graph stats, triangles, clusters, hubs",
    "reflect" => "Find missing edges from entity co-occurrences",
    "reweave" => "Find stale contexts on a topic and suggest updates",
    "simulate" => "Run a 'what if' scenario through the graph",
    "impact" => "Impact analysis for an entity or node",
    "remember" => "Store observations; mine friction patterns",
    "rethink" => "Synthesize observations into actionable knowledge",
    "knowledge" => "Knowledge graph + SICA learning operations",
    "health" => "Diagnostic checks on the knowledge base",
    "verify" => "Cold-read test of L0 abstract fidelity",
    "stats" => "Store statistics (counts, sizes, token budgets)",
    "api" => "Start the HTTP API on port 4200",
    "graph-ui" => "Launch the graph visualizer against a running API",
    "spec.drift" => "Detect code changes without spec updates",
    "spec.init" => "Scaffold the .spec/ directory",
    "spec.report" => "Spec coverage and verification summary"
  }

  @doc """
  Main entry point for the escript.

  Takes the raw argv from the OS, parses the subcommand, ensures the
  application and its dependencies are started, then delegates to the matching
  `Mix.Tasks.Optimal.*` task module.
  """
  @spec main([String.t()]) :: any()
  def main(argv) do
    case argv do
      [] -> print_help_and_exit(0)
      ["--help"] -> print_help_and_exit(0)
      ["-h"] -> print_help_and_exit(0)
      ["help"] -> print_help_and_exit(0)
      ["--version"] -> print_version_and_exit()
      ["-v"] -> print_version_and_exit()
      [sub | rest] -> dispatch(sub, rest)
    end
  end

  # ── Dispatch ─────────────────────────────────────────────────────────────

  defp dispatch(sub, rest) do
    case Map.fetch(@subcommands, sub) do
      {:ok, module} ->
        ensure_started()
        module.run(rest)
        :ok

      :error ->
        IO.puts(:stderr, "optimal: unknown subcommand '#{sub}'")
        IO.puts(:stderr, "")
        print_help(:stderr)
        System.halt(64)
    end
  end

  # ── Boot ─────────────────────────────────────────────────────────────────

  defp ensure_started do
    # escript packages don't auto-start applications the way `mix` does
    Application.ensure_all_started(:optimal_engine)
  end

  # ── Help ─────────────────────────────────────────────────────────────────

  defp print_help_and_exit(code) do
    print_help(:stdio)
    System.halt(code)
  end

  defp print_version_and_exit do
    IO.puts("optimal #{@version}")
    System.halt(0)
  end

  defp print_help(io) do
    lines = [
      "Optimal Engine CLI — signal-native context storage for AI agents.",
      "",
      "  usage: optimal <subcommand> [args...]",
      "         optimal --help",
      "         optimal --version",
      "",
      "Subcommands by group:"
    ]

    Enum.each(lines, &IO.puts(io, &1))

    Enum.each(@ordered_groups, fn {group, names} ->
      IO.puts(io, "")
      IO.puts(io, "  #{group}:")

      Enum.each(names, fn name ->
        short = Map.get(@shortdocs, name, "")
        IO.puts(io, "    #{String.pad_trailing(name, 14)} #{short}")
      end)
    end)

    IO.puts(io, "")

    IO.puts(
      io,
      "Run 'optimal <subcommand> --help' for per-command detail where supported."
    )
  end
end
