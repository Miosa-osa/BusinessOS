# Magic Link Invitation System

**Feature:** Workspace Member Invitations via Magic Link  
**Assigned To:** Javaris  
**Status:** Planning  
**Created:** January 5, 2026

---

## Overview

This document provides complete implementation instructions for the magic link invitation system. This allows workspace owners/admins to invite new members via email with a secure, one-time-use magic link.

### Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         MAGIC LINK INVITATION FLOW                          │
└─────────────────────────────────────────────────────────────────────────────┘

  Admin invites user@email.com
           │
           ▼
  ┌─────────────────┐
  │ Generate Secure │
  │ Token (64 char) │
  └─────────────────┘
           │
           ▼
  ┌─────────────────┐
  │ Store in DB     │
  │ (expires 7 days)│
  └─────────────────┘
           │
           ▼
  ┌─────────────────┐
  │ Send Email via  │
  │ Resend          │
  └─────────────────┘
           │
           ▼
  User receives email with magic link:
  https://app.businessos.com/invite/abc123...
           │
           ▼
  ┌─────────────────┐
  │ User clicks     │
  │ magic link      │
  └─────────────────┘
           │
           ▼
  ┌─────────────────┐
  │ Frontend calls  │──────────────────────┐
  │ GET /api/       │                      │
  │ invitations/    │                      │
  │ :token          │                      │
  └─────────────────┘                      │
           │                               │
      ┌────┴────┐                          │
      │         │                          │
   Valid     Invalid                       │
      │         │                          │
      ▼         ▼                          │
  Show      Show Error                     │
  Details   (expired/used/                 │
      │      revoked)                      │
      │                                    │
      ▼                                    │
  ┌─────────────┐                          │
  │ Has Account?│                          │
  └─────────────┘                          │
      │                                    │
  ┌───┴───┐                                │
  │       │                                │
 YES      NO                               │
  │       │                                │
  ▼       ▼                                │
Sign In  Sign Up ◄─────────────────────────┘
  │       (email pre-filled)
  │       │
  └───┬───┘
      │
      ▼
  ┌─────────────────┐
  │ POST /api/      │
  │ invitations/    │
  │ :token/accept   │
  └─────────────────┘
      │
      ▼
  ┌─────────────────┐
  │ Add user to     │
  │ workspace_      │
  │ members         │
  └─────────────────┘
      │
      ▼
  ┌─────────────────┐
  │ Mark invitation │
  │ as accepted     │
  └─────────────────┘
      │
      ▼
  Redirect to workspace dashboard
```

---

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Token usage | **One-time use** | More secure, clear audit trail, industry standard |
| Token length | **64 characters** (32 bytes hex) | Cryptographically secure, URL-safe |
| Expiration | **7 days** | Balance between security and user convenience |
| New user flow | **Create account + join in one flow** | Frictionless onboarding, email already verified |
| Email provider | **Resend** | Modern API, good deliverability, generous free tier |

---

## Prerequisites

### 1. Resend Account Setup

**Website:** https://resend.com

#### Step 1: Create Account
1. Go to [resend.com](https://resend.com)
2. Sign up with email or GitHub
3. Verify your email

#### Step 2: Get API Key
1. Go to Dashboard → API Keys
2. Click "Create API Key"
3. Name it (e.g., "BusinessOS Production")
4. Copy the key (starts with `re_`)

#### Step 3: Verify Domain (Required for Production)
1. Go to Dashboard → Domains
2. Click "Add Domain"
3. Enter your domain (e.g., `businessos.com`)
4. Add the DNS records Resend provides:

| Type | Name | Value |
|------|------|-------|
| TXT | `resend._domainkey.businessos.com` | (provided by Resend) |
| TXT | `_dmarc.businessos.com` | `v=DMARC1; p=none;` |
| MX | `send.businessos.com` | `feedback-smtp.us-east-1.amazonses.com` |

5. Wait for verification (usually 5-30 minutes)

#### Step 4: Add Environment Variables
```env
# Email (Resend)
RESEND_API_KEY=re_xxxxxxxxxxxxxxxxxxxxxxxxxxxxx
RESEND_FROM_EMAIL=invites@businessos.com

# App URLs
APP_URL=http://localhost:3000           # Development
# APP_URL=https://app.businessos.com    # Production

# Invitation Settings
INVITATION_EXPIRY_DAYS=7
```

#### Cost
- **Free tier:** 3,000 emails/month
- **Pro:** $20/month for 50,000 emails

---

## Database Schema

### Migration File

Create: `migrations/XXXXXX_create_workspace_invitations.sql`

```sql
-- Workspace Invitations (Magic Links)
CREATE TABLE workspace_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    
    -- Invitation Details
    email VARCHAR(255) NOT NULL,
    token VARCHAR(64) NOT NULL UNIQUE,        -- Secure random token (32 bytes hex)
    
    -- Role to assign on accept
    role_id UUID REFERENCES workspace_roles(id) ON DELETE SET NULL,
    role_name VARCHAR(100) NOT NULL,          -- Denormalized for display
    
    -- Inviter Information
    invited_by_id VARCHAR(255) NOT NULL,
    invited_by_name VARCHAR(255),
    
    -- Status
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- 'pending', 'accepted', 'expired', 'revoked'
    
    -- Timestamps
    expires_at TIMESTAMPTZ NOT NULL,          -- Default: created_at + 7 days
    accepted_at TIMESTAMPTZ,
    accepted_by_user_id VARCHAR(255),         -- The user who accepted
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_workspace_invitations_token ON workspace_invitations(token);
CREATE INDEX idx_workspace_invitations_workspace ON workspace_invitations(workspace_id);
CREATE INDEX idx_workspace_invitations_email ON workspace_invitations(email);
CREATE INDEX idx_workspace_invitations_status ON workspace_invitations(status) WHERE status = 'pending';

-- Partial unique index: only one pending invitation per email per workspace
CREATE UNIQUE INDEX idx_workspace_invitations_pending_unique 
ON workspace_invitations(workspace_id, email) 
WHERE status = 'pending';

-- Trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION update_workspace_invitations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_workspace_invitations_updated_at
    BEFORE UPDATE ON workspace_invitations
    FOR EACH ROW
    EXECUTE FUNCTION update_workspace_invitations_updated_at();
```

### Rollback

```sql
DROP TRIGGER IF EXISTS trigger_workspace_invitations_updated_at ON workspace_invitations;
DROP FUNCTION IF EXISTS update_workspace_invitations_updated_at();
DROP TABLE IF EXISTS workspace_invitations;
```

---

## API Endpoints

### Endpoint Summary

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/api/workspaces/:id/members/invite` | ✅ Required | Send magic link invitation |
| `GET` | `/api/workspaces/:id/invitations` | ✅ Required | List workspace invitations |
| `DELETE` | `/api/workspaces/:id/invitations/:invitationId` | ✅ Required | Revoke pending invitation |
| `POST` | `/api/workspaces/:id/invitations/:invitationId/resend` | ✅ Required | Resend invitation email |
| `GET` | `/api/invitations/:token` | ❌ Public | Verify magic link token |
| `POST` | `/api/invitations/:token/accept` | ✅ Required | Accept invitation |

---

### Endpoint Details

#### 1. Send Invitation
```
POST /api/workspaces/:id/members/invite
```

**Headers:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "email": "newmember@example.com",
  "role_name": "member",
  "role_id": "uuid-optional"
}
```

**Response (201 Created):**
```json
{
  "invitation": {
    "id": "uuid",
    "workspace_id": "uuid",
    "email": "newmember@example.com",
    "role_name": "member",
    "invited_by_name": "John Doe",
    "status": "pending",
    "expires_at": "2026-01-12T10:00:00Z",
    "created_at": "2026-01-05T10:00:00Z"
  },
  "message": "Invitation sent successfully"
}
```

**Error Responses:**
| Status | Error | Description |
|--------|-------|-------------|
| 400 | `invalid_email` | Email format invalid |
| 403 | `insufficient_permissions` | User can't invite members |
| 409 | `already_member` | User is already a workspace member |
| 409 | `pending_invitation` | Pending invitation already exists |

---

#### 2. List Invitations
```
GET /api/workspaces/:id/invitations
```

**Query Parameters:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `status` | string | all | Filter by status: `pending`, `accepted`, `expired`, `revoked` |
| `limit` | int | 50 | Max results |
| `offset` | int | 0 | Pagination offset |

**Response (200 OK):**
```json
{
  "invitations": [
    {
      "id": "uuid",
      "email": "user1@example.com",
      "role_name": "member",
      "status": "pending",
      "invited_by_name": "John Doe",
      "expires_at": "2026-01-12T10:00:00Z",
      "created_at": "2026-01-05T10:00:00Z"
    },
    {
      "id": "uuid",
      "email": "user2@example.com",
      "role_name": "admin",
      "status": "accepted",
      "invited_by_name": "John Doe",
      "accepted_at": "2026-01-04T15:30:00Z",
      "created_at": "2026-01-03T10:00:00Z"
    }
  ],
  "total": 2
}
```

---

#### 3. Revoke Invitation
```
DELETE /api/workspaces/:id/invitations/:invitationId
```

**Response (200 OK):**
```json
{
  "message": "Invitation revoked successfully"
}
```

**Error Responses:**
| Status | Error | Description |
|--------|-------|-------------|
| 404 | `not_found` | Invitation not found |
| 409 | `already_processed` | Invitation already accepted/expired |

---

#### 4. Resend Invitation
```
POST /api/workspaces/:id/invitations/:invitationId/resend
```

**Response (200 OK):**
```json
{
  "message": "Invitation resent successfully",
  "expires_at": "2026-01-12T10:00:00Z"
}
```

*Note: This generates a new token and extends expiration.*

---

#### 5. Verify Token (Public)
```
GET /api/invitations/:token
```

*No authentication required - this is the landing page for the magic link.*

**Response (200 OK) - Valid:**
```json
{
  "valid": true,
  "email": "newmember@example.com",
  "workspace": {
    "id": "uuid",
    "name": "Acme Corp",
    "logo_url": "https://..."
  },
  "role_name": "member",
  "invited_by": "John Doe",
  "expires_at": "2026-01-12T10:00:00Z"
}
```

**Response (410 Gone) - Invalid:**
```json
{
  "valid": false,
  "error": "invitation_expired",
  "message": "This invitation has expired. Please request a new one."
}
```

**Error Codes:**
| Code | Description |
|------|-------------|
| `invitation_not_found` | Token doesn't exist |
| `invitation_expired` | Past expiration date |
| `invitation_used` | Already accepted |
| `invitation_revoked` | Cancelled by admin |

---

#### 6. Accept Invitation
```
POST /api/invitations/:token/accept
```

**Headers:**
```
Authorization: Bearer <token>
```

*User must be authenticated. If email doesn't match invitation, returns error.*

**Response (200 OK):**
```json
{
  "message": "Successfully joined workspace",
  "workspace": {
    "id": "uuid",
    "name": "Acme Corp",
    "slug": "acme-corp"
  },
  "role": "member"
}
```

**Error Responses:**
| Status | Error | Description |
|--------|-------|-------------|
| 401 | `unauthorized` | Not authenticated |
| 403 | `email_mismatch` | Logged in email doesn't match invitation |
| 410 | `invitation_expired` | Token expired |
| 410 | `invitation_used` | Already accepted |

---

## Backend Implementation

### 1. Install Dependencies

```bash
cd desktop/backend-go
go get github.com/resend/resend-go/v2
```

### 2. Email Service

Create: `internal/services/email_service.go`

```go
package services

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client    *resend.Client
	fromEmail string
	appURL    string
	enabled   bool
}

func NewEmailService() *EmailService {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: RESEND_API_KEY not set, email sending disabled")
		return &EmailService{enabled: false}
	}

	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = "noreply@businessos.com"
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}

	return &EmailService{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		appURL:    appURL,
		enabled:   true,
	}
}

func (s *EmailService) IsEnabled() bool {
	return s.enabled
}

type InvitationEmailData struct {
	RecipientEmail string
	WorkspaceName  string
	WorkspaceLogo  string
	InviterName    string
	RoleName       string
	Token          string
	ExpiresIn      string
}

func (s *EmailService) SendInvitationEmail(ctx context.Context, data InvitationEmailData) error {
	if !s.enabled {
		log.Printf("Email disabled - would send invitation to %s with token %s", data.RecipientEmail, data.Token)
		return nil
	}

	magicLink := fmt.Sprintf("%s/invite/%s", s.appURL, data.Token)

	html := s.buildInvitationHTML(data, magicLink)
	text := s.buildInvitationText(data, magicLink)

	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("BusinessOS <%s>", s.fromEmail),
		To:      []string{data.RecipientEmail},
		Subject: fmt.Sprintf("%s invited you to join %s on BusinessOS", data.InviterName, data.WorkspaceName),
		Html:    html,
		Text:    text,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Failed to send invitation email to %s: %v", data.RecipientEmail, err)
		return err
	}

	log.Printf("Invitation email sent to %s, ID: %s", data.RecipientEmail, sent.Id)
	return nil
}

func (s *EmailService) buildInvitationHTML(data InvitationEmailData, magicLink string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #1f2937;
            background-color: #f3f4f6;
            margin: 0;
            padding: 0;
        }
        .wrapper {
            max-width: 600px;
            margin: 0 auto;
            padding: 40px 20px;
        }
        .card {
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
            padding: 40px;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo img {
            height: 40px;
        }
        h1 {
            color: #111827;
            font-size: 24px;
            font-weight: 600;
            margin: 0 0 20px 0;
            text-align: center;
        }
        .workspace-name {
            color: #2563eb;
        }
        .invite-box {
            background: #f9fafb;
            border-radius: 8px;
            padding: 20px;
            margin: 24px 0;
            text-align: center;
        }
        .invite-box p {
            margin: 0;
            color: #6b7280;
        }
        .invite-box .role {
            font-size: 18px;
            font-weight: 600;
            color: #111827;
            margin-top: 8px;
        }
        .button {
            display: inline-block;
            background: #2563eb;
            color: white !important;
            padding: 14px 32px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 16px;
            margin: 24px 0;
        }
        .button:hover {
            background: #1d4ed8;
        }
        .button-wrapper {
            text-align: center;
        }
        .link-fallback {
            margin-top: 24px;
            padding-top: 24px;
            border-top: 1px solid #e5e7eb;
        }
        .link-fallback p {
            color: #6b7280;
            font-size: 14px;
            margin: 0 0 8px 0;
        }
        .link-fallback a {
            color: #2563eb;
            word-break: break-all;
            font-size: 14px;
        }
        .footer {
            text-align: center;
            margin-top: 32px;
            color: #9ca3af;
            font-size: 13px;
        }
        .footer p {
            margin: 4px 0;
        }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="card">
            <div class="logo">
                <img src="%s/logo.png" alt="BusinessOS" />
            </div>
            
            <h1>Join <span class="workspace-name">%s</span></h1>
            
            <p><strong>%s</strong> has invited you to join their workspace on BusinessOS.</p>
            
            <div class="invite-box">
                <p>You've been invited as</p>
                <p class="role">%s</p>
            </div>
            
            <div class="button-wrapper">
                <a href="%s" class="button">Accept Invitation</a>
            </div>
            
            <div class="link-fallback">
                <p>Or copy and paste this link into your browser:</p>
                <a href="%s">%s</a>
            </div>
        </div>
        
        <div class="footer">
            <p>This invitation expires in %s.</p>
            <p>If you weren't expecting this invitation, you can safely ignore this email.</p>
        </div>
    </div>
</body>
</html>
`, s.appURL, data.WorkspaceName, data.InviterName, data.RoleName, magicLink, magicLink, magicLink, data.ExpiresIn)
}

func (s *EmailService) buildInvitationText(data InvitationEmailData, magicLink string) string {
	return fmt.Sprintf(`You've been invited to join %s on BusinessOS

%s has invited you to join their workspace as a %s.

Accept the invitation by clicking the link below:
%s

This invitation expires in %s.

If you weren't expecting this invitation, you can safely ignore this email.
`, data.WorkspaceName, data.InviterName, data.RoleName, magicLink, data.ExpiresIn)
}
```

### 3. Invitation Service

Create: `internal/services/invitation_service.go`

```go
package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Errors
var (
	ErrInvitationNotFound    = errors.New("invitation not found")
	ErrInvitationExpired     = errors.New("invitation has expired")
	ErrInvitationUsed        = errors.New("invitation has already been used")
	ErrInvitationRevoked     = errors.New("invitation has been revoked")
	ErrAlreadyMember         = errors.New("user is already a member of this workspace")
	ErrPendingInvitation     = errors.New("a pending invitation already exists for this email")
	ErrEmailMismatch         = errors.New("your email does not match this invitation")
	ErrInsufficientPermission = errors.New("you do not have permission to perform this action")
)

// Constants
const (
	TokenLength           = 32 // 32 bytes = 64 hex characters
	DefaultExpirationDays = 7
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

// Models
type WorkspaceInvitation struct {
	ID               uuid.UUID        `json:"id"`
	WorkspaceID      uuid.UUID        `json:"workspace_id"`
	Email            string           `json:"email"`
	Token            string           `json:"-"` // Never expose in JSON responses
	RoleID           *uuid.UUID       `json:"role_id,omitempty"`
	RoleName         string           `json:"role_name"`
	InvitedByID      string           `json:"invited_by_id"`
	InvitedByName    string           `json:"invited_by_name"`
	Status           InvitationStatus `json:"status"`
	ExpiresAt        time.Time        `json:"expires_at"`
	AcceptedAt       *time.Time       `json:"accepted_at,omitempty"`
	AcceptedByUserID *string          `json:"accepted_by_user_id,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type WorkspaceInfo struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	LogoURL *string   `json:"logo_url,omitempty"`
}

type InvitationVerifyResponse struct {
	Valid       bool           `json:"valid"`
	Email       string         `json:"email,omitempty"`
	Workspace   *WorkspaceInfo `json:"workspace,omitempty"`
	RoleName    string         `json:"role_name,omitempty"`
	InvitedBy   string         `json:"invited_by,omitempty"`
	ExpiresAt   *time.Time     `json:"expires_at,omitempty"`
	Error       string         `json:"error,omitempty"`
	Message     string         `json:"message,omitempty"`
}

type CreateInvitationInput struct {
	WorkspaceID   uuid.UUID
	WorkspaceName string
	WorkspaceLogo string
	Email         string
	RoleID        *uuid.UUID
	RoleName      string
	InvitedByID   string
	InvitedByName string
}

// Service
type InvitationService struct {
	db           *pgxpool.Pool
	emailService *EmailService
	expiryDays   int
}

func NewInvitationService(db *pgxpool.Pool, emailService *EmailService) *InvitationService {
	expiryDays := DefaultExpirationDays
	if days := os.Getenv("INVITATION_EXPIRY_DAYS"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 {
			expiryDays = parsed
		}
	}

	return &InvitationService{
		db:           db,
		emailService: emailService,
		expiryDays:   expiryDays,
	}
}

// GenerateSecureToken creates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, TokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateInvitation creates a new workspace invitation and sends email
func (s *InvitationService) CreateInvitation(ctx context.Context, input CreateInvitationInput) (*WorkspaceInvitation, error) {
	// 1. Check if user is already a member
	var existingMember bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM workspace_members wm
			JOIN users u ON u.id = wm.user_id
			WHERE wm.workspace_id = $1 AND LOWER(u.email) = LOWER($2)
		)
	`, input.WorkspaceID, input.Email).Scan(&existingMember)

	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Error checking existing member: %v", err)
	}
	if existingMember {
		return nil, ErrAlreadyMember
	}

	// 2. Check for existing pending invitation
	var existingPending bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM workspace_invitations 
			WHERE workspace_id = $1 
			AND LOWER(email) = LOWER($2) 
			AND status = 'pending'
			AND expires_at > NOW()
		)
	`, input.WorkspaceID, input.Email).Scan(&existingPending)

	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Error checking existing invitation: %v", err)
	}
	if existingPending {
		return nil, ErrPendingInvitation
	}

	// 3. Generate secure token
	token, err := GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	// 4. Calculate expiration
	expiresAt := time.Now().AddDate(0, 0, s.expiryDays)

	// 5. Create invitation
	invitation := &WorkspaceInvitation{}
	err = s.db.QueryRow(ctx, `
		INSERT INTO workspace_invitations (
			workspace_id, email, token, role_id, role_name,
			invited_by_id, invited_by_name, status, expires_at
		) VALUES ($1, LOWER($2), $3, $4, $5, $6, $7, 'pending', $8)
		RETURNING id, workspace_id, email, role_id, role_name,
			invited_by_id, invited_by_name, status, expires_at, created_at, updated_at
	`, input.WorkspaceID, input.Email, token, input.RoleID, input.RoleName,
		input.InvitedByID, input.InvitedByName, expiresAt,
	).Scan(
		&invitation.ID, &invitation.WorkspaceID, &invitation.Email,
		&invitation.RoleID, &invitation.RoleName, &invitation.InvitedByID,
		&invitation.InvitedByName, &invitation.Status, &invitation.ExpiresAt,
		&invitation.CreatedAt, &invitation.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creating invitation: %v", err)
		return nil, err
	}

	// Store token for email (not returned in JSON)
	invitation.Token = token

	// 6. Send invitation email
	if s.emailService != nil && s.emailService.IsEnabled() {
		err = s.emailService.SendInvitationEmail(ctx, InvitationEmailData{
			RecipientEmail: invitation.Email,
			WorkspaceName:  input.WorkspaceName,
			WorkspaceLogo:  input.WorkspaceLogo,
			InviterName:    input.InvitedByName,
			RoleName:       input.RoleName,
			Token:          token,
			ExpiresIn:      "7 days",
		})
		if err != nil {
			// Log but don't fail - invitation is created
			log.Printf("Failed to send invitation email: %v", err)
		}
	}

	return invitation, nil
}

// VerifyToken checks if a token is valid and returns invitation details
func (s *InvitationService) VerifyToken(ctx context.Context, token string) (*InvitationVerifyResponse, error) {
	var invitation WorkspaceInvitation
	var workspace WorkspaceInfo

	err := s.db.QueryRow(ctx, `
		SELECT 
			i.id, i.workspace_id, i.email, i.role_id, i.role_name,
			i.invited_by_id, i.invited_by_name, i.status, i.expires_at,
			i.accepted_at, i.accepted_by_user_id, i.created_at, i.updated_at,
			w.id, w.name, w.slug, w.logo_url
		FROM workspace_invitations i
		JOIN workspaces w ON w.id = i.workspace_id
		WHERE i.token = $1
	`, token).Scan(
		&invitation.ID, &invitation.WorkspaceID, &invitation.Email,
		&invitation.RoleID, &invitation.RoleName, &invitation.InvitedByID,
		&invitation.InvitedByName, &invitation.Status, &invitation.ExpiresAt,
		&invitation.AcceptedAt, &invitation.AcceptedByUserID,
		&invitation.CreatedAt, &invitation.UpdatedAt,
		&workspace.ID, &workspace.Name, &workspace.Slug, &workspace.LogoURL,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return &InvitationVerifyResponse{
				Valid:   false,
				Error:   "invitation_not_found",
				Message: "This invitation link is invalid.",
			}, ErrInvitationNotFound
		}
		return nil, err
	}

	// Check status
	switch invitation.Status {
	case InvitationStatusAccepted:
		return &InvitationVerifyResponse{
			Valid:   false,
			Error:   "invitation_used",
			Message: "This invitation has already been used.",
		}, ErrInvitationUsed
	case InvitationStatusRevoked:
		return &InvitationVerifyResponse{
			Valid:   false,
			Error:   "invitation_revoked",
			Message: "This invitation has been cancelled.",
		}, ErrInvitationRevoked
	case InvitationStatusExpired:
		return &InvitationVerifyResponse{
			Valid:   false,
			Error:   "invitation_expired",
			Message: "This invitation has expired. Please request a new one.",
		}, ErrInvitationExpired
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		// Update status to expired
		_, _ = s.db.Exec(ctx, `
			UPDATE workspace_invitations SET status = 'expired', updated_at = NOW()
			WHERE id = $1
		`, invitation.ID)

		return &InvitationVerifyResponse{
			Valid:   false,
			Error:   "invitation_expired",
			Message: "This invitation has expired. Please request a new one.",
		}, ErrInvitationExpired
	}

	return &InvitationVerifyResponse{
		Valid:     true,
		Email:     invitation.Email,
		Workspace: &workspace,
		RoleName:  invitation.RoleName,
		InvitedBy: invitation.InvitedByName,
		ExpiresAt: &invitation.ExpiresAt,
	}, nil
}

// AcceptInvitation marks invitation as accepted and adds user to workspace
func (s *InvitationService) AcceptInvitation(ctx context.Context, token string, userID string, userEmail string) (*WorkspaceInfo, error) {
	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Get and verify invitation
	var invitation WorkspaceInvitation
	var workspace WorkspaceInfo

	err = tx.QueryRow(ctx, `
		SELECT 
			i.id, i.workspace_id, i.email, i.role_id, i.role_name,
			i.invited_by_id, i.status, i.expires_at,
			w.id, w.name, w.slug
		FROM workspace_invitations i
		JOIN workspaces w ON w.id = i.workspace_id
		WHERE i.token = $1
		FOR UPDATE OF i
	`, token).Scan(
		&invitation.ID, &invitation.WorkspaceID, &invitation.Email,
		&invitation.RoleID, &invitation.RoleName, &invitation.InvitedByID,
		&invitation.Status, &invitation.ExpiresAt,
		&workspace.ID, &workspace.Name, &workspace.Slug,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	// Check status
	if invitation.Status != InvitationStatusPending {
		switch invitation.Status {
		case InvitationStatusAccepted:
			return nil, ErrInvitationUsed
		case InvitationStatusRevoked:
			return nil, ErrInvitationRevoked
		case InvitationStatusExpired:
			return nil, ErrInvitationExpired
		}
	}

	// Check expiration
	if time.Now().After(invitation.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	// Check email matches (case-insensitive)
	// if !strings.EqualFold(invitation.Email, userEmail) {
	// 	return nil, ErrEmailMismatch
	// }
	// Note: Uncomment above if you want strict email matching

	// 2. Add user to workspace
	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_members (
			workspace_id, user_id, role_id, role_name, status, 
			invited_by, invited_at, joined_at
		) VALUES ($1, $2, $3, $4, 'active', $5, $6, NOW())
		ON CONFLICT (workspace_id, user_id) DO UPDATE SET
			role_id = EXCLUDED.role_id,
			role_name = EXCLUDED.role_name,
			status = 'active',
			joined_at = NOW(),
			updated_at = NOW()
	`, invitation.WorkspaceID, userID, invitation.RoleID, invitation.RoleName,
		invitation.InvitedByID, invitation.CreatedAt)

	if err != nil {
		log.Printf("Error adding user to workspace: %v", err)
		return nil, err
	}

	// 3. Mark invitation as accepted
	_, err = tx.Exec(ctx, `
		UPDATE workspace_invitations 
		SET status = 'accepted', accepted_at = NOW(), accepted_by_user_id = $1, updated_at = NOW()
		WHERE id = $2
	`, userID, invitation.ID)

	if err != nil {
		log.Printf("Error marking invitation as accepted: %v", err)
		return nil, err
	}

	// 4. Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &workspace, nil
}

// RevokeInvitation cancels a pending invitation
func (s *InvitationService) RevokeInvitation(ctx context.Context, invitationID uuid.UUID, workspaceID uuid.UUID) error {
	result, err := s.db.Exec(ctx, `
		UPDATE workspace_invitations 
		SET status = 'revoked', updated_at = NOW()
		WHERE id = $1 AND workspace_id = $2 AND status = 'pending'
	`, invitationID, workspaceID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}

	return nil
}

// ResendInvitation generates a new token and resends the email
func (s *InvitationService) ResendInvitation(ctx context.Context, invitationID uuid.UUID, workspaceID uuid.UUID, workspaceName string) (*WorkspaceInvitation, error) {
	// Generate new token
	token, err := GenerateSecureToken()
	if err != nil {
		return nil, err
	}

	// Update invitation with new token and expiration
	expiresAt := time.Now().AddDate(0, 0, s.expiryDays)

	invitation := &WorkspaceInvitation{}
	err = s.db.QueryRow(ctx, `
		UPDATE workspace_invitations 
		SET token = $1, expires_at = $2, updated_at = NOW()
		WHERE id = $3 AND workspace_id = $4 AND status = 'pending'
		RETURNING id, workspace_id, email, role_id, role_name,
			invited_by_id, invited_by_name, status, expires_at, created_at, updated_at
	`, token, expiresAt, invitationID, workspaceID).Scan(
		&invitation.ID, &invitation.WorkspaceID, &invitation.Email,
		&invitation.RoleID, &invitation.RoleName, &invitation.InvitedByID,
		&invitation.InvitedByName, &invitation.Status, &invitation.ExpiresAt,
		&invitation.CreatedAt, &invitation.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrInvitationNotFound
		}
		return nil, err
	}

	// Send email
	if s.emailService != nil && s.emailService.IsEnabled() {
		err = s.emailService.SendInvitationEmail(ctx, InvitationEmailData{
			RecipientEmail: invitation.Email,
			WorkspaceName:  workspaceName,
			InviterName:    invitation.InvitedByName,
			RoleName:       invitation.RoleName,
			Token:          token,
			ExpiresIn:      "7 days",
		})
		if err != nil {
			log.Printf("Failed to resend invitation email: %v", err)
		}
	}

	return invitation, nil
}

// ListWorkspaceInvitations returns all invitations for a workspace
func (s *InvitationService) ListWorkspaceInvitations(ctx context.Context, workspaceID uuid.UUID, statusFilter string) ([]WorkspaceInvitation, error) {
	query := `
		SELECT 
			id, workspace_id, email, role_id, role_name,
			invited_by_id, invited_by_name, status, expires_at,
			accepted_at, accepted_by_user_id, created_at, updated_at
		FROM workspace_invitations
		WHERE workspace_id = $1
	`

	args := []interface{}{workspaceID}

	if statusFilter != "" && statusFilter != "all" {
		query += " AND status = $2"
		args = append(args, statusFilter)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []WorkspaceInvitation
	for rows.Next() {
		var inv WorkspaceInvitation
		err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.RoleID, &inv.RoleName,
			&inv.InvitedByID, &inv.InvitedByName, &inv.Status, &inv.ExpiresAt,
			&inv.AcceptedAt, &inv.AcceptedByUserID, &inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}
```

### 4. Handlers

Create: `internal/handlers/invitation_handlers.go`

```go
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhl/businessos-backend/internal/middleware"
	"github.com/rhl/businessos-backend/internal/services"
)

// InviteMember sends a magic link invitation
// POST /api/workspaces/:id/members/invite
func (h *Handlers) InviteMember(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var input struct {
		Email    string     `json:"email" binding:"required,email"`
		RoleName string     `json:"role_name" binding:"required"`
		RoleID   *uuid.UUID `json:"role_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Check if user has permission to invite members
	// TODO: Get workspace name from DB

	invitation, err := h.invitationService.CreateInvitation(c.Request.Context(), services.CreateInvitationInput{
		WorkspaceID:   workspaceID,
		WorkspaceName: "Workspace", // TODO: Get from DB
		Email:         input.Email,
		RoleID:        input.RoleID,
		RoleName:      input.RoleName,
		InvitedByID:   user.ID,
		InvitedByName: user.Email, // Or user.Name if available
	})

	if err != nil {
		switch err {
		case services.ErrAlreadyMember:
			c.JSON(http.StatusConflict, gin.H{"error": "already_member", "message": "User is already a member of this workspace"})
		case services.ErrPendingInvitation:
			c.JSON(http.StatusConflict, gin.H{"error": "pending_invitation", "message": "A pending invitation already exists for this email"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"invitation": invitation,
		"message":    "Invitation sent successfully",
	})
}

// ListInvitations returns all invitations for a workspace
// GET /api/workspaces/:id/invitations
func (h *Handlers) ListInvitations(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	statusFilter := c.DefaultQuery("status", "all")

	invitations, err := h.invitationService.ListWorkspaceInvitations(c.Request.Context(), workspaceID, statusFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invitations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invitations": invitations,
		"total":       len(invitations),
	})
}

// RevokeInvitation cancels a pending invitation
// DELETE /api/workspaces/:id/invitations/:invitationId
func (h *Handlers) RevokeInvitation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	invitationID, err := uuid.Parse(c.Param("invitationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	err = h.invitationService.RevokeInvitation(c.Request.Context(), invitationID, workspaceID)
	if err != nil {
		if err == services.ErrInvitationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found or already processed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke invitation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invitation revoked successfully"})
}

// ResendInvitation resends an invitation email
// POST /api/workspaces/:id/invitations/:invitationId/resend
func (h *Handlers) ResendInvitation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	workspaceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	invitationID, err := uuid.Parse(c.Param("invitationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitation ID"})
		return
	}

	// TODO: Get workspace name from DB
	invitation, err := h.invitationService.ResendInvitation(c.Request.Context(), invitationID, workspaceID, "Workspace")
	if err != nil {
		if err == services.ErrInvitationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invitation not found or already processed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resend invitation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Invitation resent successfully",
		"expires_at": invitation.ExpiresAt,
	})
}

// VerifyInvitation checks if a magic link token is valid (public endpoint)
// GET /api/invitations/:token
func (h *Handlers) VerifyInvitation(c *gin.Context) {
	token := c.Param("token")

	response, err := h.invitationService.VerifyToken(c.Request.Context(), token)
	if err != nil {
		status := http.StatusBadRequest

		switch err {
		case services.ErrInvitationNotFound:
			status = http.StatusNotFound
		case services.ErrInvitationExpired, services.ErrInvitationUsed, services.ErrInvitationRevoked:
			status = http.StatusGone
		}

		c.JSON(status, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

// AcceptInvitation accepts a magic link invitation
// POST /api/invitations/:token/accept
func (h *Handlers) AcceptInvitation(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	token := c.Param("token")

	workspace, err := h.invitationService.AcceptInvitation(c.Request.Context(), token, user.ID, user.Email)
	if err != nil {
		switch err {
		case services.ErrInvitationNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation_not_found", "message": "Invitation not found"})
		case services.ErrInvitationExpired:
			c.JSON(http.StatusGone, gin.H{"error": "invitation_expired", "message": "Invitation has expired"})
		case services.ErrInvitationUsed:
			c.JSON(http.StatusGone, gin.H{"error": "invitation_used", "message": "Invitation has already been used"})
		case services.ErrInvitationRevoked:
			c.JSON(http.StatusGone, gin.H{"error": "invitation_revoked", "message": "Invitation has been cancelled"})
		case services.ErrEmailMismatch:
			c.JSON(http.StatusForbidden, gin.H{"error": "email_mismatch", "message": "Your email does not match this invitation"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invitation"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Successfully joined workspace",
		"workspace": workspace,
	})
}
```

### 5. Route Registration

Add to `internal/handlers/handlers.go`:

```go
// Workspace Invitations (authenticated)
workspaces := api.Group("/workspaces")
{
    // ... existing workspace routes ...
    
    // Member Invitations
    workspaces.POST("/:id/members/invite", h.InviteMember)
    workspaces.GET("/:id/invitations", h.ListInvitations)
    workspaces.DELETE("/:id/invitations/:invitationId", h.RevokeInvitation)
    workspaces.POST("/:id/invitations/:invitationId/resend", h.ResendInvitation)
}

// Public invitation verification (no auth required)
invitations := r.Group("/api/invitations")
{
    invitations.GET("/:token", h.VerifyInvitation)
}

// Protected invitation acceptance
invitationsAuth := api.Group("/invitations")
{
    invitationsAuth.POST("/:token/accept", h.AcceptInvitation)
}
```

---

## Frontend Implementation

### Route: `/invite/:token`

```typescript
// src/routes/invite/[token]/+page.svelte
<script lang="ts">
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  
  let invitation: InvitationVerifyResponse | null = null;
  let error: string | null = null;
  let loading = true;
  let accepting = false;
  
  const token = $page.params.token;
  
  onMount(async () => {
    try {
      const res = await fetch(`/api/invitations/${token}`);
      invitation = await res.json();
      
      if (!invitation.valid) {
        error = invitation.message;
      }
    } catch (e) {
      error = 'Failed to verify invitation';
    } finally {
      loading = false;
    }
  });
  
  async function acceptInvitation() {
    accepting = true;
    try {
      const res = await fetch(`/api/invitations/${token}/accept`, {
        method: 'POST',
        credentials: 'include'
      });
      
      if (res.ok) {
        const data = await res.json();
        goto(`/workspace/${data.workspace.slug}`);
      } else if (res.status === 401) {
        // Redirect to login with return URL
        goto(`/login?redirect=/invite/${token}`);
      } else {
        const data = await res.json();
        error = data.message;
      }
    } catch (e) {
      error = 'Failed to accept invitation';
    } finally {
      accepting = false;
    }
  }
</script>

{#if loading}
  <div class="loading">Verifying invitation...</div>
{:else if error}
  <div class="error">
    <h1>Invalid Invitation</h1>
    <p>{error}</p>
    <a href="/">Go to Home</a>
  </div>
{:else if invitation?.valid}
  <div class="invitation">
    <h1>You're invited!</h1>
    <p>
      <strong>{invitation.invited_by}</strong> has invited you to join
      <strong>{invitation.workspace?.name}</strong> as a <strong>{invitation.role_name}</strong>.
    </p>
    
    <button on:click={acceptInvitation} disabled={accepting}>
      {accepting ? 'Joining...' : 'Accept Invitation'}
    </button>
    
    <p class="expires">
      This invitation expires on {new Date(invitation.expires_at).toLocaleDateString()}
    </p>
  </div>
{/if}
```

---

## Testing

### Manual Testing Checklist

- [ ] Create invitation - verify email is sent (or logged if disabled)
- [ ] Click magic link - verify invitation details shown
- [ ] Accept as logged-in user - verify added to workspace
- [ ] Accept as new user - verify signup flow then added to workspace
- [ ] Expired token - verify proper error message
- [ ] Already used token - verify proper error message
- [ ] Revoke invitation - verify token no longer works
- [ ] Resend invitation - verify new token generated, old invalidated
- [ ] Duplicate invitation - verify error returned

### API Testing with cURL

```bash
# 1. Create invitation
curl -X POST http://localhost:8080/api/workspaces/{workspace_id}/members/invite \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "role_name": "member"}'

# 2. Verify token (no auth)
curl http://localhost:8080/api/invitations/{magic_token}

# 3. Accept invitation
curl -X POST http://localhost:8080/api/invitations/{magic_token}/accept \
  -H "Authorization: Bearer {token}"

# 4. List invitations
curl http://localhost:8080/api/workspaces/{workspace_id}/invitations \
  -H "Authorization: Bearer {token}"

# 5. Revoke invitation
curl -X DELETE http://localhost:8080/api/workspaces/{workspace_id}/invitations/{invitation_id} \
  -H "Authorization: Bearer {token}"
```

---

## Security Considerations

1. **Token Entropy:** 32 bytes (256 bits) of randomness - sufficient for security
2. **One-Time Use:** Tokens are invalidated immediately after acceptance
3. **Expiration:** 7-day default expiration limits exposure window
4. **Email Verification:** Clicking the link implicitly verifies email ownership
5. **Rate Limiting:** TODO - Add rate limiting on invitation creation
6. **Audit Trail:** All invitations are logged with inviter info

---

## Environment Variables

```env
# Required
RESEND_API_KEY=re_xxxxxxxxxxxxx           # Get from resend.com dashboard
RESEND_FROM_EMAIL=invites@businessos.com  # Must be verified domain
APP_URL=https://app.businessos.com        # Base URL for magic links

# Optional
INVITATION_EXPIRY_DAYS=7                  # Default expiration period
```

---

## Troubleshooting

### Emails not sending
1. Check `RESEND_API_KEY` is set correctly
2. Verify domain in Resend dashboard
3. Check logs for email service errors

### Token not found
1. Verify token hasn't expired
2. Check token wasn't already used
3. Ensure token is URL-decoded properly

### Permission errors
1. Verify user has `team.invite` permission
2. Check workspace membership status

---

## Future Enhancements

- [ ] Bulk invitations (CSV upload)
- [ ] Custom invitation messages
- [ ] Invitation link sharing (shareable link vs email)
- [ ] Role-based invitation permissions
- [ ] Invitation analytics (open rates, acceptance rates)
