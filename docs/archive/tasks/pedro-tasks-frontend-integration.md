# Pedro Tasks - Frontend Integration

## Data: 2026-01-01
## Branch: pedro-dev

---

## Resumo da Implementacao

Integracao completa do sistema Pedro Tasks (Memory, Context & Intelligence) no frontend do BusinessOS.

---

## Arquivos Criados

### 1. API Modules

#### Learning API (`/src/lib/api/learning/`)
- **types.ts** - Tipos para feedback, personalizacao, padroes detectados
- **learning.ts** - Funcoes: `recordFeedback`, `getPersonalizationProfile`, `updatePersonalizationProfile`, `refreshProfile`, `detectPatterns`, `getLearnings`, `applyLearning`, `observeBehavior`
- **index.ts** - Re-exporta tipos e funcoes

#### Pedro Documents API (`/src/lib/api/pedro-documents/`)
- **types.ts** - Tipos para documentos, chunks, busca semantica
- **documents.ts** - Funcoes: `uploadDocument`, `listDocuments`, `searchDocuments`, `getRelevantChunks`, `getDocument`, `deleteDocument`, `reprocessDocument`, `getDocumentContent`
- **index.ts** - Re-exporta tipos e funcoes

#### Intelligence API (`/src/lib/api/intelligence/`)
- **types.ts** - Tipos para analise de conversa, extracao de memorias
  - `ExtractedMemoryType` (renomeado de MemoryType para evitar conflito)
  - `IntelligenceMessage` (renomeado de Message para evitar conflito)
- **intelligence.ts** - Funcoes: `analyzeConversation`, `getConversationAnalysis`, `searchConversationAnalyses` (renomeado), `extractMemoriesFromConversation`, `extractMemoriesFromVoiceNote`, `getExtractedMemories`
- **index.ts** - Re-exporta tipos e funcoes

#### App Profiles API (`/src/lib/api/app-profiles/`)
- **types.ts** - Tipos para perfis de aplicacao, tech stack, componentes
- **profiles.ts** - Funcoes: `analyzeCodebase`, `listProfiles`, `getProfile`, `refreshAppProfile` (renomeado), `deleteProfile`, `getProfileComponents`, `getProfileEndpoints`, `getProfileStructure`, `getProfileModules`, `getProfileTechStack`
- **index.ts** - Re-exporta tipos e funcoes

### 2. Stores

#### Learning Store (`/src/lib/stores/learning.ts`)
```typescript
interface LearningState {
  profile: PersonalizationProfile | null;
  patterns: DetectedPattern[];
  learnings: Learning[];
  loading: boolean;
  error: string | null;
}

// Funcoes do store:
- loadProfile()
- updateProfile(data)
- recordFeedback(input)
- detectPatterns()
- loadLearnings(agentType?, limit?)
```

### 3. UI Components

#### MemoryCard (`/src/lib/components/chat/MemoryCard.svelte`)
- Card para exibir memoria individual
- Suporta pin/unpin
- Mostra tipo, importancia, tags, entidades
- Acoes: editar, deletar

#### MemoryPanel (`/src/lib/components/chat/MemoryPanel.svelte`)
- Painel lateral para gerenciar memorias
- Busca semantica
- Filtros por tipo
- Ordenacao por relevancia/data
- Integracao com API de memoria

#### DocumentUploadModal (`/src/lib/components/chat/DocumentUploadModal.svelte`)
- Modal para upload de documentos RAG
- Drag & drop
- Selecao via file browser
- Progress bar
- Suporta metadados (titulo, descricao, tags, projeto, node)
- Formatos: PDF, TXT, MD, DOCX

### 4. Modificacoes

#### MessageActions.svelte
- Integrado com Learning API
- Botoes thumbs up/down agora salvam feedback no backend
- Estado visual de feedback dado
- Loading state durante envio

#### Settings Page (+page.svelte)
- Nova tab "Personalization"
- Exibe perfil de personalizacao do usuario
- Permite ajustar: tom, verbosidade, formato, preferencias
- Mostra padroes detectados pelo sistema
- Botao para detectar novos padroes

#### API Index (api/index.ts)
- Exporta novos modulos: learning, pedro-documents, intelligence, app-profiles

---

## Conflitos Resolvidos

Durante a integracao, houve conflitos de export entre modulos:

| Original | Renomeado Para | Motivo |
|----------|----------------|--------|
| `Message` (intelligence) | `IntelligenceMessage` | Conflito com conversations |
| `MemoryType` (intelligence) | `ExtractedMemoryType` | Conflito com memory |
| `searchConversations` (intelligence) | `searchConversationAnalyses` | Conflito com conversations |
| `refreshProfile` (app-profiles) | `refreshAppProfile` | Conflito com learning |

---

## Endpoints Backend Utilizados

### Learning API
- `POST /api/learning/feedback` - Registrar feedback
- `POST /api/learning/behavior` - Observar comportamento
- `GET /api/learning/profile` - Obter perfil
- `PUT /api/learning/profile` - Atualizar perfil
- `POST /api/learning/profile/refresh` - Atualizar de padroes
- `GET /api/learning/patterns` - Detectar padroes
- `GET /api/learning/learnings` - Listar aprendizados
- `POST /api/learning/learnings/:id/apply` - Aplicar aprendizado

### Documents API
- `POST /api/documents` - Upload documento (multipart)
- `GET /api/documents` - Listar documentos
- `POST /api/documents/search` - Busca semantica
- `POST /api/documents/chunks` - Chunks relevantes
- `GET /api/documents/:id` - Obter documento
- `DELETE /api/documents/:id` - Deletar documento
- `POST /api/documents/:id/reprocess` - Reprocessar
- `GET /api/documents/:id/content` - Conteudo completo

### Intelligence API
- `POST /api/intelligence/analyze` - Analisar conversa
- `GET /api/intelligence/conversations/:id` - Obter analise
- `GET /api/intelligence/conversations/search` - Buscar analises
- `POST /api/intelligence/extract/conversation` - Extrair memorias
- `POST /api/intelligence/extract/voice-note` - Extrair de audio
- `GET /api/intelligence/memories` - Listar memorias extraidas

### App Profiles API
- `POST /api/app-profiles` - Analisar codebase
- `GET /api/app-profiles` - Listar perfis
- `GET /api/app-profiles/:name` - Obter perfil
- `POST /api/app-profiles/:name/refresh` - Atualizar perfil
- `DELETE /api/app-profiles/:name` - Deletar perfil
- `GET /api/app-profiles/:name/components` - Componentes
- `GET /api/app-profiles/:name/endpoints` - Endpoints
- `GET /api/app-profiles/:name/structure` - Estrutura
- `GET /api/app-profiles/:name/modules` - Modulos
- `GET /api/app-profiles/:name/tech-stack` - Tech stack

---

## Status da Verificacao

```
npm run check:
- Errors: 9 (pre-existentes, nao relacionados a integracao)
- Warnings: 314 (a11y, CSS, etc - pre-existentes)
- Arquivos criados: Compilam sem erros
- Conflitos de export: Resolvidos
```

---

## Proximos Passos

1. **Integrar MemoryPanel no ContextPanel** - Adicionar tab de memorias
2. **Integrar DocumentUploadModal no ChatInput** - Botao de upload
3. **Testar fluxo completo** com backend rodando
4. **Adicionar testes unitarios** para stores e APIs

---

## Padroes Utilizados

### API Pattern
```typescript
// Todas as APIs seguem o padrao:
import { request } from '../base';
import type { ... } from './types';

export async function nomeDaFuncao(params): Promise<TipoRetorno> {
  return request<TipoRetorno>('/endpoint', {
    method: 'POST',
    body: JSON.stringify(params)
  });
}
```

### Store Pattern (Svelte)
```typescript
import { writable } from 'svelte/store';

function createNomeStore() {
  const { subscribe, update } = writable<State>(initialState);

  return {
    subscribe,
    async action() {
      update(s => ({ ...s, loading: true }));
      try {
        const result = await apiCall();
        update(s => ({ ...s, data: result, loading: false }));
      } catch (error) {
        update(s => ({ ...s, error: error.message, loading: false }));
      }
    }
  };
}

export const nomeStore = createNomeStore();
```

### Component Pattern (Svelte 5)
```svelte
<script lang="ts">
  import { ... } from '$lib/api/...';

  interface Props {
    prop1: string;
    onAction?: () => void;
  }

  let { prop1, onAction }: Props = $props();
  let state = $state<Type>(initial);

  async function handleAction() { ... }
</script>

<div class="component">
  <!-- markup -->
</div>

<style>
  .component { ... }
</style>
```

---

## Decisoes Arquiteturais

1. **Modulos separados por dominio** - Cada feature tem seu proprio diretorio em `/api/`
2. **Types em arquivo separado** - Facilita importacao e evita dependencias circulares
3. **Store centralizado para Learning** - Estado de personalizacao compartilhado
4. **Renomeacao de exports conflitantes** - Evita quebra de imports existentes
5. **Componentes em `/chat/`** - Memorias e documentos sao relacionados ao chat

---

## Arquivos de Referencia

- Plan original: `C:\Users\Pichau\.claude\plans\golden-spinning-wind.md`
- Backend handlers: `desktop/backend-go/internal/handler/`
- Frontend API base: `frontend/src/lib/api/base.ts`
