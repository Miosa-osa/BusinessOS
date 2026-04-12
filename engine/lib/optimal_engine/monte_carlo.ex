defmodule OptimalEngine.MonteCarlo do
  @moduledoc """
  Monte Carlo sampling for probability estimation over the knowledge graph.

  Given a simulation impact report and edge probabilities, samples N possible
  outcomes to estimate probability distributions of impacts across nodes.

  Each simulation flip-coins each affected edge based on its stored probability.
  If the edge fires, the impact propagates; otherwise, that connection holds and
  the node recovers partially.

  ## Usage

      {:ok, report} = Simulator.simulate("What if Ed leaves?")
      {:ok, dist}   = MonteCarlo.sample(report, simulations: 1000)
  """

  require Logger

  alias OptimalEngine.Store

  @default_simulations 1000

  # Nodes that directly affect revenue
  @revenue_nodes ~w[money-revenue ai-masters agency-accelerants miosa-platform lunivate]

  # Topology: which people work on which nodes
  @topology %{
    "ai-masters" => ["Robert Potter", "Adam", "Roberto", "Ed Honour"],
    "miosa-platform" => ["Pedram", "Pedro", "Abdul", "Nejd", "Javaris", "Roberto"],
    "agency-accelerants" => ["Bennett", "Len", "Liam", "Roberto"],
    "content-creators" => ["Bennett", "Ahmed", "Tejas", "Ikram", "Roberto"],
    "accelerants-community" => ["Bennett", "Roberto"],
    "os-architect" => ["Ahmed", "Roberto"],
    "roberto" => ["Jordan", "Roberto"],
    "money-revenue" => ["Roberto"],
    "team" => ["Roberto"],
    "lunivate" => ["Roberto"],
    "os-accelerator" => ["Roberto"]
  }

  # ---------------------------------------------------------------------------
  # Types
  # ---------------------------------------------------------------------------

  @type node_distribution :: %{
          mean_impact: float(),
          std_dev: float(),
          p10: float(),
          p50: float(),
          p90: float(),
          probability_of_severe_impact: float()
        }

  @type distribution :: %{
          simulations: non_neg_integer(),
          outcomes: %{
            best_case: map(),
            worst_case: map(),
            expected: map(),
            median: float()
          },
          node_distributions: %{String.t() => node_distribution()},
          confidence_interval_95: {float(), float()},
          histogram: [non_neg_integer()]
        }

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Runs N Monte Carlo simulations over a simulation impact report.

  ## Options
    - `:simulations` — number of samples (default: #{@default_simulations})
  """
  @spec sample(map(), keyword()) :: {:ok, distribution()} | {:error, term()}
  def sample(simulation_report, opts \\ []) do
    n = Keyword.get(opts, :simulations, @default_simulations)

    base_state = build_base_state(simulation_report)
    edge_probs = load_edge_probabilities(simulation_report)

    results =
      Enum.map(1..n, fn _i ->
        sampled_state = sample_one(base_state, edge_probs)
        score_state(sampled_state)
      end)

    {:ok, build_distribution(results, n)}
  rescue
    e ->
      Logger.error("[MonteCarlo] sample failed: #{inspect(e)}")
      {:error, {:mc_failed, e}}
  end

  # ---------------------------------------------------------------------------
  # State construction
  # ---------------------------------------------------------------------------

  defp build_base_state(%{affected_nodes: nodes} = report) do
    removed_entities =
      report
      |> Map.get(:affected_entities, [])
      |> Enum.filter(fn e -> e.impact == :direct end)
      |> Enum.map(& &1.name)

    node_map =
      Enum.reduce(nodes, %{}, fn node_info, acc ->
        raw_score = Map.get(node_info, :impact_score, 0.0)
        health = max(0.0, 1.0 - raw_score / 10.0)

        Map.put(acc, node_info.node, %{
          health: Float.round(health, 3),
          entities: Map.get(@topology, node_info.node, []),
          revenue_exposed: node_info.node in @revenue_nodes,
          context_count: Map.get(node_info, :context_count, 0)
        })
      end)

    # Fill in unaffected nodes at full health
    node_map =
      Enum.reduce(@topology, node_map, fn {node, entities}, acc ->
        if Map.has_key?(acc, node) do
          acc
        else
          Map.put(acc, node, %{
            health: 1.0,
            entities: entities,
            revenue_exposed: node in @revenue_nodes,
            context_count: 0
          })
        end
      end)

    %{
      nodes: node_map,
      removed_entities: removed_entities,
      total_health: avg_health(node_map)
    }
  end

  defp build_base_state(_), do: %{nodes: %{}, removed_entities: [], total_health: 1.0}

  # ---------------------------------------------------------------------------
  # Edge probability loading
  # ---------------------------------------------------------------------------

  # Returns a list of {edge_key, probability} where edge_key is {source, target, relation}.
  # Falls back to default probability 0.8 if the column hasn't been populated.
  defp load_edge_probabilities(%{affected_nodes: nodes}) do
    node_names = Enum.map(nodes, & &1.node)

    Enum.flat_map(node_names, fn node ->
      sql = """
      SELECT source_id, target_id, relation, COALESCE(probability, 0.8)
      FROM edges
      WHERE target_id = ?1 OR source_id = ?1
      LIMIT 200
      """

      case Store.raw_query(sql, [node]) do
        {:ok, rows} ->
          Enum.map(rows, fn [src, tgt, rel, prob] ->
            {{src, tgt, rel}, prob || 0.8}
          end)

        _ ->
          []
      end
    end)
    |> Enum.uniq_by(fn {k, _} -> k end)
  end

  defp load_edge_probabilities(_), do: []

  # ---------------------------------------------------------------------------
  # Single simulation
  # ---------------------------------------------------------------------------

  defp sample_one(base_state, edge_probs) do
    Enum.reduce(edge_probs, base_state, fn {{_src, target, _rel}, prob}, state ->
      # If the edge fires (impact propagates), degrade the target node's health
      if :rand.uniform() <= prob do
        propagate_impact(state, target)
      else
        # Connection holds — partial mitigation (+small boost to that node)
        mitigate(state, target)
      end
    end)
  end

  defp propagate_impact(state, node_name) do
    # Apply a random degradation proportional to edge probability
    delta = -(0.05 + :rand.uniform() * 0.15)
    update_node_health(state, node_name, delta)
  end

  defp mitigate(state, node_name) do
    update_node_health(state, node_name, 0.02)
  end

  defp update_node_health(state, node_name, delta) do
    nodes =
      Map.update(state.nodes, node_name, nil, fn info ->
        new_h = Float.round(min(1.0, max(0.0, info.health + delta)), 3)
        %{info | health: new_h}
      end)

    # If node_name wasn't in the map, nothing changes
    if nodes == state.nodes do
      state
    else
      %{state | nodes: nodes, total_health: avg_health(nodes)}
    end
  end

  # ---------------------------------------------------------------------------
  # Scoring a sampled state
  # ---------------------------------------------------------------------------

  defp score_state(%{nodes: nodes} = _state) do
    node_list = Enum.to_list(nodes)

    node_healths = Map.new(node_list, fn {name, info} -> {name, info.health} end)
    total_health = avg_health(nodes)

    revenue_nodes = Enum.filter(node_list, fn {_, info} -> info.revenue_exposed end)

    revenue_health =
      if revenue_nodes == [] do
        1.0
      else
        Enum.sum(Enum.map(revenue_nodes, fn {_, info} -> info.health end)) / length(revenue_nodes)
      end

    %{
      total_health: Float.round(total_health, 4),
      revenue_health: Float.round(revenue_health, 4),
      node_healths: node_healths
    }
  end

  # ---------------------------------------------------------------------------
  # Distribution statistics
  # ---------------------------------------------------------------------------

  defp build_distribution(results, n) do
    sorted = Enum.sort_by(results, & &1.total_health)
    healths = Enum.map(sorted, & &1.total_health)

    mean = Enum.sum(healths) / n
    median = safe_at(healths, div(n, 2))
    histogram =
      Enum.map(0..9, fn bin ->
        low = bin * 0.1
        high = if bin == 9, do: 1.001, else: (bin + 1) * 0.1
        Enum.count(healths, fn h -> h >= low and h < high end)
      end)

    ci_lo = safe_at(healths, max(0, round(n * 0.025)))
    ci_hi = safe_at(healths, min(n - 1, round(n * 0.975)))

    node_dists = compute_per_node_distributions(results)

    %{
      simulations: n,
      outcomes: %{
        best_case: %{total_health: safe_at(healths, n - 1), description: "Best observed"},
        worst_case: %{total_health: safe_at(healths, 0), description: "Worst observed"},
        expected: %{total_health: Float.round(mean, 3), description: "Expected outcome"},
        median: Float.round(median, 3)
      },
      node_distributions: node_dists,
      confidence_interval_95: {Float.round(ci_lo, 3), Float.round(ci_hi, 3)},
      histogram: histogram
    }
  end

  defp compute_per_node_distributions(results) do
    all_nodes =
      results
      |> Enum.flat_map(fn r -> Map.keys(r.node_healths) end)
      |> Enum.uniq()

    Map.new(all_nodes, fn node ->
      vals =
        results
        |> Enum.map(fn r -> Map.get(r.node_healths, node, 1.0) end)
        |> Enum.sort()

      n = length(vals)
      mean = Enum.sum(vals) / max(n, 1)

      std =
        :math.sqrt(
          Enum.sum(Enum.map(vals, fn v -> (v - mean) * (v - mean) end)) / max(n, 1)
        )

      severe_count = Enum.count(vals, fn v -> v < 0.3 end)

      {node,
       %{
         mean_impact: Float.round(mean, 3),
         std_dev: Float.round(std, 3),
         p10: Float.round(safe_at(vals, div(n, 10)), 3),
         p50: Float.round(safe_at(vals, div(n, 2)), 3),
         p90: Float.round(safe_at(vals, div(n * 9, 10)), 3),
         probability_of_severe_impact: Float.round(severe_count / max(n, 1), 3)
       }}
    end)
  end

  # ---------------------------------------------------------------------------
  # Helpers
  # ---------------------------------------------------------------------------

  defp avg_health(node_map) when map_size(node_map) == 0, do: 1.0

  defp avg_health(node_map) do
    values = Map.values(node_map)
    Float.round(Enum.sum(Enum.map(values, & &1.health)) / length(values), 3)
  end

  defp safe_at(list, idx) do
    idx = max(0, min(idx, length(list) - 1))
    Enum.at(list, idx) || 0.0
  end
end
