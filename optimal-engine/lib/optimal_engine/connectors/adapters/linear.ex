defmodule OptimalEngine.Connectors.Adapters.Linear do
  @moduledoc """
  Linear connector — issues, comments, cycles, projects.

  ## Required config keys
    * `:team_ids` (list of Linear team UUIDs)

  ## Credentials
    * `:api_key` — `lin_api_…`
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :linear,
    display_name: "Linear",
    auth_scheme: :token,
    required_keys: [:team_ids],
    credential_keys: [:api_key]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:linear, ext_id),
       title: raw["title"] || "Untitled",
       content: raw["description"] || "",
       path: Transform.source_uri(:linear, ext_id),
       genre: "ticket",
       modified_at: Transform.parse_iso8601(raw["updatedAt"]),
       entities:
         Enum.filter([get_in(raw, ["assignee", "name"]), get_in(raw, ["creator", "name"])], & &1)
     })}
  end
end
