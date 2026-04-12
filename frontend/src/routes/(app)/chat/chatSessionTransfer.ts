import { tick } from "svelte";
import type { chatConversationStore } from "$lib/stores/chat/chatConversationStore.svelte";
import type { chatContextStore } from "$lib/stores/chat/chatContextStore.svelte";
import type { chatModelStore } from "$lib/stores/chat/chatModelStore.svelte";
import type { chatAgentStore } from "$lib/stores/chat/chatAgentStore.svelte";

type CS = typeof chatConversationStore;
type CX = typeof chatContextStore;
type MS = typeof chatModelStore;
type AG = typeof chatAgentStore;

export function checkForSpotlightTransfer(cs: CS, cx: CX): void {
  const transferData = sessionStorage.getItem("spotlightChatTransfer");
  if (!transferData) return;
  try {
    const data = JSON.parse(transferData);
    sessionStorage.removeItem("spotlightChatTransfer");

    if (data.conversationId) {
      cs.conversationId = data.conversationId;
      cs.activeConversationId = data.conversationId;
      cs.selectConversation(data.conversationId);
    } else if (data.messages?.length > 0) {
      cs.messages = data.messages.map(
        (m: { role: string; content: string }, i: number) => ({
          id: `spotlight-${i}`,
          role: m.role as "user" | "assistant",
          content: m.content,
        }),
      );
    }

    if (data.projectId) {
      cx.selectedProjectId = data.projectId;
    }
  } catch (e) {
    console.error("Failed to parse spotlight transfer data:", e);
  }
}

export async function checkForQuickChatMessage(
  cs: CS,
  cx: CX,
  ms: MS,
  sendMessage: () => void,
): Promise<void> {
  const quickChatData = sessionStorage.getItem("quickChatMessage");
  if (!quickChatData) return;
  try {
    const data = JSON.parse(quickChatData);
    sessionStorage.removeItem("quickChatMessage");

    if (Date.now() - data.timestamp < 5000) {
      if (data.isNewConversation) {
        cs.startNewConversation();
      }
      if (data.projectId) cx.selectedProjectId = data.projectId;
      if (data.model) ms.selectedModel = data.model;

      cs.inputValue = data.message;
      await tick();
      sendMessage();
    }
  } catch (e) {
    console.error("Failed to parse quick chat message:", e);
  }
}

export function checkForVoiceTranscript(
  cs: CS,
  ag: AG,
  inputRef: HTMLTextAreaElement | undefined,
): void {
  const transcriptData = sessionStorage.getItem("voiceTranscript");
  if (!transcriptData) return;
  try {
    const data = JSON.parse(transcriptData);
    sessionStorage.removeItem("voiceTranscript");

    if (Date.now() - data.timestamp < 10000) {
      if (ag.focusModeEnabled && !cs.conversationId) {
        ag.focusModeInitialInput = data.message;
        setTimeout(() => {
          ag.focusModeInitialInput = "";
        }, 500);
      } else {
        cs.inputValue = data.message;
        setTimeout(() => inputRef?.focus(), 100);
      }
    }
  } catch (e) {
    console.error("Failed to parse voice transcript:", e);
  }
}
