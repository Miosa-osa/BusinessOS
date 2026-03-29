/**
 * Verify Elite Seed Data
 */

const { Client } = require("pg");

const DATABASE_URL = process.env.DATABASE_URL;
if (!DATABASE_URL) {
  console.error("DATABASE_URL environment variable is required");
  process.exit(1);
}

async function verify() {
  const client = new Client(DATABASE_URL);
  await client.connect();
  const userId = "86bc90ce-0996-45ee-893e-754bfa198100";

  console.log("\n📊 DASHBOARDS:");
  const dashboards = await client.query(
    "SELECT name, is_default FROM user_dashboards WHERE user_id = $1",
    [userId],
  );
  dashboards.rows.forEach((r) =>
    console.log("  -", r.name, r.is_default ? "(default)" : ""),
  );

  console.log("\n📁 PROJECTS:");
  const projects = await client.query(
    "SELECT name, status, priority FROM projects WHERE user_id = $1 ORDER BY created_at",
    [userId],
  );
  projects.rows.forEach((r) =>
    console.log("  -", r.name, "|", r.status, "|", r.priority),
  );

  console.log("\n📅 CALENDAR EVENTS (next 7 days):");
  const events = await client.query(
    "SELECT title, start_time FROM calendar_events WHERE user_id = $1 AND start_time >= CURRENT_DATE AND start_time < CURRENT_DATE + INTERVAL '7 days' ORDER BY start_time",
    [userId],
  );
  events.rows.forEach((r) => {
    const date = r.start_time.toISOString().split("T")[0];
    const time = r.start_time.toISOString().split("T")[1].substring(0, 5);
    console.log("  -", r.title, "|", date, time);
  });

  console.log("\n🏢 CLIENTS:");
  const clients = await client.query(
    "SELECT name, status FROM clients WHERE user_id = $1",
    [userId],
  );
  clients.rows.forEach((r) => console.log("  -", r.name, "|", r.status));

  console.log("\n🎯 FOCUS ITEMS:");
  const focusItems = await client.query(
    "SELECT text, completed FROM focus_items WHERE user_id = $1",
    [userId],
  );
  focusItems.rows.forEach((r) =>
    console.log("  -", r.text, "|", r.completed ? "✓" : "○"),
  );

  console.log("\n📄 DOCUMENTS:");
  const docs = await client.query(
    "SELECT display_name, document_type FROM uploaded_documents WHERE user_id = $1",
    [userId],
  );
  docs.rows.forEach((r) =>
    console.log("  -", r.display_name, "|", r.document_type),
  );

  console.log("\n👥 TEAM MEMBERS:");
  const team = await client.query(
    "SELECT name, role FROM team_members WHERE user_id = $1",
    [userId],
  );
  team.rows.forEach((r) => console.log("  -", r.name, "|", r.role));

  await client.end();
  console.log("\n✅ Verification complete!\n");
}

verify().catch((e) => {
  console.error("Error:", e.message);
  process.exit(1);
});
