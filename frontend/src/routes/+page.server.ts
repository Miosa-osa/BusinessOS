import { redirect } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
  // If user has a session, go to desktop
  const sessionCookie = cookies.get("better-auth.session_token");
  if (sessionCookie) {
    throw redirect(303, "/window");
  }
  // Otherwise show the landing page
};
