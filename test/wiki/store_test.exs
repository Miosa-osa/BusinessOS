defmodule OptimalEngine.Wiki.StoreTest do
  use ExUnit.Case, async: false

  alias OptimalEngine.Wiki.{Page, Store}

  test "put/1 + latest/3 round-trip" do
    unique = System.unique_integer([:positive])

    page = %Page{
      tenant_id: "default",
      slug: "round-trip-#{unique}",
      audience: "default",
      version: 1,
      frontmatter: %{"slug" => "round-trip-#{unique}"},
      body: "## Summary\n\nContent {{cite: optimal://x}}.",
      last_curated: DateTime.utc_now() |> DateTime.to_iso8601(),
      curated_by: "test"
    }

    assert :ok = Store.put(page)
    assert {:ok, fetched} = Store.latest(page.tenant_id, page.slug, page.audience)
    assert fetched.slug == page.slug
    assert fetched.body == page.body
    assert fetched.version == 1
  end

  test "list/1 returns latest per (slug, audience) pair" do
    suffix = System.unique_integer([:positive])

    page_v1 = %Page{
      tenant_id: "default",
      slug: "list-#{suffix}",
      audience: "default",
      version: 1,
      frontmatter: %{},
      body: "v1 body"
    }

    page_v2 = %{page_v1 | version: 2, body: "v2 body"}

    :ok = Store.put(page_v1)
    :ok = Store.put(page_v2)

    {:ok, pages} = Store.list("default")

    matching = Enum.filter(pages, &(&1.slug == "list-#{suffix}"))
    assert length(matching) == 1
    assert hd(matching).version == 2
    assert hd(matching).body == "v2 body"
  end

  test "put_citations/4 persists + clear_citations/3 removes" do
    suffix = System.unique_integer([:positive])
    slug = "cite-#{suffix}"
    tenant = "default"

    # Seed a chunk row so FK holds
    OptimalEngine.Store.raw_query(
      "INSERT OR IGNORE INTO chunks (id, tenant_id, signal_id, scale, text) VALUES (?1, ?2, 's', 'document', 't')",
      ["cite-chunk-#{suffix}", tenant]
    )

    assert :ok =
             Store.put_citations(tenant, slug, "default", [
               %{chunk_id: "cite-chunk-#{suffix}", text: "fact 1"}
             ])

    {:ok, [[count]]} =
      OptimalEngine.Store.raw_query(
        "SELECT COUNT(*) FROM citations WHERE wiki_slug = ?1",
        [slug]
      )

    assert count >= 1

    :ok = Store.clear_citations(tenant, slug, "default")

    {:ok, [[count_after]]} =
      OptimalEngine.Store.raw_query(
        "SELECT COUNT(*) FROM citations WHERE wiki_slug = ?1",
        [slug]
      )

    assert count_after == 0
  end
end
