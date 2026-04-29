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
    GET /api/l0              — L0 context cache
    GET /api/health          — Health diagnostics

  All responses are JSON with `Content-Type: application/json`.
  CORS headers are set on every response (GET + OPTIONS allowed).
  """

  use Plug.Router

  alias OptimalEngine.Graph.Analyzer, as: GraphAnalyzer
  alias OptimalEngine.Health
  alias OptimalEngine.Insight.Health, as: HealthDiagnostics
  alias OptimalEngine.Graph.Reflector, as: Reflector
  alias OptimalEngine.Retrieval
  alias OptimalEngine.Retrieval.Receiver
  alias OptimalEngine.Store
  alias OptimalEngine.Telemetry
  alias OptimalEngine.Wiki

  plug(:cors)
  plug(:match)
  plug(Plug.Parsers, parsers: [:json], json_decoder: Jason, pass: ["application/json"])
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

    results =
      if q == "" do
        []
      else
        case OptimalEngine.search(q, limit: limit) do
          {:ok, contexts} -> format_search_results(contexts)
          _ -> []
        end
      end

    json(conn, %{query: q, results: results})
  end

  get "/api/l0" do
    json(conn, %{l0: OptimalEngine.l0()})
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
  # Body: {"query": "…", "format": "markdown", "audience": "default", "bandwidth": "medium"}
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

        {:ok, result} = Retrieval.ask(q, receiver: receiver)
        json(conn, result)

      _ ->
        send_resp(conn, 400, Jason.encode!(%{error: "query is required"}))
    end
  end

  # Wiki listing — GET /api/wiki?tenant=default
  get "/api/wiki" do
    tenant = query_param(conn, "tenant", "default")

    case Wiki.list(tenant) do
      {:ok, pages} ->
        json(conn, %{
          pages:
            Enum.map(pages, fn p ->
              %{
                slug: p.slug,
                audience: p.audience,
                version: p.version,
                last_curated: p.last_curated,
                curated_by: p.curated_by,
                size_bytes: byte_size(p.body || "")
              }
            end)
        })

      _ ->
        json(conn, %{pages: []})
    end
  end

  # Wiki page render — GET /api/wiki/:slug?audience=default&format=markdown
  get "/api/wiki/:slug" do
    tenant = query_param(conn, "tenant", "default")
    audience = query_param(conn, "audience", "default")
    format = query_param(conn, "format", "markdown") |> parse_atom(:markdown)

    case Wiki.latest(tenant, slug, audience) do
      {:ok, page} ->
        resolver = fn _d, _opts -> {:ok, "", %{}} end
        {rendered, warnings} = Wiki.render(page, resolver, format: format)

        json(conn, %{
          slug: page.slug,
          audience: page.audience,
          version: page.version,
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
  get "/api/workspace" do
    json(conn, %{nodes: workspace_nodes()})
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
    |> put_resp_header("access-control-allow-methods", "GET, POST, OPTIONS")
    |> put_resp_header("access-control-allow-headers", "content-type")
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
  defp workspace_nodes do
    sql = """
    SELECT n.id, n.slug, n.name, n.kind, n.parent_id, n.style, n.status,
           COALESCE((SELECT COUNT(*) FROM contexts c WHERE c.node = n.slug), 0) AS signal_count
    FROM nodes n
    ORDER BY COALESCE(n.parent_id, ''), n.slug
    """

    case Store.raw_query(sql, []) do
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
