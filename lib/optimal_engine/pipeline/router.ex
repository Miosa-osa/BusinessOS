defmodule OptimalEngine.Pipeline.Router do
  @moduledoc """
  Routes signals to destination nodes based on topology routing rules.

  Rules are loaded from `topology.yaml` at startup and cached in GenServer state.
  Rule evaluation is deterministic — all matching rules fire (not first-match).

  Priority levels: `:critical` > `:high` > `:normal` > `:low`

  Cross-cutting rules from `config.yaml` are also applied:
  - Financial genres always copy to `money-revenue`
  - Any signal mentioning a known person copies to `team`
  """

  use GenServer
  require Logger

  alias OptimalEngine.{Signal, Routing}

  @financial_genres ~w[
    invoice profit-loss balance-sheet budget
    expense-report revenue-report financial-forecast financial-model
  ]

  # --- Client API ---

  @spec start_link(keyword()) :: GenServer.on_start()
  def start_link(opts \\ []) do
    GenServer.start_link(__MODULE__, opts, name: __MODULE__)
  end

  @doc """
  Routes a Signal and returns the list of destination node IDs.
  Always returns at least `[\"inbox\"]` as a fallback.
  """
  @spec route(Signal.t()) :: {:ok, [String.t()]} | {:error, term()}
  def route(%Signal{} = signal) do
    GenServer.call(__MODULE__, {:route, signal})
  end

  @doc "Reloads routing rules from topology files."
  @spec reload() :: :ok
  def reload do
    GenServer.call(__MODULE__, :reload)
  end

  # --- GenServer Callbacks ---

  @impl true
  def init(_opts) do
    state = load_state()
    Logger.info("[Router] Loaded #{length(state.rules)} routing rules")
    {:ok, state}
  end

  @impl true
  def handle_call({:route, signal}, _from, state) do
    destinations =
      signal
      |> apply_rules(state.rules)
      |> apply_cross_cutting(signal)
      |> Enum.uniq()
      |> normalize_destinations()

    result = if destinations == [], do: ["inbox"], else: destinations

    :telemetry.execute(
      [:optimal_engine, :router, :route],
      %{destinations: length(result)},
      %{node: signal.node, genre: signal.genre}
    )

    {:reply, {:ok, result}, state}
  end

  @impl true
  def handle_call(:reload, _from, _state) do
    new_state = load_state()
    Logger.info("[Router] Reloaded #{length(new_state.rules)} routing rules")
    {:reply, :ok, new_state}
  end

  # --- Private: Rule Evaluation ---

  defp apply_rules(signal, rules) do
    rules
    |> Enum.filter(&rule_matches?(signal, &1))
    |> Enum.flat_map(& &1.destination)
  end

  defp rule_matches?(signal, rule) do
    condition = rule.condition

    cond do
      Map.has_key?(condition, "default") ->
        false

      Map.has_key?(condition, "signal_type") ->
        to_string(signal.type) == to_string(condition["signal_type"])

      Map.has_key?(condition, "source_node") ->
        signal.node == condition["source_node"]

      Map.has_key?(condition, "entity_mention") ->
        mentions = condition["entity_mention"]

        Enum.any?(mentions, fn entity ->
          Regex.match?(~r/#{Regex.escape(entity)}/i, signal.content || "") ||
            entity in (signal.entities || [])
        end)

      Map.has_key?(condition, "content_pattern") ->
        patterns = condition["content_pattern"]

        Enum.any?(patterns, fn pattern ->
          case Regex.compile(pattern) do
            {:ok, re} -> Regex.match?(re, signal.content || "")
            _ -> false
          end
        end)

      true ->
        false
    end
  end

  defp apply_cross_cutting(destinations, signal) do
    extra =
      []
      |> maybe_add_financial(signal)
      |> maybe_add_team(signal)

    destinations ++ extra
  end

  defp maybe_add_financial(acc, signal) do
    if signal.genre in @financial_genres do
      ["11-money-revenue" | acc]
    else
      acc
    end
  end

  defp maybe_add_team(acc, signal) do
    if length(signal.entities || []) > 0 do
      ["10-team" | acc]
    else
      acc
    end
  end

  # Normalize: strip entity-path prefixes, map folder names to node IDs
  defp normalize_destinations(destinations) do
    destinations
    |> Enum.map(fn dest ->
      dest
      |> String.split("/")
      |> List.last()
      |> map_to_folder_id()
    end)
    |> Enum.reject(&is_nil/1)
    |> Enum.uniq()
  end

  # Maps destination strings to canonical folder/node names
  defp map_to_folder_id("roberto"), do: "01-roberto"
  defp map_to_folder_id("inbox"), do: "09-new-stuff"
  defp map_to_folder_id("money-revenue"), do: "11-money-revenue"
  defp map_to_folder_id("team"), do: "10-team"
  defp map_to_folder_id("governance/l7-bypass"), do: "01-roberto"
  defp map_to_folder_id("ai-masters"), do: "04-ai-masters"
  defp map_to_folder_id("os-accelerator"), do: "12-os-accelerator"
  defp map_to_folder_id("agency-accelerants"), do: "06-agency-accelerants"
  defp map_to_folder_id("content/mosaic-effect"), do: "08-content-creators"
  defp map_to_folder_id("miosa-core"), do: "02-miosa"
  defp map_to_folder_id("compute-engine"), do: "02-miosa"
  defp map_to_folder_id("optimal-system"), do: "05-os-architect"
  defp map_to_folder_id(other), do: other

  defp load_state do
    topology =
      case Routing.load() do
        {:ok, t} -> t
        {:error, _} -> %{routing_rules: []}
      end

    %{rules: topology.routing_rules}
  end
end
