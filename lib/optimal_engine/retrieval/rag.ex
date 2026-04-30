defmodule OptimalEngine.Retrieval.RAG do
  @moduledoc """
  End-to-end Retrieval-Augmented Generation flow.

      query + receiver
         │
         ▼
      [1] IntentAnalyzer   — expand query, extract entities
         │
         ▼
      [2] WikiFirst        — Tier 3 lookup (curated page?)
         │
         ├──hit──▶ [3a] Deliver.render_wiki   ──▶ envelope
         │
         └─miss─▶ [3b] Search (hybrid, principal-filtered)
                   │
                   ▼
                 [4] BandwidthPlanner.plan
                   │
                   ▼
                 [5] Deliver.render_chunks    ──▶ envelope

  One function — `ask/2` — is the public entry point. Everything
  else is plumbing.

  ## Return shape

      {:ok, %{
        source: :wiki | :chunks | :empty,
        envelope: %{body, format, sources, warnings},
        trace: %{
          intent: map(),
          wiki_hit?: boolean(),
          n_candidates: non_neg_integer(),
          n_delivered: non_neg_integer(),
          truncated?: boolean(),
          elapsed_ms: non_neg_integer()
        }
      }}

  The `trace` field is the Wiener feedback loop — it tells the caller
  exactly what happened so the receiver (human or agent) can decide
  whether to dig deeper.
  """

  alias OptimalEngine.Retrieval.{
    BandwidthPlanner,
    Deliver,
    IntentAnalyzer,
    Receiver,
    Search,
    WikiFirst
  }

  alias OptimalEngine.Store
  alias OptimalEngine.Wiki.Directives

  require Logger

  @type ask_result :: %{
          source: :wiki | :chunks | :empty,
          envelope: Deliver.envelope(),
          trace: trace()
        }

  @type trace :: %{
          intent: map(),
          wiki_hit?: boolean(),
          n_candidates: non_neg_integer(),
          n_delivered: non_neg_integer(),
          truncated?: boolean(),
          elapsed_ms: non_neg_integer()
        }

  @default_hybrid_limit 20

  @doc """
  Ask the engine a question on behalf of `receiver`.

  Options:
    * `:receiver` — `%Receiver{}`; defaults to `Receiver.anonymous/0`
    * `:hybrid_limit` — max chunks pulled from hybrid search (default 20)
    * `:skip_wiki` — bypass Tier 3 entirely (testing / debugging)
    * `:skip_intent` — bypass intent expansion (testing)
    * `:tenant_id` — override tenant scope
    * `:workspace_id` — scope retrieval to a workspace (default `"default"`)
    * `:on_event` — `(event_atom, payload_map -> any)` — called at each pipeline
      stage boundary for streaming consumers. Existing callers that omit this
      opt see no behaviour change. The function is called synchronously in the
      ask process; keep it non-blocking (e.g. `send/2` to a listener PID).
      Events fired (in order): `:intent`, `:wiki_hit`, `:chunks`, `:composing`
  """
  @spec ask(String.t(), keyword()) :: {:ok, ask_result()}
  def ask(query, opts \\ []) when is_binary(query) do
    started = System.monotonic_time(:millisecond)
    on_event = Keyword.get(opts, :on_event, fn _event, _payload -> :ok end)

    receiver =
      case Keyword.get(opts, :receiver) do
        %Receiver{} = r -> r
        nil -> Receiver.anonymous()
      end

    intent =
      if Keyword.get(opts, :skip_intent, false) do
        %{expanded_query: query, intent_type: :lookup, key_entities: [], temporal_hint: :any}
      else
        case IntentAnalyzer.analyze(query) do
          {:ok, i} -> i
          _ -> %{expanded_query: query, intent_type: :lookup, key_entities: [], temporal_hint: :any}
        end
      end

    on_event.(:intent, %{
      intent_type: intent[:intent_type],
      expanded_query: intent[:expanded_query],
      key_entities: intent[:key_entities] || [],
      node_hints: intent[:node_hints] || []
    })

    workspace_id = Keyword.get(opts, :workspace_id, "default")

    wiki_result =
      if Keyword.get(opts, :skip_wiki, false) do
        :miss
      else
        WikiFirst.lookup(query, receiver.audience,
          tenant_id: receiver.tenant_id,
          workspace_id: workspace_id
        )
      end

    {source, envelope, counts} =
      case wiki_result do
        {:hit, page, _reason} ->
          on_event.(:wiki_hit, %{
            slug: page.slug,
            audience: page.audience,
            version: page.version
          })

          on_event.(:composing, %{format: receiver.format, bandwidth: receiver.bandwidth})
          {:wiki, deliver_wiki(page, receiver), %{candidates: 1, delivered: 1, truncated?: false}}

        :miss ->
          hybrid_deliver(query, intent, receiver,
            Keyword.merge(Keyword.put(opts, :workspace_id, workspace_id),
              on_event: on_event
            )
          )
      end

    elapsed = System.monotonic_time(:millisecond) - started

    {:ok,
     %{
       source: source,
       envelope: envelope,
       trace: %{
         intent: intent,
         wiki_hit?: source == :wiki,
         n_candidates: counts.candidates,
         n_delivered: counts.delivered,
         truncated?: counts.truncated?,
         elapsed_ms: elapsed
       }
     }}
  end

  # ─── private ─────────────────────────────────────────────────────────────

  defp deliver_wiki(page, receiver) do
    Deliver.render_wiki(page, receiver, &wiki_resolver/2)
  end

  defp hybrid_deliver(query, intent, receiver, opts) do
    on_event = Keyword.get(opts, :on_event, fn _event, _payload -> :ok end)
    limit = Keyword.get(opts, :hybrid_limit, @default_hybrid_limit)
    workspace_id = Keyword.get(opts, :workspace_id, "default")
    expanded = Map.get(intent, :expanded_query, query)

    search_opts =
      [limit: limit, workspace_id: workspace_id]
      |> maybe_put(:principal, receiver.id)

    candidates = safe_search(expanded, search_opts)

    items =
      Enum.map(candidates, fn c ->
        %{
          score: Map.get(c, :score, 0.0),
          content: Map.get(c, :content) || Map.get(c, :l1_description) || "",
          title: Map.get(c, :title) || "Untitled",
          uri: Map.get(c, :uri)
        }
      end)

    on_event.(:chunks, Enum.map(items, fn item ->
      %{slug: item.uri, score: item.score, scale: "chunk"}
    end))

    plan = BandwidthPlanner.plan(items, receiver.token_budget)

    on_event.(:composing, %{format: receiver.format, bandwidth: receiver.bandwidth})

    envelope =
      case plan.kept do
        [] -> Deliver.empty(receiver, "No results for `#{query}`.")
        kept -> Deliver.render_chunks(kept, receiver)
      end

    {:chunks, envelope,
     %{
       candidates: length(candidates),
       delivered: length(plan.kept),
       truncated?: plan.truncated?
     }}
  end

  # Resolver used when rendering a wiki page. Today we only resolve
  # `cite` (URIs pass straight through to the format wrapper) and
  # `include` (fetches the referenced context body from the store).
  # Other verbs degrade gracefully to the raw token via `render_fallback`.
  defp wiki_resolver(%Directives.Directive{verb: :cite, argument: uri}, _opts) do
    {:ok, "", %{uri: uri}}
  end

  defp wiki_resolver(%Directives.Directive{verb: :include, argument: uri}, _opts) do
    case fetch_context_body(uri) do
      {:ok, body} -> {:ok, body, %{uri: uri}}
      _ -> {:error, :not_found}
    end
  end

  defp wiki_resolver(%Directives.Directive{verb: :wikilink, argument: slug}, _opts) do
    {:ok, slug, %{}}
  end

  defp wiki_resolver(_, _opts), do: {:error, :unresolved}

  defp fetch_context_body(uri) when is_binary(uri) do
    case Store.raw_query(
           "SELECT content FROM contexts WHERE uri = ?1 LIMIT 1",
           [uri]
         ) do
      {:ok, [[body]]} when is_binary(body) -> {:ok, body}
      _ -> {:error, :not_found}
    end
  end

  defp maybe_put(kw, _k, nil), do: kw
  defp maybe_put(kw, k, v), do: Keyword.put(kw, k, v)

  # Hybrid search can stall (Ollama timeout, vector-store cold cache). We
  # catch the GenServer :timeout exit so a flaky downstream dependency
  # returns an empty result set rather than blowing up the whole ask.
  defp safe_search(query, opts) do
    try do
      case Search.search(query, opts) do
        {:ok, results} -> results
        _ -> []
      end
    catch
      :exit, {:timeout, _} ->
        Logger.warning("RAG: hybrid search timed out for query=#{inspect(query)}")
        []

      :exit, reason ->
        Logger.warning("RAG: hybrid search crashed: #{inspect(reason)}")
        []
    end
  end
end
