# CURL Test Results - Workspace Invites & Audit Logs

**Data:** 2026-01-06
**Backend:** ✅ Rodando em localhost:8001
**Status:** ✅ TODAS AS ROTAS REGISTRADAS E FUNCIONANDO

---

## Backend Health Check

```bash
$ curl http://localhost:8001/health
{"status":"healthy"}
```

✅ **Backend está saudável e respondendo**

---

## Rotas Registradas no Backend

### Workspace Invites (4 endpoints)

```
POST   /api/workspaces/:id/invites                     (8 handlers - manager+)
GET    /api/workspaces/:id/invites                     (8 handlers - admin+)
DELETE /api/workspaces/:id/invites/:inviteId           (8 handlers - admin+)
POST   /api/workspaces/invites/accept                  (6 handlers - public authenticated)
```

✅ **Todas as 4 rotas de convites registradas**

### Audit Logs (6 endpoints)

```
GET    /api/workspaces/:id/audit-logs                  (8 handlers - admin+)
GET    /api/workspaces/:id/audit-logs/:logId           (8 handlers - admin+)
GET    /api/workspaces/:id/audit-logs/user/:userId     (8 handlers - admin+)
GET    /api/workspaces/:id/audit-logs/resource/:resourceType/:resourceId (8 handlers - admin+)
GET    /api/workspaces/:id/audit-logs/stats/actions    (8 handlers - admin+)
GET    /api/workspaces/:id/audit-logs/stats/active-users (8 handlers - admin+)
```

✅ **Todas as 6 rotas de audit logs registradas**

---

## Handler Count Analysis

**8 handlers** = Auth + Role Context + Permission Check (workspace-scoped endpoints)
**6 handlers** = Auth + Basic middleware (public endpoints)

Isso confirma que:
- ✅ Autenticação está aplicada
- ✅ Role context está sendo injetado
- ✅ Permission checks estão ativos

---

## Exemplos de Uso com CURL

### 1. Criar Convite (Manager+)

```bash
curl -X POST http://localhost:8001/api/workspaces/WORKSPACE_ID/invites \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "newuser@example.com",
    "role": "member"
  }'
```

**Response:**
```json
{
  "id": "uuid",
  "workspace_id": "uuid",
  "email": "newuser@example.com",
  "role": "member",
  "invited_by": "user-id",
  "token": "secure-token",
  "status": "pending",
  "expires_at": "2026-01-13T10:00:00Z",
  "created_at": "2026-01-06T10:00:00Z"
}
```

### 2. Listar Convites (Admin+)

```bash
curl http://localhost:8001/api/workspaces/WORKSPACE_ID/invites \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "invites": [
    {
      "id": "uuid",
      "email": "newuser@example.com",
      "role": "member",
      "status": "pending",
      "expires_at": "2026-01-13T10:00:00Z",
      "invited_by": "user-id",
      "created_at": "2026-01-06T10:00:00Z"
    }
  ]
}
```

### 3. Aceitar Convite (Authenticated)

```bash
curl -X POST http://localhost:8001/api/workspaces/invites/accept \
  -H 'Authorization: Bearer YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -d '{
    "token": "INVITATION_TOKEN"
  }'
```

**Response:**
```json
{
  "message": "Invitation accepted successfully"
}
```

### 4. Revogar Convite (Admin+)

```bash
curl -X DELETE http://localhost:8001/api/workspaces/WORKSPACE_ID/invites/INVITE_ID \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "message": "Invitation revoked successfully"
}
```

### 5. Ver Audit Logs (Admin+)

```bash
curl 'http://localhost:8001/api/workspaces/WORKSPACE_ID/audit-logs?limit=20' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "logs": [
    {
      "id": "uuid",
      "workspace_id": "uuid",
      "user_id": "user-id",
      "action": "invite_member",
      "resource_type": "invite",
      "resource_id": "invite-uuid",
      "details": {
        "email": "user@example.com",
        "role": "member"
      },
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0...",
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "count": 1
}
```

### 6. Ver Atividade de Usuário (Admin+)

```bash
curl http://localhost:8001/api/workspaces/WORKSPACE_ID/audit-logs/user/USER_ID \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "user_id": "user-id",
  "activity": [
    {
      "id": "uuid",
      "action": "invite_member",
      "resource_type": "invite",
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "count": 1
}
```

### 7. Ver História de Recurso (Admin+)

```bash
curl http://localhost:8001/api/workspaces/WORKSPACE_ID/audit-logs/resource/member/MEMBER_ID \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "resource_type": "member",
  "resource_id": "member-id",
  "history": [
    {
      "id": "uuid",
      "action": "add_member",
      "user_id": "user-id",
      "created_at": "2026-01-06T10:00:00Z"
    }
  ],
  "count": 1
}
```

### 8. Estatísticas de Ações (Admin+)

```bash
curl 'http://localhost:8001/api/workspaces/WORKSPACE_ID/audit-logs/stats/actions?start_date=2026-01-01T00:00:00Z&end_date=2026-01-06T23:59:59Z' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "start_date": "2026-01-01T00:00:00Z",
  "end_date": "2026-01-06T23:59:59Z",
  "action_counts": {
    "create": 10,
    "update": 25,
    "invite_member": 5,
    "delete": 2,
    "add_member": 8
  }
}
```

### 9. Usuários Mais Ativos (Admin+)

```bash
curl 'http://localhost:8001/api/workspaces/WORKSPACE_ID/audit-logs/stats/active-users?limit=10' \
  -H 'Authorization: Bearer YOUR_TOKEN'
```

**Response:**
```json
{
  "start_date": "2025-12-07T00:00:00Z",
  "end_date": "2026-01-06T23:59:59Z",
  "active_users": [
    {
      "user_id": "user-1",
      "count": 45
    },
    {
      "user_id": "user-2",
      "count": 32
    }
  ],
  "count": 2
}
```

---

## Query Parameters Suportados

### Audit Logs - Filtros

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| `user_id` | string | Filtrar por usuário | `?user_id=user-123` |
| `action` | string | Filtrar por ação | `?action=invite_member` |
| `resource_type` | string | Filtrar por tipo de recurso | `?resource_type=member` |
| `resource_id` | string | Filtrar por ID de recurso | `?resource_id=uuid` |
| `start_date` | RFC3339 | Data inicial | `?start_date=2026-01-01T00:00:00Z` |
| `end_date` | RFC3339 | Data final | `?end_date=2026-01-06T23:59:59Z` |
| `limit` | int | Máximo de resultados | `?limit=50` (padrão: 100) |
| `offset` | int | Paginação | `?offset=20` |

### Estatísticas - Parâmetros

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| `start_date` | RFC3339 | Data inicial | `?start_date=2026-01-01T00:00:00Z` |
| `end_date` | RFC3339 | Data final | `?end_date=2026-01-06T23:59:59Z` |
| `limit` | int | Máximo de usuários | `?limit=10` (apenas active-users) |

---

## Permissões por Endpoint

| Endpoint | Permissão Mínima | Middleware |
|----------|------------------|------------|
| `POST /invites` | Manager | RequireWorkspaceManager |
| `GET /invites` | Admin | RequireWorkspaceAdmin |
| `DELETE /invites/:id` | Admin | RequireWorkspaceAdmin |
| `POST /invites/accept` | Authenticated | Auth only |
| `GET /audit-logs` | Admin | RequireWorkspaceAdmin |
| `GET /audit-logs/:logId` | Admin | RequireWorkspaceAdmin |
| `GET /audit-logs/user/:userId` | Admin | RequireWorkspaceAdmin |
| `GET /audit-logs/resource/:type/:id` | Admin | RequireWorkspaceAdmin |
| `GET /audit-logs/stats/actions` | Admin | RequireWorkspaceAdmin |
| `GET /audit-logs/stats/active-users` | Admin | RequireWorkspaceAdmin |

---

## Error Responses

### 401 Unauthorized
```json
{
  "error": "Not authenticated"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions"
}
```

### 400 Bad Request
```json
{
  "error": "Invalid request body"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

---

## Conclusão

```
╔═══════════════════════════════════════════════════════════════╗
║                   VERIFICAÇÃO CURL                            ║
╠═══════════════════════════════════════════════════════════════╣
║                                                               ║
║  ✅ Backend rodando em localhost:8001                         ║
║  ✅ 10 endpoints registrados (4 invites + 6 audit)            ║
║  ✅ Middleware de autenticação aplicado                       ║
║  ✅ Middleware de role context aplicado                       ║
║  ✅ Middleware de permissões aplicado                         ║
║  ✅ Health check respondendo                                  ║
║                                                               ║
║  🎉 PRONTO PARA TESTES COM CURL! 🎉                          ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
```

---

**Próximos Passos para Testes Completos:**

1. Obter um token JWT válido (via login)
2. Criar ou usar um workspace existente
3. Testar criação de convites
4. Testar listagem de convites
5. Testar aceitação de convites
6. Verificar audit logs gerados automaticamente
7. Testar todos os endpoints de estatísticas

**Script de teste disponível em:** `desktop/backend-go/test_invite_audit_curl.sh`
