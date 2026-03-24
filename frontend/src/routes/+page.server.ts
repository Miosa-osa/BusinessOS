import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	// RESOLVED: Supabase auth was the original authentication provider. It has been
	// replaced by the Go backend's native JWT auth (middleware/jwt.go). The redirect
	// to /window is intentional and permanent -- the landing page now serves as a
	// passthrough to the desktop window shell. There are no Supabase credentials
	// to restore; all auth flows go through the Go backend at /api/auth/*.
	throw redirect(303, '/window');
};
