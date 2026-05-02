defmodule OptimalEngine.Memory do
  @moduledoc """
  Unified memory API for OptimalEngine.

  This module provides two complementary memory APIs:

  ## 1. Versioned memory (primary — SQL-backed)

  First-class, workspace-scoped, versioned memory with typed relations and
  soft forgetting. Backed by the `memories` and `memory_relations` tables
  (migration 028). Delegates to `OptimalEngine.Memory.Versioned`.

  Key operations: `create/1`, `get/1`, `list/1`, `update/2`, `extend/2`,
  `derive/2`, `forget/2`, `versions/1`, `relations/1`, `delete/1`.

  ## 2. Key-value memory (legacy — ETS-backed)

  Simple collection/key/value store used by the agent memory system,
  episodic memory, and session management. Backed by ETS via
  `OptimalEngine.Memory.Store.ETS`.

  Key operations: `store/4`, `recall/2`, `search/2`, `forget/3`,
  `collections/0`, `export/2`, `import_collection/2`.

  The two APIs coexist without conflict. `forget/2` dispatches based on
  argument types: `forget(id, keyword_opts)` → versioned; the legacy
  path uses the explicit `forget_key/2` alias.
  """

  alias OptimalEngine.Memory.Versioned

  require Logger

  # ===========================================================================
  # Versioned memory API (primary)
  # ===========================================================================

  @type versioned :: OptimalEngine.Memory.Versioned.t()

  @doc """
  Creates a new versioned memory (v1, is_latest=true).

  Required: `:content` (non-empty string)
  Optional: `:workspace_id`, `:tenant_id`, `:is_static`, `:audience`,
            `:citation_uri`, `:source_chunk_id`, `:metadata`

  When `is_static: true` and the workspace config has
  `memory.auto_promote_to_wiki: true`, the memory is asynchronously
  promoted to a wiki page (fire-and-forget via `Task.start`).
  """
  @spec create(map()) :: {:ok, versioned()} | {:error, term()}
  def create(attrs) do
    case Versioned.create(attrs) do
      {:ok, mem} = ok ->
        maybe_promote_to_wiki(mem)
        ok

      other ->
        other
    end
  end

  @doc "Fetches a versioned memory by id."
  @spec get(String.t()) :: {:ok, versioned()} | {:error, :not_found}
  defdelegate get(id), to: Versioned

  @doc """
  Lists versioned memories.

  Returns a plain list (not `{:ok, list}`) for ergonomic router use.

  Options:
    - `:workspace_id` — defaults to "default"
    - `:audience` — string filter
    - `:include_forgotten` — default false
    - `:include_old_versions` — default false
    - `:limit` — default 50
  """
  @spec list(keyword()) :: [versioned()]
  def list(opts \\ []) do
    case Versioned.list(opts) do
      {:ok, mems} -> mems
      _ -> []
    end
  end

  @doc """
  Counts versioned memories matching the same filter opts as `list/1`.
  Returns `{:ok, integer()}` so callers can distinguish 0 from error.
  """
  @spec count(keyword()) :: {:ok, non_neg_integer()} | {:error, term()}
  def count(opts \\ []), do: Versioned.count(opts)

  @doc """
  Creates a new version of a memory (bumps version, demotes old to
  is_latest=false, adds `:updates` relation).
  """
  @spec update(String.t(), map()) :: {:ok, versioned()} | {:error, term()}
  defdelegate update(id, attrs), to: Versioned

  @doc """
  Creates a child memory with `:extends` relation. Source keeps is_latest.
  """
  @spec extend(String.t(), map()) :: {:ok, versioned()} | {:error, term()}
  defdelegate extend(id, attrs), to: Versioned

  @doc """
  Creates a derived memory with `:derives` relation. Source keeps is_latest.
  """
  @spec derive(String.t(), map()) :: {:ok, versioned()} | {:error, term()}
  defdelegate derive(id, attrs), to: Versioned

  @doc """
  Soft-forgets a versioned memory (sets is_forgotten=1).

  Options:
    - `:reason` — stored in `forget_reason`
    - `:forget_after` — ISO8601 timestamp

  Dispatches on the second argument type:
    - `forget(id, keyword_list)` — versioned memory soft-forget
    - `forget(id, binary)` — alias for legacy `forget_key/2`
  """
  @spec forget(String.t(), keyword() | String.t()) :: :ok | {:error, term()}
  def forget(id, opts \\ [])

  # Versioned soft-forget: forget(memory_id, keyword_opts)
  def forget(id, opts) when is_binary(id) and is_list(opts) do
    Versioned.forget(id, opts)
  end

  # Legacy ETS delete: forget(collection, key) — both strings
  def forget(collection, key) when is_binary(collection) and is_binary(key) do
    backend().delete(collection, key)
  end

  @doc "Returns the full version chain ordered v1 → ... → latest."
  @spec versions(String.t()) :: {:ok, [versioned()]} | {:error, term()}
  defdelegate versions(id), to: Versioned

  @doc """
  Returns all typed relations touching `memory_id` (inbound + outbound).

  Each relation map includes:
    - `:id`, `:source_memory_id`, `:target_memory_id`, `:target_id` (alias),
      `:relation`, `:direction`, `:created_at`
  """
  @spec relations(String.t()) :: {:ok, [map()]} | {:error, term()}
  def relations(id) when is_binary(id) do
    case Versioned.relations(id) do
      {:ok, rels} ->
        # Add :target_id alias so the router's `r.target_id` resolves correctly.
        enriched = Enum.map(rels, &Map.put(&1, :target_id, &1.target_memory_id))
        {:ok, enriched}

      other ->
        other
    end
  end

  @doc "Hard deletes a versioned memory (cascades relations)."
  @spec delete(String.t()) :: :ok | {:error, term()}
  defdelegate delete(id), to: Versioned

  # ===========================================================================
  # Legacy key-value memory API (ETS-backed)
  # ===========================================================================

  @type collection :: String.t()
  @type key :: String.t()
  @type value :: term()

  @doc """
  Store a memory entry in a collection.

  ## Parameters
    - `collection` - the memory collection name (e.g., "decisions", "patterns")
    - `key` - unique key within the collection
    - `value` - any term to store

  ## Options
    - `:tags` - list of string tags for search (default: `[]`)
  """
  @spec store(collection(), key(), value(), keyword()) :: :ok | {:error, term()}
  def store(collection, key, value, opts \\ []) do
    tags = Keyword.get(opts, :tags, [])
    metadata = %{tags: tags}
    backend().put(collection, key, value, metadata)
  end

  @doc """
  Recall a memory entry by collection and key.

  Returns `{:ok, entry}` or `{:error, :not_found}`.
  """
  @spec recall(collection(), key()) ::
          {:ok, OptimalEngine.Memory.Store.entry()} | {:error, :not_found}
  def recall(collection, key) do
    backend().get(collection, key)
  end

  @doc """
  Search memories in a collection by query string (keyword match).

  Returns `{:ok, matches}` with entries whose key, value, or tags match
  any query term.
  """
  @spec search(collection(), String.t()) :: {:ok, [OptimalEngine.Memory.Store.entry()]}
  def search(collection, query) do
    backend().search(collection, query)
  end

  @doc """
  Delete a legacy key-value memory entry.

  Two-arg form: `forget(collection, key)` — legacy ETS delete.
  Single-arg / keyword form: `forget(id)` / `forget(id, opts)` — versioned.
  """
  @spec forget_key(collection(), key()) :: :ok
  def forget_key(collection, key) do
    backend().delete(collection, key)
  end

  @doc """
  List all memory collections.
  """
  @spec collections() :: {:ok, [collection()]}
  def collections do
    backend().collections()
  end

  @doc """
  Recalls all long-term memories as a single concatenated string.
  Used by the Bridge and Cortex for synthesis.
  """
  @spec recall() :: String.t()
  def recall do
    OptimalEngine.Memory.Store.recall()
  end

  @doc """
  Stores an insight into the default "memory" collection.
  Convenience wrapper over `store/4`.
  """
  @spec remember(String.t(), keyword()) :: :ok | {:error, term()}
  def remember(insight, opts \\ []) when is_binary(insight) do
    tags = Keyword.get(opts, :tags, [])
    key = Keyword.get(opts, :key, "insight_#{:erlang.unique_integer([:positive])}")
    store("memory", key, insight, tags: tags)
  end

  @doc """
  Export a collection to a JSON file at the given path.
  """
  @spec export(collection(), String.t()) :: :ok | {:error, term()}
  def export(collection, path) do
    {:ok, entries} = backend().list(collection, [])

    data =
      Enum.map(entries, fn entry ->
        entry
        |> put_in([:metadata, :created_at], DateTime.to_iso8601(entry.metadata.created_at))
        |> put_in([:metadata, :updated_at], DateTime.to_iso8601(entry.metadata.updated_at))
      end)

    dir = Path.dirname(path)
    File.mkdir_p!(dir)
    File.write!(path, Jason.encode!(data, pretty: true))
    :ok
  end

  @doc """
  Import a collection from a JSON file at the given path.
  Entries are merged into the specified collection.
  """
  @spec import_collection(collection(), String.t()) :: :ok | {:error, term()}
  def import_collection(collection, path) do
    case File.read(path) do
      {:ok, content} ->
        entries = Jason.decode!(content, keys: :atoms)

        Enum.each(entries, fn entry ->
          metadata =
            entry.metadata
            |> Map.update!(:created_at, &parse_dt/1)
            |> Map.update!(:updated_at, &parse_dt/1)

          backend().put(collection, entry.key, entry.value, metadata)
        end)

        :ok

      {:error, reason} ->
        {:error, reason}
    end
  end

  defp parse_dt(dt) when is_binary(dt) do
    {:ok, parsed, _} = DateTime.from_iso8601(dt)
    parsed
  end

  defp parse_dt(%DateTime{} = dt), do: dt

  defp backend, do: OptimalEngine.Memory.Store.backend()

  # ---------------------------------------------------------------------------
  # WikiBridge auto-promotion (Phase 17.1/7 bridge)
  # ---------------------------------------------------------------------------

  # Fire-and-forget: if the memory is static and the workspace config enables
  # auto_promote_to_wiki, promote it to the "memory-promotions" wiki page.
  # Any failure is logged and swallowed — memory create must never be blocked.
  defp maybe_promote_to_wiki(%Versioned{is_static: true} = mem) do
    alias OptimalEngine.Workspace.Config
    mem_cfg = Config.get_section(mem.workspace_id, :memory, %{auto_promote_to_wiki: false})

    if Map.get(mem_cfg, :auto_promote_to_wiki, false) do
      Task.start(fn ->
        opts = [
          workspace_id: mem.workspace_id,
          tenant_id: mem.tenant_id,
          audience: mem.audience
        ]

        slug = "memory-promotions"

        case OptimalEngine.Memory.WikiBridge.promote_memory_to_wiki(mem.id, slug, opts) do
          {:ok, _page} ->
            Logger.info("[Memory] auto-promoted #{mem.id} to wiki/#{slug}")

          {:error, reason} ->
            Logger.warning(
              "[Memory] WikiBridge.promote_memory_to_wiki failed for #{mem.id}: #{inspect(reason)}"
            )
        end
      end)
    end

    :ok
  rescue
    e ->
      Logger.warning("[Memory] maybe_promote_to_wiki error: #{inspect(e)}")
      :ok
  end

  defp maybe_promote_to_wiki(_mem), do: :ok
end
