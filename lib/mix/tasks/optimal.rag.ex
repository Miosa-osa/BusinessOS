defmodule Mix.Tasks.Optimal.Rag do
  @shortdoc "Ask the engine a question (wiki-first, hybrid-fallback, LLM-ready)"

  @moduledoc """
  End-to-end retrieval for LLM consumption.

  Flow: Intent → WikiFirst (Tier 3) → hybrid search fallback → BandwidthPlanner
  → Deliver (receiver-shaped envelope).

  Where `mix optimal.search` returns metadata for humans, `mix optimal.rag`
  returns a payload ready to be fed to a language model.

  ## Usage

      mix optimal.rag "Q4 pricing decision"
      mix optimal.rag "pricing" --audience sales --format claude
      mix optimal.rag "pricing" --bandwidth small
      mix optimal.rag "pricing" --principal user:ada@acme.com
      mix optimal.rag "pricing" --skip-wiki --trace

  ## Options

    --format      plain | markdown | claude | openai | text  (default: markdown; `text` alias of `plain`)
    --bandwidth   small | medium | large                     (default: medium)
    --audience    wiki audience tag                          (default: default)
    --tenant      tenant id                                  (default: default)
    --principal   hydrate receiver from a principal id
    --skip-wiki   force hybrid fallback (debugging)
    --trace       print the trace block alongside the envelope
    --hybrid-limit max chunks pulled on wiki miss            (default: 20)

  ## Examples

      # Drop into a Claude Agent SDK prompt
      $ ctx=$(mix optimal.rag "our pricing conversation with Ed" --format claude)
      $ claude-cli "Given this: $ctx — what should our counter-offer be?"

      # Pipe to a local LLM
      $ mix optimal.rag "bug report" --format plain | ollama run qwen3:8b
  """

  use Mix.Task

  alias OptimalEngine.Retrieval
  alias OptimalEngine.Retrieval.Receiver

  @impl Mix.Task
  def run(args) do
    Mix.Task.run("app.start")

    {parsed, rest, _} =
      OptionParser.parse(args,
        strict: [
          format: :string,
          bandwidth: :string,
          audience: :string,
          tenant: :string,
          principal: :string,
          skip_wiki: :boolean,
          trace: :boolean,
          hybrid_limit: :integer
        ],
        aliases: [f: :format, a: :audience, b: :bandwidth]
      )

    query = Enum.join(rest, " ")

    if query == "" do
      Mix.shell().error(~s(Usage: mix optimal.rag "<query>" [options]))
      System.halt(1)
    end

    receiver = build_receiver(parsed)

    {:ok, result} =
      Retrieval.ask(query,
        receiver: receiver,
        skip_wiki: Keyword.get(parsed, :skip_wiki, false),
        hybrid_limit: Keyword.get(parsed, :hybrid_limit, 20)
      )

    IO.puts(result.envelope.body)

    if result.envelope.sources != [] do
      IO.puts("\n— sources —")
      Enum.each(result.envelope.sources, fn u -> IO.puts("  • #{u}") end)
    end

    if Keyword.get(parsed, :trace, false) do
      IO.puts("\n— trace —")
      IO.puts("  source:       #{result.source}")
      IO.puts("  wiki_hit?:    #{result.trace.wiki_hit?}")
      IO.puts("  candidates:   #{result.trace.n_candidates}")
      IO.puts("  delivered:    #{result.trace.n_delivered}")
      IO.puts("  truncated?:   #{result.trace.truncated?}")
      IO.puts("  elapsed_ms:   #{result.trace.elapsed_ms}")
    end
  end

  defp build_receiver(parsed) do
    overrides =
      parsed
      |> Enum.reduce(%{}, fn
        {:format, v}, acc -> Map.put(acc, :format, parse_format(v))
        {:bandwidth, v}, acc -> Map.put(acc, :bandwidth, String.to_existing_atom(v))
        {:audience, v}, acc -> Map.put(acc, :audience, v)
        {:tenant, v}, acc -> Map.put(acc, :tenant_id, v)
        _, acc -> acc
      end)

    case Keyword.get(parsed, :principal) do
      nil ->
        Receiver.new(overrides)

      principal_id ->
        case Receiver.from_principal(principal_id, Enum.to_list(overrides)) do
          {:ok, r} -> r
          {:error, _} -> Mix.raise("Principal not found: #{principal_id}")
        end
    end
  end

  # `text` is accepted as an alias for `plain` so old shell scripts keep working.
  defp parse_format("text"), do: :plain
  defp parse_format(v), do: String.to_existing_atom(v)
end
