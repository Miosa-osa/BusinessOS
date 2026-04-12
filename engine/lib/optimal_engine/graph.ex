defmodule OptimalEngine.Graph do
  @moduledoc """
  Knowledge graph edge management.

  Edges connect contexts, entities, and nodes in a typed directed graph.

  ## Edge types (relation values)

  - `mentioned_in`  — entity → context: entity appears in a context
  - `lives_in`      — context → node: context belongs to a node
  - `cross_ref`     — context → context: signal is cross-referenced to another node
  - `works_on`      — entity → node: person works on a node (from topology)
  - `supersedes`    — context → context: newer context supersedes an older one

  ## Node conventions

  Node names are bare strings like "ai-masters", "miosa-platform", etc. — they are
  NOT context IDs. Entity names like "Ed Honour" are also bare strings. Only
  context-to-context edges use real context IDs as both source and target.

  The edges table has no FK constraints (migrated from the original FK schema) so
  that entity, node, and context IDs can all coexist as edge endpoints.
  """

  require Logger

  alias OptimalEngine.Store

  @doc """
  Creates all edges for a newly indexed or ingested context.

  Inserts:
  - `entity --mentioned_in--> context.id` for each entity
  - `context.id --lives_in--> context.node`
  - `context.id --cross_ref--> dest` for each extra destination in routed_to
  - `context.id --supersedes--> supersedes_id` if context.supersedes is set

  This variant is used outside the Store GenServer (e.g. post-indexing tasks).
  For calls from within the Store GenServer, use `create_edges_for_context_db/2`
  to pass the raw db reference directly and avoid the re-entrant call deadlock.
  """
  @spec create_edges_for_context(map() | struct()) :: :ok
  def create_edges_for_context(context) do
    id = struct_or_map_get(context, :id)

    if is_binary(id) do
      node = struct_or_map_get(context, :node)
      entities = struct_or_map_get(context, :entities) || []
      routed_to = struct_or_map_get(context, :routed_to) || []
      supersedes = struct_or_map_get(context, :supersedes)

      now = DateTime.to_iso8601(DateTime.utc_now())

      # entity → context (mentioned_in)
      Enum.each(entities, fn entity ->
        insert_edge_via_store(entity, id, "mentioned_in", 1.0, now)
      end)

      # context → node (lives_in)
      if is_binary(node) and node != "" do
        insert_edge_via_store(id, node, "lives_in", 1.0, now)
      end

      # context → context (cross_ref) — extra destinations beyond the primary node
      cross_refs = Enum.reject(routed_to, fn dest -> dest == node end)

      Enum.each(cross_refs, fn dest ->
        insert_edge_via_store(id, dest, "cross_ref", 0.8, now)
      end)

      # context → context (supersedes)
      if is_binary(supersedes) and supersedes != "" do
        insert_edge_via_store(id, supersedes, "supersedes", 1.0, now)
      end
    end

    :ok
  end

  @doc """
  Creates edges for a context using a raw SQLite db reference.

  Used from within the Store GenServer to avoid re-entrant `GenServer.call` deadlocks.
  The `db` argument is the raw Exqlite connection reference.
  """
  @spec create_edges_for_context_db(reference(), map() | struct()) :: :ok
  def create_edges_for_context_db(db, context) do
    id = struct_or_map_get(context, :id)

    if is_binary(id) do
      node = struct_or_map_get(context, :node)
      entities = struct_or_map_get(context, :entities) || []
      routed_to = struct_or_map_get(context, :routed_to) || []
      supersedes = struct_or_map_get(context, :supersedes)

      now = DateTime.to_iso8601(DateTime.utc_now())

      Enum.each(entities, fn entity ->
        insert_edge_direct(db, entity, id, "mentioned_in", 1.0, now)
      end)

      if is_binary(node) and node != "" do
        insert_edge_direct(db, id, node, "lives_in", 1.0, now)
      end

      cross_refs = Enum.reject(routed_to, fn dest -> dest == node end)

      Enum.each(cross_refs, fn dest ->
        insert_edge_direct(db, id, dest, "cross_ref", 0.8, now)
      end)

      if is_binary(supersedes) and supersedes != "" do
        insert_edge_direct(db, id, supersedes, "supersedes", 1.0, now)
      end
    end

    :ok
  end

  @doc """
  Seeds `works_on` edges from topology — entity → node relationships.

  These encode which people work on which nodes. Idempotent (INSERT OR IGNORE).
  """
  @spec seed_from_topology() :: {:ok, non_neg_integer()} | {:error, term()}
  def seed_from_topology do
    now = DateTime.to_iso8601(DateTime.utc_now())

    edges = topology_works_on_edges()

    inserted =
      Enum.reduce(edges, 0, fn {entity, node}, count ->
        case insert_edge_via_store(entity, node, "works_on", 1.0, now) do
          :ok -> count + 1
          {:error, _} -> count
        end
      end)

    Logger.info("[Graph] Seeded #{inserted} works_on edges from topology")
    {:ok, inserted}
  end

  @doc """
  Rebuilds ALL edges from scratch:
  1. Clears the edges table
  2. Iterates all contexts and creates their edges
  3. Seeds topology works_on edges
  """
  @spec rebuild() :: {:ok, non_neg_integer()} | {:error, term()}
  def rebuild do
    Logger.info("[Graph] Rebuilding all edges...")

    with :ok <- clear_edges(),
         {:ok, context_count} <- rebuild_context_edges(),
         {:ok, topology_count} <- seed_from_topology() do
      total = context_count + topology_count
      Logger.info("[Graph] Edge rebuild complete. #{total} edges created.")
      {:ok, total}
    end
  end

  @doc """
  Returns all edges for a given source or target ID.

  Options:
  - `:direction` — `:out` (default), `:in`, or `:both`
  - `:relation`  — filter by relation type
  """
  @spec edges_for(String.t(), keyword()) :: {:ok, [map()]} | {:error, term()}
  def edges_for(id, opts \\ []) when is_binary(id) do
    direction = Keyword.get(opts, :direction, :out)
    relation = Keyword.get(opts, :relation)

    {where_clause, params} = build_edges_where(id, direction, relation)

    sql = "SELECT source_id, target_id, relation, weight FROM edges WHERE #{where_clause}"

    case Store.raw_query(sql, params) do
      {:ok, rows} ->
        edges = Enum.map(rows, &row_to_edge/1)
        {:ok, edges}

      err ->
        err
    end
  end

  @doc """
  Returns all contexts that mention a given entity (traverse mentioned_in edges).
  """
  @spec related_contexts(String.t()) :: {:ok, [String.t()]} | {:error, term()}
  def related_contexts(entity_name) when is_binary(entity_name) do
    sql = """
    SELECT target_id FROM edges
    WHERE source_id = ?1 AND relation = 'mentioned_in'
    """

    case Store.raw_query(sql, [entity_name]) do
      {:ok, rows} -> {:ok, Enum.map(rows, fn [id] -> id end)}
      err -> err
    end
  end

  @doc """
  Returns the 1-hop subgraph around a context ID.
  Returns all edges where source_id OR target_id equals the context_id.
  """
  @spec subgraph(String.t(), pos_integer()) :: {:ok, [map()]} | {:error, term()}
  def subgraph(context_id, _depth \\ 1) when is_binary(context_id) do
    sql = """
    SELECT source_id, target_id, relation, weight FROM edges
    WHERE source_id = ?1 OR target_id = ?1
    """

    case Store.raw_query(sql, [context_id]) do
      {:ok, rows} ->
        edges = Enum.map(rows, &row_to_edge/1)
        {:ok, edges}

      err ->
        err
    end
  end

  @doc """
  Returns aggregate stats about the graph.
  """
  @spec stats() :: {:ok, map()} | {:error, term()}
  def stats do
    queries = [
      {"total_edges", "SELECT COUNT(*) FROM edges"},
      {"mentioned_in", "SELECT COUNT(*) FROM edges WHERE relation = 'mentioned_in'"},
      {"lives_in", "SELECT COUNT(*) FROM edges WHERE relation = 'lives_in'"},
      {"works_on", "SELECT COUNT(*) FROM edges WHERE relation = 'works_on'"},
      {"cross_ref", "SELECT COUNT(*) FROM edges WHERE relation = 'cross_ref'"},
      {"supersedes", "SELECT COUNT(*) FROM edges WHERE relation = 'supersedes'"}
    ]

    result =
      Enum.reduce_while(queries, %{}, fn {key, sql}, acc ->
        case Store.raw_query(sql, []) do
          {:ok, [[val]]} -> {:cont, Map.put(acc, key, val)}
          {:ok, []} -> {:cont, Map.put(acc, key, 0)}
          {:error, _} = err -> {:halt, err}
        end
      end)

    case result do
      %{} = s -> {:ok, s}
      err -> err
    end
  end

  @doc """
  Returns the top N most connected entities by edge count.
  """
  @spec top_entities(pos_integer()) :: {:ok, [{String.t(), non_neg_integer()}]} | {:error, term()}
  def top_entities(limit \\ 10) do
    sql = """
    SELECT source_id, COUNT(*) as cnt
    FROM edges
    WHERE relation IN ('mentioned_in', 'works_on')
    GROUP BY source_id
    ORDER BY cnt DESC
    LIMIT ?1
    """

    case Store.raw_query(sql, [limit]) do
      {:ok, rows} -> {:ok, Enum.map(rows, fn [id, cnt] -> {id, cnt} end)}
      err -> err
    end
  end

  @doc """
  Returns sample edges for display.
  """
  @spec sample_edges(pos_integer()) :: {:ok, [map()]} | {:error, term()}
  def sample_edges(limit \\ 10) do
    sql = """
    SELECT source_id, target_id, relation, weight
    FROM edges
    ORDER BY ROWID DESC
    LIMIT ?1
    """

    case Store.raw_query(sql, [limit]) do
      {:ok, rows} -> {:ok, Enum.map(rows, &row_to_edge/1)}
      err -> err
    end
  end

  # ---------------------------------------------------------------------------
  # Private: topology works_on edges
  # ---------------------------------------------------------------------------

  defp topology_works_on_edges do
    # Static mapping: entity name → [nodes they work on]
    # Derived from CLAUDE.md routing table and OptimalOS topology
    [
      {"Ed Honour", "ai-masters"},
      {"Robert Potter", "ai-masters"},
      {"Adam", "ai-masters"},
      {"Pedro", "miosa-platform"},
      {"Abdul", "miosa-platform"},
      {"Nejd", "miosa-platform"},
      {"Pedram", "miosa-platform"},
      {"Javaris", "miosa-platform"},
      {"Bennett", "agency-accelerants"},
      {"Bennett", "content-creators"},
      {"Bennett", "accelerants-community"},
      {"Len", "agency-accelerants"},
      {"Liam", "agency-accelerants"},
      {"Ahmed", "os-architect"},
      {"Ahmed", "content-creators"},
      {"Tejas", "content-creators"},
      {"Ikram", "content-creators"},
      {"Jordan", "roberto"},
      {"Roberto", "roberto"},
      {"Roberto", "miosa-platform"},
      {"Roberto", "ai-masters"},
      {"Roberto", "os-architect"},
      {"Roberto", "agency-accelerants"},
      {"Roberto", "accelerants-community"},
      {"Roberto", "content-creators"},
      {"Roberto", "money-revenue"},
      {"Roberto", "team"},
      {"Roberto", "lunivate"},
      {"Roberto", "os-accelerator"}
    ]
  end

  # ---------------------------------------------------------------------------
  # Private: rebuild helpers
  # ---------------------------------------------------------------------------

  defp clear_edges do
    case Store.raw_query("DELETE FROM edges", []) do
      {:ok, _} -> :ok
      :ok -> :ok
      err -> err
    end
  end

  defp rebuild_context_edges do
    sql = """
    SELECT id, node, entities, routed_to, supersedes
    FROM contexts
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        Enum.each(rows, fn [id, node, entities_json, routed_to_json, supersedes] ->
          entities = decode_json_list(entities_json)
          routed_to = decode_json_list(routed_to_json)

          context = %{
            id: id,
            node: node,
            entities: entities,
            routed_to: routed_to,
            supersedes: supersedes
          }

          create_edges_for_context(context)
        end)

        # Return actual count from the table after insertion
        case Store.raw_query("SELECT COUNT(*) FROM edges", []) do
          {:ok, [[count]]} -> {:ok, count}
          {:ok, []} -> {:ok, 0}
          err -> err
        end

      err ->
        err
    end
  end

  # ---------------------------------------------------------------------------
  # Private: SQL helpers
  # ---------------------------------------------------------------------------

  @edge_insert_sql """
  INSERT OR IGNORE INTO edges (source_id, target_id, relation, weight, valid_from)
  VALUES (?1, ?2, ?3, ?4, ?5)
  """

  # Insert via Store GenServer (for use outside of Store callbacks)
  defp insert_edge_via_store(source, target, relation, weight, now) do
    case Store.raw_query(@edge_insert_sql, [source, target, relation, weight, now]) do
      {:ok, _} ->
        :ok

      :ok ->
        :ok

      {:error, reason} ->
        Logger.debug("[Graph] Edge insert skipped: #{inspect(reason)}")
        {:error, reason}
    end
  end

  # Insert directly via db reference (for use inside Store callbacks — avoids deadlock)
  defp insert_edge_direct(db, source, target, relation, weight, now) do
    with {:ok, stmt} <- Exqlite.Sqlite3.prepare(db, @edge_insert_sql),
         :ok <- Exqlite.Sqlite3.bind(stmt, [source, target, relation, weight, now]),
         _ <- Exqlite.Sqlite3.step(db, stmt),
         :ok <- Exqlite.Sqlite3.release(db, stmt) do
      :ok
    else
      {:error, reason} ->
        Logger.debug("[Graph] Edge direct insert skipped: #{inspect(reason)}")
        :ok
    end
  end

  defp build_edges_where(id, :out, nil), do: {"source_id = ?1", [id]}
  defp build_edges_where(id, :in, nil), do: {"target_id = ?1", [id]}
  defp build_edges_where(id, :both, nil), do: {"source_id = ?1 OR target_id = ?1", [id]}
  defp build_edges_where(id, :out, rel), do: {"source_id = ?1 AND relation = ?2", [id, rel]}
  defp build_edges_where(id, :in, rel), do: {"target_id = ?1 AND relation = ?2", [id, rel]}

  defp build_edges_where(id, :both, rel),
    do: {"(source_id = ?1 OR target_id = ?1) AND relation = ?2", [id, rel]}

  defp row_to_edge([source_id, target_id, relation, weight]) do
    %{source_id: source_id, target_id: target_id, relation: relation, weight: weight}
  end

  defp decode_json_list(nil), do: []
  defp decode_json_list(""), do: []

  defp decode_json_list(json) when is_binary(json) do
    case Jason.decode(json) do
      {:ok, list} when is_list(list) -> list
      _ -> []
    end
  end

  # Works for both plain maps (%{key: val}) and structs (which don't support Access)
  defp struct_or_map_get(%_{} = struct, key), do: Map.get(struct, key)
  defp struct_or_map_get(map, key) when is_map(map), do: Map.get(map, key)
  defp struct_or_map_get(_, _), do: nil
end
