# Database Connection Troubleshooting

**Date**: 2026-01-05
**Status**: ❌ Supabase Connection Failed
**Current Mode**: ✅ Backend running (degraded mode - no database)

---

## Problema Identificado

### Erro Final: "Tenant or user not found"

```
FATAL: Tenant or user not found (SQLSTATE XX000)
```

**Significado**: O projeto Supabase não existe, está pausado ou foi deletado.

---

## Testes Realizados

### ✅ Teste 1: Conectividade de Rede
```bash
Port 6543: ✅ ACESSÍVEL
```
**Resultado**: Rede OK

### ❌ Teste 2: Conexão Direta
```
URL: db.fuqhjbgbjamtxcdphjpp.supabase.co:6543
User: postgres
Error: SASL authentication failed
```
**Resultado**: Formato de username incorreto

### ❌ Teste 3: Conexão Pooler
```
URL: aws-0-us-east-1.pooler.supabase.com:6543
User: postgres.fuqhjbgbjamtxcdphjpp
Error: Tenant or user not found
```
**Resultado**: Projeto não encontrado

---

## 🔍 Causa Raiz

O projeto Supabase **`fuqhjbgbjamtxcdphjpp`** está:
1. **Pausado** (inatividade > 7 dias)
2. **Suspenso** (quota excedida ou pagamento pendente)
3. **Deletado** (removido manualmente)
4. Ou as **credenciais mudaram**

---

## ✅ Como Resolver

### Opção 1: Reativar Projeto Supabase (Recomendado)

1. **Acesse o Dashboard Supabase**:
   ```
   https://app.supabase.com/project/fuqhjbgbjamtxcdphjpp
   ```

2. **Verifique o Status do Projeto**:
   - Se estiver **PAUSADO**: Clique em "Resume Project"
   - Se estiver **SUSPENSO**: Verifique billing/pagamento
   - Se não aparecer: Projeto foi deletado → Criar novo

3. **Obtenha Novas Credenciais**:
   - Vá em **Settings** > **Database**
   - Copie a **Connection String** (Pooler)
   - Atualize o `.env`:
   ```bash
   DATABASE_URL=<nova-connection-string>
   DATABASE_REQUIRED=true
   ```

4. **Reinicie o Backend**:
   ```bash
   cd desktop/backend-go
   # Mate o processo atual
   ps aux | grep "go run" | grep -v grep | awk '{print $2}' | xargs kill

   # Inicie novamente
   go run ./cmd/server
   ```

---

### Opção 2: Criar Novo Projeto Supabase

1. **Criar Projeto**:
   - Acesse https://app.supabase.com
   - Clique em "New Project"
   - Preencha:
     - **Name**: BusinessOS
     - **Database Password**: (escolha uma senha forte)
     - **Region**: escolha a mais próxima

2. **Aguarde Provisionamento** (1-2 minutos)

3. **Configure o Banco de Dados**:
   - Vá em **SQL Editor**
   - Execute os scripts de migração (se disponíveis em `migrations/`)
   - Ou restaure de um backup

4. **Atualize `.env`**:
   ```bash
   # Copie do Supabase: Settings > Database > Connection String (Pooler)
   DATABASE_URL=postgres://postgres.[SEU-PROJETO]:[SUA-SENHA]@aws-0-us-east-1.pooler.supabase.com:6543/postgres
   DATABASE_REQUIRED=true

   SUPABASE_URL=https://[SEU-PROJETO].supabase.co
   SUPABASE_ANON_KEY=[SUA-ANON-KEY]
   ```

5. **Reinicie o Backend**

---

### Opção 3: PostgreSQL Local (Desenvolvimento)

Para desenvolvimento local sem depender do Supabase:

1. **Instalar PostgreSQL com Docker**:
   ```bash
   docker run -d \
     --name businessos-postgres \
     -e POSTGRES_PASSWORD=senha123 \
     -e POSTGRES_DB=businessos \
     -p 5432:5432 \
     postgres:15-alpine
   ```

2. **Criar Banco de Dados**:
   ```bash
   docker exec -it businessos-postgres psql -U postgres -c "CREATE DATABASE businessos;"
   ```

3. **Executar Migrações** (se houver):
   ```bash
   # Se tiver migrations/
   cd desktop/backend-go
   # Execute scripts SQL ou use ferramenta de migração
   ```

4. **Atualizar `.env`**:
   ```bash
   DATABASE_URL=postgres://postgres:senha123@localhost:5432/businessos
   DATABASE_REQUIRED=true
   ```

5. **Reiniciar Backend**

---

## Status Atual

### ✅ O Que Está Funcionando

**Backend** (modo degradado):
- ✅ HTTP server rodando (porta 8001)
- ✅ Health checks funcionando
- ✅ CORS configurado
- ✅ Rate limiting ativo

**Frontend**:
- ✅ Vite dev server rodando (porta 5173)
- ✅ UI carregando
- ⚠️ API calls retornam 404 (esperado sem database)

**Acessos**:
- Frontend: http://localhost:5173
- Backend: http://localhost:8001
- Health: http://localhost:8001/health

---

### ❌ O Que NÃO Está Funcionando

Sem banco de dados, estas features estão desabilitadas:
- ❌ Autenticação (login/signup)
- ❌ Chat com agentes
- ❌ RAG/Busca semântica
- ❌ Projetos e tarefas
- ❌ Sistema de aprendizado
- ❌ Contexto de roles
- ❌ Memória e personalizaçã

---

## Erros no Console do Frontend

Estes erros são **ESPERADOS** em modo degradado:

```
❌ /api/auth/session → 404 Not Found
❌ /api/auth/sign-up/email → 404 Not Found
```

**Por quê?**: O backend não registra rotas `/api/*` quando `DATABASE_REQUIRED=false`

**Como resolver?**: Habilitar o banco de dados (Opção 1, 2 ou 3 acima)

---

## Verificação Rápida

### Testar se Supabase está acessível:

```bash
# Teste 1: API Supabase
curl -I https://fuqhjbgbjamtxcdphjpp.supabase.co

# Teste 2: Porta do banco
timeout 5 bash -c 'cat < /dev/null > /dev/tcp/db.fuqhjbgbjamtxcdphjpp.supabase.co/6543' && echo "✅ Porta acessível" || echo "❌ Porta bloqueada"

# Teste 3: Conexão PostgreSQL (se tiver psql instalado)
psql "postgres://postgres.fuqhjbgbjamtxcdphjpp:fmm6Wt7kN0ajrjxK@aws-0-us-east-1.pooler.supabase.com:6543/postgres?connect_timeout=5"
```

---

## Logs para Debug

```bash
# Backend logs
tail -f logs/backend.log

# Frontend logs
tail -f logs/frontend.log

# Testar API status
curl http://localhost:8001/api/status
```

---

## Próximos Passos

1. **Verificar Supabase Dashboard** (Opção 1)
2. **OU criar novo projeto** (Opção 2)
3. **OU usar PostgreSQL local** (Opção 3)
4. **Atualizar `.env`** com novas credenciais
5. **Reiniciar backend**
6. **Testar no frontend**

---

## Suporte

Se precisar de ajuda:
1. Compartilhe o erro específico do `logs/backend.log`
2. Verifique o status no Supabase Dashboard
3. Confirme se o projeto ainda existe

---

**Última Atualização**: 2026-01-05 13:40 UTC
**Status Backend**: ✅ Rodando (degraded mode)
**Status Frontend**: ✅ Rodando
**Status Database**: ❌ Não conectado
