defmodule OptimalEngine.Connectors.Adapters.Dropbox do
  @moduledoc """
  Dropbox connector — files + folders + shared content.

  ## Required config keys
    * `:namespace` — `:personal | :team`

  ## Credentials
    * `:refresh_token`, `:app_key`, `:app_secret`
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :dropbox,
    display_name: "Dropbox",
    auth_scheme: :oauth2,
    required_keys: [:namespace],
    credential_keys: [:refresh_token, :app_key, :app_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || raw["path_lower"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:dropbox, ext_id),
       title: raw["name"] || Path.basename(raw["path_display"] || ""),
       content: raw["content"] || "",
       path: Transform.source_uri(:dropbox, ext_id),
       genre: "file",
       modified_at: Transform.parse_iso8601(raw["server_modified"])
     })}
  end
end
