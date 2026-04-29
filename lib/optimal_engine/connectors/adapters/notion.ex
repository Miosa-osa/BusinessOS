defmodule OptimalEngine.Connectors.Adapters.Notion do
  @moduledoc """
  Notion connector — pages, databases, blocks.

  ## Required config keys
    * `:workspace_name`, `:database_ids` (list; empty = all accessible)

  ## Credentials
    * `:integration_token` — `secret_…`

  ## Cursor shape
  Opaque: page `last_edited_time` ISO timestamp. Initial sync: `nil`.
  """

  use OptimalEngine.Connectors.Adapters.Base,
    kind: :notion,
    display_name: "Notion",
    auth_scheme: :token,
    required_keys: [:workspace_name, :database_ids],
    credential_keys: [:integration_token]

  @impl true
  def sync(_state, _cursor), do: {:error, :not_implemented}

  @impl true
  def transform(raw) when is_map(raw) do
    ext_id = raw["id"] || ""
    title = extract_title(raw) || "Untitled"

    {:ok,
     Transform.new_signal(%{
       id: Transform.signal_id(:notion, ext_id),
       title: title,
       content: flatten_blocks(raw["blocks"] || []),
       path: Transform.source_uri(:notion, ext_id),
       genre: "document",
       mode: :linguistic,
       format: :markdown,
       modified_at: Transform.parse_iso8601(raw["last_edited_time"])
     })}
  end

  defp extract_title(%{"properties" => %{"Name" => %{"title" => [%{"plain_text" => t} | _]}}}),
    do: t

  defp extract_title(%{"properties" => %{"title" => %{"title" => [%{"plain_text" => t} | _]}}}),
    do: t

  defp extract_title(_), do: nil

  defp flatten_blocks(blocks) when is_list(blocks) do
    blocks
    |> Enum.map_join("\n", fn
      %{"type" => type, "paragraph" => %{"rich_text" => rt}} when type == "paragraph" ->
        Enum.map_join(rt, "", fn %{"plain_text" => t} -> t end)

      %{"type" => "heading_1", "heading_1" => %{"rich_text" => rt}} ->
        "# " <> Enum.map_join(rt, "", fn %{"plain_text" => t} -> t end)

      %{"type" => "heading_2", "heading_2" => %{"rich_text" => rt}} ->
        "## " <> Enum.map_join(rt, "", fn %{"plain_text" => t} -> t end)

      _ ->
        ""
    end)
  end
end
