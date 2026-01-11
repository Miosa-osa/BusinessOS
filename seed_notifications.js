// Notification Seed Script for Business OS
// Run with: node seed_notifications.js [seed|clear]

const { Client } = require('pg');

const DATABASE_URL = process.env.DATABASE_URL || 
  'postgresql://postgres:Lunivate69420@db.fuqhjbgbjamtxcdphjpp.supabase.co:5432/postgres';

const action = process.argv[2] || 'seed';
const targetEmail = process.argv[3] || null; // Optional: specify email

async function seed(client) {
  // Get user - either by email or first user
  let userResult;
  if (targetEmail) {
    userResult = await client.query('SELECT id FROM "user" WHERE email = $1', [targetEmail]);
  } else {
    userResult = await client.query('SELECT id FROM "user" ORDER BY "createdAt" DESC LIMIT 1');
  }
  if (userResult.rows.length === 0) {
    console.error('❌ No users found. Specify email: node seed_notifications.js seed your@email.com');
    return;
  }
  const userId = userResult.rows[0].id;
  console.log(`🌱 Seeding notifications for user: ${userId}`);

  const now = new Date();
  const notifications = [
    // TODAY - URGENT/HIGH
    { type: 'task.overdue', title: 'Task overdue: Q4 Financial Report', body: 'This task was due yesterday and needs immediate attention.', entity_type: 'task', sender_name: 'Sarah Chen', priority: 'urgent', is_read: false, offset: -10 },
    { type: 'task.due_today', title: 'Task due today: Review PR #142', body: 'Code review for the notifications feature is due today.', entity_type: 'task', sender_name: 'Mike Johnson', priority: 'high', is_read: false, offset: -30 },
    { type: 'integration.sync_failed', title: 'Google Calendar sync failed', body: 'Unable to sync calendar events. Please reconnect your account.', entity_type: 'integration', sender_name: 'System', priority: 'high', is_read: false, offset: -60 },
    
    // TODAY - NORMAL
    { type: 'task.assigned', title: 'You were assigned: Update API documentation', body: 'Sarah Chen assigned you to this task in Project Alpha.', entity_type: 'task', sender_name: 'Sarah Chen', priority: 'normal', is_read: false, offset: -120 },
    { type: 'mention.comment', title: '@you in: Design Review Discussion', body: 'Mike Johnson mentioned you: "what do you think about this approach?"', entity_type: 'comment', sender_name: 'Mike Johnson', priority: 'normal', is_read: false, offset: -180 },
    { type: 'project.added', title: 'Added to project: Mobile App Redesign', body: "You've been added as a contributor to this project.", entity_type: 'project', sender_name: 'Emily Davis', priority: 'normal', is_read: false, offset: -240 },
    { type: 'task.comment', title: 'New comment on: Update API documentation', body: 'Emily Davis commented: "I\'ve added some notes to the shared doc."', entity_type: 'task', sender_name: 'Emily Davis', priority: 'normal', is_read: true, offset: -300 },
    { type: 'client.meeting_scheduled', title: 'Meeting scheduled with Acme Corp', body: 'Tomorrow at 2:00 PM - Quarterly business review', entity_type: 'client', sender_name: 'Alex Thompson', priority: 'normal', is_read: false, offset: -360 },
    
    // YESTERDAY
    { type: 'task.completed', title: 'Task completed: Setup CI/CD pipeline', body: 'Alex Thompson marked this task as complete.', entity_type: 'task', sender_name: 'Alex Thompson', priority: 'normal', is_read: true, offset: -1200 },
    { type: 'team.member_joined', title: 'New team member: Jordan Park', body: 'Jordan Park joined the Engineering team.', entity_type: 'team', sender_name: 'System', priority: 'low', is_read: true, offset: -1320 },
    { type: 'mention.task', title: '@you in task: Database Migration Plan', body: 'Sarah Chen mentioned you in a task description.', entity_type: 'task', sender_name: 'Sarah Chen', priority: 'normal', is_read: false, offset: -1560 },
    { type: 'project.status_changed', title: 'Project status: Mobile App Redesign → In Progress', body: 'Project moved from Planning to In Progress.', entity_type: 'project', sender_name: 'Emily Davis', priority: 'normal', is_read: true, offset: -1680 },
    
    // THIS WEEK
    { type: 'task.due_soon', title: 'Task due soon: Prepare presentation slides', body: 'This task is due in 3 days.', entity_type: 'task', sender_name: 'System', priority: 'high', is_read: false, offset: -4320 },
    { type: 'client.deal_update', title: 'Deal update: Acme Corp - Enterprise Plan', body: 'Deal value updated to $50,000. Stage: Negotiation.', entity_type: 'client', sender_name: 'Mike Johnson', priority: 'normal', is_read: true, offset: -5760 },
    { type: 'integration.connected', title: 'Slack connected successfully', body: 'Your Slack workspace is now connected.', entity_type: 'integration', sender_name: 'System', priority: 'low', is_read: true, offset: -6000 },
    { type: 'project.completed', title: 'Project completed: Website Refresh', body: 'Congratulations! The Website Refresh project is now complete.', entity_type: 'project', sender_name: 'Emily Davis', priority: 'normal', is_read: true, offset: -7200 },
    { type: 'chat.artifact_ready', title: 'AI artifact ready: Market Analysis Report', body: 'Your requested analysis is ready to view.', entity_type: 'chat', sender_name: 'AI Assistant', priority: 'normal', is_read: false, offset: -7800 },
    
    // EARLIER
    { type: 'system.welcome', title: 'Welcome to Business OS!', body: 'Get started by creating your first project or exploring the dashboard.', entity_type: 'system', sender_name: 'System', priority: 'low', is_read: true, offset: -14400 },
    { type: 'system.feature_announcement', title: 'New feature: AI-powered task suggestions', body: 'Try our new AI feature that suggests task breakdowns automatically.', entity_type: 'system', sender_name: 'System', priority: 'low', is_read: true, offset: -20160 },
    { type: 'dailylog.reminder', title: "Don't forget your daily log", body: 'Take a moment to log your progress for today.', entity_type: 'dailylog', sender_name: 'System', priority: 'normal', is_read: true, offset: -21600 },
    { type: 'team.role_changed', title: 'Your role updated: Project Lead', body: "You've been promoted to Project Lead for the Engineering team.", entity_type: 'team', sender_name: 'Sarah Chen', priority: 'normal', is_read: true, offset: -24000 },
  ];

  let count = 0;
  for (const n of notifications) {
    const createdAt = new Date(now.getTime() + n.offset * 60 * 1000);
    try {
      await client.query(`
        INSERT INTO notifications (user_id, type, title, body, entity_type, sender_name, priority, is_read, metadata, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
      `, [userId, n.type, n.title, n.body, n.entity_type, n.sender_name, n.priority, n.is_read, JSON.stringify({ seeded: true }), createdAt]);
      count++;
    } catch (err) {
      console.error(`Failed to insert ${n.type}:`, err.message);
    }
  }
  
  console.log(`✅ Created ${count} notifications`);
}

async function clear(client) {
  const result = await client.query("DELETE FROM notifications WHERE metadata->>'seeded' = 'true'");
  console.log(`🗑️  Deleted ${result.rowCount} seeded notifications`);
}

async function main() {
  const client = new Client({ connectionString: DATABASE_URL, ssl: { rejectUnauthorized: false } });
  
  try {
    await client.connect();
    console.log('📡 Connected to database');
    
    if (action === 'clear') {
      await clear(client);
    } else {
      await seed(client);
    }
    
    // Show counts
    const countResult = await client.query('SELECT COUNT(*) as total, COUNT(*) FILTER (WHERE NOT is_read) as unread FROM notifications');
    console.log(`📊 Total: ${countResult.rows[0].total}, Unread: ${countResult.rows[0].unread}`);
    
  } catch (err) {
    console.error('❌ Error:', err.message);
  } finally {
    await client.end();
  }
}

main();
