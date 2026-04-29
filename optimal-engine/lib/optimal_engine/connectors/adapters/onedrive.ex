defmodule OptimalEngine.Connectors.Adapters.OneDrive do
  @moduledoc """
  OneDrive / SharePoint connector — files, documents, sites.

  ## Required config keys
    * `:tenant_id_ms`, `:drive_id`

  ## Credentials
    * `:client_id`, `:client_secret` (Azure AD app)
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :onedrive,
    display_name: "OneDrive",
    auth_scheme: :oauth2,
    required_keys: [:tenant_id_ms, :drive_id],
    credential_keys: [:client_id, :client_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:onedrive, ext_id),
       title: raw["name"] || "Untitled",
       content: raw["content"] || "",
       path: Transform.source_uri(:onedrive, ext_id),
       genre: "file",
       modified_at: Transform.parse_iso8601(raw["lastModifiedDateTime"])
     })}
  end
end
