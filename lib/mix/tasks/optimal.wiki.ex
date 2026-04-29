defmodule Mix.Tasks.Optimal.Wiki do
  @shortdoc "Operate on Tier-3 wiki pages (list / view / verify / curate)"

  @moduledoc """
  Wiki Tier-3 operations.

  ## Usage

      mix optimal.wiki list                         — list every page in the default tenant
      mix optimal.wiki view <slug>                  — view a page's rendered body
      mix optimal.wiki view <slug> --audience sales
      mix optimal.wiki view <slug> --format claude  — render with directive resolution
      mix optimal.wiki verify <slug>                — run integrity + schema checks
      mix optimal.wiki verify-all                   — run checks across every page

  ## Options

    --tenant <id>      — default: `default`
    --audience <name>  — default: `default`
    --format <fmt>     — plain | markdown | claude | openai (for `view`)
  """

  use Mix.Task

  alias OptimalEngine.Tenancy.Tenant
  alias OptimalEngine.Wiki

  @impl Mix.Task
  def run(args) do
    Mix.Task.run("app.start")

    {opts, positional, _} =
      OptionParser.parse(args,
        strict: [tenant: :string, audience: :string, format: :string]
      )

    tenant_id = Keyword.get(opts, :tenant, Tenant.default_id())
    audience = Keyword.get(opts, :audience, "default")
    format = Keyword.get(opts, :format, "markdown") |> String.to_atom()

    case positional do
      ["list" | _] -> list(tenant_id)
      ["view", slug | _] -> view(tenant_id, slug, audience, format)
      ["verify", slug | _] -> verify(tenant_id, slug, audience)
      ["verify-all" | _] -> verify_all(tenant_id)
      _ -> Mix.shell().info(@moduledoc)
    end
  end

  defp list(tenant_id) do
    case Wiki.list(tenant_id) do
      {:ok, []} ->
        IO.puts("\n  No wiki pages for tenant #{tenant_id}.\n")

      {:ok, pages} ->
        IO.puts("")
        IO.puts("  Wiki pages — tenant: #{tenant_id}  (#{length(pages)} total)")
        IO.puts("  " <> String.duplicate("─", 70))

        Enum.each(pages, fn p ->
          IO.puts(
            "    #{String.pad_trailing(p.slug, 34)} audience=#{String.pad_trailing(p.audience, 14)} v#{p.version}"
          )
        end)

        IO.puts("")

      _ ->
        IO.puts("  (wiki unavailable)")
    end
  end

  defp view(tenant_id, slug, audience, format) do
    case Wiki.latest(tenant_id, slug, audience) do
      {:ok, page} ->
        {rendered, warnings} = Wiki.render(page, &noop_resolver/2, format: format)
        IO.puts(rendered)

        if warnings != [] do
          IO.puts("\n---\nWarnings:")
          Enum.each(warnings, &IO.puts("  • #{&1}"))
        end

      {:error, :not_found} ->
        Mix.shell().error("Page not found: #{slug} (audience=#{audience}, tenant=#{tenant_id})")
        System.halt(1)
    end
  end

  defp verify(tenant_id, slug, audience) do
    with {:ok, page} <- Wiki.latest(tenant_id, slug, audience) do
      report = Wiki.verify_against_schema(page, default_schema())
      print_report(report)
    else
      {:error, :not_found} ->
        Mix.shell().error("Page not found: #{slug}")
        System.halt(1)
    end
  end

  defp verify_all(tenant_id) do
    case Wiki.list(tenant_id) do
      {:ok, pages} ->
        total = length(pages)

        results =
          Enum.map(pages, fn p -> Wiki.verify_against_schema(p, default_schema()) end)

        ok = Enum.count(results, & &1.ok?)
        total_issues = Enum.flat_map(results, & &1.issues) |> length()

        IO.puts("")
        IO.puts("  Wiki verification — #{tenant_id}")
        IO.puts("  " <> String.duplicate("─", 50))
        IO.puts("    Pages:         #{total}")
        IO.puts("    OK:            #{ok}")
        IO.puts("    Total issues:  #{total_issues}")
        IO.puts("")

        results
        |> Enum.reject(& &1.ok?)
        |> Enum.take(10)
        |> Enum.each(fn r ->
          IO.puts("    ✗ #{r.page_slug}")

          Enum.take(r.issues, 3)
          |> Enum.each(fn i ->
            IO.puts("        [#{i.severity}] #{i.message}")
          end)
        end)

        IO.puts("")

      _ ->
        Mix.shell().error("Could not list pages")
    end
  end

  defp print_report(report) do
    IO.puts("")
    status = if report.ok?, do: "✓ OK", else: "✗ FAILED"
    IO.puts("  #{status}  —  #{report.page_slug}")
    IO.puts("  " <> String.duplicate("─", 50))

    if report.issues == [] do
      IO.puts("    No issues.")
    else
      Enum.each(report.issues, fn i ->
        IO.puts("    [#{i.severity}] #{i.kind}  —  #{i.message}")
      end)
    end

    IO.puts("")
  end

  defp default_schema do
    %{
      "required_sections" => ["Summary", "Open threads", "Related"],
      "required_frontmatter" => ["slug", "audience", "version", "last_curated"],
      "max_bytes" => 50_000
    }
  end

  defp noop_resolver(_directive, _opts), do: {:ok, "", %{}}
end
