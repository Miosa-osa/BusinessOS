import { beforeEach, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";
vi.mock("$app/environment", () => ({ browser: true }));
vi.mock("$lib/api/base", () => ({
  getApiBaseUrl: () => "/api/v1",
  getCSRFToken: () => "",
  initCSRF: async () => {},
  getActiveWorkspaceHeaders: () => ({ "X-Workspace-ID": "workspace-a" }),
}));
vi.mock("$lib/stores/workspaces", () => ({
  currentWorkspaceId: writable("workspace-a"),
}));
beforeEach(() => {
  vi.resetModules();
  localStorage.clear();
});
it("restores a saved thread and its last runtime without losing the saved ID", async () => {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          conversation: { id: "saved" },
          messages: [
            {
              id: "1",
              role: "user",
              content: "remember this",
              created_at: "2026-09-30T10:00:00Z",
              metadata: { runtime: "codex" },
            },
            {
              id: "2",
              role: "assistant",
              content: "remembered",
              created_at: "2026-09-30T10:00:01Z",
              metadata: { runtime: "codex" },
            },
          ],
        }),
      ),
    ),
  );
  const { osaStore } = await import("./osa");
  await osaStore.loadConversation("saved");
  expect(get(osaStore).conversationId).toBe("saved");
  expect(get(osaStore).activeRuntime).toBe("codex");
  expect(get(osaStore).conversation.map((m) => m.content)).toEqual([
    "remember this",
    "remembered",
  ]);
  expect(localStorage.getItem("osa_conversation_workspace-a")).toBe("saved");
  osaStore.clearConversation();
  expect(get(osaStore).conversation).toEqual([]);
  expect(localStorage.getItem("osa_conversation_workspace-a")).toBeNull();
  expect(fetch).toHaveBeenCalledTimes(1);
});
it("does not restore an older request over a newly cleared conversation", async () => {
  let finish!: (response: Response) => void;
  vi.stubGlobal(
    "fetch",
    vi.fn(() => new Promise((resolve) => (finish = resolve))),
  );
  const { osaStore } = await import("./osa");
  const pending = osaStore.loadConversation("old");
  osaStore.clearConversation();
  finish(
    new Response(JSON.stringify({ conversation: { id: "old" }, messages: [] })),
  );
  await pending;
  expect(get(osaStore).conversationId).toBeNull();
  expect(localStorage.getItem("osa_conversation_workspace-a")).toBeNull();
});
