import { test } from "node:test";
import assert from "node:assert/strict";
import { commandFor, jsonEvent } from "./runner";
test("user text is passed as data, never interpolated into a shell command", () => {
  const text = "$(touch /tmp/should-not-exist); `whoami`";
  for (const runtime of ["claude", "codex", "ollama"] as const) {
    const cmd = commandFor(runtime, text, "test");
    assert.equal(cmd.stdin, text);
    assert.ok(!cmd.args.includes(text));
  }
  assert.deepEqual(commandFor("hermes", text, "test").args, [
    "chat",
    "--quiet",
    "--provider",
    "ollama",
    "-m",
    "test",
    "-q",
    text,
  ]);
});
test("parses only assistant text and reports actual CLI failures", () => {
  assert.deepEqual(
    jsonEvent(
      "claude",
      JSON.stringify({
        type: "stream_event",
        event: {
          type: "content_block_delta",
          delta: { type: "text_delta", text: "hello" },
        },
      }),
    ),
    { type: "token", content: "hello" },
  );
  assert.equal(
    jsonEvent(
      "claude",
      JSON.stringify({
        type: "system",
        subtype: "init",
        apiKeySource: "private",
      }),
    ),
    null,
  );
  assert.deepEqual(
    jsonEvent(
      "codex",
      JSON.stringify({
        type: "item.completed",
        item: { type: "agent_message", text: "answer" },
      }),
    ),
    { type: "token", content: "answer" },
  );
  assert.deepEqual(
    jsonEvent(
      "codex",
      JSON.stringify({
        type: "turn.failed",
        error: { message: "Sign in required" },
      }),
    ),
    { type: "error", content: "Sign in required" },
  );
});

test("desktop coding agents use the configured non-interactive permission mode", () => {
  assert.ok(
    commandFor("claude", "test").args.includes(
      "--dangerously-skip-permissions",
    ),
  );
  assert.ok(
    commandFor("codex", "test").args.includes(
      "--dangerously-bypass-approvals-and-sandbox",
    ),
  );
});
