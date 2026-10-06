import {
  getActiveWorkspaceHeaders,
  getApiBaseUrl,
  getCSRFToken,
  initCSRF,
} from "$lib/api/base";
import { osaStore } from "$lib/stores/osa";

export type OrbVoicePhase =
  | "idle"
  | "starting"
  | "listening"
  | "transcribing"
  | "thinking"
  | "speaking"
  | "error";
type Callbacks = {
  phase: (phase: OrbVoicePhase) => void;
  caption: (sender: "user" | "osa", text: string) => void;
  transcript?: (text: string) => void;
  error: (message: string) => void;
};

/** Bounded local recordings provide an incremental transcript and one OSA turn. */
export class OsaOrbVoice {
  private recorder: MediaRecorder | null = null;
  private stream: MediaStream | null = null;
  private controller: AbortController | null = null;
  private limit: ReturnType<typeof setTimeout> | null = null;
  private chunkTimer: ReturnType<typeof setTimeout> | null = null;
  private phase: OrbVoicePhase = "idle";
  private epoch = 0;
  private disposed = false;
  private finishing = false;
  private transcript = "";
  private queue: Promise<void> = Promise.resolve();
  constructor(private callbacks: Callbacks) {}
  private setPhase(phase: OrbVoicePhase) {
    this.phase = phase;
    this.callbacks.phase(phase);
  }

  async toggle() {
    if (this.phase === "listening") {
      this.finishRecording();
      return;
    }
    if (!["idle", "error"].includes(this.phase)) {
      this.cancel();
      return;
    }
    const epoch = ++this.epoch;
    this.finishing = false;
    this.transcript = "";
    this.queue = Promise.resolve();
    this.controller = new AbortController();
    this.callbacks.transcript?.("");
    this.setPhase("starting");
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      if (epoch !== this.epoch || this.disposed) {
        stream.getTracks().forEach((t) => t.stop());
        return;
      }
      this.stream = stream;
      this.recordChunk(epoch);
      this.limit = setTimeout(() => this.finishRecording(), 120_000);
      this.setPhase("listening");
    } catch (error) {
      if (epoch === this.epoch && !this.disposed)
        this.fail(this.message(error, "Microphone unavailable"));
    }
  }

  private recordChunk(epoch: number) {
    if (!this.stream || epoch !== this.epoch) return;
    const chunks: Blob[] = [];
    const controller = this.controller!;
    const recorder = new MediaRecorder(this.stream);
    this.recorder = recorder;
    recorder.ondataavailable = (event) => {
      if (event.data.size) chunks.push(event.data);
    };
    recorder.onerror = () => {
      if (epoch === this.epoch)
        this.fail("Microphone recording failed. Please try again.");
    };
    recorder.onstop = () => {
      if (this.chunkTimer) clearTimeout(this.chunkTimer);
      this.chunkTimer = null;
      if (epoch !== this.epoch || this.disposed) return;
      const audio = new Blob(chunks, {
        type: recorder.mimeType || "audio/webm",
      });
      this.queue = this.queue
        .then(async () => {
          if (epoch !== this.epoch || !audio.size) return;
          const text = await this.transcribe(audio, controller);
          if (epoch !== this.epoch) return;
          if (text) {
            this.transcript = [this.transcript, text].filter(Boolean).join(" ");
            this.callbacks.transcript?.(this.transcript);
          }
        })
        .catch((error) => {
          if (epoch === this.epoch)
            this.fail(this.message(error, "Voice transcription failed"));
        });
      if (this.finishing) {
        this.releaseMicrophone();
        void this.queue.then(() => {
          if (epoch === this.epoch) return this.respond(epoch);
        });
      } else {
        try {
          this.recordChunk(epoch);
        } catch (error) {
          this.fail(this.message(error, "Microphone recording failed"));
        }
      }
    };
    recorder.start();
    this.chunkTimer = setTimeout(() => {
      if (recorder.state === "recording") recorder.stop();
    }, 4000);
  }

  private finishRecording() {
    if (this.phase !== "listening") return;
    this.finishing = true;
    this.setPhase("transcribing");
    if (this.recorder?.state === "recording") this.recorder.stop();
  }

  private async transcribe(
    audio: Blob,
    controller: AbortController,
  ): Promise<string> {
    await initCSRF();
    controller.signal.throwIfAborted();
    const headers = getActiveWorkspaceHeaders();
    const csrf = getCSRFToken();
    if (csrf) headers["X-CSRF-Token"] = csrf;
    const form = new FormData();
    form.append(
      "audio",
      audio,
      `osa-voice.${audio.type.includes("mp4") ? "mp4" : "webm"}`,
    );
    const response = await fetch(`${getApiBaseUrl()}/transcribe`, {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
      signal: AbortSignal.any([controller.signal, AbortSignal.timeout(90_000)]),
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok)
      throw new Error(data.error || "Voice transcription failed");
    return typeof data.text === "string"
      ? data.text
          .replace(/\[(?:BLANK_AUDIO|SILENCE|NO_SPEECH|MUSIC)\]/gi, "")
          .trim()
      : "";
  }

  private async respond(epoch: number) {
    try {
      if (!this.transcript)
        throw new Error("No speech was detected. Tap the orb and try again.");
      this.callbacks.caption("user", this.transcript);
      this.callbacks.transcript?.("");
      this.setPhase("thinking");
      osaStore.setRuntime("osa");
      const answer = await osaStore.sendMessage(this.transcript, {
        voice: true,
      });
      if (epoch !== this.epoch) return;
      if (!answer)
        throw new Error("OSA did not return an answer. Please try again.");
      this.callbacks.caption("osa", answer.content);
      if (
        !window.speechSynthesis ||
        typeof SpeechSynthesisUtterance === "undefined"
      )
        throw new Error(
          "Speech playback is unavailable on this device. Your answer is shown above.",
        );
      const utterance = new SpeechSynthesisUtterance(
        answer.content
          .replace(/\[([^\]]+)\]\([^)]+\)/g, "$1")
          .replace(/[*#`]/g, ""),
      );
      utterance.onend = () => {
        if (epoch === this.epoch) this.setPhase("idle");
      };
      utterance.onerror = () => {
        if (epoch === this.epoch)
          this.fail("Audio playback failed. Your answer is shown above.");
      };
      this.setPhase("speaking");
      window.speechSynthesis.speak(utterance);
    } catch (error) {
      if (epoch === this.epoch && !this.disposed)
        this.fail(this.message(error, "Voice request failed"));
    }
  }

  private message(error: unknown, fallback: string) {
    return error instanceof Error ? error.message : fallback;
  }
  private releaseMicrophone() {
    if (this.limit) clearTimeout(this.limit);
    if (this.chunkTimer) clearTimeout(this.chunkTimer);
    this.limit = this.chunkTimer = null;
    this.stream?.getTracks().forEach((track) => track.stop());
    this.stream = null;
    this.recorder = null;
  }
  private fail(message: string) {
    ++this.epoch;
    this.controller?.abort();
    if (this.recorder?.state === "recording") this.recorder.stop();
    this.releaseMicrophone();
    this.setPhase("error");
    this.callbacks.error(message);
  }
  cancel() {
    ++this.epoch;
    this.controller?.abort();
    if (this.recorder?.state === "recording") this.recorder.stop();
    this.releaseMicrophone();
    if (this.phase === "thinking") osaStore.cancelStream();
    if (this.phase === "speaking") window.speechSynthesis?.cancel();
    this.callbacks.transcript?.("");
    this.setPhase("idle");
  }
  destroy() {
    this.disposed = true;
    this.cancel();
  }
}
