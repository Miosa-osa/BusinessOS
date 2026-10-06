import { spawn } from "node:child_process";
import { terminalPath } from "../terminal/env";

export type Runtime = "claude" | "codex" | "ollama" | "hermes";
export type AgentEvent = { type: "token" | "error" | "done"; content?: string };
export function commandFor(
  runtime: Runtime,
  prompt: string,
  model?: string,
): { bin: string; args: string[]; stdin?: string } {
  switch (runtime) {
    case "claude":
      return {
        bin: "claude",
        args: [
          "-p",
          "--dangerously-skip-permissions",
          "--output-format",
          "stream-json",
          "--verbose",
          "--include-partial-messages",
        ],
        stdin: prompt,
      };
    case "codex":
      return {
        bin: "codex",
        args: [
          "exec",
          "--dangerously-bypass-approvals-and-sandbox",
          "--json",
          "--skip-git-repo-check",
          "-",
        ],
        stdin: prompt,
      };
    case "hermes":
      if (!model || model.startsWith("-"))
        throw new Error("Select an Ollama model for Hermes before sending.");
      return {
        bin: "hermes",
        args: [
          "chat",
          "--quiet",
          "--provider",
          "ollama",
          "-m",
          model,
          "-q",
          prompt,
        ],
      };
    case "ollama":
      if (!model || model.startsWith("-"))
        throw new Error("Select an Ollama model before sending.");
      return {
        bin: "ollama",
        args: ["run", model, "--hidethinking", "--nowordwrap"],
        stdin: prompt,
      };
    default:
      throw new Error("Unsupported agent runtime");
  }
}

export function jsonEvent(runtime: Runtime, line: string): AgentEvent | null {
  let data: any;
  try {
    data = JSON.parse(line);
  } catch {
    return null;
  }
  if (runtime === "claude") {
    if (
      data.type === "stream_event" &&
      data.event?.type === "content_block_delta" &&
      data.event.delta?.type === "text_delta"
    )
      return { type: "token", content: data.event.delta.text };
    if (data.type === "result" && data.is_error)
      return {
        type: "error",
        content: data.result || data.errors?.join("\n") || "Claude Code failed",
      };
  }
  if (runtime === "codex") {
    if (data.type === "item.completed" && data.item?.type === "agent_message")
      return { type: "token", content: data.item.text };
    if (data.type === "turn.failed" || data.type === "error")
      return {
        type: "error",
        content: data.error?.message || data.message || "Codex failed",
      };
  }
  return null;
}

export function runAgent(
  runtime: Runtime,
  prompt: string,
  cwd: string,
  emit: (event: AgentEvent) => void,
  model?: string,
) {
  const command = commandFor(runtime, prompt, model);
  const env: NodeJS.ProcessEnv = {
    ...process.env,
    PATH: terminalPath(),
    NO_COLOR: "1",
    TERM: "dumb",
  };
  // These variables describe the parent app, not the new CLI session.
  delete env.ELECTRON_RUN_AS_NODE;
  delete env.CLAUDECODE;
  // BusinessOS uses the local Ollama endpoint; leave Hermes global credentials untouched.
  if (runtime === "hermes") env.CUSTOM_BASE_URL = "http://127.0.0.1:11434/v1";
  const child = spawn(command.bin, command.args, {
    cwd,
    env,
    stdio: ["pipe", "pipe", "pipe"],
    detached: process.platform !== "win32",
  });
  let buffer = "",
    stderr = "",
    answer = "",
    finalText = "",
    failed = false,
    ended = false,
    claudeEnded = false;
  const output = (event: AgentEvent) => {
    if (ended) return;
    if (event.type === "token") answer += event.content || "";
    if (event.type === "error") failed = true;
    emit(event);
  };
  const line = (value: string) => {
    if (runtime === "claude" || runtime === "codex") {
      const event = jsonEvent(runtime, value);
      if (event) output(event);
      try {
        const data = JSON.parse(value);
        if (runtime === "claude") {
          if (data.type === "result" && !data.is_error) {
            finalText = data.result || "";
            if (!answer && finalText)
              output({ type: "token", content: finalText });
            cancel();
          }
          if (
            data.type === "stream_event" &&
            data.event?.type === "message_delta" &&
            data.event.delta?.stop_reason === "end_turn"
          )
            claudeEnded = true;
          if (
            data.type === "stream_event" &&
            data.event?.type === "message_stop" &&
            claudeEnded &&
            answer
          )
            cancel();
        }
        if (runtime === "codex" && data.type === "turn.completed") cancel();
      } catch {}
    } else {
      // Hermes adds a session footer in quiet mode; it is not part of the answer.
      if (
        runtime === "hermes" &&
        /^(Session (?:ID|saved)|Resume (?:with|this)|To resume)/i.test(
          value.trim(),
        )
      )
        return;
      output({ type: "token", content: value + "\n" });
    }
  };
  child.stdout.setEncoding("utf8");
  child.stderr.setEncoding("utf8");
  child.stdout.on("data", (chunk: string) => {
    buffer += chunk;
    let end;
    while ((end = buffer.indexOf("\n")) >= 0) {
      line(buffer.slice(0, end).replace(/\r$/, ""));
      buffer = buffer.slice(end + 1);
    }
  });
  child.stderr.on("data", (chunk: string) => {
    stderr = (stderr + chunk).slice(-12000);
  });
  const finish = () => {
    if (!ended) {
      ended = true;
      clearTimeout(timer);
      emit({ type: "done" });
    }
  };
  child.on("error", (error) => {
    output({
      type: "error",
      content: `Could not start ${runtime}: ${error.message}`,
    });
    finish();
  });
  child.on("close", (code) => {
    if (ended) return;
    if (buffer) line(buffer);
    if (!answer && finalText) output({ type: "token", content: finalText });
    if (code !== 0 && !failed)
      output({
        type: "error",
        content: `${runtime} exited with code ${code}. ${stderr.trim() || "Check its sign-in and settings."}`,
      });
    if (!answer && !failed)
      output({
        type: "error",
        content: `${runtime} returned no answer. ${stderr.trim()}`,
      });
    finish();
  });
  child.stdin.on("error", () => {});
  child.stdin.end(command.stdin || "");
  function cancel() {
    if (ended) return;
    try {
      if (child.pid && process.platform !== "win32")
        process.kill(-child.pid, "SIGTERM");
      else child.kill("SIGTERM");
    } catch {}
    finish();
  }
  const timer = setTimeout(
    () => {
      output({ type: "error", content: `${runtime} timed out` });
      cancel();
    },
    10 * 60 * 1000,
  );
  return { cancel };
}
