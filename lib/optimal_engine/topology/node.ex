defmodule OptimalEngine.Topology.Node do
  @moduledoc """
  An organizational unit inside a tenant's workspace — a team, a project, an
  external entity, a domain, a named person, etc.

  Nodes form a tree (`parent_id`). They're the anchor for every signal the
  engine ingests: routing rules point at them, wiki pages are organized by
  them, retrieval boosts chunks rooted in the caller's nodes.

  Before Phase 3.5 these existed as freeform strings on `contexts.node`;
  after Phase 3.5 they're first-class rows with kind + style + path +
  status + versionable metadata.

  See `docs/architecture/WORKSPACE.md` for the full model.
  """

  alias OptimalEngine.Store
  alias OptimalEngine.Tenancy.Tenant

  @type kind :: :unit | :team | :project | :entity | :domain | :person
  @type style :: :internal | :external | :mixed
  @type status :: :active | :archived | :draft

  @type t :: %__MODULE__{
          id: String.t(),
          tenant_id: String.t(),
          slug: String.t(),
          name: String.t(),
          kind: kind(),
          parent_id: String.t() | nil,
          description: String.t() | nil,
          style: style(),
          status: status(),
          path: String.t(),
          metadata: map(),
          created_at: String.t() | nil
        }

  defstruct id: nil,
            tenant_id: Tenant.default_id(),
            slug: nil,
            name: nil,
            kind: :unit,
            parent_id: nil,
            description: nil,
            style: :internal,
            status: :active,
            path: "",
            metadata: %{},
            created_at: nil

  @allowed_kinds [:unit, :team, :project, :entity, :domain, :person]
  @allowed_styles [:internal, :external, :mixed]
  @allowed_statuses [:active, :archived, :draft]

  @doc """
  Upsert a node. `slug` + `tenant_id` is the natural key; `id` defaults to
  `"{tenant_id}:{slug}"` for a deterministic, readable identifier.
  """
  @spec upsert(map()) :: {:ok, t()} | {:error, term()}
  def upsert(%{slug: slug, name: name, kind: kind} = attrs)
      when is_binary(slug) and is_binary(name) and kind in @allowed_kinds do
    tenant_id = Map.get(attrs, :tenant_id, Tenant.default_id())
    id = Map.get(attrs, :id) || "#{tenant_id}:#{slug}"
    parent_id = Map.get(attrs, :parent_id)
    description = Map.get(attrs, :description)
    style = Map.get(attrs, :style, :internal)
    status = Map.get(attrs, :status, :active)
    path = Map.get(attrs, :path, "nodes/#{slug}")
    metadata = Map.get(attrs, :metadata, %{})

    cond do
      style not in @allowed_styles ->
        {:error, {:invalid_style, style}}

      status not in @allowed_statuses ->
        {:error, {:invalid_status, status}}

      true ->
        sql = """
        INSERT INTO nodes (id, tenant_id, slug, name, kind, parent_id, description,
                           style, status, path, metadata)
        VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11)
        ON CONFLICT(id) DO UPDATE SET
          name        = excluded.name,
          kind        = excluded.kind,
          parent_id   = excluded.parent_id,
          description = excluded.description,
          style       = excluded.style,
          status      = excluded.status,
          path        = excluded.path,
          metadata    = excluded.metadata
        """

        params = [
          id,
          tenant_id,
          slug,
          name,
          Atom.to_string(kind),
          parent_id,
          description,
          Atom.to_string(style),
          Atom.to_string(status),
          path,
          Jason.encode!(metadata)
        ]

        case Store.raw_query(sql, params) do
          {:ok, _} ->
            {:ok,
             %__MODULE__{
               id: id,
               tenant_id: tenant_id,
               slug: slug,
               name: name,
               kind: kind,
               parent_id: parent_id,
               description: description,
               style: style,
               status: status,
               path: path,
               metadata: metadata
             }}

          other ->
            other
        end
    end
  end

  @doc "Fetch a node by id, tenant-scoped."
  @spec get(String.t(), String.t()) :: {:ok, t()} | {:error, :not_found}
  def get(id, tenant_id \\ Tenant.default_id()) when is_binary(id) do
    case Store.raw_query(
           "SELECT " <>
             select_columns() <>
             " FROM nodes WHERE id = ?1 AND tenant_id = ?2",
           [id, tenant_id]
         ) do
      {:ok, [row]} -> {:ok, row_to_struct(row)}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc "Fetch by slug within the tenant."
  @spec get_by_slug(String.t(), String.t()) :: {:ok, t()} | {:error, :not_found}
  def get_by_slug(slug, tenant_id \\ Tenant.default_id()) when is_binary(slug) do
    case Store.raw_query(
           "SELECT " <>
             select_columns() <>
             " FROM nodes WHERE slug = ?1 AND tenant_id = ?2",
           [slug, tenant_id]
         ) do
      {:ok, [row]} -> {:ok, row_to_struct(row)}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc "Lists all nodes in a tenant, optionally filtered by kind and/or status."
  @spec list(keyword()) :: {:ok, [t()]}
  def list(opts \\ []) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())

    {clauses, params} = {["tenant_id = ?1"], [tenant_id]}

    {clauses, params} =
      case Keyword.get(opts, :kind) do
        nil ->
          {clauses, params}

        k when is_atom(k) ->
          {clauses ++ ["kind = ?#{length(params) + 1}"], params ++ [Atom.to_string(k)]}
      end

    {clauses, params} =
      case Keyword.get(opts, :status) do
        nil ->
          {clauses, params}

        s when is_atom(s) ->
          {clauses ++ ["status = ?#{length(params) + 1}"], params ++ [Atom.to_string(s)]}
      end

    sql =
      "SELECT " <>
        select_columns() <>
        " FROM nodes WHERE " <> Enum.join(clauses, " AND ") <> " ORDER BY slug"

    case Store.raw_query(sql, params) do
      {:ok, rows} -> {:ok, Enum.map(rows, &row_to_struct/1)}
      other -> other
    end
  end

  @doc "Returns the direct children of a node (one level down)."
  @spec children(String.t(), keyword()) :: {:ok, [t()]}
  def children(node_id, opts \\ []) when is_binary(node_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())

    case Store.raw_query(
           "SELECT " <>
             select_columns() <>
             " FROM nodes WHERE parent_id = ?1 AND tenant_id = ?2 ORDER BY slug",
           [node_id, tenant_id]
         ) do
      {:ok, rows} -> {:ok, Enum.map(rows, &row_to_struct/1)}
      other -> other
    end
  end

  @doc "Walks up the parent chain. Returns ancestors ordered root→self."
  @spec ancestors(String.t(), keyword()) :: {:ok, [t()]}
  def ancestors(node_id, opts \\ []) when is_binary(node_id) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    do_ancestors(node_id, tenant_id, [])
  end

  # Ancestors are built by walking parent pointers up from the starting node.
  # Each recursive call prepends the newly-seen node, so the accumulator
  # naturally ends up in root→self order. Do NOT reverse at the end.
  defp do_ancestors(nil, _tenant_id, acc), do: {:ok, acc}

  defp do_ancestors(id, tenant_id, acc) do
    case get(id, tenant_id) do
      {:ok, node} -> do_ancestors(node.parent_id, tenant_id, [node | acc])
      {:error, :not_found} -> {:ok, acc}
    end
  end

  # ─── private ─────────────────────────────────────────────────────────────

  defp select_columns,
    do:
      "id, tenant_id, slug, name, kind, parent_id, description, style, status, path, metadata, created_at"

  defp row_to_struct([
         id,
         tenant_id,
         slug,
         name,
         kind,
         parent_id,
         description,
         style,
         status,
         path,
         metadata,
         created_at
       ]) do
    %__MODULE__{
      id: id,
      tenant_id: tenant_id,
      slug: slug,
      name: name,
      kind: safe_atom(kind, :unit),
      parent_id: parent_id,
      description: description,
      style: safe_atom(style, :internal),
      status: safe_atom(status, :active),
      path: path,
      metadata: decode_json(metadata),
      created_at: created_at
    }
  end

  defp safe_atom(nil, fallback), do: fallback

  defp safe_atom(str, fallback) when is_binary(str) do
    try do
      String.to_existing_atom(str)
    rescue
      ArgumentError -> fallback
    end
  end

  defp decode_json(nil), do: %{}
  defp decode_json(""), do: %{}

  defp decode_json(s) when is_binary(s) do
    case Jason.decode(s) do
      {:ok, m} -> m
      _ -> %{}
    end
  end
end
