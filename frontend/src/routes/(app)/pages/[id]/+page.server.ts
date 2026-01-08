import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ params, url }) => {
	// Redirect to knowledge-v2/[id] preserving query params
	const searchParams = url.searchParams.toString();
	const redirectUrl = searchParams
		? `/knowledge-v2/${params.id}?${searchParams}`
		: `/knowledge-v2/${params.id}`;
	redirect(307, redirectUrl);
};
