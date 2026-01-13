# 📧 OSA BusinessOS Email Templates

> **Logo Assets Available:**
> - `frontend/static/osa-logo.png` - Primary logo for emails
> - `frontend/osalogobusiness.png` - Business variant
> - Recommended: Host on CDN for email delivery (e.g., `https://cdn.osa.io/logo.png`)

---

## 🎨 Brand Guidelines for Email Templates

### Colors
```css
--osa-primary: #6366f1;      /* Indigo - Primary brand color */
--osa-secondary: #8b5cf6;    /* Purple - Accent color */
--osa-success: #10b981;      /* Green - Success states */
--osa-warning: #f59e0b;      /* Amber - Warning states */
--osa-danger: #ef4444;       /* Red - Error/urgent states */
--osa-text: #1f2937;         /* Dark gray - Body text */
--osa-muted: #6b7280;        /* Gray - Secondary text */
--osa-background: #f9fafb;   /* Light gray - Background */
```

### Typography
- **Headlines:** Inter, -apple-system, sans-serif
- **Body:** 16px line-height 1.6
- **Button text:** 14px semi-bold

---

## 📋 Required Email Templates

### 1. 🔐 Authentication Templates

#### 1.1 `email_verification`
| Field | Value |
|-------|-------|
| **Subject** | Verify your OSA account |
| **Trigger** | New user signup |
| **Priority** | CRITICAL |

**Template Variables:**
```
{{user_name}}          - User's display name
{{verification_link}}  - One-click verification URL
{{verification_code}}  - 6-digit code (alternative)
{{expires_in}}         - Link expiration (e.g., "24 hours")
{{support_email}}      - Support contact
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Welcome to OSA! Please verify your email address to get started.

[Verify Email Button → {{verification_link}}]

Or enter this code: {{verification_code}}

This link expires in {{expires_in}}.

---
If you didn't create an account, please ignore this email.
```

---

#### 1.2 `password_reset`
| Field | Value |
|-------|-------|
| **Subject** | Reset your OSA password |
| **Trigger** | User requests password reset |
| **Priority** | CRITICAL |

**Template Variables:**
```
{{user_name}}        - User's display name
{{reset_link}}       - Password reset URL
{{expires_in}}       - Link expiration (e.g., "1 hour")
{{ip_address}}       - Request origin IP
{{request_time}}     - When request was made
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

We received a request to reset your password.

[Reset Password Button → {{reset_link}}]

This link expires in {{expires_in}}.

Request details:
• Time: {{request_time}}
• IP Address: {{ip_address}}

---
If you didn't request this, your account may be compromised.
Please secure your account immediately.
```

---

#### 1.3 `password_changed`
| Field | Value |
|-------|-------|
| **Subject** | Your OSA password was changed |
| **Trigger** | Password successfully updated |
| **Priority** | HIGH (Security) |

**Template Variables:**
```
{{user_name}}      - User's display name
{{change_time}}    - When password was changed
{{ip_address}}     - Change origin IP
{{device}}         - Browser/device info
{{secure_link}}    - Link to secure account if unauthorized
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Your password was successfully changed.

Change details:
• Time: {{change_time}}
• Device: {{device}}
• IP Address: {{ip_address}}

---
⚠️ If you didn't make this change, your account may be compromised.
[Secure My Account → {{secure_link}}]
```

---

#### 1.4 `magic_link`
| Field | Value |
|-------|-------|
| **Subject** | Your OSA login link |
| **Trigger** | User requests passwordless login |
| **Priority** | CRITICAL |

**Template Variables:**
```
{{user_name}}      - User's display name  
{{magic_link}}     - One-click login URL
{{expires_in}}     - Link expiration (e.g., "15 minutes")
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Click below to sign in to OSA. No password needed!

[Sign In to OSA → {{magic_link}}]

This link expires in {{expires_in}} and can only be used once.

---
If you didn't request this, please ignore this email.
```

---

### 2. 👥 Workspace & Team Templates

#### 2.1 `workspace_invitation`
| Field | Value |
|-------|-------|
| **Subject** | {{inviter_name}} invited you to {{workspace_name}} |
| **Trigger** | Admin invites user to workspace |
| **Priority** | HIGH |
| **Status** | ✅ EXISTS |

**Template Variables:**
```
{{inviter_name}}      - Person who sent invite
{{inviter_email}}     - Inviter's email
{{workspace_name}}    - Workspace being joined
{{role}}              - Assigned role (Admin, Member, etc.)
{{invitation_link}}   - Accept invitation URL
{{expires_in}}        - Invitation expiration
{{personal_message}}  - Optional message from inviter
```

**Content Structure:**
```
[OSA Logo]

Hi there,

{{inviter_name}} ({{inviter_email}}) has invited you to join 
{{workspace_name}} on OSA as a {{role}}.

{{#if personal_message}}
Message from {{inviter_name}}:
"{{personal_message}}"
{{/if}}

[Join {{workspace_name}} → {{invitation_link}}]

This invitation expires in {{expires_in}}.

---
If you don't recognize this invitation, please ignore this email.
```

---

#### 2.2 `role_changed`
| Field | Value |
|-------|-------|
| **Subject** | Your role in {{workspace_name}} has changed |
| **Trigger** | User's role/permissions updated |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}        - User's display name
{{workspace_name}}   - Workspace name
{{old_role}}         - Previous role
{{new_role}}         - New role
{{changed_by}}       - Who made the change
{{change_time}}      - When change occurred
{{workspace_link}}   - Link to workspace
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Your role in {{workspace_name}} has been updated.

Role Change:
• Previous: {{old_role}}
• New: {{new_role}}
• Changed by: {{changed_by}}

[Go to {{workspace_name}} → {{workspace_link}}]

---
If you have questions about this change, contact your workspace admin.
```

---

#### 2.3 `workspace_removal`
| Field | Value |
|-------|-------|
| **Subject** | You've been removed from {{workspace_name}} |
| **Trigger** | User removed from workspace |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}        - User's display name
{{workspace_name}}   - Workspace name
{{removed_by}}       - Who removed them
{{removal_time}}     - When removal occurred
{{reason}}           - Optional reason (if provided)
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

You've been removed from {{workspace_name}}.

Details:
• Removed by: {{removed_by}}
• Time: {{removal_time}}
{{#if reason}}
• Reason: {{reason}}
{{/if}}

You no longer have access to this workspace's projects and tasks.

---
If you believe this was a mistake, contact the workspace admin.
```

---

### 3. ✅ Task Templates

#### 3.1 `task_assigned`
| Field | Value |
|-------|-------|
| **Subject** | {{assigner_name}} assigned you: {{task_title}} |
| **Trigger** | User assigned to a task |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}        - Assignee's name
{{assigner_name}}    - Who assigned the task
{{task_title}}       - Task title
{{task_description}} - Task description (truncated)
{{project_name}}     - Parent project
{{due_date}}         - Due date (if set)
{{priority}}         - Task priority
{{task_link}}        - Direct link to task
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

{{assigner_name}} assigned you a new task.

📋 {{task_title}}
{{#if task_description}}
{{task_description}}
{{/if}}

• Project: {{project_name}}
• Priority: {{priority}}
{{#if due_date}}
• Due: {{due_date}}
{{/if}}

[View Task → {{task_link}}]

---
Reply to this email to add a comment to the task.
```

---

#### 3.2 `task_due_reminder`
| Field | Value |
|-------|-------|
| **Subject** | ⏰ Reminder: {{task_title}} is due {{due_relative}} |
| **Trigger** | Task due in 24-48 hours |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}        - Assignee's name
{{task_title}}       - Task title
{{project_name}}     - Parent project
{{due_date}}         - Exact due date/time
{{due_relative}}     - "tomorrow", "in 2 days", etc.
{{hours_remaining}}  - Hours until due
{{task_link}}        - Direct link to task
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Friendly reminder: you have a task due {{due_relative}}.

📋 {{task_title}}

• Project: {{project_name}}
• Due: {{due_date}} ({{hours_remaining}} hours remaining)

[View Task → {{task_link}}]

---
Update your notification preferences in OSA settings.
```

---

#### 3.3 `task_overdue`
| Field | Value |
|-------|-------|
| **Subject** | 🚨 Overdue: {{task_title}} |
| **Trigger** | Task past due date |
| **Priority** | URGENT |

**Template Variables:**
```
{{user_name}}        - Assignee's name
{{task_title}}       - Task title
{{project_name}}     - Parent project
{{due_date}}         - Original due date
{{days_overdue}}     - Days past due
{{task_link}}        - Direct link to task
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

This task is now overdue.

🚨 {{task_title}}

• Project: {{project_name}}
• Was due: {{due_date}}
• Overdue by: {{days_overdue}} day(s)

[Complete Task → {{task_link}}]

---
Need more time? Update the due date or contact your project lead.
```

---

### 4. 💬 Comments & Mentions Templates

#### 4.1 `mention`
| Field | Value |
|-------|-------|
| **Subject** | {{mentioner_name}} mentioned you in {{entity_title}} |
| **Trigger** | User @mentioned in comment |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}        - Mentioned user's name
{{mentioner_name}}   - Who mentioned them
{{mentioner_avatar}} - Mentioner's avatar URL
{{entity_type}}      - "task", "project", "document"
{{entity_title}}     - Title of entity
{{comment_snippet}}  - First 200 chars of comment
{{comment_link}}     - Direct link to comment
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

{{mentioner_name}} mentioned you:

"{{comment_snippet}}"

In {{entity_type}}: {{entity_title}}

[View Comment → {{comment_link}}]

---
Reply to this email to respond.
```

---

#### 4.2 `comment_reply`
| Field | Value |
|-------|-------|
| **Subject** | {{replier_name}} replied to your comment |
| **Trigger** | Someone replies to user's comment |
| **Priority** | MEDIUM |

**Template Variables:**
```
{{user_name}}             - Original commenter
{{replier_name}}          - Who replied
{{original_snippet}}      - Their original comment (truncated)
{{reply_snippet}}         - The reply content
{{entity_type}}           - "task", "project", etc.
{{entity_title}}          - Title of entity
{{comment_link}}          - Direct link to thread
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

{{replier_name}} replied to your comment on {{entity_title}}:

Your comment:
"{{original_snippet}}"

{{replier_name}}'s reply:
"{{reply_snippet}}"

[View Conversation → {{comment_link}}]

---
Reply to this email to continue the conversation.
```

---

### 5. 🎉 Onboarding & Welcome Templates

#### 5.1 `welcome`
| Field | Value |
|-------|-------|
| **Subject** | Welcome to OSA, {{user_name}}! 🎉 |
| **Trigger** | User completes signup/onboarding |
| **Priority** | HIGH |

**Template Variables:**
```
{{user_name}}           - User's display name
{{workspace_name}}      - Their workspace name
{{business_type}}       - Business type from onboarding
{{getting_started_url}} - Quick start guide link
{{help_center_url}}     - Help documentation
{{community_url}}       - Community/forum link
```

**Content Structure:**
```
[OSA Logo]

Welcome to OSA, {{user_name}}! 🎉

Your workspace "{{workspace_name}}" is ready to go.

Here's how to get started:

1️⃣ Create your first project
2️⃣ Invite your team members  
3️⃣ Set up integrations (Calendar, Slack, etc.)
4️⃣ Explore AI-powered features

[Get Started → {{getting_started_url}}]

Need help?
• [Help Center]({{help_center_url}})
• [Community Forum]({{community_url}})

---
We're here to help you build something great!
The OSA Team
```

---

### 6. 📊 Digest & Summary Templates

#### 6.1 `weekly_digest`
| Field | Value |
|-------|-------|
| **Subject** | Your OSA weekly summary - {{week_range}} |
| **Trigger** | Scheduled weekly (configurable day) |
| **Priority** | MEDIUM |

**Template Variables:**
```
{{user_name}}            - User's name
{{week_range}}           - "Jan 6-12, 2026"
{{tasks_completed}}      - Number completed
{{tasks_created}}        - Number created
{{tasks_overdue}}        - Number overdue
{{comments_received}}    - Comments on their items
{{mentions_count}}       - Times mentioned
{{upcoming_deadlines}}   - Array of upcoming tasks
{{top_projects}}         - Most active projects
{{dashboard_link}}       - Link to dashboard
```

**Content Structure:**
```
[OSA Logo]

Hi {{user_name}},

Here's your week in review ({{week_range}}):

📊 YOUR STATS
────────────────────
✅ Tasks completed: {{tasks_completed}}
📝 Tasks created: {{tasks_created}}
⚠️ Overdue: {{tasks_overdue}}
💬 Comments: {{comments_received}}
@ Mentions: {{mentions_count}}

📅 UPCOMING DEADLINES
────────────────────
{{#each upcoming_deadlines}}
• {{this.title}} - Due {{this.due_date}}
{{/each}}

🔥 ACTIVE PROJECTS
────────────────────
{{#each top_projects}}
• {{this.name}} ({{this.task_count}} tasks)
{{/each}}

[View Dashboard → {{dashboard_link}}]

---
Manage digest settings in your preferences.
```

---

## 📁 Template File Structure

```
email-templates/
├── layouts/
│   ├── base.html           # Base HTML wrapper with logo, footer
│   └── base.txt            # Plain text version
├── auth/
│   ├── email_verification.html
│   ├── email_verification.txt
│   ├── password_reset.html
│   ├── password_reset.txt
│   ├── password_changed.html
│   ├── password_changed.txt
│   ├── magic_link.html
│   └── magic_link.txt
├── workspace/
│   ├── workspace_invitation.html
│   ├── workspace_invitation.txt
│   ├── role_changed.html
│   ├── role_changed.txt
│   ├── workspace_removal.html
│   └── workspace_removal.txt
├── tasks/
│   ├── task_assigned.html
│   ├── task_assigned.txt
│   ├── task_due_reminder.html
│   ├── task_due_reminder.txt
│   ├── task_overdue.html
│   └── task_overdue.txt
├── comments/
│   ├── mention.html
│   ├── mention.txt
│   ├── comment_reply.html
│   └── comment_reply.txt
├── onboarding/
│   ├── welcome.html
│   └── welcome.txt
└── digest/
    ├── weekly_digest.html
    └── weekly_digest.txt
```

---

## ✅ Implementation Checklist

| # | Template | Priority | Status |
|---|----------|----------|--------|
| 1 | `email_verification` | 🔴 CRITICAL | ⬜ TODO |
| 2 | `password_reset` | 🔴 CRITICAL | ⬜ TODO |
| 3 | `password_changed` | 🟠 HIGH | ⬜ TODO |
| 4 | `magic_link` | 🔴 CRITICAL | ⬜ TODO |
| 5 | `workspace_invitation` | 🟠 HIGH | ✅ EXISTS |
| 6 | `role_changed` | 🟠 HIGH | ⬜ TODO |
| 7 | `workspace_removal` | 🟠 HIGH | ⬜ TODO |
| 8 | `task_assigned` | 🟠 HIGH | ⬜ TODO |
| 9 | `task_due_reminder` | 🟠 HIGH | ⬜ TODO |
| 10 | `task_overdue` | 🔴 CRITICAL | ⬜ TODO |
| 11 | `mention` | 🟠 HIGH | ⬜ TODO |
| 12 | `comment_reply` | 🟡 MEDIUM | ⬜ TODO |
| 13 | `welcome` | 🟠 HIGH | ⬜ TODO |
| 14 | `weekly_digest` | 🟡 MEDIUM | ⬜ TODO |

---

## 🛠️ Technical Notes

### Email Provider Recommendations
- **Transactional:** Resend, SendGrid, Postmark, AWS SES
- **Template Engine:** MJML (responsive), Handlebars, Go html/template

### Best Practices
1. Always include plain text version
2. Keep subject lines under 50 characters
3. Include unsubscribe link (required by law)
4. Test on multiple email clients
5. Use inline CSS for compatibility
6. Host images on CDN (not embedded)
7. Include preheader text for preview

### Logo Usage
```html
<img src="https://your-cdn.com/osa-logo.png" 
     alt="OSA" 
     width="120" 
     height="40"
     style="display: block;">
```

---

*Last updated: January 13, 2026*
