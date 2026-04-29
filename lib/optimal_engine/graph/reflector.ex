defmodule OptimalEngine.Graph.Reflector do
  @moduledoc """
  Scans for entity co-occurrences not captured as edges in the knowledge graph.

  When two entities appear together in multiple contexts but have no direct edge,
  that's a missing relationship. The Reflector finds these gaps and suggests
  new edges to strengthen the graph.

  Stateless module — no GenServer. Returns `{:ok, [suggestion()]}` or `{:error, reason}`.

  ## Suggestion shape

      %{
        source: "Alice",
        target: "Alice",
        cooccurrences: 4,
        confidence: 0.8,
        suggested_relation: "works_with",
        reason: "Co-occur in 4 contexts without direct edge"
      }

  ## Usage

      {:ok, suggestions} = OptimalEngine.Graph.Reflector.reflect(min_cooccurrences: 2, limit: 20)
      {:ok, contexts}    = OptimalEngine.Graph.Reflector.shared_contexts("Alice", "Alice")
  """

  require Logger
  alias OptimalEngine.Embed.Ollama, as: Ollama
  alias OptimalEngine.Store

  @valid_relations ~w(works_with reports_to related_to depends_on)

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Finds entity pairs that frequently co-occur without a direct edge.

  ## Options

  - `:min_cooccurrences` — minimum shared contexts to qualify (default: 2)
  - `:limit` — max suggestions to return (default: 20)

  Returns `{:ok, [map()]}`. Never raises — logs and returns `{:ok, []}` on failure.
  """
  @spec reflect(keyword()) :: {:ok, [map()]}
  def reflect(opts \\ []) do
    min_cooccurrences = Keyword.get(opts, :min_cooccurrences, 2)
    limit = Keyword.get(opts, :limit, 20)

    sql = """
    SELECT e1.name as entity_a, e2.name as entity_b, COUNT(DISTINCT e1.context_id) as cooccurrences
    FROM entities e1
    JOIN entities e2 ON e1.context_id = e2.context_id AND e1.name < e2.name
    GROUP BY e1.name, e2.name
    HAVING COUNT(DISTINCT e1.context_id) >= ?1
    ORDER BY cooccurrences DESC
    LIMIT ?2
    """

    with {:ok, rows} <- Store.raw_query(sql, [min_cooccurrences, limit * 3]) do
      suggestions =
        rows
        |> Enum.map(fn [a, b, count] -> {a, b, count} end)
        |> Enum.reject(fn {a, b, _} -> edge_exists?(a, b) end)
        |> Enum.take(limit)
        |> Enum.map(&build_suggestion/1)

      {:ok, suggestions}
    end
  rescue
    err ->
      Logger.warning("[Reflector] reflect/1 failed: #{inspect(err)}")
      {:ok, []}
  end

  @doc """
  Returns the contexts where both `entity_a` and `entity_b` appear together.

  Useful for understanding why the Reflector flagged a pair.
  """
  @spec shared_contexts(String.t(), String.t()) :: {:ok, [map()]} | {:error, term()}
  def shared_contexts(entity_a, entity_b) do
    sql = """
    SELECT DISTINCT c.id, c.title, c.node
    FROM contexts c
    JOIN entities e1 ON e1.context_id = c.id AND e1.name = ?1
    JOIN entities e2 ON e2.context_id = c.id AND e2.name = ?2
    """

    case Store.raw_query(sql, [entity_a, entity_b]) do
      {:ok, rows} ->
        contexts = Enum.map(rows, fn [id, title, node] -> %{id: id, title: title, node: node} end)
        {:ok, contexts}

      err ->
        err
    end
  end

  # ---------------------------------------------------------------------------
  # Private helpers
  # ---------------------------------------------------------------------------

  defp build_suggestion({a, b, count}) do
    base = %{
      source: a,
      target: b,
      cooccurrences: count,
      confidence: min(count / 5.0, 1.0),
      reason: "Co-occur in #{count} contexts without direct edge"
    }

    if Ollama.available?() do
      classify_relationship(base)
    else
      Map.put(base, :suggested_relation, "related")
    end
  end

  defp edge_exists?(a, b) do
    sql = """
    SELECT COUNT(*) FROM edges
    WHERE (source_id = ?1 AND target_id = ?2) OR (source_id = ?2 AND target_id = ?1)
    """

    case Store.raw_query(sql, [a, b]) do
      {:ok, [[count]]} -> count > 0
      _ -> false
    end
  end

  defp classify_relationship(suggestion) do
    prompt = """
    Two entities frequently co-occur: "#{suggestion.source}" and "#{suggestion.target}" \
    appear together in #{suggestion.cooccurrences} contexts.

    What type of relationship likely exists? Choose one:
    - works_with: professional collaboration
    - reports_to: hierarchical relationship
    - related_to: topical/domain relationship
    - depends_on: dependency relationship

    Reply with ONLY the relationship type, nothing else.
    """

    case Ollama.generate(prompt, system: "You classify relationships. Reply with only the type.") do
      {:ok, response} ->
        relation =
          response
          |> String.trim()
          |> String.downcase()
          |> String.replace(~r/[^a-z_]/, "")

        rel = if relation in @valid_relations, do: relation, else: "related"
        Map.put(suggestion, :suggested_relation, rel)

      _ ->
        Map.put(suggestion, :suggested_relation, "related")
    end
  end
end
