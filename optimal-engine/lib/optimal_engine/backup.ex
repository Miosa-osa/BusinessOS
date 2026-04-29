defmodule OptimalEngine.Backup do
  @moduledoc """
  Backup + restore for the engine's SQLite database.

  SQLite's online backup uses `VACUUM INTO 'target.db'` — atomic,
  safe to run while the engine keeps serving, and produces a
  self-contained single file with no WAL trail attached. That one
  file plus any migration metadata is the complete engine state.

  ## Usage

      # Live backup while the engine runs
      OptimalEngine.Backup.create("/var/backups/engine-2026-01-05.db")

      # Restore from a backup (requires supervisor shutdown first —
      # we fail loudly if it's still up)
      OptimalEngine.Backup.restore("/var/backups/engine-2026-01-05.db")

  The backup is not encrypted. Treat the file as sensitive —
  connector credentials live in it (encrypted envelope, but the
  envelope key lives outside the DB by design).
  """

  alias OptimalEngine.Store

  @type backup_result :: %{
          target: String.t(),
          size_bytes: non_neg_integer(),
          rows_backed_up: non_neg_integer(),
          duration_ms: non_neg_integer()
        }

  @doc """
  Create a backup at `target_path`. Runs online — the engine keeps
  serving throughout. Parent directory must exist.
  """
  @spec create(String.t()) :: {:ok, backup_result()} | {:error, term()}
  def create(target_path) when is_binary(target_path) do
    parent = Path.dirname(target_path)

    with :ok <- ensure_parent_dir(parent),
         :ok <- ensure_no_existing(target_path),
         started <- System.monotonic_time(:millisecond),
         {:ok, _} <- Store.raw_query("VACUUM INTO ?1", [target_path]),
         size <- File.stat!(target_path).size,
         rows <- count_backup_rows(target_path) do
      {:ok,
       %{
         target: target_path,
         size_bytes: size,
         rows_backed_up: rows,
         duration_ms: System.monotonic_time(:millisecond) - started
       }}
    end
  end

  @doc """
  Restore from `source_path`. Refuses to run while the supervisor is
  up — restore must happen pre-boot.

  The target is whatever `:optimal_engine, :db_path` resolves to. We
  don't overwrite without a prior copy of the existing DB at
  `<db_path>.pre-restore-<epoch>` so the operator has a rollback.
  """
  @spec restore(String.t()) :: {:ok, String.t()} | {:error, term()}
  def restore(source_path) when is_binary(source_path) do
    db_path = Application.get_env(:optimal_engine, :db_path)

    with :ok <- refuse_if_supervisor_up(),
         :ok <- ensure_source_exists(source_path),
         {:ok, backup_of_current} <- archive_current(db_path) do
      File.cp!(source_path, db_path)
      {:ok, backup_of_current}
    end
  end

  @doc """
  Quick integrity check on a backup file — opens it as a read-only
  SQLite, runs `PRAGMA integrity_check`, and closes. Returns
  `{:ok, :ok}` for clean DBs or `{:ok, issues}` for anything else.
  """
  @spec verify(String.t()) :: {:ok, term()} | {:error, term()}
  def verify(path) when is_binary(path) do
    with {:ok, db} <- Exqlite.Sqlite3.open(path, mode: :readonly),
         {:ok, stmt} <- Exqlite.Sqlite3.prepare(db, "PRAGMA integrity_check"),
         :ok <- :ok,
         rows <- drain(db, stmt, []),
         _ <- Exqlite.Sqlite3.release(db, stmt),
         :ok <- Exqlite.Sqlite3.close(db) do
      case rows do
        [["ok"]] -> {:ok, :ok}
        issues -> {:ok, issues}
      end
    end
  end

  # ─── private ─────────────────────────────────────────────────────────────

  defp ensure_parent_dir(parent) do
    if File.dir?(parent), do: :ok, else: {:error, {:parent_missing, parent}}
  end

  defp ensure_no_existing(path) do
    if File.exists?(path), do: {:error, {:target_exists, path}}, else: :ok
  end

  defp ensure_source_exists(path) do
    if File.exists?(path), do: :ok, else: {:error, {:source_missing, path}}
  end

  defp refuse_if_supervisor_up do
    case Process.whereis(OptimalEngine.Supervisor) do
      nil -> :ok
      _ -> {:error, :supervisor_still_running}
    end
  end

  defp archive_current(db_path) do
    if File.exists?(db_path) do
      epoch = System.os_time(:second)
      dest = "#{db_path}.pre-restore-#{epoch}"
      File.cp!(db_path, dest)
      {:ok, dest}
    else
      {:ok, nil}
    end
  end

  # Sum of `COUNT(*)` across the primary engine tables — gives a
  # rough "how much data is in this backup" number without opening
  # every table.
  defp count_backup_rows(path) do
    try do
      {:ok, db} = Exqlite.Sqlite3.open(path, mode: :readonly)

      total =
        ~w(contexts signals wiki_pages connectors chunks citations)
        |> Enum.reduce(0, fn table, acc ->
          case exec(db, "SELECT COUNT(*) FROM #{table}") do
            [[n]] when is_integer(n) -> acc + n
            _ -> acc
          end
        end)

      Exqlite.Sqlite3.close(db)
      total
    rescue
      _ -> 0
    end
  end

  defp exec(db, sql) do
    with {:ok, stmt} <- Exqlite.Sqlite3.prepare(db, sql),
         rows <- drain(db, stmt, []) do
      Exqlite.Sqlite3.release(db, stmt)
      rows
    else
      _ -> []
    end
  end

  defp drain(db, stmt, acc) do
    case Exqlite.Sqlite3.step(db, stmt) do
      {:row, row} -> drain(db, stmt, [row | acc])
      :done -> Enum.reverse(acc)
      _ -> Enum.reverse(acc)
    end
  end
end
