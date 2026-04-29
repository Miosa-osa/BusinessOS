defmodule OptimalEngine.Workspace.NodeMember do
  @moduledoc """
  Ties a principal to a node with an explicit membership kind.

  A principal can be a member of many nodes; a node has many members.
  Membership is time-bounded (`started_at` + `ended_at`) so historical
  queries ("who was on the sales team on 2026-03-15?") are answerable.

  Membership kinds:

    * `:owner`    — operates the node; can add members, edit metadata
    * `:internal` — employee / contributor inside the company
    * `:external` — client, partner, vendor; outside the company
    * `:observer` — read-only visibility (auditors, AI agents)
  """

  alias OptimalEngine.Store
  alias OptimalEngine.Tenancy.Tenant

  @type membership :: :owner | :internal | :external | :observer

  @allowed_memberships [:owner, :internal, :external, :observer]

  @doc """
  Add a principal to a node with a membership kind. Idempotent on
  `(node_id, principal_id, membership)`.

  Options:
    * `:membership` — default `:internal`
    * `:role`       — optional string ("lead", "designer", "CSM", …)
    * `:tenant_id`  — defaults to default tenant
  """
  @spec add(String.t(), String.t(), keyword()) :: :ok | {:error, term()}
  def add(node_id, principal_id, opts \\ [])
      when is_binary(node_id) and is_binary(principal_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    membership = Keyword.get(opts, :membership, :internal)
    role = Keyword.get(opts, :role)

    if membership not in @allowed_memberships do
      {:error, {:invalid_membership, membership}}
    else
      sql = """
      INSERT OR IGNORE INTO node_members (tenant_id, node_id, principal_id, membership, role)
      VALUES (?1, ?2, ?3, ?4, ?5)
      """

      case Store.raw_query(sql, [tenant_id, node_id, principal_id, Atom.to_string(membership), role]) do
        {:ok, _} -> :ok
        other -> other
      end
    end
  end

  @doc "Remove (end) a membership. Sets `ended_at` rather than deleting."
  @spec remove(String.t(), String.t(), keyword()) :: :ok | {:error, term()}
  def remove(node_id, principal_id, opts \\ [])
      when is_binary(node_id) and is_binary(principal_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    membership = Keyword.get(opts, :membership)

    {sql, params} =
      case membership do
        nil ->
          {"""
           UPDATE node_members
           SET ended_at = datetime('now')
           WHERE node_id = ?1 AND principal_id = ?2 AND tenant_id = ?3 AND ended_at IS NULL
           """, [node_id, principal_id, tenant_id]}

        m when m in @allowed_memberships ->
          {"""
           UPDATE node_members
           SET ended_at = datetime('now')
           WHERE node_id = ?1 AND principal_id = ?2 AND tenant_id = ?3
             AND membership = ?4 AND ended_at IS NULL
           """, [node_id, principal_id, tenant_id, Atom.to_string(m)]}
      end

    case Store.raw_query(sql, params) do
      {:ok, _} -> :ok
      other -> other
    end
  end

  @doc """
  Returns every active member of the node, optionally filtered by
  membership kind.
  """
  @spec members_of(String.t(), keyword()) :: {:ok, [map()]}
  def members_of(node_id, opts \\ []) when is_binary(node_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    kind_filter = Keyword.get(opts, :membership)

    {where, params} = {["node_id = ?1", "tenant_id = ?2", "ended_at IS NULL"], [node_id, tenant_id]}

    {where, params} =
      case kind_filter do
        nil ->
          {where, params}

        m when m in @allowed_memberships ->
          {where ++ ["membership = ?#{length(params) + 1}"], params ++ [Atom.to_string(m)]}
      end

    sql = """
    SELECT principal_id, membership, role, started_at, ended_at
    FROM node_members
    WHERE #{Enum.join(where, " AND ")}
    ORDER BY started_at
    """

    case Store.raw_query(sql, params) do
      {:ok, rows} ->
        {:ok,
         Enum.map(rows, fn [pid, m, role, started_at, ended_at] ->
           %{
             principal_id: pid,
             membership: safe_atom(m, :internal),
             role: role,
             started_at: started_at,
             ended_at: ended_at
           }
         end)}

      other ->
        other
    end
  end

  @doc """
  Returns every active node the principal is a member of, optionally filtered
  by membership kind.
  """
  @spec nodes_of(String.t(), keyword()) :: {:ok, [map()]}
  def nodes_of(principal_id, opts \\ []) when is_binary(principal_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    kind_filter = Keyword.get(opts, :membership)

    {where, params} =
      {["principal_id = ?1", "tenant_id = ?2", "ended_at IS NULL"], [principal_id, tenant_id]}

    {where, params} =
      case kind_filter do
        nil ->
          {where, params}

        m when m in @allowed_memberships ->
          {where ++ ["membership = ?#{length(params) + 1}"], params ++ [Atom.to_string(m)]}
      end

    sql = """
    SELECT node_id, membership, role, started_at, ended_at
    FROM node_members
    WHERE #{Enum.join(where, " AND ")}
    ORDER BY started_at
    """

    case Store.raw_query(sql, params) do
      {:ok, rows} ->
        {:ok,
         Enum.map(rows, fn [nid, m, role, started_at, ended_at] ->
           %{
             node_id: nid,
             membership: safe_atom(m, :internal),
             role: role,
             started_at: started_at,
             ended_at: ended_at
           }
         end)}

      other ->
        other
    end
  end

  defp safe_atom(nil, fallback), do: fallback

  defp safe_atom(str, fallback) when is_binary(str) do
    try do
      String.to_existing_atom(str)
    rescue
      ArgumentError -> fallback
    end
  end
end
