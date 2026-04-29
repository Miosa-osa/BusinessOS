<script lang="ts">
  import { ask, type RagResult } from '$lib/api';

  let query = $state('');
  let audience = $state('default');
  let loading = $state(false);
  let result = $state<RagResult | null>(null);
  let error = $state<string | null>(null);

  async function submit(e: Event) {
    e.preventDefault();
    if (!query.trim()) return;
    loading = true;
    error = null;
    result = null;

    try {
      result = await ask(query, { audience, format: 'markdown' });
    } catch (e) {
      error = (e as Error).message;
    } finally {
      loading = false;
    }
  }
</script>

<section class="ask">
  <header class="ask__header">
    <h1>Ask the engine</h1>
    <p>Wiki-first retrieval · falls through to hybrid search · scoped to ACLs · packed to receiver bandwidth.</p>
  </header>

  <form onsubmit={submit} class="ask__form">
    <div class="ask__row">
      <input
        type="text"
        bind:value={query}
        placeholder="e.g. healthtech pricing"
        disabled={loading}
        class="ask__input"
      />
      <select bind:value={audience} disabled={loading} class="ask__select">
        <option value="default">default</option>
        <option value="sales">sales</option>
        <option value="legal">legal</option>
        <option value="exec">exec</option>
      </select>
      <button type="submit" disabled={loading || !query.trim()} class="ask__submit">
        {loading ? 'Thinking…' : 'Ask'}
      </button>
    </div>
  </form>

  {#if error}
    <pre class="ask__error">{error}</pre>
  {/if}

  {#if result}
    <div class="ask__trace">
      <span class="ask__chip ask__chip--{result.source}">
        {result.source}
      </span>
      <span class="ask__stat">
        <small>candidates</small>{result.trace.n_candidates}
      </span>
      <span class="ask__stat">
        <small>delivered</small>{result.trace.n_delivered}
      </span>
      <span class="ask__stat">
        <small>elapsed</small>{result.trace.elapsed_ms} ms
      </span>
      {#if result.trace['wiki_hit?']}
        <span class="ask__chip ask__chip--hit">wiki hit</span>
      {/if}
    </div>

    <article class="ask__envelope">
      <pre>{result.envelope.body}</pre>
    </article>

    {#if result.envelope.sources.length > 0}
      <section class="ask__sources">
        <h3>Sources</h3>
        <ul>
          {#each result.envelope.sources as uri}
            <li><code>{uri}</code></li>
          {/each}
        </ul>
      </section>
    {/if}
  {/if}
</section>

<style>
  .ask {
    max-width: 860px;
    margin: 0 auto;
    padding: 2rem 1.5rem;
  }

  .ask__header h1 {
    margin: 0 0 0.35rem;
    font-size: 1.15rem;
    color: #eee;
  }

  .ask__header p {
    margin: 0 0 1.5rem;
    color: #888;
    font-size: 0.85rem;
  }

  .ask__form {
    margin-bottom: 1rem;
  }

  .ask__row {
    display: flex;
    gap: 0.5rem;
  }

  .ask__input {
    flex: 1;
    background: #0e1116;
    color: #eee;
    border: 1px solid #1e232c;
    border-radius: 6px;
    padding: 0.6rem 0.85rem;
    font-size: 0.92rem;
  }
  .ask__input:focus {
    outline: none;
    border-color: #375;
  }

  .ask__select,
  .ask__submit {
    background: #0e1116;
    color: #eee;
    border: 1px solid #1e232c;
    border-radius: 6px;
    padding: 0.6rem 0.85rem;
    font-size: 0.88rem;
  }

  .ask__submit {
    background: #1b3429;
    border-color: #265e45;
    color: #6f8;
    font-weight: 600;
    cursor: pointer;
  }
  .ask__submit:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .ask__error {
    color: #f88;
    background: #311;
    border-radius: 6px;
    padding: 0.75rem;
    margin: 1rem 0;
  }

  .ask__trace {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    padding: 0.5rem 0.75rem;
    background: #0e1116;
    border: 1px solid #1a1e25;
    border-radius: 6px;
    margin-bottom: 0.75rem;
    font-size: 0.78rem;
    flex-wrap: wrap;
  }

  .ask__chip {
    padding: 2px 10px;
    border-radius: 999px;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  .ask__chip--wiki {
    background: #143;
    color: #6f8;
  }
  .ask__chip--chunks {
    background: #223;
    color: #8af;
  }
  .ask__chip--empty {
    background: #332;
    color: #fc6;
  }
  .ask__chip--hit {
    background: #1c2a3a;
    color: #8af;
  }

  .ask__stat {
    display: inline-flex;
    align-items: baseline;
    gap: 0.35rem;
    color: #ddd;
  }
  .ask__stat small {
    color: #666;
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .ask__envelope {
    background: #0e1116;
    border: 1px solid #1a1e25;
    border-radius: 6px;
    padding: 1rem 1.25rem;
  }
  .ask__envelope pre {
    margin: 0;
    white-space: pre-wrap;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    font-size: 0.93rem;
    line-height: 1.55;
    color: #ddd;
  }

  .ask__sources {
    margin-top: 1rem;
    padding: 1rem 1.25rem;
    background: #0e1116;
    border: 1px solid #1a1e25;
    border-radius: 6px;
  }
  .ask__sources h3 {
    margin: 0 0 0.5rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    color: #888;
  }
  .ask__sources ul {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .ask__sources li {
    padding: 0.2rem 0;
  }
  .ask__sources code {
    color: #8af;
    font-size: 0.82rem;
    font-family: 'SF Mono', Menlo, monospace;
  }
</style>
