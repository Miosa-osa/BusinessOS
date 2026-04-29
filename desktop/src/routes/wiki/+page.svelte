<script lang="ts">
  import { onMount } from 'svelte';
  import { listWiki, getWiki, type WikiPageSummary } from '$lib/api';

  let pages: WikiPageSummary[] = [];
  let selected: string | null = null;
  let body = '';
  let error: string | null = null;

  onMount(async () => {
    try {
      const res = await listWiki();
      pages = res.pages;
    } catch (e) {
      error = (e as Error).message;
    }
  });

  async function open(slug: string) {
    selected = slug;
    body = '';

    try {
      const res = await getWiki(slug);
      body = res.body;
    } catch (e) {
      error = (e as Error).message;
    }
  }
</script>

<h2>Wiki</h2>

{#if error}
  <pre class="error">{error}</pre>
{/if}

<div class="grid">
  <aside>
    {#if pages.length === 0}
      <p class="muted">No pages yet.</p>
    {:else}
      <ul>
        {#each pages as p}
          <li class:selected={selected === p.slug}>
            <button on:click={() => open(p.slug)}>
              <strong>{p.slug}</strong>
              <span class="meta">
                v{p.version} · {p.audience} · {p.size_bytes}b
              </span>
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </aside>

  <article>
    {#if selected && body}
      <pre>{body}</pre>
    {:else if selected}
      <p class="muted">Loading {selected}…</p>
    {:else}
      <p class="muted">Pick a page to render.</p>
    {/if}
  </article>
</div>

<style>
  .grid {
    display: grid;
    grid-template-columns: 260px 1fr;
    gap: 1rem;
  }
  aside ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }
  aside li button {
    width: 100%;
    background: none;
    border: none;
    color: #ddd;
    text-align: left;
    padding: 0.5rem 0.75rem;
    border-radius: 4px;
    cursor: pointer;
    font-size: 0.9rem;
  }
  aside li.selected button,
  aside li button:hover {
    background: #151a21;
  }
  .meta {
    display: block;
    color: #888;
    font-size: 0.75rem;
    margin-top: 2px;
  }
  article pre {
    background: #0e1116;
    border: 1px solid #222;
    border-radius: 4px;
    padding: 1rem;
    white-space: pre-wrap;
    font-family: 'SF Mono', Menlo, monospace;
    font-size: 0.88rem;
    line-height: 1.5;
  }
  .muted {
    color: #888;
  }
  .error {
    color: #f88;
  }
</style>
