import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	// DEV BYPASS: go straight to window desktop
	// TODO: Remove when Supabase credentials are restored
	throw redirect(303, '/window');
};
