package optimal

import (
	"math"
	"math/rand/v2"
)

// MCTSResult is the output of a Monte Carlo Tree Search over a SimulationResult.
// It recommends the best response action and ranks all candidate actions.
type MCTSResult struct {
	BestAction string   `json:"best_action"`
	Score      float64  `json:"score"`
	Visits     int      `json:"visits"`
	Actions    []Action `json:"actions"`
}

// Action is a single candidate response with its UCB1-derived score and visit count.
type Action struct {
	Name   string  `json:"name"`
	Score  float64 `json:"score"`
	Visits int     `json:"visits"`
}

// candidateActions are the fixed response types evaluated by the tree search.
// Matches the Elixir MCTS action set.
var candidateActions = []string{
	"reassign",
	"hire",
	"restructure",
	"pause",
	"pivot",
	"redistribute",
}

// ucbC is the exploration constant for UCB1 (√2 is the standard choice).
const ucbC = math.Sqrt2

// mctsNode is an internal tree node.
type mctsNode struct {
	action     string
	visits     int
	totalScore float64
	children   []*mctsNode
	parent     *mctsNode
}

func (n *mctsNode) avgScore() float64 {
	if n.visits == 0 {
		return 0
	}
	return n.totalScore / float64(n.visits)
}

// ucb1 returns the Upper Confidence Bound score for node n given total parent visits.
func (n *mctsNode) ucb1(totalVisits int) float64 {
	if n.visits == 0 {
		return math.MaxFloat64 // force exploration of unvisited nodes
	}
	exploitation := n.avgScore()
	exploration := ucbC * math.Sqrt(math.Log(float64(totalVisits))/float64(n.visits))
	return exploitation + exploration
}

// PlanResponse uses Monte Carlo Tree Search to find the best response action
// for a given SimulationResult. It runs budget iterations of the UCB1 loop
// (select → expand → rollout → backpropagate) over the fixed action set.
//
// When budget <= 0 it defaults to 32 (matching the Elixir engine default).
// When result is nil an empty result is returned.
func PlanResponse(result *SimulationResult, budget int) *MCTSResult {
	if budget <= 0 {
		budget = 32
	}
	if result == nil {
		return &MCTSResult{Actions: []Action{}}
	}

	// Build a flat root with one child per candidate action.
	root := &mctsNode{action: "root"}
	for _, a := range candidateActions {
		root.children = append(root.children, &mctsNode{action: a, parent: root})
	}

	for i := 0; i < budget; i++ {
		// 1. SELECT: pick the child with the highest UCB1 score.
		node := selectNode(root)

		// 2. ROLLOUT: simulate the action's effect and score it.
		score := rollout(node.action, result)

		// 3. BACKPROPAGATE: update visit counts and scores up the tree.
		backpropagate(node, score)
	}

	// Collect action results, find the best.
	actions := make([]Action, 0, len(root.children))
	var best *mctsNode
	for _, child := range root.children {
		actions = append(actions, Action{
			Name:   child.action,
			Score:  child.avgScore(),
			Visits: child.visits,
		})
		if best == nil || child.avgScore() > best.avgScore() {
			best = child
		}
	}

	if best == nil {
		return &MCTSResult{Actions: actions}
	}

	return &MCTSResult{
		BestAction: best.action,
		Score:      best.avgScore(),
		Visits:     best.visits,
		Actions:    actions,
	}
}

// ── MCTS phases ───────────────────────────────────────────────────────────────

// selectNode picks the child of root with the highest UCB1 score.
// The tree is only one level deep (root → actions), so no recursive descent
// is needed — we always select directly from root's children.
func selectNode(root *mctsNode) *mctsNode {
	if len(root.children) == 0 {
		return root
	}
	best := root.children[0]
	bestScore := best.ucb1(root.visits + 1)
	for _, child := range root.children[1:] {
		if s := child.ucb1(root.visits + 1); s > bestScore {
			bestScore = s
			best = child
		}
	}
	return best
}

// rollout simulates the expected effectiveness of an action against the impact
// result using a heuristic scoring model. Returns a score in [0, 1].
//
// The scoring model:
//   - Base score depends on action type and total impact magnitude.
//   - A random perturbation (±15%) models outcome uncertainty.
//   - Scores are clamped to [0, 1].
func rollout(action string, result *SimulationResult) float64 {
	total := result.TotalImpact
	n := float64(len(result.AffectedNodes))
	if n == 0 {
		n = 1
	}
	avgImpact := total / n

	// Base effectiveness of each action type against the impact profile.
	var base float64
	switch action {
	case "reassign":
		// Works well when impact is concentrated on a few nodes.
		base = 0.7 - 0.3*total
	case "hire":
		// Replaces capacity; effective when total impact is high.
		base = 0.5 + 0.3*total
	case "restructure":
		// Highest ceiling but high variance; effective for broad impact.
		base = 0.4 + 0.4*(1-avgImpact)
	case "pause":
		// Safe option; always scores mid-range.
		base = 0.55
	case "pivot":
		// High risk, high reward; favoured when impact is catastrophic.
		base = 0.3 + 0.5*total
	case "redistribute":
		// Spreads the load; works best with many low-impact nodes.
		base = 0.6 - 0.2*avgImpact
	default:
		base = 0.5
	}

	// Stochastic perturbation (±15%).
	perturbation := (rand.Float64() - 0.5) * 0.30
	score := base + perturbation

	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}
	return score
}

// backpropagate walks from node to root, incrementing visit counts and
// accumulating the rollout score.
func backpropagate(node *mctsNode, score float64) {
	for n := node; n != nil; n = n.parent {
		n.visits++
		n.totalScore += score
	}
}
