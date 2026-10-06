import { ipcMain, app } from "electron";
import { runAgent, type Runtime } from "./runner";

export function setupAgentHandlers(resolveHome: () => Promise<string>) {
  const watched = new Set<number>();
  const runs = new Map<string, { owner: number; cancel: () => void }>();
  ipcMain.handle(
    "agents:run",
    async (
      event,
      request: { id: string; runtime: Runtime; prompt: string; model?: string },
    ) => {
      if (event.senderFrame !== event.sender.mainFrame)
        throw new Error("Agent requests must come from the desktop window");
      if (
        !request ||
        !["claude", "codex", "ollama", "hermes"].includes(request.runtime) ||
        typeof request.id !== "string" ||
        request.id.length > 80 ||
        typeof request.prompt !== "string" ||
        request.prompt.length > 200000
      )
        throw new Error("Invalid agent request");
      const key = `${event.sender.id}:${request.id}`;
      if (runs.has(key)) throw new Error("This request is already running");
      const cwd = await resolveHome();
      const run = runAgent(
        request.runtime,
        request.prompt,
        cwd,
        (update) => {
          if (!event.sender.isDestroyed())
            event.sender.send("agents:event", { id: request.id, ...update });
          if (update.type === "done") runs.delete(key);
        },
        request.model,
      );
      runs.set(key, { owner: event.sender.id, cancel: run.cancel });
      if (!watched.has(event.sender.id)) {
        const owner = event.sender.id;
        watched.add(owner);
        event.sender.once("destroyed", () => {
          for (const [id, run] of runs) {
            if (run.owner === owner) {
              run.cancel();
              runs.delete(id);
            }
          }
          watched.delete(owner);
        });
      }
      return { started: true };
    },
  );
  ipcMain.handle("agents:cancel", (event, id: string) => {
    const key = `${event.sender.id}:${id}`;
    runs.get(key)?.cancel();
    runs.delete(key);
  });
  app.on("before-quit", () => {
    for (const run of runs.values()) run.cancel();
    runs.clear();
  });
}
