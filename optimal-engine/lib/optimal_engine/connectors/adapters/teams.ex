defmodule OptimalEngine.Connectors.Adapters.Teams do
  @moduledoc """
  Microsoft Teams connector — channel messages, chats, meetings.

  ## Required config keys
    * `:tenant_id_ms` (Azure AD tenant), `:team_ids` (list)

  ## Credentials
    * `:client_id`, `:client_secret` (Azure AD app)
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :teams,
    display_name: "Microsoft Teams",
    auth_scheme: :oauth2,
    required_keys: [:tenant_id_ms, :team_ids],
    credential_keys: [:client_id, :client_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""
    body_text = get_in(raw, ["body", "content"]) || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:teams, ext_id),
       title: String.slice(Transform.strip_html(body_text), 0, 80),
       content: Transform.strip_html(body_text),
       path: Transform.source_uri(:teams, ext_id),
       genre: "message",
       modified_at: Transform.parse_iso8601(raw["lastModifiedDateTime"]),
       entities: [get_in(raw, ["from", "user", "displayName"])] |> Enum.reject(&is_nil/1)
     })}
  end
end
