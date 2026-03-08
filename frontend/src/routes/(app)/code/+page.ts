import type { PageLoad } from "./$types";
import { BUILTIN_MODULES } from "$lib/stores/desktop3dStore";
import { MODULE_INFO } from "$lib/stores/desktop3dStore";

/**
 * Client-side load function for the Code Browser page.
 *
 * Provides the list of built-in modules so the sidebar can render them
 * immediately without an async API call. Dynamic/custom modules loaded
 * from the API are fetched on the client after mount.
 */
export const load: PageLoad = async () => {
  // Build a stable module list from the registry
  const modules = [...BUILTIN_MODULES].map((id) => ({
    id,
    title: MODULE_INFO[id]?.title ?? id,
    color: MODULE_INFO[id]?.color ?? "#8B5CF6",
    icon: MODULE_INFO[id]?.icon ?? "box",
  }));

  return { modules };
};
