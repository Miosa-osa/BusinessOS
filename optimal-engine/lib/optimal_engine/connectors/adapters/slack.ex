defmodule OptimalEngine.Connectors.Adapters.Slack do
  @moduledoc """
  Slack connector — channels, threads, DMs, files.

  ## Required config keys
    * `:workspace_id`, `:channels` (list; empty = all)

  ## Credentials
    * `:bot_token` — starts with `xoxb-…`

  ## Cursor shape
  Opaque: `"<channel_id>:<oldest_ts>"`. Initial sync: `nil`.
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :slack,
    display_name: "Slack",
    auth_scheme: :oauth2,
    required_keys: [:workspace_id, :channels],
    credential_keys: [:bot_token]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["client_msg_id"] || raw["ts"] || ""
    text = raw["text"] || ""
    user = raw["user"] || "slack-user"

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:slack, ext_id),
       title: String.slice(text, 0, 80),
       content: text,
       path: Transform.source_uri(:slack, ext_id),
       genre: "message",
       entities: [user],
       modified_at: parse_ts(raw["ts"])
     })}
  end

  defp parse_ts(nil), do: DateTime.utc_now()

  defp parse_ts(ts) when is_binary(ts) do
    case Float.parse(ts) do
      {f, _} -> f |> round() |> DateTime.from_unix!()
      _ -> DateTime.utc_now()
    end
  end
end
