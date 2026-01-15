// Complete test data seed for BusinessOS
const { Pool } = require('pg');

const pool = new Pool({
  connectionString: 'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres',
  ssl: { rejectUnauthorized: false }
});

const USER_ID = 'test-user-001';
const WORKSPACE_ID = '9ed0dd7f-c72e-4c01-bb48-46cf9b3e90c1';

async function seed() {
  const client = await pool.connect();
  
  try {
    console.log('🌱 Seeding test data...\n');

    // 1. WORKSPACE MEMBER
    console.log('👤 Seeding workspace member...');
    await client.query(`
      INSERT INTO workspace_members (id, workspace_id, user_id, role, status, joined_at, created_at, updated_at)
      VALUES (gen_random_uuid(), $1, $2, 'owner', 'active', NOW(), NOW(), NOW())
      ON CONFLICT (workspace_id, user_id) DO UPDATE SET role = 'owner', status = 'active'
    `, [WORKSPACE_ID, USER_ID]);
    console.log('  ✅ Test user is workspace owner\n');

    // 2. PROJECTS
    console.log('📁 Seeding projects...');
    await client.query(`DELETE FROM projects WHERE user_id = $1`, [USER_ID]);
    
    const projects = [
      { name: 'Website Redesign', status: 'ACTIVE', priority: 'HIGH' },
      { name: 'Mobile App MVP', status: 'ACTIVE', priority: 'MEDIUM' },
      { name: 'Q1 Marketing', status: 'ACTIVE', priority: 'LOW' },
    ];
    
    const projectIds = [];
    for (const p of projects) {
      const res = await client.query(`
        INSERT INTO projects (id, name, status, priority, user_id, owner_id, workspace_id, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $4, $5, NOW(), NOW())
        RETURNING id
      `, [p.name, p.status, p.priority, USER_ID, WORKSPACE_ID]);
      projectIds.push(res.rows[0].id);
      console.log(`  ✅ ${p.name}`);
    }
    console.log('');

    // 3. TASKS
    console.log('✅ Seeding tasks...');
    await client.query(`DELETE FROM tasks WHERE user_id = $1`, [USER_ID]);
    
    const tasks = [
      { title: 'Design homepage mockup', status: 'in_progress', priority: 'high', projectIdx: 0 },
      { title: 'Set up CI/CD pipeline', status: 'todo', priority: 'medium', projectIdx: 1 },
      { title: 'Write API documentation', status: 'todo', priority: 'low', projectIdx: 1 },
      { title: 'Review pull requests', status: 'in_progress', priority: 'high', projectIdx: 0 },
      { title: 'Update dependencies', status: 'done', priority: 'low', projectIdx: 2 },
    ];
    
    for (const t of tasks) {
      await client.query(`
        INSERT INTO tasks (id, title, status, priority, user_id, project_id, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
      `, [t.title, t.status, t.priority, USER_ID, projectIds[t.projectIdx]]);
      console.log(`  ✅ ${t.title}`);
    }
    console.log('');

    // 4. CLIENTS
    console.log('🏢 Seeding clients...');
    await client.query(`DELETE FROM clients WHERE user_id = $1`, [USER_ID]);
    
    const clients = [
      { name: 'Acme Corp', email: 'contact@acme.com', status: 'active', type: 'company' },
      { name: 'TechStart Inc', email: 'hello@techstart.io', status: 'active', type: 'company' },
      { name: 'John Smith', email: 'john@freelance.com', status: 'lead', type: 'individual' },
    ];
    
    for (const c of clients) {
      await client.query(`
        INSERT INTO clients (id, name, email, status, type, user_id, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
      `, [c.name, c.email, c.status, c.type, USER_ID]);
      console.log(`  ✅ ${c.name}`);
    }
    console.log('');

    // 5. NOTIFICATIONS
    console.log('🔔 Seeding notifications...');
    await client.query(`DELETE FROM notifications WHERE user_id = $1`, [USER_ID]);
    
    const notifications = [
      { title: 'Welcome to BusinessOS!', body: 'Get started by exploring your dashboard.', type: 'system' },
      { title: 'Task assigned', body: 'You have been assigned "Design homepage mockup"', type: 'task_assigned' },
      { title: 'Project update', body: 'Website Redesign project is now active', type: 'project_update' },
    ];
    
    for (const n of notifications) {
      await client.query(`
        INSERT INTO notifications (id, user_id, workspace_id, title, body, type, is_read, created_at, updated_at)
        VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, false, NOW(), NOW())
      `, [USER_ID, WORKSPACE_ID, n.title, n.body, n.type]);
      console.log(`  ✅ ${n.title}`);
    }
    console.log('');

    console.log('🎉 All test data seeded successfully!\n');
    console.log('='.repeat(50));
    console.log('TEST CREDENTIALS:');
    console.log('='.repeat(50));
    console.log('Email:        testuser@businessos.dev');
    console.log('Password:     password123');
    console.log('Workspace:    Test Workspace');
    console.log('='.repeat(50));

  } catch (err) {
    console.error('❌ Error:', err.message);
    console.error(err);
  } finally {
    client.release();
    await pool.end();
  }
}

seed();
