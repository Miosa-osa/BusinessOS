# Custom Agents - Guia de Teste na UI
## Como Testar o Sistema de Custom Agents no Frontend

**Data:** 2026-01-09
**Frontend URL:** http://localhost:5173
**Backend URL:** http://localhost:8001

---

## 🚀 Passo 1: Fazer Login

⚠️ **IMPORTANTE:** Você precisa estar autenticado para acessar Custom Agents!

### Como Fazer Login:

1. **Acesse a página de login:**
   ```
   http://localhost:5173/login
   ```

2. **Credenciais de teste** (se houver usuário criado):
   - Email: [seu email de teste]
   - Password: [sua senha]

3. **Ou crie uma nova conta:**
   ```
   http://localhost:5173/signup
   ```

---

## 🎯 Passo 2: Galeria de Presets (Comece Aqui!)

Esta é a melhor página para começar - mostra os 10 agent presets disponíveis.

### URL:
```
http://localhost:5173/agents/presets
```

### O Que Você Vai Ver:

**📊 Seção Featured (Destacados):**
- Agentes marcados como "featured" aparecem no topo
- Cards maiores com descrição completa

**🎨 Seção All Presets (Todos os Presets):**
- Grid com todos os 10 presets disponíveis:
  1. Code Reviewer
  2. Technical Writer
  3. Data Analyst
  4. Business Strategist
  5. Creative Writer
  6. Researcher (Core Specialist)
  7. Writer (Core Specialist)
  8. Coder (Core Specialist)
  9. Analyst (Core Specialist)
  10. Planner (Core Specialist)

**🔍 Filtros Disponíveis:**
- Search bar (busca por nome/descrição)
- Category filter (coding, writing, analysis, business, research, planning)

### Como Testar:

1. ✅ **Visualização:** Veja os cards dos presets
2. ✅ **Filtro de Categoria:** Clique em uma categoria
3. ✅ **Busca:** Digite "code" e veja filtrar
4. ✅ **Preview:** Clique em um preset para ver detalhes
5. ✅ **Criar do Preset:** Clique em "Use this Template"

**Resultado Esperado:**
- Modal abre com o preset selecionado
- Botão "Create from Template" disponível
- Ao criar, redireciona para a biblioteca

---

## 📚 Passo 3: Library (Biblioteca de Agentes)

Página principal para ver seus agentes customizados.

### URL:
```
http://localhost:5173/agents
```

### O Que Você Vai Ver:

**🎛️ Toolbar:**
- Search bar
- Category filter dropdown
- Status filter (Active / Inactive / All)
- Sort options (Name / Created / Usage)
- Button: "New Agent" (vai para /agents/new)
- Button: "Browse Presets" (vai para /agents/presets)

**📊 Stats Cards (se houver agentes):**
- Total Agents
- Active Agents
- Most Used Agent
- Recent Activity

**🗂️ Grid de Agentes:**
- Cards com:
  - Avatar/emoji
  - Nome e categoria
  - Descrição
  - Status badge (active/inactive)
  - Botões: View, Edit, Test, Delete

### Como Testar:

1. ✅ **Vazio Inicial:** Se não houver agentes, mostra empty state
2. ✅ **Criar Primeiro Agente:**
   - Clique "New Agent" ou
   - Clique "Browse Presets" → escolha um preset
3. ✅ **Filtros:**
   - Filtre por categoria
   - Filtre por status
   - Busque por nome
4. ✅ **Ordenação:**
   - Ordene por nome
   - Ordene por data de criação
   - Ordene por uso

---

## ➕ Passo 4: Criar Novo Agente

Formulário para criar um agente do zero.

### URL:
```
http://localhost:5173/agents/new
```

### O Que Você Vai Ver:

**📝 Formulário com Campos:**

**Basic Information:**
- Display Name (obrigatório)
- Name (slug, auto-gerado)
- Description
- Avatar (emoji picker)

**Configuration:**
- System Prompt (obrigatório, textarea grande)
- Model Preference (dropdown)
- Temperature (slider 0-1)
- Max Tokens (input numérico)

**Capabilities:**
- Category (dropdown: general, coding, writing, analysis, business, custom)
- Tools Enabled (multi-select checkboxes)
- Context Sources (multi-select)

**Behavior:**
- Enable Thinking (toggle)
- Enable Streaming (toggle)
- Is Active (toggle)

**Botões:**
- Save (primário)
- Cancel (volta para /agents)

### Como Testar:

1. ✅ **Preenchimento Básico:**
   ```
   Display Name: My Test Agent
   Description: Testing custom agents
   System Prompt: You are a helpful test agent.
   ```

2. ✅ **Validação:**
   - Tente salvar sem Display Name → erro
   - Tente salvar sem System Prompt → erro

3. ✅ **Configuração Avançada:**
   - Ajuste temperatura
   - Selecione modelo
   - Habilite thinking
   - Escolha ferramentas

4. ✅ **Salvar:**
   - Clique Save
   - Deve redirecionar para /agents
   - Novo agente aparece na lista

**Resultado Esperado:**
- Agente criado no backend
- Redireciona para biblioteca
- Agente aparece na lista

---

## 👁️ Passo 5: Visualizar Agente (Detail Page)

Página de detalhes do agente com tabs.

### URL:
```
http://localhost:5173/agents/[ID]
```
(Substitua [ID] pelo UUID do agente criado)

### O Que Você Vai Ver:

**🎨 Header:**
- Avatar grande
- Nome e categoria
- Status badge
- Botões: Edit, Delete, Clone, Toggle Active

**📑 Tabs:**

**1. Overview Tab:**
- Descrição completa
- System prompt (preview)
- Configurações (model, temperature, tokens)
- Capabilities e tools
- Usage stats (times used, last used)
- Created/updated dates

**2. Test Tab:**
- Agent Sandbox integrado
- Input para mensagem de teste
- Botão "Send"
- Response area (SSE streaming)
- History de testes
- Clear history button

**3. Settings Tab:**
- Formulário de edição (mesmo que /agents/[id]/edit)
- Inline editing

### Como Testar:

1. ✅ **Navegação de Tabs:**
   - Clique em cada tab
   - Veja conteúdo diferente

2. ✅ **Test Tab (Importante!):**
   ```
   Message: Hello, can you help me test?
   ```
   - Clique Send
   - Veja resposta streaming do agente
   - Status: loading → content streaming → done

3. ✅ **Toggle Active:**
   - Clique botão "Activate" / "Deactivate"
   - Badge muda
   - Status atualiza

4. ✅ **Clone:**
   - Clique "Clone"
   - Cria cópia do agente
   - Redireciona para novo agente

5. ✅ **Delete:**
   - Clique Delete
   - Mostra confirmação
   - Remove agente
   - Volta para /agents

**Resultado Esperado:**
- Todas as informações carregam
- Test funciona com SSE streaming
- Actions funcionam (toggle, clone, delete)

---

## ✏️ Passo 6: Editar Agente

Formulário de edição (similar ao create).

### URL:
```
http://localhost:5173/agents/[ID]/edit
```

### O Que Você Vai Ver:

- Mesmo formulário do "New Agent"
- Pré-preenchido com dados atuais
- Botões: Save Changes, Cancel

### Como Testar:

1. ✅ **Edição Simples:**
   - Mude description
   - Salve
   - Volte para detail page
   - Veja mudança

2. ✅ **Edição de System Prompt:**
   - Adicione mais instruções
   - Salve
   - Teste no sandbox
   - Veja comportamento diferente

3. ✅ **Mudança de Categoria:**
   - Mude categoria
   - Salve
   - Volte para library
   - Filtre pela nova categoria
   - Veja agente aparece

**Resultado Esperado:**
- Mudanças salvam no backend
- UI atualiza imediatamente
- Filtros refletem mudanças

---

## 🎯 Passo 7: Agent Sandbox (Teste Avançado)

Sandbox para testar prompts sem criar agente.

### Onde Encontrar:

**Opção 1:** Tab "Test" na detail page de qualquer agente

**Opção 2:** Sandbox standalone (se implementado):
```
http://localhost:5173/agents/sandbox
```

### O Que Você Vai Ver:

**📝 Inputs:**
- System Prompt (textarea)
- User Message (input)
- Model Selection (dropdown)
- Temperature (slider)

**🎛️ Configurações:**
- Max tokens
- Enable thinking
- Enable streaming

**💬 Response Area:**
- Streaming response
- Thinking blocks (se habilitado)
- Status indicators
- Copy button

### Como Testar:

1. ✅ **Teste Básico:**
   ```
   System Prompt: You are a helpful assistant.
   Message: What is 2+2?
   ```
   - Send
   - Veja resposta

2. ✅ **Teste com Thinking:**
   - Habilite thinking
   - Send mensagem complexa
   - Veja thinking blocks aparecerem
   - Veja resposta final

3. ✅ **Teste de Streaming:**
   - Veja texto aparecer palavra por palavra
   - Status muda: idle → loading → streaming → complete

4. ✅ **Teste de Erro:**
   - Deixe system prompt vazio
   - Tente enviar
   - Veja mensagem de erro

**Resultado Esperado:**
- SSE streaming funciona
- Thinking aparecem se habilitado
- Erros mostram mensagens claras

---

## 🧪 Fluxo de Teste Completo (20 minutos)

### Teste Rápido (5 minutos):

1. ✅ Login
2. ✅ Acesse `/agents/presets`
3. ✅ Escolha "Code Reviewer"
4. ✅ Clique "Use Template"
5. ✅ Crie agente
6. ✅ Teste no sandbox
7. ✅ Veja resposta

### Teste Médio (10 minutos):

1. ✅ Login
2. ✅ Crie agente do zero (`/agents/new`)
3. ✅ Preencha todos os campos
4. ✅ Salve
5. ✅ Vá para detail page
6. ✅ Teste todas as tabs
7. ✅ Edite o agente
8. ✅ Clone o agente
9. ✅ Delete um agente

### Teste Completo (20 minutos):

1. ✅ **Setup:**
   - Login
   - Veja biblioteca vazia

2. ✅ **Presets:**
   - Acesse galeria
   - Teste filtros
   - Crie 3 agentes de presets diferentes

3. ✅ **Custom Agent:**
   - Crie agente do zero
   - Configure completamente
   - Salve

4. ✅ **Library:**
   - Veja 4 agentes na lista
   - Teste todos os filtros
   - Teste ordenação
   - Busque por nome

5. ✅ **Detail Pages:**
   - Abra cada agente
   - Teste sandbox de cada um
   - Veja respostas diferentes

6. ✅ **Actions:**
   - Toggle active/inactive
   - Clone um agente
   - Edit um agente
   - Delete um agente

7. ✅ **Edge Cases:**
   - Tente criar agente sem nome
   - Tente criar agente com nome duplicado
   - Teste com system prompts muito longos
   - Delete todos agentes (empty state)

---

## 🎨 Componentes Implementados

### 📦 6 Componentes Criados:

1. **AgentCard.svelte** (328 linhas)
   - Display de agente individual
   - Variantes: default, compact
   - Actions: view, edit, test, delete

2. **AgentBuilder.svelte** (106 linhas)
   - Formulário de criação/edição
   - Validação de campos
   - Props aceita agent para edição

3. **SystemPromptEditor.svelte** (252 linhas)
   - Editor de system prompt
   - Templates sugeridos
   - Variables dinâmicas
   - Tips & best practices

4. **AgentSandbox.svelte** (479 linhas)
   - Teste de agentes
   - SSE streaming
   - Thinking display
   - History

5. **AgentSelector.svelte** (771 linhas)
   - Seletor de agentes
   - Grid ou list view
   - Keyboard navigation
   - Search e filters

6. **PresetCard.svelte** (99 linhas)
   - Display de preset
   - Preview modal
   - Create from template

### 🗂️ 5 Pages Criadas:

1. **`/agents/+page.svelte`** (377 linhas)
   - Library com grid
   - Filtros e busca
   - Stats cards

2. **`/agents/new/+page.svelte`** (255 linhas)
   - Formulário de criação
   - Validation
   - Templates

3. **`/agents/[id]/+page.svelte`**
   - Detail page com tabs
   - Overview, Test, Settings
   - Actions

4. **`/agents/[id]/edit/+page.svelte`**
   - Edit form
   - Pre-filled com dados

5. **`/agents/presets/+page.svelte`** (424 linhas)
   - Galeria de presets
   - Featured section
   - Modal de preview

---

## 🐛 Troubleshooting

### Problema: "Not authenticated"

**Solução:**
1. Vá para `/login`
2. Faça login
3. Cookie `better-auth.session_token` será setado
4. Recarregue a página

### Problema: Agentes não aparecem

**Solução:**
1. Abra DevTools (F12)
2. Veja Network tab
3. Verifique request para `/api/ai/custom-agents`
4. Status 401? → Faça login
5. Status 200? → Verifique response.agents array

### Problema: Sandbox não funciona

**Solução:**
1. Verifique se backend está rodando (http://localhost:8001/health)
2. Veja console do browser (F12)
3. Verifique erros de CORS
4. Teste endpoint direto:
   ```bash
   curl -X POST http://localhost:8001/api/ai/custom-agents/sandbox \
     -H "Content-Type: application/json" \
     -d '{"message":"test","system_prompt":"You are helpful"}'
   ```

### Problema: Presets não carregam

**Solução:**
1. Verifique se migrations rodaram:
   ```bash
   cd desktop/backend-go
   go run ./cmd/migrate
   ```
2. Verifique tabela agent_presets no Supabase:
   ```bash
   curl "https://fuqhjbgbjamtxcdphjpp.supabase.co/rest/v1/agent_presets?select=count" \
     -H "apikey: YOUR_KEY"
   ```

---

## 📊 Checklist de Teste

### UI/UX:
- [ ] Todas as páginas carregam sem erros
- [ ] Componentes renderizam corretamente
- [ ] Botões são clicáveis e respondem
- [ ] Modals abrem e fecham
- [ ] Forms validam inputs
- [ ] Loading states aparecem
- [ ] Error states mostram mensagens claras
- [ ] Empty states são informativos
- [ ] Dark mode funciona (se implementado)
- [ ] Responsive em mobile (opcional)

### Funcionalidades:
- [ ] Login funciona
- [ ] Presets carregam (10 total)
- [ ] Criar agente do zero funciona
- [ ] Criar agente de preset funciona
- [ ] Editar agente funciona
- [ ] Delete agente funciona
- [ ] Clone agente funciona
- [ ] Toggle active/inactive funciona
- [ ] Filtros funcionam
- [ ] Busca funciona
- [ ] Ordenação funciona
- [ ] Sandbox funciona
- [ ] SSE streaming funciona
- [ ] Thinking blocks aparecem (se habilitado)

### API/Backend:
- [ ] Todas requests retornam 200 (se autenticado)
- [ ] 401 se não autenticado (correto)
- [ ] Dados salvam no Supabase
- [ ] Dados atualizam no Supabase
- [ ] Dados deletam no Supabase
- [ ] Filtros server-side funcionam
- [ ] Endpoints de test funcionam

---

## 🎯 Próximos Passos Após Teste

1. **Se tudo funcionar:** 🎉
   - Documente bugs encontrados
   - Liste melhorias desejadas
   - Planeje próximas features

2. **Se houver erros:**
   - Anote erros específicos
   - Capture screenshots
   - Copie mensagens do console
   - Reporte para correção

3. **Feedback de UX:**
   - O que está confuso?
   - O que falta?
   - O que pode melhorar?

---

**Guia Criado:** 2026-01-09
**Frontend:** http://localhost:5173
**Backend:** http://localhost:8001
**Documentação Completa:** ✅ PRONTA PARA TESTES
