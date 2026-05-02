import { createEngine, type EngineContext } from "@miosa/optimal-engine";
import { browser } from "$app/environment";

let instance: EngineContext | null = null;

function resolveBaseUrl(): string {
  if (!browser) return "http://localhost:4200";
  const isDev =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";
  if (isDev) return "http://localhost:4200";
  return `${window.location.origin}/api/v1/optimal-engine`;
}

export function getEngine(): EngineContext {
  if (!instance) {
    instance = createEngine({
      baseUrl: resolveBaseUrl(),
      workspace: "default",
    });
  }
  return instance;
}

export function resetEngine(): void {
  instance = null;
}
