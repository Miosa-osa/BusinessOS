<script lang="ts">
import { osaStore, type OsaModelCatalog } from "$lib/stores/osa";
import { onDestroy, tick } from "svelte";
let {
  class: className = "",
  compact = false,
}: { class?: string; compact?: boolean } = $props();
let isOpen = $state(false);
let root: HTMLDivElement;
let trigger: HTMLButtonElement;
let panel = $state<HTMLDivElement>();
let catalog = $state<OsaModelCatalog | null>(null);
let stagedProvider = $state("");
let stagedModel = $state("");
let search = $state("");
let loading = $state(false);
let applying = $state(false);
let error = $state("");
let applied = $state(false);
let left = $state(8);
let bottom = $state(80);
let epoch = 0;
const localRuntime = $derived(
  ["ollama", "hermes"].includes($osaStore.activeRuntime),
);
const activeModel = $derived(
  localRuntime
    ? ($osaStore.localModel ??
        ($osaStore.activeProvider === "ollama" ? $osaStore.activeModel : null))
    : $osaStore.activeModel,
);
const providers = $derived(
  [...(catalog?.providers ?? [])]
    .filter((p) => !localRuntime || p.slug === "ollama")
    .sort(
      (a, b) =>
        Number(b.configured) - Number(a.configured) ||
        a.name.localeCompare(b.name),
    ),
);
const selectedProvider = $derived(
  providers.find((p) => p.slug === stagedProvider),
);
const models = $derived(
  (catalog?.models ?? []).filter(
    (m) =>
      m.provider === stagedProvider &&
      m.name.toLowerCase().includes(search.toLowerCase()),
  ),
);
function providerName(name: string) {
  return name.replace(/_/g, " ").replace(/\bCli\b/gi, "CLI");
}
function position() {
  if (!trigger) return;
  const rect = trigger.getBoundingClientRect();
  left = Math.max(8, Math.min(rect.left, window.innerWidth - 356));
  bottom = Math.max(8, window.innerHeight - rect.top + 10);
}
async function refresh() {
  const request = ++epoch;
  loading = true;
  error = "";
  applied = false;
  catalog = null;
  try {
    const data = await osaStore.loadModelCatalog();
    if (request !== epoch) return;
    catalog = data;
    stagedProvider = localRuntime ? "ollama" : data.provider;
    stagedModel = activeModel ?? "";
    search = "";
  } catch (e) {
    if (request === epoch)
      error = e instanceof Error ? e.message : "Could not load OSA models";
  } finally {
    if (request === epoch) loading = false;
  }
}
async function open() {
  isOpen = true;
  position();
  void refresh();
  await tick();
  panel?.focus();
}
function close() {
  isOpen = false;
  ++epoch;
  trigger?.focus();
}
async function apply() {
  if (
    applying ||
    !catalog ||
    !selectedProvider?.configured ||
    !stagedModel.trim()
  )
    return;
  const provider = stagedProvider,
    model = stagedModel.trim();
  applying = true;
  error = "";
  applied = false;
  try {
    if (localRuntime) osaStore.setLocalModel(model);
    else await osaStore.setModel(provider, model);
    if (catalog) catalog = { ...catalog, provider, current: model };
    applied = true;
  } catch (e) {
    error = e instanceof Error ? e.message : "OSA could not apply this model";
  } finally {
    applying = false;
  }
}
function outside(e: PointerEvent) {
  if (
    isOpen &&
    root &&
    !root.contains(e.target as Node) &&
    !panel?.contains(e.target as Node)
  )
    close();
}
function mountPanel(node: HTMLDivElement) {
  document.body.appendChild(node);
  return {
    destroy() {
      node.remove();
    },
  };
}
function keydown(e: KeyboardEvent) {
  if (e.key === "Escape" && isOpen) {
    e.stopPropagation();
    close();
  }
}
onDestroy(() => {
  ++epoch;
});
</script>

<svelte:window onpointerdown={outside} onkeydown={keydown} onresize={position} />
<div bind:this={root} class="model-selector {className}">
  <button type="button" bind:this={trigger} class="model-trigger" role="combobox" aria-haspopup="dialog" aria-expanded={isOpen} aria-controls="osa-model-panel" aria-label="Select model: {activeModel ?? 'none selected'}" title={activeModel ?? 'Select OSA model'} onclick={()=>isOpen?close():open()}>
    <svg class="model-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" aria-hidden="true"><rect x="4" y="4" width="16" height="16" rx="3"/><rect x="9" y="9" width="6" height="6"/><path d="M9 1v3m6-3v3M9 20v3m6-3v3M1 9h3m-3 6h3m16-6h3m-3 6h3"/></svg>
    {#if !compact}<span class="model-label">{activeModel ?? 'Model'}</span>{/if}
  </button>
  {#if isOpen}
    <div use:mountPanel bind:this={panel} id="osa-model-panel" class="model-dropdown" role="dialog" aria-label="OSA model configuration" tabindex="-1" style:left="{left}px" style:bottom="{bottom}px">
      <header><div><strong>{localRuntime?'Ollama model':'OSA model'}</strong><p>{localRuntime?'Used by your local agent runtime':'Live from your OSA runtime'}</p></div><button type="button" class="icon-button" aria-label="Close model picker" onclick={close}>×</button></header>
      {#if loading}<p role="status">Loading OSA providers and models…</p>
      {:else if catalog}
        <div class="current"><span>Current model</span><strong>{activeModel || 'None selected'}</strong></div>
        <label for="osa-provider">Provider</label>
        <select id="osa-provider" value={stagedProvider} disabled={applying || localRuntime} onchange={e=>{stagedProvider=e.currentTarget.value;stagedModel='';search='';applied=false;error='';}}>
          {#each providers as provider}<option value={provider.slug}>{providerName(provider.name)}{provider.configured?'':' · Not configured'}</option>{/each}
        </select>
        {#if selectedProvider && !selectedProvider.configured}<p class="notice">Configure this provider in OSA before using its models.</p>{/if}
        <label for="osa-model-search">Models</label>
        <input id="osa-model-search" type="search" placeholder="Search OSA models" bind:value={search} disabled={applying}/>
        <div class="models" aria-label="Models from OSA">
          {#each models as model}
            <button type="button" class="model-option" class:selected={stagedModel===model.name} disabled={applying || !selectedProvider?.configured} aria-pressed={stagedModel===model.name} onclick={()=>{stagedModel=model.name;applied=false;error='';}}>{model.name}</button>
          {:else}<p class="notice">{search?'No matching models.':'OSA has not listed models for this provider.'}</p>{/each}
        </div>
        <label for="osa-model-name">Model name</label>
        <input id="osa-model-name" value={stagedModel} disabled={applying || !selectedProvider?.configured} placeholder="Exact model ID" oninput={e=>{stagedModel=e.currentTarget.value;applied=false;}}/>
        <button type="button" class="apply-btn" disabled={applying || !selectedProvider?.configured || !stagedModel.trim()} onclick={apply}>{applying?'Applying…':'Apply model'}</button>
        {#if applied}<p class="success" role="status">Model applied</p>{/if}
      {/if}
      {#if error}<p class="error" role="alert">{error}</p>{/if}
      <footer><span>{localRuntime?'Uses your local Ollama service.':'Provider settings are managed in OSA.'}</span><button type="button" disabled={loading || applying} onclick={refresh}>Refresh</button></footer>
    </div>
  {/if}
</div>

<style>
.model-selector {
  position: relative;
  flex-shrink: 0;
}
.model-trigger {
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 180px;
  padding: 4px 8px;
  border: 1px solid #dedee3;
  border-radius: 8px;
  background: #f5f5f7;
  color: #333;
  cursor: pointer;
}
.model-icon {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
}
.model-label {
  font-size: 11px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.model-dropdown {
  position: fixed;
  z-index: 10100;
  width: 340px;
  max-width: calc(100vw - 16px);
  max-height: min(620px, calc(100vh - 150px));
  overflow: auto;
  box-sizing: border-box;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 9px;
  background: #fff;
  color: #202124;
  border: 1px solid #d9dce2;
  border-radius: 14px;
  box-shadow: 0 16px 48px #0003;
  isolation: isolate;
  text-align: left;
  cursor: default;
}
header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}
header strong {
  font-size: 15px;
}
p {
  margin: 0;
  font-size: 12px;
  line-height: 1.45;
}
header p,
footer,
.notice {
  color: #656a74;
}
.icon-button {
  border: 0;
  background: transparent;
  color: inherit;
  font-size: 22px;
  cursor: pointer;
  padding: 0 4px;
}
.current {
  padding: 10px 12px;
  background: #f3f5f8;
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.current span {
  font-size: 11px;
  color: #656a74;
}
.current strong {
  font-size: 12px;
  overflow-wrap: anywhere;
}
label {
  font-size: 12px;
  font-weight: 600;
  margin-top: 3px;
}
select,
input {
  width: 100%;
  box-sizing: border-box;
  padding: 9px 10px;
  border: 1px solid #cbd0d8;
  border-radius: 7px;
  background: #fff;
  color: #202124;
  font-size: 12px;
}
.models {
  max-height: 172px;
  min-height: 40px;
  overflow: auto;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.model-option {
  padding: 8px 10px;
  text-align: left;
  border: 1px solid transparent;
  border-radius: 6px;
  background: #f6f7f9;
  color: inherit;
  font:
    12px ui-monospace,
    monospace;
  overflow-wrap: anywhere;
  cursor: pointer;
  flex-shrink: 0;
}
.model-option:hover:not(:disabled) {
  background: #e9eef6;
}
.model-option.selected {
  background: #e8f0fe;
  border-color: #92b7f3;
  color: #174b9b;
}
.apply-btn {
  padding: 10px;
  border: 0;
  border-radius: 7px;
  background: #202124;
  color: #fff;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}
button:disabled {
  opacity: 0.5;
  cursor: default;
}
.success {
  color: #16753d;
}
.error {
  color: #b42318;
  background: #fff0ef;
  padding: 8px;
  border-radius: 6px;
}
footer {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
  font-size: 10px;
  border-top: 1px solid #e5e7eb;
  padding-top: 9px;
}
footer button {
  border: 0;
  background: transparent;
  color: #245eab;
  cursor: pointer;
  font-size: 11px;
}
button:focus-visible,
input:focus-visible,
select:focus-visible {
  outline: 2px solid #3478df;
  outline-offset: 2px;
}
:global(.dark) .model-dropdown {
  background: #202226;
  color: #f2f3f5;
  border-color: #41454d;
}
:global(.dark) .current,
:global(.dark) .model-option {
  background: #2c3037;
}
:global(.dark) select,
:global(.dark) input {
  background: #25282d;
  color: #f2f3f5;
  border-color: #555b65;
}
:global(.dark) .model-option.selected {
  background: #173964;
  color: #dceaff;
  border-color: #5489d4;
}
:global(.dark) .apply-btn {
  background: #e7edf7;
  color: #202124;
}
:global(.dark) .notice,
:global(.dark) header p,
:global(.dark) footer,
:global(.dark) .current span {
  color: #b3bbc8;
}
</style>
