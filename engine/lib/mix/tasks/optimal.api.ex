defmodule Mix.Tasks.Optimal.Api do
  @moduledoc """
  Starts the OptimalOS HTTP API server.

  The server listens on port 4200 and exposes the knowledge graph as JSON.

  ## Usage

      mix optimal.api

  ## Endpoints

      GET /api/graph           — Full knowledge graph (edges + entities + nodes)
      GET /api/graph/hubs      — Hub entities (degree > 2σ above mean)
      GET /api/graph/triangles — Open triangles (synthesis opportunities, ?limit=20)
      GET /api/graph/clusters  — Connected components
      GET /api/graph/reflect   — Co-occurrence gaps (?min=2)
      GET /api/node/:id        — Subgraph for a single node
      GET /api/search          — Full-text search (?q=query&limit=10)
      GET /api/l0              — L0 context cache
      GET /api/health          — Health diagnostics
  """

  use Mix.Task

  @shortdoc "Start the OptimalOS HTTP API server on port 4200"

  @impl Mix.Task
  def run(_args) do
    Mix.Task.run("app.start")

    IO.puts("""
    [OptimalEngine API] Listening on http://localhost:4200

    Endpoints:
      GET /api/graph              Full knowledge graph
      GET /api/graph/hubs         Hub entities
      GET /api/graph/triangles    Synthesis opportunities
      GET /api/graph/clusters     Connected components
      GET /api/graph/reflect      Co-occurrence gaps
      GET /api/node/:id           Node subgraph
      GET /api/search?q=<query>   Full-text search
      GET /api/l0                 L0 context cache
      GET /api/health             Health diagnostics

    Press Ctrl+C to stop.
    """)

    Process.sleep(:infinity)
  end
end
