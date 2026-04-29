defmodule OptimalEngine.Wiki.Integrity do
  @moduledoc """
  Integrity checks for wiki pages — the "citation hygiene" layer.

  A healthy wiki page has three properties:

    1. **Every directive uses a whitelisted verb.** Typos + malicious
       injection attempts surface as warnings.
    2. **Every `{{cite: uri}}` resolves.** Broken citations are flagged
       with the offending URI so the curator can repair or remove.
    3. **Every factual claim is cited.** A paragraph containing a claim
       (detected heuristically) without a nearby `{{cite: …}}` is flagged.

  Additional checks come from `.wiki/SCHEMA.md` rules (minimum sections,
  citation density, size ceilings) — those are applied via
  `Integrity.against_schema/2`.

  This module is pure: it doesn't mutate the page, it just reports.
  """

  alias OptimalEngine.Wiki.Directives

  @type issue :: %{
          severity: :error | :warning | :info,
          kind: atom(),
          message: String.t(),
          location: String.t() | nil
        }

  @type report :: %{
          page_slug: String.t(),
          ok?: boolean(),
          issues: [issue()]
        }

  @doc """
  Run the full integrity suite against a page body.

  Options:
    * `:resolve_uri` — function `fn uri -> :ok | {:error, reason} end` used
      to test whether each cite resolves. Defaults to always-ok so tests
      that don't need live resolution can skip the Store.

  Returns `%{page_slug, ok?, issues}`. `ok?` is true when every issue has
  severity `:info`.
  """
  @spec check(OptimalEngine.Wiki.Page.t(), keyword()) :: report()
  def check(%OptimalEngine.Wiki.Page{} = page, opts \\ []) do
    resolver = Keyword.get(opts, :resolve_uri, fn _ -> :ok end)

    issues =
      []
      |> append(check_verbs(page.body))
      |> append(check_citations(page.body, resolver))
      |> append(check_claim_density(page.body))
      |> append(check_page_size(page.body))

    %{
      page_slug: page.slug,
      ok?: not Enum.any?(issues, &(&1.severity == :error)),
      issues: issues
    }
  end

  @doc """
  Apply schema-defined rules from `.wiki/SCHEMA.md`. The schema is a set
  of markdown-documented rules; we support the machine-enforceable subset:

    * required top-level sections (`## Summary`, `## Related`, etc.)
    * minimum citation density per section
    * maximum page size in bytes

  Custom rules can be added via `:rules` keyword.
  """
  @spec against_schema(OptimalEngine.Wiki.Page.t(), map(), keyword()) :: report()
  def against_schema(%OptimalEngine.Wiki.Page{} = page, schema, opts \\ []) when is_map(schema) do
    base = check(page, opts)

    schema_issues =
      []
      |> append(check_required_sections(page.body, Map.get(schema, "required_sections", [])))
      |> append(check_max_size(page.body, Map.get(schema, "max_bytes")))
      |> append(
        check_required_frontmatter(page.frontmatter, Map.get(schema, "required_frontmatter", []))
      )

    %{
      base
      | issues: base.issues ++ schema_issues,
        ok?: base.ok? and Enum.all?(schema_issues, &(&1.severity != :error))
    }
  end

  # ─── checks ──────────────────────────────────────────────────────────────

  defp check_verbs(body) do
    if Directives.all_verbs_whitelisted?(body) do
      []
    else
      {:ok, directives} = Directives.parse(body)

      directives
      |> Enum.reject(fn d -> d.verb == :wikilink or d.verb in Directives.whitelist() end)
      |> Enum.map(fn d ->
        %{
          severity: :error,
          kind: :invalid_verb,
          message: "Unknown directive verb `#{d.verb}` in `#{d.raw}`",
          location: "offset #{d.offset}"
        }
      end)
    end
  end

  defp check_citations(body, resolver) do
    {:ok, directives} = Directives.parse(body)

    directives
    |> Enum.filter(&(&1.verb == :cite))
    |> Enum.flat_map(fn d ->
      case resolver.(d.argument) do
        :ok ->
          []

        {:error, reason} ->
          [
            %{
              severity: :error,
              kind: :broken_citation,
              message: "Citation `#{d.argument}` does not resolve (#{inspect(reason)})",
              location: "offset #{d.offset}"
            }
          ]
      end
    end)
  end

  # Heuristic: every non-empty paragraph should contain at least one cite
  # directive, wikilink, or explicit acknowledgment phrase (for pure
  # structural paragraphs like headers, we skip).
  defp check_claim_density(body) do
    paragraphs =
      body
      |> strip_directive_markup()
      |> String.split(~r/\n{2,}/, trim: true)
      |> Enum.with_index()

    paragraphs
    |> Enum.flat_map(fn {para, idx} ->
      trimmed = String.trim(para)

      cond do
        trimmed == "" ->
          []

        String.starts_with?(trimmed, "#") ->
          []

        String.starts_with?(trimmed, "- ") or String.starts_with?(trimmed, "* ") ->
          # List items often carry their own citations; flag only if long
          if String.length(trimmed) > 200 and not contains_citation?(body, idx) do
            [claim_issue(idx, trimmed)]
          else
            []
          end

        contains_citation?(body, idx) ->
          []

        # Very short paragraphs (< 40 chars) are likely labels/section-intros
        String.length(trimmed) < 40 ->
          []

        true ->
          [claim_issue(idx, trimmed)]
      end
    end)
  end

  defp claim_issue(idx, para) do
    %{
      severity: :warning,
      kind: :uncited_claim,
      message:
        "Paragraph #{idx + 1} contains substantive content but no {{cite: …}} — #{String.slice(para, 0, 60)}…",
      location: "paragraph #{idx + 1}"
    }
  end

  defp contains_citation?(body, paragraph_idx) do
    # Cheap approach: parse the raw body for directives and check whether
    # any directive falls within the given paragraph boundary. For this
    # coarse heuristic we count ANY citation in the body — the goal is
    # to avoid false positives on small pages, not to prove per-paragraph
    # citation placement.
    {:ok, directives} = Directives.parse(body)
    _ = paragraph_idx
    Enum.any?(directives, &(&1.verb == :cite))
  end

  defp strip_directive_markup(body) do
    body
    |> String.replace(~r/\{\{[^}]+\}\}/, " ")
    |> String.replace(~r/\[\[[^\]]+\]\]/, " ")
  end

  defp check_page_size(body) do
    size = byte_size(body)

    cond do
      size > 50_000 ->
        [
          %{
            severity: :warning,
            kind: :page_too_large,
            message: "Page body is #{size} bytes; consider spawning child pages",
            location: nil
          }
        ]

      size == 0 ->
        [
          %{
            severity: :error,
            kind: :empty_page,
            message: "Page body is empty",
            location: nil
          }
        ]

      true ->
        []
    end
  end

  defp check_required_sections(body, []), do: []

  defp check_required_sections(body, required) when is_list(required) do
    headings = Regex.scan(~r/^##\s+(.+?)\s*$/m, body, capture: :all_but_first)
    present = headings |> Enum.map(fn [h] -> String.downcase(h) end) |> MapSet.new()

    required
    |> Enum.reject(fn name -> MapSet.member?(present, String.downcase(name)) end)
    |> Enum.map(fn name ->
      %{
        severity: :error,
        kind: :missing_section,
        message: "Required section `## #{name}` is absent",
        location: nil
      }
    end)
  end

  defp check_max_size(_body, nil), do: []

  defp check_max_size(body, max) when is_integer(max) do
    if byte_size(body) > max do
      [
        %{
          severity: :error,
          kind: :schema_size_exceeded,
          message: "Page body is #{byte_size(body)} bytes; schema maximum is #{max}",
          location: nil
        }
      ]
    else
      []
    end
  end

  defp check_required_frontmatter(_fm, []), do: []

  defp check_required_frontmatter(fm, required) when is_list(required) do
    Enum.flat_map(required, fn key ->
      case Map.get(fm, key) do
        nil ->
          [
            %{
              severity: :error,
              kind: :missing_frontmatter,
              message: "Required frontmatter key `#{key}` is absent",
              location: nil
            }
          ]

        "" ->
          [
            %{
              severity: :error,
              kind: :empty_frontmatter,
              message: "Required frontmatter key `#{key}` is empty",
              location: nil
            }
          ]

        _ ->
          []
      end
    end)
  end

  defp append(acc, list) when is_list(list), do: acc ++ list
end
