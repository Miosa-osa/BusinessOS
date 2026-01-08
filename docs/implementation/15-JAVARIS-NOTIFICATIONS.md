# Javaris: Frontend Notifications Guide

> **Status:** Ready for Implementation (Backend Done)
> **Owner:** Javaris
> **Linear Issues:** CUS-38 (Done), CUS-39 (Done)
> **Last Updated:** January 8, 2026

---

## Overview

The notifications system provides three delivery channels:
- **SSE (Server-Sent Events)** - Real-time in-app notifications
- **Web Push** - Background notifications (even when app is closed)
- **Email** - Important notifications delivered via email

---

## Quick Start

### 1. Initialize Notifications

```typescript
// In your app's root layout or main component
import { notificationStore } from '$lib/stores/notifications';
import { initializePush } from '$lib/services/pushService';
import { onMount } from 'svelte';

onMount(() => {
  // Connect to SSE stream for real-time notifications
  notificationStore.initialize();

  // Initialize Web Push (optional)
  initializePush();
});
```

### 2. Display Notification Count (Badge)

```svelte
<script>
  import { notificationStore } from '$lib/stores/notifications';
  const { unreadCount, hasUnread } = notificationStore;
</script>

<button class="notification-bell">
  <BellIcon />
  {#if $hasUnread}
    <span class="badge">{$unreadCount}</span>
  {/if}
</button>
```

### 3. List Notifications

```svelte
<script>
  import { notificationStore } from '$lib/stores/notifications';
  const { notifications, recentNotifications } = notificationStore;
</script>

<!-- Recent notifications (last 10) -->
{#each $recentNotifications as notification}
  <NotificationItem {notification} />
{/each}

<!-- Or all notifications -->
{#each $notifications as notification}
  <NotificationItem {notification} />
{/each}
```

---

## Notification Store API

### Stores (Reactive)

| Store | Type | Description |
|-------|------|-------------|
| `notifications` | `Notification[]` | All loaded notifications |
| `unreadCount` | `number` | Count of unread notifications |
| `isConnected` | `boolean` | SSE connection status |
| `connectionError` | `string \| null` | Connection error message |
| `hasUnread` | `boolean` | Derived: `unreadCount > 0` |
| `recentNotifications` | `Notification[]` | Derived: last 10 notifications |

### Methods

```typescript
// Fetch notifications from server
await notificationStore.fetchNotifications(limit?: number, offset?: number);

// Fetch just the unread count (for badge)
await notificationStore.fetchUnreadCount();

// Mark single notification as read
await notificationStore.markAsRead(notificationId: string);

// Mark multiple as read
await notificationStore.markMultipleAsRead(ids: string[]);

// Mark all as read
await notificationStore.markAllAsRead();

// Delete a notification
await notificationStore.deleteNotification(id: string);

// Get user's notification preferences
const prefs = await notificationStore.getPreferences();

// Update preferences
await notificationStore.updatePreferences({
  email_enabled: true,
  push_enabled: true,
  quiet_hours_enabled: true,
  quiet_hours_start: "22:00",
  quiet_hours_end: "08:00"
});

// Connection management
notificationStore.connect();    // Start SSE connection
notificationStore.disconnect(); // Stop SSE connection
```

---

## Notification Type

```typescript
interface Notification {
  id: string;
  user_id: string;
  workspace_id?: string;
  type: string;                    // e.g., "task.assigned", "comment.mention"
  title: string;                   // Display title
  body?: string;                   // Optional body text
  entity_type?: string;            // "task", "project", "comment", etc.
  entity_id?: string;              // UUID of related entity
  sender_id?: string;              // Who triggered this
  sender_name?: string;
  sender_avatar_url?: string;
  is_read: boolean;
  read_at?: string;
  priority: 'low' | 'normal' | 'high' | 'urgent';
  metadata?: Record<string, unknown>;
  created_at: string;
}
```

---

## Notification Types (35 Total)

### Task Notifications
| Type | Title Example | When Triggered |
|------|---------------|----------------|
| `task.assigned` | "You were assigned: Review proposal" | User assigned to task |
| `task.completed` | "Task completed: Review proposal" | Task marked done |
| `task.due_soon` | "Task due soon: Review proposal" | 24h before due date |
| `task.overdue` | "Task overdue: Review proposal" | Past due date |
| `task.comment` | "New comment on: Review proposal" | Comment added |
| `task.status_changed` | "Status changed: Review proposal" | Status updated |

### Comment Notifications
| Type | Title Example | When Triggered |
|------|---------------|----------------|
| `comment.mention` | "@you in: Review proposal" | User @mentioned |
| `comment.reply` | "Reply to your comment" | Reply to user's comment |

### Project Notifications
| Type | Title Example | When Triggered |
|------|---------------|----------------|
| `project.created` | "New project: Q1 Planning" | Project created |
| `project.member_added` | "Added to project: Q1 Planning" | User added to project |
| `project.completed` | "Project completed: Q1 Planning" | Project marked done |

### Workspace Notifications
| Type | Title Example | When Triggered |
|------|---------------|----------------|
| `workspace.member_invited` | "Invitation sent to user@email.com" | Invitation sent |
| `workspace.member_joined` | "John joined the workspace" | New member joined |

---

## Real-Time Events (SSE)

The notification store automatically handles SSE events:

```typescript
// Listen for new notifications anywhere in your app
if (browser) {
  window.addEventListener('businessos:notification', (event) => {
    const notification = event.detail;

    // Show toast notification
    toast.info(notification.title, {
      description: notification.body
    });
  });
}
```

### SSE Connection Status

```svelte
<script>
  import { notificationStore } from '$lib/stores/notifications';
  const { isConnected, connectionError } = notificationStore;
</script>

{#if !$isConnected}
  <div class="connection-warning">
    {$connectionError || 'Connecting...'}
  </div>
{/if}
```

---

## Web Push Notifications

### Check Support & Status

```svelte
<script>
  import {
    pushSupported,
    pushPermission,
    pushSubscribed,
    pushLoading
  } from '$lib/services/pushService';
</script>

{#if $pushSupported}
  <p>Permission: {$pushPermission}</p>
  <p>Subscribed: {$pushSubscribed}</p>
{:else}
  <p>Push notifications not supported in this browser</p>
{/if}
```

### Subscribe/Unsubscribe

```svelte
<script>
  import {
    subscribeToPush,
    unsubscribeFromPush,
    pushSubscribed,
    pushLoading
  } from '$lib/services/pushService';
</script>

<button
  on:click={() => $pushSubscribed ? unsubscribeFromPush() : subscribeToPush()}
  disabled={$pushLoading}
>
  {$pushSubscribed ? 'Disable Push' : 'Enable Push'}
</button>
```

### Test Push

```typescript
import { sendTestPush } from '$lib/services/pushService';

// Send a test notification to verify it's working
await sendTestPush();
```

---

## Notification Preferences UI

```svelte
<script>
  import { notificationStore } from '$lib/stores/notifications';
  import { onMount } from 'svelte';

  let preferences = {
    email_enabled: true,
    push_enabled: true,
    in_app_enabled: true,
    quiet_hours_enabled: false,
    quiet_hours_start: '22:00',
    quiet_hours_end: '08:00'
  };

  onMount(async () => {
    const prefs = await notificationStore.getPreferences();
    if (prefs) preferences = prefs;
  });

  async function savePreferences() {
    await notificationStore.updatePreferences(preferences);
  }
</script>

<form on:submit|preventDefault={savePreferences}>
  <label>
    <input type="checkbox" bind:checked={preferences.email_enabled} />
    Email notifications
  </label>

  <label>
    <input type="checkbox" bind:checked={preferences.push_enabled} />
    Push notifications
  </label>

  <label>
    <input type="checkbox" bind:checked={preferences.quiet_hours_enabled} />
    Quiet hours
  </label>

  {#if preferences.quiet_hours_enabled}
    <div class="quiet-hours">
      <input type="time" bind:value={preferences.quiet_hours_start} />
      to
      <input type="time" bind:value={preferences.quiet_hours_end} />
    </div>
  {/if}

  <button type="submit">Save Preferences</button>
</form>
```

---

## Notification Item Component Example

```svelte
<!-- NotificationItem.svelte -->
<script lang="ts">
  import { notificationStore } from '$lib/stores/notifications';
  import { goto } from '$app/navigation';
  import type { Notification } from '$lib/stores/notifications';

  export let notification: Notification;

  function getIcon(type: string) {
    if (type.startsWith('task.')) return 'TaskIcon';
    if (type.startsWith('project.')) return 'FolderIcon';
    if (type.startsWith('comment.')) return 'MessageIcon';
    if (type.startsWith('workspace.')) return 'BuildingIcon';
    return 'BellIcon';
  }

  async function handleClick() {
    // Mark as read
    if (!notification.is_read) {
      await notificationStore.markAsRead(notification.id);
    }

    // Navigate to entity
    if (notification.entity_type && notification.entity_id) {
      goto(`/${notification.entity_type}s/${notification.entity_id}`);
    }
  }

  function formatTime(dateString: string) {
    const date = new Date(dateString);
    const now = new Date();
    const diff = now.getTime() - date.getTime();

    if (diff < 60000) return 'Just now';
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
    return date.toLocaleDateString();
  }
</script>

<button
  class="notification-item"
  class:unread={!notification.is_read}
  on:click={handleClick}
>
  <span class="icon">{getIcon(notification.type)}</span>

  <div class="content">
    <p class="title">{notification.title}</p>
    {#if notification.body}
      <p class="body">{notification.body}</p>
    {/if}
    <span class="time">{formatTime(notification.created_at)}</span>
  </div>

  {#if notification.sender_avatar_url}
    <img src={notification.sender_avatar_url} alt="" class="avatar" />
  {/if}
</button>

<style>
  .notification-item {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px;
    border: none;
    background: transparent;
    text-align: left;
    width: 100%;
    cursor: pointer;
  }

  .notification-item:hover {
    background: var(--hover-bg);
  }

  .notification-item.unread {
    background: var(--unread-bg);
  }

  .title {
    font-weight: 500;
    margin: 0;
  }

  .body {
    color: var(--text-muted);
    font-size: 0.875rem;
    margin: 4px 0 0;
  }

  .time {
    color: var(--text-muted);
    font-size: 0.75rem;
  }

  .avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
  }
</style>
```

---

## API Endpoints Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/notifications` | List notifications |
| GET | `/api/notifications/unread-count` | Get unread count |
| POST | `/api/notifications/:id/read` | Mark as read |
| POST | `/api/notifications/read` | Mark multiple as read |
| POST | `/api/notifications/read-all` | Mark all as read |
| DELETE | `/api/notifications/:id` | Delete notification |
| GET | `/api/notifications/preferences` | Get preferences |
| PUT | `/api/notifications/preferences` | Update preferences |
| GET | `/api/notifications/stream` | SSE stream |
| GET | `/api/notifications/push/vapid-public-key` | Get VAPID key |
| POST | `/api/notifications/push/subscribe` | Subscribe to push |
| POST | `/api/notifications/push/unsubscribe` | Unsubscribe |
| POST | `/api/notifications/push/test` | Send test push |

---

## Files

| File | Purpose |
|------|---------|
| `src/lib/stores/notifications.ts` | Main notification store |
| `src/lib/services/pushService.ts` | Web Push management |
| `static/sw.js` | Service Worker for background push |

---

## Troubleshooting

### SSE Not Connecting
1. Check browser console for errors
2. Verify user is authenticated (session cookie)
3. Check `connectionError` store for details

### Push Not Working
1. Check `pushSupported` - browser must support it
2. Check `pushPermission` - must be 'granted'
3. Verify VAPID keys are configured on backend
4. Check browser's notification settings

### Notifications Not Updating
1. Verify SSE connection is active (`isConnected`)
2. Try `notificationStore.fetchNotifications()` manually
3. Check network tab for SSE stream
