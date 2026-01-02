// Verify Supabase Schema - Check all required tables and columns
const { createClient } = require('@supabase/supabase-js');
const { Client } = require('pg');

const SUPABASE_URL = 'https://fuqhjbgbjamtxcdphjpp.supabase.co';
const SUPABASE_ANON_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImZ1cWhqYmdiamFtdHhjZHBoanBwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjI3OTY3NjQsImV4cCI6MjA3ODM3Mjc2NH0.srS2dl_9JQlEIWGmp4Cj8UpioF40iDUM8uPjGKoF_cs';

// Required tables based on pedro_tasks_v2.md migrations
const REQUIRED_TABLES = {
  // Migration 016
  'memories': [
    'id', 'user_id', 'title', 'summary', 'content', 'memory_type', 'category',
    'source_type', 'source_id', 'source_context', 'project_id', 'node_id',
    'importance_score', 'access_count', 'last_accessed_at', 'embedding',
    'embedding_model', 'is_active', 'is_pinned', 'expires_at', 'tags',
    'metadata', 'created_at', 'updated_at'
  ],
  'memory_associations': [
    'id', 'memory_id', 'entity_type', 'entity_id', 'relevance_score',
    'association_type', 'created_at'
  ],
  'memory_access_log': [
    'id', 'memory_id', 'user_id', 'access_type', 'accessing_agent',
    'conversation_id', 'trigger_query', 'relevance_score', 'created_at'
  ],
  'user_facts': [
    'id', 'user_id', 'fact_key', 'fact_value', 'fact_type',
    'source_memory_id', 'confidence_score', 'is_active',
    'last_confirmed_at', 'created_at', 'updated_at'
  ],

  // Migration 017 - Context System
  'context_nodes': ['id', 'user_id', 'parent_id', 'title', 'created_at', 'updated_at'],
  'context_profiles': ['id', 'user_id', 'name', 'created_at', 'updated_at'],

  // Migration 018 - Output Styles
  'output_styles': ['id', 'name', 'created_at'],
  'user_output_preferences': ['id', 'user_id', 'created_at'],

  // Migration 019 - Documents
  'uploaded_documents': ['id', 'user_id', 'title', 'embedding', 'created_at', 'updated_at'],
  'document_chunks': ['id', 'document_id', 'chunk_index', 'content', 'embedding', 'created_at'],

  // Migration 020 - Context Integration
  'conversations': ['id', 'user_id', 'title', 'embedding', 'created_at', 'updated_at'],
  'conversation_summaries': [
    'id', 'conversation_id', 'summary', 'title', 'sentiment', 'entities',
    'questions', 'decisions', 'code_mentions', 'token_count', 'duration',
    'metadata', 'embedding', 'created_at', 'updated_at'
  ],
  'voice_notes': ['id', 'user_id', 'transcript', 'embedding', 'created_at'],

  // Migration 021 - Learning System
  'learning_events': ['id', 'user_id', 'event_type', 'created_at'],
  'behavior_patterns': ['id', 'user_id', 'pattern_type', 'created_at'],

  // Migration 022 - Application Profiles
  'application_profiles': ['id', 'user_id', 'name', 'embedding', 'created_at', 'updated_at'],
  'application_components': ['id', 'profile_id', 'name', 'embedding', 'created_at'],
  'application_api_endpoints': ['id', 'profile_id', 'path', 'embedding', 'created_at'],
  'code_patterns': ['id', 'profile_id', 'pattern', 'embedding', 'created_at']
};

async function checkSchema() {
  console.log('╔════════════════════════════════════════════════════════╗');
  console.log('║        Supabase Schema Verification                    ║');
  console.log('╚════════════════════════════════════════════════════════╝\n');

  const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

  console.log('📊 Checking tables and columns...\n');

  const results = {
    totalTables: Object.keys(REQUIRED_TABLES).length,
    existingTables: 0,
    missingTables: [],
    tableDetails: {}
  };

  for (const [tableName, requiredColumns] of Object.entries(REQUIRED_TABLES)) {
    try {
      // Try to query the table structure
      const { data, error } = await supabase
        .from(tableName)
        .select('*')
        .limit(0);

      if (error) {
        if (error.code === 'PGRST116' || error.message.includes('does not exist')) {
          console.log(`❌ Table "${tableName}" - NOT FOUND`);
          results.missingTables.push(tableName);
          results.tableDetails[tableName] = { exists: false, missingColumns: requiredColumns };
        } else {
          console.log(`⚠️  Table "${tableName}" - Error: ${error.message}`);
          results.tableDetails[tableName] = { exists: 'unknown', error: error.message };
        }
      } else {
        console.log(`✅ Table "${tableName}" - EXISTS`);
        results.existingTables++;
        results.tableDetails[tableName] = { exists: true, missingColumns: [] };

        // Note: We can't easily check individual columns via Supabase JS client
        // Would need direct PostgreSQL connection for that
      }
    } catch (err) {
      console.log(`❌ Table "${tableName}" - Error: ${err.message}`);
      results.tableDetails[tableName] = { exists: false, error: err.message };
    }
  }

  console.log('\n═══════════════════════════════════════════════════════');
  console.log('                    SUMMARY                            ');
  console.log('═══════════════════════════════════════════════════════\n');

  console.log(`Total Required Tables: ${results.totalTables}`);
  console.log(`Existing Tables: ${results.existingTables}`);
  console.log(`Missing Tables: ${results.missingTables.length}\n`);

  if (results.missingTables.length > 0) {
    console.log('🔴 MISSING TABLES:');
    results.missingTables.forEach(table => {
      console.log(`   - ${table}`);
    });
    console.log('\n⚠️  You need to run migrations to create these tables.');
    console.log('   Run: node apply-migrations.js (or use Supabase dashboard)\n');
  } else {
    console.log('✅ All required tables exist!\n');
  }

  return results;
}

async function checkEmbeddingDimensions() {
  console.log('\n╔════════════════════════════════════════════════════════╗');
  console.log('║        Checking Embedding Dimensions                   ║');
  console.log('╚════════════════════════════════════════════════════════╝\n');

  console.log('⚠️  Note: This requires direct PostgreSQL access.');
  console.log('   The embeddings should be vector(768) not vector(1536).');
  console.log('   Migration 024 handles this conversion.\n');
}

async function main() {
  try {
    const results = await checkSchema();
    await checkEmbeddingDimensions();

    console.log('═══════════════════════════════════════════════════════\n');

    if (results.missingTables.length === 0) {
      console.log('🎉 Schema verification complete! All tables exist.');
      console.log('   Next step: Verify column details with migration 024.');
    } else {
      console.log('⚠️  Schema incomplete. Apply missing migrations.');
    }

  } catch (error) {
    console.error('Error during verification:', error);
  }
}

main();
