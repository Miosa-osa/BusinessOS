# Communications Module

## Purpose

Communications gives a workspace an operational view of the conversations, messages, and meetings that belong to its work.
It does not treat an external account as a workspace or copy every account record into every workspace.

## Ownership Model

An external account belongs to the authenticated user.
A connector writes its raw source records to its provider table, such as `emails`, `calendar_events`, or `channel_messages`.
Those records remain the immutable provider-facing record.

A workspace owns the operating context around a source record.
That context includes triage state, assignment, waiting status, client or project references, and a deliberate route to the workspace.
The `communication_workspace_items` table is the overlay that stores this context.

One source record has at most one active workspace route for an owner.
Routing the source to another workspace moves the overlay instead of duplicating it.
Removing a route deletes only the overlay and never deletes the underlying connector record.

## Safety Invariants

- Provider data is not deleted when a workspace item is removed.
- A source cannot silently appear in multiple workspaces for the same owner.
- A user must be an active member of a workspace to view or modify its communication overlay.
- Connector data that has not been routed remains visible only in the user's unrouted triage queue.
- Outbound content is a draft until separately approved and sent.
- No connector or agent may send a message merely because an item was routed, classified, or summarized.

## Current Vertical Slice

The Communications Triage console provides three views for the selected workspace.

- Triage shows routed source items that still need action.
- Meetings shows routed calendar source items.
- Unrouted shows account-owned source items that have not yet been assigned to any workspace.

The user can route an unrouted item into the current workspace, change its operational state, or remove its workspace route.
The screen also shows whether Gmail, Slack, Calendar, and WhatsApp connectors are configured for the user.

The API is intentionally workspace-scoped.
It verifies membership before returning data or changing a route.
It also treats a missing provider table as an unavailable connector and returns an empty source list rather than failing the entire communications screen.

## Planned Extensions

- Aggregate individual provider records into conversation threads where the provider supports them.
- Resolve participants into shared people and organization identities.
- Add explicit client, project, and context selection in the detail panel.
- Ingest WhatsApp and meeting-transcript sources through the same immutable-source model.
- Produce task, claim, decision, and memory proposals for Optimal Engine review without automatically accepting them as truth.
- Add a reviewed draft workflow with an explicit send action and audit trail.

## Verification Expectations

Before a release, validate that routing one source to a workspace does not create a workspace or alter another workspace's data.
Validate that moving and removing a route preserve the raw provider record.
Validate that outbound drafts cannot be dispatched without explicit approval.
