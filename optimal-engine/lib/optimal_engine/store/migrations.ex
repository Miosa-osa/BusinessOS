defmodule OptimalEngine.Store.Migrations do
  @moduledoc """
  Versioned schema migrations for the Optimal Engine.

  Each migration is `{version, description, ddl_statements}` where
  `ddl_statements` is a list of `{label, sql}` pairs. Migrations run in order,
  tracked in the `schema_migrations` table. Idempotent: already-applied
  versions are skipped; statements that would duplicate (columns, indexes)
  are recognized and tolerated via `safe_execute/3`.

  ## When to add a migration

  - Schema change (new table, new column, new index, new trigger).
  - Data backfill that must run once per database.

  ## When NOT to add a migration

  - Pure code change with no schema impact.
  - Data operations that can run on-demand from a Mix task.

  ## Contract with `Store.open_and_migrate/1`

  `Store.open_and_migrate/1` creates its baseline tables (contexts, entities,
  edges, decisions, sessions, vectors, observations), then calls
  `Migrations.run/1` which applies anything this module tracks. Baseline
  DDL stays inline in `Store` for minimal disruption; additive work
  (Phase 1 onward) lives here.
  """

  require Logger

  @type migration :: {
          version :: pos_integer(),
          description :: String.t(),
          statements :: [{label :: String.t(), sql :: String.t()}]
        }

  # ---------------------------------------------------------------------------
  # Migration registry
  # ---------------------------------------------------------------------------

  @doc """
  Returns all migrations in ascending-version order. Add new migrations at
  the end; never rewrite an existing version.
  """
  @spec all() :: [migration()]
  def all do
    [
      migration_001_schema_migrations_table(),
      migration_002_tenancy(),
      migration_003_identity(),
      migration_004_acls(),
      migration_005_chunks_classifications_intents(),
      migration_006_assets(),
      migration_007_clusters(),
      migration_008_wiki(),
      migration_009_connectors(),
      migration_010_retention_legal_hold(),
      migration_011_audiences(),
      migration_012_events(),
      migration_013_tenant_id_on_existing_tables(),
      migration_014_tenant_first_indexes(),
      migration_015_default_tenant_seed(),
      migration_016_backfill_document_chunks(),
      migration_017_workspace_nodes(),
      migration_018_node_members(),
      migration_019_skills(),
      migration_020_principal_skills(),
      migration_021_workspace_indexes(),
      migration_022_backfill_nodes_from_contexts(),
      migration_023_chunk_embeddings(),
      migration_024_compliance_columns(),
      migration_025_data_architectures()
    ]
  end

  # ---------------------------------------------------------------------------
  # Runner
  # ---------------------------------------------------------------------------

  @doc """
  Applies any migrations with `version > max(schema_migrations.version)`.
  Records each successful migration in the `schema_migrations` table.
  """
  @spec run(any()) :: :ok
  def run(db) do
    ensure_migrations_table!(db)
    applied = applied_versions(db)

    pending =
      all()
      |> Enum.reject(fn {version, _desc, _stmts} -> MapSet.member?(applied, version) end)

    Enum.each(pending, fn {version, description, statements} ->
      Logger.info("[Migrations] Applying #{pad(version)} — #{description}")
      Enum.each(statements, fn {label, sql} -> safe_execute(db, label, sql) end)
      record_migration!(db, version, description)
    end)

    :ok
  end

  @doc """
  Returns the set of applied migration versions.
  """
  @spec applied_versions(any()) :: MapSet.t(pos_integer())
  def applied_versions(db) do
    {:ok, stmt} = Exqlite.Sqlite3.prepare(db, "SELECT version FROM schema_migrations")

    versions = collect_versions(db, stmt, [])
    Exqlite.Sqlite3.release(db, stmt)
    MapSet.new(versions)
  rescue
    _ -> MapSet.new()
  end

  # ---------------------------------------------------------------------------
  # Private — migrations
  # ---------------------------------------------------------------------------

  defp migration_001_schema_migrations_table do
    {1, "schema_migrations tracking table",
     [
       {"schema_migrations",
        """
        CREATE TABLE IF NOT EXISTS schema_migrations (
          version INTEGER PRIMARY KEY,
          applied_at TEXT NOT NULL DEFAULT (datetime('now')),
          description TEXT
        )
        """}
     ]}
  end

  defp migration_002_tenancy do
    {2, "tenants",
     [
       {"tenants",
        """
        CREATE TABLE IF NOT EXISTS tenants (
          id TEXT PRIMARY KEY,
          name TEXT NOT NULL,
          plan TEXT NOT NULL DEFAULT 'default',
          region TEXT,
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          metadata TEXT NOT NULL DEFAULT '{}'
        )
        """}
     ]}
  end

  defp migration_003_identity do
    {3, "principals, groups, roles, role_grants",
     [
       {"principals",
        """
        CREATE TABLE IF NOT EXISTS principals (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
          kind TEXT NOT NULL,
          display_name TEXT NOT NULL,
          external_id TEXT,
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          metadata TEXT NOT NULL DEFAULT '{}',
          UNIQUE(tenant_id, external_id)
        )
        """},
       {"groups",
        """
        CREATE TABLE IF NOT EXISTS groups (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
          name TEXT NOT NULL,
          source TEXT NOT NULL DEFAULT 'local',
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(tenant_id, name)
        )
        """},
       {"principal_groups",
        """
        CREATE TABLE IF NOT EXISTS principal_groups (
          principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
          group_id TEXT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
          PRIMARY KEY (principal_id, group_id)
        )
        """},
       {"roles",
        """
        CREATE TABLE IF NOT EXISTS roles (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
          name TEXT NOT NULL,
          description TEXT,
          UNIQUE(tenant_id, name)
        )
        """},
       {"role_grants",
        """
        CREATE TABLE IF NOT EXISTS role_grants (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          principal_id TEXT REFERENCES principals(id) ON DELETE CASCADE,
          group_id TEXT REFERENCES groups(id) ON DELETE CASCADE,
          role_id TEXT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
          granted_at TEXT NOT NULL DEFAULT (datetime('now')),
          CHECK ((principal_id IS NOT NULL) <> (group_id IS NOT NULL))
        )
        """}
     ]}
  end

  defp migration_004_acls do
    {4, "acls",
     [
       {"acls",
        """
        CREATE TABLE IF NOT EXISTS acls (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          resource_uri TEXT NOT NULL,
          principal_id TEXT,
          group_id TEXT,
          permission TEXT NOT NULL,
          granted_at TEXT NOT NULL DEFAULT (datetime('now')),
          CHECK ((principal_id IS NOT NULL) <> (group_id IS NOT NULL))
        )
        """}
     ]}
  end

  defp migration_005_chunks_classifications_intents do
    {5, "chunks + classifications + intents",
     [
       {"chunks",
        """
        CREATE TABLE IF NOT EXISTS chunks (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          signal_id TEXT NOT NULL,
          parent_id TEXT,
          scale TEXT NOT NULL,
          offset_bytes INTEGER NOT NULL DEFAULT 0,
          length_bytes INTEGER NOT NULL DEFAULT 0,
          text TEXT NOT NULL DEFAULT '',
          modality TEXT NOT NULL DEFAULT 'text',
          asset_ref TEXT,
          classification_level TEXT NOT NULL DEFAULT 'internal',
          created_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """},
       {"classifications",
        """
        CREATE TABLE IF NOT EXISTS classifications (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          chunk_id TEXT NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
          mode TEXT,
          genre TEXT,
          signal_type TEXT,
          format TEXT,
          structure TEXT,
          sn_ratio REAL,
          confidence REAL,
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(chunk_id)
        )
        """},
       {"intents",
        """
        CREATE TABLE IF NOT EXISTS intents (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          chunk_id TEXT NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
          intent TEXT NOT NULL,
          confidence REAL NOT NULL,
          evidence TEXT,
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(chunk_id)
        )
        """}
     ]}
  end

  defp migration_006_assets do
    {6, "assets",
     [
       {"assets",
        """
        CREATE TABLE IF NOT EXISTS assets (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          content_type TEXT NOT NULL,
          size_bytes INTEGER NOT NULL,
          storage_path TEXT NOT NULL,
          created_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """}
     ]}
  end

  defp migration_007_clusters do
    {7, "clusters + cluster_members",
     [
       {"clusters",
        """
        CREATE TABLE IF NOT EXISTS clusters (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          theme TEXT NOT NULL,
          intent_dominant TEXT,
          member_count INTEGER NOT NULL DEFAULT 0,
          centroid BLOB,
          updated_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """},
       {"cluster_members",
        """
        CREATE TABLE IF NOT EXISTS cluster_members (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          cluster_id TEXT NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
          chunk_id TEXT NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
          weight REAL NOT NULL DEFAULT 1.0,
          UNIQUE(cluster_id, chunk_id)
        )
        """}
     ]}
  end

  defp migration_008_wiki do
    {8, "wiki_pages + citations",
     [
       {"wiki_pages",
        """
        CREATE TABLE IF NOT EXISTS wiki_pages (
          tenant_id TEXT NOT NULL,
          slug TEXT NOT NULL,
          audience TEXT NOT NULL DEFAULT 'default',
          version INTEGER NOT NULL DEFAULT 1,
          frontmatter TEXT NOT NULL DEFAULT '{}',
          body TEXT NOT NULL,
          last_curated TEXT NOT NULL DEFAULT (datetime('now')),
          curated_by TEXT,
          PRIMARY KEY (tenant_id, slug, audience, version)
        )
        """},
       {"citations",
        """
        CREATE TABLE IF NOT EXISTS citations (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          wiki_slug TEXT NOT NULL,
          wiki_audience TEXT NOT NULL DEFAULT 'default',
          chunk_id TEXT NOT NULL REFERENCES chunks(id) ON DELETE CASCADE,
          claim_hash TEXT NOT NULL,
          last_verified TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """}
     ]}
  end

  defp migration_009_connectors do
    {9, "connectors + connector_runs",
     [
       {"connectors",
        """
        CREATE TABLE IF NOT EXISTS connectors (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          kind TEXT NOT NULL,
          config TEXT NOT NULL DEFAULT '{}',
          cursor TEXT,
          enabled INTEGER NOT NULL DEFAULT 1,
          created_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """},
       {"connector_runs",
        """
        CREATE TABLE IF NOT EXISTS connector_runs (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          connector_id TEXT NOT NULL REFERENCES connectors(id) ON DELETE CASCADE,
          started_at TEXT NOT NULL DEFAULT (datetime('now')),
          completed_at TEXT,
          signals_ingested INTEGER NOT NULL DEFAULT 0,
          errors_encountered INTEGER NOT NULL DEFAULT 0,
          cursor_before TEXT,
          cursor_after TEXT,
          status TEXT NOT NULL DEFAULT 'running',
          error_detail TEXT
        )
        """}
     ]}
  end

  defp migration_010_retention_legal_hold do
    {10, "retention_policies + legal_holds",
     [
       {"retention_policies",
        """
        CREATE TABLE IF NOT EXISTS retention_policies (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          scope_type TEXT NOT NULL,
          scope_value TEXT,
          ttl_days INTEGER,
          action TEXT NOT NULL DEFAULT 'archive',
          created_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """},
       {"legal_holds",
        """
        CREATE TABLE IF NOT EXISTS legal_holds (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          signal_id TEXT NOT NULL,
          held_by TEXT NOT NULL,
          reason TEXT NOT NULL,
          placed_at TEXT NOT NULL DEFAULT (datetime('now')),
          released_at TEXT
        )
        """}
     ]}
  end

  defp migration_011_audiences do
    {11, "audiences",
     [
       {"audiences",
        """
        CREATE TABLE IF NOT EXISTS audiences (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          name TEXT NOT NULL,
          role_ids TEXT NOT NULL DEFAULT '[]',
          description TEXT,
          UNIQUE(tenant_id, name)
        )
        """}
     ]}
  end

  defp migration_012_events do
    {12, "events (append-only audit log)",
     [
       {"events",
        """
        CREATE TABLE IF NOT EXISTS events (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          ts TEXT NOT NULL DEFAULT (datetime('now')),
          principal TEXT NOT NULL,
          kind TEXT NOT NULL,
          target_uri TEXT,
          latency_ms INTEGER,
          metadata TEXT NOT NULL DEFAULT '{}'
        )
        """}
     ]}
  end

  defp migration_013_tenant_id_on_existing_tables do
    {13, "add tenant_id to existing primary tables",
     [
       {"contexts.tenant_id",
        "ALTER TABLE contexts ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"entities.tenant_id",
        "ALTER TABLE entities ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"edges.tenant_id",
        "ALTER TABLE edges ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"decisions.tenant_id",
        "ALTER TABLE decisions ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"sessions.tenant_id",
        "ALTER TABLE sessions ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"vectors.tenant_id",
        "ALTER TABLE vectors ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"},
       {"observations.tenant_id",
        "ALTER TABLE observations ADD COLUMN tenant_id TEXT NOT NULL DEFAULT 'default'"}
     ]}
  end

  defp migration_014_tenant_first_indexes do
    {14, "tenant-first indexes on every primary table",
     [
       {"idx_contexts_tenant_node",
        "CREATE INDEX IF NOT EXISTS idx_contexts_tenant_node ON contexts(tenant_id, node)"},
       {"idx_contexts_tenant_type",
        "CREATE INDEX IF NOT EXISTS idx_contexts_tenant_type ON contexts(tenant_id, type)"},
       {"idx_chunks_tenant_scale",
        "CREATE INDEX IF NOT EXISTS idx_chunks_tenant_scale ON chunks(tenant_id, scale)"},
       {"idx_chunks_tenant_signal",
        "CREATE INDEX IF NOT EXISTS idx_chunks_tenant_signal ON chunks(tenant_id, signal_id)"},
       {"idx_chunks_parent", "CREATE INDEX IF NOT EXISTS idx_chunks_parent ON chunks(parent_id)"},
       {"idx_classifications_tenant_chunk",
        "CREATE INDEX IF NOT EXISTS idx_classifications_tenant_chunk ON classifications(tenant_id, chunk_id)"},
       {"idx_intents_tenant_chunk",
        "CREATE INDEX IF NOT EXISTS idx_intents_tenant_chunk ON intents(tenant_id, chunk_id)"},
       {"idx_cluster_members_tenant_cluster",
        "CREATE INDEX IF NOT EXISTS idx_cluster_members_tenant_cluster ON cluster_members(tenant_id, cluster_id)"},
       {"idx_cluster_members_tenant_chunk",
        "CREATE INDEX IF NOT EXISTS idx_cluster_members_tenant_chunk ON cluster_members(tenant_id, chunk_id)"},
       {"idx_wiki_pages_tenant_slug",
        "CREATE INDEX IF NOT EXISTS idx_wiki_pages_tenant_slug ON wiki_pages(tenant_id, slug)"},
       {"idx_citations_tenant_slug",
        "CREATE INDEX IF NOT EXISTS idx_citations_tenant_slug ON citations(tenant_id, wiki_slug)"},
       {"idx_citations_tenant_chunk",
        "CREATE INDEX IF NOT EXISTS idx_citations_tenant_chunk ON citations(tenant_id, chunk_id)"},
       {"idx_events_tenant_ts",
        "CREATE INDEX IF NOT EXISTS idx_events_tenant_ts ON events(tenant_id, ts)"},
       {"idx_events_tenant_principal",
        "CREATE INDEX IF NOT EXISTS idx_events_tenant_principal ON events(tenant_id, principal, ts)"},
       {"idx_events_tenant_kind",
        "CREATE INDEX IF NOT EXISTS idx_events_tenant_kind ON events(tenant_id, kind, ts)"},
       {"idx_acls_tenant_resource",
        "CREATE INDEX IF NOT EXISTS idx_acls_tenant_resource ON acls(tenant_id, resource_uri)"},
       {"idx_acls_tenant_principal",
        "CREATE INDEX IF NOT EXISTS idx_acls_tenant_principal ON acls(tenant_id, principal_id)"},
       {"idx_acls_tenant_group",
        "CREATE INDEX IF NOT EXISTS idx_acls_tenant_group ON acls(tenant_id, group_id)"},
       {"idx_principal_groups_principal",
        "CREATE INDEX IF NOT EXISTS idx_principal_groups_principal ON principal_groups(principal_id)"},
       {"idx_principal_groups_group",
        "CREATE INDEX IF NOT EXISTS idx_principal_groups_group ON principal_groups(group_id)"},
       {"idx_role_grants_tenant_principal",
        "CREATE INDEX IF NOT EXISTS idx_role_grants_tenant_principal ON role_grants(tenant_id, principal_id)"},
       {"idx_role_grants_tenant_group",
        "CREATE INDEX IF NOT EXISTS idx_role_grants_tenant_group ON role_grants(tenant_id, group_id)"},
       {"idx_connector_runs_connector",
        "CREATE INDEX IF NOT EXISTS idx_connector_runs_connector ON connector_runs(connector_id, started_at)"},
       {"idx_legal_holds_tenant_signal",
        "CREATE INDEX IF NOT EXISTS idx_legal_holds_tenant_signal ON legal_holds(tenant_id, signal_id)"}
     ]}
  end

  defp migration_015_default_tenant_seed do
    {15, "seed the default tenant",
     [
       {"default_tenant",
        """
        INSERT OR IGNORE INTO tenants (id, name, plan)
        VALUES ('default', 'Default Tenant', 'default')
        """}
     ]}
  end

  defp migration_017_workspace_nodes do
    {17, "workspace: nodes (organizational units)",
     [
       {"nodes",
        """
        CREATE TABLE IF NOT EXISTS nodes (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          slug TEXT NOT NULL,
          name TEXT NOT NULL,
          kind TEXT NOT NULL,
          parent_id TEXT REFERENCES nodes(id) ON DELETE CASCADE,
          description TEXT,
          style TEXT NOT NULL DEFAULT 'internal',
          status TEXT NOT NULL DEFAULT 'active',
          path TEXT NOT NULL DEFAULT '',
          metadata TEXT NOT NULL DEFAULT '{}',
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(tenant_id, slug)
        )
        """}
     ]}
  end

  defp migration_018_node_members do
    {18, "workspace: node_members (principal ↔ node with internal/external/owner/observer)",
     [
       {"node_members",
        """
        CREATE TABLE IF NOT EXISTS node_members (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
          principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
          membership TEXT NOT NULL DEFAULT 'internal',
          role TEXT,
          started_at TEXT NOT NULL DEFAULT (datetime('now')),
          ended_at TEXT,
          UNIQUE(node_id, principal_id, membership)
        )
        """}
     ]}
  end

  defp migration_019_skills do
    {19, "workspace: skills (capability registry)",
     [
       {"skills",
        """
        CREATE TABLE IF NOT EXISTS skills (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL,
          name TEXT NOT NULL,
          kind TEXT,
          description TEXT,
          UNIQUE(tenant_id, name)
        )
        """}
     ]}
  end

  defp migration_020_principal_skills do
    {20, "workspace: principal_skills (many-to-many capability grants)",
     [
       {"principal_skills",
        """
        CREATE TABLE IF NOT EXISTS principal_skills (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          principal_id TEXT NOT NULL REFERENCES principals(id) ON DELETE CASCADE,
          skill_id TEXT NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
          level TEXT NOT NULL DEFAULT 'intermediate',
          evidence TEXT,
          acquired_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(principal_id, skill_id)
        )
        """}
     ]}
  end

  defp migration_021_workspace_indexes do
    {21, "workspace: tenant-first indexes",
     [
       {"idx_nodes_tenant_kind",
        "CREATE INDEX IF NOT EXISTS idx_nodes_tenant_kind ON nodes(tenant_id, kind)"},
       {"idx_nodes_tenant_parent",
        "CREATE INDEX IF NOT EXISTS idx_nodes_tenant_parent ON nodes(tenant_id, parent_id)"},
       {"idx_nodes_tenant_style",
        "CREATE INDEX IF NOT EXISTS idx_nodes_tenant_style ON nodes(tenant_id, style)"},
       {"idx_node_members_tenant_node",
        "CREATE INDEX IF NOT EXISTS idx_node_members_tenant_node ON node_members(tenant_id, node_id)"},
       {"idx_node_members_tenant_principal",
        "CREATE INDEX IF NOT EXISTS idx_node_members_tenant_principal ON node_members(tenant_id, principal_id)"},
       {"idx_skills_tenant_kind",
        "CREATE INDEX IF NOT EXISTS idx_skills_tenant_kind ON skills(tenant_id, kind)"},
       {"idx_principal_skills_tenant_principal",
        "CREATE INDEX IF NOT EXISTS idx_principal_skills_tenant_principal ON principal_skills(tenant_id, principal_id)"},
       {"idx_principal_skills_tenant_skill",
        "CREATE INDEX IF NOT EXISTS idx_principal_skills_tenant_skill ON principal_skills(tenant_id, skill_id)"}
     ]}
  end

  # Backfill: every distinct `contexts.node` value becomes a nodes row so
  # existing routing / retrieval paths continue to resolve, with the node
  # now first-class and upgradable (kind / style / memberships can evolve).
  defp migration_022_backfill_nodes_from_contexts do
    {22, "workspace: backfill nodes rows from distinct contexts.node values",
     [
       {"backfill_nodes",
        """
        INSERT OR IGNORE INTO nodes (id, tenant_id, slug, name, kind, style, status, path)
        SELECT
          COALESCE(c.tenant_id, 'default') || ':' || c.node AS id,
          COALESCE(c.tenant_id, 'default')                  AS tenant_id,
          c.node                                            AS slug,
          c.node                                            AS name,
          'domain'                                          AS kind,
          'internal'                                        AS style,
          'active'                                          AS status,
          'nodes/' || c.node                                AS path
        FROM (SELECT DISTINCT tenant_id, node FROM contexts WHERE node IS NOT NULL AND node <> '') c
        """}
     ]}
  end

  # Phase 5: chunk-level embeddings in the aligned 768-dim nomic space.
  # Keyed on chunk_id so re-embedding overwrites in place. `modality` is the
  # original chunk modality (text/image/audio/code/data/mixed) so a single
  # query can filter to a subset of modalities if needed.
  defp migration_025_data_architectures do
    {25, "data-architectures — model-agnostic data-point schemas (Phase 14)",
     [
       {"architectures",
        """
        CREATE TABLE IF NOT EXISTS architectures (
          id TEXT PRIMARY KEY,
          tenant_id TEXT NOT NULL DEFAULT 'default',
          name TEXT NOT NULL,
          version INTEGER NOT NULL DEFAULT 1,
          description TEXT,
          modality_primary TEXT NOT NULL DEFAULT 'text',
          spec TEXT NOT NULL DEFAULT '{}',
          created_at TEXT NOT NULL DEFAULT (datetime('now')),
          UNIQUE(tenant_id, name, version)
        )
        """},
       {"idx_architectures_tenant",
        "CREATE INDEX IF NOT EXISTS idx_architectures_tenant ON architectures(tenant_id, name)"},
       {"contexts.architecture_id", "ALTER TABLE contexts ADD COLUMN architecture_id TEXT"},
       {"idx_contexts_architecture",
        "CREATE INDEX IF NOT EXISTS idx_contexts_architecture ON contexts(tenant_id, architecture_id)"},
       {"processor_runs",
        """
        CREATE TABLE IF NOT EXISTS processor_runs (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          tenant_id TEXT NOT NULL,
          context_id TEXT NOT NULL,
          architecture_id TEXT NOT NULL,
          processor TEXT NOT NULL,
          field TEXT NOT NULL,
          status TEXT NOT NULL DEFAULT 'pending',
          started_at TEXT NOT NULL DEFAULT (datetime('now')),
          completed_at TEXT,
          output_ref TEXT,
          metadata TEXT NOT NULL DEFAULT '{}'
        )
        """},
       {"idx_processor_runs_context",
        "CREATE INDEX IF NOT EXISTS idx_processor_runs_context ON processor_runs(context_id, processor)"}
     ]}
  end

  defp migration_024_compliance_columns do
    {24, "compliance — contexts.created_by + archived_at for Phase 11",
     [
       {"contexts.created_by", "ALTER TABLE contexts ADD COLUMN created_by TEXT"},
       {"contexts.archived_at", "ALTER TABLE contexts ADD COLUMN archived_at TEXT"},
       {"idx_contexts_tenant_created_by",
        "CREATE INDEX IF NOT EXISTS idx_contexts_tenant_created_by ON contexts(tenant_id, created_by)"}
     ]}
  end

  defp migration_023_chunk_embeddings do
    {23, "chunk_embeddings — aligned 768-dim per-chunk vectors (Phase 5)",
     [
       {"chunk_embeddings",
        """
        CREATE TABLE IF NOT EXISTS chunk_embeddings (
          chunk_id TEXT PRIMARY KEY REFERENCES chunks(id) ON DELETE CASCADE,
          tenant_id TEXT NOT NULL,
          model TEXT NOT NULL,
          modality TEXT NOT NULL,
          dim INTEGER NOT NULL DEFAULT 768,
          vector BLOB NOT NULL,
          created_at TEXT NOT NULL DEFAULT (datetime('now'))
        )
        """},
       {"idx_chunk_embeddings_tenant_modality",
        "CREATE INDEX IF NOT EXISTS idx_chunk_embeddings_tenant_modality ON chunk_embeddings(tenant_id, modality)"},
       {"idx_chunk_embeddings_tenant_model",
        "CREATE INDEX IF NOT EXISTS idx_chunk_embeddings_tenant_model ON chunk_embeddings(tenant_id, model)"}
     ]}
  end

  # Backfill: for each existing contexts row, create a corresponding
  # :document-scale chunk so retrieval paths that expect chunks have something
  # to read. Chunk id = "{context_id}:doc" (deterministic + idempotent).
  defp migration_016_backfill_document_chunks do
    {16, "backfill :document-scale chunks from existing contexts",
     [
       {"backfill_chunks",
        """
        INSERT OR IGNORE INTO chunks
          (id, tenant_id, signal_id, parent_id, scale, offset_bytes, length_bytes, text, modality, classification_level, created_at)
        SELECT
          c.id || ':doc',
          COALESCE(c.tenant_id, 'default'),
          c.id,
          NULL,
          'document',
          0,
          COALESCE(LENGTH(c.content), 0),
          COALESCE(c.content, ''),
          CASE WHEN COALESCE(c.mode, '') = 'code' THEN 'code' ELSE 'text' END,
          'internal',
          COALESCE(c.created_at, datetime('now'))
        FROM contexts c
        """}
     ]}
  end

  # ---------------------------------------------------------------------------
  # Private — helpers
  # ---------------------------------------------------------------------------

  defp ensure_migrations_table!(db) do
    # Apply migration 001 unconditionally — it's the bootstrap.
    {1, _desc, stmts} = migration_001_schema_migrations_table()
    Enum.each(stmts, fn {label, sql} -> safe_execute(db, label, sql) end)
    :ok
  end

  defp collect_versions(db, stmt, acc) do
    case Exqlite.Sqlite3.step(db, stmt) do
      {:row, [version]} -> collect_versions(db, stmt, [version | acc])
      :done -> acc
      {:error, _} -> acc
    end
  end

  defp record_migration!(db, version, description) do
    {:ok, stmt} =
      Exqlite.Sqlite3.prepare(
        db,
        "INSERT OR IGNORE INTO schema_migrations (version, description) VALUES (?1, ?2)"
      )

    :ok = Exqlite.Sqlite3.bind(stmt, [version, description])
    :done = Exqlite.Sqlite3.step(db, stmt)
    Exqlite.Sqlite3.release(db, stmt)
    :ok
  end

  # Executes a statement; tolerates duplicate-column / duplicate-index errors
  # so re-runs never fail. Any unexpected error is logged but not raised
  # (migrations continue) so a single broken statement doesn't brick the store.
  defp safe_execute(db, label, sql) do
    case Exqlite.Sqlite3.execute(db, sql) do
      :ok ->
        :ok

      {:error, msg} when is_binary(msg) ->
        cond do
          String.contains?(msg, "duplicate column") ->
            :ok

          String.contains?(msg, "already exists") ->
            :ok

          true ->
            Logger.warning("[Migrations] #{label} failed: #{msg}")
            :ok
        end

      {:error, reason} ->
        Logger.warning("[Migrations] #{label} failed: #{inspect(reason)}")
        :ok
    end
  end

  defp pad(n) when is_integer(n), do: n |> to_string() |> String.pad_leading(3, "0")
end
