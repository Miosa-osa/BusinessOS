import { toast } from "svelte-sonner";
import type { chatConversationStore } from "$lib/stores/chat/chatConversationStore.svelte";
import type { chatContextStore } from "$lib/stores/chat/chatContextStore.svelte";
import type { chatModelStore } from "$lib/stores/chat/chatModelStore.svelte";
import type { chatAgentStore } from "$lib/stores/chat/chatAgentStore.svelte";
import type { chatArtifactStore } from "$lib/stores/chat/chatArtifactStore.svelte";
import type { chatUIStore } from "$lib/stores/chat/chatUIStore.svelte";
import type { ChatMessage, StreamingToolCall } from "$lib/stores/chat/types";
import type { SignalClassifiedEvent } from "$lib/utils/chatSSEParser";

type CS = typeof chatConversationStore;
type CX = typeof chatContextStore;
type MS = typeof chatModelStore;
type AG = typeof chatAgentStore;
type AR = typeof chatArtifactStore;
type UI = typeof chatUIStore;

export interface StreamingState {
  currentThinking: string;
  hasThinking: boolean;
  thinkingExpanded: boolean;
  streamingToolCalls: Map<string, StreamingToolCall>;
  streamingSignalMode: {
    mode: SignalClassifiedEvent["mode"];
    confidence?: number;
    genre?: SignalClassifiedEvent["genre"];
    docType?: string;
    weight?: number;
  } | null;
}

export interface StreamingStateSetters {
  setCurrentThinking: (v: string) => void;
  setHasThinking: (v: boolean) => void;
  setThinkingExpanded: (v: boolean) => void;
  setStreamingToolCalls: (v: Map<string, StreamingToolCall>) => void;
  setStreamingSignalMode: (v: StreamingState["streamingSignalMode"]) => void;
  getCurrentThinking: () => string;
  getHasThinking: () => boolean;
  getStreamingToolCalls: () => Map<string, StreamingToolCall>;
}

/**
 * Creates the stream callbacks object for ChatStreamManager.send().
 * Closes over store references and local state setters to keep +page.svelte lean.
 */
export function createStreamCallbacks(
  cs: CS,
  cx: CX,
  ms: MS,
  ag: AG,
  ar: AR,
  ui: UI,
  state: StreamingStateSetters,
) {
  return {
    addUserMessage(id: string, content: string) {
      cs.messages = [...cs.messages, { id, role: "user", content }];
    },
    addAssistantPlaceholder(id: string) {
      cs.messages = [...cs.messages, { id, role: "assistant", content: "" }];
    },
    updateAssistantContent(
      id: string,
      content: string,
      artifacts: ChatMessage["artifacts"],
      usage: ChatMessage["usage"],
    ) {
      cs.messages = cs.messages.map((m: ChatMessage) => {
        if (m.id !== id) return m;
        return {
          ...m,
          content,
          ...(artifacts !== undefined ? { artifacts } : {}),
          ...(usage !== undefined ? { usage } : {}),
        };
      });
    },
    onConversationCreated(id: string, title: string) {
      cs.onConversationCreated(id, title);
    },
    onThinkingStart() {
      state.setHasThinking(true);
      const searchContent = state
        .getCurrentThinking()
        .split("\n")
        .filter(
          (l) =>
            l.includes("Searching the web") ||
            (l.includes("Found") && l.includes("sources")),
        )
        .join("\n");
      state.setCurrentThinking(searchContent ? searchContent + "\n" : "");
      state.setThinkingExpanded(true);
    },
    onThinkingChunk(chunk: string) {
      state.setCurrentThinking(state.getCurrentThinking() + chunk);
      state.setHasThinking(true);
    },
    onThinkingEnd() {},
    onToolCall(toolCall: StreamingToolCall) {
      const next = new Map(state.getStreamingToolCalls());
      next.set(toolCall.toolCallId, toolCall);
      state.setStreamingToolCalls(next);
    },
    onToolResult(
      toolCallId: string,
      status: "completed" | "error",
      result?: string,
    ) {
      const updated = new Map(state.getStreamingToolCalls());
      const existing = updated.get(toolCallId);
      if (existing) {
        updated.set(toolCallId, { ...existing, status, result: result ?? "" });
      }
      state.setStreamingToolCalls(updated);
    },
    onSignalClassified(
      mode: SignalClassifiedEvent["mode"],
      confidence: number | undefined,
      genre: SignalClassifiedEvent["genre"],
      docType: string | undefined,
      weight: number | undefined,
    ) {
      state.setStreamingSignalMode({
        mode,
        confidence,
        genre,
        docType,
        weight,
      });
    },
    onArtifactStart(title: string, artifactType: string) {
      ar.onArtifactStart(title, artifactType);
      ui.rightPanelOpen = true;
      ui.rightPanelTab = "artifacts";
      const availableWidth = window.innerWidth - 256;
      ui.artifactPanelWidth = Math.floor(availableWidth / 2);
    },
    onArtifactComplete(artifact: {
      title: string;
      type: string;
      content: string;
    }) {
      // Updates local UI state only — no API call. Backend persists the artifact.
      ar.onArtifactComplete({
        title: artifact.title,
        artifactType: artifact.type,
        content: artifact.content,
      });
      ui.rightPanelOpen = true;
      ui.rightPanelTab = "artifacts";
    },
    onArtifactContentUpdate(content: string) {
      ar.generatingArtifactContent = content;
    },
    onStreamStart(abortController: AbortController) {
      cs.isStreaming = true;
      cs.abortController = abortController;
    },
    onStreamEnd() {
      cs.isStreaming = false;
      cs.abortController = null;

      const thinking = state.getCurrentThinking();
      const hasThinking = state.getHasThinking();
      if (thinking && hasThinking) {
        const lastAssistant = cs.messages.findLast(
          (m: ChatMessage) => m.role === "assistant",
        );
        if (lastAssistant) {
          const thinkingBlock = `<thinking>${thinking}</thinking>\n\n`;
          cs.messages = cs.messages.map((m: ChatMessage) =>
            m.id === lastAssistant.id
              ? { ...m, content: thinkingBlock + m.content }
              : m,
          );
        }
      }

      ar.generatingArtifact = false;
      ar.generatingArtifactTitle = "";
      ar.generatingArtifactType = "";
      ar.generatingArtifactContent = "";

      // Sync artifact store with what the backend persisted during postProcessStream.
      // A short delay allows the backend's post-processing goroutine to finish writing
      // before we fetch. The UI already shows the artifact via viewingArtifactFromMessage.
      setTimeout(() => {
        ar.loadArtifacts();
      }, 800);
    },
    onStreamError(message: string) {
      toast.error(message || "Failed to process message. Please try again.");
    },
    onArtifactFinalize(artifact: {
      title: string;
      type: string;
      content: string;
    }) {
      // No longer calls autoSaveArtifact — backend handles persistence.
      // Inline task creation is still triggered for actionable artifact types.
      const actionableTypes = ["plan", "framework", "proposal", "sop"];
      if (actionableTypes.includes(artifact.type.toLowerCase())) {
        ar.triggerInlineTaskCreation(artifact, cx.availableTeamMembers);
      }
    },
    onModelWarmedUp(model: string) {
      ms.markModelWarmedUp(model);
    },
  };
}
