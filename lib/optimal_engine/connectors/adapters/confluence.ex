defmodule OptimalEngine.Connectors.Adapters.Confluence do
  @moduledoc """
  Confluence connector — spaces + pages + blog posts.

  ## Required config keys
    * `:site_url`, `:spaces` (list of space keys; empty = all)

  ## Credentials
    * `:email`, `:api_token`
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :confluence,
    display_name: "Confluence",
    auth_scheme: :basic,
    required_keys: [:site_url, :spaces],
    credential_keys: [:email, :api_token]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:confluence, ext_id),
       title: raw["title"] || "Untitled",
       content: Transform.strip_html(get_in(raw, ["body", "storage", "value"]) || ""),
       path: Transform.source_uri(:confluence, ext_id),
       genre: "document",
       mode: :linguistic,
       modified_at: Transform.parse_iso8601(get_in(raw, ["version", "when"]))
     })}
  end
end
