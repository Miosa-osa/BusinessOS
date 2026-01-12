# Custom Agents - Metas de Implementação Completa

**Objetivo:** Implementar sistema completo de Custom Agents conforme spec `02-CUSTOM-AGENTS.md`

---

## 🎯 METAS PRINCIPAIS

### META 1: API Client & State Management
**Arquivos:**
- `frontend/src/lib/api/agents/customAgents.ts`
- `frontend/src/lib/stores/agents.ts`

**Requisitos:**
- ✅ TypeScript types completos para CustomAgent, AgentPreset
- ✅ 10 funções de API: getCustomAgents, createCustomAgent, updateCustomAgent, deleteCustomAgent, getAgentsByCategory, getAgentPresets, getAgentPreset, createFromPreset, testAgent, testSandbox
- ✅ Store Svelte com CRUD completo
- ✅ Derived stores: selectedAgent, agentsByCategory
- ✅ Loading states, error handling

---

### META 2: Agent Library Page
**Arquivos:**
- `frontend/src/routes/(app)/agents/+page.svelte`
- `frontend/src/lib/components/agents/AgentCard.svelte`

**Requisitos:**
- ✅ Grid responsivo de agentes (built-in + custom)
- ✅ Filtro por categoria (tabs)
- ✅ Busca por nome/descrição
- ✅ Botão "Create Agent"
- ✅ AgentCard com avatar, nome, descrição, categoria, usage count
- ✅ Dropdown menu: Edit, Delete
- ✅ Badge de categoria com cores diferentes
- ✅ Featured star icon para agentes destacados
- ✅ Empty state quando sem agentes
- ✅ Loading skeleton
- ✅ Hover effects e animações

---

### META 3: Agent Builder (Create/Edit)
**Arquivos:**
- `frontend/src/routes/(app)/agents/new/+page.svelte`
- `frontend/src/routes/(app)/agents/[id]/edit/+page.svelte`
- `frontend/src/lib/components/agents/AgentBuilder.svelte`

**Requisitos:**
- ✅ Seção Identity: name, display_name, description, avatar_url, category
- ✅ Seção Behavior: system_prompt, welcome_message, suggested_prompts
- ✅ Seção Configuration: model_preference, temperature slider, max_tokens
- ✅ Seção Tools: enabled_tools checkboxes, can_create_artifacts, can_execute_code, can_search_web
- ✅ Seção Access: is_public toggle, is_featured toggle
- ✅ Validação de campos obrigatórios
- ✅ Estados de loading durante save
- ✅ Mensagens de erro
- ✅ Botões Cancel e Save/Update
- ✅ Reutilizar mesmo componente para create e edit

---

### META 4: System Prompt Editor
**Arquivos:**
- `frontend/src/lib/components/agents/SystemPromptEditor.svelte`

**Requisitos:**
- ✅ Textarea grande com syntax highlighting (Monaco ou CodeMirror)
- ✅ Character counter
- ✅ Template snippets: Role Definition, Capabilities, Limitations, Tone & Style, Example Interaction
- ✅ Variable insertion: {{user_name}}, {{workspace_name}}, {{date}}, {{time}}
- ✅ Best practices sidebar com dicas
- ✅ Monospace font
- ✅ Auto-resize textarea

---

### META 5: Agent Testing & Sandbox
**Arquivos:**
- `frontend/src/lib/components/agents/AgentSandbox.svelte`
- `frontend/src/lib/components/agents/SandboxModal.svelte`

**Requisitos:**
- ✅ Split view: Editor | Chat Preview
- ✅ Input de teste de mensagem
- ✅ Botão "Test" → call /api/ai/custom-agents/:id/test
- ✅ Real-time preview de respostas
- ✅ Métricas: tokens_used, cost_usd, latency_ms
- ✅ Histórico de testes
- ✅ Modal standalone para sandbox sem salvar
- ✅ Usar endpoint /api/ai/custom-agents/sandbox
- ✅ Loading states durante teste

---

### META 6: Preset Gallery
**Arquivos:**
- `frontend/src/routes/(app)/presets/+page.svelte`
- `frontend/src/lib/components/agents/PresetCard.svelte`
- `frontend/src/lib/components/agents/PresetDetailModal.svelte`

**Requisitos:**
- ✅ Grid de preset templates
- ✅ PresetCard com icon, nome, descrição, categoria
- ✅ Botão "Use Template"
- ✅ Modal de preview com system_prompt completo
- ✅ Botão "Install & Customize" → cria agent e redireciona para edit
- ✅ Usar endpoint GET /api/ai/agents/presets
- ✅ Usar endpoint POST /api/ai/custom-agents/from-preset/:presetId
- ✅ Loading states
- ✅ Empty state se não houver presets

---

### META 7: Chat Integration
**Arquivos:**
- `frontend/src/routes/(app)/chat/+page.svelte` (UPDATE)
- `frontend/src/lib/components/agents/AgentSelector.svelte`

**Requisitos:**
- ✅ Agent selector dropdown no header do chat
- ✅ Mostrar avatar e nome do agente selecionado
- ✅ Lista de todos custom agents disponíveis
- ✅ Opção "Default Agent"
- ✅ Welcome message do agente ao selecionar
- ✅ Passar agent_id no payload de mensagens
- ✅ Reset do chat ao trocar agente
- ✅ Keyboard shortcut Cmd+Shift+A para abrir selector
- ✅ Command palette integration

---

### META 8: Agent Detail Page
**Arquivos:**
- `frontend/src/routes/(app)/agents/[id]/+page.svelte`
- `frontend/src/lib/components/agents/AgentStats.svelte`

**Requisitos:**
- ✅ Header: avatar, nome, descrição, badge categoria
- ✅ Tabs: Overview, Configuration, Testing, History
- ✅ Tab Overview: system_prompt, welcome_message, capabilities
- ✅ Tab Configuration: model, temperature, max_tokens, tools
- ✅ Tab Testing: AgentSandbox integrado
- ✅ Tab History: usage stats, last used, total messages
- ✅ Botões: Edit, Delete, Chat
- ✅ AgentStats component com métricas: usage_count, avg_tokens, total_cost
- ✅ Loading state ao carregar agente
- ✅ 404 se agente não encontrado

---

### META 9: Backend Endpoints (Verificação)
**Endpoints a consumir:**
- ✅ GET /api/ai/custom-agents
- ✅ POST /api/ai/custom-agents
- ✅ GET /api/ai/custom-agents/:id
- ✅ PUT /api/ai/custom-agents/:id
- ✅ DELETE /api/ai/custom-agents/:id
- ✅ GET /api/ai/custom-agents/category/:category
- ✅ POST /api/ai/custom-agents/:id/test
- ✅ POST /api/ai/custom-agents/sandbox
- ✅ GET /api/ai/agents/presets
- ✅ GET /api/ai/agents/presets/:id
- ✅ POST /api/ai/custom-agents/from-preset/:presetId
- ✅ GET /api/ai/agents (built-in agents)
- ✅ GET /api/ai/agents/:id

**Tarefas backend:**
- ✅ Testar todos endpoints com Postman/curl
- ✅ Verificar responses corretas
- ✅ Verificar error handling
- ✅ Verificar authentication/authorization
- ✅ Seed database com 10-20 presets

---

### META 10: Routing & Navigation
**Arquivos:**
- Rotas configuradas

**Requisitos:**
- ✅ /agents - Library
- ✅ /agents/new - Create
- ✅ /agents/[id] - Detail
- ✅ /agents/[id]/edit - Edit
- ✅ /presets - Preset gallery
- ✅ Navegação fluida entre páginas
- ✅ Breadcrumbs
- ✅ Back buttons

---

### META 11: UI/UX Polish
**Requisitos:**
- ✅ Dark mode completo
- ✅ Responsividade mobile
- ✅ Animações suaves (hover, transitions)
- ✅ Loading skeletons
- ✅ Empty states com ilustrações/mensagens
- ✅ Error states com retry buttons
- ✅ Success/error toasts
- ✅ Confirmação antes de deletar
- ✅ Keyboard shortcuts documentados
- ✅ Accessibility (a11y): ARIA labels, tab navigation

---

### META 12: Testing
**Arquivos:**
- `frontend/src/lib/api/agents/customAgents.test.ts`
- `frontend/src/lib/stores/agents.test.ts`
- `frontend/src/lib/components/agents/*.test.ts`

**Requisitos:**
- ✅ Unit tests: API client (todas funções)
- ✅ Unit tests: Store (CRUD actions)
- ✅ Component tests: AgentCard, AgentBuilder, SystemPromptEditor
- ✅ E2E: Criar agente do zero
- ✅ E2E: Criar agente de preset
- ✅ E2E: Editar agente
- ✅ E2E: Deletar agente
- ✅ E2E: Testar agente em sandbox
- ✅ E2E: Usar agente no chat
- ✅ Coverage target: 85%+

---

### META 13: Documentation
**Arquivos:**
- `docs/features/CUSTOM_AGENTS.md`
- `docs/api/AGENTS_API.md`

**Requisitos:**
- ✅ User guide: Como criar custom agent
- ✅ User guide: Como usar presets
- ✅ User guide: Como testar agente
- ✅ Developer docs: API client usage
- ✅ Developer docs: Store usage
- ✅ Developer docs: Component props
- ✅ Screenshots de cada página
- ✅ Video demo (opcional)

---

### META 14: Performance
**Requisitos:**
- ✅ Agent library: < 2s load com 100+ agentes
- ✅ Agent builder: < 500ms render
- ✅ Testing sandbox: < 3s response
- ✅ Chat agent selector: < 200ms render
- ✅ Lazy loading de componentes pesados
- ✅ Virtualized list se > 50 agentes
- ✅ Debounce na busca
- ✅ Cache de agentes no store (5 min)

---

### META 15: Integration com Features Existentes
**Requisitos:**
- ✅ Workspace integration: agentes por workspace
- ✅ Memory integration: agentes usam workspace memories
- ✅ MCP tools: agentes podem usar tools configurados
- ✅ Artifact system: agentes com can_create_artifacts
- ✅ Code execution: agentes com can_execute_code
- ✅ Web search: agentes com can_search_web
- ✅ Focus mode: agentes têm focus mode preference
- ✅ Chat history: salvar agent_id em mensagens

---

## 📋 CHECKLIST GERAL

### Infraestrutura
- [ ] API client completo (customAgents.ts)
- [ ] Store completo (agents.ts)
- [ ] TypeScript types
- [ ] Rotas configuradas

### Páginas
- [ ] /agents (library)
- [ ] /agents/new
- [ ] /agents/[id]
- [ ] /agents/[id]/edit
- [ ] /presets

### Componentes Core
- [ ] AgentCard
- [ ] AgentBuilder
- [ ] SystemPromptEditor
- [ ] AgentSandbox
- [ ] SandboxModal
- [ ] AgentSelector

### Componentes Presets
- [ ] PresetCard
- [ ] PresetDetailModal

### Componentes Stats
- [ ] AgentStats

### Backend
- [ ] Todos 13 endpoints testados
- [ ] Database seed com presets
- [ ] Error handling
- [ ] Authentication

### Testing
- [ ] Unit tests (API, Store)
- [ ] Component tests
- [ ] E2E tests
- [ ] 85%+ coverage

### UX/UI
- [ ] Dark mode
- [ ] Responsividade
- [ ] Animações
- [ ] Loading states
- [ ] Error states
- [ ] Empty states
- [ ] Toasts
- [ ] Accessibility

### Performance
- [ ] < 2s load times
- [ ] Lazy loading
- [ ] Virtualization
- [ ] Cache strategy

### Integration
- [ ] Workspace
- [ ] Memory
- [ ] MCP Tools
- [ ] Artifacts
- [ ] Code execution
- [ ] Web search

### Documentation
- [ ] User guide
- [ ] Developer docs
- [ ] API docs
- [ ] Screenshots

---

## 🎯 CRITÉRIO DE SUCESSO

**Funcional:**
- ✅ 15/15 endpoints integrados (100%)
- ✅ Todas páginas funcionando
- ✅ CRUD completo de agentes
- ✅ Testing sandbox funcionando
- ✅ Presets instaláveis
- ✅ Chat integration completa

**Qualidade:**
- ✅ Zero bugs críticos
- ✅ 85%+ test coverage
- ✅ Performance targets atingidos
- ✅ Acessibilidade compliant
- ✅ Dark mode perfeito
- ✅ Mobile responsive

**Documentação:**
- ✅ User guide completo
- ✅ Developer docs completos
- ✅ Screenshots de todas features

---

**PRONTO PARA IMPLEMENTAR.**
