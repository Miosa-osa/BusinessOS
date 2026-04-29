defmodule OptimalEngine.Connectors do
  @moduledoc """
  Top-level facade for the connectors layer.

      Connectors.list()                   — every registered adapter
      Connectors.register(attrs)          — persist a new connector row
      Connectors.run(connector_id, opts)  — execute one sync cycle

  See `OptimalEngine.Connectors.Behaviour` for the adapter contract.

  The 14 adapters live under `OptimalEngine.Connectors.Adapters.*`:
  Slack, Gmail, Drive, Notion, Jira, Linear, GitHub, Zoom, Confluence,
  Teams, Dropbox, OneDrive, Salesforce, HubSpot.
  """

  alias OptimalEngine.Connectors.{Credential, Registry, Runner}

  @doc "Return `[{kind, display_name, auth_scheme}]` for every registered adapter."
  defdelegate list, to: Registry, as: :summary

  @doc "Look up an adapter module by `kind`."
  defdelegate fetch_adapter(kind), to: Registry, as: :fetch

  @doc "Every `kind` atom known to the registry."
  defdelegate kinds, to: Registry

  @doc "Persist a new connector row. See `Runner.upsert_row/1`."
  defdelegate register(attrs), to: Runner, as: :upsert_row

  @doc "Run one sync cycle for the given connector."
  def run(connector_id, opts \\ []), do: Runner.run(connector_id, opts)

  @doc "`true` when the master key for credential encryption is configured."
  defdelegate credentials_ready?, to: Credential, as: :ready?
end
