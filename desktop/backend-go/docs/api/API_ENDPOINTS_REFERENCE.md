# API Endpoints Reference

**Base URL:** `http://localhost:8080/api`
**Authentication:** Bearer token required (Supabase JWT)
**Version:** 1.0.0
**Last Updated:** January 6, 2026

---

## Table of Contents

1. [Authentication](#authentication)
2. [Workspace Management](#workspace-management)
3. [Workspace Members](#workspace-members)
4. [Workspace Roles](#workspace-roles)
5. [Workspace Invites](#workspace-invites)
6. [Workspace Audit Logs](#workspace-audit-logs)
7. [Workspace Memories](#workspace-memories)
8. [Project Members](#project-members)
9. [RAG Search](#rag-search)
10. [Multi-Modal Search](#multi-modal-search)
11. [Error Codes](#error-codes)

---

## Authentication

All endpoints require authentication via Bearer token in the Authorization header:

```bash
Authorization: Bearer <supabase_jwt_token>
```

The token is validated using Supabase public key and contains the user ID.

---

## Workspace Management

### Create Workspace

Create a new workspace.

**Endpoint:** `POST /api/workspaces`

**Request Body:**
```json
{
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "description": "Main workspace for Acme Corp",
  "logo_url": "https://example.com/logo.png",
  "plan_type": "professional",
  "settings": {
    "allow_public_sharing": false,
    "require_2fa": true,
    "default_project_visibility": "private"
  }
}
```

**Response:** `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "description": "Main workspace for Acme Corp",
  "logo_url": "https://example.com/logo.png",
  "plan_type": "professional",
  "max_members": 50,
  "max_projects": 100,
  "max_storage_gb": 500,
  "settings": {
    "allow_public_sharing": false,
    "require_2fa": true,
    "default_project_visibility": "private"
  },
  "owner_id": "user-123",
  "created_at": "2026-01-06T10:00:00Z",
  "updated_at": "2026-01-06T10:00:00Z"
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/workspaces \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "description": "Main workspace for Acme Corp"
  }'
```

---

### List Workspaces

List all workspaces for the authenticated user.

**Endpoint:** `GET /api/workspaces`

**Query Parameters:**
- `limit` (optional) - Max results, default: 50
- `offset` (optional) - Pagination offset, default: 0

**Response:** `200 OK`
```json
{
  "workspaces": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "Acme Corporation",
      "slug": "acme-corp",
      "description": "Main workspace",
      "logo_url": "https://example.com/logo.png",
      "plan_type": "professional",
      "role": "owner",
      "role_display_name": "Owner",
      "member_count": 12,
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### Get Workspace

Get workspace details.

**Endpoint:** `GET /api/workspaces/:id`

**Response:** `200 OK`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "description": "Main workspace for Acme Corp",
  "logo_url": "https://example.com/logo.png",
  "plan_type": "professional",
  "max_members": 50,
  "max_projects": 100,
  "max_storage_gb": 500,
  "settings": {},
  "owner_id": "user-123",
  "member_count": 12,
  "project_count": 8,
  "created_at": "2026-01-06T10:00:00Z",
  "updated_at": "2026-01-06T10:00:00Z"
}
```

---

### Update Workspace

Update workspace details.

**Endpoint:** `PUT /api/workspaces/:id`

**Required Permission:** Workspace admin or owner

**Request Body:**
```json
{
  "name": "Acme Corporation Updated",
  "description": "Updated description",
  "logo_url": "https://example.com/new-logo.png",
  "settings": {
    "allow_public_sharing": true
  }
}
```

**Response:** `200 OK` (Updated workspace object)

---

### Delete Workspace

Delete a workspace (requires owner role).

**Endpoint:** `DELETE /api/workspaces/:id`

**Required Permission:** Workspace owner only

**Response:** `204 No Content`

---

## Workspace Members

### List Members

List all members in a workspace.

**Endpoint:** `GET /api/workspaces/:id/members`

**Query Parameters:**
- `status` (optional) - Filter by status: active, invited, suspended
- `role` (optional) - Filter by role name

**Response:** `200 OK`
```json
{
  "members": [
    {
      "id": "member-uuid",
      "user_id": "user-123",
      "workspace_id": "workspace-uuid",
      "role_id": "role-uuid",
      "role_name": "admin",
      "role_display_name": "Administrator",
      "status": "active",
      "joined_at": "2026-01-06T10:00:00Z",
      "profile": {
        "display_name": "John Doe",
        "title": "Engineering Manager",
        "department": "Engineering",
        "avatar_url": "https://example.com/avatar.jpg"
      }
    }
  ],
  "total": 12
}
```

---

### Invite Member

Send workspace invitation via email.

**Endpoint:** `POST /api/workspaces/:id/members/invite`

**Required Permission:** invite_members permission

**Request Body:**
```json
{
  "email": "newmember@example.com",
  "role_id": "role-uuid",
  "message": "Welcome to our workspace!"
}
```

**Response:** `201 Created`
```json
{
  "id": "invite-uuid",
  "workspace_id": "workspace-uuid",
  "email": "newmember@example.com",
  "role_id": "role-uuid",
  "invited_by": "user-123",
  "invited_at": "2026-01-06T10:00:00Z",
  "expires_at": "2026-01-13T10:00:00Z",
  "token": "secure-random-token",
  "status": "pending"
}
```

---

### Update Member Role

Update a member's role in the workspace.

**Endpoint:** `PUT /api/workspaces/:id/members/:userId`

**Required Permission:** manage_roles permission

**Request Body:**
```json
{
  "role_id": "new-role-uuid"
}
```

**Response:** `200 OK` (Updated member object)

---

### Remove Member

Remove a member from the workspace.

**Endpoint:** `DELETE /api/workspaces/:id/members/:userId`

**Required Permission:** manage_members permission

**Response:** `204 No Content`

---

## Workspace Roles

### List Roles

List all roles in a workspace.

**Endpoint:** `GET /api/workspaces/:id/roles`

**Response:** `200 OK`
```json
{
  "roles": [
    {
      "id": "role-uuid",
      "workspace_id": "workspace-uuid",
      "name": "admin",
      "display_name": "Administrator",
      "description": "Full access except ownership transfer",
      "color": "#3B82F6",
      "icon": "shield",
      "permissions": {
        "projects": {
          "create": true,
          "read": true,
          "update": true,
          "delete": true
        },
        "workspace": {
          "invite_members": true,
          "manage_roles": true,
          "manage_billing": false
        }
      },
      "hierarchy_level": 5,
      "is_system": true,
      "is_default": false,
      "member_count": 3
    }
  ],
  "total": 6
}
```

---

## Workspace Invites

### List Invites

List pending invites for a workspace.

**Endpoint:** `GET /api/workspaces/:id/invites`

**Query Parameters:**
- `status` (optional) - Filter by status: pending, accepted, expired, revoked

**Response:** `200 OK`
```json
{
  "invites": [
    {
      "id": "invite-uuid",
      "workspace_id": "workspace-uuid",
      "email": "newuser@example.com",
      "role_id": "role-uuid",
      "role_name": "member",
      "invited_by": "user-123",
      "invited_by_name": "John Doe",
      "invited_at": "2026-01-06T10:00:00Z",
      "expires_at": "2026-01-13T10:00:00Z",
      "status": "pending"
    }
  ],
  "total": 3
}
```

---

### Delete Invite

Revoke a pending invite.

**Endpoint:** `DELETE /api/workspaces/:id/invites/:inviteId`

**Required Permission:** invite_members permission

**Response:** `204 No Content`

---

### Accept Invite

Accept a workspace invitation (public endpoint, uses token).

**Endpoint:** `POST /api/workspaces/invites/accept`

**Request Body:**
```json
{
  "token": "secure-random-token"
}
```

**Response:** `200 OK`
```json
{
  "workspace_id": "workspace-uuid",
  "workspace_name": "Acme Corporation",
  "role_name": "member",
  "message": "Successfully joined workspace"
}
```

---

## Workspace Audit Logs

### List Audit Logs

List audit logs for a workspace.

**Endpoint:** `GET /api/workspaces/:id/audit-logs`

**Query Parameters:**
- `user_id` (optional) - Filter by user
- `action` (optional) - Filter by action type
- `resource_type` (optional) - Filter by resource type
- `limit` (optional) - Max results, default: 50
- `offset` (optional) - Pagination offset

**Response:** `200 OK`
```json
{
  "logs": [
    {
      "id": "log-uuid",
      "workspace_id": "workspace-uuid",
      "user_id": "user-123",
      "user_name": "John Doe",
      "action": "create_project",
      "resource_type": "project",
      "resource_id": "project-uuid",
      "details": {
        "project_name": "New Project",
        "visibility": "private"
      },
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "total": 145
}
```

---

### Get Audit Log

Get a specific audit log entry.

**Endpoint:** `GET /api/workspaces/:id/audit-logs/:logId`

**Response:** `200 OK` (Single log object)

---

### Get User Audit Logs

Get audit logs for a specific user.

**Endpoint:** `GET /api/workspaces/:id/audit-logs/user/:userId`

**Response:** `200 OK` (Array of log objects)

---

### Get Resource Audit Logs

Get audit logs for a specific resource.

**Endpoint:** `GET /api/workspaces/:id/audit-logs/resource/:resourceType/:resourceId`

**Response:** `200 OK` (Array of log objects)

---

### Get Action Statistics

Get statistics about actions in the workspace.

**Endpoint:** `GET /api/workspaces/:id/audit-logs/stats/actions`

**Query Parameters:**
- `days` (optional) - Number of days to look back, default: 30

**Response:** `200 OK`
```json
{
  "action_counts": {
    "create_project": 12,
    "invite_member": 8,
    "update_project": 45,
    "delete_task": 23
  },
  "total_actions": 88,
  "period_days": 30
}
```

---

### Get Active Users Statistics

Get statistics about active users.

**Endpoint:** `GET /api/workspaces/:id/audit-logs/stats/active-users`

**Query Parameters:**
- `days` (optional) - Number of days to look back, default: 30

**Response:** `200 OK`
```json
{
  "active_users": [
    {
      "user_id": "user-123",
      "user_name": "John Doe",
      "action_count": 45,
      "last_action": "2026-01-06T09:30:00Z"
    }
  ],
  "total_active_users": 8,
  "period_days": 30
}
```

---

## Workspace Memories

### Create Memory

Create a new workspace or private memory.

**Endpoint:** `POST /api/workspaces/:id/memories`

**Request Body:**
```json
{
  "title": "API Authentication Best Practices",
  "summary": "Guidelines for implementing secure authentication",
  "content": "Detailed content about authentication patterns...",
  "memory_type": "pattern",
  "category": "security",
  "visibility": "workspace",
  "tags": ["auth", "security", "api"],
  "importance_score": 0.85,
  "scope_type": "project",
  "scope_id": "project-uuid"
}
```

**Visibility Options:**
- `workspace` - Shared with all workspace members (owner_user_id must be null)
- `private` - Accessible only to owner (owner_user_id required)
- `shared` - Owner + specific users (created as private, then shared)

**Memory Types:**
- `general` - General information
- `decision` - Architectural decisions
- `pattern` - Reusable patterns
- `context` - Contextual information
- `learning` - Learned behaviors
- `preference` - User preferences

**Response:** `201 Created`
```json
{
  "id": "memory-uuid",
  "workspace_id": "workspace-uuid",
  "title": "API Authentication Best Practices",
  "summary": "Guidelines for implementing secure authentication",
  "content": "Detailed content...",
  "memory_type": "pattern",
  "category": "security",
  "visibility": "workspace",
  "owner_user_id": null,
  "shared_with": null,
  "tags": ["auth", "security", "api"],
  "importance_score": 0.85,
  "access_count": 0,
  "scope_type": "project",
  "scope_id": "project-uuid",
  "is_pinned": false,
  "is_active": true,
  "created_by": "user-123",
  "created_at": "2026-01-06T10:00:00Z",
  "updated_at": "2026-01-06T10:00:00Z",
  "last_accessed_at": null
}
```

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/workspaces/${WORKSPACE_ID}/memories \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "API Authentication Best Practices",
    "summary": "Guidelines for secure auth",
    "content": "Always use JWT tokens...",
    "memory_type": "pattern",
    "visibility": "private",
    "tags": ["auth", "security"]
  }'
```

---

### List Workspace Memories

List all workspace-level memories (accessible to all members).

**Endpoint:** `GET /api/workspaces/:id/memories`

**Query Parameters:**
- `type` (optional) - Filter by memory_type
- `category` (optional) - Filter by category
- `tags` (optional) - Filter by tags (comma-separated)
- `limit` (optional) - Max results, default: 50
- `offset` (optional) - Pagination offset

**Response:** `200 OK`
```json
{
  "memories": [
    {
      "id": "memory-uuid",
      "workspace_id": "workspace-uuid",
      "title": "API Authentication Best Practices",
      "summary": "Guidelines for implementing secure authentication",
      "content": "Detailed content...",
      "memory_type": "pattern",
      "category": "security",
      "visibility": "workspace",
      "tags": ["auth", "security", "api"],
      "importance_score": 0.85,
      "access_count": 12,
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "total": 23
}
```

---

### List Private Memories

List user's private and shared memories.

**Endpoint:** `GET /api/workspaces/:id/memories/private`

**Query Parameters:** Same as List Workspace Memories

**Response:** `200 OK`
```json
{
  "memories": [
    {
      "id": "memory-uuid",
      "workspace_id": "workspace-uuid",
      "title": "Personal Notes on Database Design",
      "summary": "My thoughts on schema design",
      "content": "...",
      "memory_type": "general",
      "visibility": "private",
      "owner_user_id": "user-123",
      "shared_with": null,
      "tags": ["database", "personal"],
      "importance_score": 0.7,
      "access_count": 5,
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "total": 8
}
```

---

### List Accessible Memories

List all memories accessible to the user (workspace + private + shared).

**Endpoint:** `GET /api/workspaces/:id/memories/accessible`

**Query Parameters:** Same as List Workspace Memories

**Response:** `200 OK`
```json
{
  "memories": [
    {
      "id": "memory-uuid",
      "title": "...",
      "visibility": "workspace",
      "is_owner": false,
      "..."
    },
    {
      "id": "memory-uuid-2",
      "title": "...",
      "visibility": "private",
      "is_owner": true,
      "..."
    },
    {
      "id": "memory-uuid-3",
      "title": "...",
      "visibility": "shared",
      "is_owner": false,
      "shared_with": ["user-123", "user-456"],
      "..."
    }
  ],
  "total": 45
}
```

---

### Share Memory

Share a private memory with specific users.

**Endpoint:** `POST /api/workspaces/:id/memories/:memoryId/share`

**Required:** Must be owner of the memory, memory must be private

**Request Body:**
```json
{
  "user_ids": ["user-456", "user-789"]
}
```

**Response:** `200 OK`
```json
{
  "id": "memory-uuid",
  "visibility": "shared",
  "shared_with": ["user-456", "user-789"],
  "message": "Memory shared successfully"
}
```

---

### Unshare Memory

Revert a shared memory back to private.

**Endpoint:** `DELETE /api/workspaces/:id/memories/:memoryId/share`

**Required:** Must be owner of the memory

**Response:** `200 OK`
```json
{
  "id": "memory-uuid",
  "visibility": "private",
  "shared_with": null,
  "message": "Memory is now private"
}
```

---

### Delete Memory

Delete a memory.

**Endpoint:** `DELETE /api/workspaces/:id/memories/:memoryId`

**Required Permissions:**
- For workspace memories: workspace admin or owner
- For private/shared memories: owner only

**Response:** `204 No Content`

---

## Project Members

### List Project Members

List all members assigned to a project.

**Endpoint:** `GET /api/projects/:id/members`

**Response:** `200 OK`
```json
{
  "members": [
    {
      "id": "member-uuid",
      "project_id": "project-uuid",
      "user_id": "user-123",
      "workspace_id": "workspace-uuid",
      "role": "lead",
      "role_display_name": "Project Lead",
      "can_edit": true,
      "can_delete": true,
      "can_invite": true,
      "assigned_by": "user-456",
      "assigned_at": "2026-01-06T10:00:00Z",
      "status": "active",
      "user_profile": {
        "display_name": "John Doe",
        "title": "Senior Engineer",
        "avatar_url": "https://example.com/avatar.jpg"
      }
    }
  ],
  "total": 5
}
```

---

### Add Project Member

Add a member to a project.

**Endpoint:** `POST /api/projects/:id/members`

**Required Permission:** can_invite on the project

**Request Body:**
```json
{
  "user_id": "user-789",
  "role": "contributor"
}
```

**Role Options:**
- `lead` - Full project control (can_edit: true, can_delete: true, can_invite: true)
- `contributor` - Can edit project content (can_edit: true, can_delete: false, can_invite: false)
- `reviewer` - Can view and comment (can_edit: false, can_delete: false, can_invite: false)
- `viewer` - Read-only access (can_edit: false, can_delete: false, can_invite: false)

**Response:** `201 Created`
```json
{
  "id": "member-uuid",
  "project_id": "project-uuid",
  "user_id": "user-789",
  "workspace_id": "workspace-uuid",
  "role": "contributor",
  "can_edit": true,
  "can_delete": false,
  "can_invite": false,
  "assigned_by": "user-123",
  "assigned_at": "2026-01-06T10:00:00Z",
  "status": "active"
}
```

---

### Update Project Member Role

Update a member's role in a project.

**Endpoint:** `PUT /api/projects/:id/members/:memberId/role`

**Required Permission:** lead role on the project

**Request Body:**
```json
{
  "role": "reviewer"
}
```

**Response:** `200 OK` (Updated member object)

---

### Remove Project Member

Remove a member from a project.

**Endpoint:** `DELETE /api/projects/:id/members/:memberId`

**Required Permission:** lead role on the project

**Response:** `204 No Content`

---

### Check Project Access

Check if a user has access to a project.

**Endpoint:** `GET /api/projects/:id/access/:userId`

**Response:** `200 OK`
```json
{
  "has_access": true,
  "role": "contributor",
  "permissions": {
    "can_edit": true,
    "can_delete": false,
    "can_invite": false
  }
}
```

---

## RAG Search

### Hybrid Search

Perform hybrid search combining semantic and keyword approaches.

**Endpoint:** `POST /api/rag/search/hybrid`

**Request Body:**
```json
{
  "query": "authentication best practices",
  "semantic_weight": 0.7,
  "keyword_weight": 0.3,
  "max_results": 10,
  "min_similarity": 0.3
}
```

**Response:** `200 OK`
```json
{
  "query": "authentication best practices",
  "results": [
    {
      "context_id": "context-uuid",
      "block_id": "block-123",
      "block_type": "paragraph",
      "content": "When implementing authentication...",
      "context_name": "Security Documentation",
      "context_type": "document",
      "semantic_score": 0.89,
      "keyword_score": 0.45,
      "hybrid_score": 0.82,
      "search_strategy": "hybrid"
    }
  ],
  "count": 10,
  "options": {
    "semantic_weight": 0.7,
    "keyword_weight": 0.3,
    "max_results": 10
  }
}
```

**See:** `docs/api_rag_endpoints.md` for complete RAG API documentation

---

### Agentic RAG Retrieval

Intelligent adaptive retrieval with query understanding.

**Endpoint:** `POST /api/rag/retrieve`

**Request Body:**
```json
{
  "query": "How to implement authentication?",
  "max_results": 10,
  "min_quality_score": 0.6,
  "use_personalization": true
}
```

**Response:** `200 OK`
```json
{
  "results": [
    {
      "context_id": "context-uuid",
      "content": "...",
      "semantic_score": 0.85,
      "final_score": 0.87,
      "rank_change": 2
    }
  ],
  "query_intent": "procedural",
  "strategy_used": "hybrid",
  "strategy_reasoning": "How-to queries need both semantic and keyword precision",
  "quality_score": 0.82,
  "iteration_count": 1,
  "personalized": true,
  "processing_time_ms": 245
}
```

---

## Multi-Modal Search

### Upload Image (Base64)

Upload an image using base64 encoding.

**Endpoint:** `POST /api/images/upload`

**Request Body:**
```json
{
  "image_data": "data:image/png;base64,iVBORw0KGgoAAAANS...",
  "filename": "diagram.png",
  "description": "System architecture diagram"
}
```

**Response:** `201 Created`
```json
{
  "id": "image-uuid",
  "user_id": "user-123",
  "filename": "diagram.png",
  "mime_type": "image/png",
  "size_bytes": 45678,
  "description": "System architecture diagram",
  "storage_path": "/uploads/user-123/diagram.png",
  "embedding_generated": true,
  "created_at": "2026-01-06T10:00:00Z"
}
```

---

### Upload Image (Multipart)

Upload an image using multipart form data.

**Endpoint:** `POST /api/images/upload-file`

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `file` - Image file
- `description` (optional) - Image description

**Response:** `201 Created` (Same as base64 upload)

**cURL Example:**
```bash
curl -X POST http://localhost:8080/api/images/upload-file \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@/path/to/image.png" \
  -F "description=Architecture diagram"
```

---

### Search Images by Text

Find images using text description (cross-modal search).

**Endpoint:** `POST /api/search/images-by-text`

**Request Body:**
```json
{
  "query": "database schema diagram",
  "max_results": 10
}
```

**Response:** `200 OK`
```json
{
  "query": "database schema diagram",
  "results": [
    {
      "id": "image-uuid",
      "filename": "schema.png",
      "description": "Database schema for users",
      "similarity_score": 0.87,
      "image_url": "/api/images/image-uuid/data",
      "created_at": "2026-01-05T14:30:00Z"
    }
  ],
  "count": 5
}
```

---

### Search Similar Images

Find similar images using an image query.

**Endpoint:** `POST /api/search/similar-images`

**Request Body:**
```json
{
  "image_id": "source-image-uuid",
  "max_results": 10
}
```

**Response:** `200 OK` (Similar to images-by-text)

---

### Multi-Modal Search

Search using both text and image.

**Endpoint:** `POST /api/search/multimodal`

**Request Body:**
```json
{
  "text_query": "authentication flow",
  "image_id": "reference-image-uuid",
  "text_weight": 0.5,
  "image_weight": 0.5,
  "max_results": 10
}
```

**Response:** `200 OK`
```json
{
  "query": "authentication flow + image",
  "results": [
    {
      "type": "text",
      "content": "...",
      "text_score": 0.85,
      "image_score": 0.0,
      "combined_score": 0.425
    },
    {
      "type": "image",
      "filename": "auth-flow.png",
      "text_score": 0.72,
      "image_score": 0.91,
      "combined_score": 0.815
    }
  ],
  "count": 15
}
```

---

## Error Codes

All endpoints return standard HTTP status codes and error responses.

### Success Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `204 No Content` - Request successful, no content to return

### Client Error Codes

- `400 Bad Request` - Invalid request body or parameters
- `401 Unauthorized` - Missing or invalid authentication token
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., duplicate slug)
- `422 Unprocessable Entity` - Validation error

### Server Error Codes

- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service not initialized

### Error Response Format

```json
{
  "error": "Error message here",
  "code": "ERROR_CODE",
  "details": {
    "field": "Additional context"
  }
}
```

### Common Error Examples

**401 Unauthorized:**
```json
{
  "error": "Unauthorized: missing or invalid token"
}
```

**403 Forbidden:**
```json
{
  "error": "Forbidden: insufficient permissions",
  "code": "PERMISSION_DENIED",
  "details": {
    "required_permission": "manage_members",
    "user_role": "viewer"
  }
}
```

**404 Not Found:**
```json
{
  "error": "Workspace not found",
  "code": "WORKSPACE_NOT_FOUND"
}
```

**422 Validation Error:**
```json
{
  "error": "Validation failed",
  "code": "VALIDATION_ERROR",
  "details": {
    "slug": "Slug must be alphanumeric with hyphens only",
    "name": "Name is required"
  }
}
```

---

## Rate Limiting

Currently no rate limiting is implemented. Future versions will include:

- 100 requests/minute per user for standard endpoints
- 20 requests/minute for RAG/search endpoints
- 10 requests/minute for image upload endpoints

---

## Pagination

List endpoints support pagination via query parameters:

- `limit` - Maximum number of results (default: 50, max: 100)
- `offset` - Number of results to skip (default: 0)

**Example:**
```
GET /api/workspaces/:id/members?limit=20&offset=40
```

**Response includes total:**
```json
{
  "members": [...],
  "total": 156,
  "limit": 20,
  "offset": 40
}
```

---

## Best Practices

### Authentication
- Always include Bearer token in Authorization header
- Token expires after 1 hour, refresh before expiration
- Store token securely, never in localStorage

### Error Handling
- Check HTTP status code first
- Parse error response for details
- Implement retry logic for 503 errors
- Show user-friendly messages for 403/404 errors

### Performance
- Use pagination for large result sets
- Cache workspace/role data in frontend
- Implement debouncing for search endpoints
- Use multipart upload for large images

### Security
- Validate all input on client side
- Don't expose sensitive data in logs
- Use HTTPS in production
- Implement CSRF protection for mutations

---

## Postman Collection

A Postman collection with all endpoints and example requests is available at:
`docs/postman/BusinessOS_API_Collection.json`

Import instructions:
1. Open Postman
2. File > Import
3. Select the JSON file
4. Configure environment variables:
   - `base_url`: http://localhost:8080
   - `token`: Your Supabase JWT token
   - `workspace_id`: Test workspace ID

---

## Changelog

**v1.0.0 (2026-01-06)**
- Initial API documentation
- 21 workspace endpoints
- 7 memory endpoints
- 5 project member endpoints
- 10+ RAG/search endpoints
- 9 multi-modal endpoints

---

**Document Version:** 1.0.0
**Last Updated:** January 6, 2026
**Maintainer:** Development Team
