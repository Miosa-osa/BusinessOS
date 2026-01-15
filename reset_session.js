const { Pool } = require('pg');
const p = new Pool({
  connectionString: 'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres',
  ssl: { rejectUnauthorized: false }
});

async function reset() {
  // Delete sessions for test user
  await p.query(`DELETE FROM session WHERE "userId" = 'test-user-001'`);
  console.log('✅ Cleared sessions');
  
  // Delete onboarding sessions
  await p.query(`DELETE FROM onboarding_sessions WHERE user_id = 'test-user-001'`);
  console.log('✅ Cleared onboarding sessions');
  
  await p.end();
  console.log('\n🎉 Session reset! Sign in again to see onboarding.');
}
reset();
