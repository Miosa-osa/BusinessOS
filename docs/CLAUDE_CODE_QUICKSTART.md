# 🚀 Claude Code - Quick Start Guide

**Tempo estimado:** 5 minutos

---

## ⚡ Instalação Rápida

### Opção 1: Script Automatizado (Recomendado)

**Windows (Git Bash ou WSL):**
```bash
cd C:\Users\Pichau\Desktop\BusinessOS-main-dev
bash .claude/install-optimization.sh
```

**Linux/macOS:**
```bash
cd ~/BusinessOS-main-dev
bash .claude/install-optimization.sh
```

### Opção 2: Instalação Manual

Veja instruções completas em: `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`

---

## 🎯 O Que Foi Instalado?

### 4 Skills (Conhecimento Auto-Aplicado)

| Skill | Quando Ativa |
|-------|--------------|
| `go-backend-expert` | Ao trabalhar com arquivos `.go` |
| `svelte-frontend-expert` | Ao trabalhar com arquivos `.svelte` ou frontend |
| `database-migration-expert` | Ao criar migrations ou modificar schema |
| `testing-expert` | Ao escrever ou rodar testes |

### 3 Custom Agents (Especialistas)

| Agent | Para Que Serve |
|-------|----------------|
| `backend-specialist` | APIs, handlers, serviços Go |
| `frontend-specialist` | Componentes Svelte, UI |
| `migration-specialist` | Migrations PostgreSQL |

### Hooks (Automações)

- ✅ Auto-formatação de código Go com `gofmt`
- ✅ Bloqueio de edição de arquivos sensíveis (`.env`, secrets)
- ✅ Log de todos comandos bash executados
- ✅ (Opcional) Auto-run de testes após mudanças

---

## 🧪 Teste Rápido

### 1. Verificar Instalação

```bash
# Listar skills
ls .claude/skills/

# Listar agents
ls .claude/agents/

# Ver configuração
cat .claude/settings.json
```

### 2. Testar Skill (Go Backend)

Abra nova sessão do Claude Code e peça:

```
Add a new endpoint for health check in the backend
```

Claude deve automaticamente aplicar a skill `go-backend-expert` e seguir os padrões:
- Handler → Service → Repository
- Usar `slog` para logging
- Context propagation
- Proper error handling

### 3. Testar Agent (Frontend)

```
Use the frontend-specialist to create a settings modal component
```

O agent `frontend-specialist` será invocado com conhecimento especializado em Svelte 5.

### 4. Testar Hook (Auto-format)

Peça para Claude editar um arquivo Go:

```
Add a comment to the main.go file explaining the startup process
```

Após a edição, o hook `gofmt` deve formatar o arquivo automaticamente.

---

## 💡 Exemplos de Uso

### Desenvolvimento Backend

```
# Skill ativa automaticamente
"Add authentication middleware to the handlers"

# Agent explícito
"Use the backend-specialist to implement JWT token validation"

# Com migration
"Use the migration-specialist to add a users table"
```

### Desenvolvimento Frontend

```
# Skill ativa automaticamente
"Create a reusable Button component with variants"

# Agent explícito
"Have the frontend-specialist build the dashboard page"

# Com stores
"Add a store for managing user authentication state"
```

### Database

```
# Skill ativa automaticamente
"Add a column for storing user preferences"

# Agent explícito
"Use the migration-specialist to create a migration for the notifications table"

# Com sqlc
"Generate sqlc queries for the new agents table"
```

---

## 🔍 Verificar Skills Ativas

Durante uma conversa com Claude, você pode perguntar:

```
What skills are currently active?
Which agent are you using right now?
```

Claude informará quais skills/agents estão ativos.

---

## 🛠️ Comandos Úteis

### Gerenciar Skills

```bash
# Adicionar nova skill
mkdir -p .claude/skills/my-skill
nano .claude/skills/my-skill/SKILL.md

# Desabilitar skill temporariamente
mv .claude/skills/skill-name .claude/skills/_skill-name.disabled
```

### Gerenciar Agents

```bash
# Adicionar novo agent
nano .claude/agents/my-agent.md

# Ver agents disponíveis
ls .claude/agents/
```

### Gerenciar Hooks

```bash
# Editar configuração de hooks
nano .claude/settings.json

# Ver log de comandos (criado pelos hooks)
tail -f .claude/command-log.txt
```

---

## 📊 MCP Servers (Opcional)

### Adicionar PostgreSQL Access

```bash
claude mcp add --scope project --transport stdio business-os-db -- \
  npx -y @modelcontextprotocol/server-postgres \
  postgresql://localhost:5432/business_os
```

Agora você pode fazer:
```
Show me all custom agents in the database
What's the schema of the agents table?
```

### Adicionar GitHub Integration

```bash
claude mcp add --scope user --transport http github \
  https://api.githubcopilot.com/mcp/
```

Agora você pode fazer:
```
Create a PR for my changes
Review PR #123
Show open issues labeled "bug"
```

### Ver MCP Servers

```bash
# Listar servers instalados
claude mcp list

# Ver detalhes de um server
claude mcp get business-os-db

# Remover server
claude mcp remove business-os-db
```

---

## 🎓 Próximos Passos

1. ✅ **Teste as skills** - Peça para Claude trabalhar em diferentes arquivos
2. ✅ **Invoque agents explicitamente** - `Use the backend-specialist to...`
3. ⚡ **Adicione MCP servers** - PostgreSQL, GitHub, etc
4. 🔧 **Customize hooks** - Adicione suas próprias automações
5. 📚 **Leia guia completo** - `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`

---

## 🆘 Troubleshooting

### Skill não está ativando

**Problema:** Skill não é aplicada automaticamente

**Solução:**
1. Verifique o campo `description` no frontmatter do SKILL.md
2. Seja mais específico: inclua palavras-chave relevantes
3. Exemplo: Em vez de "Helps with Go", use "Go backend development with Handler→Service→Repository pattern, slog logging, when working with files in desktop/backend-go/"

### Hook não está executando

**Problema:** Hook configurado mas não executa

**Solução:**
1. Verifique sintaxe JSON em `.claude/settings.json`
2. Verifique permissões de execução: `chmod +x .claude/hooks/*.sh`
3. Teste comando manualmente
4. Use `claude --debug` para ver logs

### Agent não é encontrado

**Problema:** "Agent not found" ao invocar

**Solução:**
1. Verifique nome do arquivo: deve ser `.md` não `.txt`
2. Verifique frontmatter YAML (entre `---`)
3. Verifique campo `name` corresponde ao nome do arquivo
4. Reinicie sessão do Claude Code

---

## 📚 Documentação Adicional

- **Guia Completo:** `docs/CLAUDE_CODE_OPTIMIZATION_GUIDE.md`
- **Convenções do Projeto:** `CLAUDE.md`
- **Skills Oficiais:** https://docs.claude.ai/claude-code/skills
- **Agents Oficiais:** https://docs.claude.ai/claude-code/agents
- **MCP Servers:** https://github.com/modelcontextprotocol/servers

---

**Dica Final:** Use skills e agents consistentemente por 1 semana. Após esse período, você terá dados suficientes para ajustar e criar skills/agents personalizados para seus workflows específicos.

**Última atualização:** 2026-01-11
