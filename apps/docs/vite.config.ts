import { sveltekit } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    port: 1422,
  },
  preview: {
    port: 1422,
  },
});
