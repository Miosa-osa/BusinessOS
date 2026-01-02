// Apply Database Migrations to Supabase
// This script reads migration files and applies them to your Supabase database

const { Client } = require('pg');
const fs = require('fs');
const path = require('path');

// Supabase connection details
const SUPABASE_CONFIG = {
  host: 'db.fuqhjbgbjamtxcdphjpp.supabase.co',
  port: 5432, // Direct connection port
  database: 'postgres',
  user: 'postgres',
  password: 'fmm6Wt7kN0ajrjxK',
  ssl: {
    rejectUnauthorized: false
  }
};

// Migration files in order (based on pedro_tasks_v2.md)
const MIGRATION_FILES = [
  '016_memories.sql',
  '017_context_system.sql',
  '018_output_styles.sql',
  '019_documents_no_vector.sql', // Using no_vector version for compatibility
  '020_context_integration_no_vector.sql',
  '021_learning_system.sql',
  '022_application_profiles_no_vector.sql',
  '023_pedro_tasks_schema_fix.sql',
  '024_embedding_dimensions_768.sql'
];

const MIGRATIONS_DIR = path.join(__dirname, 'desktop', 'backend-go', 'internal', 'database', 'migrations');

async function checkConnection(client) {
  try {
    await client.connect();
    const result = await client.query('SELECT version()');
    console.log('✅ Connected to Supabase PostgreSQL');
    console.log(`   Version: ${result.rows[0].version.split(' ')[0]} ${result.rows[0].version.split(' ')[1]}\n`);
    return true;
  } catch (error) {
    console.error('❌ Connection failed:', error.message);
    console.error('   Code:', error.code, '\n');
    return false;
  }
}

async function createMigrationsTable(client) {
  const createTableSQL = `
    CREATE TABLE IF NOT EXISTS schema_migrations (
      id SERIAL PRIMARY KEY,
      migration_name VARCHAR(255) UNIQUE NOT NULL,
      applied_at TIMESTAMPTZ DEFAULT NOW()
    );
  `;

  try {
    await client.query(createTableSQL);
    console.log('✅ schema_migrations table ready\n');
  } catch (error) {
    console.error('❌ Failed to create schema_migrations table:', error.message);
    throw error;
  }
}

async function getAppliedMigrations(client) {
  try {
    const result = await client.query(
      'SELECT migration_name FROM schema_migrations ORDER BY applied_at'
    );
    return result.rows.map(row => row.migration_name);
  } catch (error) {
    console.error('⚠️  Could not read applied migrations:', error.message);
    return [];
  }
}

async function applyMigration(client, migrationFile, migrationPath) {
  console.log(`\n📄 Applying: ${migrationFile}`);
  console.log('─'.repeat(60));

  try {
    // Read migration file
    const sql = fs.readFileSync(migrationPath, 'utf8');

    if (!sql || sql.trim().length === 0) {
      console.log('⚠️  Migration file is empty, skipping...');
      return true;
    }

    // Execute migration
    await client.query('BEGIN');

    try {
      await client.query(sql);

      // Record migration as applied
      await client.query(
        'INSERT INTO schema_migrations (migration_name) VALUES ($1) ON CONFLICT (migration_name) DO NOTHING',
        [migrationFile]
      );

      await client.query('COMMIT');
      console.log(`✅ Successfully applied: ${migrationFile}`);
      return true;
    } catch (execError) {
      await client.query('ROLLBACK');
      throw execError;
    }
  } catch (error) {
    console.error(`❌ Failed to apply ${migrationFile}:`);
    console.error(`   ${error.message}`);

    // Show helpful context
    if (error.message.includes('does not exist')) {
      console.error('   💡 Tip: Check if this migration depends on a previous one');
    } else if (error.message.includes('already exists')) {
      console.error('   💡 Tip: This object might already exist, you can skip this');
    }

    return false;
  }
}

async function generateSQLFile(migrations) {
  console.log('\n📝 Generating combined SQL file for manual application...\n');

  let combinedSQL = `-- Combined Migrations for Supabase
-- Generated: ${new Date().toISOString()}
-- Apply this via Supabase SQL Editor if direct connection fails

`;

  for (const migrationFile of migrations) {
    const migrationPath = path.join(MIGRATIONS_DIR, migrationFile);

    if (!fs.existsSync(migrationPath)) {
      console.log(`⚠️  File not found: ${migrationFile}`);
      continue;
    }

    const sql = fs.readFileSync(migrationPath, 'utf8');
    combinedSQL += `\n-- ================================================\n`;
    combinedSQL += `-- Migration: ${migrationFile}\n`;
    combinedSQL += `-- ================================================\n\n`;
    combinedSQL += sql;
    combinedSQL += `\n\n`;
  }

  const outputPath = path.join(__dirname, 'supabase-migrations-combined.sql');
  fs.writeFileSync(outputPath, combinedSQL, 'utf8');

  console.log(`✅ Combined SQL file created: supabase-migrations-combined.sql`);
  console.log(`\n📋 Manual Application Steps:`);
  console.log(`   1. Go to: https://supabase.com/dashboard/project/fuqhjbgbjamtxcdphjpp/sql`);
  console.log(`   2. Open: supabase-migrations-combined.sql`);
  console.log(`   3. Copy the entire content`);
  console.log(`   4. Paste into Supabase SQL Editor`);
  console.log(`   5. Click "RUN"\n`);
}

async function main() {
  console.log('╔════════════════════════════════════════════════════════╗');
  console.log('║        Supabase Migration Application Tool             ║');
  console.log('╚════════════════════════════════════════════════════════╝\n');

  // Verify migration files exist
  console.log('📂 Verifying migration files...\n');
  const missingFiles = [];

  for (const file of MIGRATION_FILES) {
    const filePath = path.join(MIGRATIONS_DIR, file);
    if (fs.existsSync(filePath)) {
      console.log(`✅ Found: ${file}`);
    } else {
      console.log(`❌ Missing: ${file}`);
      missingFiles.push(file);
    }
  }

  if (missingFiles.length > 0) {
    console.error('\n❌ Some migration files are missing. Cannot proceed.');
    return;
  }

  console.log(`\n✅ All ${MIGRATION_FILES.length} migration files found!\n`);

  // Try to connect to Supabase
  console.log('🔌 Attempting to connect to Supabase...\n');
  const client = new Client(SUPABASE_CONFIG);

  const connected = await checkConnection(client);

  if (!connected) {
    console.log('⚠️  Direct database connection failed.\n');
    console.log('🔄 Generating combined SQL file instead...\n');
    await generateSQLFile(MIGRATION_FILES);
    process.exit(0);
  }

  try {
    // Create migrations tracking table
    await createMigrationsTable(client);

    // Check which migrations are already applied
    const appliedMigrations = await getAppliedMigrations(client);
    console.log(`📊 Previously applied migrations: ${appliedMigrations.length}\n`);

    if (appliedMigrations.length > 0) {
      console.log('Already applied:');
      appliedMigrations.forEach(m => console.log(`   ✓ ${m}`));
      console.log('');
    }

    // Apply pending migrations
    const pendingMigrations = MIGRATION_FILES.filter(m => !appliedMigrations.includes(m));

    if (pendingMigrations.length === 0) {
      console.log('✅ All migrations are already applied!\n');
      await client.end();
      return;
    }

    console.log(`📋 Migrations to apply: ${pendingMigrations.length}\n`);
    console.log('Starting migration process...');

    let successCount = 0;
    let failCount = 0;

    for (const migrationFile of pendingMigrations) {
      const migrationPath = path.join(MIGRATIONS_DIR, migrationFile);
      const success = await applyMigration(client, migrationFile, migrationPath);

      if (success) {
        successCount++;
      } else {
        failCount++;
        console.log('\n⚠️  Migration failed. Do you want to continue? (Ctrl+C to abort)');
        // Give user a moment to see the error
        await new Promise(resolve => setTimeout(resolve, 2000));
      }
    }

    console.log('\n╔════════════════════════════════════════════════════════╗');
    console.log('║                   MIGRATION SUMMARY                    ║');
    console.log('╠════════════════════════════════════════════════════════╣');
    console.log(`║  Total Migrations: ${pendingMigrations.length.toString().padEnd(38)} ║`);
    console.log(`║  Successful:       ${successCount.toString().padEnd(38)} ║`);
    console.log(`║  Failed:           ${failCount.toString().padEnd(38)} ║`);
    console.log('╚════════════════════════════════════════════════════════╝\n');

    if (failCount > 0) {
      console.log('⚠️  Some migrations failed. Check errors above.');
      console.log('💡 Tip: You can apply failed migrations manually via Supabase dashboard.\n');
      await generateSQLFile(MIGRATION_FILES);
    } else {
      console.log('🎉 All migrations applied successfully!\n');
    }

  } catch (error) {
    console.error('\n❌ Error during migration process:', error.message);
  } finally {
    await client.end();
    console.log('🔌 Database connection closed.\n');
  }
}

// Run the script
main().catch(error => {
  console.error('Fatal error:', error);
  process.exit(1);
});
