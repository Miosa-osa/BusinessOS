defmodule OptimalEngine.Connectors.Adapters.HubSpot do
  @moduledoc """
  HubSpot connector — CRM contacts, companies, deals, engagements, notes.

  ## Required config keys
    * `:portal_id`, `:objects` (list of HubSpot object API names)

  ## Credentials
    * `:access_token` (private app) **or**
    * `:client_id` + `:client_secret` + `:refresh_token` (OAuth app)
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :hubspot,
    display_name: "HubSpot",
    auth_scheme: :oauth2,
    required_keys: [:portal_id, :objects],
    credential_keys: []

  @impl true
  def init(config) do
    flat = flatten_credentials(config)

    with :ok <- require_keys(flat, required_config_keys()),
         :ok <- require_hubspot_auth(flat) do
      {:ok, flat}
    end
  end

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = to_string(raw["id"] || "")
    props = raw["properties"] || %{}

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:hubspot, ext_id),
       title:
         props["dealname"] || props["firstname"] || props["company"] || props["subject"] ||
           ext_id,
       content: props["description"] || props["notes_last_contacted"] || "",
       path: Transform.source_uri(:hubspot, ext_id),
       genre: "crm",
       modified_at: Transform.parse_iso8601(props["hs_lastmodifieddate"] || raw["updatedAt"])
     })}
  end

  defp require_hubspot_auth(config) do
    token = Map.get(config, "access_token") || Map.get(config, :access_token)

    oauth? =
      Enum.all?([:client_id, :client_secret, :refresh_token], fn k ->
        Map.has_key?(config, k) or Map.has_key?(config, Atom.to_string(k))
      end)

    cond do
      is_binary(token) and token != "" -> :ok
      oauth? -> :ok
      true -> {:error, :missing_credentials}
    end
  end
end
