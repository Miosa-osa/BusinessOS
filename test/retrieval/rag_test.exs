defmodule OptimalEngine.Retrieval.RAGTest do
  use ExUnit.Case, async: false

  alias OptimalEngine.Retrieval
  alias OptimalEngine.Retrieval.{Receiver, RAG}
  alias OptimalEngine.Wiki.{Page, Store}

  describe "ask/2 — wiki-first path" do
    test "returns source: :wiki when a curated page exists" do
      suffix = System.unique_integer([:positive])
      slug = "rag-wiki-hit-#{suffix}"

      :ok =
        Store.put(%Page{
          tenant_id: "default",
          slug: slug,
          audience: "default",
          version: 1,
          frontmatter: %{"slug" => slug},
          body: "## Summary\n\nThe answer {{cite: optimal://a}}."
        })

      receiver = Receiver.new(%{format: :markdown, audience: "default"})
      {:ok, result} = RAG.ask(slug, receiver: receiver, skip_intent: true)

      assert result.source == :wiki
      assert result.trace.wiki_hit? == true
      assert result.envelope.body =~ "The answer"
      assert "optimal://a" in result.envelope.sources
    end
  end

  describe "ask/2 — chunks fallback path" do
    test "falls through to hybrid search when no wiki hit" do
      receiver = Receiver.new(%{format: :plain, token_budget: 200})

      {:ok, result} =
        RAG.ask("some_nonexistent_wiki_query_#{System.unique_integer([:positive])}",
          receiver: receiver,
          skip_intent: true
        )

      refute result.trace.wiki_hit?
      assert result.source in [:chunks, :empty]
    end

    test "empty receiver returns a format-appropriate empty envelope" do
      receiver = Receiver.new(%{format: :markdown})

      {:ok, result} =
        RAG.ask("this-query-matches-nothing-#{System.unique_integer([:positive])}",
          receiver: receiver,
          skip_intent: true,
          skip_wiki: true
        )

      assert result.envelope.format == :markdown
      assert is_binary(result.envelope.body)
    end
  end

  describe "ask/2 — trace fields" do
    test "includes timing and counts" do
      {:ok, result} =
        RAG.ask("anything", skip_intent: true, skip_wiki: true)

      assert is_integer(result.trace.elapsed_ms)
      assert result.trace.elapsed_ms >= 0
      assert is_integer(result.trace.n_candidates)
      assert is_integer(result.trace.n_delivered)
      assert is_boolean(result.trace.truncated?)
    end
  end

  describe "Retrieval facade" do
    test "Retrieval.ask/2 delegates to RAG.ask/2" do
      {:ok, r1} = Retrieval.ask("x", skip_intent: true, skip_wiki: true)
      assert Map.has_key?(r1, :source)
      assert Map.has_key?(r1, :envelope)
      assert Map.has_key?(r1, :trace)
    end

    test "Retrieval.plan/2 delegates to BandwidthPlanner" do
      plan = Retrieval.plan([%{content: "hi", score: 1.0}], 10_000)
      assert length(plan.kept) == 1
    end

    test "Retrieval.anonymous_receiver/1 returns a receiver" do
      r = Retrieval.anonymous_receiver()
      assert r.kind == :unknown
    end
  end
end
