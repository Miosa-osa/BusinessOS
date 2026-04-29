defmodule OptimalEngine.Connectors.Adapters.Drive do
  @moduledoc """
  Google Drive connector — docs, sheets, slides, PDFs, arbitrary files.

  ## Required config keys
    * `:scope` — `:my_drive | :shared_drive | :starred`
    * `:drive_id` — required when `scope: :shared_drive`

  ## Credentials
    * `:oauth_refresh_token`, `:client_id`, `:client_secret`

  ## Cursor shape
  Google Drive `changes.startPageToken`. Initial sync: `nil`.
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :drive,
    display_name: "Google Drive",
    auth_scheme: :oauth2,
    required_keys: [:scope],
    credential_keys: [:oauth_refresh_token, :client_id, :client_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""
    name = raw["name"] || "Untitled"
    mime = raw["mimeType"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:drive, ext_id),
       title: name,
       content: raw["exportedText"] || raw["content"] || "",
       path: Transform.source_uri(:drive, ext_id),
       genre: drive_genre(mime),
       mode: drive_mode(mime),
       format: drive_format(mime),
       modified_at: Transform.parse_iso8601(raw["modifiedTime"])
     })}
  end

  defp drive_genre("application/vnd.google-apps.document"), do: "document"
  defp drive_genre("application/vnd.google-apps.spreadsheet"), do: "table"
  defp drive_genre("application/vnd.google-apps.presentation"), do: "slides"
  defp drive_genre("application/pdf"), do: "document"
  defp drive_genre(_), do: "file"

  defp drive_mode("application/vnd.google-apps.spreadsheet"), do: :data
  defp drive_mode("image/" <> _), do: :visual
  defp drive_mode(_), do: :linguistic

  defp drive_format("application/vnd.google-apps.document"), do: :markdown
  defp drive_format("application/vnd.google-apps.spreadsheet"), do: :json
  defp drive_format(_), do: :unknown
end
