import { redirect } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, fetch }) => {
  const sessionCookie = cookies.get("better-auth.session_token");
  if (sessionCookie) {
    // Validate the session is actually valid before redirecting
    try {
      const res = await fetch("http://localhost:8001/api/auth/session", {
        headers: { Cookie: `better-auth.session_token=${sessionCookie}` },
      });
      if (res.ok) {
        const data = await res.json();
        if (data?.user) {
          throw redirect(303, "/window");
        }
      }
    } catch (err) {
      // If it's a redirect, let it through
      if (err && typeof err === "object" && "status" in err) throw err;
    }
    // Invalid session — clear the stale cookie
    cookies.delete("better-auth.session_token", { path: "/" });
  }
  // Show the landing page
};
