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

| Feature | Description |
|---------|-------------|
| **Projects** | Track work across your business |
| **Tasks** | Kanban boards, assignments, due dates |
| **Team** | Org chart, capacity planning, workload management |
| **AI Chat** | Chat with local AI models, on-device |
| **Artifacts** | Generate code, docs, and more |
| **Contexts** | Store your business knowledge for AI |
| **Daily Log** | Track your day and patterns |
| **Dashboard** | Your daily command center |
| **Search** | Find anything instantly |

---

## Tech Stack

### Frontend
- **Framework**: SvelteKit
- **Styling**: TailwindCSS
- **Language**: TypeScript

### Backend
- **Framework**: FastAPI (Python)
- **Database**: PostgreSQL
- **AI**: Ollama (local LLMs)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      YOUR SERVERS                           │
│  ┌───────────┐  ┌───────────┐  ┌───────────┐              │
│  │   Data    │  │   LLMs    │  │  Agents   │              │
│  └───────────┘  └───────────┘  └───────────┘              │
│                                                             │
│          🔒 Nothing leaves without your permission          │
└─────────────────────────────────────────────────────────────┘
                            ↕
                     [YOU CONTROL]
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                   EXTERNAL (Optional)                       │
│         Cloud LLMs  •  APIs  •  Integrations               │
└─────────────────────────────────────────────────────────────┘
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
   ollama pull qwen2.5-coder:7b  # or your preferred model
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

## AI & LLM Support

Business OS is designed to work with **local LLMs by default** via Ollama:

- **Qwen** — Great for coding and general tasks
- **Llama** — Meta's open models
- **Mistral** — Fast and capable
- **DeepSeek** — Strong reasoning

You can also connect cloud providers if you choose — the decision is yours.

### MCP Integration

Business OS supports the **Model Context Protocol (MCP)** for secure agent-to-tool connections. Connect any tools you want via AMCP (Advanced Model Context Protocol).

---

## Data Privacy

- **Self-hosted** — Runs entirely on your servers
- **Local LLMs** — Train and use AI without Big Tech watching
- **You decide** — Choose what connects externally
- **No data harvesting** — Your business data stays yours

---

## Project Structure

```
BusinessOS/
├── backend/
│   ├── app/
│   │   ├── api/          # API routes
│   │   ├── models/       # Database models
│   │   ├── services/     # Business logic
│   │   └── main.py       # FastAPI app
│   └── requirements.txt
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── components/   # Svelte components
│   │   │   └── stores/       # State management
│   │   └── routes/           # SvelteKit pages
│   ├── package.json
│   └── svelte.config.js
└── README.md
```

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
