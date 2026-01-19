# Javaris UI Tasks

> **Owner:** Javaris
> **Created:** January 8, 2026
> **Status:** In Progress

---

## Overview

Two main features to implement:
1. **Notifications UI** (CUS-38, CUS-39) - Real-time notification system
2. **Team Invitations UI** (CUS-29) - Workspace invitation flow

---

## 1. Notifications UI

### 1.1 Core Components

- [ ] **NotificationBell.svelte** - Bell icon with unread badge for header
  - Display unread count badge
  - Animate on new notification
  - Click to open dropdown

- [ ] **NotificationDropdown.svelte** - Popover showing recent notifications
  - List recent 10 notifications
  - "Mark all as read" button
  - "View all" link to full page
  - Empty state

- [ ] **NotificationItem.svelte** - Single notification row
  - Icon based on type (task, project, comment, workspace)
  - Title, body, timestamp
  - Sender avatar (if applicable)
  - Unread indicator
  - Click to navigate + mark read

- [ ] **NotificationList.svelte** - Full notifications page/panel
  - Infinite scroll or pagination
  - Filter by type (all, unread, mentions)
  - Bulk actions (mark read, delete)

### 1.2 Layout Integration

- [ ] Add NotificationBell to app layout header (between nav and user profile)
- [ ] Initialize SSE connection on app mount
- [ ] Initialize Web Push on app mount
- [ ] Add toast notifications for real-time events

### 1.3 Settings Page Enhancement

- [ ] Expand Notifications tab with full preferences:
  - [ ] Email notifications toggle
  - [ ] Push notifications toggle + test button
  - [ ] In-app notifications toggle
  - [ ] Quiet hours (enable, start time, end time)
  - [ ] Per-type notification settings (optional)

### 1.4 Polish & UX

- [ ] Loading states (skeleton)
- [ ] Error states
- [ ] Empty states
- [ ] Dark mode support
- [ ] Keyboard navigation
- [ ] Sound toggle for notifications

---

## 2. Team Invitations UI

### 2.1 Core Components

- [ ] **InviteAcceptPage** - `/invite/[token]/+page.svelte`
  - Accept invitation via token
  - Show workspace name on success
  - Handle expired/invalid tokens
  - Redirect to workspace after acceptance

### 2.2 Team Page Enhancements

- [ ] Add "Invite Member" button to team page header
- [ ] Show pending invitations section on team page
- [ ] Connect existing `WorkspaceInvitesList` to team page
- [ ] Add toast on successful invite send

### 2.3 Integration

- [ ] Link team invites to workspace settings
- [ ] Add invite count badge (optional)
- [ ] Email preview for invitation (optional)

---

## File Locations

| Component | Path |
|-----------|------|
| NotificationBell | `src/lib/components/notifications/NotificationBell.svelte` |
| NotificationDropdown | `src/lib/components/notifications/NotificationDropdown.svelte` |
| NotificationItem | `src/lib/components/notifications/NotificationItem.svelte` |
| NotificationList | `src/lib/components/notifications/NotificationList.svelte` |
| Invite Accept Page | `src/routes/invite/[token]/+page.svelte` |

---

## Dependencies (Already Exist)

- ✅ `src/lib/stores/notifications.ts` - Full notification store with SSE
- ✅ `src/lib/services/pushService.ts` - Web Push service
- ✅ `src/lib/api/workspaces/workspaces.ts` - Invite API functions
- ✅ `src/lib/components/workspace/InviteMemberModal.svelte` - Invite modal
- ✅ `src/lib/components/workspace/WorkspaceInvitesList.svelte` - Invites list

---

## Design Tokens

```css
/* Notification-specific colors */
--notification-unread-bg: rgba(59, 130, 246, 0.1);
--notification-hover-bg: var(--color-bg-secondary);
--notification-badge-bg: #dc2626;
--notification-badge-text: white;
```

---

## Priority Order

1. 🔴 NotificationBell + Dropdown (highest visibility)
2. 🔴 Invite Accept Page (blocks invite flow)
3. 🟡 NotificationItem component
4. 🟡 Settings preferences expansion
5. 🟢 Full NotificationList page
6. 🟢 Team page integration

---

## Notes

- Use existing `bits-ui` Popover for dropdown
- Match existing design patterns (rounded-xl, dark mode overrides)
- Follow Svelte 5 runes syntax (`$state`, `$derived`, `$effect`)
- Use lucide-svelte icons consistently
