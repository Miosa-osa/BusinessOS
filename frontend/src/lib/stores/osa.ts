import { writable, get } from "svelte/store";
import { browser } from "$app/environment";
import { getApiBaseUrl, getCSRFToken, initCSRF } from "$lib/api/base";
import { currentWorkspaceId } from "$lib/stores/workspaces";
import type { SkillExecution } from "$lib/types/skills";
import type { AttachedFile } from "$lib/stores/chat/types";
import { parseChatSSEStream, isSSEStream } from "$lib/utils/chatSSEParser";

// ─── Types ───────────────────────────────────────────────────────────────────

export type OsaMode = "BUILD" | "ASSIST" | "ANALYZE" | "EXECUTE" | "MAINTAIN";

export interface OsaModeInfo {
  mode: OsaMode;
  label: string;
  description: string;
}

export interface OsaMessage {
  id: string;
  role: "user" | "osa";
  content: string;
  mode: OsaMode;
  confidence?: number;
  timestamp: Date;
  /** Response duration in milliseconds (set on OSA responses) */
  durationMs?: number;
  /** Model that generated this response */
  model?: string;
  /** Set when BUILD mode creates a module — links to /settings/modules */
  module_id?: string;
  /** Set when EXECUTE mode proposes a skill — rendered as inline decision card */
  skill_execution?: SkillExecution;
}

export interface OsaState {
  activeMode: OsaMode;
  modeConfidence: number;
  /** Available modes — loaded from API with hardcoded fallback */
  modes: OsaModeInfo[];
  modesLoaded: boolean;
  conversation: OsaMessage[];
  /** Persisted conversation ID from backend — maintains context across messages */
  conversationId: string | null;
  isStreaming: boolean;
  streamingContent: string;
  isExpanded: boolean;
  error: string | null;
  /** Signal Theory genre classification from the last signal_classified event */
  activeGenre: string | null;
  /** Document type hint from the last signal_classified event */
  activeDocType: string | null;
  /** Signal weight from the last signal_classified event */
  signalWeight: number | null;
  /** Files attached to the next message */
  attachments: AttachedFile[];
  /** Active OSA model name (from health check) */
  activeModel: string | null;
  /** Active OSA provider (from health check) */
  activeProvider: string | null;
}

// ─── Constants ───────────────────────────────────────────────────────────────

const STORAGE_KEY_MODE = "osa_active_mode";
const DEFAULT_MODE: OsaMode = "ASSIST";

/** Canonical dot/accent colors per mode — shared by ModeSelector, ModeIndicator, etc. */
export const MODE_COLORS: Record<OsaMode, string> = {
  BUILD: "#3b82f6",
  ASSIST: "#22c55e",
  ANALYZE: "#a855f7",
  EXECUTE: "#f59e0b",
  MAINTAIN: "#6b7280",
};
const VALID_MODES: OsaMode[] = [
  "BUILD",
  "ASSIST",
  "ANALYZE",
  "EXECUTE",
  "MAINTAIN",
];

// ─── Fallback mode definitions (used when API is unavailable) ────────────────

const FALLBACK_MODES: OsaModeInfo[] = [
  { mode: "BUILD", label: "Build", description: "Create modules & features" },
  { mode: "ASSIST", label: "Assist", description: "Help with tasks" },
  { mode: "ANALYZE", label: "Analyze", description: "Surface insights" },
  { mode: "EXECUTE", label: "Execute", description: "Run actions & workflows" },
  { mode: "MAINTAIN", label: "Maintain", description: "Monitor systems" },
];

// ─── Helpers ─────────────────────────────────────────────────────────────────

function getInitialMode(): OsaMode {
  if (!browser) return DEFAULT_MODE;
  try {
    const stored = localStorage.getItem(STORAGE_KEY_MODE) as OsaMode | null;
    return stored && VALID_MODES.includes(stored) ? stored : DEFAULT_MODE;
  } catch {
    return DEFAULT_MODE;
  }
}

// ─── Store ───────────────────────────────────────────────────────────────────

function createOsaStore() {
  const initialState: OsaState = {
    activeMode: getInitialMode(),
    modeConfidence: 0,
    modes: FALLBACK_MODES,
    modesLoaded: false,
    conversation: [],
    conversationId: null,
    isStreaming: false,
    streamingContent: "",
    isExpanded: false,
    error: null,
    activeGenre: null,
    activeDocType: null,
    signalWeight: null,
    attachments: [],
    activeModel: null,
    activeProvider: null,
  };

  const { subscribe, set, update } = writable<OsaState>(initialState);

  // Active stream controller — abort to cancel in-flight requests
  let activeStreamController: AbortController | null = null;

  // Internal helper to get current state synchronously
  function getState(): OsaState {
    let current: OsaState = initialState;
    const unsub = subscribe((s) => (current = s));
    unsub();
    return current;
  }

  return {
    subscribe,

    setMode(mode: OsaMode) {
      update((s) => {
        if (browser) {
          try {
            localStorage.setItem(STORAGE_KEY_MODE, mode);
          } catch {
            // localStorage unavailable — silently continue
          }
        }
        return { ...s, activeMode: mode, error: null };
      });
    },

    /** Update active model + provider locally, then POST to config endpoint (fails gracefully) */
    async setModel(provider: string, model: string, url?: string) {
      update((s) => ({ ...s, activeProvider: provider, activeModel: model }));

      if (!browser) return;
      try {
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
        };
        const csrfToken = getCSRFToken();
        if (csrfToken) headers["X-CSRF-Token"] = csrfToken;

        const body: Record<string, string> = { provider, model };
        if (url) body.url = url;

        await fetch(`${getApiBaseUrl()}/osa/config`, {
          method: "POST",
          headers,
          credentials: "include",
          signal: AbortSignal.timeout(4000),
          body: JSON.stringify(body),
        });
      } catch {
        // Endpoint may not exist yet — local state is already updated, so silently continue
      }
    },

    /** Fetch modes from OSA API. Falls back silently to hardcoded list. */
    async loadModes() {
      // Skip if already loaded or not in browser
      if (!browser || getState().modesLoaded) return;

      try {
        const headers: Record<string, string> = {};
        const csrfToken = getCSRFToken();
        if (csrfToken) headers["X-CSRF-Token"] = csrfToken;

        const res = await fetch(`${getApiBaseUrl()}/osa/modes`, {
          method: "GET",
          headers,
          credentials: "include",
          signal: AbortSignal.timeout(3000), // 3s max — don't block UI
        });

        if (!res.ok) throw new Error(`HTTP ${res.status}`);

        const data: { modes?: OsaModeInfo[] } = await res.json();
        if (data.modes && Array.isArray(data.modes) && data.modes.length > 0) {
          const modes = data.modes;
          update((s) => ({ ...s, modes, modesLoaded: true }));
          return;
        }
      } catch {
        // API unavailable — silently use fallback modes
      }

      // Mark loaded even on failure so we don't retry every render
      update((s) => ({ ...s, modesLoaded: true }));
    },

    /** Fetch OSA health to get active model/provider info */
    async loadHealth() {
      if (!browser) return;
      try {
        const res = await fetch(`${getApiBaseUrl()}/osa/health`, {
          credentials: "include",
          signal: AbortSignal.timeout(3000),
        });
        if (!res.ok) return;
        const data = await res.json();
        update((s) => ({
          ...s,
          activeModel: data.model || data.osa_model || null,
          activeProvider: data.provider || data.osa_provider || null,
        }));
      } catch {
        // silently ignore
      }
    },

    setExpanded(expanded: boolean) {
      update((s) => ({ ...s, isExpanded: expanded }));
    },

    clearConversation() {
      update((s) => ({
        ...s,
        conversation: [],
        conversationId: null,
        streamingContent: "",
        isStreaming: false,
        error: null,
        attachments: [],
      }));
    },

    /** Cancel any in-flight SSE stream */
    cancelStream() {
      if (activeStreamController) {
        activeStreamController.abort();
        activeStreamController = null;
      }
    },

    /** Full cleanup — cancel streams and reset state. Call on unmount if needed. */
    destroy() {
      this.cancelStream();
      set(initialState);
    },

    addAttachment(file: AttachedFile) {
      update((s) => ({ ...s, attachments: [...s.attachments, file] }));
    },

    removeAttachment(id: string) {
      update((s) => ({
        ...s,
        attachments: s.attachments.filter((f) => f.id !== id),
      }));
    },

    clearAttachments() {
      update((s) => ({ ...s, attachments: [] }));
    },

    async sendMessage(content: string) {
      // Cancel any in-flight stream before starting a new one
      this.cancelStream();
      activeStreamController = new AbortController();

      const startTime = performance.now();
      const state = getState();
      const workspaceId = get(currentWorkspaceId);

      // Map OSA mode to focus_mode for the chat backend
      const FOCUS_MODE_MAP: Record<OsaMode, string> = {
        BUILD: "build",
        ASSIST: "general",
        ANALYZE: "analyze",
        EXECUTE: "general",
        MAINTAIN: "general",
      };

      // Add user message immediately
      const userMessage: OsaMessage = {
        id: crypto.randomUUID(),
        role: "user",
        content,
        mode: state.activeMode,
        timestamp: new Date(),
      };

      update((s) => ({
        ...s,
        conversation: [...s.conversation, userMessage],
        isStreaming: true,
        streamingContent: "",
        isExpanded: true,
        error: null,
      }));

      try {
        // Build request headers — ensure CSRF is available
        const headers: Record<string, string> = {
          "Content-Type": "application/json",
        };
        let csrfToken = getCSRFToken();
        if (!csrfToken) {
          await initCSRF();
          csrfToken = getCSRFToken();
        }
        if (csrfToken) {
          headers["X-CSRF-Token"] = csrfToken;
        }

        // Build full request body matching ChatStreamManager pattern
        const requestBody: Record<string, unknown> = {
          message: content,
          conversation_id: state.conversationId,
          workspace_id: workspaceId,
          focus_mode: FOCUS_MODE_MAP[state.activeMode] ?? "general",
          structured_output: true,
        };

        // Include file attachment metadata if any
        if (state.attachments.length > 0) {
          requestBody.attachments = state.attachments.map((a) => ({
            name: a.name,
            type: a.type,
            size: a.size,
            content: a.content,
          }));
        }

        const response = await fetch(`${getApiBaseUrl()}/chat/message`, {
          method: "POST",
          headers,
          credentials: "include",
          signal: activeStreamController?.signal,
          body: JSON.stringify(requestBody),
        });

        if (!response.ok) {
          const errorData = await response
            .json()
            .catch(() => ({ detail: "Chat failed" }));
          throw new Error(
            errorData.detail || `Chat failed (HTTP ${response.status})`,
          );
        }

        // Extract conversation ID from backend for multi-turn continuity
        const newConvId = response.headers.get("X-Conversation-Id");
        if (newConvId) {
          update((s) => ({ ...s, conversationId: newConvId }));
        }

        if (!response.body) {
          throw new Error("No response stream");
        }

        // Read the streaming response — detect SSE vs plain text
        const rawReader = response.body.getReader();
        let fullContent = "";
        let detectedMode: OsaMode | undefined;
        let detectedConfidence: number | undefined;

        // Peek at first chunk for SSE detection
        const firstRead = await rawReader.read();
        const firstChunk = firstRead.done
          ? ""
          : new TextDecoder().decode(firstRead.value);
        const streamIsSSE = !firstRead.done && isSSEStream(firstChunk);

        if (streamIsSSE) {
          // SSE path: replay first chunk + delegate to parseChatSSEStream
          const firstValue = firstRead.value!;
          const replayStream = new ReadableStream<Uint8Array>({
            async start(controller) {
              controller.enqueue(firstValue);
              while (true) {
                const { done, value } = await rawReader.read();
                if (done) {
                  controller.close();
                  break;
                }
                controller.enqueue(value);
              }
            },
            cancel() {
              rawReader.cancel();
            },
          });
          const typedReader =
            replayStream.getReader() as ReadableStreamDefaultReader<Uint8Array>;

          for await (const event of parseChatSSEStream(typedReader)) {
            switch (event.type) {
              case "token":
                if (event.content) {
                  fullContent += event.content;
                  update((s) => ({ ...s, streamingContent: fullContent }));
                }
                break;
              case "signal_classified":
                detectedMode = event.mode as OsaMode;
                detectedConfidence = event.confidence;
                update((s) => ({
                  ...s,
                  activeMode: detectedMode ?? s.activeMode,
                  modeConfidence: detectedConfidence ?? s.modeConfidence,
                  activeGenre: event.genre ?? s.activeGenre,
                  activeDocType: event.docType ?? s.activeDocType,
                  signalWeight: event.weight ?? s.signalWeight,
                }));
                break;
              case "error":
                console.error("[OSA Store] SSE error event:", event.message);
                // Only show error if no content was received yet
                if (!fullContent.trim()) {
                  fullContent += event.message
                    ? `⚠️ ${event.message}`
                    : "⚠️ An error occurred while generating a response.";
                  update((s) => ({ ...s, streamingContent: fullContent }));
                }
                break;
              case "done":
                break;
            }
          }
        } else {
          // Plain-text fallback
          const decoder = new TextDecoder();
          if (!firstRead.done) {
            fullContent += decoder.decode(firstRead.value, { stream: true });
            update((s) => ({ ...s, streamingContent: fullContent }));
          }
          while (true) {
            const { done, value } = await rawReader.read();
            if (done) break;
            fullContent += decoder.decode(value, { stream: true });
            update((s) => ({ ...s, streamingContent: fullContent }));
          }
        }

        // Finalize: add OSA response to conversation
        const durationMs = Math.round(performance.now() - startTime);
        const currentState = getState();
        const osaMessage: OsaMessage = {
          id: crypto.randomUUID(),
          role: "osa",
          content: fullContent,
          mode: detectedMode ?? currentState.activeMode,
          confidence: detectedConfidence ?? currentState.modeConfidence,
          timestamp: new Date(),
          durationMs,
          model: currentState.activeModel ?? undefined,
        };

        update((s) => ({
          ...s,
          conversation: [...s.conversation, osaMessage],
          isStreaming: false,
          streamingContent: "",
          attachments: [],
        }));
      } catch (err) {
        // AbortError is expected when user cancels — don't show as error
        if (err instanceof DOMException && err.name === "AbortError") {
          update((s) => ({
            ...s,
            isStreaming: false,
            streamingContent: "",
          }));
          return;
        }
        const message =
          err instanceof Error ? err.message : "Failed to send message";
        console.error("[OSA Store] Send failed:", err);
        update((s) => ({
          ...s,
          isStreaming: false,
          streamingContent: "",
          error: message,
        }));
      }
    },
  };
}

export const osaStore = createOsaStore();
