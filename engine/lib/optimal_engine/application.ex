defmodule OptimalEngine.Application do
  @moduledoc """
  OTP Application entry point for the Optimal Context Engine.

  Supervision tree:
  - Store              — SQLite + ETS hot cache
  - Router             — Routing rules loader
  - Indexer            — File crawler + indexing pipeline
  - SearchEngine       — Query executor
  - L0Cache            — Always-loaded context generator
  - Intake             — Signal ingestion pipeline
  - Simulator          — Scenario planning + impact analysis engine
  - SessionRegistry    — Registry for active sessions (via name lookup)
  - SessionSupervisor  — DynamicSupervisor for per-session GenServers

  Strategy: `:one_for_one` — each process is independent.

  The Store is listed first so it is started before any process that writes to SQLite.
  Sessions use `:temporary` restart — they are not restarted on crash (conversation
  state is ephemeral; if a session process crashes, the client starts a new one).
  """

  use Application
  require Logger

  @impl true
  def start(_type, _args) do
    children = [
      # Core engine
      OptimalEngine.Store,
      OptimalEngine.Router,
      OptimalEngine.Indexer,
      OptimalEngine.SearchEngine,
      OptimalEngine.L0Cache,
      OptimalEngine.Intake,
      OptimalEngine.Simulator,
      {Registry, keys: :unique, name: OptimalEngine.SessionRegistry},
      {DynamicSupervisor, name: OptimalEngine.SessionSupervisor, strategy: :one_for_one},

      # MIOSA Memory subsystem (no Application module — we start GenServers here)
      MiosaMemory.Episodic,
      MiosaMemory.Cortex,
      MiosaMemory.Learning,

      # Bridge.Knowledge is now a stateless module — no GenServer needed
    ]

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

  @impl true
  def stop(_state) do
    Logger.info("[OptimalEngine] Application stopping")
    :ok
  end
end
