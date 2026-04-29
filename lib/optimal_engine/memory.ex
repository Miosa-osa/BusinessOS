defmodule OptimalEngine.Memory do
  @moduledoc """
  Main API for the OptimalEngine.Memory agent memory system.

  Provides persistence, context management, and memory injection
  following the MemoryOS pattern.

  ## Usage

      OptimalEngine.Memory.store("decisions", "arch-001", %{title: "Use ETS", reason: "Fast local access"})
      {:ok, entry} = OptimalEngine.Memory.recall("decisions", "arch-001")
      {:ok, matches} = OptimalEngine.Memory.search("decisions", "ETS")
  """

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

  ## Examples

      iex> OptimalEngine.Memory.store("test_store", "key1", "value1")
      :ok
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

  ## Examples

      iex> OptimalEngine.Memory.store("test_recall", "k", "v")
      iex> {:ok, entry} = OptimalEngine.Memory.recall("test_recall", "k")
      iex> entry.value
      "v"
  """
  @spec recall(collection(), key()) ::
          {:ok, OptimalEngine.Memory.Store.entry()} | {:error, :not_found}
  def recall(collection, key) do
    backend().get(collection, key)
  end

  @doc """
  Search memories in all collections by query string (keyword match).

  Returns `{:ok, matches}` with entries whose key, value, or tags match any query term.

  ## Examples

      iex> OptimalEngine.Memory.store("test_search", "elixir-tip", "Use pattern matching", tags: ["elixir"])
      iex> {:ok, matches} = OptimalEngine.Memory.search("test_search", "elixir")
      iex> length(matches) > 0
      true
  """
  @spec search(collection(), String.t()) :: {:ok, [OptimalEngine.Memory.Store.entry()]}
  def search(collection, query) do
    backend().search(collection, query)
  end

  @doc """
  Delete a memory entry.

  ## Examples

      iex> OptimalEngine.Memory.store("test_forget", "tmp", "data")
      iex> OptimalEngine.Memory.forget("test_forget", "tmp")
      :ok
      iex> OptimalEngine.Memory.recall("test_forget", "tmp")
      {:error, :not_found}
  """
  @spec forget(collection(), key()) :: :ok
  def forget(collection, key) do
    backend().delete(collection, key)
  end

  @doc """
  List all memory collections.

  ## Examples

      iex> OptimalEngine.Memory.store("test_collections_list", "k", "v")
      iex> {:ok, cols} = OptimalEngine.Memory.collections()
      iex> "test_collections_list" in cols
      true
  """
  @spec collections() :: {:ok, [collection()]}
  def collections do
    backend().collections()
  end

  @doc """
  Export a collection to a JSON file at the given path.

  Returns `:ok` on success.
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
end
