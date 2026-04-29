<script lang="ts">
  import { onMount } from 'svelte';
  import { status, type StatusPayload } from '$lib/api';

  let payload: StatusPayload | null = null;
  let error: string | null = null;

  onMount(async () => {
    try {
      payload = await status();
    } catch (e) {
      error = (e as Error).message;
    }
  });
</script>

<h2>Status</h2>

{#if error}
  <pre class="error">{error}</pre>
{:else if !payload}
  <p>Loading…</p>
{:else}
  <p>Overall: <strong>{payload.status}</strong></p>

  <table>
    <thead>
      <tr>
        <th>Check</th>
        <th>Result</th>
      </tr>
    </thead>
    <tbody>
      {#each Object.entries(payload.checks) as [name, value]}
        <tr>
          <td><code>{name}</code></td>
          <td>{value}</td>
        </tr>
      {/each}
    </tbody>
  </table>

  {#if payload.degraded.length > 0}
    <p class="warn">Degraded: {payload.degraded.join(', ')}</p>
  {/if}
{/if}

<style>
  table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 0.5rem;
  }
  th,
  td {
    border-bottom: 1px solid #222;
    padding: 0.5rem 0.75rem;
    text-align: left;
  }
  th {
    color: #888;
    font-weight: 500;
    font-size: 0.85rem;
  }
  code {
    color: #8af;
  }
  .warn {
    color: #fc6;
  }
  .error {
    color: #f88;
  }
</style>
