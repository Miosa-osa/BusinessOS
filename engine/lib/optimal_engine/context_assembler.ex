defmodule OptimalEngine.ContextAssembler do
  @moduledoc """
  Tiered context assembly following OpenViking's hierarchy.

  ## Tiers (the correct model)

  - **L0** — Structural inventory: what exists in the library.
    Nodes, skills, resources, memory counts, system state.
    Always loaded. The agent sees "here's what's available."

  - **L1** — Per-file summaries/abstracts: one-liner per relevant file.
    Query-driven. BM25 + vector + graph fusion returns matching files
    with their L0 abstracts. The agent sees "here are the relevant files
    and what each one is about."

  - **L2** — Full content: loaded on demand for top results.
    Only for the files the agent actually needs to read in full.
    Deep retrieval with decision history.

  ## Usage

      # Full tiered assembly for a query
      {:ok, context} = ContextAssembler.assemble("AI Masters pricing")

      # Just the inventory
      {:ok, l0} = ContextAssembler.l0()

      # Just summaries for a query
      {:ok, l1} = ContextAssembler.l1("pricing")
  """

  require Logger

  alias OptimalEngine.{L0Cache, SearchEngine, Store}
  alias OptimalEngine.Bridge.Knowledge

  @default_budgets %{l0: 3_000, l1: 10_000, l2: 50_000}
  @rrf_k 60

  @type assembled :: %{
          l0: String.t(),
          l1: String.t(),
          l2: String.t(),
          total_tokens: non_neg_integer(),
          sources: [String.t()],
          search_scores: [map()]
        }

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Assembles full tiered context for a query.

  L0 = inventory (what exists)
  L1 = relevant file summaries (what matches)
  L2 = full content (what you need to read)
  """
  @spec assemble(String.t(), keyword()) :: {:ok, assembled()}
  def assemble(query, opts \\ []) do
    budgets = Keyword.get(opts, :tier_budgets, @default_budgets)

    # L0 — structural inventory (always loaded, query-independent)
    l0 = build_l0(budgets.l0)

    # L1 — per-file summaries for matching files (query-driven)
    {l1, search_scores} = build_l1(query, budgets.l1, opts)

    # L2 — full content for top results (deep retrieval)
    l2 = build_l2(query, search_scores, budgets.l2, opts)

    total_tokens = estimate_tokens(l0) + estimate_tokens(l1) + estimate_tokens(l2)
    sources = extract_sources(search_scores)

    {:ok,
     %{
       l0: l0,
       l1: l1,
       l2: l2,
       total_tokens: total_tokens,
       sources: sources,
       search_scores: search_scores
     }}
  end

  @doc "Returns L0 — the structural inventory of the library."
  @spec l0() :: {:ok, String.t()}
  def l0 do
    content = build_l0(@default_budgets.l0)
    {:ok, content}
  end

  @doc "Returns L1 — per-file summaries for a query."
  @spec l1(String.t(), keyword()) :: {:ok, String.t()}
  def l1(query, opts \\ []) do
    budget = Keyword.get(opts, :budget, @default_budgets.l1)
    {content, _scores} = build_l1(query, budget, opts)
    {:ok, content}
  end

  @doc """
  Performs Reciprocal Rank Fusion across BM25 and graph-boosted results.
  """
  @spec fused_search(String.t(), keyword()) :: {:ok, [map()]}
  def fused_search(query, opts \\ []) do
    limit = Keyword.get(opts, :limit, 20)

    bm25_results =
      case SearchEngine.search(query, limit: limit) do
        {:ok, results} -> results
        _ -> []
      end

    graph_results = Knowledge.graph_boost(bm25_results, query)

    fused = reciprocal_rank_fusion([bm25_results, graph_results])

    {:ok, Enum.take(fused, limit)}
  end

  # ---------------------------------------------------------------------------
  # Private: Tier Builders
  # ---------------------------------------------------------------------------

  # L0 — Structural inventory from cache
  defp build_l0(budget) do
    content = L0Cache.get()
    truncate_to_budget(content, budget)
  end

  # L1 — Per-file abstracts/summaries for relevant files
  defp build_l1(query, budget, opts) do
    limit = Keyword.get(opts, :limit, 20)

    {:ok, results} = fused_search(query, limit: limit)

    # Build L1 as a list of file summaries — one entry per matching file
    {content, _} =
      Enum.reduce(results, {"", 0}, fn result, {acc, tokens} ->
        title = Map.get(result, :title, "Untitled")
        node = Map.get(result, :node, "")
        score = Map.get(result, :score, 0)
        l0_abstract = Map.get(result, :l0_abstract, "") || ""
        uri = Map.get(result, :uri, "")
        type = Map.get(result, :type, :signal)

        # L1 shows the abstract (one-liner) per file, not the full overview
        entry =
          "- **#{title}** [#{type} | #{node} | score: #{Float.round(score, 3)}]\n" <>
            "  #{l0_abstract}\n" <>
            if(uri != "", do: "  `#{uri}`\n", else: "")

        entry_tokens = estimate_tokens(entry)

        if tokens + entry_tokens <= budget do
          {acc <> entry, tokens + entry_tokens}
        else
          {acc, tokens}
        end
      end)

    header = "## Matching Files (#{length(results)} results)\n\n"

    scores =
      Enum.map(results, fn r ->
        %{
          id: Map.get(r, :id),
          title: Map.get(r, :title),
          node: Map.get(r, :node),
          score: Map.get(r, :score),
          uri: Map.get(r, :uri)
        }
      end)

    {header <> content, scores}
  end

  # L2 — Full content for top results
  defp build_l2(_query, search_scores, budget, _opts) do
    top_ids =
      search_scores
      |> Enum.take(5)
      |> Enum.map(& &1.id)
      |> Enum.reject(&is_nil/1)

    {content, _} =
      Enum.reduce(top_ids, {"", 0}, fn id, {acc, tokens} ->
        case Store.get_context(id) do
          {:ok, ctx} ->
            full_text = Map.get(ctx, :content, "") || ""
            title = Map.get(ctx, :title, "")
            node = Map.get(ctx, :node, "")
            uri = Map.get(ctx, :uri, "")

            entry =
              "## #{title} (#{node})\n" <>
                if(uri != "", do: "> `#{uri}`\n\n", else: "\n") <>
                full_text <> "\n\n---\n"

            entry_tokens = estimate_tokens(entry)

            if tokens + entry_tokens <= budget do
              {acc <> entry, tokens + entry_tokens}
            else
              {acc, tokens}
            end

          _ ->
            {acc, tokens}
        end
      end)

    # Decision history for related topics
    decision_content = load_related_decisions(top_ids, budget - estimate_tokens(content))
    content <> decision_content
  end

  # ---------------------------------------------------------------------------
  # Private: Reciprocal Rank Fusion
  # ---------------------------------------------------------------------------

  defp reciprocal_rank_fusion(result_lists) do
    score_map =
      result_lists
      |> Enum.reduce(%{}, fn results, acc ->
        results
        |> Enum.with_index(1)
        |> Enum.reduce(acc, fn {result, rank}, inner_acc ->
          id = Map.get(result, :id, make_ref())
          rrf_score = 1.0 / (@rrf_k + rank)

          Map.update(inner_acc, id, {rrf_score, result}, fn {existing_score, existing_result} ->
            {existing_score + rrf_score, existing_result}
          end)
        end)
      end)

    score_map
    |> Enum.map(fn {_id, {fused_score, result}} ->
      %{result | score: Float.round(fused_score, 6)}
    end)
    |> Enum.sort_by(& &1.score, :desc)
  end

  # ---------------------------------------------------------------------------
  # Private: Helpers
  # ---------------------------------------------------------------------------

  defp load_related_decisions(context_ids, remaining_budget) when remaining_budget > 0 do
    placeholders =
      context_ids
      |> Enum.with_index(1)
      |> Enum.map_join(", ", fn {_, i} -> "?#{i}" end)

    if placeholders == "" do
      ""
    else
      sql = """
      SELECT topic, decision, rationale, decided_at
      FROM decisions
      WHERE context_id IN (#{placeholders})
      ORDER BY decided_at DESC
      LIMIT 5
      """

      case Store.raw_query(sql, context_ids) do
        {:ok, rows} when rows != [] ->
          lines =
            Enum.map(rows, fn [topic, decision, rationale, date] ->
              "- **#{topic}** (#{date}): #{decision}\n  _Rationale: #{rationale}_"
            end)

          "\n## Related Decisions\n\n" <> Enum.join(lines, "\n")

        _ ->
          ""
      end
    end
  end

  defp load_related_decisions(_, _), do: ""

  defp extract_sources(scores) do
    scores
    |> Enum.map(fn s -> s[:uri] || s[:title] || "unknown" end)
    |> Enum.uniq()
  end

  defp estimate_tokens(text) when is_binary(text), do: div(String.length(text), 4)
  defp estimate_tokens(_), do: 0

  defp truncate_to_budget(text, budget) do
    max_chars = budget * 4

    if String.length(text) > max_chars do
      String.slice(text, 0, max_chars) <> "\n\n[...truncated at token budget]"
    else
      text
    end
  end
end
