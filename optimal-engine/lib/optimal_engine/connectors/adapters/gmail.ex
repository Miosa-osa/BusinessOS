defmodule OptimalEngine.Connectors.Adapters.Gmail do
  @moduledoc """
  Gmail connector — messages, threads, labels, attachments.

  ## Required config keys
    * `:user_email`, `:labels` (list of label names to sync; default all)

  ## Credentials
    * `:oauth_refresh_token` — long-lived OAuth2 refresh token
    * `:client_id`, `:client_secret` — Google app credentials

  ## Cursor shape
  Opaque: a Gmail `historyId`. Initial sync: `nil` → full fetch bounded
  by `:initial_lookback_days` (default 30).
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :gmail,
    display_name: "Gmail",
    auth_scheme: :oauth2,
    required_keys: [:user_email],
    credential_keys: [:oauth_refresh_token, :client_id, :client_secret]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""
    subject = get_header(raw, "Subject") || "(no subject)"
    from = get_header(raw, "From") || ""

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:gmail, ext_id),
       title: String.slice(subject, 0, 120),
       content: Transform.strip_html(raw["snippet"] || raw["body"] || ""),
       path: Transform.source_uri(:gmail, ext_id),
       genre: "email",
       mode: :linguistic,
       entities: extract_addresses(from),
       modified_at: Transform.parse_iso8601(raw["internalDate"])
     })}
  end

  defp get_header(%{"payload" => %{"headers" => headers}}, name) when is_list(headers) do
    Enum.find_value(headers, fn
      %{"name" => ^name, "value" => v} -> v
      _ -> nil
    end)
  end

  defp get_header(_, _), do: nil

  defp extract_addresses(header) when is_binary(header) do
    Regex.scan(~r/[\w.+-]+@[\w.-]+\.[A-Za-z]{2,}/, header)
    |> Enum.flat_map(& &1)
  end

  defp extract_addresses(_), do: []
end
