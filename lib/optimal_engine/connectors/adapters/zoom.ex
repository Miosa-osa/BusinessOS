defmodule OptimalEngine.Connectors.Adapters.Zoom do
  @moduledoc """
  Zoom connector — meeting recordings + transcripts.

  ## Required config keys
    * `:account_id`

  ## Credentials
    * `:client_id`, `:client_secret` (server-to-server OAuth)
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :zoom,
    display_name: "Zoom",
    auth_scheme: :oauth2,
    required_keys: [:account_id],
    credential_keys: [:client_id, :client_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["uuid"] || raw["id"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:zoom, to_string(ext_id)),
       title: raw["topic"] || "Zoom meeting",
       content: raw["transcript"] || "",
       path: Transform.source_uri(:zoom, to_string(ext_id)),
       genre: "transcript",
       mode: :linguistic,
       modified_at: Transform.parse_iso8601(raw["start_time"])
     })}
  end
end
