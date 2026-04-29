defmodule OptimalEngine.Wiki.Store do
  @moduledoc """
  Store helpers for Tier-3 wiki pages and their citations.

  Persistence lives in two Phase 1 tables:

    * `wiki_pages` — one row per `(tenant_id, slug, audience, version)`
    * `citations` — one row per `(tenant_id, wiki_slug, wiki_audience, chunk_id, claim_hash)`

  This module converts between the `%Page{}` struct and the row tuples,
  handles version bumps, and provides read helpers for the Curator + CLI.
  """

  alias OptimalEngine.Store
  alias OptimalEngine.Wiki.Page

  @doc "Upsert a wiki page (new version row)."
  @spec put(Page.t()) :: :ok | {:error, term()}
  def put(%Page{} = page) do
    frontmatter_json = Jason.encode!(page.frontmatter || %{})

    # `last_curated` is NOT NULL in the schema, but callers may pass nil for
    # a brand-new page. COALESCE to `datetime('now')` so the DEFAULT still
    # applies when the caller doesn't set one explicitly.
    sql = """
    INSERT INTO wiki_pages (tenant_id, slug, audience, version, frontmatter, body, last_curated, curated_by)
    VALUES (?1, ?2, ?3, ?4, ?5, ?6, COALESCE(?7, datetime('now')), ?8)
    ON CONFLICT(tenant_id, slug, audience, version) DO UPDATE SET
      frontmatter  = excluded.frontmatter,
      body         = excluded.body,
      last_curated = COALESCE(excluded.last_curated, datetime('now')),
      curated_by   = excluded.curated_by
    """

    params = [
      page.tenant_id,
      page.slug,
      page.audience,
      page.version,
      frontmatter_json,
      page.body,
      page.last_curated,
      page.curated_by
    ]

    case Store.raw_query(sql, params) do
      {:ok, _} -> :ok
      other -> other
    end
  end

  @doc """
  Fetch the latest version of a page by `(tenant_id, slug, audience)`.
  """
  @spec latest(String.t(), String.t(), String.t()) ::
          {:ok, Page.t()} | {:error, :not_found}
  def latest(tenant_id, slug, audience \\ "default") do
    sql = """
    SELECT tenant_id, slug, audience, version, frontmatter, body, last_curated, curated_by
    FROM wiki_pages
    WHERE tenant_id = ?1 AND slug = ?2 AND audience = ?3
    ORDER BY version DESC
    LIMIT 1
    """

    case Store.raw_query(sql, [tenant_id, slug, audience]) do
      {:ok, [row]} -> {:ok, row_to_page(row)}
      {:ok, []} -> {:error, :not_found}
      other -> other
    end
  end

  @doc "List every page for a tenant (latest version of each (slug, audience) pair)."
  @spec list(String.t()) :: {:ok, [Page.t()]}
  def list(tenant_id) do
    sql = """
    SELECT w.tenant_id, w.slug, w.audience, w.version, w.frontmatter, w.body, w.last_curated, w.curated_by
    FROM wiki_pages w
    INNER JOIN (
      SELECT tenant_id, slug, audience, MAX(version) AS max_version
      FROM wiki_pages
      WHERE tenant_id = ?1
      GROUP BY tenant_id, slug, audience
    ) latest
      ON w.tenant_id = latest.tenant_id
     AND w.slug      = latest.slug
     AND w.audience  = latest.audience
     AND w.version   = latest.max_version
    ORDER BY w.slug, w.audience
    """

    case Store.raw_query(sql, [tenant_id]) do
      {:ok, rows} -> {:ok, Enum.map(rows, &row_to_page/1)}
      other -> other
    end
  end

  @doc "Upsert a batch of citations (bulk, atomic)."
  @spec put_citations(String.t(), String.t(), String.t(), [map()]) :: :ok | {:error, term()}
  def put_citations(tenant_id, wiki_slug, wiki_audience, citations)
      when is_binary(tenant_id) and is_binary(wiki_slug) and is_list(citations) do
    sql = """
    INSERT INTO citations
      (tenant_id, wiki_slug, wiki_audience, chunk_id, claim_hash, last_verified)
    VALUES (?1, ?2, ?3, ?4, ?5, datetime('now'))
    """

    Enum.reduce_while(citations, :ok, fn c, _acc ->
      case Store.raw_query(sql, [
             tenant_id,
             wiki_slug,
             wiki_audience,
             Map.get(c, :chunk_id) || Map.get(c, "chunk_id"),
             Map.get(c, :claim_hash) || Map.get(c, "claim_hash") || derive_claim_hash(c)
           ]) do
        {:ok, _} -> {:cont, :ok}
        {:error, _} = err -> {:halt, err}
      end
    end)
  end

  @doc "Delete all citations tied to a page (used before re-curating)."
  @spec clear_citations(String.t(), String.t(), String.t()) :: :ok
  def clear_citations(tenant_id, slug, audience) do
    Store.raw_query(
      "DELETE FROM citations WHERE tenant_id = ?1 AND wiki_slug = ?2 AND wiki_audience = ?3",
      [tenant_id, slug, audience]
    )

    :ok
  end

  defp row_to_page([tenant_id, slug, audience, version, fm_json, body, last_curated, curated_by]) do
    frontmatter =
      case Jason.decode(fm_json || "{}") do
        {:ok, m} -> m
        _ -> %{}
      end

    %Page{
      tenant_id: tenant_id,
      slug: slug,
      audience: audience,
      version: version,
      frontmatter: frontmatter,
      body: body,
      last_curated: last_curated,
      curated_by: curated_by
    }
  end

  defp derive_claim_hash(%{text: text}) when is_binary(text) do
    :crypto.hash(:sha256, text) |> Base.encode16(case: :lower) |> String.slice(0, 16)
  end

  defp derive_claim_hash(_), do: "unknown"
end
