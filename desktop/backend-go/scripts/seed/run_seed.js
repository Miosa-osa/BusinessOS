// Run seed data - execute with: node run_seed.js
const { Client } = require("pg");
const fs = require("fs");
const path = require("path");

async function runSeed() {
  if (!process.env.DATABASE_URL) {
    console.error("DATABASE_URL required");
    process.exit(1);
  }
  const client = new Client({
    connectionString: process.env.DATABASE_URL,
  });

  try {
    await client.connect();
    console.log("Connected to database");

    const seedSQL = fs.readFileSync(
      path.join(__dirname, "seed_test_data.sql"),
      "utf8",
    );
    await client.query(seedSQL);

    console.log("✓ Seed data inserted successfully");

    // Verify counts
    const counts = await client.query(`
      SELECT
        (SELECT COUNT(*) FROM team_members WHERE user_id = 'test-user-seed-001') as team_members,
        (SELECT COUNT(*) FROM clients WHERE user_id = 'test-user-seed-001') as clients,
        (SELECT COUNT(*) FROM projects WHERE user_id = 'test-user-seed-001') as projects,
        (SELECT COUNT(*) FROM tasks WHERE user_id = 'test-user-seed-001') as tasks,
        (SELECT COUNT(*) FROM calendar_events WHERE user_id = 'test-user-seed-001') as events,
        (SELECT COUNT(*) FROM emails WHERE user_id = 'test-user-seed-001') as emails
    `);

    console.log("\nCreated:");
    console.log(`  Team Members: ${counts.rows[0].team_members}`);
    console.log(`  Clients: ${counts.rows[0].clients}`);
    console.log(`  Projects: ${counts.rows[0].projects}`);
    console.log(`  Tasks: ${counts.rows[0].tasks}`);
    console.log(`  Calendar Events: ${counts.rows[0].events}`);
    console.log(`  Emails: ${counts.rows[0].emails}`);
  } catch (err) {
    console.error("Error:", err.message);
    process.exit(1);
  } finally {
    await client.end();
  }
}

runSeed();
