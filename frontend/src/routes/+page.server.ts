// Root route — routing is handled client-side in +page.svelte
// (Static adapter / Cloudflare Pages: no server runtime, auth is client-side)
export const load = async () => {
  return {};
};
