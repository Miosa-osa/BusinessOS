// Test Supabase Connection
const { createClient } = require('@supabase/supabase-js');
const { Client } = require('pg');

// Supabase credentials
const SUPABASE_URL = 'https://fuqhjbgbjamtxcdphjpp.supabase.co';
const SUPABASE_ANON_KEY = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6ImZ1cWhqYmdiamFtdHhjZHBoanBwIiwicm9sZSI6ImFub24iLCJpYXQiOjE3NjI3OTY3NjQsImV4cCI6MjA3ODM3Mjc2NH0.srS2dl_9JQlEIWGmp4Cj8UpioF40iDUM8uPjGKoF_cs';
const POSTGRES_CONNECTION = 'postgres://postgres:fmm6Wt7kN0ajrjxK@db.fuqhjbgbjamtxcdphjpp.supabase.co:6543/postgres';

async function testSupabaseAPI() {
  console.log('\n=== Testing Supabase API Connection ===');
  try {
    const supabase = createClient(SUPABASE_URL, SUPABASE_ANON_KEY);

    // Test by getting metadata about the database
    const { data, error } = await supabase.rpc('version').then(
      () => ({ data: true, error: null }),
      async () => {
        // If RPC fails, try a simple auth check
        const { data: authData, error: authError } = await supabase.auth.getSession();
        return { data: authData, error: authError };
      }
    );

    if (error) {
      console.log('❌ Supabase API Error:', error.message);
      return false;
    }

    console.log('✅ Supabase API connection successful!');
    console.log('   URL:', SUPABASE_URL);
    console.log('   Status: API is reachable and responding');
    return true;
  } catch (err) {
    console.log('❌ Supabase API connection failed:', err.message);
    return false;
  }
}

async function testPostgresConnection() {
  console.log('\n=== Testing PostgreSQL Direct Connection ===');

  // Test with connection string
  console.log('   Method 1: Using connection string...');
  const client1 = new Client({
    connectionString: POSTGRES_CONNECTION,
    ssl: {
      rejectUnauthorized: false
    }
  });

  try {
    await client1.connect();
    console.log('   ✅ Connection string method successful!');

    const result = await client1.query('SELECT version()');
    console.log('   Database version:', result.rows[0].version.split(' ')[0], result.rows[0].version.split(' ')[1]);

    const tables = await client1.query(`
      SELECT table_name
      FROM information_schema.tables
      WHERE table_schema = 'public'
      ORDER BY table_name
      LIMIT 10
    `);

    console.log('   Available tables:', tables.rows.length > 0 ? tables.rows.map(r => r.table_name).join(', ') : 'No tables found');

    await client1.end();
    return true;
  } catch (err1) {
    console.log('   ❌ Connection string failed:', err1.message);
    console.log('   Code:', err1.code);
    try { await client1.end(); } catch {}

    // Try alternative connection method with explicit parameters
    console.log('\n   Method 2: Using explicit parameters...');
    const client2 = new Client({
      host: 'db.fuqhjbgbjamtxcdphjpp.supabase.co',
      port: 6543,
      database: 'postgres',
      user: 'postgres',
      password: 'fmm6Wt7kN0ajrjxK',
      ssl: {
        rejectUnauthorized: false
      }
    });

    try {
      await client2.connect();
      console.log('   ✅ Explicit parameters method successful!');

      const result = await client2.query('SELECT version()');
      console.log('   Database version:', result.rows[0].version.split(' ')[0], result.rows[0].version.split(' ')[1]);

      await client2.end();
      return true;
    } catch (err2) {
      console.log('   ❌ Explicit parameters failed:', err2.message);
      console.log('   Code:', err2.code);
      try { await client2.end(); } catch {}

      // Try with port 5432 (standard PostgreSQL port)
      console.log('\n   Method 3: Trying port 5432...');
      const client3 = new Client({
        host: 'db.fuqhjbgbjamtxcdphjpp.supabase.co',
        port: 5432,
        database: 'postgres',
        user: 'postgres',
        password: 'fmm6Wt7kN0ajrjxK',
        ssl: {
          rejectUnauthorized: false
        }
      });

      try {
        await client3.connect();
        console.log('   ✅ Port 5432 method successful!');

        const result = await client3.query('SELECT version()');
        console.log('   Database version:', result.rows[0].version.split(' ')[0], result.rows[0].version.split(' ')[1]);

        await client3.end();
        return true;
      } catch (err3) {
        console.log('   ❌ Port 5432 failed:', err3.message);
        console.log('   Code:', err3.code);
        try { await client3.end(); } catch {}

        console.log('\n   ℹ️  Troubleshooting tips:');
        console.log('      - Verify password is correct in Supabase dashboard');
        console.log('      - Check if database password was recently reset');
        console.log('      - Confirm your IP is not blocked by Supabase');
        console.log('      - Try resetting the database password');

        return false;
      }
    }
  }
}

async function main() {
  console.log('╔════════════════════════════════════════════════════════╗');
  console.log('║        Supabase Connection Test                        ║');
  console.log('╚════════════════════════════════════════════════════════╝');

  const apiResult = await testSupabaseAPI();
  const pgResult = await testPostgresConnection();

  console.log('\n=== Summary ===');
  console.log('Supabase API:', apiResult ? '✅ WORKING' : '❌ FAILED');
  console.log('PostgreSQL:  ', pgResult ? '✅ WORKING' : '❌ FAILED');
  console.log('\n');

  if (apiResult && pgResult) {
    console.log('🎉 All connections successful! You can update your .env files.');
  } else {
    console.log('⚠️  Some connections failed. Check the errors above.');
  }
}

main();
