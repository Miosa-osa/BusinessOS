import { beforeEach, expect, it, vi } from "vitest";
import { OsaOrbVoice } from "./osaOrbVoice";
const mocks = vi.hoisted(() => ({
  sendMessage: vi.fn(),
  cancelStream: vi.fn(),
  setRuntime: vi.fn(),
}));
vi.mock("$lib/stores/osa", () => ({ osaStore: mocks }));
vi.mock("$lib/api/base", () => ({
  getActiveWorkspaceHeaders: () => ({ "X-Workspace-ID": "w" }),
  getApiBaseUrl: () => "/api",
  getCSRFToken: () => "csrf",
  initCSRF: async () => {},
}));
class Recorder {
  static current: Recorder;
  state = "inactive";
  mimeType = "audio/webm";
  ondataavailable?: (e: { data: Blob }) => void;
  onstop?: () => void;
  constructor() {
    Recorder.current = this;
  }
  start() {
    this.state = "recording";
  }
  stop() {
    this.state = "inactive";
    this.ondataavailable?.({ data: new Blob(["audio"]) });
    this.onstop?.();
  }
}
beforeEach(() => {
  vi.clearAllMocks();
  vi.stubGlobal("MediaRecorder", Recorder);
});
it("records, transcribes, sends through OSA and speaks the answer", async () => {
  const stop = vi.fn(),
    speak = vi.fn();
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: vi.fn().mockResolvedValue({ getTracks: () => [{ stop }] }),
    },
  });
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(new Response('{"text":"hello"}')),
  );
  vi.stubGlobal(
    "SpeechSynthesisUtterance",
    class {
      constructor(public text: string) {}
    },
  );
  Object.defineProperty(window, "speechSynthesis", {
    configurable: true,
    value: { speak, cancel: vi.fn() },
  });
  mocks.sendMessage.mockResolvedValue({ content: "Hello back" });
  const phase = vi.fn(),
    caption = vi.fn(),
    error = vi.fn();
  const voice = new OsaOrbVoice({ phase, caption, error });
  await voice.toggle();
  expect(phase).toHaveBeenLastCalledWith("listening");
  await voice.toggle();
  await vi.waitFor(() => expect(speak).toHaveBeenCalledOnce());
  expect(stop).toHaveBeenCalledOnce();
  expect(mocks.setRuntime).toHaveBeenCalledWith("osa");
  expect(mocks.sendMessage).toHaveBeenCalledWith("hello", { voice: true });
  expect(caption).toHaveBeenCalledWith("osa", "Hello back");
  expect(error).not.toHaveBeenCalled();
  voice.destroy();
});
it("releases a late microphone grant after cancellation without sending audio", async () => {
  const stop = vi.fn();
  let grant!: (value: unknown) => void;
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: () =>
        new Promise((resolve) => {
          grant = resolve;
        }),
    },
  });
  const voice = new OsaOrbVoice({
    phase: vi.fn(),
    caption: vi.fn(),
    error: vi.fn(),
  });
  const pending = voice.toggle();
  voice.cancel();
  grant({ getTracks: () => [{ stop }] });
  await pending;
  expect(stop).toHaveBeenCalledOnce();
  expect(mocks.sendMessage).not.toHaveBeenCalled();
  voice.destroy();
});
it("does not send silence markers to the agent", async () => {
  const stop = vi.fn();
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: {
      getUserMedia: vi.fn().mockResolvedValue({ getTracks: () => [{ stop }] }),
    },
  });
  vi.stubGlobal(
    "fetch",
    vi.fn().mockResolvedValue(new Response('{"text":"[BLANK_AUDIO]"}')),
  );
  const error = vi.fn();
  const voice = new OsaOrbVoice({ phase: vi.fn(), caption: vi.fn(), error });
  await voice.toggle();
  await voice.toggle();
  await vi.waitFor(() =>
    expect(error).toHaveBeenCalledWith(expect.stringContaining("No speech")),
  );
  expect(mocks.sendMessage).not.toHaveBeenCalled();
  voice.destroy();
});

it("shows incremental speech while listening and waits to send until stopped", async () => {
  vi.useFakeTimers();
  const stop = vi.fn(),
    transcript = vi.fn(),
    phase = vi.fn();
  Object.defineProperty(navigator, "mediaDevices", {
    configurable: true,
    value: { getUserMedia: async () => ({ getTracks: () => [{ stop }] }) },
  });
  vi.stubGlobal(
    "fetch",
    vi
      .fn()
      .mockImplementationOnce(
        async () => new Response('{"text":"first phrase"}'),
      )
      .mockImplementationOnce(
        async () => new Response('{"text":"second phrase"}'),
      ),
  );
  vi.stubGlobal(
    "SpeechSynthesisUtterance",
    class {
      constructor(public text: string) {}
    },
  );
  Object.defineProperty(window, "speechSynthesis", {
    configurable: true,
    value: { speak: vi.fn(), cancel: vi.fn() },
  });
  mocks.sendMessage.mockResolvedValue({ content: "Answer" });
  const voice = new OsaOrbVoice({
    phase,
    transcript,
    caption: vi.fn(),
    error: vi.fn(),
  });
  try {
    await voice.toggle();
    await vi.advanceTimersByTimeAsync(4000);
    expect(transcript).toHaveBeenLastCalledWith("first phrase");
    expect(phase).toHaveBeenLastCalledWith("listening");
    expect(mocks.sendMessage).not.toHaveBeenCalled();
    await voice.toggle();
    await vi.advanceTimersByTimeAsync(0);
    expect(mocks.sendMessage).toHaveBeenCalledWith(
      "first phrase second phrase",
      { voice: true },
    );
    expect(stop).toHaveBeenCalledOnce();
  } finally {
    voice.destroy();
    vi.useRealTimers();
  }
});
