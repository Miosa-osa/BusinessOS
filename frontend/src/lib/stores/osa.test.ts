import { beforeEach, describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";

vi.mock("$app/environment", () => ({ browser: true }));
vi.mock("$lib/api/base", () => ({
  getApiBaseUrl: () => "/api/v1",
  getCSRFToken: () => "csrf",
  initCSRF: vi.fn(),
  getActiveWorkspaceHeaders: () => ({ "X-Workspace-ID": "workspace-a" }),
}));
vi.mock("$lib/stores/workspaces", () => ({
  currentWorkspaceId: writable("workspace-a"),
}));

describe("OSA runtime connection", () => {
  beforeEach(() => {
    vi.resetModules();
    localStorage.clear();
    vi.stubGlobal("fetch", vi.fn());
  });
  it("does not report a disabled runtime as available", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ enabled: false, status: "disabled" })),
    );
    const { osaStore } = await import("./osa");
    await osaStore.loadHealth();
    expect(get(osaStore).osaAvailable).toBe(false);
  });
  it("keeps the confirmed model and exposes a rejected change", async () => {
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify({ error: "OSA rejected model" }), {
        status: 502,
      }),
    );
    const { osaStore } = await import("./osa");
    await expect(osaStore.setModel("ollama", "missing-model")).rejects.toThrow(
      "OSA rejected model",
    );
    expect(get(osaStore).activeModel).toBeNull();
  });
  it("routes through OSA and sends usable workspace-scoped context", async () => {
    vi.mocked(fetch).mockImplementation(async (url) => {
      if (String(url).endsWith("/optimal/search"))
        return new Response(
          JSON.stringify({
            results: [
              { l0_abstract: "Relevant fact", uri: "optimal://work/item" },
            ],
          }),
        );
      if (String(url).endsWith("/chat/message"))
        return new Response(
          'event: token\ndata: {"content":"Hello"}\n\nevent: done\ndata: {}\n\n',
          {
            headers: {
              "Content-Type": "text/event-stream",
              "X-Conversation-Id": "conversation-a",
              "X-OSA-Routing": "true",
            },
          },
        );
      return new Response("{}");
    });
    const { osaStore } = await import("./osa");
    await osaStore.sendMessage("Hello");
    const requests = vi.mocked(fetch).mock.calls;
    const search = requests.find(([url]) =>
      String(url).endsWith("/optimal/search"),
    )!;
    expect(search[1]?.headers).toMatchObject({
      "X-Workspace-ID": "workspace-a",
    });
    expect(search[1]?.signal).toBeDefined();
    const chat = requests.find(([url]) =>
      String(url).endsWith("/chat/message"),
    )!;
    const body = JSON.parse(String(chat[1]?.body));
    expect(body.runtime).toBe("osa");
    expect(body.message).toContain("Relevant fact");
    expect(body.message).toContain("optimal://work/item");
    expect(get(osaStore).conversation.at(-1)?.content).toBe("Hello");
  });
});

it("clears the busy state when a request is cancelled", async () => {
  vi.resetModules();
  localStorage.clear();
  vi.stubGlobal(
    "fetch",
    vi.fn(
      (_url, init) =>
        new Promise((_resolve, reject) => {
          init?.signal?.addEventListener("abort", () =>
            reject(new DOMException("Aborted", "AbortError")),
          );
        }),
    ),
  );
  const { osaStore } = await import("./osa");
  const pending = osaStore.sendMessage("Hello");
  expect(get(osaStore).isStreaming).toBe(true);
  osaStore.cancelStream();
  expect(get(osaStore).isStreaming).toBe(false);
  await pending;
});

it("reports tool progress and only confirms overdrive from the backend", async () => {
  vi.resetModules();
  localStorage.clear();
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url) => {
      if (String(url).endsWith("/chat/message"))
        return new Response(
          'event: tool_call\ndata: {"data":{"tool_name":"computer_use","status":"calling"}}\n\n' +
            'event: tool_result\ndata: {"data":{"tool_name":"computer_use","status":"error"}}\n\n' +
            'event: token\ndata: {"content":"Tool failed"}\n\nevent: done\ndata: {}\n\n',
          {
            headers: {
              "Content-Type": "text/event-stream",
              "X-OSA-Routing": "true",
              "X-OSA-Permission-Mode": "overdrive",
            },
          },
        );
      return new Response("{}");
    }),
  );
  const { osaStore } = await import("./osa");
  const activity: string[] = [];
  const unsubscribe = osaStore.subscribe((s) => activity.push(s.activity));
  await osaStore.sendMessage("Check screen");
  unsubscribe();
  expect(activity).toContain("Using computer use…");
  expect(activity).toContain("Tool failed: computer use. OSA is handling it…");
  expect(get(osaStore).permissionMode).toBe("overdrive");
});
