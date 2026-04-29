defmodule OptimalEngine.Wiki do
  @moduledoc """
  Top-level facade for Tier 3 — the LLM-maintained wiki layer.

  A wiki page is a curated, audience-aware summary with hot citations
  back to chunks (Tier 2) and raw sources (Tier 1). Pages are the
  agent's "front door": `mix optimal.rag` consults the wiki first and
  only falls through to hybrid retrieval on wiki miss.

  Public API:

      Wiki.get(tenant_id, slug, audience)        — read latest version
      Wiki.put(page)                              — persist a new version
      Wiki.list(tenant_id)                        — list every page
      Wiki.curate(slug, citations, opts)          — run the curator
      Wiki.render(page, resolver, format)         — render directives
      Wiki.verify(page)                           — run integrity checks
      Wiki.verify_against_schema(page, schema)    — schema-enforced checks
  """

  alias OptimalEngine.Wiki.{Curator, Directives, Integrity, Page, Store}

  defdelegate from_markdown(markdown, opts), to: Page
  defdelegate to_markdown(page), to: Page

  defdelegate put(page), to: Store
  defdelegate latest(tenant_id, slug, audience), to: Store
  defdelegate list(tenant_id), to: Store

  @doc "Render a page's body, resolving every directive via the supplied resolver."
  @spec render(Page.t(), Directives.resolver(), keyword()) :: {String.t(), [String.t()]}
  def render(%Page{} = page, resolver, opts \\ []) do
    Directives.render(page.body, resolver, opts)
  end

  @doc "Verify a page's integrity (citations, directive verbs, page size, uncited claims)."
  @spec verify(Page.t(), keyword()) :: Integrity.report()
  def verify(%Page{} = page, opts \\ []), do: Integrity.check(page, opts)

  @doc "Verify against the `.wiki/SCHEMA.md` schema rules (required sections, max size, etc.)."
  @spec verify_against_schema(Page.t(), map(), keyword()) :: Integrity.report()
  def verify_against_schema(%Page{} = page, schema, opts \\ []) do
    Integrity.against_schema(page, schema, opts)
  end

  @doc "Curate a page with new citations. See `Curator.curate/3`."
  @spec curate(Page.t(), [Curator.citation()], keyword()) :: Curator.outcome()
  def curate(%Page{} = page, citations, opts \\ []) do
    Curator.curate(page, citations, opts)
  end
end
