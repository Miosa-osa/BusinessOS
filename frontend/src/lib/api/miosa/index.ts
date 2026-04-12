// MIOSA Cloud sync API client.
// All cloud interactions go through the Go backend — the frontend never calls
// api.miosa.ai directly. The API key is stored server-side only.

import { request } from "../base";

export type SyncMode = "local" | "cloud";

export interface MIOSAConnectionStatus {
  mode: SyncMode;
  connected: boolean;
  api_key_set: boolean;
  last_sync?: string; // ISO 8601
  error?: string;
}

export interface SyncResult {
  success: boolean;
  synced_at: string; // ISO 8601
  manifest_id?: string;
  error?: string;
}

/**
 * Returns current OSA mode and MIOSA Cloud connection status.
 * Makes no external network call; safe to call on every settings page load.
 */
export async function getMIOSAStatus(): Promise<MIOSAConnectionStatus> {
  return request<MIOSAConnectionStatus>("/miosa/status");
}

/**
 * Validates the API key stored in the backend against MIOSA Cloud.
 * Call this after the user saves a new API key via Settings.
 */
export async function pingMIOSACloud(): Promise<{
  connected: boolean;
  error?: string;
}> {
  return request<{ connected: boolean; error?: string }>("/miosa/ping", {
    method: "POST",
  });
}

/**
 * Pushes a WorkspaceManifest (config only, no business data) to MIOSA Cloud.
 * In local mode this is a no-op on the server and returns success immediately.
 */
export async function syncToMIOSACloud(
  workspaceId: string,
): Promise<SyncResult> {
  return request<SyncResult>("/miosa/sync", {
    method: "POST",
    body: { workspace_id: workspaceId },
  });
}
