defmodule Mix.Tasks.Optimal.Remember do
  @shortdoc "Store observations and mine friction patterns"
  @moduledoc """
  Three-mode friction capture for the knowledge base.

  Usage:
      mix optimal.remember "always check duplicates before inserting"
      mix optimal.remember --contextual
      mix optimal.remember --mine
      mix optimal.remember --list
      mix optimal.remember --list --category process
      mix optimal.remember --escalations
  """

  use Mix.Task

  @impl Mix.Task
  def run(args) do
    Mix.Task.run("app.start")

    cond do
      "--contextual" in args ->
        run_contextual()

      "--mine" in args ->
        run_mine()

      "--list" in args ->
        run_list(args)

      "--escalations" in args ->
        run_escalations()

      args != [] ->
        observation =
          args
          |> Enum.reject(&String.starts_with?(&1, "--"))
          |> Enum.join(" ")

        run_explicit(observation)

      true ->
        IO.puts("Usage: mix optimal.remember \"observation\"")
        IO.puts("       mix optimal.remember --contextual")
        IO.puts("       mix optimal.remember --mine")
        IO.puts("       mix optimal.remember --list [--category CAT]")
        IO.puts("       mix optimal.remember --escalations")
    end
  end

  # ---------------------------------------------------------------------------
  # Mode runners
  # ---------------------------------------------------------------------------

  defp run_explicit(observation) do
    IO.puts("\nRememberLoop — Storing Observation\n")

    case OptimalEngine.Insight.Remember.remember(observation) do
      {:ok, result} ->
        IO.puts("  Stored: [#{result.category}] #{result.content}")
        IO.puts("  Confidence: #{result.confidence}")

        if result.escalation.escalated do
          IO.puts(
            "  Escalation: #{result.escalation.count} observations in this category (total confidence: #{result.escalation.total_confidence})"
          )

          if result.escalation.ready_for_rethink do
            IO.puts("  -> Ready for rethink! Run: mix optimal.rethink \"#{result.category}\"")
          end
        end

      {:error, reason} ->
        IO.puts("  Failed: #{inspect(reason)}")
    end
  end

  defp run_contextual do
    IO.puts("\nRememberLoop — Contextual Scan\n")

    case OptimalEngine.Insight.Remember.contextual_scan() do
      {:ok, []} ->
        IO.puts("  No friction signals found in recent contexts.")

      {:ok, observations} ->
        IO.puts("  Found #{length(observations)} friction signals:\n")

        Enum.each(observations, fn obs ->
          IO.puts("  [#{obs.category}] #{String.slice(obs.content, 0, 80)}")
        end)
    end
  end

  defp run_mine do
    IO.puts("\nRememberLoop — Session Mining\n")

    case OptimalEngine.Insight.Remember.mine_sessions() do
      {:ok, []} ->
        IO.puts("  No patterns extracted from sessions.")

      {:ok, observations} ->
        IO.puts("  Extracted #{length(observations)} patterns:\n")

        Enum.each(observations, fn obs ->
          IO.puts("  [#{obs.category}] #{String.slice(obs.content, 0, 80)}")
        end)
    end
  end

  defp run_list(args) do
    IO.puts("\nRememberLoop — Observations\n")

    category =
      case Enum.find_index(args, &(&1 == "--category")) do
        nil -> nil
        idx -> Enum.at(args, idx + 1)
      end

    opts = if category, do: [category: category], else: []

    case OptimalEngine.Insight.Remember.list(opts) do
      {:ok, []} ->
        IO.puts("  No observations stored yet.")

      {:ok, observations} ->
        IO.puts("  #{length(observations)} observations:\n")

        Enum.each(observations, fn obs ->
          IO.puts("  ##{obs.id} [#{obs.category}] #{String.slice(obs.content, 0, 60)}")
          IO.puts("    confidence: #{obs.confidence} | source: #{obs.source} | #{obs.created_at}")
        end)
    end
  end

  defp run_escalations do
    IO.puts("\nRememberLoop — Escalation Candidates\n")

    case OptimalEngine.Insight.Remember.escalation_candidates() do
      {:ok, []} ->
        IO.puts("  No categories have reached escalation threshold (3+ observations).")

      {:ok, candidates} ->
        Enum.each(candidates, fn c ->
          status = if c.ready_for_rethink, do: "READY FOR RETHINK", else: "accumulating"

          IO.puts(
            "  [#{c.category}] #{c.count} observations, confidence: #{Float.round(c.total_confidence * 1.0, 2)} — #{status}"
          )
        end)
    end
  end
end
