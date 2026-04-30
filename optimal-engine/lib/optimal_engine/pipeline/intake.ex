defmodule OptimalEngine.Pipeline.Intake do
  @moduledoc """
  Intake pipeline — classifies raw text, routes it to nodes, writes signal
  files to disk, and updates the SQLite index.

  This is the core user-facing ingestion flow:

  1. Classify: run S=(M,G,T,F,W) classification via `Classifier`
  2. Route: determine primary node + cross-reference nodes via `Router`
  3. Write: emit structured markdown files to `{node}/signals/` directories
  4. Cross-ref: write cross-references to secondary destinations
  5. Index: store the Context in SQLite via `Store`

  Non-critical steps run async (fire-and-forget):
  - Episodic recording
  - SICA observation
  - Quality audit logging

  ## Quality Actions

  FailureModes violations now trigger actions:
  - S/N < 0.3 → reject with `{:error, :signal_too_noisy}`
  - S/N < 0.6 AND routing_failure → quarantine to inbox with `:low_quality` tag
  - :bandwidth_overload → truncate L1 to 250 chars
  - :structure_failure → auto-apply genre skeleton (already happens)

  ## Usage

      {:ok, result} = OptimalEngine.Pipeline.Intake.process("Customer called about AI Masters pricing")

      {:ok, result} = OptimalEngine.Pipeline.Intake.process(
        "Customer called...",
        genre: "transcript",
        title: "Q4 Pricing Call",
        node: "ai-masters",
        entities: ["Alice", "Alice"]
      )

  ## Result shape

      %{
        signal: %Signal{},
        context: %Context{},
        files_written: ["04-ai-masters/signals/2026-03-18-ed-pricing-call.md"],
        routed_to: ["04-ai-masters", "11-money-revenue"],
        cross_references: ["11-money-revenue/signals/2026-03-18-ed-pricing-call.md"],
        uri: "optimal://nodes/ai-masters/signals/2026-03-18-ed-pricing-call.md",
        quality_violations: [{:bandwidth_overload, "..."}],
        quality_action: :accepted
      }

  ## Options

  - `:genre`    — Override auto-detected genre (e.g. "transcript")
  - `:node`     — Override primary node routing (e.g. "ai-masters")
  - `:title`    — Explicit title instead of auto-extracted one
  - `:entities` — Explicit entity list (merged with auto-extracted)
  - `:type`     — Force context type (:signal, :resource, :memory, :skill)
  """

  use GenServer
  require Logger

  alias OptimalEngine.Pipeline.Classifier, as: Classifier
  alias OptimalEngine.Context
  alias OptimalEngine.Pipeline.Indexer, as: Indexer
  alias OptimalEngine.Pipeline.Intake.Writer, as: Writer
  alias OptimalEngine.Pipeline.Router, as: Router
  alias OptimalEngine.Pipeline.SemanticProcessor, as: SemanticProcessor
  alias OptimalEngine.Signal
  alias OptimalEngine.Store
  alias OptimalEngine.Routing
  alias OptimalEngine.URI

  alias OptimalEngine.Bridge.Memory, as: BridgeMemory
  alias OptimalEngine.Bridge.Signal, as: BridgeSignal

  # Result type for process/2
  @type result :: %{
          signal: Signal.t(),
          context: Context.t(),
          files_written: [String.t()],
          routed_to: [String.t()],
          cross_references: [String.t()],
          uri: String.t(),
          quality_violations: [{atom(), String.t()}],
          quality_action: :accepted | :quarantined | :rejected
        }

  # ---------------------------------------------------------------------------
  # Client API
  # ---------------------------------------------------------------------------

  @spec start_link(keyword()) :: GenServer.on_start()
  def start_link(opts \\ []) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @doc """
  Processes raw text through the full intake pipeline:
  classify → route → write files → index.

  Returns `{:ok, result}` or `{:error, reason}`.

  Options include all classification overrides plus:
  - `:workspace_id` — target workspace (default: "default")
  """
  @spec process(String.t(), keyword()) :: {:ok, result()} | {:error, term()}
  def process(raw_text, opts \\ []) when is_binary(raw_text) do
    GenServer.call(__MODULE__, {:process, raw_text, opts}, 30_000)
  end

  # ---------------------------------------------------------------------------
  # GenServer Callbacks
  # ---------------------------------------------------------------------------

  @impl true
  def init(_opts) do
    {:ok, %{}}
  end

  @impl true
  def handle_call({:process, raw_text, opts}, _from, state) do
    result = run_pipeline(raw_text, opts)
    {:reply, result, state}
  end

  # ---------------------------------------------------------------------------
  # Private: Pipeline
  # ---------------------------------------------------------------------------

  defp run_pipeline(raw_text, opts) do
    with {:ok, signal} <- classify_step(raw_text, opts),
         signal = enhance_with_miosa(signal, raw_text),
         {:ok, signal} <- quality_gate(signal),
         {:ok, signal, routed_to} <- route_step(signal, opts),
         {:ok, primary_path} <- write_step(signal, opts),
         {:ok, cross_paths} <- cross_ref_step(signal, routed_to),
         {:ok, context} <- index_step(signal, primary_path, opts) do
      uri = URI.from_path(primary_path)
      primary_relative = relative(primary_path)
      cross_relatives = Enum.map(cross_paths, &relative/1)

      # Semantic processing: LLM-enhanced L0/L1 + embedding (async, non-blocking)
      Task.start(fn -> semantic_enhance(context) end)

      # Quality audit — determine violations and action
      violations = BridgeSignal.audit(signal)
      quality_action = determine_quality_action(signal, violations)

      # Apply bandwidth fix if needed (truncate L1 for overloaded signals)
      signal = apply_bandwidth_fix(signal, violations)

      result = %{
        signal: signal,
        context: context,
        files_written: [primary_relative],
        routed_to: routed_to,
        cross_references: cross_relatives,
        uri: uri,
        quality_violations: violations,
        quality_action: quality_action
      }

      # Async: non-critical steps (episodic record, SICA observe, telemetry)
      Task.start(fn -> async_post_intake(signal, result, violations) end)

      :telemetry.execute(
        [:optimal_engine, :intake, :process],
        %{files_written: 1 + length(cross_paths)},
        %{node: signal.node, genre: signal.genre}
      )

      {:ok, result}
    end
  end

  # Quality gate — reject or quarantine signals based on S/N and failure modes
  defp quality_gate(%Signal{} = signal) do
    sn = signal.sn_ratio || 0.6

    cond do
      sn < 0.3 ->
        Logger.warning("[Intake] Rejecting signal — S/N #{sn} too low (< 0.3)")
        {:error, :signal_too_noisy}

      true ->
        {:ok, signal}
    end
  end

  # Determine quality action based on violations and S/N
  defp determine_quality_action(%Signal{} = signal, violations) do
    sn = signal.sn_ratio || 0.6
    violation_modes = Enum.map(violations, &elem(&1, 0))

    cond do
      sn < 0.6 and :routing_failure in violation_modes ->
        :quarantined

      true ->
        :accepted
    end
  end

  # Apply bandwidth overload fix to signal
  defp apply_bandwidth_fix(%Signal{} = signal, violations) do
    violation_modes = Enum.map(violations, &elem(&1, 0))

    if :bandwidth_overload in violation_modes do
      # Truncate L1 to 250 chars instead of 500
      l1 =
        (signal.l1_description || "")
        |> String.slice(0, 250)

      %{signal | l1_description: l1}
    else
      signal
    end
  end

  # Augment signal classification with OptimalEngine.Signal.Core's classifier
  defp enhance_with_miosa(signal, raw_text) do
    BridgeSignal.enhance_classification(signal, raw_text)
  rescue
    _ -> signal
  end

  # Async post-intake: episodic recording, SICA observation, quality logging
  defp async_post_intake(signal, result, violations) do
    # Episodic event
    BridgeMemory.record_event(:intake, %{
      genre: signal.genre,
      node: signal.node,
      title: signal.title,
      entities: signal.entities || [],
      routed_to: result.routed_to,
      quality_action: result.quality_action
    })

    # SICA observation
    BridgeMemory.observe_mutation(%{
      tool: "intake",
      input: signal.title || "",
      output: "routed to #{Enum.join(result.routed_to, ", ")}",
      success: true
    })

    # Log violations if any
    if violations != [] do
      Logger.debug(
        "[Intake] Signal quality violations: #{inspect(Enum.map(violations, &elem(&1, 0)))}"
      )
    end

    # Check SICA for new patterns and log them
    check_sica_patterns()
  rescue
    _ -> :ok
  end

  # Check if SICA has detected new patterns since last check
  defp check_sica_patterns do
    metrics = BridgeMemory.learning_metrics()
    pattern_count = Map.get(metrics, :patterns, 0)

    if pattern_count > 0 do
      patterns = BridgeMemory.patterns()

      Enum.each(patterns, fn {pattern_name, pattern_data} ->
        Logger.info("[Intake/SICA] Pattern detected: #{pattern_name} — #{inspect(pattern_data)}")
      end)
    end
  rescue
    _ -> :ok
  end

  # Step 1: Classify
  defp classify_step(raw_text, opts) do
    topology = load_topology()
    known_entities = topology_entities(topology)

    # Merge explicit entities into known entities for extraction
    extra_entities = Keyword.get(opts, :entities, [])
    all_known = Enum.uniq(known_entities ++ extra_entities)

    # Intake always produces a signal unless the caller explicitly forces another type.
    # Without this, content with an empty path is classified as :resource and loses
    # all S=(M,G,T,F,W) dimensions.
    forced_type = Keyword.get(opts, :type, :signal)
    classify_opts = [path: "", known_entities: all_known, type: forced_type]

    ctx = Classifier.classify_context(raw_text, classify_opts)

    signal =
      case ctx.signal do
        %Signal{} = sig ->
          sig

        nil ->
          # Build a minimal signal from the context for non-signal types
          Context.to_signal(ctx)
      end

    now = DateTime.utc_now()
    id = generate_id(raw_text, now)

    # Apply option overrides
    signal =
      signal
      |> maybe_override(:title, Keyword.get(opts, :title))
      |> maybe_override(:genre, Keyword.get(opts, :genre))
      |> maybe_override(:node, Keyword.get(opts, :node))
      |> maybe_merge_entities(extra_entities)
      |> then(fn s ->
        %{
          s
          | id: id,
            created_at: now,
            modified_at: now,
            content: raw_text,
            format: :markdown
        }
      end)
      |> regenerate_summaries()

    {:ok, signal}
  end

  # Step 2: Route
  defp route_step(%Signal{} = signal, opts) do
    forced_node = Keyword.get(opts, :node)

    primary_node =
      if forced_node, do: forced_node, else: signal.node || "inbox"

    signal = %{signal | node: primary_node}

    routed_to =
      case Router.route(signal) do
        {:ok, destinations} -> destinations
        {:error, _} -> [primary_node]
      end

    signal = %{signal | routed_to: routed_to}

    {:ok, signal, routed_to}
  end

  # Step 3: Write primary file
  defp write_step(%Signal{} = signal, _opts) do
    Writer.write_signal(signal)
  end

  # Step 4: Write cross-references
  defp cross_ref_step(%Signal{} = signal, routed_to) do
    primary_folder = Writer.node_to_folder(signal.node)

    cross_nodes =
      routed_to
      |> Enum.reject(fn dest ->
        Writer.node_to_folder(dest) == primary_folder
      end)

    Writer.write_cross_references(signal, cross_nodes)
  end

  # Step 5: Index — write the primary file to SQLite via Indexer
  defp index_step(%Signal{} = signal, primary_path, opts) do
    # Build a complete context for storage
    uri = URI.from_path(primary_path)
    workspace_id = Keyword.get(opts, :workspace_id, "default")

    context = %Context{
      id: signal.id,
      uri: uri,
      type: :signal,
      path: primary_path,
      title: signal.title,
      content: signal.content,
      l0_abstract: signal.l0_summary || "",
      l1_overview: signal.l1_description || "",
      signal: %{signal | path: primary_path},
      node: signal.node,
      sn_ratio: signal.sn_ratio || 0.6,
      entities: signal.entities || [],
      created_at: signal.created_at,
      modified_at: signal.modified_at,
      routed_to: signal.routed_to || [],
      workspace_id: workspace_id,
      metadata: %{}
    }

    case Store.insert_context(context) do
      :ok ->
        index_cross_refs(signal)
        {:ok, context}

      {:error, _} = err ->
        err
    end
  end

  defp index_cross_refs(%Signal{} = signal) do
    primary_folder = Writer.node_to_folder(signal.node)

    Enum.each(signal.routed_to || [], fn dest ->
      if Writer.node_to_folder(dest) != primary_folder do
        cross_path = cross_file_path(signal, dest)
        if File.exists?(cross_path), do: Indexer.index_file(cross_path)
      end
    end)
  end

  defp cross_file_path(%Signal{} = signal, dest_node) do
    root = Application.get_env(:optimal_engine, :root_path, "")
    dest_folder = Writer.node_to_folder(dest_node)
    filename = Writer.relative_path(%{signal | node: dest_node}) |> Path.basename()
    Path.join([root, dest_folder, "signals", filename])
  end

  # ---------------------------------------------------------------------------
  # Private: Helpers
  # ---------------------------------------------------------------------------

  # Async semantic enhancement: LLM summaries + embedding + sidecar files
  defp semantic_enhance(%Context{} = context) do
    case SemanticProcessor.process(context) do
      {:ok, enhanced} ->
        # Update the stored context with enhanced L0/L1
        if enhanced.l0_abstract != context.l0_abstract or
             enhanced.l1_overview != context.l1_overview do
          Store.raw_query(
            "UPDATE contexts SET l0_abstract = ?1, l1_overview = ?2 WHERE id = ?3",
            [enhanced.l0_abstract, enhanced.l1_overview, context.id]
          )
        end

      _ ->
        :ok
    end
  rescue
    _ -> :ok
  end

  defp maybe_override(signal, _key, nil), do: signal
  defp maybe_override(signal, :title, val), do: %{signal | title: val}
  defp maybe_override(signal, :genre, val), do: %{signal | genre: val}
  defp maybe_override(signal, :node, val), do: %{signal | node: val}

  defp maybe_merge_entities(signal, []), do: signal

  defp maybe_merge_entities(signal, extra) do
    merged = Enum.uniq((signal.entities || []) ++ extra)
    %{signal | entities: merged}
  end

  defp regenerate_summaries(%Signal{} = signal) do
    l0 =
      "#{String.upcase(signal.genre || "note")} | #{signal.node || "inbox"} | #{signal.title || "Untitled"} [S/N: #{Float.round(signal.sn_ratio || 0.6, 1)}]"

    l1 =
      signal.content
      |> String.replace(~r/---.*?---/s, "")
      |> String.trim()
      |> String.slice(0, 500)

    %{signal | l0_summary: l0, l1_description: l1}
  end

  defp relative(abs_path) do
    root = Application.get_env(:optimal_engine, :root_path, "")
    String.replace_prefix(abs_path, root <> "/", "")
  end

  defp generate_id(text, dt) do
    :crypto.hash(:sha256, text <> DateTime.to_iso8601(dt))
    |> Base.encode16(case: :lower)
    |> String.slice(0, 32)
  end

  defp load_topology do
    case Routing.load() do
      {:ok, t} -> t
      {:error, _} -> %{endpoints: %{}, root_path: nil}
    end
  end

  defp topology_entities(topology) do
    topology
    |> Map.get(:endpoints, %{})
    |> Map.values()
    |> Enum.flat_map(fn ep ->
      name = Map.get(ep, :name, "")
      first = name |> String.split(" ") |> List.first("")
      [name, first]
    end)
    |> Enum.reject(&(String.length(&1) < 3))
    |> Enum.uniq()
  end
end
