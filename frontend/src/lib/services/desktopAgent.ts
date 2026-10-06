import type { AgentRuntime } from "$lib/stores/osa";
type AgentEvent = { id: string; type: string; content?: string };
type Bridge = {
  detect: () => Promise<Record<string, boolean>>;
  run?: (request: {
    id: string;
    runtime: string;
    prompt: string;
    model?: string;
  }) => Promise<unknown>;
  cancel?: (id: string) => Promise<void>;
  onEvent?: (callback: (event: AgentEvent) => void) => () => void;
};
export function agentBridge(): Bridge | undefined {
  if (typeof window === "undefined") return undefined;
  // Built-in modules may be shown inside a same-origin desktop iframe.
  // Cross-origin embedded apps cannot read the parent and get no bridge.
  let frame: Window = window;
  for (let depth = 0; depth < 8; depth++) {
    try {
      if (frame.location.origin !== window.location.origin) return undefined;
      const bridge = (frame as unknown as { electron?: { agents?: Bridge } })
        .electron?.agents;
      if (bridge) return bridge;
      if (frame.parent === frame) return undefined;
      frame = frame.parent;
    } catch {
      return undefined;
    }
  }
  return undefined;
}
export async function runDesktopAgent(
  runtime: AgentRuntime,
  prompt: string,
  model: string | undefined,
  signal: AbortSignal,
  onToken: (content: string) => void,
): Promise<string> {
  const bridge = agentBridge();
  if (!bridge?.run || !bridge.onEvent || !bridge.cancel)
    throw new Error(
      "This runtime needs the updated BusinessOS desktop app. Restart the desktop app to connect it.",
    );
  signal.throwIfAborted();
  const id = crypto.randomUUID();
  return new Promise((resolve, reject) => {
    let content = "",
      finished = false;
    const finish = (error?: Error) => {
      if (finished) return;
      finished = true;
      unsubscribe();
      signal.removeEventListener("abort", abort);
      error ? reject(error) : resolve(content);
    };
    const abort = () => {
      void bridge.cancel!(id);
      finish(new DOMException("Cancelled", "AbortError"));
    };
    const unsubscribe = bridge.onEvent!((event) => {
      if (event.id !== id || finished) return;
      if (event.type === "token") {
        content += event.content || "";
        onToken(content);
      }
      if (event.type === "error") {
        void bridge.cancel!(id);
        finish(new Error(event.content || `${runtime} failed`));
      }
      if (event.type === "done")
        finish(
          content.trim()
            ? undefined
            : new Error(`${runtime} returned no answer`),
        );
    });
    signal.addEventListener("abort", abort, { once: true });
    bridge.run!({ id, runtime, prompt, model })
      .then(() => {
        if (signal.aborted) void bridge.cancel!(id);
      })
      .catch((error) =>
        finish(error instanceof Error ? error : new Error(String(error))),
      );
  });
}
