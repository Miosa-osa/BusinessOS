import type { LayoutLoad } from "./$types";
import { redirect } from "@sveltejs/kit";
import { getLegacyApiBaseUrl } from "$lib/config/runtime";

// Client-side auth check for the (app) group.
// Replaces the deleted +layout.server.ts (which broke static/Cloudflare builds).
export const load: LayoutLoad = async ({ fetch, url }) => {
  const isEmbed = url.searchParams.get("embed") === "true";

  if (isEmbed) {
    return { user: null, session: null };
  }

  if (
    typeof localStorage !== "undefined" &&
    localStorage.getItem("businessos_logged_out") === "1"
  ) {
    throw redirect(302, "/login");
  }

  // Validate the HttpOnly session cookie with the backend.
  // Browser JavaScript cannot read the cookie directly.
  try {
    const response = await fetch(`${getLegacyApiBaseUrl()}/auth/session`, {
      method: "GET",
      credentials: "include",
    });

    if (!response.ok) {
      // No valid session — send to login, preserving the intended destination.
      const returnTo = encodeURIComponent(url.pathname + url.search);
      throw redirect(302, `/login?next=${returnTo}`);
    }

    const contentType = response.headers.get("content-type") || "";
    if (!contentType.includes("application/json")) {
      throw redirect(302, "/login");
    }

    const data = await response.json();

    if (!data?.user) {
      throw redirect(302, "/login");
    }

    return {
      user: data.user,
      session: data.session || { id: "active" },
    };
  } catch (err) {
    // Re-throw SvelteKit redirects so they are handled correctly.
    if (
      err instanceof Response ||
      (err != null && typeof err === "object" && "status" in err)
    ) {
      throw err;
    }
    // Network / parse error — redirect to login rather than crash.
    throw redirect(302, "/login");
  }
};

// This runs client-side, not server-side.
export const ssr = false;
