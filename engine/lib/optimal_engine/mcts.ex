defmodule OptimalEngine.MCTS do
  @moduledoc """
  Monte Carlo Tree Search for response planning.

  Given a simulation impact report (from Simulator), explores the space of possible
  response actions to find the optimal recovery strategy.

  ## How it works

  1. Root state = knowledge graph health snapshot derived from the simulation report
  2. Actions = possible responses (reassign, hire, restructure, pause, pivot, redistribute)
  3. Rollout = random simulation of the action's effect on node health scores
  4. Evaluation = weighted score of revenue health, coverage, and overall health
  5. Backpropagate = update UCB1 values up the tree
  6. After N simulations, extract the best action sequence

  ## Usage

      {:ok, impact} = Simulator.simulate("What if Ed leaves?")
      {:ok, plan}   = MCTS.plan_response(impact, budget: 32)
  """

  require Logger

  # Topology: which people work on which nodes (mirrors Graph topology)
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

  # Nodes that directly affect revenue
  @revenue_nodes ~w[money-revenue ai-masters agency-accelerants miosa-platform lunivate]

  # ---------------------------------------------------------------------------
  # Types
  # ---------------------------------------------------------------------------

  @type action :: %{
          type: :reassign | :hire | :restructure | :pause | :pivot | :redistribute,
          target_node: String.t(),
          target_entity: String.t() | nil,
          description: String.t(),
          estimated_cost: :low | :medium | :high
        }

  @type plan :: %{
          best_sequence: [action()],
          explored_paths: non_neg_integer(),
          confidence: float(),
          expected_outcome: %{
            revenue_impact: float(),
            coverage_score: float(),
            health_score: float(),
            overall: float()
          },
          alternatives: [%{sequence: [action()], score: float()}],
          tree_stats: %{nodes: integer(), max_depth: integer(), avg_branching: float()}
        }

  # ---------------------------------------------------------------------------
  # Tree node struct
  # ---------------------------------------------------------------------------

  defmodule TreeNode do
    @moduledoc false
    defstruct [
      :id,
      :state,
      :action,
      :parent_id,
      children_ids: [],
      visits: 0,
      total_value: 0.0,
      untried_actions: []
    ]
  end

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Plans the optimal response to a simulation impact report using MCTS.

  ## Options
    - `:budget`  — number of MCTS iterations (default: 32)
    - `:ucb_c`   — UCB1 exploration constant (default: 1.41)
  """
  @spec plan_response(map(), keyword()) :: {:ok, plan()} | {:error, term()}
  def plan_response(simulation_report, opts \\ []) do
    budget = Keyword.get(opts, :budget, 32)
    ucb_c = Keyword.get(opts, :ucb_c, 1.41)

    initial_state = build_initial_state(simulation_report)
    root_actions = generate_actions(initial_state)

    root_id = make_ref()

    root = %TreeNode{
      id: root_id,
      state: initial_state,
      action: nil,
      parent_id: nil,
      untried_actions: root_actions,
      visits: 0,
      total_value: 0.0
    }

    tree = %{root_id => root}

    {final_tree, _root_id} =
      Enum.reduce(1..budget, {tree, root_id}, fn _i, {t, rid} ->
        {node_id, t2} = select(t, rid, ucb_c)
        {expanded_id, t3} = expand(t2, node_id)
        value = simulate_rollout(t3, expanded_id)
        t4 = backpropagate(t3, expanded_id, value)
        {t4, rid}
      end)

    best = extract_best_path(final_tree, root_id)
    alternatives = extract_alternatives(final_tree, root_id, 3)

    {:ok,
     %{
       best_sequence: best.actions,
       explored_paths: budget,
       confidence: best.confidence,
       expected_outcome: best.outcome,
       alternatives: alternatives,
       tree_stats: compute_tree_stats(final_tree)
     }}
  rescue
    e ->
      Logger.error("[MCTS] plan_response failed: #{inspect(e)}")
      {:error, {:mcts_failed, e}}
  end

  # ---------------------------------------------------------------------------
  # State construction
  # ---------------------------------------------------------------------------

  defp build_initial_state(%{affected_nodes: nodes} = report) do
    removed_entities =
      report
      |> Map.get(:affected_entities, [])
      |> Enum.filter(fn e -> e.impact == :direct end)
      |> Enum.map(& &1.name)

    node_map =
      Enum.reduce(nodes, %{}, fn node_info, acc ->
        # Normalize impact_score (0..10) to health (0..1): inverted
        raw_score = Map.get(node_info, :impact_score, 0.0)
        health = max(0.0, 1.0 - raw_score / 10.0)
        revenue_exposed = node_info.node in @revenue_nodes

        entities = Map.get(@topology, node_info.node, [])

        Map.put(acc, node_info.node, %{
          health: Float.round(health, 3),
          entities: entities,
          revenue_exposed: revenue_exposed,
          context_count: Map.get(node_info, :context_count, 0)
        })
      end)

    # Ensure all topology nodes exist in the state (unaffected = health 1.0)
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
      removed_nodes: [],
      actions_taken: [],
      total_health: compute_total_health(node_map)
    }
  end

  defp build_initial_state(_report) do
    %{nodes: %{}, removed_entities: [], removed_nodes: [], actions_taken: [], total_health: 1.0}
  end

  defp compute_total_health(node_map) when map_size(node_map) == 0, do: 1.0

  defp compute_total_health(node_map) do
    nodes = Map.values(node_map)
    Float.round(Enum.sum(Enum.map(nodes, & &1.health)) / length(nodes), 3)
  end

  # ---------------------------------------------------------------------------
  # Action generation
  # ---------------------------------------------------------------------------

  defp generate_actions(state) do
    affected =
      state.nodes
      |> Enum.filter(fn {_name, info} -> info.health < 0.9 end)
      |> Enum.sort_by(fn {_name, info} -> info.health end)
      |> Enum.take(4)

    Enum.flat_map(affected, fn {node_name, node_info} ->
      available = find_available_entities(state, node_name)

      reassign_actions =
        available
        |> Enum.take(2)
        |> Enum.map(fn entity ->
          %{
            type: :reassign,
            target_node: node_name,
            target_entity: entity,
            description: "Reassign #{node_name} responsibilities to #{entity}",
            estimated_cost: :low
          }
        end)

      hire_action = %{
        type: :hire,
        target_node: node_name,
        target_entity: nil,
        description: "Hire replacement for #{node_name}",
        estimated_cost: :high
      }

      restructure_action = %{
        type: :restructure,
        target_node: node_name,
        target_entity: nil,
        description: "Restructure #{node_name} — distribute to connected nodes",
        estimated_cost: :medium
      }

      pause_action = %{
        type: :pause,
        target_node: node_name,
        target_entity: nil,
        description: "Pause #{node_name} operations temporarily",
        estimated_cost: :low
      }

      # Only offer redistribute if there are multiple affected nodes
      base = reassign_actions ++ [hire_action, restructure_action]

      base =
        if node_info.revenue_exposed do
          base
        else
          base ++ [pause_action]
        end

      base
    end)
    |> Enum.uniq_by(fn a -> {a.type, a.target_node, a.target_entity} end)
  end

  defp find_available_entities(state, node_name) do
    all_entities = Map.get(@topology, node_name, [])
    removed = state.removed_entities

    all_entities
    |> Enum.reject(fn e -> e in removed end)
    |> Enum.filter(fn e ->
      # Only suggest entities that exist in other nodes (can be redistributed)
      Enum.any?(@topology, fn {other_node, entities} ->
        other_node != node_name and e in entities
      end)
    end)
  end

  # ---------------------------------------------------------------------------
  # MCTS core: Select
  # ---------------------------------------------------------------------------

  # Walk the tree using UCB1 to find the most promising leaf to expand.
  # Returns {node_id, updated_tree}.
  defp select(tree, node_id, ucb_c) do
    node = Map.fetch!(tree, node_id)

    cond do
      # Leaf with untried actions — expand here
      node.untried_actions != [] ->
        {node_id, tree}

      # Leaf with no children — terminal, return as-is
      node.children_ids == [] ->
        {node_id, tree}

      # Internal node — pick best child by UCB1
      true ->
        best_child_id =
          node.children_ids
          |> Enum.max_by(fn cid ->
            child = Map.fetch!(tree, cid)
            ucb1(child, node.visits, ucb_c)
          end)

        select(tree, best_child_id, ucb_c)
    end
  end

  defp ucb1(%TreeNode{visits: 0}, _parent_visits, _c), do: 1.0e9

  defp ucb1(%TreeNode{visits: v, total_value: tv}, parent_visits, c) do
    exploitation = tv / v
    exploration = c * :math.sqrt(:math.log(max(parent_visits, 1)) / v)
    exploitation + exploration
  end

  # ---------------------------------------------------------------------------
  # MCTS core: Expand
  # ---------------------------------------------------------------------------

  # Pick one untried action from the node, apply it, create a child node.
  # Returns {child_id, updated_tree}.
  defp expand(tree, node_id) do
    node = Map.fetch!(tree, node_id)

    case node.untried_actions do
      [] ->
        {node_id, tree}

      [action | rest] ->
        new_state = apply_action(node.state, action)
        child_actions = generate_actions(new_state)
        child_id = make_ref()

        child = %TreeNode{
          id: child_id,
          state: new_state,
          action: action,
          parent_id: node_id,
          untried_actions: child_actions,
          visits: 0,
          total_value: 0.0
        }

        updated_parent = %{node | untried_actions: rest, children_ids: [child_id | node.children_ids]}

        tree
        |> Map.put(child_id, child)
        |> Map.put(node_id, updated_parent)
        |> then(&{child_id, &1})
    end
  end

  # ---------------------------------------------------------------------------
  # MCTS core: Rollout
  # ---------------------------------------------------------------------------

  defp simulate_rollout(tree, node_id) do
    node = Map.fetch!(tree, node_id)
    random_rollout(node.state, 0, 5)
  end

  defp random_rollout(state, depth, max_depth) do
    cond do
      depth >= max_depth -> evaluate_state(state)
      terminal?(state) -> evaluate_state(state)
      true ->
        actions = generate_actions(state)

        if actions == [] do
          evaluate_state(state)
        else
          action = Enum.random(actions)
          new_state = apply_action(state, action)
          random_rollout(new_state, depth + 1, max_depth)
        end
    end
  end

  defp terminal?(state) do
    nodes = Map.values(state.nodes)
    Enum.all?(nodes, fn n -> n.health >= 0.8 end) or
      Enum.all?(nodes, fn n -> n.health <= 0.2 end)
  end

  # ---------------------------------------------------------------------------
  # MCTS core: Backpropagate
  # ---------------------------------------------------------------------------

  defp backpropagate(tree, node_id, value) do
    node = Map.fetch!(tree, node_id)
    updated = %{node | visits: node.visits + 1, total_value: node.total_value + value}
    tree = Map.put(tree, node_id, updated)

    case node.parent_id do
      nil -> tree
      parent_id -> backpropagate(tree, parent_id, value)
    end
  end

  # ---------------------------------------------------------------------------
  # State transition
  # ---------------------------------------------------------------------------

  defp apply_action(state, %{type: :reassign, target_node: node}) do
    update_node_health(state, node, +0.3)
  end

  defp apply_action(state, %{type: :hire, target_node: node}) do
    update_node_health(state, node, +0.5)
  end

  defp apply_action(state, %{type: :restructure, target_node: node}) do
    connected = find_connected_nodes(node)

    state
    |> update_node_health(node, +0.4)
    |> spread_load(connected, -0.08)
  end

  defp apply_action(state, %{type: :pause, target_node: node}) do
    node_info = Map.get(state.nodes, node, %{revenue_exposed: false})

    if node_info[:revenue_exposed] do
      update_node_health(state, node, -0.15)
    else
      update_node_health(state, node, +0.1)
    end
  end

  defp apply_action(state, %{type: :pivot, target_node: node}) do
    delta = Enum.random([-0.2, 0.0, 0.2, 0.4, 0.6])
    update_node_health(state, node, delta)
  end

  defp apply_action(state, %{type: :redistribute, target_node: node}) do
    connected = find_connected_nodes(node)

    state
    |> update_node_health(node, +0.35)
    |> spread_load(connected, -0.05)
  end

  defp update_node_health(state, node_name, delta) do
    nodes =
      Map.update(state.nodes, node_name, %{health: max(0.0, delta), entities: [], revenue_exposed: false, context_count: 0}, fn info ->
        new_health = Float.round(min(1.0, max(0.0, info.health + delta)), 3)
        %{info | health: new_health}
      end)

    %{state | nodes: nodes, total_health: compute_total_health(nodes)}
  end

  defp spread_load(state, connected_nodes, delta) do
    Enum.reduce(connected_nodes, state, fn node, acc ->
      update_node_health(acc, node, delta)
    end)
  end

  defp find_connected_nodes(node_name) do
    # Adjacent nodes share people in the topology
    node_entities = Map.get(@topology, node_name, [])

    @topology
    |> Enum.filter(fn {other, entities} ->
      other != node_name and Enum.any?(entities, fn e -> e in node_entities end)
    end)
    |> Enum.map(fn {other, _} -> other end)
    |> Enum.take(3)
  end

  # ---------------------------------------------------------------------------
  # State evaluation
  # ---------------------------------------------------------------------------

  defp evaluate_state(%{nodes: nodes}) when map_size(nodes) == 0, do: 0.5

  defp evaluate_state(%{nodes: nodes}) do
    node_list = Map.values(nodes)

    revenue_nodes = Enum.filter(node_list, & &1.revenue_exposed)

    revenue_health =
      if revenue_nodes == [] do
        1.0
      else
        Enum.sum(Enum.map(revenue_nodes, & &1.health)) / length(revenue_nodes)
      end

    coverage = Enum.count(node_list, fn n -> n.health > 0.5 end) / max(length(node_list), 1)
    avg_health = Enum.sum(Enum.map(node_list, & &1.health)) / max(length(node_list), 1)

    0.4 * revenue_health + 0.3 * coverage + 0.3 * avg_health
  end

  # ---------------------------------------------------------------------------
  # Result extraction
  # ---------------------------------------------------------------------------

  defp extract_best_path(tree, root_id) do
    root = Map.fetch!(tree, root_id)
    best_child_id = best_child(tree, root)

    case best_child_id do
      nil ->
        %{actions: [], confidence: 0.0, outcome: build_outcome(root.state)}

      cid ->
        child = Map.fetch!(tree, cid)
        actions = collect_actions(tree, cid, [])
        confidence = if child.visits > 0, do: Float.round(child.total_value / child.visits, 3), else: 0.0

        %{actions: actions, confidence: confidence, outcome: build_outcome(child.state)}
    end
  end

  defp collect_actions(_tree, nil, acc), do: Enum.reverse(acc)

  defp collect_actions(tree, node_id, acc) do
    node = Map.fetch!(tree, node_id)

    case {node.action, best_child(tree, node)} do
      {nil, _} -> Enum.reverse(acc)
      {action, nil} -> Enum.reverse([action | acc])
      {action, child_id} -> collect_actions(tree, child_id, [action | acc])
    end
  end

  defp best_child(_tree, %TreeNode{children_ids: []}), do: nil

  defp best_child(tree, %TreeNode{children_ids: children}) do
    Enum.max_by(children, fn cid ->
      child = Map.fetch!(tree, cid)
      if child.visits > 0, do: child.total_value / child.visits, else: 0.0
    end)
  end

  defp build_outcome(state) do
    node_list = Map.values(state.nodes)

    revenue_nodes = Enum.filter(node_list, & &1.revenue_exposed)

    revenue_health =
      if revenue_nodes == [] do
        1.0
      else
        Float.round(Enum.sum(Enum.map(revenue_nodes, & &1.health)) / length(revenue_nodes), 3)
      end

    coverage =
      Float.round(Enum.count(node_list, fn n -> n.health > 0.5 end) / max(length(node_list), 1), 3)

    avg_health =
      Float.round(Enum.sum(Enum.map(node_list, & &1.health)) / max(length(node_list), 1), 3)

    overall = Float.round(0.4 * revenue_health + 0.3 * coverage + 0.3 * avg_health, 3)

    %{
      revenue_impact: revenue_health,
      coverage_score: coverage,
      health_score: avg_health,
      overall: overall
    }
  end

  defp extract_alternatives(tree, root_id, n) do
    root = Map.fetch!(tree, root_id)

    root.children_ids
    |> Enum.map(fn cid ->
      child = Map.fetch!(tree, cid)
      score = if child.visits > 0, do: Float.round(child.total_value / child.visits, 3), else: 0.0
      actions = collect_actions(tree, cid, [])
      %{sequence: actions, score: score}
    end)
    |> Enum.sort_by(& &1.score, :desc)
    |> Enum.drop(1)
    |> Enum.take(n)
  end

  defp compute_tree_stats(tree) do
    nodes = Map.values(tree)
    node_count = length(nodes)

    max_depth =
      nodes
      |> Enum.map(&depth_of_node(tree, &1.id, 0))
      |> Enum.max(fn -> 0 end)

    total_children = Enum.sum(Enum.map(nodes, fn n -> length(n.children_ids) end))
    internal_nodes = Enum.count(nodes, fn n -> n.children_ids != [] end)

    avg_branching =
      if internal_nodes == 0 do
        0.0
      else
        Float.round(total_children / internal_nodes, 1)
      end

    %{nodes: node_count, max_depth: max_depth, avg_branching: avg_branching}
  end

  defp depth_of_node(_tree, _node_id, depth) when depth > 20, do: depth

  defp depth_of_node(tree, node_id, depth) do
    node = Map.fetch!(tree, node_id)

    case node.children_ids do
      [] ->
        depth

      children ->
        children
        |> Enum.map(fn cid -> depth_of_node(tree, cid, depth + 1) end)
        |> Enum.max()
    end
  end
end
