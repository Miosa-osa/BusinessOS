import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ url }) => {
	// Build redirect URL preserving all query params
	const searchParams = url.searchParams.toString();
	const redirectUrl = searchParams ? `/knowledge-v2?${searchParams}` : '/knowledge-v2';
	redirect(307, redirectUrl);
};
