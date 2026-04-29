defmodule OptimalEngine.Pipeline.Indexer do
  @moduledoc """
  Crawls the OptimalOS directory, classifies content, and persists Contexts.

  ## File type support

  - `.md`                        → classify as :signal (org folders) or :resource (docs/)
  - `.py .ex .exs .js .ts` etc   → classify as :resource with code mode
  - `.json .yaml .yml .toml`     → classify as :resource with data mode
  - `.txt .rst .adoc`            → classify as :resource with document mode
  - `.pdf .docx` etc             → classify as :resource (binary — metadata only)
  - `_memories/` subtree         → classify as :memory
  - `_skills/` subtree           → classify as :skill
  - Other files                  → classify as :resource

  ## Indexing modes

  - **Full**: Rebuilds the entire index from scratch.
  - **Incremental**: Only indexes files newer than the last indexed mtime.

  The indexer is a GenServer so that index runs can be requested from Mix tasks
  or triggered by file-watch events without blocking callers. Long-running
  indexing happens in a linked Task so the GenServer mailbox stays responsive.

  Directory scan skips:
  - `.git/`, hidden directories (except `.system`, `_memories`, `_skills`)
  - Files larger than 1MB
  - Binary files (metadata-only indexing instead)
  """

  use GenServer
  require Logger

  alias OptimalEngine.Pipeline.Classifier, as: Classifier
  alias OptimalEngine.Context
  alias OptimalEngine.Pipeline.Router, as: Router
  alias OptimalEngine.Pipeline.SemanticProcessor, as: SemanticProcessor
  alias OptimalEngine.Signal
  alias OptimalEngine.Store
  alias OptimalEngine.Topology
  alias OptimalEngine.URI
  alias OptimalEngine.Bridge.Knowledge, as: BridgeKnowledge
  alias OptimalEngine.Bridge.Memory, as: BridgeMemory

  @skip_dirs ~w[.git .claude node_modules _ARCHIVE .system engine tasks tools]
  @sidecar_suffixes [".abstract.md", ".overview.md"]
  @max_file_bytes 1_024 * 1_024

  # Extensions to index with full content
  @text_extensions ~w[
    .md .txt .rst .adoc
    .ex .exs .py .js .ts .jsx .tsx .go .rs .rb .java .c .cpp .h .sh .bash .zsh
    .json .yaml .yml .toml .csv .xml
  ]

  # Binary extensions — index with metadata only (no content read)
  @binary_extensions ~w[.pdf .docx .pptx .xlsx .png .jpg .jpeg .gif .svg .mp4 .mp3]

  # ---------------------------------------------------------------------------
  # Client API
  # ---------------------------------------------------------------------------

  @spec start_link(keyword()) :: GenServer.on_start()
  def start_link(opts \\ []) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @doc """
  Starts a full index run asynchronously.
  Returns `{:ok, :started}` immediately.
  """
  @spec full_index() :: {:ok, :started} | {:error, :already_running}
  def full_index do
    GenServer.call(__MODULE__, :full_index)
  end

  @doc """
  Indexes a single file immediately (synchronous).
  Useful for CLI tools like `mix optimal.ingest`.
  """
  @spec index_file(String.t()) :: {:ok, Context.t()} | {:error, term()}
  def index_file(path) when is_binary(path) do
    GenServer.call(__MODULE__, {:index_file, path}, 10_000)
  end

  @doc "Returns the current indexer status."
  @spec status() :: map()
  def status do
    GenServer.call(__MODULE__, :status)
  end

  # ---------------------------------------------------------------------------
  # GenServer Callbacks
  # ---------------------------------------------------------------------------

  @impl true
  def init(_opts) do
    topology =
      case Topology.load() do
        {:ok, t} ->
          t

        {:error, reason} ->
          Logger.warning("[Indexer] Could not load topology: #{inspect(reason)}")
          %{root_path: Application.get_env(:optimal_engine, :root_path), endpoints: %{}}
      end

    {:ok,
     %{
       topology: topology,
       status: :idle,
       last_run: nil,
       indexed_count: 0,
       task_ref: nil
     }}
  end

  @impl true
  def handle_call(:full_index, _from, %{status: :running} = state) do
    {:reply, {:error, :already_running}, state}
  end

  @impl true
  def handle_call(:full_index, _from, state) do
    known_entities = known_entity_names(state.topology)
    root = state.topology.root_path || Application.get_env(:optimal_engine, :root_path)

    task =
      Task.async(fn ->
        run_full_index(root, known_entities)
      end)

    {:reply, {:ok, :started},
     %{state | status: :running, task_ref: task.ref, last_run: DateTime.utc_now()}}
  end

  @impl true
  def handle_call({:index_file, path}, _from, state) do
    known_entities = known_entity_names(state.topology)
    result = index_single_file(path, known_entities)
    {:reply, result, state}
  end

  @impl true
  def handle_call(:status, _from, state) do
    info = %{
      status: state.status,
      last_run: state.last_run,
      indexed_count: state.indexed_count
    }

    {:reply, info, state}
  end

  @impl true
  def handle_info({ref, {:ok, count}}, %{task_ref: ref} = state) do
    Process.demonitor(ref, [:flush])
    Logger.info("[Indexer] Full index complete. Indexed #{count} files.")

    :telemetry.execute(
      [:optimal_engine, :indexer, :complete],
      %{count: count},
      %{}
    )

    # Post-index: OWL materialization + SICA pattern analysis (async)
    Task.start(fn -> post_index_analysis(count) end)

    {:noreply, %{state | status: :idle, indexed_count: count, task_ref: nil}}
  end

  @impl true
  def handle_info({ref, {:error, reason}}, %{task_ref: ref} = state) do
    Process.demonitor(ref, [:flush])
    Logger.error("[Indexer] Full index failed: #{inspect(reason)}")
    {:noreply, %{state | status: :error, task_ref: nil}}
  end

  @impl true
  def handle_info({:DOWN, ref, :process, _pid, reason}, %{task_ref: ref} = state) do
    Logger.error("[Indexer] Index task crashed: #{inspect(reason)}")
    {:noreply, %{state | status: :error, task_ref: nil}}
  end

  @impl true
  def handle_info(_msg, state), do: {:noreply, state}

  # ---------------------------------------------------------------------------
  # Private: Indexing Pipeline
  # ---------------------------------------------------------------------------

  defp run_full_index(root, known_entities) do
    Logger.info("[Indexer] Starting full index of #{root}")

    paths = collect_indexable_files(root)
    total = length(paths)
    Logger.info("[Indexer] Found #{total} files to index")

    results =
      paths
      |> Enum.chunk_every(50)
      |> Enum.flat_map(fn batch ->
        contexts =
          batch
          |> Enum.map(&build_context(&1, known_entities))
          |> Enum.reject(&is_nil/1)
          |> Enum.map(&maybe_semantic_process/1)

        case Store.insert_contexts(contexts) do
          {:ok, n} ->
            Logger.debug("[Indexer] Batch of #{n} contexts inserted")
            contexts

          {:error, reason} ->
            Logger.warning("[Indexer] Batch insert failed: #{inspect(reason)}")
            []
        end
      end)

    {:ok, length(results)}
  end

  defp index_single_file(path, known_entities) do
    with ctx when not is_nil(ctx) <- build_context(path, known_entities),
         :ok <- Store.insert_context(ctx) do
      {:ok, ctx}
    else
      nil -> {:error, :could_not_build_context}
      {:error, _} = err -> err
    end
  end

  defp build_context(path, known_entities) do
    ext = Path.extname(path)

    if ext in @binary_extensions do
      build_binary_context(path)
    else
      build_text_context(path, known_entities)
    end
  end

  defp build_text_context(path, known_entities) do
    case read_file_safe(path) do
      {:ok, content} ->
        stat = File.stat!(path, time: :posix)

        classified =
          Classifier.classify_context(content,
            path: path,
            known_entities: known_entities
          )

        # Override node from directory structure if classifier defaulted to inbox
        node =
          if classified.node == "inbox" do
            node_from_path(path)
          else
            classified.node
          end

        uri = URI.from_path(path)
        ctx_with_meta = %{classified | node: node, path: path, uri: uri}

        # Route signals
        routed_to =
          if classified.type == :signal && classified.signal do
            route_signal(%{classified.signal | node: node, path: path})
          else
            [node]
          end

        # Update the embedded signal's routing if present
        updated_signal =
          if classified.signal do
            %{classified.signal | node: node, path: path, routed_to: routed_to}
          else
            nil
          end

        %Context{
          ctx_with_meta
          | id: context_id(path),
            path: path,
            uri: uri,
            node: node,
            signal: updated_signal,
            routed_to: routed_to,
            created_at: posix_to_datetime(stat.ctime),
            modified_at: posix_to_datetime(stat.mtime)
        }

      {:error, reason} ->
        Logger.debug("[Indexer] Skipping #{path}: #{inspect(reason)}")
        nil
    end
  end

  defp build_binary_context(path) do
    case File.stat(path) do
      {:ok, stat} when stat.size > 0 and stat.size <= @max_file_bytes ->
        ext = Path.extname(path)
        filename = Path.basename(path)
        title = filename |> String.replace(~r/[-_]/, " ") |> String.trim()
        uri = URI.from_path(path)
        node = node_from_path(path)

        %Context{
          id: context_id(path),
          uri: uri,
          type: :resource,
          path: path,
          title: title,
          content: "",
          l0_abstract: "RESOURCE | binary | #{title}",
          l1_overview: "Binary file: #{filename} (#{ext}, #{stat.size} bytes)",
          signal: nil,
          node: node,
          sn_ratio: 0.3,
          entities: [],
          created_at: posix_to_datetime(stat.ctime),
          modified_at: posix_to_datetime(stat.mtime),
          routed_to: [node],
          metadata: %{
            "extension" => ext,
            "filename" => filename,
            "size_bytes" => stat.size,
            "format" => "binary"
          }
        }

      _ ->
        nil
    end
  end

  defp route_signal(%Signal{} = signal) do
    case Router.route(signal) do
      {:ok, destinations} -> destinations
      {:error, _} -> [signal.node]
    end
  end

  # ---------------------------------------------------------------------------
  # Private: File collection
  # ---------------------------------------------------------------------------

  defp collect_indexable_files(root) do
    collect_dir(root, [])
  end

  defp collect_dir(path, acc) do
    case File.ls(path) do
      {:ok, entries} ->
        Enum.reduce(entries, acc, fn entry, inner_acc ->
          full = Path.join(path, entry)
          process_entry(full, entry, inner_acc)
        end)

      {:error, _} ->
        acc
    end
  end

  defp process_entry(full_path, entry, acc) do
    cond do
      skip_entry?(entry) -> acc
      File.dir?(full_path) -> collect_dir(full_path, acc)
      indexable_extension?(entry) -> maybe_add_file(full_path, acc)
      true -> acc
    end
  end

  defp skip_entry?(entry) do
    entry in @skip_dirs or
      (String.starts_with?(entry, ".") and entry != ".system") or
      (String.starts_with?(entry, "_") and entry not in ["_memories", "_skills"]) or
      sidecar_file?(entry)
  end

  defp sidecar_file?(filename) do
    Enum.any?(@sidecar_suffixes, &String.ends_with?(filename, &1))
  end

  defp maybe_add_file(full_path, acc) do
    case File.stat(full_path) do
      {:ok, %{size: size}} when size > 0 and size <= @max_file_bytes -> [full_path | acc]
      _ -> acc
    end
  end

  defp indexable_extension?(filename) do
    ext = Path.extname(filename)
    ext in @text_extensions or ext in @binary_extensions
  end

  # ---------------------------------------------------------------------------
  # Private: Helpers
  # ---------------------------------------------------------------------------

  defp read_file_safe(path) do
    case File.read(path) do
      {:ok, content} when byte_size(content) > 0 -> {:ok, content}
      {:ok, _} -> {:error, :empty_file}
      {:error, _} = err -> err
    end
  end

  defp node_from_path(path) do
    root = Application.get_env(:optimal_engine, :root_path, "")

    relative =
      path
      |> String.replace_prefix(root <> "/", "")
      |> String.split("/")
      |> List.first("")

    folder_to_node(relative)
  end

  defp folder_to_node("01-roberto"), do: "roberto"
  defp folder_to_node("02-miosa"), do: "miosa-platform"
  defp folder_to_node("03-lunivate"), do: "lunivate"
  defp folder_to_node("04-ai-masters"), do: "ai-masters"
  defp folder_to_node("05-os-architect"), do: "os-architect"
  defp folder_to_node("06-agency-accelerants"), do: "agency-accelerants"
  defp folder_to_node("07-accelerants-community"), do: "accelerants-community"
  defp folder_to_node("08-content-creators"), do: "content-creators"
  defp folder_to_node("09-new-stuff"), do: "inbox"
  defp folder_to_node("10-team"), do: "team"
  defp folder_to_node("11-money-revenue"), do: "money-revenue"
  defp folder_to_node("12-os-accelerator"), do: "os-accelerator"
  defp folder_to_node("docs"), do: "resources"
  defp folder_to_node("_memories"), do: "inbox"
  defp folder_to_node("_skills"), do: "inbox"
  defp folder_to_node(_), do: "inbox"

  defp context_id(path) do
    :crypto.hash(:sha256, path)
    |> Base.encode16(case: :lower)
    |> String.slice(0, 32)
  end

  defp posix_to_datetime(posix) when is_integer(posix), do: DateTime.from_unix!(posix)
  defp posix_to_datetime(_), do: DateTime.utc_now()

  # Post-index analysis: OWL reasoning + SICA pattern surfacing
  defp post_index_analysis(indexed_count) do
    # 1. OWL Materialization — sync edges to OptimalEngine.Knowledge and reason
    Logger.info("[Indexer] Starting post-index OWL materialization...")

    case BridgeKnowledge.sync_and_materialize() do
      {:ok, inferred} ->
        Logger.info("[Indexer] OWL materialization complete: #{inspect(inferred)} inferred triples")

      {:error, reason} ->
        Logger.warning("[Indexer] OWL materialization failed: #{inspect(reason)}")
    end

    # 2. SICA pattern analysis — surface any detected patterns
    metrics = BridgeMemory.learning_metrics()
    pattern_count = Map.get(metrics, :patterns, 0)
    solution_count = Map.get(metrics, :skills, 0)

    Logger.info(
      "[Indexer/SICA] After indexing #{indexed_count} files: #{pattern_count} patterns, #{solution_count} solutions"
    )

    if pattern_count > 0 do
      patterns = BridgeMemory.patterns()

      Enum.each(patterns, fn {name, data} ->
        Logger.info("[Indexer/SICA] Pattern: #{name} — #{inspect(data)}")
      end)
    end

    # 3. Record the index event in SICA
    BridgeMemory.observe_mutation(%{
      tool: "indexer",
      input: "full_index",
      output: "indexed #{indexed_count} files",
      success: true
    })
  rescue
    e ->
      Logger.warning("[Indexer] Post-index analysis failed: #{inspect(e)}")
  end

  # Run semantic processing (LLM summaries + embedding) if Ollama available and no fresh sidecar
  defp maybe_semantic_process(%Context{} = ctx) do
    if ctx.path && !SemanticProcessor.sidecar_fresh?(ctx.path) do
      case SemanticProcessor.process(ctx) do
        {:ok, enhanced} -> enhanced
        _ -> ctx
      end
    else
      ctx
    end
  rescue
    _ -> ctx
  end

  defp known_entity_names(topology) do
    topology
    |> Map.get(:endpoints, %{})
    |> Enum.flat_map(fn {_id, %{name: name}} ->
      first = name |> String.split(" ") |> List.first()
      [name, first]
    end)
    |> Enum.uniq()
    |> Enum.reject(&(String.length(&1) < 3))
  end
end
