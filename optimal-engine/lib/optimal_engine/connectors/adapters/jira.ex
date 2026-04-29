defmodule OptimalEngine.Connectors.Adapters.Jira do
  @moduledoc """
  Jira connector — issues, comments, sprints, releases.

  ## Required config keys
    * `:site_url` (e.g. `https://acme.atlassian.net`), `:projects` (list of keys)

  ## Credentials
    * `:email`, `:api_token`
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :jira,
    display_name: "Jira",
    auth_scheme: :basic,
    required_keys: [:site_url, :projects],
    credential_keys: [:email, :api_token]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["key"] || raw["id"] || ""
    fields = raw["fields"] || %{}

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:jira, ext_id),
       title: "[#{ext_id}] #{fields["summary"] || ""}",
       content: Transform.strip_html(fields["description"] || ""),
       path: Transform.source_uri(:jira, ext_id),
       genre: "ticket",
       modified_at: Transform.parse_iso8601(fields["updated"]),
       entities:
         Enum.filter([fields["assignee"]["displayName"], fields["reporter"]["displayName"]], & &1)
     })}
  rescue
    _ ->
      {:error, :bad_payload}
  end
end
