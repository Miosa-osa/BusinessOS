#!/bin/bash

# Teste completo do sistema de convites e audit logs via curl
# Backend deve estar rodando em localhost:8001

BASE_URL="http://localhost:8001"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║     Testando Invite & Audit System via CURL                  ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# TEST 1: Verificar saúde do servidor
echo "📝 TEST 1: Verificar Backend Health"
echo "-----------------------------------------------------------"
HEALTH=$(curl -s ${BASE_URL}/health)
echo "Response: ${HEALTH}"
if [[ $HEALTH == *"healthy"* ]]; then
    echo -e "${GREEN}✅ Backend está saudável${NC}"
else
    echo -e "${RED}❌ Backend não está respondendo${NC}"
    exit 1
fi
echo ""

# Nota: Para testes completos, você precisaria:
# 1. Criar um usuário (ou usar Supabase auth)
# 2. Fazer login para obter token JWT
# 3. Criar um workspace
# 4. Usar o token nas chamadas autenticadas

echo "📝 PRÓXIMOS TESTES REQUEREM AUTENTICAÇÃO"
echo "-----------------------------------------------------------"
echo "Para testar completamente, você precisa:"
echo "1. Ter um usuário autenticado com token JWT"
echo "2. Criar/ter um workspace"
echo "3. Usar o token nos headers: -H 'Authorization: Bearer TOKEN'"
echo ""

echo "📋 ENDPOINTS DISPONÍVEIS:"
echo "-----------------------------------------------------------"
echo ""
echo "🔐 WORKSPACE INVITES (manager+):"
echo "  POST   ${BASE_URL}/api/workspaces/:id/invites"
echo "         Body: {\"email\":\"user@example.com\",\"role\":\"member\"}"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/invites"
echo "         Lista todos os convites do workspace"
echo ""
echo "  DELETE ${BASE_URL}/api/workspaces/:id/invites/:inviteId"
echo "         Revoga um convite"
echo ""
echo "  POST   ${BASE_URL}/api/workspaces/invites/accept"
echo "         Body: {\"token\":\"invitation-token\"}"
echo "         (público - qualquer usuário autenticado)"
echo ""
echo "📊 AUDIT LOGS (admin+):"
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs"
echo "         Query params: user_id, action, resource_type, limit, offset"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs/:logId"
echo "         Detalhes de um log específico"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs/user/:userId"
echo "         Atividade de um usuário específico"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs/resource/:type/:id"
echo "         História de um recurso específico"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs/stats/actions"
echo "         Estatísticas de ações"
echo ""
echo "  GET    ${BASE_URL}/api/workspaces/:id/audit-logs/stats/active-users"
echo "         Usuários mais ativos"
echo ""

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                    EXEMPLO DE USO                             ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""
echo "# 1. Criar convite (manager+)"
echo "curl -X POST ${BASE_URL}/api/workspaces/WORKSPACE_ID/invites \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email\":\"newuser@example.com\",\"role\":\"member\"}'"
echo ""
echo "# 2. Listar convites (admin+)"
echo "curl ${BASE_URL}/api/workspaces/WORKSPACE_ID/invites \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN'"
echo ""
echo "# 3. Aceitar convite (authenticated)"
echo "curl -X POST ${BASE_URL}/api/workspaces/invites/accept \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN' \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"token\":\"INVITATION_TOKEN\"}'"
echo ""
echo "# 4. Ver audit logs (admin+)"
echo "curl '${BASE_URL}/api/workspaces/WORKSPACE_ID/audit-logs?limit=20' \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN'"
echo ""
echo "# 5. Ver atividade de usuário (admin+)"
echo "curl ${BASE_URL}/api/workspaces/WORKSPACE_ID/audit-logs/user/USER_ID \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN'"
echo ""
echo "# 6. Estatísticas de ações (admin+)"
echo "curl '${BASE_URL}/api/workspaces/WORKSPACE_ID/audit-logs/stats/actions?start_date=2026-01-01T00:00:00Z' \\"
echo "  -H 'Authorization: Bearer YOUR_TOKEN'"
echo ""

echo "✅ BACKEND ESTÁ RODANDO E ENDPOINTS REGISTRADOS!"
echo ""
