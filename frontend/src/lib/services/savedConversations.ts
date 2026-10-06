import {
  getActiveWorkspaceHeaders,
  getApiBaseUrl,
  getCSRFToken,
  initCSRF,
} from "$lib/api/base";
export type SavedConversation = {
  id: string;
  title: string;
  updated_at: string;
  created_at: string;
  message_count: number;
};
export type SavedMessage = {
  id: string;
  role: string;
  content: string;
  created_at: string;
  metadata?: { runtime?: string; workspace_id?: string; model?: string };
};
async function request(path: string, method = "GET", body?: unknown) {
  if (method !== "GET") await initCSRF();
  const headers: Record<string, string> = {
    ...getActiveWorkspaceHeaders(),
    "Content-Type": "application/json",
  };
  const csrf = getCSRFToken();
  if (csrf) headers["X-CSRF-Token"] = csrf;
  const response = await fetch(`${getApiBaseUrl()}/chat${path}`, {
    method,
    headers,
    credentials: "include",
    body: body === undefined ? undefined : JSON.stringify(body),
    signal: AbortSignal.timeout(15000),
  });
  const data = await response.json().catch(() => {
    throw new Error(
      "Conversation service is temporarily unavailable. Please try again.",
    );
  });
  if (!response.ok)
    throw new Error(data.error || "Could not load or save this conversation");
  return data;
}
export async function listSavedConversations(
  page = 1,
): Promise<{ data: SavedConversation[]; pagination: { has_more: boolean } }> {
  return request(`/conversations?page=${page}&page_size=50`);
}
export async function getSavedConversation(
  id: string,
): Promise<{ conversation: SavedConversation; messages: SavedMessage[] }> {
  return request(`/conversations/${encodeURIComponent(id)}`);
}
export async function createSavedConversation(
  title: string,
): Promise<SavedConversation> {
  return request("/conversations", "POST", { title });
}
export async function appendSavedMessage(
  id: string,
  body: {
    role: string;
    content: string;
    runtime: string;
    workspace_id: string | null;
    model?: string;
  },
) {
  return request(
    `/conversations/${encodeURIComponent(id)}/messages`,
    "POST",
    body,
  );
}
export async function renameSavedConversation(id: string, title: string) {
  return request(`/conversations/${encodeURIComponent(id)}`, "PUT", { title });
}
