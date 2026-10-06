-- Complete the intentionally fictional Northstar Growth demo with an operating
-- cadence and a meeting follow-up agent. Keep this migration idempotent so it
-- can run safely across environments that already contain part of the demo.

WITH northstar AS (
    SELECT id, owner_id
    FROM workspaces
    WHERE slug = 'northstar-growth'
    LIMIT 1
), entries(period, kind, content, owner, position) AS (
    VALUES
        ('daily', 'focus', 'Complete the Atlas Roofing access map and confirm system owners before build work begins.', 'Noah Kim', 10),
        ('daily', 'blocker', 'Atlas Roofing CRM export is still awaiting client approval.', 'Avery Quinn', 20),
        ('daily', 'priority', 'Turn the latest discovery call into reviewed tasks, owners, and due dates.', 'Elena Park', 30),
        ('daily', 'note', 'No agent may launch client-facing work without a human approval receipt.', 'Talia Morgan', 40),
        ('weekly', 'focus', 'Move qualified opportunities through the Solutions Pipeline with evidence for every next step.', 'Marcus Reed', 10),
        ('weekly', 'priority', 'Publish the September VSL and route responses into the Northstar qualification workflow.', 'Maya Chen', 20),
        ('monthly', 'focus', 'Improve delivery capacity without adding disconnected tools or ungoverned automation.', 'Noah Kim', 10)
)
INSERT INTO rhythm_entries (
    workspace_id, period, kind, content, owner, position, created_by
)
SELECT n.id, e.period, e.kind, e.content, e.owner, e.position, n.owner_id
FROM northstar n
CROSS JOIN entries e
WHERE NOT EXISTS (
    SELECT 1
    FROM rhythm_entries existing
    WHERE existing.workspace_id = n.id
      AND existing.period = e.period
      AND existing.kind = e.kind
      AND existing.content = e.content
);

INSERT INTO workspace_agents (
    workspace_id,
    name,
    role,
    description,
    model,
    system_prompt,
    status,
    created_by
)
SELECT
    w.id,
    'After-Meeting Analyzer',
    'Operations',
    'Turns completed meeting transcripts into proposed workspace tasks with owners, dates, and source evidence for human review.',
    'claude-sonnet-4-5-20250929',
    'You analyze completed meeting transcripts for Northstar Growth. Return proposed tasks with an owner, due date, source evidence, and confidence. Keep every task within the selected workspace. Never invent commitments. Flag missing owners and dates. Do not write, dispatch, or execute actions until a human approves them.',
    'active',
    w.owner_id
FROM workspaces w
WHERE w.slug = 'northstar-growth'
  AND NOT EXISTS (
      SELECT 1
      FROM workspace_agents existing
      WHERE existing.workspace_id = w.id
        AND lower(existing.name) = lower('After-Meeting Analyzer')
  );
