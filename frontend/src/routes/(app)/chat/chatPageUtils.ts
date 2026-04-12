import { greetingSuggestions } from "./chatActions";
import type { chatUIStore } from "$lib/stores/chat/chatUIStore.svelte";
import type { chatConversationStore } from "$lib/stores/chat/chatConversationStore.svelte";
import type { chatArtifactStore } from "$lib/stores/chat/chatArtifactStore.svelte";

type UI = typeof chatUIStore;
type CS = typeof chatConversationStore;
type AR = typeof chatArtifactStore;

/**
 * Returns the combined artifact list: saved artifacts + inline message artifacts.
 * Call this inside a `$derived(() => ...)` in the component.
 */
export function computeAllArtifacts(ar: AR, cs: CS) {
  const combined: {
    id: string;
    title: string;
    type: string;
    content: string;
    fromMessage?: boolean;
  }[] = [];
  for (const a of ar.artifacts) {
    combined.push({
      id: a.id,
      title: a.title,
      type: a.type,
      content: "",
      fromMessage: false,
    });
  }
  for (const msg of cs.messages) {
    if (msg.role === "assistant" && msg.artifacts) {
      for (const a of msg.artifacts) {
        combined.push({
          id: `msg-${msg.id}-${a.title}`,
          title: a.title,
          type: a.type,
          content: a.content,
          fromMessage: true,
        });
      }
    }
  }
  return combined;
}

/**
 * Copies text to the clipboard using the Clipboard API, with an execCommand fallback.
 */
export async function copyToClipboard(text: string): Promise<void> {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
      return;
    }
  } catch {
    /* fall through to fallback */
  }
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.style.cssText = "position:fixed;top:0;left:0;opacity:0";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  try {
    document.execCommand("copy");
  } finally {
    document.body.removeChild(textarea);
  }
}

/**
 * Runs the typewriter greeting animation on the empty-state suggestion text.
 * Intended to be called inside a `$effect` in +page.svelte.
 *
 * Returns a cleanup function suitable for the `$effect` return value.
 */
export function runTypewriterEffect(ui: UI, cs: CS): () => void {
  if (cs.hasConversation) return () => {};

  const currentSuggestion = greetingSuggestions[ui.currentSuggestionIndex];
  let charIndex = 0;
  let direction: "typing" | "deleting" | "pausing" = "typing";
  let timeoutId: ReturnType<typeof setTimeout>;

  function tick() {
    if (direction === "typing") {
      if (charIndex <= currentSuggestion.length) {
        ui.displayedSuggestion = currentSuggestion.slice(0, charIndex);
        charIndex++;
        timeoutId = setTimeout(tick, 50 + Math.random() * 30);
      } else {
        direction = "pausing";
        timeoutId = setTimeout(tick, 2500);
      }
    } else if (direction === "pausing") {
      direction = "deleting";
      timeoutId = setTimeout(tick, 50);
    } else if (direction === "deleting") {
      if (charIndex > 0) {
        charIndex--;
        ui.displayedSuggestion = currentSuggestion.slice(0, charIndex);
        timeoutId = setTimeout(tick, 25);
      } else {
        ui.currentSuggestionIndex =
          (ui.currentSuggestionIndex + 1) % greetingSuggestions.length;
      }
    }
  }

  tick();
  return () => clearTimeout(timeoutId);
}
