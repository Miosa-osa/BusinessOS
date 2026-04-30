defmodule OptimalEngine.API.Router do
  @moduledoc """
  Plug-based HTTP JSON API for OptimalOS knowledge graph data.

  Endpoints:
    GET /api/graph           — Full graph (all edges + entity summary + node summary)
    GET /api/graph/hubs      — Hub entities (degree > 2σ)
    GET /api/graph/triangles — Open triangles (synthesis opportunities)
    GET /api/graph/clusters  — Connected components
    GET /api/graph/reflect   — Co-occurrence gaps (?min=2)
    GET /api/node/:node_id   — Subgraph for one node
    GET /api/search          — Full-text search (?q=query&limit=10)
    GET /api/grep            — Hybrid semantic+literal grep with full signal trace (?q=query&workspace=W&intent=I&scale=S&limit=N)
    GET /api/l0              — L0 context cache
    GET /api/health          — Health diagnostics
    GET /api/profile         — 4-tier workspace profile snapshot

  All responses are JSON with `Content-Type: application/json`.
  CORS headers are set on every response (GET + OPTIONS allowed).
  """

  use Plug.Router

  alias OptimalEngine.Graph.Analyzer, as: GraphAnalyzer
  alias OptimalEngine.Health
  alias OptimalEngine.Insight.Health, as: HealthDiagnostics
  alias OptimalEngine.Graph.Reflector, as: Reflector
  alias OptimalEngine.Profile
  alias OptimalEngine.Retrieval
  alias OptimalEngine.Retrieval.Grep
  alias OptimalEngine.Retrieval.RagStream
  alias OptimalEngine.Retrieval.Receiver
  alias OptimalEngine.Auth.ApiKey
  alias OptimalEngine.Store
  alias OptimalEngine.Telemetry
  alias OptimalEngine.Wiki

  plug(:cors)
  plug(:match)
  plug(Plug.Parsers, parsers: [:json], json_decoder: Jason, pass: ["application/json"])
  plug(OptimalEngine.API.RateLimitPlug)
  plug(OptimalEngine.API.AuthPlug)
  plug(:dispatch)

  # ---------------------------------------------------------------------------
  # CORS preflight
  # ---------------------------------------------------------------------------

  options _ do
    send_resp(conn, 204, "")
  end

  # ---------------------------------------------------------------------------
  # Routes
  # ---------------------------------------------------------------------------

  # Full graph payload — every context as a renderable node + visual edges + entities
  get "/api/graph" do
    edges = fetch_visual_edges()
    entities = fetch_entity_summary()
    nodes = fetch_node_summary()
    contexts = fetch_all_contexts()

    json(conn, %{
      contexts: contexts,
      edges: edges,
      entities: format_entities(entities),
      nodes: nodes,
      stats: %{
        context_count: length(contexts),
        edge_count: length(edges),
        entity_count: length(entities)
      }
    })
  end

  get "/api/graph/hubs" do
    {:ok, hubs} = GraphAnalyzer.hubs()
    json(conn, %{hubs: format_hubs(hubs)})
  end

  get "/api/graph/triangles" do
    limit = conn |> query_param("limit", "20") |> parse_int(20)
    {:ok, triangles} = GraphAnalyzer.triangles(limit: limit)
    json(conn, %{triangles: format_triangles(triangles)})
  end

  get "/api/graph/clusters" do
    {:ok, clusters} = GraphAnalyzer.clusters()
    json(conn, %{clusters: format_clusters(clusters)})
  end

  get "/api/graph/reflect" do
    min = conn |> query_param("min", "2") |> parse_int(2)
    {:ok, gaps} = Reflector.reflect(min_cooccurrences: min)
    json(conn, %{gaps: format_gaps(gaps)})
  end

  get "/api/node/:node_id" do
    contexts = fetch_node_contexts(node_id)
    edges = fetch_node_edges(node_id)

    json(conn, %{
      node: node_id,
      contexts: contexts,
      edges: format_edges(edges)
    })
  end

  get "/api/search" do
    q = query_param(conn, "q", "")
    limit = conn |> query_param("limit", "10") |> parse_int(10)
    workspace = query_param(conn, "workspace", "default")

    results =
      if q == "" do
        []
      else
        case OptimalEngine.search(q, limit: limit, workspace_id: workspace) do
          {:ok, contexts} -> format_search_results(contexts)
          _ -> []
        end
      end

    json(conn, %{query: q, results: results})
  end

  # GET /api/grep — hybrid semantic + literal grep over a workspace.
  #
  # Unlike /api/search (which returns context-level metadata), /api/grep
  # returns chunk-level matches with the full signal trace per match:
  # slug, scale, intent, sn_ratio, modality, and a 200-char snippet.
  #
  # Params:
  #   q=<query>       search terms (required)
  #   workspace=<id>  workspace id (default: "default")
  #   intent=<val>    filter by intent (one of 10 canonical values)
  #   scale=<val>     filter by scale: document | section | paragraph | chunk
  #   modality=<val>  filter by modality (e.g. text | image | audio)
  #   limit=<n>       max results (default 25)
  #   literal=true    force literal FTS match — skip semantic/vector search
  #   path=<prefix>   restrict to node slug or slug prefix
  get "/api/grep" do
    q = query_param(conn, "q", "")
    workspace_id = query_param(conn, "workspace", "default")
    intent = query_param(conn, "intent", "")
    scale = query_param(conn, "scale", "")
    modality = query_param(conn, "modality", "")
    limit = conn |> query_param("limit", "25") |> parse_int(25)
    literal? = conn |> query_param("literal", "false") |> parse_bool(false)
    path_prefix = query_param(conn, "path", "")

    if q == "" do
      send_resp(conn, 400, Jason.encode!(%{error: "q is required"}))
    else
      grep_opts =
        [
          workspace_id: workspace_id,
          limit: limit,
          literal: literal?
        ]
        |> api_maybe_put(:intent, parse_atom(intent, nil))
        |> api_maybe_put(:scale, parse_atom(scale, nil))
        |> api_maybe_put(:modality, parse_atom(modality, nil))
        |> api_maybe_put(:path_prefix, if(path_prefix == "", do: nil, else: path_prefix))

      case Grep.grep(q, grep_opts) do
        {:ok, matches} ->
          json(conn, %{
            query: q,
            workspace_id: workspace_id,
            results:
              Enum.map(matches, fn m ->
                %{
                  slug: m.slug,
                  scale: to_string(m.scale),
                  intent: if(m.intent, do: to_string(m.intent), else: nil),
                  sn_ratio: m.sn_ratio,
                  modality: if(m.modality, do: to_string(m.modality), else: nil),
                  snippet: m.snippet,
                  score: m.score
                }
              end)
          })

        {:error, reason} ->
          send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
      end
    end
  end

  get "/api/l0" do
    json(conn, %{l0: OptimalEngine.l0()})
  end

  # GET /api/profile — 4-tier workspace snapshot.
  # Params:
  #   workspace=<id>        workspace id or slug (default: "default")
  #   audience=<tag>        audience tag for wiki variant (default: "default")
  #   bandwidth=l0|l1|full  response density (default: "l1")
  #   node=<slug>           restrict Tier 1/2 to one node (optional)
  get "/api/profile" do
    workspace = query_param(conn, "workspace", "default")
    audience = query_param(conn, "audience", "default")
    bandwidth_str = query_param(conn, "bandwidth", "l1")
    node = query_param(conn, "node", "")
    tenant = query_param(conn, "tenant", "default")

    bandwidth =
      case bandwidth_str do
        "l0" -> :l0
        "full" -> :full
        _ -> :l1
      end

    opts =
      [audience: audience, bandwidth: bandwidth, tenant_id: tenant]
      |> then(fn o -> if node == "", do: o, else: Keyword.put(o, :node_filter, node) end)

    case Profile.get(workspace, opts) do
      {:ok, profile} ->
        json(conn, profile_to_map(profile))

      {:error, :not_found} ->
        send_resp(conn, 404, Jason.encode!(%{error: "workspace not found"}))

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  get "/api/health" do
    {:ok, checks} = HealthDiagnostics.run()
    json(conn, %{health: format_health(checks)})
  end

  # ── Phase 12: runtime + retrieval + wiki endpoints ──────────────────────

  # Engine liveness + readiness (Phase 10 Health module).
  get "/api/status" do
    report = Health.ready(skip: [:embedder])

    json(conn, %{
      status: Health.status(),
      ok?: report.ok?,
      checks: Map.new(report.checks, fn {k, v} -> {k, inspect(v)} end),
      degraded: report.degraded
    })
  end

  # Runtime metrics snapshot (Phase 10 Telemetry module).
  get "/api/metrics" do
    # `try/rescue` doesn't catch GenServer exits (they aren't exceptions).
    # `try/catch :exit` does. Without this, a downed Telemetry GenServer
    # would crash this request handler instead of returning the fallback.
    snap =
      try do
        Telemetry.snapshot()
      rescue
        _ -> %{counters: %{}, histograms: %{}, uptime_ms: 0}
      catch
        :exit, _ -> %{counters: %{}, histograms: %{}, uptime_ms: 0}
      end

    json(conn, snap)
  end

  # POST /api/rag — end-to-end retrieval for LLM consumption.
  # Body: {"query": "…", "format": "markdown", "audience": "default", "bandwidth": "medium", "workspace": "default"}
  post "/api/rag" do
    body = conn.body_params || %{}

    case Map.get(body, "query") do
      q when is_binary(q) and q != "" ->
        receiver =
          Receiver.new(%{
            format: parse_atom(body["format"], :markdown),
            bandwidth: parse_atom(body["bandwidth"], :medium),
            audience: body["audience"] || "default"
          })

        workspace_id = body["workspace"] || "default"

        {:ok, result} = Retrieval.ask(q, receiver: receiver, workspace_id: workspace_id)
        json(conn, result)

      _ ->
        send_resp(conn, 400, Jason.encode!(%{error: "query is required"}))
    end
  end

  # GET /api/rag/stream — SSE streaming variant of POST /api/rag.
  # Emits pipeline-stage events as the retrieval composes the envelope:
  #   intent → wiki_hit (if wiki) → chunks (if hybrid) → composing → envelope → done
  #
  # Params:
  #   query=<text>       required — the question
  #   workspace=<id>     workspace id (default: "default")
  #   audience=<tag>     wiki audience (default: "default")
  #   format=<atom>      response format: markdown | plain | claude | openai (default: markdown)
  #   bandwidth=<atom>   small | medium | large (default: medium)
  get "/api/rag/stream" do
    query = query_param(conn, "query", "")

    if query == "" do
      send_resp(conn, 400, Jason.encode!(%{error: "query is required"}))
    else
      receiver =
        Receiver.new(%{
          format: parse_atom(query_param(conn, "format", "markdown"), :markdown),
          bandwidth: parse_atom(query_param(conn, "bandwidth", "medium"), :medium),
          audience: query_param(conn, "audience", "default")
        })

      workspace_id = query_param(conn, "workspace", "default")

      # skip_intent is a test-only shortcut that bypasses Ollama analysis.
      # It is ignored in production (non-test) builds.
      rag_opts =
        [workspace_id: workspace_id]
        |> then(fn opts ->
          if Mix.env() == :test and query_param(conn, "skip_intent", "") == "true" do
            Keyword.put(opts, :skip_intent, true)
          else
            opts
          end
        end)

      conn = stream_init(conn)
      {:ok, _task} = RagStream.start_link(query, receiver, self(), rag_opts)

      rag_stream_loop(conn)
    end
  end

  # Wiki listing — GET /api/wiki?tenant=default&workspace=default
  get "/api/wiki" do
    tenant = query_param(conn, "tenant", "default")
    workspace = query_param(conn, "workspace", "default")

    case Wiki.list(tenant, workspace) do
      {:ok, pages} ->
        json(conn, %{
          tenant_id: tenant,
          workspace_id: workspace,
          pages:
            Enum.map(pages, fn p ->
              %{
                slug: p.slug,
                audience: p.audience,
                version: p.version,
                last_curated: p.last_curated,
                curated_by: p.curated_by,
                size_bytes: byte_size(p.body || ""),
                workspace_id: p.workspace_id
              }
            end)
        })

      _ ->
        json(conn, %{pages: []})
    end
  end

  # GET /api/wiki/contradictions?workspace=X
  # Active contradiction surfacing events from the last 90 days. MUST sit
  # before the /:slug route below — Plug routers match first-defined-wins,
  # and "contradictions" would otherwise be matched as a slug.
  get "/api/wiki/contradictions" do
    workspace = query_param(conn, "workspace", "default")

    sql = """
    SELECT envelope_slug, metadata, score, pushed_at
    FROM surfacing_events
    WHERE workspace_id = ?1
      AND category = 'contradictions'
      AND datetime(pushed_at) >= datetime('now', '-90 days')
    ORDER BY pushed_at DESC
    LIMIT 200
    """

    rows =
      case Store.raw_query(sql, [workspace]) do
        {:ok, r} -> r
        _ -> []
      end

    items =
      Enum.map(rows, fn [slug, metadata_json, score, pushed_at] ->
        meta =
          case Jason.decode(metadata_json || "{}") do
            {:ok, m} -> m
            _ -> %{}
          end

        %{
          page_slug: slug,
          contradictions: Map.get(meta, "contradictions", []),
          entities: Map.get(meta, "entities", []),
          score: score,
          detected_at: pushed_at
        }
      end)

    json(conn, %{workspace_id: workspace, contradictions: items, count: length(items)})
  end

  # Wiki page render — GET /api/wiki/:slug?workspace=default&audience=default&format=markdown
  get "/api/wiki/:slug" do
    tenant = query_param(conn, "tenant", "default")
    workspace = query_param(conn, "workspace", "default")
    audience = query_param(conn, "audience", "default")
    format = query_param(conn, "format", "markdown") |> parse_atom(:markdown)

    case Wiki.latest(tenant, slug, audience, workspace) do
      {:ok, page} ->
        resolver = fn _d, _opts -> {:ok, "", %{}} end
        {rendered, warnings} = Wiki.render(page, resolver, format: format)

        json(conn, %{
          slug: page.slug,
          audience: page.audience,
          version: page.version,
          workspace_id: page.workspace_id,
          body: rendered,
          warnings: warnings
        })

      _ ->
        send_resp(conn, 404, Jason.encode!(%{error: "page not found"}))
    end
  end

  # ── Phase 12.5: endpoints the ported BusinessOS knowledge components call ─
  #
  # The Svelte components in desktop/src/lib/components/knowledge/ (copied from
  # BusinessOS-5) hit these paths directly. We keep the shape stable so the
  # component code stays unmodified between the two projects.

  # GET /api/optimal/graph — shape expected by OptimalGraphView.svelte:
  #   %{entities: [%{name, type, connections}], edges: [%{source, target, relation, weight}], stats: %{...}}
  get "/api/optimal/graph" do
    entities = optimal_entity_summary()
    edges = optimal_relation_edges()
    edge_types = Enum.frequencies_by(edges, & &1.relation)

    json(conn, %{
      entities: entities,
      edges: edges,
      stats: %{
        entity_count: length(entities),
        edge_count: length(edges),
        edge_types: edge_types
      }
    })
  end

  # GET /api/optimal/nodes — shape expected by NodeDrillDown level-0 card grid:
  #   %{nodes: [%{slug, name, type, signal_count}]}
  get "/api/optimal/nodes" do
    json(conn, %{nodes: optimal_node_summary()})
  end

  # GET /api/optimal/nodes/:slug/files — drill-down file tree. Component
  # expects `{files: [...]}`; each entry `%{name, path, is_dir, size, children?}`.
  get "/api/optimal/nodes/:slug/files" do
    json(conn, %{files: optimal_node_files(slug)})
  end

  # ── Phase 14: workspace explorer endpoints ──────────────────────────────

  # GET /api/workspace — full node forest with parent/child + signal counts.
  # Returned as a flat list; the UI builds the tree client-side using parent_id.
  # Accepts optional ?workspace=<id> to scope to a specific workspace (default: "default").
  get "/api/workspace" do
    workspace = query_param(conn, "workspace", "default")
    json(conn, %{nodes: workspace_nodes(workspace)})
  end

  # GET /api/signals/:id — full signal granularity: chunks (4 scales),
  # entities (by type), classification, intent, clusters, wiki citations.
  # This is the "see everything the engine knows about this one data point"
  # endpoint that the workspace drill-down mounts.
  get "/api/signals/:id" do
    case signal_detail(id) do
      nil -> send_resp(conn, 404, Jason.encode!(%{error: "signal not found"}))
      detail -> json(conn, detail)
    end
  end

  # GET /api/activity — reverse-chron events (audit log). Params:
  #   limit=N  (default 100, max 1000)
  #   kind=ingest|erasure|retention_action|…  (optional filter)
  get "/api/activity" do
    limit = conn |> query_param("limit", "100") |> parse_int(100) |> min(1000)
    kind = query_param(conn, "kind", "")
    json(conn, %{events: recent_events(limit, kind)})
  end

  # GET /api/architectures — every data architecture the engine recognises,
  # built-ins first then any tenant-registered schemas.
  get "/api/architectures" do
    archs =
      Enum.map(OptimalEngine.Architecture.list(), fn a ->
        %{
          id: a.id,
          name: a.name,
          version: a.version,
          description: a.description,
          modality_primary: a.modality_primary,
          granularity: a.granularity,
          field_count: length(a.fields)
        }
      end)

    processors =
      Enum.map(OptimalEngine.Architecture.processor_summary(), fn {id, modality, emits} ->
        %{id: id, modality: modality, emits: emits}
      end)

    json(conn, %{architectures: archs, processors: processors})
  end

  # GET /api/architectures/:id — field-level detail for one architecture.
  get "/api/architectures/:id" do
    case OptimalEngine.Architecture.fetch(id) do
      {:ok, arch} ->
        json(conn, architecture_detail(arch))

      {:error, _} ->
        send_resp(conn, 404, Jason.encode!(%{error: "architecture not found"}))
    end
  end

  # ── Organizations + Workspaces (Phase 1.5) ─────────────────────────────

  # GET /api/organizations — every organization (tenant) the caller can see.
  # v0.1 single-tenant: returns the singleton default org.
  get "/api/organizations" do
    case OptimalEngine.Tenancy.Tenant.get(OptimalEngine.Tenancy.Tenant.default_id()) do
      {:ok, t} ->
        json(conn, %{
          organizations: [
            %{
              id: t.id,
              name: t.name,
              plan: t.plan,
              region: t.region,
              created_at: t.created_at
            }
          ]
        })

      _ ->
        json(conn, %{organizations: []})
    end
  end

  # GET /api/workspaces?tenant=default — list workspaces in an org.
  # Status filter via ?status=active|archived|all (default: active).
  get "/api/workspaces" do
    tenant_id = query_param(conn, "tenant", OptimalEngine.Tenancy.Tenant.default_id())

    status =
      conn
      |> query_param("status", "active")
      |> case do
        "all" -> :all
        "archived" -> :archived
        _ -> :active
      end

    case OptimalEngine.Workspace.list(tenant_id: tenant_id, status: status) do
      {:ok, list} ->
        json(conn, %{
          tenant_id: tenant_id,
          workspaces: Enum.map(list, &workspace_to_map/1)
        })

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/workspaces — create a workspace under a tenant.
  # Body: {"slug":"engineering","name":"Engineering Brain","description":"…","tenant":"default"}
  post "/api/workspaces" do
    body = conn.body_params || %{}
    slug = Map.get(body, "slug")
    name = Map.get(body, "name")

    cond do
      not is_binary(slug) or slug == "" ->
        send_resp(conn, 400, Jason.encode!(%{error: "slug required"}))

      not is_binary(name) or name == "" ->
        send_resp(conn, 400, Jason.encode!(%{error: "name required"}))

      true ->
        tenant_id = Map.get(body, "tenant", OptimalEngine.Tenancy.Tenant.default_id())
        description = Map.get(body, "description")

        case OptimalEngine.Workspace.create(%{
               slug: slug,
               name: name,
               description: description,
               tenant_id: tenant_id
             }) do
          {:ok, ws} ->
            conn
            |> put_resp_content_type("application/json")
            |> send_resp(201, Jason.encode!(workspace_to_map(ws)))

          {:error, reason} ->
            send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
        end
    end
  end

  # GET /api/workspaces/:id — fetch one.
  get "/api/workspaces/:id" do
    case OptimalEngine.Workspace.get(id) do
      {:ok, ws} -> json(conn, workspace_to_map(ws))
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "workspace not found"}))
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # PATCH /api/workspaces/:id — rename / re-describe.
  patch "/api/workspaces/:id" do
    body = conn.body_params || %{}

    attrs =
      %{}
      |> maybe_put(:name, Map.get(body, "name"))
      |> maybe_put(:description, Map.get(body, "description"))

    case OptimalEngine.Workspace.update(id, attrs) do
      {:ok, ws} -> json(conn, workspace_to_map(ws))
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "workspace not found"}))
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/workspaces/:id/archive — soft delete.
  post "/api/workspaces/:id/archive" do
    case OptimalEngine.Workspace.archive(id) do
      :ok -> send_resp(conn, 204, "")
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # GET /api/workspaces/:id/config — merged (defaults + on-disk) config.
  # :id is the workspace id (e.g. "default" or "default:engineering").
  # Resolved to slug via Workspace.get/1 before reading the filesystem.
  get "/api/workspaces/:id/config" do
    with {:ok, ws} <- resolve_workspace(id),
         {:ok, cfg} <- OptimalEngine.Workspace.Config.get(ws.slug) do
      json(conn, %{workspace_id: id, config: stringify_keys(cfg)})
    else
      {:error, :not_found} ->
        send_resp(conn, 404, Jason.encode!(%{error: "workspace not found"}))

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # PATCH /api/workspaces/:id/config — deep-merge request body into the
  # on-disk YAML and return the full merged config. Body keys are top-level
  # section names (string keys); nested values replace on a key-by-key basis.
  patch "/api/workspaces/:id/config" do
    body = conn.body_params || %{}

    with {:ok, ws} <- resolve_workspace(id),
         :ok <- OptimalEngine.Workspace.Config.put(ws.slug, body),
         {:ok, cfg} <- OptimalEngine.Workspace.Config.get(ws.slug) do
      json(conn, %{workspace_id: id, config: stringify_keys(cfg)})
    else
      {:error, :not_found} ->
        send_resp(conn, 404, Jason.encode!(%{error: "workspace not found"}))

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # ── Cued recall (Engramme-flavored memory-failure flows) ────────────────
  #
  # Five typed endpoints, one per memory-failure pattern from Kreiman's
  # 2026 "Questions in the Wild" study. Each builds a shaped query string,
  # passes it to the same Retrieval.ask flow as /api/rag, and returns the
  # same envelope. Difference vs /api/rag: the question shape is unambiguous,
  # so IntentAnalyzer decodes intent with maximum confidence and the
  # retrieval boost picks chunks that match the intent type directly.

  # GET /api/recall/actions?actor=X&topic=Y&since=Z&workspace=default
  # Past actions / decisions / commits.
  get "/api/recall/actions" do
    actor = query_param(conn, "actor", "")
    topic = query_param(conn, "topic", "")
    since = query_param(conn, "since", "")
    workspace = query_param(conn, "workspace", "default")

    q =
      ["what actions, decisions, or commitments"]
      |> append_if(actor != "", "by #{actor}")
      |> append_if(topic != "", "about #{topic}")
      |> append_if(since != "", "since #{since}")
      |> Enum.join(" ")

    do_recall(conn, q, body_audience(conn), workspace)
  end

  # GET /api/recall/who?topic=X&role=Y&workspace=default
  # Contact / ownership lookup — "who owns this".
  get "/api/recall/who" do
    topic = query_param(conn, "topic", "")
    role = query_param(conn, "role", "owner")
    workspace = query_param(conn, "workspace", "default")
    q = "who is the #{role} of #{topic}, and how do I reach them"
    do_recall(conn, q, body_audience(conn), workspace)
  end

  # GET /api/recall/when?event=X&workspace=default
  # Schedule / temporal lookup.
  get "/api/recall/when" do
    event = query_param(conn, "event", "")
    workspace = query_param(conn, "workspace", "default")
    q = "when does #{event} happen, and what is the current schedule"
    do_recall(conn, q, body_audience(conn), workspace)
  end

  # GET /api/recall/where?thing=X&workspace=default
  # Object-location lookup — which node/file/path.
  get "/api/recall/where" do
    thing = query_param(conn, "thing", "")
    workspace = query_param(conn, "workspace", "default")
    q = "where is #{thing} kept, owned, or discussed in the workspace"
    do_recall(conn, q, body_audience(conn), workspace)
  end

  # GET /api/recall/owns?actor=X&workspace=default
  # Open-task / current-commitment lookup for an actor.
  get "/api/recall/owns" do
    actor = query_param(conn, "actor", "")
    workspace = query_param(conn, "workspace", "default")
    q = "what is #{actor} currently committed to, and what tasks are open"
    do_recall(conn, q, body_audience(conn), workspace)
  end

  # ── Subscriptions + proactive surfacing (Phase 15) ──────────────────────

  # GET /api/subscriptions?workspace=default
  get "/api/subscriptions" do
    workspace_id = query_param(conn, "workspace", OptimalEngine.Workspace.default_id())

    case OptimalEngine.Memory.Subscription.list(workspace_id: workspace_id) do
      {:ok, subs} ->
        json(conn, %{
          workspace_id: workspace_id,
          subscriptions: Enum.map(subs, &subscription_to_map/1)
        })

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/subscriptions
  # Body: {"workspace": "default", "scope": "topic", "scope_value": "pricing",
  #        "categories": ["recent_actions","ownership"], "principal_id": "alice"}
  post "/api/subscriptions" do
    body = conn.body_params || %{}

    attrs =
      %{
        workspace_id: Map.get(body, "workspace", OptimalEngine.Workspace.default_id()),
        principal_id: Map.get(body, "principal_id"),
        scope: parse_atom(Map.get(body, "scope", "workspace"), :workspace),
        scope_value: Map.get(body, "scope_value"),
        activity: Map.get(body, "activity")
      }
      |> maybe_put(:categories, parse_categories(Map.get(body, "categories")))

    case OptimalEngine.Memory.Subscription.create(attrs) do
      {:ok, sub} ->
        conn
        |> put_resp_content_type("application/json")
        |> send_resp(201, Jason.encode!(subscription_to_map(sub)))

      {:error, reason} ->
        send_resp(conn, 400, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/subscriptions/:id/pause
  post "/api/subscriptions/:id/pause" do
    case OptimalEngine.Memory.Subscription.pause(id) do
      :ok -> send_resp(conn, 204, "")
      other -> send_resp(conn, 500, Jason.encode!(%{error: inspect(other)}))
    end
  end

  # POST /api/subscriptions/:id/resume
  post "/api/subscriptions/:id/resume" do
    case OptimalEngine.Memory.Subscription.resume(id) do
      :ok -> send_resp(conn, 204, "")
      other -> send_resp(conn, 500, Jason.encode!(%{error: inspect(other)}))
    end
  end

  # DELETE /api/subscriptions/:id
  delete "/api/subscriptions/:id" do
    case OptimalEngine.Memory.Subscription.delete(id) do
      :ok -> send_resp(conn, 204, "")
      other -> send_resp(conn, 500, Jason.encode!(%{error: inspect(other)}))
    end
  end

  # POST /api/surface/test  body: {"subscription":"sub:...","slug":"healthtech-pricing-decision"}
  # Convenience: trigger a synthetic push to all listeners of a subscription.
  post "/api/surface/test" do
    body = conn.body_params || %{}
    sub_id = Map.get(body, "subscription")
    slug = Map.get(body, "slug")

    cond do
      not is_binary(sub_id) ->
        send_resp(conn, 400, Jason.encode!(%{error: "subscription required"}))

      not is_binary(slug) ->
        send_resp(conn, 400, Jason.encode!(%{error: "slug required"}))

      true ->
        OptimalEngine.Memory.Surfacer.test_push(sub_id, slug)
        send_resp(conn, 204, "")
    end
  end

  # GET /api/surface/stream?subscription=<id>
  # Server-Sent Events stream. Client sends EventSource and receives
  # newline-delimited JSON envelopes whenever the Surfacer pushes.
  get "/api/surface/stream" do
    sub_id = query_param(conn, "subscription", "")

    if sub_id == "" do
      send_resp(conn, 400, Jason.encode!(%{error: "subscription query param required"}))
    else
      conn = stream_init(conn)
      OptimalEngine.Memory.Surfacer.subscribe(sub_id, self())
      send_sse(conn, "ready", %{subscription: sub_id})
      stream_loop(conn, sub_id)
    end
  end

  # ── Memory primitive (Phase 17) ─────────────────────────────────────────

  # POST /api/memory — create a new memory entry.
  # Body: {content, workspace?, is_static?, audience?, citation_uri?, source_chunk_id?, metadata?}
  # Returns: 201 + memory struct.
  post "/api/memory" do
    body = conn.body_params || %{}
    content = Map.get(body, "content")

    if not (is_binary(content) and content != "") do
      send_resp(conn, 400, Jason.encode!(%{error: "content is required"}))
    else
      workspace_id = Map.get(body, "workspace", "default")

      attrs =
        %{content: content, workspace_id: workspace_id}
        |> maybe_put(:is_static, Map.get(body, "is_static"))
        |> maybe_put(:audience, Map.get(body, "audience"))
        |> maybe_put(:citation_uri, Map.get(body, "citation_uri"))
        |> maybe_put(:source_chunk_id, Map.get(body, "source_chunk_id"))
        |> maybe_put(:metadata, Map.get(body, "metadata"))

      case OptimalEngine.Memory.create(attrs) do
        {:ok, mem} ->
          conn
          |> put_resp_content_type("application/json")
          |> send_resp(201, Jason.encode!(memory_to_map(mem)))

        {:error, reason} ->
          send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
      end
    end
  end

  # GET /api/memory/:id/versions — version chain for a memory (BEFORE bare /:id).
  # Returns: {memory_id, root_id, versions: [...]} in chronological order.
  get "/api/memory/:id/versions" do
    case OptimalEngine.Memory.versions(id) do
      {:ok, vs} ->
        root_id = vs |> List.first() |> then(fn v -> v && v.root_memory_id end)

        json(conn, %{
          memory_id: id,
          root_id: root_id,
          versions: Enum.map(vs, &memory_to_map/1)
        })

      {:error, :not_found} ->
        send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # GET /api/memory/:id/relations — inbound + outbound relation graph (BEFORE bare /:id).
  # Returns: {memory_id, inbound: [...], outbound: [...]}
  get "/api/memory/:id/relations" do
    case OptimalEngine.Memory.relations(id) do
      {:ok, rels} ->
        {inbound, outbound} =
          Enum.split_with(rels, fn r -> to_string(r.target_id) == id end)

        json(conn, %{memory_id: id, inbound: inbound, outbound: outbound})

      {:error, :not_found} ->
        send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # GET /api/memory/:id — fetch one memory by id.
  get "/api/memory/:id" do
    case OptimalEngine.Memory.get(id) do
      {:ok, mem} -> json(conn, memory_to_map(mem))
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # GET /api/memory — list memories for a workspace.
  # Params: workspace, audience, include_forgotten (bool), include_old_versions (bool), limit (int)
  get "/api/memory" do
    workspace_id = query_param(conn, "workspace", "default")
    audience = query_param(conn, "audience", "")
    include_forgotten = conn |> query_param("include_forgotten", "false") |> parse_bool(false)
    include_old_versions = conn |> query_param("include_old_versions", "false") |> parse_bool(false)
    limit = conn |> query_param("limit", "50") |> parse_int(50)

    opts =
      [
        workspace_id: workspace_id,
        limit: limit,
        include_forgotten: include_forgotten,
        include_old_versions: include_old_versions
      ]
      |> api_maybe_put(:audience, if(audience == "", do: nil, else: audience))

    memories = OptimalEngine.Memory.list(opts)

    json(conn, %{
      workspace_id: workspace_id,
      count: length(memories),
      memories: Enum.map(memories, &memory_to_map/1)
    })
  end

  # POST /api/memory/:id/forget — mark a memory as forgotten.
  # Body: {reason?, forget_after?}  Returns: 204.
  post "/api/memory/:id/forget" do
    body = conn.body_params || %{}

    opts =
      []
      |> api_maybe_put(:reason, Map.get(body, "reason"))
      |> api_maybe_put(:forget_after, Map.get(body, "forget_after"))

    case OptimalEngine.Memory.forget(id, opts) do
      {:ok, _} -> send_resp(conn, 204, "")
      :ok -> send_resp(conn, 204, "")
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/memory/:id/update — create a new version of a memory.
  # Body: {content, audience?, citation_uri?, metadata?}  Returns: 201 + new memory struct.
  post "/api/memory/:id/update" do
    conn |> memory_mutation_endpoint(id, &OptimalEngine.Memory.update/2)
  end

  # POST /api/memory/:id/extend — create a child memory with :extends relation.
  # Body: {content, audience?, citation_uri?, metadata?}  Returns: 201 + new memory struct.
  post "/api/memory/:id/extend" do
    conn |> memory_mutation_endpoint(id, &OptimalEngine.Memory.extend/2)
  end

  # POST /api/memory/:id/derive — create a derived memory with :derives relation.
  # Body: {content, audience?, citation_uri?, metadata?}  Returns: 201 + new memory struct.
  post "/api/memory/:id/derive" do
    conn |> memory_mutation_endpoint(id, &OptimalEngine.Memory.derive/2)
  end

  # DELETE /api/memory/:id — hard delete. Returns: 204.
  delete "/api/memory/:id" do
    case OptimalEngine.Memory.delete(id) do
      :ok -> send_resp(conn, 204, "")
      {:ok, _} -> send_resp(conn, 204, "")
      {:error, :not_found} -> send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end


  # POST /api/memory/:id/promote — manually promote a memory to a wiki page section.
  # Body: {"slug": "my-page", "audience"?: "default", "workspace"?: "default"}
  # Returns: 200 + updated page, 404 if memory not found.
  post "/api/memory/:id/promote" do
    body = conn.body_params || %{}
    slug = Map.get(body, "slug")

    if not (is_binary(slug) and slug != "") do
      send_resp(conn, 400, Jason.encode!(%{error: "slug is required"}))
    else
      workspace_id = Map.get(body, "workspace", "default")
      audience = Map.get(body, "audience", "default")
      tenant_id = Map.get(body, "tenant", "default")

      opts = [workspace_id: workspace_id, tenant_id: tenant_id, audience: audience]

      case OptimalEngine.Memory.WikiBridge.promote_memory_to_wiki(id, slug, opts) do
        {:ok, page} ->
          json(conn, %{
            slug: page.slug,
            audience: page.audience,
            version: page.version,
            workspace_id: page.workspace_id,
            body: page.body,
            last_curated: page.last_curated,
            curated_by: page.curated_by
          })

        {:error, :not_found} ->
          send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))

        {:error, reason} ->
          send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
      end
    end
  end

  # ── API key management (Phase 18) ──────────────────────────────────────────
  #
  # These endpoints manage API keys for the current tenant. When auth is on
  # (auth_required: true) they require the `admin` scope. In dev (auth_required:
  # false) they are open — same behaviour as every other route.

  # POST /api/auth/keys — mint a new key. Returns id, key, prefix.
  # Body: {name, scopes?, workspace_scope?, expires_at?}
  post "/api/auth/keys" do
    body = conn.body_params || %{}
    name = Map.get(body, "name")
    tenant_id = conn.assigns[:current_tenant] || "default"

    if not (is_binary(name) and name != "") do
      send_resp(conn, 400, Jason.encode!(%{error: "name is required"}))
    else
      attrs =
        %{tenant_id: tenant_id, name: name}
        |> maybe_put(:scopes, Map.get(body, "scopes"))
        |> maybe_put(:workspace_scope, Map.get(body, "workspace_scope"))
        |> maybe_put(:expires_at, Map.get(body, "expires_at"))
        |> maybe_put(:principal_id, Map.get(body, "principal_id"))

      case ApiKey.mint(attrs) do
        {:ok, %{id: id, key: key, secret: _secret}} ->
          conn
          |> put_resp_content_type("application/json")
          |> send_resp(201, Jason.encode!(%{id: id, key: key, prefix: String.slice(key, 0, 8)}))

        {:error, reason} ->
          send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
      end
    end
  end

  # GET /api/auth/keys — list non-revoked keys for current tenant (no secrets).
  get "/api/auth/keys" do
    tenant_id = conn.assigns[:current_tenant] || "default"

    case ApiKey.list(tenant_id) do
      {:ok, keys} ->
        json(conn, %{
          tenant_id: tenant_id,
          keys: Enum.map(keys, &api_key_to_map/1)
        })

      {:error, reason} ->
        send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # POST /api/auth/keys/:id/revoke — soft revoke a key.
  post "/api/auth/keys/:id/revoke" do
    case ApiKey.revoke(id) do
      :ok -> send_resp(conn, 204, "")
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  # DELETE /api/auth/keys/:id — hard delete a key.
  delete "/api/auth/keys/:id" do
    case ApiKey.delete(id) do
      :ok -> send_resp(conn, 204, "")
      {:error, reason} -> send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
    end
  end

  match _ do
    send_resp(conn, 404, Jason.encode!(%{error: "not found"}))
  end

  # ---------------------------------------------------------------------------
  # Helpers: response
  # ---------------------------------------------------------------------------

  defp json(conn, data) do
    body = Jason.encode!(data)

    conn
    |> put_resp_content_type("application/json")
    |> send_resp(200, body)
  end

  defp cors(conn, _opts) do
    conn
    |> put_resp_header("access-control-allow-origin", "*")
    |> put_resp_header("access-control-allow-methods", "GET, POST, PATCH, DELETE, OPTIONS")
    |> put_resp_header("access-control-allow-headers", "content-type")
  end

  defp workspace_to_map(ws) do
    %{
      id: ws.id,
      tenant_id: ws.tenant_id,
      slug: ws.slug,
      name: ws.name,
      description: ws.description,
      status: Atom.to_string(ws.status),
      created_at: ws.created_at,
      archived_at: ws.archived_at,
      metadata: ws.metadata
    }
  end

  defp profile_to_map(%Profile{} = p) do
    %{
      workspace_id: p.workspace_id,
      tenant_id: p.tenant_id,
      audience: p.audience,
      static: p.static,
      dynamic: p.dynamic,
      curated: p.curated,
      activity: p.activity,
      entities: p.entities,
      generated_at: p.generated_at
    }
  end

  defp memory_to_map(mem) do
    %{
      id: mem.id,
      tenant_id: mem.tenant_id,
      workspace_id: mem.workspace_id,
      content: mem.content,
      is_static: mem.is_static,
      is_forgotten: mem.is_forgotten,
      forget_after: mem.forget_after,
      forget_reason: mem.forget_reason,
      version: mem.version,
      parent_memory_id: mem.parent_memory_id,
      root_memory_id: mem.root_memory_id,
      is_latest: mem.is_latest,
      citation_uri: mem.citation_uri,
      source_chunk_id: mem.source_chunk_id,
      audience: mem.audience,
      metadata: mem.metadata,
      created_at: mem.created_at,
      updated_at: mem.updated_at,
      was_existing: Map.get(mem, :was_existing, false)
    }
  end

  # Shared handler for update/extend/derive — each takes (id, attrs) and returns {:ok, mem}.
  defp memory_mutation_endpoint(conn, id, fun) do
    body = conn.body_params || %{}
    content = Map.get(body, "content")

    if not (is_binary(content) and content != "") do
      send_resp(conn, 400, Jason.encode!(%{error: "content is required"}))
    else
      attrs =
        %{content: content}
        |> maybe_put(:audience, Map.get(body, "audience"))
        |> maybe_put(:citation_uri, Map.get(body, "citation_uri"))
        |> maybe_put(:metadata, Map.get(body, "metadata"))

      case fun.(id, attrs) do
        {:ok, mem} ->
          conn
          |> put_resp_content_type("application/json")
          |> send_resp(201, Jason.encode!(memory_to_map(mem)))

        {:error, :not_found} ->
          send_resp(conn, 404, Jason.encode!(%{error: "memory not found"}))

        {:error, reason} ->
          send_resp(conn, 500, Jason.encode!(%{error: inspect(reason)}))
      end
    end
  end

  defp api_key_to_map(key) do
    %{
      id: key.id,
      tenant_id: key.tenant_id,
      principal_id: key.principal_id,
      prefix: key.prefix,
      name: key.name,
      scopes: key.scopes,
      workspace_scope: key.workspace_scope,
      expires_at: key.expires_at,
      created_at: key.created_at,
      last_used_at: key.last_used_at,
      revoked_at: key.revoked_at
    }
  end

  defp maybe_put(map, _key, nil), do: map
  defp maybe_put(map, key, val), do: Map.put(map, key, val)

  # Keyword-list variant for the /api/grep opts builder.
  defp api_maybe_put(kw, _key, nil), do: kw
  defp api_maybe_put(kw, key, val), do: Keyword.put(kw, key, val)

  defp parse_bool("true", _default), do: true
  defp parse_bool("1", _default), do: true
  defp parse_bool(_, default), do: default

  # Resolve a workspace id (e.g. "default" or "default:engineering") to a
  # Workspace struct. Returns `{:error, :not_found}` for unknown ids.
  defp resolve_workspace(id) do
    OptimalEngine.Workspace.get(id)
  end

  # Recursively convert atom keys to strings for JSON serialisation.
  # Config maps use atom keys internally; JSON must emit strings.
  defp stringify_keys(map) when is_map(map) do
    Map.new(map, fn
      {k, v} when is_atom(k) -> {Atom.to_string(k), stringify_keys(v)}
      {k, v} -> {k, stringify_keys(v)}
    end)
  end

  defp stringify_keys(list) when is_list(list), do: Enum.map(list, &stringify_keys/1)
  defp stringify_keys(other), do: other

  defp append_if(list, true, item), do: list ++ [item]
  defp append_if(list, false, _item), do: list
  defp append_if(list, "", _item), do: list

  defp body_audience(conn), do: query_param(conn, "audience", "default")

  defp do_recall(conn, q, audience, workspace) do
    if String.length(String.trim(q)) < 8 do
      send_resp(
        conn,
        400,
        Jason.encode!(%{error: "missing required parameter(s) — recall query was too short"})
      )
    else
      receiver = Receiver.new(%{format: :markdown, bandwidth: :medium, audience: audience})
      {:ok, result} = Retrieval.ask(q, receiver: receiver, workspace_id: workspace)
      json(conn, Map.put(result, :recall_query, q))
    end
  end

  # ── SSE plumbing for /api/surface/stream ────────────────────────────────

  defp subscription_to_map(s) do
    %{
      id: s.id,
      tenant_id: s.tenant_id,
      workspace_id: s.workspace_id,
      principal_id: s.principal_id,
      scope: Atom.to_string(s.scope),
      scope_value: s.scope_value,
      categories: Enum.map(s.categories, &Atom.to_string/1),
      activity: s.activity,
      status: Atom.to_string(s.status),
      created_at: s.created_at
    }
  end

  defp parse_categories(nil), do: nil

  defp parse_categories(list) when is_list(list) do
    Enum.map(list, fn c ->
      try do
        String.to_existing_atom(c)
      rescue
        _ -> :unassigned
      end
    end)
  end

  defp parse_categories(_), do: nil

  defp stream_init(conn) do
    conn
    |> put_resp_header("content-type", "text/event-stream")
    |> put_resp_header("cache-control", "no-cache")
    |> put_resp_header("connection", "keep-alive")
    |> send_chunked(200)
  end

  defp send_sse(conn, event, data) do
    payload = Jason.encode!(data)
    chunk = "event: #{event}\ndata: #{payload}\n\n"

    case Plug.Conn.chunk(conn, chunk) do
      {:ok, conn} -> conn
      {:error, _} -> conn
    end
  end

  # Long-running receive loop. Listens for surface pushes, heartbeats every
  # 25s to keep proxies from cutting the connection, exits when the client
  # disconnects (chunk write fails).
  defp stream_loop(conn, sub_id) do
    receive do
      {:surface, payload} ->
        case Plug.Conn.chunk(conn, "event: surface\ndata: #{Jason.encode!(payload)}\n\n") do
          {:ok, conn} -> stream_loop(conn, sub_id)
          {:error, _} -> stream_close(conn, sub_id)
        end

      {:plug_conn, :sent} ->
        stream_loop(conn, sub_id)

      _other ->
        stream_loop(conn, sub_id)
    after
      25_000 ->
        case Plug.Conn.chunk(conn, ": keepalive\n\n") do
          {:ok, conn} -> stream_loop(conn, sub_id)
          {:error, _} -> stream_close(conn, sub_id)
        end
    end
  end

  # Receive loop for /api/rag/stream — drains RagStream messages and forwards
  # each as an SSE event. Keepalive every 25s; hard timeout after 30s of silence
  # from the pipeline. Returns the final conn when done or on error.
  defp rag_stream_loop(conn) do
    receive do
      {:rag_stream_event, event, payload} ->
        conn = send_sse(conn, to_string(event), payload)
        rag_stream_loop(conn)

      {:rag_stream_done, final} ->
        conn = send_sse(conn, "envelope", final.envelope)
        send_sse(conn, "done", %{
          elapsed_ms: final.trace.elapsed_ms,
          source: final.source
        })

      {:rag_stream_error, reason} ->
        send_sse(conn, "error", %{error: inspect(reason)})

      # Task links emit a {:DOWN, ...} or normal exit message — ignore safely.
      {_ref, _result} ->
        rag_stream_loop(conn)

      {:plug_conn, :sent} ->
        rag_stream_loop(conn)

      _other ->
        rag_stream_loop(conn)
    after
      30_000 ->
        send_sse(conn, "error", %{error: "timeout"})
    end
  end

  defp stream_close(conn, sub_id) do
    OptimalEngine.Memory.Surfacer.unsubscribe(sub_id, self())
    conn
  end

  defp parse_atom(v, default) when is_atom(default) do
    case v do
      nil -> default
      "" -> default
      s when is_binary(s) -> String.to_existing_atom(s)
      a when is_atom(a) -> a
    end
  rescue
    _ -> default
  end

  defp query_param(conn, key, default) do
    conn = Plug.Conn.fetch_query_params(conn)
    Map.get(conn.query_params, key, default)
  end

  defp parse_int(str, default) do
    case Integer.parse(str) do
      {n, _} when n > 0 -> n
      _ -> default
    end
  end

  # ---------------------------------------------------------------------------
  # Helpers: data fetching (raw SQL via Store.raw_query/2)
  # ---------------------------------------------------------------------------

  # Every context in the DB — each becomes a renderable node in the graph
  defp fetch_all_contexts do
    case Store.raw_query(
           """
           SELECT id, title, node, type, genre, sn_ratio, modified_at, l0_abstract, uri
           FROM contexts
           ORDER BY node, modified_at DESC
           """,
           []
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [id, title, node, type, genre, sn, mod, abstract, uri] ->
          %{
            id: id,
            title: title,
            node: node,
            type: type,
            genre: genre,
            sn_ratio: sn,
            modified_at: mod,
            l0_abstract: abstract,
            uri: uri
          }
        end)

      _ ->
        []
    end
  end

  # Build visual edges between context IDs (not entity/node names).
  #
  # Two strategies:
  #   1. shared_entity — contexts that both mention the same entity are linked.
  #      Weight = number of shared entities, normalised to 1.0 max.
  #   2. cross_ref     — cross_ref edges in the edges table go from a context_id
  #      to a node name; we resolve the node name to actual contexts in that node.
  #
  # All edges are deduplicated: only the (min_id, max_id) pair is kept so
  # A→B and B→A are not both emitted.
  defp fetch_visual_edges do
    entity_edges = fetch_shared_entity_edges()
    cross_ref_edges = fetch_cross_ref_edges()

    (entity_edges ++ cross_ref_edges)
    |> deduplicate_edges()
  end

  defp fetch_shared_entity_edges do
    sql = """
    SELECT e1.context_id AS source, e2.context_id AS target, e1.name AS entity
    FROM entities e1
    JOIN entities e2 ON e1.name = e2.name AND e1.context_id < e2.context_id
    LIMIT 2000
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        # Group by (source, target) pair and count shared entities
        rows
        |> Enum.group_by(fn [source, target, _entity] -> {source, target} end)
        |> Enum.map(fn {{source, target}, matches} ->
          shared_count = length(matches)
          # Pull distinct entity names for metadata (first 3 for brevity)
          entity_names = matches |> Enum.map(fn [_, _, e] -> e end) |> Enum.uniq() |> Enum.take(3)
          weight = min(shared_count / 5.0, 1.0)

          %{
            source: source,
            target: target,
            relation: "shared_entity",
            weight: Float.round(weight, 2),
            entities: entity_names,
            shared_count: shared_count
          }
        end)

      _ ->
        []
    end
  end

  defp fetch_cross_ref_edges do
    # cross_ref edges: source_id is a context_id, target_id is a node name.
    # Resolve target node name → all context IDs in that node.
    sql = """
    SELECT e.source_id AS source, c.id AS target
    FROM edges e
    JOIN contexts c ON c.node = e.target_id
    WHERE e.relation = 'cross_ref'
      AND EXISTS (SELECT 1 FROM contexts WHERE id = e.source_id)
    LIMIT 500
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        rows
        |> Enum.map(fn [source, target] ->
          %{source: source, target: target, relation: "cross_ref", weight: 1.0}
        end)

      _ ->
        []
    end
  end

  # Remove duplicate pairs: keep only one direction per (source, target) pair.
  # For shared_entity edges this is already handled by the `e1.context_id < e2.context_id`
  # constraint, but cross_ref edges may produce duplicates across both strategies.
  defp deduplicate_edges(edges) do
    edges
    |> Enum.uniq_by(fn %{source: s, target: t} ->
      if s < t, do: {s, t}, else: {t, s}
    end)
  end

  defp fetch_entity_summary do
    case Store.raw_query(
           """
           SELECT name, type, COUNT(*) as count
           FROM entities
           GROUP BY name, type
           ORDER BY count DESC
           LIMIT 200
           """,
           []
         ) do
      {:ok, rows} -> rows
      _ -> []
    end
  end

  defp fetch_node_summary do
    case Store.raw_query(
           """
           SELECT node, COUNT(*) as count, MAX(modified_at) as last_modified
           FROM contexts
           GROUP BY node
           ORDER BY node
           """,
           []
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [node, count, last_mod] ->
          %{node: node, context_count: count, last_modified: last_mod}
        end)

      _ ->
        []
    end
  end

  defp fetch_node_contexts(node_id) do
    case Store.raw_query(
           """
           SELECT id, title, type, genre, sn_ratio, modified_at
           FROM contexts
           WHERE node = ?1
           ORDER BY modified_at DESC
           LIMIT 50
           """,
           [node_id]
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [id, title, type, genre, sn, mod] ->
          %{id: id, title: title, type: type, genre: genre, sn_ratio: sn, modified_at: mod}
        end)

      _ ->
        []
    end
  end

  defp fetch_node_edges(node_id) do
    # Return edges where at least one endpoint lives in this node.
    # We join via the contexts table to resolve node membership.
    case Store.raw_query(
           """
           SELECT DISTINCT e.source_id, e.target_id, e.relation, e.weight
           FROM edges e
           JOIN contexts c ON (c.id = e.source_id OR c.id = e.target_id)
           WHERE c.node = ?1
           LIMIT 200
           """,
           [node_id]
         ) do
      {:ok, rows} -> rows
      _ -> []
    end
  end

  # ---------------------------------------------------------------------------
  # Helpers: data fetching for the ported knowledge components
  # ---------------------------------------------------------------------------

  # Entities for OptimalGraphView — one row per unique (name, type), with
  # `connections` = how many distinct contexts reference that entity.
  defp optimal_entity_summary do
    case Store.raw_query(
           """
           SELECT name, type, COUNT(DISTINCT context_id) AS connections
           FROM entities
           WHERE name IS NOT NULL AND name <> ''
           GROUP BY name, type
           ORDER BY connections DESC
           LIMIT 300
           """,
           []
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [name, type, connections] ->
          %{name: name, type: type || "concept", connections: connections}
        end)

      _ ->
        []
    end
  end

  # Edges between entities by co-occurrence in the same context. Two entities
  # sharing a context produce one `related_to` edge; weight is the number of
  # shared contexts (capped at 5.0 to keep visuals stable).
  defp optimal_relation_edges do
    sql = """
    SELECT e1.name AS source, e2.name AS target, COUNT(*) AS shared
    FROM entities e1
    JOIN entities e2
      ON e1.context_id = e2.context_id
     AND e1.name < e2.name
    WHERE e1.name IS NOT NULL AND e2.name IS NOT NULL
    GROUP BY e1.name, e2.name
    HAVING shared >= 1
    ORDER BY shared DESC
    LIMIT 800
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        Enum.map(rows, fn [s, t, shared] ->
          %{
            source: s,
            target: t,
            relation: "related_to",
            weight: min(shared * 1.0, 5.0)
          }
        end)

      _ ->
        []
    end
  end

  # Node list for the drill-down level-0 card grid. Pulls from the workspace
  # `nodes` table (Phase 3.5) joined against context counts, so operators see
  # real signal volumes per node rather than just names.
  defp optimal_node_summary do
    sql = """
    SELECT n.slug, n.name, COALESCE(n.kind, 'node') AS type,
           COALESCE((SELECT COUNT(*) FROM contexts c WHERE c.node = n.slug), 0) AS signal_count
    FROM nodes n
    ORDER BY n.slug
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        Enum.map(rows, fn [slug, name, type, count] ->
          %{slug: slug, name: name || slug, type: type, signal_count: count}
        end)

      _ ->
        # Fallback for tenants that haven't populated the nodes table yet —
        # derive distinct nodes straight from contexts.
        case Store.raw_query(
               "SELECT node, COUNT(*) FROM contexts WHERE node IS NOT NULL GROUP BY node",
               []
             ) do
          {:ok, rows} ->
            Enum.map(rows, fn [slug, count] ->
              %{slug: slug, name: slug, type: "node", signal_count: count}
            end)

          _ ->
            []
        end
    end
  end

  # File tree for NodeDrillDown — one entry per signal in a given node,
  # flattened with `is_dir: false`. The component also accepts nested
  # children but the engine stores a flat list per node, so we return that.
  defp optimal_node_files(slug) do
    case Store.raw_query(
           """
           SELECT id, title, uri, genre, modified_at, LENGTH(content) AS size
           FROM contexts
           WHERE node = ?1
           ORDER BY modified_at DESC
           LIMIT 200
           """,
           [slug]
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [id, title, uri, genre, modified_at, size] ->
          %{
            name: title || id,
            path: uri || id,
            is_dir: false,
            size: size || 0,
            genre: genre,
            modified_at: modified_at
          }
        end)

      _ ->
        []
    end
  end

  # ---------------------------------------------------------------------------
  # Helpers: workspace / signal / activity / architecture (Phase 14)
  # ---------------------------------------------------------------------------

  # Flat node list with enough structure for a client-side tree build.
  # Scoped to the given workspace_id (default: "default").
  defp workspace_nodes(workspace_id) do
    sql = """
    SELECT n.id, n.slug, n.name, n.kind, n.parent_id, n.style, n.status,
           COALESCE((SELECT COUNT(*) FROM contexts c WHERE c.node = n.slug AND c.workspace_id = ?1), 0) AS signal_count
    FROM nodes n
    WHERE n.workspace_id = ?1
    ORDER BY COALESCE(n.parent_id, ''), n.slug
    """

    case Store.raw_query(sql, [workspace_id]) do
      {:ok, rows} ->
        Enum.map(rows, fn [id, slug, name, kind, parent_id, style, status, count] ->
          %{
            id: id,
            slug: slug,
            name: name || slug,
            kind: kind || "node",
            parent_id: parent_id,
            style: style || "internal",
            status: status || "active",
            signal_count: count
          }
        end)

      _ ->
        []
    end
  end

  # Full granularity for one signal — enough for a drill-down to visualize
  # every layer the engine tracks.
  defp signal_detail(id) do
    with {:ok, [row]} <- signal_row(id) do
      [
        ctx_id,
        uri,
        title,
        genre,
        mode,
        type,
        format,
        structure,
        node,
        sn_ratio,
        content,
        l0,
        l1,
        modified_at,
        architecture_id
      ] = row

      %{
        id: ctx_id,
        uri: uri,
        title: title,
        node: node,
        genre: genre,
        modified_at: modified_at,
        architecture_id: architecture_id,
        signal_dimensions: %{
          mode: mode,
          genre: genre,
          type: type,
          format: format,
          structure: structure
        },
        sn_ratio: sn_ratio,
        content: content,
        l0_abstract: l0,
        l1_overview: l1,
        chunks: signal_chunks(ctx_id),
        entities: signal_entities(ctx_id),
        classification: signal_classification(ctx_id),
        intent: signal_intent(ctx_id),
        clusters: signal_clusters(ctx_id),
        citations: signal_wiki_citations(ctx_id)
      }
    else
      _ -> nil
    end
  end

  defp signal_row(id) do
    Store.raw_query(
      """
      SELECT id, uri, title, genre, mode, signal_type, format, structure, node,
             sn_ratio, content, l0_abstract, l1_overview, modified_at,
             architecture_id
      FROM contexts
      WHERE id = ?1 LIMIT 1
      """,
      [id]
    )
  end

  defp signal_chunks(signal_id) do
    case Store.raw_query(
           """
           SELECT id, parent_id, scale, modality, length_bytes
           FROM chunks WHERE signal_id = ?1
           ORDER BY
             CASE scale
               WHEN 'document'  THEN 0
               WHEN 'section'   THEN 1
               WHEN 'paragraph' THEN 2
               WHEN 'sentence'  THEN 3
               ELSE 4
             END,
             id
           """,
           [signal_id]
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [cid, parent, scale, modality, len] ->
          %{id: cid, parent_id: parent, scale: scale, modality: modality, length_bytes: len}
        end)

      _ ->
        []
    end
  end

  defp signal_entities(signal_id) do
    case Store.raw_query(
           "SELECT name, type FROM entities WHERE context_id = ?1 ORDER BY type, name",
           [signal_id]
         ) do
      {:ok, rows} -> Enum.map(rows, fn [n, t] -> %{name: n, type: t} end)
      _ -> []
    end
  end

  defp signal_classification(signal_id) do
    case Store.raw_query(
           """
           SELECT mode, genre, signal_type, format, structure, sn_ratio, confidence
           FROM classifications WHERE chunk_id = ?1 LIMIT 1
           """,
           ["#{signal_id}:doc"]
         ) do
      {:ok, [[mode, genre, type, format, structure, sn, conf]]} ->
        %{
          mode: mode,
          genre: genre,
          type: type,
          format: format,
          structure: structure,
          sn_ratio: sn,
          confidence: conf
        }

      _ ->
        nil
    end
  end

  defp signal_intent(signal_id) do
    case Store.raw_query(
           "SELECT intent, confidence FROM intents WHERE chunk_id = ?1 LIMIT 1",
           ["#{signal_id}:doc"]
         ) do
      {:ok, [[intent, confidence]]} -> %{intent: intent, confidence: confidence}
      _ -> nil
    end
  end

  defp signal_clusters(signal_id) do
    sql = """
    SELECT c.id, c.theme, c.intent_dominant, cm.weight
    FROM cluster_members cm
    JOIN clusters c ON c.id = cm.cluster_id
    WHERE cm.chunk_id = ?1
    """

    case Store.raw_query(sql, ["#{signal_id}:doc"]) do
      {:ok, rows} ->
        Enum.map(rows, fn [cid, theme, dom, w] ->
          %{id: cid, theme: theme, intent_dominant: dom, weight: w}
        end)

      _ ->
        []
    end
  end

  defp signal_wiki_citations(signal_id) do
    case Store.raw_query(
           """
           SELECT wiki_slug, wiki_audience, claim_hash, last_verified
           FROM citations WHERE chunk_id = ?1
           """,
           ["#{signal_id}:doc"]
         ) do
      {:ok, rows} ->
        Enum.map(rows, fn [slug, audience, claim, verified] ->
          %{wiki_slug: slug, audience: audience, claim_hash: claim, last_verified: verified}
        end)

      _ ->
        []
    end
  end

  # Activity feed — append-only events table. Most-recent first.
  defp recent_events(limit, kind) do
    {where, params} =
      if kind == "" do
        {"WHERE 1=1", [limit]}
      else
        {"WHERE kind = ?2", [limit, kind]}
      end

    sql = """
    SELECT id, tenant_id, ts, principal, kind, target_uri, latency_ms, metadata
    FROM events #{where}
    ORDER BY ts DESC, id DESC
    LIMIT ?1
    """

    case Store.raw_query(sql, params) do
      {:ok, rows} ->
        Enum.map(rows, fn [id, tenant, ts, principal, kind, uri, latency, metadata] ->
          %{
            id: id,
            tenant_id: tenant,
            ts: ts,
            principal: principal,
            kind: kind,
            target_uri: uri,
            latency_ms: latency,
            metadata: decode_json_meta(metadata)
          }
        end)

      _ ->
        []
    end
  end

  defp decode_json_meta(nil), do: %{}
  defp decode_json_meta(""), do: %{}

  defp decode_json_meta(str) do
    case Jason.decode(str) do
      {:ok, m} when is_map(m) -> m
      _ -> %{}
    end
  end

  # Full architecture detail for the /api/architectures/:id endpoint.
  defp architecture_detail(arch) do
    %{
      id: arch.id,
      name: arch.name,
      version: arch.version,
      description: arch.description,
      modality_primary: arch.modality_primary,
      granularity: arch.granularity,
      retention: arch.retention,
      fields:
        Enum.map(arch.fields, fn f ->
          %{
            name: f.name,
            modality: f.modality,
            dims: f.dims,
            required: f.required,
            processor: f.processor,
            description: f.description
          }
        end)
    }
  end

  # ---------------------------------------------------------------------------
  # Helpers: formatting
  # ---------------------------------------------------------------------------

  # Raw edge rows come from raw_query as [source_id, target_id, relation, weight]
  defp format_edges(rows) when is_list(rows) do
    Enum.map(rows, fn
      [s, t, r, w] ->
        %{source: s, target: t, relation: r, weight: w}

      %{source_id: s, target_id: t, relation: r, weight: w} ->
        %{source: s, target: t, relation: r, weight: w}

      other ->
        %{raw: inspect(other)}
    end)
  end

  defp format_edges(_), do: []

  # Entity rows: [name, type, count]
  defp format_entities(rows) when is_list(rows) do
    Enum.map(rows, fn
      [name, type, count] -> %{name: name, type: type, count: count}
      other -> %{raw: inspect(other)}
    end)
  end

  defp format_entities(_), do: []

  # GraphAnalyzer.hubs/0 returns [%{id:, degree:, sigma:}]
  defp format_hubs(hubs) when is_list(hubs) do
    Enum.map(hubs, fn
      %{id: id, degree: degree, sigma: sigma} ->
        %{entity: id, degree: degree, sigma: Float.round(sigma, 2)}

      other ->
        %{raw: inspect(other)}
    end)
  end

  defp format_hubs(_), do: []

  # GraphAnalyzer.triangles/1 returns [%{a:, b:, c:, suggestion:}]
  defp format_triangles(triangles) when is_list(triangles) do
    Enum.map(triangles, fn
      %{a: a, b: b, c: c, suggestion: suggestion} ->
        %{a: a, b: b, missing_link: c, suggestion: suggestion}

      other ->
        %{raw: inspect(other)}
    end)
  end

  defp format_triangles(_), do: []

  # GraphAnalyzer.clusters/0 returns [MapSet.t()]
  defp format_clusters(clusters) when is_list(clusters) do
    clusters
    |> Enum.with_index()
    |> Enum.map(fn {members, idx} ->
      member_list = MapSet.to_list(members)
      %{id: idx, size: length(member_list), members: member_list}
    end)
  end

  defp format_clusters(_), do: []

  # Reflector.reflect/1 returns [%{source:, target:, cooccurrences:, confidence:, suggested_relation:, reason:}]
  defp format_gaps(gaps) when is_list(gaps) do
    Enum.map(gaps, fn
      %{source: s, target: t, cooccurrences: c, confidence: conf} = gap ->
        %{
          source: s,
          target: t,
          cooccurrences: c,
          confidence: Float.round(conf, 2),
          suggested_relation: Map.get(gap, :suggested_relation, "related"),
          reason: Map.get(gap, :reason, nil)
        }

      other ->
        %{raw: inspect(other)}
    end)
  end

  defp format_gaps(_), do: []

  # OptimalEngine.search/2 returns [%Context{}]
  defp format_search_results(contexts) when is_list(contexts) do
    Enum.map(contexts, fn ctx ->
      %{
        id: Map.get(ctx, :id),
        title: Map.get(ctx, :title),
        node: Map.get(ctx, :node),
        genre: Map.get(ctx, :genre),
        uri: Map.get(ctx, :uri),
        l0_abstract: Map.get(ctx, :l0_abstract),
        sn_ratio: Map.get(ctx, :sn_ratio)
      }
    end)
  end

  defp format_search_results(_), do: []

  # HealthDiagnostics.run/0 returns [%{name:, severity:, message:, details:, fix:}]
  defp format_health(checks) when is_list(checks) do
    Enum.map(checks, fn
      %{name: name, severity: severity, message: message} = check ->
        %{
          check: to_string(name),
          status: to_string(severity),
          message: message,
          details: Map.get(check, :details, []),
          fix: Map.get(check, :fix, nil)
        }

      other ->
        %{raw: inspect(other)}
    end)
  end

  defp format_health(_), do: []
end
