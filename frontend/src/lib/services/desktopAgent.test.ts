import { afterEach, expect, it, vi } from "vitest";
import { runDesktopAgent } from "./desktopAgent";
afterEach(() => {
  delete (window as any).electron;
});
it("subscribes before dispatch, ignores other runs, and cleans up on completion", async () => {
  let listener: (event: any) => void = () => {};
  const unsubscribe = vi.fn(),
    cancel = vi.fn();
  const run = vi.fn(async ({ id }) => {
    listener({ id: "unrelated", type: "token", content: "wrong" });
    listener({ id, type: "token", content: "Hello" });
    listener({ id, type: "done" });
  });
  (window as any).electron = {
    agents: {
      run,
      cancel,
      onEvent: (fn: any) => {
        listener = fn;
        return unsubscribe;
      },
    },
  };
  const token = vi.fn();
  await expect(
    runDesktopAgent(
      "codex",
      "hello",
      undefined,
      new AbortController().signal,
      token,
    ),
  ).resolves.toBe("Hello");
  expect(token).toHaveBeenCalledExactlyOnceWith("Hello");
  expect(unsubscribe).toHaveBeenCalledOnce();
});
it("cancels the native run when the conversation is stopped", async () => {
  const cancel = vi.fn(),
    unsubscribe = vi.fn();
  (window as any).electron = {
    agents: {
      run: vi.fn().mockResolvedValue({ started: true }),
      cancel,
      onEvent: () => unsubscribe,
    },
  };
  const controller = new AbortController();
  const pending = runDesktopAgent(
    "claude",
    "hello",
    undefined,
    controller.signal,
    vi.fn(),
  );
  controller.abort();
  await expect(pending).rejects.toMatchObject({ name: "AbortError" });
  expect(cancel).toHaveBeenCalled();
  expect(unsubscribe).toHaveBeenCalledOnce();
});

it('connects same-origin embedded modules to their desktop parent', async () => {
  const {agentBridge}=await import('./desktopAgent');
  const bridge={detect:vi.fn()};
  const parent=vi.spyOn(window,'parent','get').mockReturnValue({location:{origin:window.location.origin},electron:{agents:bridge}} as unknown as Window);
  try {expect(agentBridge()).toBe(bridge);} finally {parent.mockRestore();}
});
it('does not expose the desktop bridge across origins', async () => {
  const {agentBridge}=await import('./desktopAgent');
  const parent=vi.spyOn(window,'parent','get').mockReturnValue({location:{origin:'https://different.example'},electron:{agents:{}}} as unknown as Window);
  try {expect(agentBridge()).toBeUndefined();} finally {parent.mockRestore();}
});
