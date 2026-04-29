<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { status, type StatusPayload } from '$lib/api';
  import { theme, toggle } from '$lib/theme';

  let engineStatus = $state<StatusPayload['status']>('down');
  let { children } = $props<{ children?: () => any }>();

  // Routes that want the full viewport (no padding, no max-width column).
  const FULL_BLEED = new Set(['/graph', '/workspace', '/activity']);
  let fullBleed = $derived(FULL_BLEED.has($page.url.pathname));

  onMount(async () => {
    try {
      const s = await status();
      engineStatus = s.status;
    } catch {
      engineStatus = 'down';
    }
  });
</script>

<header>
  <h1>Optimal Engine</h1>
  <nav>
    <a href="/" class:active={$page.url.pathname === '/'}>Ask</a>
    <a href="/workspace" class:active={$page.url.pathname.startsWith('/workspace')}>Workspace</a>
    <a href="/graph" class:active={$page.url.pathname === '/graph'}>Graph</a>
    <a href="/wiki" class:active={$page.url.pathname.startsWith('/wiki')}>Wiki</a>
    <a href="/architectures" class:active={$page.url.pathname.startsWith('/architectures')}>Architectures</a>
    <a href="/activity" class:active={$page.url.pathname === '/activity'}>Activity</a>
    <a href="/status" class:active={$page.url.pathname === '/status'}>Status</a>
  </nav>
  <button
    class="theme-toggle"
    onclick={toggle}
    aria-label="Toggle theme"
    title={$theme === 'dark' ? 'Switch to light' : 'Switch to dark'}
  >
    {#if $theme === 'dark'}☀︎{:else}☾{/if}
  </button>
  <span class="badge badge-{engineStatus}">{engineStatus}</span>
</header>

<main class:main--full={fullBleed}>
  {@render children?.()}
</main>

<style>
  header {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.65rem 1.25rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg);
  }
  header h1 {
    font-size: 0.95rem;
    margin: 0;
    color: var(--text);
    font-weight: 600;
    letter-spacing: 0.2px;
  }
  nav {
    display: flex;
    gap: 0.15rem;
    flex: 1;
    margin-left: 0.75rem;
  }
  nav a {
    color: var(--text-muted);
    text-decoration: none;
    font-size: 0.85rem;
    padding: 4px 10px;
    border-radius: 6px;
    transition: background 0.12s, color 0.12s;
  }
  nav a:hover {
    color: var(--text);
    background: var(--bg-elevated-2);
  }
  nav a.active {
    color: var(--accent);
    background: var(--accent-soft);
  }
  .theme-toggle {
    background: none;
    border: 1px solid var(--border);
    color: var(--text-muted);
    width: 30px;
    height: 30px;
    border-radius: 6px;
    cursor: pointer;
    font-size: 0.9rem;
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .theme-toggle:hover {
    color: var(--text);
    background: var(--bg-elevated-2);
  }
  .badge {
    padding: 2px 10px;
    border-radius: 999px;
    font-size: 0.72rem;
    font-weight: 600;
    text-transform: lowercase;
  }
  .badge-up { background: var(--good-bg); color: var(--good); }
  .badge-degraded { background: var(--warn-bg); color: var(--warn); }
  .badge-down { background: var(--bad-bg); color: var(--bad); }

  main {
    padding: 1.5rem;
    max-width: 960px;
    margin: 0 auto;
  }
  main.main--full {
    padding: 0;
    max-width: none;
    margin: 0;
    height: calc(100vh - 48px);
    width: 100%;
  }
</style>
