import { describe, expect, it } from "vitest";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

import { getDesktopModuleIds } from "$lib/config/workspaceModules";
import { BUILTIN_MODULES } from "$lib/stores/desktop3d/types";
import { MODULE_INFO } from "$lib/stores/desktop3d/moduleRegistry";
import { getModuleWindow } from "./moduleWindowRegistry";

describe("desktop module registry", () => {
  it("has metadata and a live route for every built-in desktop module", () => {
    for (const module of BUILTIN_MODULES) {
      expect(MODULE_INFO[module], `${module} metadata`).toBeDefined();
      expect(getModuleWindow(module), `${module} window route`).toBeDefined();
    }
  });

  it.each(["primitive", "agency-miosa", "miosa"] as const)(
    "can render every module enabled by the %s profile",
    (moduleProfile) => {
      for (const module of getDesktopModuleIds({ module_profile: moduleProfile })) {
        expect(MODULE_INFO[module], `${module} metadata`).toBeDefined();
        expect(getModuleWindow(module), `${module} window route`).toBeDefined();
      }
    },
  );

  it("routes Analytics to the Analytics module", () => {
    expect(getModuleWindow("analytics")?.url).toBe("/analytics");
  });

  it("backs every workspace-profile module with a SvelteKit page", () => {
    const routesRoot = resolve(process.cwd(), "src/routes/(app)");
    const profileModules = new Set([
      ...getDesktopModuleIds({ module_profile: "primitive" }),
      ...getDesktopModuleIds({ module_profile: "agency-miosa" }),
      ...getDesktopModuleIds({ module_profile: "miosa" }),
    ]);

    for (const module of profileModules) {
      const definition = getModuleWindow(module);
      expect(definition, `${module} window definition`).toBeDefined();
      const pathname = new URL(definition!.url, "http://businessos.local").pathname;
      const routeDirectory = `${routesRoot}${pathname}`;
      const exists = existsSync(`${routeDirectory}/+page.svelte`)
        || existsSync(`${routeDirectory}/+page.ts`);
      expect(exists, `${module} route at ${pathname}`).toBe(true);
    }
  });
});
