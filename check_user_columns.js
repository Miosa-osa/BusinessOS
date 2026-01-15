const { Pool } = require('pg');

const pool = new Pool({
  connectionString: 'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres'
});

async function check() {
  // Check onboarding tables
  const tables = await pool.query(`SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE '%onboard%'`);
  console.log('onboarding tables:', tables.rows);
  
  // Check onboarding_sessions columns
  const cols = await pool.query(`SELECT column_name FROM information_schema.columns WHERE table_name = 'onboarding_sessions'`);
  console.log('onboarding_sessions columns:', cols.rows.map(x => x.column_name));
  
  // Check for any completed status
  const sessions = await pool.query(`SELECT * FROM onboarding_sessions WHERE user_id = 'test-user-001'`);
  console.log('onboarding sessions for test user:', sessions.rows);
  
  await pool.end();
}

check().catch(console.error);
