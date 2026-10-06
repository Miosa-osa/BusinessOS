import { describe, expect, it } from "vitest";

import {
  AGENT_RUNTIMES,
  DEFAULT_AGENT_RUNTIME,
  runtimeLabel,
  runtimeOption,
} from "./workspaceAgents";

describe("workspace agent runtimes", () => {
  it("keeps OSA as the hosted default", () => {
    expect(DEFAULT_AGENT_RUNTIME).toBe("osa");
    expect(runtimeOption(DEFAULT_AGENT_RUNTIME)).toMatchObject({
      label: "OSA",
      hosted: true,
    });
  });

  it("marks machine harnesses as requiring a connection", () => {
    expect(AGENT_RUNTIMES.filter((runtime) => !runtime.hosted).map((runtime) => runtime.id)).toEqual([
      "claude-code",
      "codex",
      "hermes",
    ]);
    expect(runtimeLabel("codex")).toBe("Codex");
  });
});
