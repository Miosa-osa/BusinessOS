defmodule OptimalEngine.CortexFeed do
  @moduledoc """
  Provides engine data to MiosaMemory.Cortex in the format it expects.

  Instead of Cortex reading from MEMORY.md via MiosaMemory.Store.recall(),
  this module queries SQLite directly for recent signals, decisions, and
  active operations — the real data.

  ## Usage

  Called by Cortex's provider config or manually:

      OptimalEngine.CortexFeed.recall()
      # => "## Recent Signals (7 days)\\n..."

  Also provides session data for Cortex synthesis:

      OptimalEngine.CortexFeed.session_context()
      # => "## Active Sessions\\n..."
  """

  require Logger

  alias OptimalEngine.Store

  @recent_days 7

  @doc """
  Returns a formatted string of recent engine data for Cortex consumption.

  Combines:
  - Recent signals (last 7 days)
  - Recent decisions
  - Active operations (nodes with recent activity)
  """
  @spec recall() :: String.t()
  def recall do
    sections =
      [
        recent_signals_section(),
        recent_decisions_section(),
        active_operations_section()
      ]
      |> Enum.reject(&(&1 == ""))
      |> Enum.join("\n\n")

    if sections == "", do: "No recent engine data.", else: sections
  rescue
    e ->
      Logger.debug("[CortexFeed] Error building recall: #{inspect(e)}")
      ""
  end

  @doc """
  Returns session context for Cortex synthesis.
  """
  @spec session_context() :: String.t()
  def session_context do
    case Store.raw_query(
           "SELECT id, started_at, summary, message_count FROM sessions ORDER BY started_at DESC LIMIT 5",
           []
         ) do
      {:ok, rows} when rows != [] ->
        items =
          Enum.map(rows, fn [id, started, summary, count] ->
            "- **#{id}** (#{started}, #{count} msgs): #{summary || "no summary"}"
          end)

        "## Active Sessions\n\n#{Enum.join(items, "\n")}"

      _ ->
        ""
    end
  rescue
    _ -> ""
  end

  # ---------------------------------------------------------------------------
  # Private
  # ---------------------------------------------------------------------------

  defp recent_signals_section do
    cutoff = recent_cutoff()

    case Store.raw_query(
           "SELECT title, node, genre, sn_ratio, created_at FROM contexts WHERE type = 'signal' AND created_at >= ?1 ORDER BY created_at DESC LIMIT 20",
           [cutoff]
         ) do
      {:ok, rows} when rows != [] ->
        items =
          Enum.map(rows, fn [title, node, genre, sn, created] ->
            sn_str = if is_number(sn), do: Float.round(sn * 1.0, 1), else: "?"
            "- [#{genre || "?"}] **#{title || "Untitled"}** → #{node} (S/N: #{sn_str}, #{created})"
          end)

        "## Recent Signals (#{@recent_days} days)\n\n#{Enum.join(items, "\n")}"

      _ ->
        ""
    end
  end

  defp recent_decisions_section do
    cutoff = recent_cutoff()

    case Store.raw_query(
           "SELECT title, decision, decided_by, decided_at FROM decisions WHERE decided_at >= ?1 ORDER BY decided_at DESC LIMIT 10",
           [cutoff]
         ) do
      {:ok, rows} when rows != [] ->
        items =
          Enum.map(rows, fn [title, decision, by, at] ->
            "- **#{title}**: #{decision} (by #{by || "?"}, #{at})"
          end)

        "## Recent Decisions\n\n#{Enum.join(items, "\n")}"

      _ ->
        ""
    end
  end

  defp active_operations_section do
    cutoff = recent_cutoff()

    case Store.raw_query(
           "SELECT node, COUNT(*) as cnt FROM contexts WHERE type = 'signal' AND created_at >= ?1 GROUP BY node ORDER BY cnt DESC LIMIT 10",
           [cutoff]
         ) do
      {:ok, rows} when rows != [] ->
        items =
          Enum.map(rows, fn [node, count] ->
            "- **#{node}**: #{count} signals"
          end)

        "## Active Operations\n\n#{Enum.join(items, "\n")}"

      _ ->
        ""
    end
  end

  defp recent_cutoff do
    DateTime.utc_now()
    |> DateTime.add(-@recent_days * 86400, :second)
    |> DateTime.to_iso8601()
  end
end
