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

  alias OptimalEngine.{GraphAnalyzer, HealthDiagnostics, Reflector, Store}

  plug(:cors)
  plug(:match)
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
    |> put_resp_header("access-control-allow-methods", "GET, OPTIONS")
    |> put_resp_header("access-control-allow-headers", "content-type")
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
