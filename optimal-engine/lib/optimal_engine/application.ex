defmodule OptimalEngine.Application do
  @moduledoc """
  OTP Application entry point for the Optimal Engine.

  Single unified supervision tree covering all four formerly-separate subsystems:
  the engine core, Knowledge (graph + OWL reasoning), Memory (episodic + cortex +
  learning), and Signal (classification + pub/sub + journal).

  Strategy: `:one_for_one` — each process is independent.

  Children are grouped below by origin for readability; at runtime they form a
  single flat supervision tree under `OptimalEngine.Supervisor`.
  """

  use Application
  require Logger

  @impl true
  def start(_type, _args) do
    children =
      base_children() ++ OptimalEngine.API.Endpoint.children()

    opts = [strategy: :one_for_one, name: OptimalEngine.Supervisor]

    case Supervisor.start_link(children, opts) do
      {:ok, pid} ->
        Logger.info("[OptimalEngine] Supervision tree started (#{inspect(pid)})")
        {:ok, pid}

      {:error, reason} = err ->
        Logger.error("[OptimalEngine] Failed to start: #{inspect(reason)}")
        err
    end
  end

  defp base_children do
    [
      # ── Observability (comes up first so every other child's metrics are captured) ─
      OptimalEngine.Telemetry,

      # ── Core engine (SQLite-backed, must start first) ─────────────────────
      OptimalEngine.Store,
      OptimalEngine.Pipeline.Router,
      OptimalEngine.Pipeline.Indexer,
      OptimalEngine.Retrieval.Search,
      OptimalEngine.Retrieval.L0Cache,
      OptimalEngine.Pipeline.Intake,
      OptimalEngine.Insight.Simulate,
      {Registry, keys: :unique, name: OptimalEngine.SessionRegistry},
      {DynamicSupervisor, name: OptimalEngine.SessionSupervisor, strategy: :one_for_one},

      # ── Knowledge subsystem (graph store + OWL 2 RL reasoner) ────────────
      # Concrete stores are opened on-demand via OptimalEngine.Knowledge.open/2,
      # so only the registry is permanent.
      {Registry, keys: :unique, name: OptimalEngine.Knowledge.Registry},

      # ── Memory subsystem (episodic, cortex, learning, session store) ─────
      {Registry, keys: :unique, name: OptimalEngine.Memory.SessionRegistry},
      OptimalEngine.Memory.Store.ETS,
      {DynamicSupervisor, name: OptimalEngine.Memory.SessionSupervisor, strategy: :one_for_one},
      OptimalEngine.Memory.Cortex,
      OptimalEngine.Memory.Learning,
      OptimalEngine.Memory.Surfacer,

      # ── Signal subsystem (pub/sub broker + journal for causality tracking) ─
      {OptimalEngine.Signal.PubSub, name: OptimalEngine.Signal.PubSub},
      {OptimalEngine.Signal.Journal, name: OptimalEngine.Signal.Journal},

      # ── HTTP API rate limiter (owns the ETS bucket table) ────────────────
      OptimalEngine.API.RateLimiter,

      # ── Wiki maintenance (periodic staleness scan + re-curation) ─────────
      OptimalEngine.Wiki.Scheduler
    ]
  end

  @impl true
  def stop(_state) do
    Logger.info("[OptimalEngine] Application stopping")
    :ok
  end
end
