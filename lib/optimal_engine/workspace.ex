defmodule OptimalEngine.Workspace do
  @moduledoc """
  A **workspace** is a knowledge base inside an organization (tenant).

  An organization can hold many workspaces — e.g. "Engineering Brain",
  "Sales Brain", "M&A Brain", "Personal". Each workspace has its own
  curated wiki, its own connectors, its own audiences, and its own
  set of nodes (the tenant's organizational topology can be re-used
  across workspaces or scoped per-workspace via `node.workspace_id`).

  Multiplicity rules:

  - Every signal-bearing row carries a `workspace_id`.
  - Default deployments use the singleton workspace `"default"`
    seeded by migration 026.
  - A principal in a tenant can belong to multiple workspaces via
    `workspace_members` with role ∈ `:owner | :member | :viewer`.

  Tenant isolation is the absolute boundary; workspace is a soft scope
  inside it. There's no cross-workspace leakage by default — every
  query that scopes by workspace must pass `workspace_id`.

  This module is the facade over the `workspaces` and `workspace_members`
  tables created in migration 026.
  """

  alias OptimalEngine.Store
  alias OptimalEngine.Tenancy.Tenant
  alias OptimalEngine.Workspace.Config
  alias OptimalEngine.Workspace.Filesystem

  @default_id "default"
  @allowed_statuses [:active, :archived]
  @allowed_roles [:owner, :member, :viewer]

  @type status :: :active | :archived
  @type role :: :owner | :member | :viewer

  @type t :: %__MODULE__{
          id: String.t(),
          tenant_id: String.t(),
          slug: String.t(),
          name: String.t(),
          description: String.t() | nil,
          status: status(),
          created_at: String.t() | nil,
          archived_at: String.t() | nil,
          metadata: map()
        }

  defstruct id: nil,
            tenant_id: Tenant.default_id(),
            slug: nil,
            name: nil,
            description: nil,
            status: :active,
            created_at: nil,
            archived_at: nil,
            metadata: %{}

  # ── Defaults ────────────────────────────────────────────────────────────

  @doc "Returns the reserved default workspace id."
  @spec default_id() :: String.t()
  def default_id, do: @default_id

  # ── Lookup ──────────────────────────────────────────────────────────────

  @doc "Returns the workspace with the given id, or `{:error, :not_found}`."
  @spec get(String.t()) :: {:ok, t()} | {:error, :not_found}
  def get(id) when is_binary(id) do
    case Store.raw_query(
           "SELECT id, tenant_id, slug, name, description, status, created_at, archived_at, metadata FROM workspaces WHERE id = ?1",
           [id]
         ) do
      {:ok, [row]} -> {:ok, row_to_struct(row)}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc "Returns the workspace identified by `(tenant_id, slug)`."
  @spec get_by_slug(String.t(), String.t()) :: {:ok, t()} | {:error, :not_found}
  def get_by_slug(slug, tenant_id) when is_binary(slug) and is_binary(tenant_id) do
    case Store.raw_query(
           "SELECT id, tenant_id, slug, name, description, status, created_at, archived_at, metadata FROM workspaces WHERE tenant_id = ?1 AND slug = ?2",
           [tenant_id, slug]
         ) do
      {:ok, [row]} -> {:ok, row_to_struct(row)}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc """
  Lists workspaces in a tenant. Options:
    * `:status` — `:active` (default) | `:archived` | `:all`
    * `:tenant_id` — defaults to the default tenant
  """
  @spec list(keyword()) :: {:ok, [t()]} | {:error, term()}
  def list(opts \\ []) do
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())
    status = Keyword.get(opts, :status, :active)

    {sql, params} =
      case status do
        :all ->
          {"SELECT id, tenant_id, slug, name, description, status, created_at, archived_at, metadata FROM workspaces WHERE tenant_id = ?1 ORDER BY name",
           [tenant_id]}

        s when s in @allowed_statuses ->
          {"SELECT id, tenant_id, slug, name, description, status, created_at, archived_at, metadata FROM workspaces WHERE tenant_id = ?1 AND status = ?2 ORDER BY name",
           [tenant_id, Atom.to_string(s)]}

        other ->
          throw({:invalid_status, other})
      end

    case Store.raw_query(sql, params) do
      {:ok, rows} -> {:ok, Enum.map(rows, &row_to_struct/1)}
      other -> other
    end
  catch
    {:invalid_status, s} -> {:error, {:invalid_status, s}}
  end

  # ── Mutations ───────────────────────────────────────────────────────────

  @doc """
  Creates a workspace. Required: `slug`, `name`. Optional: `tenant_id`
  (defaults to default tenant), `description`, `metadata`. Slug becomes
  part of the id (`tenant_id:slug` for non-default tenants, plain
  `default` for the singleton default-tenant default-workspace case).
  """
  @spec create(map()) :: {:ok, t()} | {:error, term()}
  def create(%{slug: slug, name: name} = attrs) when is_binary(slug) and is_binary(name) do
    tenant_id = Map.get(attrs, :tenant_id, Tenant.default_id())
    description = Map.get(attrs, :description)
    metadata = Map.get(attrs, :metadata, %{})
    id = Map.get(attrs, :id) || derive_id(tenant_id, slug)

    sql = """
    INSERT INTO workspaces (id, tenant_id, slug, name, description, status, metadata)
    VALUES (?1, ?2, ?3, ?4, ?5, 'active', ?6)
    """

    params = [id, tenant_id, slug, name, description, Jason.encode!(metadata)]

    case Store.raw_query(sql, params) do
      {:ok, _} ->
        # Provision the on-disk directory tree (nodes/, .wiki/, assets/, ...).
        # Failure is non-fatal — the row exists; the user can re-trigger
        # provisioning via the API. Logged but doesn't roll back the row.
        case Filesystem.provision(slug) do
          {:ok, _path} ->
            # Write default config.yaml into .optimal/ after FS is ready.
            case Config.put(slug, Config.defaults()) do
              :ok ->
                :ok

              {:error, reason} ->
                require Logger
                Logger.warning("[Workspace.create] config write failed for #{slug}: #{inspect(reason)}")
            end

          {:error, reason} ->
            require Logger
            Logger.warning("[Workspace.create] FS provision failed for #{slug}: #{inspect(reason)}")
        end

        {:ok,
         %__MODULE__{
           id: id,
           tenant_id: tenant_id,
           slug: slug,
           name: name,
           description: description,
           status: :active,
           metadata: metadata
         }}

      other ->
        other
    end
  end

  def create(_), do: {:error, :missing_required_fields}

  @doc """
  Updates a workspace's mutable fields: `:name`, `:description`,
  `:metadata`. Slug + tenant + id are immutable.
  """
  @spec update(String.t(), map()) :: {:ok, t()} | {:error, term()}
  def update(id, attrs) when is_binary(id) and is_map(attrs) do
    with {:ok, current} <- get(id) do
      name = Map.get(attrs, :name, current.name)
      description = Map.get(attrs, :description, current.description)
      metadata = Map.get(attrs, :metadata, current.metadata)

      case Store.raw_query(
             "UPDATE workspaces SET name = ?1, description = ?2, metadata = ?3 WHERE id = ?4",
             [name, description, Jason.encode!(metadata), id]
           ) do
        {:ok, _} ->
          {:ok, %{current | name: name, description: description, metadata: metadata}}

        other ->
          other
      end
    end
  end

  @doc "Soft-deletes a workspace by setting `status = 'archived'`."
  @spec archive(String.t()) :: :ok | {:error, term()}
  def archive(id) when is_binary(id) do
    case Store.raw_query(
           "UPDATE workspaces SET status = 'archived', archived_at = datetime('now') WHERE id = ?1",
           [id]
         ) do
      {:ok, _} -> :ok
      other -> other
    end
  end

  # ── Membership ──────────────────────────────────────────────────────────

  @doc """
  Grants a principal access to a workspace at a role. Role defaults to
  `:member`.
  """
  @spec add_member(String.t(), String.t(), keyword()) :: {:ok, map()} | {:error, term()}
  def add_member(workspace_id, principal_id, opts \\ []) do
    role = Keyword.get(opts, :role, :member)
    tenant_id = Keyword.get(opts, :tenant_id, Tenant.default_id())

    if role not in @allowed_roles do
      {:error, {:invalid_role, role}}
    else
      sql = """
      INSERT INTO workspace_members (tenant_id, workspace_id, principal_id, role)
      VALUES (?1, ?2, ?3, ?4)
      ON CONFLICT(workspace_id, principal_id) DO UPDATE SET
        role = excluded.role,
        ended_at = NULL
      """

      case Store.raw_query(sql, [tenant_id, workspace_id, principal_id, Atom.to_string(role)]) do
        {:ok, _} ->
          {:ok,
           %{
             tenant_id: tenant_id,
             workspace_id: workspace_id,
             principal_id: principal_id,
             role: role
           }}

        other ->
          other
      end
    end
  end

  @doc "Removes a principal from a workspace (sets ended_at)."
  @spec remove_member(String.t(), String.t()) :: :ok | {:error, term()}
  def remove_member(workspace_id, principal_id) do
    case Store.raw_query(
           "UPDATE workspace_members SET ended_at = datetime('now') WHERE workspace_id = ?1 AND principal_id = ?2",
           [workspace_id, principal_id]
         ) do
      {:ok, _} -> :ok
      other -> other
    end
  end

  @doc "Lists active members of a workspace as `{principal_id, role}` tuples."
  @spec members_of(String.t()) :: {:ok, [{String.t(), atom()}]} | {:error, term()}
  def members_of(workspace_id) when is_binary(workspace_id) do
    case Store.raw_query(
           "SELECT principal_id, role FROM workspace_members WHERE workspace_id = ?1 AND ended_at IS NULL ORDER BY role, principal_id",
           [workspace_id]
         ) do
      {:ok, rows} ->
        {:ok, Enum.map(rows, fn [pid, role] -> {pid, String.to_atom(role)} end)}

      other ->
        other
    end
  end

  @doc """
  Lists the workspaces a principal can access (active, not-ended) as
  `{workspace, role}` tuples, ordered by name.
  """
  @spec workspaces_of(String.t()) :: {:ok, [{t(), atom()}]} | {:error, term()}
  def workspaces_of(principal_id) when is_binary(principal_id) do
    sql = """
    SELECT w.id, w.tenant_id, w.slug, w.name, w.description, w.status,
           w.created_at, w.archived_at, w.metadata, m.role
    FROM workspaces w
    INNER JOIN workspace_members m ON m.workspace_id = w.id
    WHERE m.principal_id = ?1 AND m.ended_at IS NULL AND w.status = 'active'
    ORDER BY w.name
    """

    case Store.raw_query(sql, [principal_id]) do
      {:ok, rows} ->
        {:ok,
         Enum.map(rows, fn row ->
           {ws_row, [role]} = Enum.split(row, 9)
           {row_to_struct(ws_row), String.to_atom(role)}
         end)}

      other ->
        other
    end
  end

  # ── Helpers ─────────────────────────────────────────────────────────────

  # The default-tenant default-workspace gets the bare id "default" for
  # backwards compat with rows backfilled in migration 026. Everything
  # else gets `<tenant_id>:<slug>`.
  defp derive_id(tenant_id, slug) do
    if tenant_id == Tenant.default_id() and slug == @default_id,
      do: @default_id,
      else: "#{tenant_id}:#{slug}"
  end

  defp row_to_struct([id, tenant_id, slug, name, description, status, created_at, archived_at, meta_json]) do
    %__MODULE__{
      id: id,
      tenant_id: tenant_id,
      slug: slug,
      name: name,
      description: description,
      status: String.to_atom(status),
      created_at: created_at,
      archived_at: archived_at,
      metadata: decode_json(meta_json)
    }
  end

  defp decode_json(nil), do: %{}
  defp decode_json(""), do: %{}

  defp decode_json(s) when is_binary(s) do
    case Jason.decode(s) do
      {:ok, m} when is_map(m) -> m
      _ -> %{}
    end
  end
end
