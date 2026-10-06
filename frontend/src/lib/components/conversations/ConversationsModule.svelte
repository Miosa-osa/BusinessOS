<script lang="ts">
import { onMount, tick } from "svelte";
import {
  ArrowUp,
  Check,
  Copy,
  MessageSquare,
  Mic,
  PanelLeft,
  Pencil,
  Plus,
  RefreshCw,
  Search,
  Square,
  X,
  AudioLines,
} from "lucide-svelte";
import { osaStore, AGENT_RUNTIME_OPTIONS } from "$lib/stores/osa";
import {
  listSavedConversations,
  renameSavedConversation,
  type SavedConversation,
} from "$lib/services/savedConversations";
import { OsaOrbVoice, type OrbVoicePhase } from "$lib/services/osaOrbVoice";
import { renderMarkdown } from "$lib/utils/markdownRenderer";
import AgentRuntimeSelector from "$lib/components/desktop/osa/AgentRuntimeSelector.svelte";
import ModelSelector from "$lib/components/desktop/osa/ModelSelector.svelte";

let sessions = $state<SavedConversation[]>([]),
  search = $state(""),
  draft = $state(""),
  error = $state("");
let loading = $state(false),
  opening = $state(false),
  hasMore = $state(false),
  page = $state(1),
  renaming = $state(false),
  title = $state(""),
  historyOpen = $state(true),
  copied = $state("");
let thread: HTMLDivElement, input: HTMLTextAreaElement;
let followEnd = true,
  copyTimer: ReturnType<typeof setTimeout>;
let voice: OsaOrbVoice,
  voicePhase = $state<OrbVoicePhase>("idle"),
  transcript = $state(""),
  voiceCaption = $state(""),
  voiceError = $state("");
const selected = $derived(
  sessions.find((s) => s.id === $osaStore.conversationId),
);
const filtered = $derived(
  sessions.filter((s) => s.title.toLowerCase().includes(search.toLowerCase())),
);
const runtimeLabel = $derived(
  AGENT_RUNTIME_OPTIONS.find((r) => r.id === $osaStore.activeRuntime)?.label ??
    "OSA",
);
const empty = $derived(
  !$osaStore.conversation.length && !$osaStore.isStreaming && !opening,
);
const voiceActive = $derived(!["idle", "error"].includes(voicePhase));
const voiceLabel = $derived(
  {
    starting: "Opening microphone",
    listening: "Listening",
    transcribing: "Transcribing",
    thinking: "OSA is thinking",
    speaking: "OSA is speaking",
    idle: "Voice conversation",
    error: "Voice unavailable",
  }[voicePhase],
);
function dateGroup(date: string) {
  const d = new Date(date),
    now = new Date();
  if (d.toDateString() === now.toDateString()) return "Today";
  now.setDate(now.getDate() - 1);
  if (d.toDateString() === now.toDateString()) return "Yesterday";
  return d.toLocaleDateString([], { month: "long", year: "numeric" });
}
async function refresh(more = false) {
  loading = true;
  error = "";
  try {
    const result = await listSavedConversations(more ? page + 1 : 1);
    sessions = more ? [...sessions, ...result.data] : result.data;
    page = more ? page + 1 : 1;
    hasMore = result.pagination.has_more;
  } catch (e) {
    error = e instanceof Error ? e.message : "Could not load conversations";
  } finally {
    loading = false;
  }
}
async function scrollEnd() {
  await tick();
  if (thread) thread.scrollTop = thread.scrollHeight;
}
async function open(id: string) {
  if ($osaStore.isStreaming) return;
  voice?.cancel();
  opening = true;
  error = "";
  renaming = false;
  try {
    await osaStore.loadConversation(id);
    osaStore.setExpanded(false);
    followEnd = true;
  } catch (e) {
    error = e instanceof Error ? e.message : "Could not open conversation";
  } finally {
    opening = false;
    void scrollEnd();
  }
}
function create() {
  voice?.cancel();
  osaStore.clearConversation();
  osaStore.setExpanded(false);
  draft = "";
  renaming = false;
  followEnd = true;
  input?.focus();
}
async function send() {
  if (!draft.trim() || $osaStore.isStreaming || opening) return;
  voice?.cancel();
  const text = draft.trim();
  draft = "";
  followEnd = true;
  await osaStore.sendMessage(text, { suppressPopup: true });
  await refresh();
  void scrollEnd();
  input?.focus();
}
async function rename() {
  if (!selected || !title.trim()) return;
  try {
    await renameSavedConversation(selected.id, title.trim());
    renaming = false;
    await refresh();
  } catch (e) {
    error = e instanceof Error ? e.message : "Could not rename conversation";
  }
}
async function copy(id: string, text: string) {
  try {
    await navigator.clipboard.writeText(text);
    copied = id;
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => (copied = ""), 1800);
  } catch {
    error = "Could not copy this message.";
  }
}
$effect(() => {
  const content = $osaStore.streamingContent;
  const count = $osaStore.conversation.length;
  if (followEnd && (content || count)) void scrollEnd();
});
onMount(() => {
  const notifyParent = (active: boolean) => {
    if(window.parent !== window) window.parent.postMessage({type:'businessos:conversation-view',active}, window.location.origin === 'null' ? '*' : window.location.origin);
  };
  notifyParent(true);
  void refresh();
  void osaStore.restoreConversation();
  void osaStore.loadHealth();
  void osaStore.loadModelCatalog().catch(() => {});
  voice = new OsaOrbVoice({
    phase: (p) => {
      voicePhase = p;
      if (p === "starting") {
        voiceError = "";
        voiceCaption = "";
      }
      if (p === "idle") void refresh();
    },
    transcript: (text) => (transcript = text),
    caption: (_, text) => (voiceCaption = text),
    error: (message) => (voiceError = message),
  });
  return () => {
    notifyParent(false);
    voice.destroy();
    clearTimeout(copyTimer);
  };
});
</script>

<section class="conversations" class:history-hidden={!historyOpen} aria-label="Conversations">
  <aside aria-label="Saved conversations">
    <div class="history-heading"><h1>Conversations</h1><button class="icon-button" onclick={()=>historyOpen=false} aria-label="Hide conversation history" title="Hide history"><PanelLeft size={17}/></button></div>
    <button class="new-conversation" onclick={create} disabled={$osaStore.isStreaming}><Plus size={17}/><span>New conversation</span><span class="new-hint">＋</span></button>
    <label class="search"><Search size={15}/><input type="search" placeholder="Search conversations" aria-label="Search conversations" bind:value={search}/></label>
    <nav class="session-list" aria-label="Conversation history">
      {#each filtered as session,i (session.id)}
        {#if i===0 || dateGroup(session.updated_at)!==dateGroup(filtered[i-1].updated_at)}<p class="date-group">{dateGroup(session.updated_at)}</p>{/if}
        <button class="session" class:active={session.id===$osaStore.conversationId} aria-current={session.id===$osaStore.conversationId?'true':undefined} disabled={$osaStore.isStreaming||opening} onclick={()=>open(session.id)} title={session.title}>
          <MessageSquare size={15}/><span><strong>{session.title}</strong><small>{session.message_count} {session.message_count===1?'message':'messages'}</small></span>
        </button>
      {:else}<p class="hint">{loading?'Loading conversations…':search?'No matching conversations.':'Your conversations will be saved here.'}</p>{/each}
      {#if hasMore}<button class="load-more" disabled={loading} onclick={()=>refresh(true)}>Load more conversations</button>{/if}
    </nav>
    <div class="history-footer"><span><span class="saved-dot"></span> Saved to BusinessOS</span><button class="icon-button" aria-label="Refresh history" title="Refresh history" disabled={loading} onclick={()=>refresh()}><RefreshCw size={15} class={loading?'spin':''}/></button></div>
  </aside>
  <main>
    <header class="thread-header">
      <div class="thread-title">{#if !historyOpen}<button class="icon-button" onclick={()=>historyOpen=true} aria-label="Show conversation history"><PanelLeft size={18}/></button>{/if}<div class="osa-avatar"><video src="/OSAFinalNOBG.mp4" autoplay loop muted playsinline aria-hidden="true"></video></div><div class="title-text">
        {#if renaming}<form class="rename-form" onsubmit={e=>{e.preventDefault();void rename();}}><input aria-label="Conversation title" bind:value={title}/><button class="icon-button" aria-label="Save title"><Check size={17}/></button><button type="button" class="icon-button" aria-label="Cancel rename" onclick={()=>renaming=false}><X size={17}/></button></form>
        {:else}<h2>{selected?.title ?? 'New conversation'}</h2><p>{runtimeLabel}{ $osaStore.activeRuntime==='hermes'?' · local Ollama':''}<span>·</span>{$osaStore.isStreaming?'Working…':selected?'Saved conversation':'Your workspace, in conversation'}</p>{/if}
      </div></div>
      {#if selected&&!renaming}<button class="icon-button" disabled={$osaStore.isStreaming} onclick={()=>{title=selected.title;renaming=true;}} aria-label="Rename conversation" title="Rename conversation"><Pencil size={16}/></button>{/if}
    </header>
    {#if error}<p class="error" role="alert">{error}</p>{/if}
    <div class="thread" class:empty bind:this={thread} aria-busy={opening} onscroll={()=>followEnd=thread.scrollHeight-thread.scrollTop-thread.clientHeight<100}>
      {#if opening}<p class="hint">Opening conversation…</p>
      {:else if empty}<div class="welcome"><div class="welcome-orb"><video src="/OSAFinalNOBG.mp4" autoplay loop muted playsinline aria-hidden="true"></video></div><span class="eyebrow">OSA · YOUR BUSINESS ASSISTANT</span><h2>What are we working on?</h2><p>Think it through, make a plan, or pick up where you left off.</p><div class="suggestions">{#each ['Plan my day','Review open projects','Help me think through an idea'] as prompt}<button onclick={()=>{draft=prompt;input?.focus();}}>{prompt}<ArrowUp size={14}/></button>{/each}</div></div>
      {:else}<div class="messages">
        {#each $osaStore.conversation as message (message.id)}
          <article class:user={message.role==='user'}>
            {#if message.role!=='user'}<div class="message-label"><span class="assistant-mark">O</span><strong>{message.model && AGENT_RUNTIME_OPTIONS.some(r=>r.label===message.model)?message.model:'OSA'}</strong><time>{new Date(message.timestamp).toLocaleTimeString([],{hour:'numeric',minute:'2-digit'})}</time></div>{/if}
            <div class="message-content">{@html renderMarkdown(message.content)}</div>
            {#if message.role!=='user'}<div class="message-actions"><button class="icon-button" onclick={()=>copy(message.id,message.content)} aria-label={copied===message.id?'Copied message':'Copy message'} title="Copy message">{#if copied===message.id}<Check size={14}/>{:else}<Copy size={14}/>{/if}</button></div>{/if}
          </article>
        {/each}
        {#if $osaStore.isStreaming}<article aria-live="polite"><div class="message-label"><span class="assistant-mark">O</span><strong>{runtimeLabel}</strong></div>{#if $osaStore.streamingContent}<div class="message-content">{@html renderMarkdown($osaStore.streamingContent)}</div>{:else}<div class="thinking"><span></span><span></span><span></span><small role="status">{$osaStore.activity || "Working on it"}</small></div>{/if}</article>{/if}
      </div>{/if}
    </div>
    <div class="compose-area">
      {#if voiceActive||voiceError}<div class="voice-panel" role="status"><div class="voice-heading"><AudioLines size={18}/><strong>{voiceLabel}</strong><button class="icon-button" aria-label="Close voice conversation" onclick={()=>{voice?.cancel();voiceError='';}}><X size={16}/></button></div><p>{voiceError||transcript||voiceCaption||(voicePhase==='listening'?'Speak naturally. Tap the microphone when you’re done.':'Connecting your voice conversation…')}</p></div>{/if}
      <form class="composer" onsubmit={e=>{e.preventDefault();void send();}}>
        {#if $osaStore.error}<p class="error" role="alert">{$osaStore.error}</p>{/if}
        <textarea bind:this={input} aria-label="Message in conversation" placeholder="Message {runtimeLabel}…" bind:value={draft} rows="2" onkeydown={e=>{if(e.key==='Enter'&&!e.shiftKey&&!e.isComposing){e.preventDefault();void send();}}}></textarea>
        <div class="composer-controls"><div class="runtime-controls"><AgentRuntimeSelector/>{#if ['osa','ollama','hermes'].includes($osaStore.activeRuntime)}<span class="control-divider"></span><ModelSelector/>{/if}</div><div class="send-controls"><button type="button" class="voice-button" class:recording={voiceActive} onclick={()=>void voice?.toggle()} aria-label={voicePhase==='listening'?'Finish recording':voiceActive?'Stop voice conversation':'Talk to OSA'} title="Talk to OSA">{#if voiceActive}<AudioLines size={20}/>{:else}<Mic size={19}/>{/if}</button>{#if $osaStore.isStreaming}<button type="button" class="send-button" aria-label="Stop response" onclick={()=>osaStore.cancelStream()}><Square size={15} fill="currentColor"/></button>{:else}<button class="send-button" disabled={!draft.trim()||opening} aria-label="Send message" title="Send message"><ArrowUp size={20}/></button>{/if}</div></div>
      </form>
      <div class="composer-note"><span>Conversations are saved automatically.</span><span>↵ Send <span class="note-separator">·</span> Shift ↵ New line</span></div>
    </div>
  </main>
</section>

<style>
.conversations {
  --ink: #24262a;
  --muted: #81848b;
  --line: #e9e9ec;
  --surface: #fff;
  --rail: #f8f8f9;
  height: 100%;
  min-height: 480px;
  display: grid;
  grid-template-columns: 244px minmax(0, 1fr);
  background: var(--surface);
  color: var(--ink);
  font-family:
    Inter,
    -apple-system,
    BlinkMacSystemFont,
    "Segoe UI",
    sans-serif;
  container-type: inline-size;
}
.conversations.history-hidden {
  grid-template-columns: minmax(0, 1fr);
}
.history-hidden aside {
  display: none;
}
button,
input,
textarea {
  font: inherit;
}
button {
  cursor: pointer;
  color: inherit;
}
button:disabled {
  opacity: 0.4;
  cursor: default;
}
button:focus-visible {
  outline: 2px solid #8a99ae;
  outline-offset: 3px;
}
.icon-button {
  border: 0;
  background: transparent;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  flex-shrink: 0;
  border-radius: 8px;
  color: var(--muted);
}
.icon-button:hover {
  background: #eeeef0;
  color: var(--ink);
}
aside {
  display: flex;
  flex-direction: column;
  min-height: 0;
  padding: 21px 12px 10px;
  background: var(--rail);
  border-right: 1px solid var(--line);
}
.history-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 5px;
  margin-bottom: 22px;
}
h1 {
  font-size: 15px;
  letter-spacing: -0.35px;
  font-weight: 600;
  margin: 0;
}
.new-conversation {
  display: flex;
  align-items: center;
  gap: 9px;
  width: 100%;
  background: var(--surface);
  border: 1px solid #dedee3;
  border-radius: 10px;
  padding: 10px 11px;
  font-size: 12px;
  font-weight: 500;
  box-shadow: 0 1px 2px #00000003;
}
.new-hint {
  margin-left: auto;
  color: #b0b1b7;
  font-size: 16px;
}
.new-conversation:hover {
  border-color: #bfc2c9;
}
.search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 7px;
  color: #a0a2a9;
  margin: 10px 0 1px;
}
.search input {
  border: 0;
  background: transparent;
  color: var(--ink);
  font-size: 12px;
  width: 100%;
  min-width: 0;
  outline: none;
}
.search:focus-within {
  border-radius: 8px;
  box-shadow: 0 0 0 1px #c6cbd3;
}
.session-list {
  overflow: auto;
  flex: 1;
  min-height: 0;
}
.date-group {
  font-size: 10px;
  font-weight: 500;
  color: #8b8e96;
  margin: 19px 10px 8px;
}
.session {
  border: 0;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  width: 100%;
  padding: 11px 10px;
  margin-bottom: 3px;
  text-align: left;
  border-radius: 9px;
  background: transparent;
}
.session > :global(svg) {
  margin-top: 2px;
  color: #93969e;
  flex-shrink: 0;
}
.session > span {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.session strong {
  font-size: 12px;
  font-weight: 450;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.35;
}
.session small {
  color: #92959b;
  font-size: 10px;
}
.session:hover {
  background: #eeeef1;
}
.session.active {
  background: #e9eaee;
}
.session.active strong {
  font-weight: 550;
  color: #17191e;
}
.session.active > :global(svg) {
  color: #474c57;
}
.history-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-top: 1px solid var(--line);
  padding: 9px 3px 0;
  margin-top: 10px;
  color: #90939b;
  font-size: 10px;
}
.history-footer > span {
  display: flex;
  align-items: center;
  gap: 6px;
}
.saved-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: #838f86;
}
.hint {
  font-size: 12px;
  color: var(--muted);
  padding: 14px;
}
.load-more {
  border: 0;
  background: transparent;
  font-size: 11px;
  padding: 12px;
  color: var(--muted);
}
main {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  position: relative;
}
.thread-header {
  min-height: 76px;
  padding: 14px 28px;
  border-bottom: 1px solid var(--line);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  box-sizing: border-box;
}
.thread-title {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.osa-avatar {
  width: 32px;
  height: 32px;
  overflow: hidden;
  flex-shrink: 0;
}
.osa-avatar,
.welcome-orb {
  border-radius: 50%;
  overflow: hidden;
}
.osa-avatar video,
.welcome-orb video {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transform: scale(1.63);
}
.title-text {
  min-width: 0;
}
.title-text h2 {
  font-size: 13px;
  font-weight: 550;
  letter-spacing: -0.2px;
  margin: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}
.title-text p {
  font-size: 10px;
  color: var(--muted);
  margin: 5px 0 0;
  display: flex;
  gap: 7px;
}
.rename-form {
  display: flex;
  align-items: center;
  gap: 5px;
}
.rename-form input {
  font-size: 13px;
  padding: 7px;
  border: 1px solid var(--line);
  border-radius: 6px;
  min-width: 160px;
  color: var(--ink);
  background: var(--surface);
}
.thread {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 32px 32px 12px;
  scrollbar-width: thin;
  scrollbar-color: #e2e3e7 transparent;
}
.messages {
  max-width: 760px;
  margin: 0 auto;
}
.messages article {
  margin: 0 0 27px;
}
.messages article.user {
  margin: 0 0 30px auto;
  max-width: 82%;
  width: fit-content;
}
.user .message-content {
  padding: 13px 18px;
  border-radius: 19px 19px 5px 19px;
  background: #f1f2f5;
}
.message-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  margin-bottom: 13px;
}
.message-label strong {
  font-weight: 600;
}
.assistant-mark {
  width: 21px;
  height: 21px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #292d34;
  border-radius: 50%;
  font-size: 10px;
  color: white;
  font-weight: 600;
}
.message-label time {
  font-size: 10px;
  color: #a1a4ab;
  margin-left: 2px;
}
.message-content {
  font-size: 14px;
  line-height: 1.8;
  overflow-wrap: anywhere;
  color: #353940;
}
.message-content :global(p) {
  margin: 0 0 12px;
}
.message-content :global(p:last-child) {
  margin-bottom: 0;
}
.message-content :global(pre) {
  overflow: auto;
  background: #f5f6f8;
  border: 1px solid var(--line);
  padding: 15px;
  border-radius: 10px;
  font-size: 12px;
  line-height: 1.6;
}
.message-content :global(a) {
  color: #3861a4;
}
.message-content :global(ul),
.message-content :global(ol) {
  padding-left: 24px;
}
.message-actions {
  margin: 8px 0 0 -5px;
}
.message-actions .icon-button {
  width: 27px;
  height: 27px;
  color: #a3a5ac;
}
.thinking {
  display: flex;
  align-items: center;
  gap: 4px;
  height: 30px;
}
.thinking > span {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: #949aa5;
  animation: pulse 1.3s infinite;
}
.thinking > span:nth-child(2) {
  animation-delay: 0.15s;
}
.thinking > span:nth-child(3) {
  animation-delay: 0.3s;
}
.thinking small {
  color: #989ca5;
  font-size: 11px;
  margin-left: 7px;
}
.thread.empty {
  display: flex;
  align-items: center;
  justify-content: center;
}
.welcome {
  max-width: 500px;
  text-align: center;
  margin-top: -20px;
}
.welcome-orb {
  width: 66px;
  height: 66px;
  margin: 0 auto 21px;
}
.eyebrow {
  font-size: 9px;
  letter-spacing: 1.6px;
  color: #969aa2;
  font-weight: 500;
}
.welcome h2 {
  font-size: 28px;
  font-weight: 500;
  letter-spacing: -1px;
  margin: 13px 0 11px;
  line-height: 1.25;
}
.welcome p {
  font-size: 12px;
  color: #8a8e97;
  line-height: 1.7;
  margin: 0;
}
.suggestions {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 8px;
  margin-top: 25px;
}
.suggestions button {
  border: 1px solid var(--line);
  border-radius: 9px;
  background: var(--surface);
  padding: 10px 12px;
  display: flex;
  gap: 12px;
  align-items: center;
  font-size: 11px;
  color: #707580;
}
.suggestions button:hover {
  border-color: #b9bec8;
  color: var(--ink);
}
.suggestions :global(svg) {
  transform: rotate(45deg);
  color: #a1a6ae;
}
.compose-area {
  width: 100%;
  max-width: 816px;
  padding: 0 28px 16px;
  box-sizing: border-box;
  margin: 0 auto;
  flex-shrink: 0;
}
.composer {
  border: 1px solid #dfe1e6;
  border-radius: 21px;
  padding: 14px 13px 10px;
  background: var(--surface);
  box-shadow: 0 3px 14px #20273605;
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
}
.composer:focus-within {
  border-color: #b6bfce;
  box-shadow:
    0 0 0 3px #8899b00b,
    0 3px 14px #20273605;
}
.composer textarea {
  display: block;
  width: 100%;
  box-sizing: border-box;
  resize: none;
  min-height: 48px;
  max-height: 180px;
  field-sizing: content;
  border: 0 !important;
  background: transparent;
  color: var(--ink);
  outline: none !important;
  box-shadow: none !important;
  font-size: 14px;
  line-height: 1.65;
  padding: 1px 5px 10px;
}
.composer textarea::placeholder {
  color: #a1a5af;
}
.composer-controls {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.runtime-controls,
.send-controls {
  display: flex;
  gap: 7px;
  align-items: center;
  min-width: 0;
}
.runtime-controls {
  gap: 4px;
}
.control-divider {
  height: 13px;
  width: 1px;
  background: var(--line);
  margin: 0 2px;
}
.runtime-controls :global(.ars-trigger),
.runtime-controls :global(.model-trigger) {
  height: 32px;
  padding: 6px 8px !important;
  background: transparent !important;
  border: 0 !important;
  box-shadow: none !important;
  border-radius: 8px !important;
  color: #6e7480 !important;
  font-size: 11px !important;
  max-width: 260px;
}
.runtime-controls :global(.ars-trigger:hover),
.runtime-controls :global(.model-trigger:hover) {
  background: #f3f4f6 !important;
}
.runtime-controls :global(.model-label) {
  max-width: 200px;
}
.voice-button,
.send-button {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 50%;
  flex-shrink: 0;
}
.voice-button {
  color: #707680;
  background: transparent;
}
.voice-button:hover {
  background: #f0f1f4;
}
.voice-button.recording {
  color: #3861a4;
  background: #eaf0fa;
}
.send-button {
  color: #fff;
  background: #2d3139;
}
.send-button:disabled {
  background: #e9ebef;
  color: #a6acb7;
  opacity: 1;
}
.send-button:not(:disabled):hover {
  background: #14171b;
}
.composer-note {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 9px;
  color: #a4a8b0;
  padding: 10px 5px 0;
}
.note-separator {
  padding: 0 4px;
}
.voice-panel {
  margin: 0 0 12px;
  border: 1px solid #dde3ec;
  border-radius: 15px;
  padding: 11px 15px;
  background: #f6f8fc;
  color: #4c5a6e;
}
.voice-heading {
  display: flex;
  align-items: center;
  gap: 9px;
  font-size: 12px;
}
.voice-heading .icon-button {
  margin-left: auto;
}
.voice-panel p {
  font-size: 12px;
  line-height: 1.6;
  max-height: 120px;
  overflow: auto;
  margin: 8px 0 2px;
}
.error {
  color: #9c3838;
  background: #fff3f2;
  padding: 11px 14px;
  font-size: 12px;
  border-radius: 9px;
  margin: 8px;
}
:global(.spin) {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
@keyframes pulse {
  50% {
    opacity: 0.3;
    transform: translateY(-2px);
  }
}
@media (prefers-reduced-motion: reduce) {
  .thinking > span,
  :global(.spin) {
    animation: none;
  }
}
@container (max-width:820px) {
  .conversations {
    grid-template-columns: 200px minmax(0, 1fr);
  }
  .runtime-controls :global(.model-label) {
    max-width: 120px;
  }
  .thread-header {
    padding: 14px 18px;
  }
  .compose-area {
    padding: 0 18px 14px;
  }
  .composer-note > span:last-child {
    display: none;
  }
  .thread {
    padding: 25px 20px 10px;
  }
}
@media (max-width: 650px) {
  .conversations {
    grid-template-columns: minmax(0, 1fr);
  }
  aside {
    position: absolute;
    inset: 0 auto 0 0;
    width: 220px;
    z-index: 20;
    box-shadow: 10px 0 40px #00000015;
  }
  .thread-header {
    padding: 12px 16px;
  }
  .runtime-controls :global(.model-label) {
    max-width: 90px;
  }
  .welcome h2 {
    font-size: 24px;
  }
  .welcome {
    padding: 20px;
  }
  .thread {
    padding: 24px 17px 10px;
  }
  .compose-area {
    padding: 0 12px 12px;
  }
  .composer-note {
    font-size: 8px;
  }
}
:global(.dark) .conversations {
  --ink: #e7e9ee;
  --muted: #989da7;
  --line: #34373f;
  --surface: #22252b;
  --rail: #1d2025;
}
:global(.dark) .session.active,
:global(.dark) .session:hover,
:global(.dark) .user .message-content {
  background: #30343d;
}
:global(.dark) .session.active strong,
:global(.dark) .message-content {
  color: #e3e6ed;
}
:global(.dark) .composer {
  border-color: #414752;
}
:global(.dark) .composer:focus-within {
  border-color: #737e91;
}
:global(.dark) .voice-panel {
  background: #2a3341;
  border-color: #414d60;
  color: #becce2;
}
:global(.dark) .send-button {
  background: #e4e8f0;
  color: #232831;
}
:global(.dark) .search input {
  color: #e3e6ed;
}
:global(.dark) .new-conversation {
  border-color: #414752;
}
</style>
