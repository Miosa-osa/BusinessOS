const { Pool } = require('pg');
const p = new Pool({
  connectionString: 'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres',
  ssl: { rejectUnauthorized: false }
});

async function check() {
  const e1 = await p.query(`SELECT enumlabel FROM pg_enum WHERE enumtypid = 'clienttype'::regtype`);
  console.log('clienttype enum:', e1.rows.map(x => x.enumlabel).join(', '));
  
  const e2 = await p.query(`SELECT enumlabel FROM pg_enum WHERE enumtypid = 'clientstatus'::regtype`);
  console.log('clientstatus enum:', e2.rows.map(x => x.enumlabel).join(', '));
  
  await p.end();
}
check();
