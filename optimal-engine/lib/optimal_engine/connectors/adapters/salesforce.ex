defmodule OptimalEngine.Connectors.Adapters.Salesforce do
  @moduledoc """
  Salesforce connector — accounts, opportunities, contacts, cases, notes.

  ## Required config keys
    * `:instance_url`, `:objects` (list of SObject API names)

  ## Credentials
    * `:client_id`, `:client_secret`, `:refresh_token`
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :salesforce,
    display_name: "Salesforce",
    auth_scheme: :oauth2,
    required_keys: [:instance_url, :objects],
    credential_keys: [:client_id, :client_secret, :refresh_token]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["Id"] || ""
    sobject = Map.get(raw, "attributes", %{}) |> Map.get("type", "record")

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:salesforce, ext_id),
       title: raw["Name"] || raw["Subject"] || ext_id,
       content: raw["Description"] || raw["Body"] || "",
       path: Transform.source_uri(:salesforce, ext_id),
       genre: String.downcase(sobject),
       modified_at: Transform.parse_iso8601(raw["LastModifiedDate"])
     })}
  end
end
