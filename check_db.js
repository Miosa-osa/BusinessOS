const { Client } = require('pg');
async function check() {
  const client = new Client({
    connectionString: 'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres',
    ssl: { rejectUnauthorized: false }
  });
  await client.connect();
  
  console.log("=== ONBOARDING SESSIONS ===");
  const sessions = await client.query('SELECT id, user_id, status, current_step, extracted_data, created_at FROM onboarding_sessions ORDER BY created_at DESC LIMIT 3');
  console.log(JSON.stringify(sessions.rows, null, 2));
  
  console.log("\n=== CONVERSATION HISTORY ===");
  const history = await client.query('SELECT session_id, role, content, question_type, sequence_number FROM onboarding_conversation_history ORDER BY created_at DESC LIMIT 10');
  console.log(JSON.stringify(history.rows, null, 2));
  
  await client.end();
}
check().catch(console.error);