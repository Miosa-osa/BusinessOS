# Business OS

**Your business operating system for the agentic era.**

AI-native. Self-hosted. Built for fast software.

---

## Overview

Business OS is a foundational operating system template for the agentic era. Built for fast software creation where you own your data, control your AI, and customize everything.

### Key Principles

- **The Agentic Era** — AI agents that work FOR you, connecting your tools and operating on your terms
- **Fast Software** — Build and customize faster than ever. Ship changes in hours, not months
- **Your Data, Your Control** — Self-hosted by default. Nothing leaves without your permission

---

## Features

### Core Modules

| Module | Description |
|--------|-------------|
| **Dashboard** | Your daily command center with widgets for tasks, projects, and activity |
| **Projects** | Track work across your business with status, deadlines, and team assignments |
| **Tasks** | Kanban boards, list views, calendar views with priorities and due dates |
| **Team** | Org chart, capacity planning, workload management, and member profiles |
| **AI Chat** | Chat with local AI models using the Orchestrator agent system |
| **Contexts** | Store business knowledge (people, businesses, projects) that AI can reference |
| **Nodes** | Hierarchical business structure - your cognitive operating system |
| **Daily Log** | Track your day, patterns, and reflections |
| **Artifacts** | AI-generated documents: proposals, SOPs, frameworks, reports, code |
| **Settings** | Configure models, preferences, and integrations |

### Nodes System

Nodes are the core organizational structure of Business OS — a hierarchical system to manage different areas of your business:

- **Node Types**: Business, Project, Learning, Operational
- **Health Tracking**: Healthy, Needs Attention, Critical, Not Started
- **Features**: Purpose, Current Status, Weekly Focus, Decision Queue, Delegation Ready
- **Views**: Tree, List, Grid
- **Activation**: Set an active node to focus AI context on that area

### AI Agent System

Business OS includes a multi-agent architecture:

| Agent | Role |
|-------|------|
| **Orchestrator** | Main coordinator that handles requests and delegates to sub-agents |
| **DocumentAgent** | Creates business documents: proposals, SOPs, frameworks, agendas, reports |
| **AnalysisAgent** | Analyzes data, situations, and provides insights |
| **PlanningAgent** | Helps with planning, prioritization, and strategy |

### Context-Aware AI

The AI system uses specialized prompts for different scenarios:
- **Default** — General business operations assistant
- **Daily Planning** — Executive productivity and prioritization
- **Project Analysis** — Senior PM-level project assessment
- **Strategic Thinking** — High-stakes business decisions
- **Code Review** — Senior architect code analysis
- **Document Creation** — Professional business writing

### Artifacts

AI-generated content with versioning:
- **Business Documents**: Proposals, SOPs, Frameworks, Agendas, Reports, Plans
- **Code**: Code snippets, React components, HTML, SVG
- **Other**: Markdown, general documents

---

## Tech Stack

### Frontend
- **Framework**: SvelteKit 2.0
- **Styling**: TailwindCSS
- **Language**: TypeScript
- **State**: Svelte 5 Runes (`$state`, `$derived`, `$effect`)
- **Auth**: Better Auth

### Backend
- **Framework**: FastAPI (Python)
- **Database**: PostgreSQL with SQLAlchemy (async)
- **AI**: Ollama (local) or cloud LLMs
- **Auth**: Better Auth integration

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      YOUR SERVERS                           │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  PostgreSQL │  │   Ollama    │  │   Agents    │        │
│  │  (Data)     │  │   (LLMs)    │  │ (Orchestra) │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│                                                             │
│          🔒 Nothing leaves without your permission          │
└─────────────────────────────────────────────────────────────┘
                            ↕
                     [YOU CONTROL]
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                   EXTERNAL (Optional)                       │
│         Cloud LLMs  •  MCP Servers  •  Integrations        │
└─────────────────────────────────────────────────────────────┘
```

---

## Project Structure

```
BusinessOS/
├── backend/
│   ├── app/
│   │   ├── agents/           # AI agent system
│   │   │   ├── orchestrator.py   # Main coordinator
│   │   │   ├── document_agent.py # Document creation
│   │   │   ├── analysis_agent.py # Data analysis
│   │   │   └── planning_agent.py # Planning & strategy
│   │   ├── models/           # SQLAlchemy models
│   │   │   ├── node.py           # Nodes system
│   │   │   ├── context.py        # Context profiles
│   │   │   ├── artifact.py       # Generated artifacts
│   │   │   ├── conversation.py   # Chat conversations
│   │   │   ├── project.py        # Projects
│   │   │   ├── task.py           # Tasks
│   │   │   ├── team_member.py    # Team members
│   │   │   └── daily_log.py      # Daily logs
│   │   ├── routers/          # API endpoints
│   │   ├── services/         # Business logic
│   │   │   └── ollama.py         # LLM service
│   │   ├── prompts/          # System prompts
│   │   └── main.py           # FastAPI app
│   └── requirements.txt
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/   # Svelte components
│   │   │   │   ├── chat/         # Chat UI
│   │   │   │   ├── dashboard/    # Dashboard widgets
│   │   │   │   ├── projects/     # Project components
│   │   │   │   ├── tasks/        # Task components
│   │   │   │   └── team/         # Team components
│   │   │   ├── stores/       # State management
│   │   │   └── api/          # API client
│   │   └── routes/
│   │       └── (app)/        # Authenticated routes
│   │           ├── dashboard/
│   │           ├── chat/
│   │           ├── projects/
│   │           ├── tasks/
│   │           ├── team/
│   │           ├── contexts/
│   │           ├── nodes/
│   │           ├── daily/
│   │           └── settings/
│   ├── package.json
│   └── svelte.config.js
└── README.md
```

---

## Getting Started

### Prerequisites

- Node.js 18+
- Python 3.11+
- PostgreSQL
- Ollama (for local LLMs)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/robertohluna/BusinessOS.git
   cd BusinessOS
   ```

2. **Setup the backend**
   ```bash
   cd backend
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   pip install -r requirements.txt
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Setup the frontend**
   ```bash
   cd frontend
   npm install
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start Ollama** (for local AI)
   ```bash
   ollama serve
   ollama pull qwen3:4b  # or your preferred model
   ```

5. **Run the application**
   ```bash
   # Terminal 1 - Backend
   cd backend
   uvicorn app.main:app --reload

   # Terminal 2 - Frontend
   cd frontend
   npm run dev
   ```

6. **Open your browser**
   ```
   http://localhost:5173
   ```

---

## LLM Configuration

Business OS supports both **local** and **cloud** LLMs:

### Local Mode (Default)
Uses Ollama running on your machine:
```env
OLLAMA_MODE=local
OLLAMA_LOCAL_URL=http://localhost:11434
DEFAULT_MODEL=qwen3:4b
```

### Cloud Mode
Connect to cloud LLM providers:
```env
OLLAMA_MODE=cloud
OLLAMA_CLOUD_URL=https://your-provider.com
OLLAMA_CLOUD_API_KEY=your-api-key
```

### Supported Models
- **Qwen** — Great for coding and general tasks
- **Llama** — Meta's open models
- **Mistral** — Fast and capable
- **DeepSeek** — Strong reasoning
- Any Ollama-compatible model

---

## MCP Integration

Business OS supports the **Model Context Protocol (MCP)** for secure agent-to-tool connections:

- Connect external tools and APIs
- Secure, controlled data flow
- Works with AMCP (Advanced Model Context Protocol)

---

## Data Privacy

- **Self-hosted** — Runs entirely on your servers
- **Local LLMs** — Use AI without Big Tech watching
- **You decide** — Choose what connects externally
- **No data harvesting** — Your business data stays yours

---

## Use Cases

- **Agencies** — Manage clients, projects, and delivery in one place
- **Startups** — Move fast with tools you control and customize
- **Consultants & Freelancers** — Your personal business command center
- **Developers** — A foundation to build on, not fight against

---

## The Fast Software Philosophy

| Traditional SaaS | Fast Software |
|------------------|---------------|
| Request a feature | Need a feature? |
| Wait 6 months | Build it today |
| Maybe get it | Ship it tonight |

- Modify anything — it's your codebase
- Extend with your own features
- Integrate your existing tools
- Ship changes in hours, not months

---

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

## License

MIT License — See [LICENSE](LICENSE) for details.

---

## Links

- **Repository**: [github.com/robertohluna/BusinessOS](https://github.com/robertohluna/BusinessOS)
- **OSA (OS Agent)**: [osa.dev](https://osa.dev)
- **AMCP**: [amcp.ai](https://amcp.ai)

---

<p align="center">
  <strong>Business OS</strong> — Your business operating system for the agentic era.
</p>
