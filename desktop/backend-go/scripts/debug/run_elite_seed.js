/**
 * Elite Seed Data Runner
 *
 * Runs the elite_seed_data.sql with proper variable substitution
 *
 * Usage: node run_elite_seed.js [user_id]
 *
 * If no user_id is provided, it will list available users to choose from.
 */

const { Client } = require("pg");
const fs = require("fs");
const path = require("path");

const DATABASE_URL = process.env.DATABASE_URL;
if (!DATABASE_URL) {
  console.error("DATABASE_URL environment variable is required");
  process.exit(1);
}

async function listUsers(client) {
  const result = await client.query(
    "SELECT id, email FROM auth.users ORDER BY created_at DESC LIMIT 10",
  );
  console.log("\n📋 Available Users:");
  result.rows.forEach((row, i) => {
    console.log(`  ${i + 1}. ${row.id} (${row.email})`);
  });
  return result.rows;
}

async function runSeedScript(client, userId) {
  // Read the seed SQL file
  const seedPath = path.join(__dirname, "elite_seed_data.sql");
  let sql = fs.readFileSync(seedPath, "utf8");

  // Remove psql-specific commands that won't work in pg client
  sql = sql.replace(/\\set user_id_var :user_id/g, "");
  sql = sql.replace(/\\set user_id .*$/gm, "");

  // Replace all :user_id_var and :'user_id_var' with the actual user_id
  sql = sql.replace(/:'user_id_var'/g, `'${userId}'`);
  sql = sql.replace(/:user_id_var/g, `'${userId}'`);

  // Split into individual statements (simple split on semicolons outside strings)
  // For safety, we'll execute the whole thing as a single transaction

  console.log(`\n🚀 Running elite seed data for user: ${userId}\n`);

  try {
    await client.query(sql);
    console.log("✅ Seed data inserted successfully!\n");
  } catch (error) {
    console.error("❌ Error running seed:", error.message);
    if (error.position) {
      const pos = parseInt(error.position);
      const context = sql.substring(Math.max(0, pos - 100), pos + 100);
      console.error("\nContext around error:", context);
    }
    throw error;
  }
}

async function verifySeedData(client, userId) {
  console.log("📊 Verifying seeded data...\n");

  const queries = [
    {
      name: "Projects",
      query: `SELECT COUNT(*) as count FROM projects WHERE user_id = $1`,
    },
    {
      name: "Tasks",
      query: `SELECT COUNT(*) as count FROM tasks WHERE user_id = $1`,
    },
    {
      name: "Clients",
      query: `SELECT COUNT(*) as count FROM clients WHERE user_id = $1`,
    },
    {
      name: "Calendar Events",
      query: `SELECT COUNT(*) as count FROM calendar_events WHERE user_id = $1`,
    },
    {
      name: "Focus Items",
      query: `SELECT COUNT(*) as count FROM focus_items WHERE user_id = $1`,
    },
    {
      name: "Documents",
      query: `SELECT COUNT(*) as count FROM uploaded_documents WHERE user_id = $1`,
    },
    {
      name: "Dashboards",
      query: `SELECT COUNT(*) as count FROM user_dashboards WHERE user_id = $1`,
    },
    {
      name: "Team Members",
      query: `SELECT COUNT(*) as count FROM team_members WHERE user_id = $1`,
    },
  ];

  console.log("Entity               | Count");
  console.log("---------------------|------");

  for (const q of queries) {
    try {
      const result = await client.query(q.query, [userId]);
      console.log(`${q.name.padEnd(20)} | ${result.rows[0].count}`);
    } catch (e) {
      console.log(`${q.name.padEnd(20)} | ERROR: ${e.message}`);
    }
  }

  // Task status breakdown
  console.log("\n📋 Task Status Breakdown:");
  const taskStatus = await client.query(
    `
        SELECT status, COUNT(*) as count,
               COUNT(*) FILTER (WHERE due_date < CURRENT_DATE AND status != 'done') as overdue
        FROM tasks WHERE user_id = $1
        GROUP BY status
    `,
    [userId],
  );

  taskStatus.rows.forEach((row) => {
    console.log(`  ${row.status}: ${row.count} (${row.overdue} overdue)`);
  });
}

async function main() {
  const client = new Client(DATABASE_URL);

  try {
    await client.connect();
    console.log("🔌 Connected to database\n");

    let userId = process.argv[2];

    if (!userId) {
      const users = await listUsers(client);
      // Default to the most likely test user
      userId = users[0]?.id || "86bc90ce-0996-45ee-893e-754bfa198100";
      console.log(`\n📌 Using user: ${userId}`);
      console.log("   (Pass user_id as argument to use a different user)\n");
    }

    await runSeedScript(client, userId);
    await verifySeedData(client, userId);

    console.log("\n🎯 Elite seed data ready for Agent Skill Testing!");
  } catch (error) {
    console.error("Fatal error:", error.message);
    process.exit(1);
  } finally {
    await client.end();
  }
}

main();
