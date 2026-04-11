/**
 * Agent Session Service
 * Creates a hidden terminal session that runs a CLI agent (Claude Code, Codex, etc.)
 * and pipes messages/output through callbacks for the dock chat.
 */

import {
  createTerminalService,
  type TerminalService,
} from "./terminal.service";

export type AgentRuntime = "claude" | "codex" | "ollama" | "hermes";

const AGENT_COMMANDS: Record<AgentRuntime, string> = {
  claude: "claude --dangerously-skip-permissions",
  codex: "codex --full-auto",
  ollama: "ollama run",
  hermes: "hermes",
};

// ANSI escape code stripper
const ANSI_RE = /\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])/g;

function stripAnsi(text: string): string {
  return text.replace(ANSI_RE, "");
}

export interface AgentSessionCallbacks {
  onOutput: (text: string) => void;
  onReady: () => void;
  onDisconnect: () => void;
  onError: (err: string) => void;
}

export interface AgentSession {
  /** Connect and launch the agent */
  connect(): void;
  /** Send a user message to the running agent's stdin */
  sendMessage(text: string): void;
  /** Tear down the session */
  disconnect(): void;
  /** Whether the WebSocket is connected */
  isConnected(): boolean;
  /** The runtime this session is running */
  runtime: AgentRuntime;
}

export function createAgentSession(
  runtime: AgentRuntime,
  callbacks: AgentSessionCallbacks,
): AgentSession {
  let service: TerminalService | null = null;
  let connected = false;
  let agentLaunched = false;
  // Buffer to detect when agent is ready (first prompt after launch)
  let outputBuffer = "";
  let readyTimeout: ReturnType<typeof setTimeout> | null = null;

  function handleOutput(raw: string) {
    const text = stripAnsi(raw);
    if (!text.trim()) return;

    outputBuffer += text;

    // After agent launch, wait for initial output burst to settle, then mark ready
    if (agentLaunched && !readyTimeout) {
      readyTimeout = setTimeout(() => {
        callbacks.onReady();
        readyTimeout = null;
      }, 1500);
    }

    // Forward cleaned output
    callbacks.onOutput(text);
  }

  const termService = createTerminalService(
    {
      onData: (data) => {
        handleOutput(data);
      },
      onConnect: (sessionId) => {
        connected = true;
        // Launch the agent command
        const command = AGENT_COMMANDS[runtime];
        if (command) {
          agentLaunched = true;
          termService.sendInput(command + "\n");
        }
      },
      onDisconnect: () => {
        connected = false;
        agentLaunched = false;
        if (readyTimeout) {
          clearTimeout(readyTimeout);
          readyTimeout = null;
        }
        callbacks.onDisconnect();
      },
      onError: (error) => {
        callbacks.onError(error);
      },
    },
    {
      cols: 120,
      rows: 40,
      shell: "zsh",
    },
  );

  service = termService;

  return {
    runtime,

    connect() {
      termService.connect();
    },

    sendMessage(text: string) {
      if (!connected) {
        callbacks.onError("Agent not connected");
        return;
      }
      termService.sendInput(text + "\n");
    },

    disconnect() {
      if (readyTimeout) {
        clearTimeout(readyTimeout);
        readyTimeout = null;
      }
      termService.disconnect();
      connected = false;
      agentLaunched = false;
    },

    isConnected() {
      return connected;
    },
  };
}
