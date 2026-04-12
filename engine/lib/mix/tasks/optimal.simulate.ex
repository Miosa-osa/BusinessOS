defmodule Mix.Tasks.Optimal.Simulate do
  @shortdoc "Run a 'what if' scenario simulation through the knowledge graph"
  @moduledoc """
  Traces a 'what if' scenario through the OptimalOS knowledge graph and
  produces a structured impact analysis report.

  Usage:
      mix optimal.simulate "What if Ed Honour leaves AI Masters?"
      mix optimal.simulate "What if we cancel Agency Accelerants?" --depth 4
      mix optimal.simulate "What if revenue drops 50%?"
      mix optimal.simulate "What if Ed Honour leaves?" --plan
      mix optimal.simulate "What if Ed Honour leaves?" --mc
      mix optimal.simulate "What if Bennett leaves?" --plan --mc

  Options:
    --depth   BFS traversal depth (default: 3)
    --plan    Run MCTS response planning after impact analysis
    --mc      Run Monte Carlo probability sampling after impact analysis
  """

  use Mix.Task

  alias OptimalEngine.{MCTS, MonteCarlo}

  @separator String.duplicate("─", 60)
  @thin_sep String.duplicate("·", 60)

  @impl Mix.Task
  def run(args) do
    Mix.Task.run("app.start")

    {opts, positional, _} =
      OptionParser.parse(args,
        switches: [depth: :integer, plan: :boolean, mc: :boolean],
        aliases: [d: :depth, p: :plan, m: :mc]
      )

    scenario =
      case positional do
        [s | _] -> s
        [] -> Mix.raise("Usage: mix optimal.simulate \"scenario\"")
      end

    sim_opts = Keyword.take(opts, [:depth]) |> rename_key(:depth, :max_depth)
    run_plan = Keyword.get(opts, :plan, false)
    run_mc = Keyword.get(opts, :mc, false)

    IO.puts("")
    IO.puts(@separator)
    IO.puts("  SCENARIO SIMULATION")
    IO.puts(@separator)
    IO.puts("  Scenario: #{scenario}")
    IO.puts("")

    case OptimalEngine.Simulator.simulate(scenario, sim_opts) do
      {:ok, report} ->
        print_report(report)

        if run_plan do
          print_mcts_plan(report)
        end

        if run_mc do
          print_monte_carlo(report)
        end

      {:error, reason} ->
        IO.puts("Simulation failed: #{inspect(reason)}")
    end
  end

  # ---------------------------------------------------------------------------
  # MCTS output
  # ---------------------------------------------------------------------------

  defp print_mcts_plan(report) do
    IO.puts(@separator)
    IO.puts("  Response Plan (MCTS — 32 simulations)")
    IO.puts(@thin_sep)

    case MCTS.plan_response(report, budget: 32) do
      {:ok, plan} ->
        confidence_pct = trunc(plan.confidence * 100)
        IO.puts("  Best sequence (confidence: #{plan.confidence} / #{confidence_pct}%):")

        if plan.best_sequence == [] do
          IO.puts("    (no actions needed — system health is acceptable)")
        else
          plan.best_sequence
          |> Enum.with_index(1)
          |> Enum.each(fn {action, i} ->
            cost_tag = format_cost(action.estimated_cost)
            IO.puts("    #{i}. [#{action.type}] #{action.description}  (cost: #{cost_tag})")
          end)
        end

        IO.puts("")
        IO.puts("  Expected outcome:")
        o = plan.expected_outcome
        IO.puts("    Revenue health:  #{o.revenue_impact}  #{score_bar(o.revenue_impact)}")
        IO.puts("    Coverage:        #{o.coverage_score}  #{score_bar(o.coverage_score)}")
        IO.puts("    Overall health:  #{o.health_score}  #{score_bar(o.health_score)}")
        IO.puts("")

        if plan.alternatives != [] do
          IO.puts("  Alternatives:")

          plan.alternatives
          |> Enum.with_index(2)
          |> Enum.each(fn {alt, rank} ->
            action_summary =
              alt.sequence
              |> Enum.map(fn a -> "#{a.type} #{a.target_node}" end)
              |> Enum.join(" → ")

            action_str = if action_summary == "", do: "(no actions)", else: action_summary
            IO.puts("    ##{rank} (score: #{alt.score}): #{action_str}")
          end)

          IO.puts("")
        end

        ts = plan.tree_stats
        IO.puts("  Tree: #{ts.nodes} nodes explored, max depth #{ts.max_depth}, avg branching #{ts.avg_branching}")
        IO.puts("")

      {:error, reason} ->
        IO.puts("  MCTS planning failed: #{inspect(reason)}")
        IO.puts("")
    end
  end

  # ---------------------------------------------------------------------------
  # Monte Carlo output
  # ---------------------------------------------------------------------------

  defp print_monte_carlo(report) do
    n = 1000

    IO.puts(@separator)
    IO.puts("  Monte Carlo Analysis (#{n} simulations)")
    IO.puts(@thin_sep)

    case MonteCarlo.sample(report, simulations: n) do
      {:ok, dist} ->
        o = dist.outcomes
        {ci_lo, ci_hi} = dist.confidence_interval_95

        IO.puts("  Expected health:     #{o.expected.total_health}")
        IO.puts("  Median health:       #{o.median}")
        IO.puts("  95% CI:              [#{ci_lo}, #{ci_hi}]")
        IO.puts("")
        IO.puts("  Distribution:")

        max_count = Enum.max(dist.histogram)

        dist.histogram
        |> Enum.with_index()
        |> Enum.each(fn {count, bin} ->
          low = Float.round(bin * 0.1, 1)
          high = Float.round((bin + 1) * 0.1, 1)
          bar = mc_bar(count, max_count)
          range_str = "#{low}-#{high}" |> String.pad_trailing(7)
          count_str = count |> to_string() |> String.pad_leading(6)
          IO.puts("  #{range_str} | #{bar} #{count_str}")
        end)

        IO.puts("")

        if map_size(dist.node_distributions) > 0 do
          IO.puts("  Per-node breakdown:")

          dist.node_distributions
          |> Enum.sort_by(fn {_node, d} -> d.mean_impact end)
          |> Enum.each(fn {node, d} ->
            severe_pct = trunc(d.probability_of_severe_impact * 100)
            node_str = String.pad_trailing(node, 22)
            IO.puts("  #{node_str}  mean=#{d.mean_impact} ±#{d.std_dev}  P(severe)=#{severe_pct}%")
          end)

          IO.puts("")
        end

      {:error, reason} ->
        IO.puts("  Monte Carlo sampling failed: #{inspect(reason)}")
        IO.puts("")
    end
  end

  # ---------------------------------------------------------------------------
  # Existing impact report output
  # ---------------------------------------------------------------------------

  defp print_report(report) do
    print_summary(report)
    print_affected_nodes(report.affected_nodes)
    print_affected_entities(report.affected_entities)
    print_critical_dependencies(report.critical_dependencies)
    print_risk_assessment(report.risk_assessment)
    print_recommendations(report.recommendations)
    print_footer(report)
  end

  defp print_summary(report) do
    IO.puts("  Mutation type:        #{format_mutation_type(report.mutation_type)}")
    IO.puts("  Contexts affected:    #{report.total_contexts_affected}")
    IO.puts("  Nodes affected:       #{length(report.affected_nodes)}")
    IO.puts("  Graph depth:          #{report.graph_depth}")
    IO.puts("")
  end

  defp print_affected_nodes([]) do
    IO.puts("  Affected Nodes: (none)")
    IO.puts("")
  end

  defp print_affected_nodes(nodes) do
    IO.puts("  Affected Nodes  (impact score / context coverage)")
    IO.puts(@thin_sep)

    Enum.each(nodes, fn node ->
      bar = ascii_bar(node.impact_score)
      score_str = node.impact_score |> Float.round(3) |> to_string() |> String.pad_leading(5)
      name_str = String.pad_trailing(node.node, 26)
      IO.puts("  #{name_str} #{score_str}  #{bar}  #{node.description}")
    end)

    IO.puts("")
  end

  defp print_affected_entities([]) do
    :ok
  end

  defp print_affected_entities(entities) do
    IO.puts("  Affected Entities")
    IO.puts(@thin_sep)

    Enum.each(entities, fn e ->
      impact_tag = if e.impact == :direct, do: "[direct]  ", else: "[indirect]"
      IO.puts("  #{impact_tag}  #{e.name}  (#{e.contexts} context(s))")
    end)

    IO.puts("")
  end

  defp print_critical_dependencies([]) do
    IO.puts("  Critical Dependencies: none identified")
    IO.puts("")
  end

  defp print_critical_dependencies(deps) do
    IO.puts("  Critical Dependencies")
    IO.puts(@thin_sep)

    Enum.each(deps, fn dep ->
      weight_pct = trunc(dep.weight * 100)
      IO.puts("  #{dep.from}  --#{dep.relation}-->  #{dep.to}  (#{weight_pct}% coverage)")
    end)

    IO.puts("")
  end

  defp print_risk_assessment(%{severity: severity, reasoning: reasoning}) do
    label = format_severity(severity)
    IO.puts("  Risk Assessment: #{label}")
    IO.puts(@thin_sep)
    IO.puts("  #{reasoning}")
    IO.puts("")
  end

  defp print_recommendations([]) do
    IO.puts("  Recommendations: none generated")
    IO.puts("")
  end

  defp print_recommendations(recs) do
    IO.puts("  Recommendations")
    IO.puts(@thin_sep)

    recs
    |> Enum.with_index(1)
    |> Enum.each(fn {rec, i} ->
      IO.puts("  #{i}. #{rec}")
    end)

    IO.puts("")
  end

  defp print_footer(report) do
    IO.puts(@separator)
    IO.puts("  Simulation complete: #{report.scenario}")
    IO.puts(@separator)
    IO.puts("")
  end

  # ---------------------------------------------------------------------------
  # Formatting helpers
  # ---------------------------------------------------------------------------

  defp ascii_bar(score) when is_float(score) do
    filled = min(trunc(score * 2), 20)
    empty = 20 - filled
    "[" <> String.duplicate("█", filled) <> String.duplicate("░", empty) <> "]"
  end

  defp score_bar(score) when is_float(score) do
    filled = min(trunc(score * 20), 20)
    empty = 20 - filled
    "[" <> String.duplicate("█", filled) <> String.duplicate("░", empty) <> "]"
  end

  defp mc_bar(_count, max_count) when max_count == 0, do: String.duplicate(" ", 26)

  defp mc_bar(count, max_count) do
    filled = min(trunc(count / max_count * 26), 26)
    String.duplicate("█", filled) <> String.duplicate(" ", 26 - filled)
  end

  defp format_cost(:low), do: "low"
  defp format_cost(:medium), do: "medium"
  defp format_cost(:high), do: "high"

  defp rename_key(kw, old_key, new_key) do
    case Keyword.get(kw, old_key) do
      nil -> kw
      val -> kw |> Keyword.delete(old_key) |> Keyword.put(new_key, val)
    end
  end

  defp format_mutation_type(:entity_removal), do: "entity removal"
  defp format_mutation_type(:node_cancellation), do: "node cancellation"
  defp format_mutation_type(:revenue_change), do: "revenue change"
  defp format_mutation_type(:dependency_break), do: "dependency break"
  defp format_mutation_type(:general), do: "general"

  defp format_severity(:critical), do: "CRITICAL"
  defp format_severity(:high), do: "HIGH"
  defp format_severity(:medium), do: "MEDIUM"
  defp format_severity(:low), do: "LOW"
end
