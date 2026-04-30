defmodule OptimalEngine.Memory.Versioned.Store do
  @moduledoc """
  Raw SQL operations for the versioned memory primitive.

  This module is internal — only `OptimalEngine.Memory.Versioned` should
  call it. All SQL lives here; the facade owns business logic.

  Column layout mirrors the `memories` table created in migration 028.
  Boolean fields (`is_static`, `is_forgotten`, `is_latest`) are stored as
  SQLite INTEGER (0/1) and deserialized to Elixir booleans here.
  """

  alias OptimalEngine.Store

  @select_cols """
  id, tenant_id, workspace_id, content, is_static, is_forgotten,
  forget_after, forget_reason, version, parent_memory_id, root_memory_id,
  is_latest, citation_uri, source_chunk_id, audience, metadata,
  created_at, updated_at
  """

  # Columns used in the dedup lookup — must match @select_cols order exactly.
  @dedup_select_cols @select_cols

  # ---------------------------------------------------------------------------
  # Write operations
  # ---------------------------------------------------------------------------

  @doc "Inserts a new memory row. Returns the raw row list on success."
  @spec insert(map()) :: {:ok, list()} | {:error, term()}
  def insert(attrs) do
    sql = """
    INSERT INTO memories (
      id, tenant_id, workspace_id, content, is_static, is_forgotten,
      version, parent_memory_id, root_memory_id, is_latest,
      citation_uri, source_chunk_id, audience, metadata, content_hash
    ) VALUES (?1, ?2, ?3, ?4, ?5, 0, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14)
    """

    params = [
      attrs.id,
      attrs.tenant_id,
      attrs.workspace_id,
      attrs.content,
      bool_to_int(attrs[:is_static] || false),
      attrs[:version] || 1,
      attrs[:parent_memory_id],
      attrs[:root_memory_id],
      bool_to_int(attrs[:is_latest] != false),
      attrs[:citation_uri],
      attrs[:source_chunk_id],
      attrs[:audience] || "default",
      Jason.encode!(attrs[:metadata] || %{}),
      attrs[:content_hash]
    ]

    Store.raw_query(sql, params)
  end

  @doc """
  Looks up an existing live memory by workspace_id, audience, and content_hash.

  "Live" means `is_forgotten = 0` AND `is_latest = 1`. Returns `{:ok, row}`
  when found, `{:error, :not_found}` when absent.
  """
  @spec find_by_content_hash(String.t(), String.t(), String.t()) ::
          {:ok, list()} | {:error, :not_found}
  def find_by_content_hash(workspace_id, audience, content_hash)
      when is_binary(workspace_id) and is_binary(audience) and is_binary(content_hash) do
    sql = """
    SELECT #{@dedup_select_cols}
    FROM memories
    WHERE workspace_id = ?1
      AND audience = ?2
      AND content_hash = ?3
      AND is_forgotten = 0
      AND is_latest = 1
    LIMIT 1
    """

    case Store.raw_query(sql, [workspace_id, audience, content_hash]) do
      {:ok, [row]} -> {:ok, row}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc "Fetches a single memory row by id."
  @spec get(String.t()) :: {:ok, list()} | {:error, :not_found}
  def get(id) when is_binary(id) do
    sql = "SELECT #{@select_cols} FROM memories WHERE id = ?1"

    case Store.raw_query(sql, [id]) do
      {:ok, [row]} -> {:ok, row}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc """
  Lists memories filtered by opts:
    - `:workspace_id` (required)
    - `:audience` — filter by audience string
    - `:include_forgotten` — default false
    - `:include_old_versions` — default false (only is_latest=1)
    - `:limit` — default 50
  """
  @spec list(keyword()) :: {:ok, [list()]} | {:error, term()}
  def list(opts) do
    workspace_id = Keyword.get(opts, :workspace_id, "default")
    audience = Keyword.get(opts, :audience)
    include_forgotten = Keyword.get(opts, :include_forgotten, false)
    include_old = Keyword.get(opts, :include_old_versions, false)
    limit = Keyword.get(opts, :limit, 50)

    conditions = ["workspace_id = ?1"]
    params = [workspace_id]
    param_idx = 2

    {conditions, params, param_idx} =
      if audience do
        {conditions ++ ["audience = ?#{param_idx}"], params ++ [audience], param_idx + 1}
      else
        {conditions, params, param_idx}
      end

    {conditions, _params, _param_idx} =
      if not include_forgotten do
        {conditions ++ ["is_forgotten = 0"], params, param_idx}
      else
        {conditions, params, param_idx}
      end

    {conditions, _params, _param_idx} =
      if not include_old do
        {conditions ++ ["is_latest = 1"], params, param_idx}
      else
        {conditions, params, param_idx}
      end

    where = Enum.join(conditions, " AND ")

    sql =
      "SELECT #{@select_cols} FROM memories WHERE #{where} ORDER BY created_at DESC LIMIT ?#{param_idx}"

    final_params = params ++ [limit]

    Store.raw_query(sql, final_params)
  end

  @doc """
  Marks an old version as no longer latest. Called during update versioning.
  """
  @spec demote_latest(String.t()) :: {:ok, term()} | {:error, term()}
  def demote_latest(id) when is_binary(id) do
    Store.raw_query(
      "UPDATE memories SET is_latest = 0, updated_at = datetime('now') WHERE id = ?1",
      [id]
    )
  end

  @doc "Marks a memory as forgotten (soft delete)."
  @spec mark_forgotten(String.t(), String.t() | nil, String.t() | nil) ::
          {:ok, term()} | {:error, term()}
  def mark_forgotten(id, reason, forget_after) when is_binary(id) do
    Store.raw_query(
      """
      UPDATE memories
      SET is_forgotten = 1, forget_reason = ?2, forget_after = ?3,
          updated_at = datetime('now')
      WHERE id = ?1
      """,
      [id, reason, forget_after]
    )
  end

  @doc "Hard deletes a memory row (cascades to memory_relations)."
  @spec delete(String.t()) :: {:ok, term()} | {:error, term()}
  def delete(id) when is_binary(id) do
    Store.raw_query("DELETE FROM memories WHERE id = ?1", [id])
  end

  # ---------------------------------------------------------------------------
  # Relation operations
  # ---------------------------------------------------------------------------

  @doc "Inserts a typed relation between two memories."
  @spec add_relation(String.t(), String.t(), String.t(), String.t(), String.t()) ::
          {:ok, term()} | {:error, term()}
  def add_relation(source_id, target_id, relation, workspace_id, tenant_id) do
    sql = """
    INSERT OR IGNORE INTO memory_relations
      (tenant_id, workspace_id, source_memory_id, target_memory_id, relation)
    VALUES (?1, ?2, ?3, ?4, ?5)
    """

    Store.raw_query(sql, [tenant_id, workspace_id, source_id, target_id, relation])
  end

  @doc """
  Returns all relations touching `memory_id` — both outbound (source) and
  inbound (target). Each row: [id, source_memory_id, target_memory_id, relation, created_at].
  """
  @spec get_relations(String.t()) :: {:ok, [list()]} | {:error, term()}
  def get_relations(memory_id) when is_binary(memory_id) do
    sql = """
    SELECT id, source_memory_id, target_memory_id, relation, created_at
    FROM memory_relations
    WHERE source_memory_id = ?1 OR target_memory_id = ?1
    ORDER BY created_at ASC
    """

    Store.raw_query(sql, [memory_id])
  end

  @doc """
  Returns all memories in the version chain rooted at `root_id`, ordered
  by version ascending.
  """
  @spec get_version_chain(String.t(), String.t()) :: {:ok, [list()]} | {:error, term()}
  def get_version_chain(root_id, workspace_id) when is_binary(root_id) do
    sql = """
    SELECT #{@select_cols} FROM memories
    WHERE workspace_id = ?1 AND (root_memory_id = ?2 OR id = ?2)
    ORDER BY version ASC
    """

    Store.raw_query(sql, [workspace_id, root_id])
  end

  # ---------------------------------------------------------------------------
  # Helpers
  # ---------------------------------------------------------------------------

  defp bool_to_int(true), do: 1
  defp bool_to_int(false), do: 0
  defp bool_to_int(1), do: 1
  defp bool_to_int(0), do: 0
end
