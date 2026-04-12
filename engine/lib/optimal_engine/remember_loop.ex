defmodule OptimalEngine.RememberLoop do
  @moduledoc """
  Three-mode friction capture system for the OptimalOS knowledge base.

  Modes:
  - **Explicit**: Direct observation storage with auto-classification
  - **Contextual**: Scan recent sessions for correction/friction signals
  - **Session mining**: Bulk extract patterns from session transcripts

  Observations accumulate in the `observations` table. When 3+ observations
  share a category, they're flagged for escalation to RethinkEngine.

  Stateless module — no GenServer.
  """

  require Logger
  alias OptimalEngine.{MemoryExtractor, Ollama, Store}

  @valid_categories ~w(process people tool decision pattern friction)

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Mode 1: Explicit observation storage.

  Classifies the observation into a category and stores it.

  ## Examples

      RememberLoop.remember("always check for duplicates before inserting")
      RememberLoop.remember("Ed prefers email over Slack", category: "people")
  """
  @spec remember(String.t(), keyword()) :: {:ok, map()} | {:error, term()}
  def remember(observation, opts \\ []) do
    category = Keyword.get(opts, :category) || classify_observation(observation)
    confidence = Keyword.get(opts, :confidence, 0.6)
    source = Keyword.get(opts, :source, "explicit")

    sql = """
    INSERT INTO observations (category, content, confidence, source)
    VALUES (?1, ?2, ?3, ?4)
    """

    case Store.raw_query(sql, [category, observation, confidence, source]) do
      {:ok, _} ->
        result = %{
          category: category,
          content: observation,
          confidence: confidence,
          source: source,
          escalation: check_escalation(category)
        }

        Logger.info(
          "[RememberLoop] Stored observation: #{category} — #{String.slice(observation, 0, 50)}"
        )

        {:ok, result}

      err ->
        {:error, err}
    end
  rescue
    err ->
      Logger.warning("[RememberLoop] remember/2 failed: #{inspect(err)}")
      {:error, inspect(err)}
  end

  @doc """
  Mode 2: Contextual scan — find correction/friction signals in recent sessions.

  Looks for patterns like "no, not that", "actually", "wrong", "should have"
  in recent session content.
  """
  @spec contextual_scan(keyword()) :: {:ok, [map()]}
  def contextual_scan(opts \\ []) do
    limit = Keyword.get(opts, :limit, 50)

    sql = """
    SELECT id, title, content FROM contexts
    WHERE type = 'signal' AND content != ''
    ORDER BY created_at DESC
    LIMIT ?1
    """

    case Store.raw_query(sql, [limit]) do
      {:ok, rows} ->
        observations =
          rows
          |> Enum.flat_map(fn [_id, _title, content] -> extract_friction(content) end)
          |> Enum.take(20)

        stored =
          observations
          |> Enum.map(fn obs ->
            case remember(obs.content,
                   category: obs.category,
                   confidence: obs.confidence,
                   source: "contextual"
                 ) do
              {:ok, result} -> result
              _ -> nil
            end
          end)
          |> Enum.reject(&is_nil/1)

        {:ok, stored}

      _ ->
        {:ok, []}
    end
  rescue
    err ->
      Logger.warning("[RememberLoop] contextual_scan failed: #{inspect(err)}")
      {:ok, []}
  end

  @doc """
  Mode 3: Session mining — bulk extract patterns using MemoryExtractor.
  """
  @spec mine_sessions(keyword()) :: {:ok, [map()]}
  def mine_sessions(opts \\ []) do
    limit = Keyword.get(opts, :limit, 20)

    sessions_sql = """
    SELECT id, content FROM sessions
    WHERE content IS NOT NULL AND content != '' AND summary != ''
    ORDER BY started_at DESC
    LIMIT ?1
    """

    rows =
      case Store.raw_query(sessions_sql, [limit]) do
        {:ok, r} when r != [] ->
          r

        _ ->
          contexts_sql =
            "SELECT id, content FROM contexts WHERE content != '' ORDER BY created_at DESC LIMIT ?1"

          case Store.raw_query(contexts_sql, [limit]) do
            {:ok, r} -> r
            _ -> []
          end
      end

    observations =
      rows
      |> Enum.flat_map(fn [_id, content] ->
        case MemoryExtractor.extract(content) do
          {:ok, memories} ->
            Enum.map(memories, fn m ->
              %{
                content: m.content,
                category: memory_category_to_observation(m.category),
                confidence: m.confidence,
                source: "mined"
              }
            end)

          _ ->
            []
        end
      end)
      |> Enum.take(30)

    stored =
      observations
      |> Enum.map(fn obs ->
        case remember(obs.content,
               category: obs.category,
               confidence: obs.confidence,
               source: obs.source
             ) do
          {:ok, result} -> result
          _ -> nil
        end
      end)
      |> Enum.reject(&is_nil/1)

    {:ok, stored}
  rescue
    err ->
      Logger.warning("[RememberLoop] mine_sessions failed: #{inspect(err)}")
      {:ok, []}
  end

  @doc """
  Lists all observations, optionally filtered by category.
  """
  @spec list(keyword()) :: {:ok, [map()]}
  def list(opts \\ []) do
    category = Keyword.get(opts, :category)
    limit = Keyword.get(opts, :limit, 50)

    {sql, params} =
      if category do
        {
          "SELECT id, category, content, confidence, source, created_at FROM observations WHERE category = ?1 ORDER BY created_at DESC LIMIT ?2",
          [category, limit]
        }
      else
        {
          "SELECT id, category, content, confidence, source, created_at FROM observations ORDER BY created_at DESC LIMIT ?1",
          [limit]
        }
      end

    case Store.raw_query(sql, params) do
      {:ok, rows} ->
        obs =
          Enum.map(rows, fn [id, cat, content, conf, source, created] ->
            %{
              id: id,
              category: cat,
              content: content,
              confidence: conf,
              source: source,
              created_at: created
            }
          end)

        {:ok, obs}

      _ ->
        {:ok, []}
    end
  end

  @doc """
  Returns categories with 3+ observations (escalation candidates for RethinkEngine).
  """
  @spec escalation_candidates() :: {:ok, [map()]}
  def escalation_candidates do
    sql = """
    SELECT category, COUNT(*) as cnt, SUM(confidence) as total_confidence
    FROM observations
    GROUP BY category
    HAVING COUNT(*) >= 3
    ORDER BY total_confidence DESC
    """

    case Store.raw_query(sql, []) do
      {:ok, rows} ->
        candidates =
          Enum.map(rows, fn [cat, cnt, conf] ->
            %{
              category: cat,
              count: cnt,
              total_confidence: conf,
              ready_for_rethink: (conf || 0) >= 1.5
            }
          end)

        {:ok, candidates}

      _ ->
        {:ok, []}
    end
  end

  # ---------------------------------------------------------------------------
  # Private helpers
  # ---------------------------------------------------------------------------

  defp classify_observation(text) do
    if Ollama.available?() do
      llm_classify(text)
    else
      regex_classify(text)
    end
  end

  defp llm_classify(text) do
    prompt = """
    Classify this observation into exactly one category:
    - process: about how work should be done
    - people: about people, preferences, relationships
    - tool: about tools, commands, technical approaches
    - decision: about choices made or to be made
    - pattern: recurring patterns or anti-patterns
    - friction: things that went wrong or caused problems

    Observation: "#{text}"

    Reply with ONLY the category name.
    """

    case Ollama.generate(prompt,
           system: "Classify observations. Reply with only the category name."
         ) do
      {:ok, response} ->
        cat = response |> String.trim() |> String.downcase()
        if cat in @valid_categories, do: cat, else: "pattern"

      _ ->
        regex_classify(text)
    end
  rescue
    _ -> regex_classify(text)
  end

  defp regex_classify(text) do
    lower = String.downcase(text)

    cond do
      String.contains?(lower, ~w(wrong error fail broken bug)) -> "friction"
      String.contains?(lower, ~w(decided choose chose pick selected)) -> "decision"
      String.contains?(lower, ~w(always never prefer should must)) -> "process"
      String.contains?(lower, ~w(mix command tool api config)) -> "tool"
      Regex.match?(~r/[A-Z][a-z]+ (prefers|wants|needs|likes)/, text) -> "people"
      true -> "pattern"
    end
  end

  defp extract_friction(content) do
    friction_patterns = ~w(no\ not actually wrong should\ have instead not\ that mistake)

    content
    |> String.split(~r/(?<=[.!?])(?:\s+|\n)/)
    |> Enum.map(&String.trim/1)
    |> Enum.reject(&(String.length(&1) < 10))
    |> Enum.flat_map(fn sentence ->
      lower = String.downcase(sentence)

      if Enum.any?(friction_patterns, &String.contains?(lower, &1)) do
        [%{content: sentence, category: "friction", confidence: 0.5}]
      else
        []
      end
    end)
  end

  defp memory_category_to_observation(category) do
    case category do
      :decision -> "decision"
      :preference -> "people"
      :skill -> "tool"
      :fact -> "pattern"
      :relationship -> "people"
      :context -> "process"
      _ -> "pattern"
    end
  end

  defp check_escalation(category) do
    sql = "SELECT COUNT(*), SUM(confidence) FROM observations WHERE category = ?1"

    case Store.raw_query(sql, [category]) do
      {:ok, [[count, total_conf]]} when count >= 3 ->
        %{
          escalated: true,
          count: count,
          total_confidence: total_conf,
          ready_for_rethink: (total_conf || 0) >= 1.5
        }

      _ ->
        %{escalated: false}
    end
  end
end
