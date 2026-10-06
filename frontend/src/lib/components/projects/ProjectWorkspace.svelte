<script lang="ts">
  import { onMount } from 'svelte';
  import { api } from '$lib/api';
  import { request } from '$lib/api/base';
  import type { Project, ProjectNote } from '$lib/api/projects/types';
  import type { Task, TaskStatus } from '$lib/api/dashboard/types';
  import type { TeamMemberListResponse } from '$lib/api/team';
  import { marked } from 'marked';
  import DOMPurify from 'dompurify';
  import { currentWorkspace } from '$lib/stores/workspaces';
  import { fetchFile } from '$lib/kb/client';
  import { resolveDocumentLink } from '$lib/kb/document-links';
  import { ArrowLeft, ArrowUpRight, Search, Plus, Pencil, Check, X, RefreshCw, FileText, Users, ListChecks } from 'lucide-svelte';

  let { project, onback, onupdate }: { project: Project; onback: () => void; onupdate: (project: Project) => void } = $props();
  let tab = $state<'overview' | 'tasks' | 'people' | 'notes'>('overview');
  let tasks = $state<Task[]>([]);
  let members = $state<TeamMemberListResponse[]>([]);
  let notes = $state<ProjectNote[]>([]);
  let loading = $state(true);
  let error = $state('');
  let saving = $state(false);
  let search = $state('');
  let ownerFilter = $state('');
  let statusFilter = $state('');
  let selectedId = $state<string | null>(null);
  let editing = $state(false);
  let editingProject = $state(false);
  let creating = $state(false);
  let draft = $state({ title: '', description: '', status: 'todo' as TaskStatus, priority: 'medium' as Task['priority'], assignee_id: '', due_date: '' });
  let projectDraft = $state({ name: '', description: '', status: 'active' as Project['status'], priority: 'medium' as Project['priority'] });
  let note = $state('');
  let sourcePath = $state('');
  let sourceBody = $state('');
  let sourceLoading = $state(false);
  let sourceRequest = 0;
  const selected = $derived(tasks.find(t => t.id === selectedId));
  const statuses: { value: TaskStatus; label: string }[] = [{ value: 'todo', label: 'To do' }, { value: 'in_progress', label: 'In progress' }, { value: 'done', label: 'Done' }, { value: 'cancelled', label: 'Cancelled' }];
  const visible = $derived(tasks.filter(t => (!ownerFilter || (ownerFilter === 'unassigned' ? !t.assignee_id : t.assignee_id === ownerFilter)) && (!statusFilter || t.status === statusFilter) && `${t.title} ${t.description ?? ''}`.toLowerCase().includes(search.toLowerCase())));
  const complete = $derived(tasks.filter(t => t.status === 'done').length);
  const people = $derived(members.filter(m => tasks.some(t => t.assignee_id === m.id)));
  const html = (value: string) => DOMPurify.sanitize(marked.parse(value, { async: false, breaks: true }), { FORBID_TAGS: ['style'], FORBID_ATTR: ['style'] });
  const owner = (id: string | null) => members.find(m => m.id === id)?.name ?? 'Unassigned';
  const date = (value: string | null) => value ? new Date(value.slice(0, 10) + 'T12:00:00').toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) : 'Not scheduled';
  const initials = (name: string) => name.split(/\s+/).filter(Boolean).map(word => word[0]).slice(0, 2).join('');
  function taskLabel(title: string) {
    const match = title.match(/^(SETUP|MON|TUE|WED|THU|FRI|FRI-MON|DAILY|WEEKLY|AS NEEDED) \| [^|]+ \| (.+)$/);
    return match ? { cadence: match[1], title: match[2] } : { cadence: '', title };
  }

  onMount(load);
  async function load() {
    loading = true; error = '';
    try {
      const [items, team, detail] = await Promise.all([api.getTasks({ projectId: project.id }), api.getTeamMembers(), request<{ notes: ProjectNote[] }>(`/projects/${project.id}?include_notes=true`, { skipCache: true })]);
      tasks = items; members = team; notes = detail.notes ?? [];
    } catch (e) { error = e instanceof Error ? e.message : 'Unable to load project. Try again.'; }
    finally { loading = false; }
  }
  function openTask(task: Task) {
    sourceRequest++; selectedId = task.id; sourcePath = ''; sourceBody = ''; sourceLoading = false; editing = false; creating = false;
  }
  function editTask(task?: Task) {
    sourceRequest++; sourcePath = ''; sourceLoading = false;
    draft = { title: task?.title ?? '', description: task?.description ?? '', status: task?.status ?? 'todo', priority: task?.priority ?? 'medium', assignee_id: task?.assignee_id ?? '', due_date: task?.due_date?.slice(0,10) ?? '' };
    creating = !task; editing = true; if (!task) selectedId = null;
  }
  async function saveTask(event: SubmitEvent) {
    event.preventDefault(); if (!draft.title.trim() || saving) return;
    saving = true; error = '';
    try {
      const payload = { ...draft, title: draft.title.trim(), due_date: draft.due_date ? `${draft.due_date}T12:00:00Z` : '', project_id: project.id };
      const saved = creating ? await api.createTask(payload) : await api.updateTask(selectedId!, payload);
      tasks = creating ? [...tasks, saved] : tasks.map(t => t.id === saved.id ? saved : t);
      openTask(saved);
    } catch (e) { error = e instanceof Error ? e.message : 'Unable to save task.'; }
    finally { saving = false; }
  }
  async function updateStatus(task: Task, status: TaskStatus) {
    saving = true; error = '';
    try { const saved = await api.updateTask(task.id, { status }); tasks = tasks.map(t => t.id === task.id ? saved : t); }
    catch (e) { error = e instanceof Error ? e.message : 'Unable to update task.'; }
    finally { saving = false; }
  }
  function editProject() { projectDraft = { name: project.name, description: project.description ?? '', status: project.status, priority: project.priority }; editingProject = true; }
  async function saveProject(event: SubmitEvent) {
    event.preventDefault(); saving = true; error = '';
    try { await api.updateProject(project.id, { ...projectDraft, name: projectDraft.name.trim() }); onupdate({ ...project, ...projectDraft, name: projectDraft.name.trim() }); editingProject = false; }
    catch (e) { error = e instanceof Error ? e.message : 'Unable to save project.'; }
    finally { saving = false; }
  }
  async function addNote(event: SubmitEvent) {
    event.preventDefault(); if (!note.trim()) return;
    saving = true; error = '';
    try { const saved = await api.addProjectNote(project.id, note.trim()); notes = [...notes, saved]; note = ''; }
    catch (e) { error = e instanceof Error ? e.message : 'Unable to save note.'; }
    finally { saving = false; }
  }
  async function openSource(path: string) {
    const seq = ++sourceRequest; sourcePath = path; sourceLoading = true; sourceBody = ''; error = '';
    try { const file = await fetchFile($currentWorkspace!.slug, path); if (seq === sourceRequest) sourceBody = file.content; }
    catch (e) { if (seq === sourceRequest) error = e instanceof Error ? e.message : 'Unable to open document.'; }
    finally { if (seq === sourceRequest) sourceLoading = false; }
  }
  function sourceLinks(node: HTMLElement) {
    const click = (event: MouseEvent) => {
      const anchor = (event.target as Element).closest('a'); if (!anchor || !node.contains(anchor)) return;
      const link = resolveDocumentLink(anchor.getAttribute('href') ?? '', sourcePath || 'task.md');
      if (link === null) return;
      event.preventDefault(); if (link === 'blocked') { error = 'Document path is outside this workspace.'; return; }
      void openSource(link.path);
    };
    node.addEventListener('click', click); return { destroy() { node.removeEventListener('click', click); } };
  }
</script>

<section class="project-workspace" aria-label="Project workspace">
  <header class="project-header">
    <button class="icon" onclick={onback} aria-label="Back to projects" title="Back to projects"><ArrowLeft size={19}/></button>
    <div class="heading"><p>Projects / {project.client_name || 'Internal'}</p><h1>{project.name}</h1></div>
    <span class="badge" data-status={project.status}><i></i>{project.status}</span>
    <button class="icon" onclick={load} disabled={loading || saving} aria-label="Refresh project" title="Refresh project"><RefreshCw size={17}/></button>
  </header>
  <nav class="project-tabs" aria-label="Project views">
    <button class:active={tab === 'overview'} onclick={() => { tab = 'overview'; sourcePath = ''; }}><FileText size={16}/>Overview</button>
    <button class:active={tab === 'tasks'} onclick={() => tab = 'tasks'}><ListChecks size={16}/>Tasks <span>{tasks.length}</span></button>
    <button class:active={tab === 'people'} onclick={() => tab = 'people'}><Users size={16}/>People <span>{people.length}</span></button>
    <button class:active={tab === 'notes'} onclick={() => tab = 'notes'}>Notes <span>{notes.length}</span></button>
  </nav>
  {#if error}<div class="error" role="alert">{error}</div>{/if}
  {#if loading}<p class="empty" role="status">Loading project...</p>
  {:else if tab === 'overview'}
    <div class="overview scroll">
      <div class="overview-main">
      {#if editingProject}
        <form onsubmit={saveProject}>
          <label>Project name<input bind:value={projectDraft.name} required /></label>
          <div class="fields"><label>Status<select bind:value={projectDraft.status}>{#each ['active','paused','completed','archived'] as status}<option value={status}>{status}</option>{/each}</select></label><label>Priority<select bind:value={projectDraft.priority}>{#each ['low','medium','high','critical'] as priority}<option value={priority}>{priority}</option>{/each}</select></label></div>
          <label>Description<textarea rows="16" bind:value={projectDraft.description}></textarea></label>
          <div class="actions"><button type="button" onclick={() => editingProject = false} disabled={saving}>Cancel</button><button class="primary" disabled={saving || !projectDraft.name.trim()}><Check size={16}/>Save project</button></div>
        </form>
      {:else}
        <div class="section-head"><h2>{sourcePath ? 'Project document' : 'Project brief'}</h2><button class="quiet" onclick={editProject}><Pencil size={14}/>Edit project</button></div>
        {#if typeof project.project_metadata?.knowledge_path === 'string'}<button class="guide-link" onclick={() => openSource(project.project_metadata!.knowledge_path as string)}><FileText size={16}/>Open operating guide <ArrowUpRight size={14}/></button>{/if}
        {#if sourcePath}<button onclick={() => { sourceRequest++; sourcePath = ''; }}><ArrowLeft size={15}/>Back to overview</button>{/if}
        <div class="prose" use:sourceLinks>{#if sourceLoading && sourcePath}<p role="status">Loading document...</p>{:else}{@html html(sourcePath ? sourceBody : (project.description || 'No overview added.'))}{/if}</div>
      {/if}
      </div>
      <aside class="project-summary" aria-label="Project summary">
        <h2>Project details</h2>
        <dl><div><dt>Status</dt><dd><span class="badge" data-status={project.status}><i></i>{project.status}</span></dd></div><div><dt>Priority</dt><dd class="capitalize">{project.priority}</dd></div><div><dt>Due date</dt><dd>{date(project.due_date ?? null)}</dd></div><div><dt>Type</dt><dd class="capitalize">{(project.project_type || 'Internal').replaceAll('_', ' ')}</dd></div></dl>
        <div class="summary-progress"><div><h3>Progress</h3><span>{tasks.length ? Math.round(complete / tasks.length * 100) : 0}%</span></div><progress max={Math.max(tasks.length, 1)} value={complete} aria-label="Project completion"></progress><p>{complete} of {tasks.length} tasks completed</p></div>
        <h3>Assigned people <span>{people.length}</span></h3>
        {#each people as person}<button class="summary-person" onclick={() => { ownerFilter = person.id; tab = 'tasks'; }}><span class="avatar" aria-hidden="true">{initials(person.name)}</span><span>{person.name}</span><small>{tasks.filter(t => t.assignee_id === person.id && t.status !== 'done' && t.status !== 'cancelled').length}</small></button>{/each}
        {#if !people.length}<p class="summary-empty">No assigned people</p>{/if}
      </aside>
    </div>
  {:else if tab === 'tasks'}
    <div class="filters"><div class="search"><Search size={16}/><input aria-label="Search project tasks" placeholder="Search tasks" bind:value={search}/></div><select aria-label="Filter by assignee" bind:value={ownerFilter}><option value="">All people</option><option value="unassigned">Unassigned</option>{#each people as person}<option value={person.id}>{person.name}</option>{/each}</select><select aria-label="Filter by status" bind:value={statusFilter}><option value="">All statuses</option>{#each statuses as status}<option value={status.value}>{status.label}</option>{/each}</select><button class="primary" onclick={() => editTask()}><Plus size={16}/>New task</button></div>
    <div class="task-layout" class:has-detail={selected || creating}>
      <div class="task-list scroll" class:mobile-hidden={selected || creating}>
        <div class="task-list-heading"><span>Task <small>{visible.length}</small></span><span>Owner</span><span>Status</span><span>Due date</span></div>
        {#each visible as task (task.id)}{@const label = taskLabel(task.title)}<button class="task-row" aria-label={task.title} class:selected={selectedId === task.id} onclick={() => openTask(task)}><strong>{#if label.cadence}<span class="cadence">{label.cadence}</span>{/if}{label.title}</strong><span class="row-owner"><span class="avatar" aria-hidden="true">{initials(owner(task.assignee_id))}</span>{owner(task.assignee_id)}</span><span class="row-status" data-status={task.status}><i></i>{statuses.find(s => s.value === task.status)?.label}</span><small class="row-due">{date(task.due_date)}</small></button>{:else}<p class="empty">{tasks.length ? 'No tasks match these filters.' : 'No tasks yet.'}</p>{/each}
      </div>
      {#if selected || creating}
        <section class="task-detail scroll" aria-label="Project task details">
          <div class="section-head"><span>Task details</span><button class="icon" onclick={() => { sourceRequest++; selectedId = null; creating = false; editing = false; sourcePath = ''; }} aria-label="Close task details" title="Close task details"><X size={18}/></button></div>
          {#if editing}
            <form onsubmit={saveTask}><label>Task title<input required bind:value={draft.title}/></label><div class="fields"><label>Assignee<select bind:value={draft.assignee_id}><option value="">Unassigned</option>{#each members as m}<option value={m.id}>{m.name}</option>{/each}</select></label><label>Due date<input type="date" bind:value={draft.due_date}/></label></div><div class="fields"><label>Status<select bind:value={draft.status}>{#each statuses as s}<option value={s.value}>{s.label}</option>{/each}</select></label><label>Priority<select bind:value={draft.priority}>{#each ['low','medium','high','critical'] as p}<option value={p}>{p}</option>{/each}</select></label></div><label>Task brief<textarea rows="18" bind:value={draft.description}></textarea></label><div class="actions"><button type="button" disabled={saving} onclick={() => { editing = false; creating = false; }}>Cancel</button><button class="primary" disabled={saving || !draft.title.trim()}><Check size={16}/>Save task</button></div></form>
          {:else if selected}
            {#if taskLabel(selected.title).cadence}<span class="cadence detail-cadence">{taskLabel(selected.title).cadence}</span>{/if}
            <h2 class="task-title">{taskLabel(selected.title).title}</h2><div class="task-properties"><div><span class="property-label">Assignee</span><span class="property-value"><span class="avatar" aria-hidden="true">{initials(owner(selected.assignee_id))}</span>{owner(selected.assignee_id)}</span></div><div><span class="property-label">Due date</span><span>{date(selected.due_date)}</span></div><div><span class="property-label">Status</span><select aria-label="Task status" value={selected.status} disabled={saving} onchange={e => updateStatus(selected!, e.currentTarget.value as TaskStatus)}>{#each statuses as s}<option value={s.value}>{s.label}</option>{/each}</select></div><button class="quiet" onclick={() => editTask(selected)}><Pencil size={14}/>Edit task</button></div>
            {#if sourcePath}<button onclick={() => { sourceRequest++; sourcePath = ''; }}><ArrowLeft size={15}/>Back to task brief</button>{/if}
            <div class="prose" use:sourceLinks>{#if sourceLoading && sourcePath}<p role="status">Loading document...</p>{:else}{@html html(sourcePath ? sourceBody : (selected.description || 'No task brief added.'))}{/if}</div>
          {/if}
        </section>
      {/if}
    </div>
  {:else if tab === 'people'}
    <div class="people scroll"><h2>Task ownership</h2>{#each people as person}<button class="person" onclick={() => { ownerFilter = person.id; tab = 'tasks'; selectedId = null; }}><span class="avatar" aria-hidden="true">{initials(person.name)}</span><span class="person-name"><strong>{person.name}</strong><small>{person.role}</small></span><span class="workload">{tasks.filter(t => t.assignee_id === person.id && t.status !== 'done' && t.status !== 'cancelled').length} open tasks</span></button>{/each}<button class="person" onclick={() => { ownerFilter = 'unassigned'; tab = 'tasks'; selectedId = null; }}><span class="avatar" aria-hidden="true">?</span><strong class="person-name">Unassigned</strong><span class="workload">{tasks.filter(t => !t.assignee_id).length} tasks</span></button></div>
  {:else}
    <div class="notes scroll"><h2>Project notes</h2><form onsubmit={addNote}><label>New note<textarea rows="4" bind:value={note} placeholder="Decisions, updates and handoffs"></textarea></label><button class="primary" disabled={saving || !note.trim()}><Plus size={16}/>Add note</button></form>{#each notes as item}<article><small>{new Date(item.created_at).toLocaleString()}</small><div class="prose">{@html html(item.content)}</div></article>{:else}<p class="empty">No notes yet.</p>{/each}</div>
  {/if}
</section>

<style>
  .project-workspace{height:100%;min-height:0;display:flex;flex-direction:column;background:var(--dbg);color:var(--dt);font-size:14px;letter-spacing:0;overflow:hidden}
  .project-header{display:flex;gap:14px;align-items:center;padding:18px 24px;border-bottom:1px solid var(--dbd)}.heading{flex:1;min-width:0}.heading p{font-size:12px;color:var(--dt3);margin:0 0 4px}h1{font-size:22px;line-height:1.25;margin:0;overflow-wrap:anywhere}h2{font-size:18px;line-height:1.4;margin:0 0 16px;overflow-wrap:anywhere}
  button,input,textarea,select{font:inherit;color:inherit}button{display:inline-flex;align-items:center;justify-content:center;gap:7px;padding:8px 12px;border:1px solid var(--dbd);border-radius:6px;background:var(--dbg);cursor:pointer}button:hover{background:color-mix(in srgb,var(--dt) 6%,var(--dbg))}button:disabled{opacity:.55;cursor:not-allowed}.icon{width:36px;height:36px;padding:0;flex-shrink:0}.primary{background:var(--dt);color:var(--dbg)}.primary:hover{background:color-mix(in srgb,var(--dt) 90%,var(--dbg))}.badge{text-transform:capitalize;font-size:12px;border:1px solid var(--dbd);border-radius:4px;padding:4px 8px}
  nav{display:flex;gap:8px;padding:0 24px;border-bottom:1px solid var(--dbd);flex-shrink:0;overflow:auto}nav button{border:0;border-radius:0;padding:13px 10px;color:var(--dt3);border-bottom:2px solid transparent}nav button.active{border-bottom-color:var(--dt);color:var(--dt)}nav span{font-size:11px}.scroll{overflow:auto;min-height:0;overscroll-behavior:contain}.overview,.people,.notes{padding:24px;flex:1}.metrics{display:flex;gap:44px;flex-wrap:wrap}.metrics div{display:flex;flex-direction:column;gap:6px}.metrics strong{font-size:20px}.metrics span{font-size:12px;color:var(--dt3)}progress{width:100%;height:5px;accent-color:#279c77;margin:20px 0 28px}.section-head{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:16px}.section-head h2{margin:0}
  .prose{font-size:14px;line-height:1.75;overflow-wrap:anywhere;max-width:90ch;margin-top:18px}.prose :global(h2){font-size:18px;margin:24px 0 8px}.prose :global(h3){font-size:15px;margin:22px 0 8px}.prose :global(p){margin:0 0 12px}.prose :global(ul),.prose :global(ol){padding-left:22px}.prose :global(ul){list-style:disc}.prose :global(ol){list-style:decimal}.prose :global(a){color:var(--accent,#15836a);text-decoration:underline}.prose :global(table){display:block;overflow:auto;white-space:normal;border-collapse:collapse}.prose :global(td),.prose :global(th){padding:8px;border:1px solid var(--dbd)}.prose :global(pre){overflow:auto;white-space:pre}
  .filters{display:flex;gap:10px;padding:14px 24px;flex-wrap:wrap;border-bottom:1px solid var(--dbd)}.search{display:flex;align-items:center;gap:8px;border:1px solid var(--dbd);border-radius:6px;padding:0 10px;flex:1;min-width:160px}.search input{border:0;background:transparent;width:100%;min-width:0}input,select,textarea{background:var(--dbg);border:1px solid var(--dbd);border-radius:5px;padding:8px;min-width:0;max-width:100%}textarea{resize:vertical;line-height:1.5}.task-layout{display:flex;min-height:0;flex:1}.task-list{flex:1;padding:8px 16px}.task-row{display:grid;grid-template-columns:1fr auto;width:100%;text-align:left;justify-content:stretch;gap:6px 12px;border:0;border-bottom:1px solid var(--dbd);border-radius:0;padding:14px 10px}.task-row strong{grid-column:1/-1;overflow-wrap:anywhere;font-weight:550}.task-row span,.task-row small{font-size:12px;color:var(--dt3)}.task-row.selected{background:color-mix(in srgb,var(--dt) 7%,var(--dbg));box-shadow:inset 3px 0 var(--dt)}.separator{margin:0 6px}.has-detail .task-list{flex:0 0 34%;min-width:220px}.task-detail{flex:1;min-width:0;border-left:1px solid var(--dbd);padding:20px 24px}.task-properties{display:flex;align-items:center;gap:10px;flex-wrap:wrap;color:var(--dt2);font-size:12px}.fields{display:grid;grid-template-columns:1fr 1fr;gap:14px}label{display:flex;flex-direction:column;gap:6px;margin-bottom:14px;font-size:12px;font-weight:550}.actions{display:flex;gap:8px;justify-content:flex-end}.error{margin:12px 24px;padding:12px;border:1px solid #dc6666;border-radius:6px;color:#b93333}.empty{padding:24px;color:var(--dt3)}.person{display:flex;width:100%;justify-content:space-between;text-align:left;border:0;border-bottom:1px solid var(--dbd);border-radius:0;padding:18px 0}.person small{display:block;color:var(--dt3);margin-top:5px}.notes form{max-width:760px}.notes article{padding:20px 0;border-bottom:1px solid var(--dbd)}.notes small{color:var(--dt3)}:is(button,input,select,textarea):focus-visible{outline:2px solid var(--accent,#279c77);outline-offset:2px}
  @media(max-width:700px){.project-header{padding:14px;gap:8px}h1{font-size:18px}.badge{display:none}nav{padding:0 10px;gap:0}nav button{padding:12px 8px}.filters{padding:12px}.filters .search{flex-basis:100%;min-height:38px}.overview,.people,.notes,.task-detail{padding:18px 14px}.metrics{gap:22px}.metrics strong{font-size:17px}.has-detail .mobile-hidden{display:none}.task-detail{border:0}.fields{grid-template-columns:1fr}.task-row{grid-template-columns:1fr}.person{gap:12px}.filters select{flex:1 1 110px;min-width:110px}button{min-height:36px}}
  .project-workspace { container: project / inline-size; --surface-subtle: color-mix(in srgb, var(--dt) 2.5%, var(--dbg)); --line-soft: color-mix(in srgb, var(--dbd) 65%, transparent); }
  .project-header { padding: 14px 22px 10px; gap: 12px; border-bottom: 0; }
  .project-header > .icon { border-color: transparent; background: transparent; color: var(--dt2); }
  .heading p { font-size: 11px; margin-bottom: 4px; }
  h1 { font-size: 20px; font-weight: 650; }
  .project-tabs { padding: 0 22px; gap: 24px; background: var(--dbg); }
  .project-tabs button { padding: 12px 0 11px; min-height: 42px; font-size: 13px; font-weight: 500; background: transparent; box-shadow: none; }
  .project-tabs button:hover { color: var(--dt); background: transparent; }
  .project-tabs button.active { border-bottom-color: var(--dt); background: transparent; font-weight: 600; }
  nav span { border-radius: 4px; background: var(--surface-subtle); padding: 1px 5px; min-width: 18px; }
  .badge { display: inline-flex; align-items: center; gap: 6px; background: var(--surface-subtle); border-color: var(--line-soft); padding: 5px 8px; font-size: 11px; }
  .badge i, .row-status i { width: 6px; height: 6px; border-radius: 50%; background: #9b9ba2; flex-shrink: 0; }
  .badge[data-status='active'] i, .row-status[data-status='done'] i { background: #27846c; }
  .badge[data-status='paused'] i, .row-status[data-status='in_progress'] i { background: #b38b30; }
  .quiet { border-color: transparent; color: var(--dt2); font-size: 12px; }
  .overview { display: grid; grid-template-columns: minmax(0,1fr) 250px; padding: 0; }
  .overview-main { min-width: 0; padding: 30px 36px; }
  .overview-main > .section-head h2 { font-size: 15px; font-weight: 600; }
  .guide-link { font-size: 12px; padding: 9px 12px; border-color: var(--line-soft); }
  .guide-link :global(svg:last-child) { margin-left: 12px; color: var(--dt3); }
  .project-summary { border-left: 1px solid var(--line-soft); background: var(--surface-subtle); padding: 30px 22px; }
  .project-summary h2, .project-summary h3 { font-size: 12px; font-weight: 600; margin: 0 0 18px; }
  .project-summary h3 span { margin-left: 6px; color: var(--dt3); font-weight: 400; }
  .project-summary dl { display: grid; gap: 17px; font-size: 12px; margin: 0; }
  .project-summary dl > div { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
  .project-summary dt { color: var(--dt3); }
  .project-summary dd { margin: 0; text-align: right; }
  .capitalize { text-transform: capitalize; }
  .summary-progress { border-block: 1px solid var(--line-soft); padding: 22px 0; margin: 24px 0; }
  .summary-progress > div { display: flex; justify-content: space-between; font-size: 12px; }
  .summary-progress h3 { margin: 0; }
  .summary-progress progress { margin: 14px 0 8px; height: 4px; display: block; border: 0; border-radius: 4px; overflow: hidden; background: var(--dbd); }
  progress::-webkit-progress-bar { background: var(--dbd); }
  progress::-webkit-progress-value { background: #27846c; }
  .summary-progress p, .summary-empty { font-size: 11px; color: var(--dt3); margin: 0; }
  .summary-person { display: flex; width: 100%; border: 0; background: transparent; padding: 8px 0; justify-content: flex-start; font-size: 12px; text-align: left; }
  .summary-person small { margin-left: auto; color: var(--dt3); }
  .avatar { width: 25px; height: 25px; display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; border-radius: 50%; font-size: 9px; font-weight: 600; background: color-mix(in srgb, #7f8ba4 13%, var(--dbg)); color: var(--dt2); border: 1px solid var(--line-soft); }
  .filters { padding: 14px 28px; gap: 8px; background: var(--dbg); }
  .filters .search { max-width: 310px; border-color: var(--line-soft); }
  .filters input, .filters select, .filters button { font-size: 12px; }
  .filters > .primary { margin-left: auto; }
  .task-list { padding: 0 24px; }
  .task-list-heading, .task-row { display: grid; grid-template-columns: minmax(200px,1fr) 155px 108px 104px; gap: 16px; align-items: center; }
  .task-list-heading { padding: 12px 10px; font-size: 11px; color: var(--dt3); border-bottom: 1px solid var(--dbd); position: sticky; top: 0; z-index: 1; background: var(--dbg); }
  .task-list-heading small { padding-left: 8px; }
  .task-row { padding: 16px 10px; min-height: 67px; border-bottom-color: var(--line-soft); }
  .task-row strong { grid-column: auto; display: flex; align-items: center; gap: 10px; font-size: 13px; font-weight: 500; line-height: 1.5; }
  .task-row .cadence, .cadence { font-size: 9px; font-weight: 600; line-height: 1.6; padding: 2px 5px; border: 1px solid var(--line-soft); border-radius: 4px; color: var(--dt3); white-space: nowrap; flex-shrink: 0; }
  .row-owner, .row-status { display: inline-flex; align-items: center; gap: 7px; }
  .task-row .row-owner, .task-row .row-status, .task-row .row-due { font-size: 11px; }
  .row-owner .avatar { width: 22px; height: 22px; font-size: 8px; }
  .task-row.selected { background: color-mix(in srgb,#27846c 6%,var(--dbg)); box-shadow: inset 2px 0 #27846c; }
  .has-detail .task-list { flex: 0 0 31%; min-width: 240px; padding: 0 12px; background: var(--surface-subtle); }
  .has-detail .task-list-heading { grid-template-columns: 1fr; background: var(--surface-subtle); }
  .has-detail .task-list-heading > span:not(:first-child) { display: none; }
  .has-detail .task-row { grid-template-columns: 1fr auto; gap: 8px; padding: 15px 10px; }
  .has-detail .task-row strong { grid-column: 1/-1; flex-wrap: wrap; gap: 6px; font-size: 12px; }
  .has-detail .row-due { display: none; }
  .has-detail .row-owner .avatar { display: none; }
  .has-detail .row-status { font-size: 10px; }
  .task-detail { padding: 20px 32px 40px; border-left-color: var(--line-soft); }
  .task-detail > .section-head { font-size: 11px; color: var(--dt3); margin-bottom: 18px; }
  .detail-cadence { display: inline-block; margin-bottom: 10px; }
  .task-title { font-size: 21px; line-height: 1.4; font-weight: 600; margin-bottom: 22px; max-width: 35ch; }
  .task-properties { display: grid; grid-template-columns: 1fr 1fr; gap: 18px 20px; padding: 18px 0; border-block: 1px solid var(--line-soft); margin-bottom: 28px; }
  .task-properties > div { display: flex; flex-direction: column; gap: 7px; min-width: 0; }
  .property-label { font-size: 10px; color: var(--dt3); }
  .property-value { display: flex; align-items: center; gap: 8px; }
  .task-properties select { font-size: 12px; max-width: 150px; }
  .task-properties > button { justify-self: end; align-self: end; }
  .prose { font-size: 13px; line-height: 1.85; color: var(--dt2); max-width: 76ch; margin-top: 24px; }
  .prose :global(h1) { font-size: 23px; color: var(--dt); line-height: 1.4; }
  .prose :global(h2) { font-size: 15px; font-weight: 600; color: var(--dt); margin: 28px 0 10px; }
  .prose :global(h3) { font-size: 13px; font-weight: 600; color: var(--dt); margin: 22px 0 8px; }
  .prose :global(li) { padding-left: 3px; margin-bottom: 6px; }
  .prose :global(strong) { color: var(--dt); font-weight: 600; }
  .prose :global(a) { color: var(--dt); text-decoration-color: var(--dt3); text-underline-offset: 3px; text-decoration-thickness: 1px; }
  .prose :global(a:hover) { text-decoration-color: currentColor; }
  .prose :global(a:focus-visible) { outline: 2px solid var(--dt); outline-offset: 4px; border-radius: 2px; }
  .prose :global(ul:has(> li > a:only-child)) { list-style: none; padding: 0; margin: 12px 0 24px; border-top: 1px solid var(--line-soft); }
  .prose :global(ul > li:has(> a:only-child)) { padding: 0; margin: 0; border-bottom: 1px solid var(--line-soft); }
  .prose :global(li > a:only-child) { display: block; padding: 11px 10px; font-size: 13px; font-weight: 500; text-decoration: none; }
  .prose :global(li > a:only-child:hover) { background: var(--surface-subtle); text-decoration: underline; }
  .prose :global(input[type='checkbox']) { margin-right: 7px; accent-color: #27846c; }
  .people, .notes { padding: 30px 36px; }
  .people h2, .notes h2 { font-size: 16px; font-weight: 600; margin-bottom: 24px; }
  .person { gap: 14px; padding: 20px 10px; border-bottom-color: var(--line-soft); }
  .person > .avatar { width: 34px; height: 34px; font-size: 11px; }
  .person-name { flex: 1; min-width: 0; font-size: 13px; }
  .person-name strong { font-weight: 550; }
  .person-name small { font-size: 11px; }
  .workload { font-size: 12px; color: var(--dt3); white-space: nowrap; }
  .notes textarea { background: var(--surface-subtle); border-color: var(--line-soft); }
  .notes article { max-width: 760px; }
  @container project (max-width: 1000px) {
    .overview { grid-template-columns: minmax(0,1fr) 215px; }
    .overview-main { padding: 26px; }
    .project-summary { padding: 26px 16px; }
    .task-list-heading, .task-row { grid-template-columns: minmax(170px,1fr) 140px 94px; gap: 10px; }
    .task-list-heading > span:last-child, .row-due { display: none; }
    .task-detail { padding: 20px 24px 36px; }
  }
  @container project (max-width: 700px) {
    .project-header { padding: 12px 16px 6px; gap: 8px; }
    .project-header > .badge { display: none; }
    h1 { font-size: 18px; }
    .project-tabs { padding: 0 16px; gap: 20px; }
    .project-tabs button { font-size: 12px; }
    .overview { display: flex; flex-direction: column; }
    .overview-main { padding: 24px 18px; }
    .project-summary { border-left: 0; border-top: 1px solid var(--line-soft); padding: 24px 18px; }
    .project-summary .badge { display: inline-flex; }
    .filters { padding: 12px 16px; }
    .filters .search { max-width: none; }
    .filters > .primary { margin-left: 0; }
    .task-list { padding: 0 10px; }
    .task-list-heading, .task-row { grid-template-columns: 1fr auto; gap: 10px; }
    .task-list-heading > span:not(:first-child) { display: none; }
    .task-row strong { grid-column: 1/-1; flex-wrap: wrap; }
    .task-row .row-owner { grid-column: 1; }
    .task-row .row-status { grid-column: 2; }
    .has-detail .mobile-hidden { display: none; }
    .task-detail { border-left: 0; }
    .has-detail .task-detail { padding: 18px 20px 32px; }
    .task-title { font-size: 20px; }
    .task-properties { gap: 18px 12px; }
    .task-row .row-owner { font-size: 11px; }
    .people, .notes { padding: 24px 18px; }
    .person { padding: 18px 0; gap: 10px; }
  }
</style>
