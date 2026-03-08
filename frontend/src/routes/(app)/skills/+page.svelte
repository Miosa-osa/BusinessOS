<!--
	/skills — Skill Catalog page
	Grid of SORX skill cards with search, pill filters, detail drawer, and tier filter.
	Loads from GET /api/v1/skills, falls back to mock data if backend unavailable.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import SkillCard from '$lib/components/skills/SkillCard.svelte';
	import SkillDecisionCard from '$lib/components/desktop/osa/SkillDecisionCard.svelte';
	import { listSkills, type BackendSkill } from '$lib/api/skills';
	import {
		MOCK_SKILLS,
		type Skill,
		type SkillCategory,
		type SkillTier,
		type SkillExecution,
		categoryLabels,
		categoryColors,
		categoryIcons,
		tierLabels,
		tierColors
	} from '$lib/types/skills';

	// ─── State ────────────────────────────────────────────────────────────────
	let skills = $state<Skill[]>([]);
	let usingMockData = $state(false);
	let isLoadingSkills = $state(true);
	let searchQuery = $state('');
	let selectedCategory = $state<SkillCategory | 'all'>('all');
	let selectedTier = $state<SkillTier | 'all'>('all');

	// Detail drawer
	let selectedSkill = $state<Skill | null>(null);
	let drawerOpen = $state(false);

	// Demo execution for showcasing decision cards
	let demoExecution = $state<SkillExecution | null>(null);

	// ─── Derived ──────────────────────────────────────────────────────────────
	let filteredSkills = $derived.by(() => {
		let result = skills;

		if (searchQuery.trim()) {
			const q = searchQuery.toLowerCase();
			result = result.filter(
				(s) =>
					s.name.toLowerCase().includes(q) ||
					s.description.toLowerCase().includes(q)
			);
		}

		if (selectedCategory !== 'all') {
			result = result.filter((s) => s.category === selectedCategory);
		}

		if (selectedTier !== 'all') {
			result = result.filter((s) => s.tier === selectedTier);
		}

		return result;
	});

	let enabledCount = $derived(skills.filter((s) => s.enabled).length);

	// Category counts for pill badges
	let categoryCounts = $derived.by(() => {
		const counts: Record<string, number> = {};
		for (const s of skills) {
			counts[s.category] = (counts[s.category] || 0) + 1;
		}
		return counts;
	});

	// ─── Handlers ─────────────────────────────────────────────────────────────
	function handleToggle(skill: Skill) {
		skills = skills.map((s) =>
			s.id === skill.id ? { ...s, enabled: !s.enabled } : s
		);
		// Update drawer if open
		if (selectedSkill?.id === skill.id) {
			selectedSkill = { ...skill, enabled: !skill.enabled };
		}
	}

	function handleSkillClick(skill: Skill) {
		selectedSkill = skill;
		drawerOpen = true;
	}

	function closeDrawer() {
		drawerOpen = false;
		selectedSkill = null;
		demoExecution = null;
	}

	function handleTestExecute(skill: Skill) {
		demoExecution = {
			id: crypto.randomUUID(),
			skill,
			action: `Execute ${skill.name}`,
			temperature: skill.tier === 'enterprise' ? 'hot' : skill.tier === 'pro' ? 'warm' : 'cold',
			reasoning: `OSA wants to run ${skill.name}: ${skill.description}`,
			timestamp: new Date()
		};
	}

	function handleDemoApprove() {
		demoExecution = null;
	}

	function handleDemoReject() {
		demoExecution = null;
	}

	const allCategories = Object.entries(categoryLabels) as [SkillCategory, string][];

	const tiers: { value: SkillTier | 'all'; label: string }[] = [
		{ value: 'all', label: 'All Tiers' },
		...Object.entries(tierLabels).map(([value, label]) => ({
			value: value as SkillTier,
			label
		}))
	];

	// ─── Load from API ───────────────────────────────────────────────────────
	/** Map backend skill shape to frontend Skill type */
	function mapBackendSkill(bs: BackendSkill): Skill {
		const nameLower = bs.name.toLowerCase();

		// Match category by keyword anywhere in name or description
		function inferCategory(): SkillCategory {
			const text = `${nameLower} ${bs.description.toLowerCase()}`;
			if (/email|gmail|smtp|inbox|mail/.test(text)) return 'email';
			if (/slack|discord|messaging|chat|notification|alert/.test(text)) return 'messaging';
			if (/crm|contacts|hubspot|salesforce|lead|deal|client/.test(text)) return 'crm';
			if (/calendar|schedule|event|meeting/.test(text)) return 'calendar';
			if (/sync|notion|linear|airtable|import/.test(text)) return 'sync';
			if (/export|sheets|csv|report|analytics|insight|metric/.test(text)) return 'export';
			if (/task|project|dashboard|manage|webhook|automat/.test(text)) return 'automation';
			return 'automation';
		}

		const category = inferCategory();

		// Infer tier from priority (backend uses priority 1-10)
		const tier: SkillTier = bs.priority >= 8 ? 'enterprise' : bs.priority >= 5 ? 'pro' : 'free';

		return {
			id: `skill-${bs.name.replace(/[.\s]/g, '-')}`,
			name: bs.name,
			description: bs.description,
			tier,
			category,
			enabled: true // backend only returns enabled skills
		};
	}

	onMount(async () => {
		try {
			const data = await listSkills();
			if (data.skills && data.skills.length > 0) {
				skills = data.skills.map(mapBackendSkill);
			}
		} catch {
			// Backend unavailable — fall back to mock data
			skills = MOCK_SKILLS;
			usingMockData = true;
		} finally {
			isLoadingSkills = false;
		}
	});
</script>

<div class="flex h-full flex-col bg-white dark:bg-gray-950">
	<!-- Header -->
	<div class="flex-shrink-0 border-b border-gray-200 bg-white px-8 py-6 dark:border-gray-800 dark:bg-gray-950">
		<!-- Breadcrumb -->
		<nav class="mb-4 flex items-center gap-1.5 text-xs text-gray-400 dark:text-gray-500" aria-label="Breadcrumb">
			<a href="/" class="hover:text-gray-600 dark:hover:text-gray-300 transition-colors">Dashboard</a>
			<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2"><polyline points="9 18 15 12 9 6" /></svg>
			<span class="text-gray-600 dark:text-gray-300 font-medium">Skills</span>
		</nav>

		<div class="mb-5 flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Skill Catalog</h1>
				<p class="mt-1 text-sm text-gray-600 dark:text-gray-400">
					{skills.length} skills available &middot; {enabledCount} active
				</p>
			</div>
			<!-- Action Button -->
			<button
				class="btn-pill btn-pill-secondary btn-pill-sm inline-flex items-center gap-2 opacity-60"
				disabled
				title="Coming soon"
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<circle cx="11" cy="11" r="8" />
					<line x1="21" y1="21" x2="16.65" y2="16.65" />
					<line x1="11" y1="8" x2="11" y2="14" />
					<line x1="8" y1="11" x2="14" y2="11" />
				</svg>
				Browse Marketplace
			</button>
		</div>

		<!-- Search Row -->
		<div class="mb-4">
			<div class="relative max-w-md">
				<svg class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
					<circle cx="11" cy="11" r="8" />
					<line x1="21" y1="21" x2="16.65" y2="16.65" />
				</svg>
				<input
					type="text"
					bind:value={searchQuery}
					placeholder="Search skills..."
					class="w-full rounded-lg border border-gray-200 bg-white py-2 pl-10 pr-4 text-sm text-gray-900 placeholder-gray-400 transition-colors focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-gray-700 dark:bg-gray-900 dark:text-white dark:placeholder-gray-500"
				/>
			</div>
		</div>

		<!-- Category Pill Filters -->
		<div class="flex flex-wrap items-center gap-2">
			<button
				onclick={() => (selectedCategory = 'all')}
				class="rounded-full px-3 py-1.5 text-xs font-medium transition-all duration-150
					{selectedCategory === 'all'
						? 'bg-blue-600 text-white shadow-sm dark:bg-blue-500 dark:text-white'
						: 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}"
			>
				All
				<span class="ml-1 text-[10px] opacity-60">{skills.length}</span>
			</button>
			{#each allCategories as [value, label]}
				<button
					onclick={() => (selectedCategory = selectedCategory === value ? 'all' : value)}
					class="inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-medium transition-all duration-150
						{selectedCategory === value
							? 'bg-blue-600 text-white shadow-sm dark:bg-blue-500 dark:text-white'
							: 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}"
				>
					<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d={categoryIcons[value]} />
					</svg>
					{label}
					{#if categoryCounts[value]}
						<span class="text-[10px] opacity-60">{categoryCounts[value]}</span>
					{/if}
				</button>
			{/each}

			<!-- Tier pill group (separated) -->
			<span class="mx-1 h-4 w-px bg-gray-200 dark:bg-gray-700"></span>
			{#each tiers as t}
				<button
					onclick={() => (selectedTier = selectedTier === t.value ? 'all' : t.value)}
					class="rounded-full px-3 py-1.5 text-xs font-medium transition-all duration-150
						{selectedTier === t.value
							? 'bg-blue-600 text-white shadow-sm dark:bg-blue-500 dark:text-white'
							: 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-gray-700'}"
				>
					{t.label}
				</button>
			{/each}
		</div>
	</div>

	<!-- Content -->
	<div class="flex-1 overflow-y-auto px-8 py-6">
		<!-- Mock data fallback banner -->
		{#if usingMockData && !isLoadingSkills}
			<div class="mb-4 flex items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 dark:border-amber-800 dark:bg-amber-950/20">
				<svg class="h-4 w-4 flex-shrink-0 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
					<path d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4.5c-.77-.833-2.694-.833-3.464 0L3.34 16.5c-.77.833.192 2.5 1.732 2.5z" />
				</svg>
				<p class="text-xs text-amber-700 dark:text-amber-300">
					Showing sample data — backend skills endpoint unavailable. Skills will load automatically when connected.
				</p>
			</div>
		{/if}

		{#if filteredSkills.length === 0}
			<!-- Empty State -->
			<div class="flex items-center justify-center py-24">
				<div class="max-w-md text-center">
					<div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-gray-800">
						<svg class="h-8 w-8 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
							<path d="M13 10V3L4 14h7v7l9-11h-7z" />
						</svg>
					</div>
					<p class="mb-2 text-sm font-medium text-gray-900 dark:text-white">No skills found</p>
					<p class="text-sm text-gray-600 dark:text-gray-400">
						{#if searchQuery || selectedCategory !== 'all' || selectedTier !== 'all'}
							Try adjusting your filters.
						{:else}
							No skills are available yet.
						{/if}
					</p>
					{#if searchQuery || selectedCategory !== 'all' || selectedTier !== 'all'}
						<button
							onclick={() => { searchQuery = ''; selectedCategory = 'all'; selectedTier = 'all'; }}
							class="mt-3 text-sm font-medium text-blue-600 hover:text-blue-700 dark:text-blue-400 dark:hover:text-blue-300"
						>
							Clear all filters
						</button>
					{/if}
				</div>
			</div>
		{:else}
			<!-- Skills Grid -->
			<div class="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each filteredSkills as skill (skill.id)}
					<SkillCard
						{skill}
						onToggle={handleToggle}
						onClick={handleSkillClick}
					/>
				{/each}
			</div>

			<!-- Count -->
			<div class="mt-8 text-center">
				<p class="text-sm text-gray-500 dark:text-gray-400">
					Showing {filteredSkills.length} of {skills.length} skills
				</p>
			</div>
		{/if}
	</div>
</div>

<!-- Skill Detail Drawer (Slide-over) -->
{#if drawerOpen && selectedSkill}
	<!-- Backdrop -->
	<div
		class="fixed inset-0 z-40 bg-black/20 backdrop-blur-sm transition-opacity"
		onclick={closeDrawer}
		onkeydown={(e) => { if (e.key === 'Escape') closeDrawer(); }}
		role="button"
		tabindex="-1"
		aria-label="Close drawer"
	></div>

	<!-- Drawer Panel -->
	<div class="fixed inset-y-0 right-0 z-50 flex w-full max-w-md flex-col border-l border-gray-200 bg-white shadow-xl dark:border-gray-700 dark:bg-gray-900 animate-in slide-in-from-right duration-200">
		<!-- Drawer Header -->
		<div class="flex items-start justify-between border-b border-gray-200 px-6 py-5 dark:border-gray-700">
			<div class="flex items-center gap-3">
				<div class="flex h-10 w-10 items-center justify-center rounded-lg
					{selectedSkill.category === 'email' ? 'bg-blue-100 dark:bg-blue-900/30' :
					selectedSkill.category === 'messaging' ? 'bg-violet-100 dark:bg-violet-900/30' :
					selectedSkill.category === 'crm' ? 'bg-orange-100 dark:bg-orange-900/30' :
					selectedSkill.category === 'calendar' ? 'bg-teal-100 dark:bg-teal-900/30' :
					selectedSkill.category === 'sync' ? 'bg-cyan-100 dark:bg-cyan-900/30' :
					selectedSkill.category === 'automation' ? 'bg-pink-100 dark:bg-pink-900/30' :
					'bg-gray-100 dark:bg-gray-800'}">
					<svg class="h-5 w-5
						{selectedSkill.category === 'email' ? 'text-blue-600 dark:text-blue-400' :
						selectedSkill.category === 'messaging' ? 'text-violet-600 dark:text-violet-400' :
						selectedSkill.category === 'crm' ? 'text-orange-600 dark:text-orange-400' :
						selectedSkill.category === 'calendar' ? 'text-teal-600 dark:text-teal-400' :
						selectedSkill.category === 'sync' ? 'text-cyan-600 dark:text-cyan-400' :
						selectedSkill.category === 'automation' ? 'text-pink-600 dark:text-pink-400' :
						'text-gray-600 dark:text-gray-400'}" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
						<path d={categoryIcons[selectedSkill.category]} />
					</svg>
				</div>
				<div>
					<h2 class="font-mono text-base font-semibold text-gray-900 dark:text-white">{selectedSkill.name}</h2>
					<div class="mt-1 flex items-center gap-1.5">
						<span class="rounded-full border px-2 py-0.5 text-[10px] font-medium {categoryColors[selectedSkill.category]}">
							{categoryLabels[selectedSkill.category]}
						</span>
						<span class="rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase {tierColors[selectedSkill.tier]}">
							{tierLabels[selectedSkill.tier]}
						</span>
					</div>
				</div>
			</div>
			<button
				onclick={closeDrawer}
				class="btn-pill btn-pill-ghost btn-pill-icon"
				aria-label="Close detail panel"
			>
				<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<line x1="18" y1="6" x2="6" y2="18" />
					<line x1="6" y1="6" x2="18" y2="18" />
				</svg>
			</button>
		</div>

		<!-- Drawer Body -->
		<div class="flex-1 overflow-y-auto px-6 py-5">
			<!-- Status Banner -->
			<div class="mb-5 flex items-center justify-between rounded-lg border px-4 py-3
				{selectedSkill.enabled
					? 'border-green-200 bg-green-50 dark:border-green-800 dark:bg-green-950/20'
					: 'border-gray-200 bg-gray-50 dark:border-gray-700 dark:bg-gray-800/50'}">
				<div class="flex items-center gap-2">
					<span class="h-2.5 w-2.5 rounded-full {selectedSkill.enabled ? 'bg-green-500' : 'bg-gray-400'}"></span>
					<span class="text-sm font-medium {selectedSkill.enabled ? 'text-green-700 dark:text-green-300' : 'text-gray-600 dark:text-gray-400'}">
						{selectedSkill.enabled ? 'Active' : 'Inactive'}
					</span>
				</div>
				<button
					onclick={() => handleToggle(selectedSkill!)}
					class="rounded-lg px-3 py-1 text-xs font-medium transition-colors
						{selectedSkill.enabled
							? 'bg-white text-gray-700 shadow-sm hover:bg-gray-50 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600'
							: 'bg-blue-600 text-white hover:bg-blue-700'}"
				>
					{selectedSkill.enabled ? 'Disable' : 'Enable'}
				</button>
			</div>

			<!-- Description -->
			<div class="mb-6">
				<h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Description</h3>
				<p class="text-sm leading-relaxed text-gray-700 dark:text-gray-300">{selectedSkill.description}</p>
			</div>

			<!-- Details Grid -->
			<div class="mb-6">
				<h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Details</h3>
				<div class="grid grid-cols-2 gap-3">
					<div class="rounded-lg border border-gray-200 px-3 py-2.5 dark:border-gray-700">
						<p class="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">Tier</p>
						<p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white capitalize">{selectedSkill.tier}</p>
					</div>
					<div class="rounded-lg border border-gray-200 px-3 py-2.5 dark:border-gray-700">
						<p class="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">Category</p>
						<p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">{categoryLabels[selectedSkill.category]}</p>
					</div>
					<div class="rounded-lg border border-gray-200 px-3 py-2.5 dark:border-gray-700">
						<p class="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">Temperature</p>
						<p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">
							{selectedSkill.tier === 'enterprise' ? 'Hot' : selectedSkill.tier === 'pro' ? 'Warm' : 'Cold'}
						</p>
					</div>
					<div class="rounded-lg border border-gray-200 px-3 py-2.5 dark:border-gray-700">
						<p class="text-[10px] font-medium uppercase tracking-wider text-gray-400 dark:text-gray-500">Approval</p>
						<p class="mt-0.5 text-sm font-semibold text-gray-900 dark:text-white">
							{selectedSkill.tier === 'enterprise' ? 'Required' : selectedSkill.tier === 'pro' ? 'Confirm' : 'Auto-run'}
						</p>
					</div>
				</div>
			</div>

			<!-- Usage Stats (placeholder) -->
			<div class="mb-6">
				<h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Usage</h3>
				<div class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-6 text-center dark:border-gray-700 dark:bg-gray-800/50">
					<p class="text-xs text-gray-400 dark:text-gray-500">Usage stats available when connected to backend</p>
				</div>
			</div>

			<!-- Config (placeholder) -->
			{#if selectedSkill.config && Object.keys(selectedSkill.config).length > 0}
				<div class="mb-6">
					<h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Configuration</h3>
					<pre class="rounded-lg border border-gray-200 bg-gray-50 px-4 py-3 text-xs text-gray-700 dark:border-gray-700 dark:bg-gray-800/50 dark:text-gray-300 overflow-x-auto">{JSON.stringify(selectedSkill.config, null, 2)}</pre>
				</div>
			{/if}

			<!-- Decision Card Demo -->
			{#if demoExecution}
				<div class="mb-6">
					<h3 class="mb-3 text-xs font-semibold uppercase tracking-wider text-gray-400 dark:text-gray-500">Execution Preview</h3>
					<SkillDecisionCard
						execution={demoExecution}
						onApprove={handleDemoApprove}
						onReject={handleDemoReject}
					/>
				</div>
			{/if}
		</div>

		<!-- Drawer Footer -->
		<div class="flex items-center gap-3 border-t border-gray-200 px-6 py-4 dark:border-gray-700">
			<button
				onclick={() => handleTestExecute(selectedSkill!)}
				disabled={!selectedSkill.enabled}
				class="flex flex-1 items-center justify-center gap-2 rounded-lg px-4 py-2.5 text-sm font-medium transition-colors disabled:cursor-not-allowed disabled:opacity-40
					{selectedSkill.enabled
						? 'bg-blue-600 text-white hover:bg-blue-700'
						: 'bg-gray-200 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}"
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<polygon points="5 3 19 12 5 21 5 3" />
				</svg>
				Test Execute
			</button>
			<button
				onclick={closeDrawer}
				class="btn-pill btn-pill-ghost btn-pill-sm"
			>
				Close
			</button>
		</div>
	</div>
{/if}
