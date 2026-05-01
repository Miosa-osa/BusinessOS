import type { LayoutLoad } from "./$types";
import { getBackendUrl } from "$lib/config/runtime";

// Client-side auth check for the (app) group
// Replaces the deleted +layout.server.ts (which broke static/Cloudflare builds)
export const load: LayoutLoad = async ({ fetch, url }) => {
  const isEmbed = url.searchParams.get("embed") === "true";

  // Check for session cookie
  const hasSession =
    typeof document !== "undefined" &&
    document.cookie.includes("better-auth.session_token");

  if (!hasSession && !isEmbed) {
    // No redirect here — let the page handle it client-side
    return {
      user: null,
      session: null,
    };
  }

  // Validate session with backend
  try {
    const response = await fetch(`${getBackendUrl()}/api/auth/session`, {
      method: "GET",
      credentials: "include",
    });

    if (!response.ok) {
      return { user: null, session: null };
    }

    const contentType = response.headers.get("content-type") || "";
    if (!contentType.includes("application/json")) {
      return { user: null, session: null };
    }

    const data = await response.json();

    if (data?.user) {
      return {
        user: data.user,
        session: data.session || { id: "active" },
      };
    }

    return { user: null, session: null };
  } catch {
    return { user: null, session: null };
  }
};

// This runs client-side, not server-side
export const ssr = false;
