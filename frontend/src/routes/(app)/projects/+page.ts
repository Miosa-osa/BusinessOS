import type { PageLoad } from './$types';
import { api, type ClientListResponse } from '$lib/api';

/**
 * Client-side load function for the Projects page.
 *
 * Returns promises (not awaited) so SvelteKit navigates immediately
 * without blocking. The page shows a loading state while data streams in.
 * On hover, data-sveltekit-preload-data="hover" still prefetches early.
 */
export const load: PageLoad = async () => {
	// Start fetches but don't block navigation — return immediately
	const projectsPromise = api.getProjects().catch(() => []);
	const clientsPromise = api.getClients().catch((): ClientListResponse[] => []);

	// Resolve in parallel — but with a fast timeout to prevent hanging
	const [projects, clients] = await Promise.all([
		Promise.race([projectsPromise, new Promise<never[]>(r => setTimeout(() => r([]), 3000))]),
		Promise.race([clientsPromise, new Promise<ClientListResponse[]>(r => setTimeout(() => r([]), 3000))]),
	]);

	return { projects, clients };
};
