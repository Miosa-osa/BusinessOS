defmodule OptimalEngine.Connectors.Adapters.GitHub do
  @moduledoc """
  GitHub connector — issues, PRs, discussions, commits, wiki.

  ## Required config keys
    * `:org_or_user`, `:repos` (list of `"name"` or `"*"` for all)

  ## Credentials
    * `:pat` — personal access token **or**
    * `:app_id` + `:installation_id` + `:private_key_pem` (GitHub App)
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :github,
    display_name: "GitHub",
    auth_scheme: :token,
    required_keys: [:org_or_user, :repos],
    credential_keys: []

  @impl true
  def init(config) do
    flat = flatten_credentials(config)

    with :ok <- require_keys(flat, required_config_keys()),
         :ok <- require_github_auth(flat) do
      {:ok, flat}
    end
  end

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["node_id"] || to_string(raw["id"] || "")
    kind_str = if raw["pull_request"], do: "pr", else: "issue"

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:github, ext_id),
       title: "[#{kind_str}] #{raw["title"] || ""}",
       content: raw["body"] || "",
       path: Transform.source_uri(:github, ext_id),
       genre: kind_str,
       modified_at: Transform.parse_iso8601(raw["updated_at"]),
       entities: [get_in(raw, ["user", "login"])] |> Enum.reject(&is_nil/1)
     })}
  end

  defp require_github_auth(config) do
    pat = Map.get(config, "pat") || Map.get(config, :pat)

    app? =
      Enum.all?(
        [:app_id, :installation_id, :private_key_pem],
        fn k -> Map.has_key?(config, k) or Map.has_key?(config, Atom.to_string(k)) end
      )

    cond do
      is_binary(pat) and pat != "" -> :ok
      app? -> :ok
      true -> {:error, :missing_credentials}
    end
  end
end
