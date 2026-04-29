defmodule OptimalEngine.Retrieval.BandwidthPlanner do
  @moduledoc """
  Plans which retrieved items survive a receiver's token budget.

  RAG almost always produces more relevant chunks than the receiver
  can swallow. The planner is the place we enforce Shannon's ceiling:
  pack the highest-scoring items until the budget is exhausted, then
  stop — with a loud `:truncated` count so the caller knows what was
  dropped.

  ## Inputs

      items :: [%{score: float, content: String.t(), uri: String.t() | nil}]

  Any map that carries `:content` and `:score` works; surplus keys are
  preserved so the caller can thread metadata through.

  ## Output

      %{
        kept:        [item],
        dropped:     [item],
        used_tokens: integer,
        budget:      integer,
        truncated?:  boolean
      }

  ## Token estimation

  We use the 4-chars-per-token heuristic — same as the
  `ContextAssembler`. This is intentionally conservative; swap in a
  real tokenizer before shipping to paid LLMs.
  """

  @chars_per_token 4
  @reserve_overhead 64

  @type item :: %{required(:content) => String.t(), optional(any()) => any()}
  @type plan :: %{
          kept: [item()],
          dropped: [item()],
          used_tokens: non_neg_integer(),
          budget: non_neg_integer(),
          truncated?: boolean()
        }

  @doc """
  Pack `items` greedily by descending score, honoring `budget`.

  Items without an explicit `:score` are treated as 0 and ordered
  last. A per-item overhead of #{@reserve_overhead} tokens is reserved
  to cover the citation/source wrapper the Deliver layer adds.
  """
  @spec plan([item()], non_neg_integer()) :: plan()
  def plan(items, budget) when is_list(items) and is_integer(budget) and budget >= 0 do
    sorted = Enum.sort_by(items, &Map.get(&1, :score, 0.0), :desc)

    {kept_rev, dropped_rev, used} =
      Enum.reduce(sorted, {[], [], 0}, fn item, {kept, dropped, tokens} ->
        cost = estimate_tokens(item) + @reserve_overhead

        if tokens + cost <= budget do
          {[item | kept], dropped, tokens + cost}
        else
          {kept, [item | dropped], tokens}
        end
      end)

    %{
      kept: Enum.reverse(kept_rev),
      dropped: Enum.reverse(dropped_rev),
      used_tokens: used,
      budget: budget,
      truncated?: dropped_rev != []
    }
  end

  @doc """
  Estimate the token cost of an item. Uses the 4-char heuristic on the
  `:content` field. Nil/empty content = 0.
  """
  @spec estimate_tokens(item() | String.t() | nil) :: non_neg_integer()
  def estimate_tokens(nil), do: 0
  def estimate_tokens(%{content: content}), do: estimate_tokens(content)
  def estimate_tokens(text) when is_binary(text), do: div(String.length(text), @chars_per_token)
  def estimate_tokens(_), do: 0

  @doc "Hard-truncate a single string to fit a token budget."
  @spec truncate(String.t(), non_neg_integer()) :: String.t()
  def truncate(text, budget) when is_binary(text) and is_integer(budget) and budget >= 0 do
    max_chars = budget * @chars_per_token

    cond do
      budget == 0 -> ""
      String.length(text) <= max_chars -> text
      true -> String.slice(text, 0, max_chars) <> "…"
    end
  end
end
