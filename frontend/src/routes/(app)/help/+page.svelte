<script lang="ts">
	import { page } from '$app/stores';
	import {
		BookOpen,
		Rocket,
		Brain,
		Keyboard,
		Monitor,
		Users,
		Network,
		FileText,
		MessageSquare,
		Search,
		Grid3X3,
		Mic,
		Sparkles,
		Target,
		BarChart3,
		PenTool,
		Hammer,
		Briefcase,
		Calendar,
		CheckSquare,
		FolderKanban,
		Settings,
		ChevronRight,
		ChevronDown,
		Mail,
		Github
	} from 'lucide-svelte';

	import HelpGettingStarted from '$lib/components/help/HelpGettingStarted.svelte';
	import HelpAIFeatures from '$lib/components/help/HelpAIFeatures.svelte';
	import HelpDesktop from '$lib/components/help/HelpDesktop.svelte';
	import HelpModules from '$lib/components/help/HelpModules.svelte';
	import HelpKnowledge from '$lib/components/help/HelpKnowledge.svelte';
	import HelpClients from '$lib/components/help/HelpClients.svelte';
	import HelpShortcuts from '$lib/components/help/HelpShortcuts.svelte';
	import HelpVoice from '$lib/components/help/HelpVoice.svelte';
	import HelpSearch from '$lib/components/help/HelpSearch.svelte';
	import HelpSettings from '$lib/components/help/HelpSettings.svelte';

	const isEmbedded = $derived($page.url.searchParams.get('embed') === 'true');

	interface Section {
		id: string;
		title: string;
		icon: any;
		subsections?: { id: string; title: string }[];
	}

	const sections: Section[] = [
		{
			id: 'getting-started',
			title: 'Getting Started',
			icon: Rocket,
			subsections: [
				{ id: 'overview', title: 'Platform Overview' },
				{ id: 'first-steps', title: 'First Steps' },
				{ id: 'core-concepts', title: 'Core Concepts' }
			]
		},
		{
			id: 'ai-features',
			title: 'AI Features',
			icon: Sparkles,
			subsections: [
				{ id: 'focus-modes', title: 'Focus Modes' },
				{ id: 'ai-chat', title: 'AI Chat' },
				{ id: 'ai-contexts', title: 'AI Contexts' }
			]
		},
		{
			id: 'desktop',
			title: 'Desktop Environment',
			icon: Monitor,
			subsections: [
				{ id: 'windows', title: 'Window Management' },
				{ id: 'dock', title: 'Dock & Navigation' },
				{ id: 'customization', title: 'Customization' },
				{ id: '3d-mode', title: '3D Desktop Mode' }
			]
		},
		{
			id: 'modules',
			title: 'Core Modules',
			icon: Grid3X3,
			subsections: [
				{ id: 'dashboard-mod', title: 'Dashboard' },
				{ id: 'chat-mod', title: 'Chat' },
				{ id: 'tasks-mod', title: 'Tasks' },
				{ id: 'projects-mod', title: 'Projects' },
				{ id: 'team-mod', title: 'Team' },
				{ id: 'calendar-mod', title: 'Calendar' }
			]
		},
		{
			id: 'knowledge',
			title: 'Knowledge & Data',
			icon: Brain,
			subsections: [
				{ id: 'nodes', title: 'Nodes System' },
				{ id: 'contexts', title: 'Contexts (Documents)' },
				{ id: 'knowledge-graph', title: 'Knowledge Graph' }
			]
		},
		{
			id: 'clients',
			title: 'Clients & CRM',
			icon: Briefcase,
			subsections: [
				{ id: 'client-profiles', title: 'Client Profiles' },
				{ id: 'deals-pipeline', title: 'Deals Pipeline' },
				{ id: 'interactions', title: 'Interactions' }
			]
		},
		{ id: 'shortcuts', title: 'Keyboard Shortcuts', icon: Keyboard },
		{ id: 'voice', title: 'Voice Features', icon: Mic },
		{ id: 'search', title: 'Spotlight Search', icon: Search },
		{ id: 'settings-help', title: 'Settings & Config', icon: Settings }
	];

	let activeSection = $state('getting-started');
	let expandedSections = $state<Set<string>>(new Set(['getting-started', 'ai-features', 'modules']));
	let searchQuery = $state('');

	function scrollToSection(id: string) {
		activeSection = id;
		const element = document.getElementById(id);
		element?.scrollIntoView({ behavior: 'smooth', block: 'start' });
	}

	function toggleSection(id: string) {
		const newSet = new Set(expandedSections);
		if (newSet.has(id)) {
			newSet.delete(id);
		} else {
			newSet.add(id);
		}
		expandedSections = newSet;
	}

	let filteredSections = $derived.by(() => {
		if (!searchQuery.trim()) return sections;
		const query = searchQuery.toLowerCase();
		return sections.filter(s =>
			s.title.toLowerCase().includes(query) ||
			s.subsections?.some(sub => sub.title.toLowerCase().includes(query))
		);
	});
</script>

<svelte:head>
	<title>Help Center - Business OS</title>
</svelte:head>

<div class="help-page" class:embedded={isEmbedded}>
	<!-- Sidebar Navigation -->
	<aside class="help-sidebar">
		<div class="sidebar-header">
			<div class="header-icon">
				<BookOpen size={20} />
			</div>
			<div class="header-text">
				<h2>Help Center</h2>
				<span class="version">Business OS v1.0</span>
			</div>
		</div>

		<div class="sidebar-search">
			<Search size={14} class="search-icon" />
			<input
				type="text"
				placeholder="Search documentation..."
				bind:value={searchQuery}
			/>
		</div>

		<nav class="sidebar-nav">
			{#each filteredSections as section}
				<div class="nav-section">
					<button
						class="nav-item"
						class:active={activeSection === section.id}
						class:has-subsections={section.subsections}
						onclick={() => {
							if (section.subsections) {
								toggleSection(section.id);
							}
							scrollToSection(section.id);
						}}
					>
						<span class="nav-icon">
							<svelte:component this={section.icon} size={16} />
						</span>
						<span class="nav-text">{section.title}</span>
						{#if section.subsections}
							<span class="nav-chevron">
								{#if expandedSections.has(section.id)}
									<ChevronDown size={14} />
								{:else}
									<ChevronRight size={14} />
								{/if}
							</span>
						{/if}
					</button>

					{#if section.subsections && expandedSections.has(section.id)}
						<div class="nav-subsections">
							{#each section.subsections as sub}
								<button
									class="nav-subitem"
									class:active={activeSection === sub.id}
									onclick={() => scrollToSection(sub.id)}
								>
									{sub.title}
								</button>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		</nav>

		<div class="sidebar-footer">
			<a href="mailto:roberto@osa.dev" class="footer-link">
				<Mail size={14} />
				<span>Contact Support</span>
			</a>
			<a href="https://github.com/Miosa-osa/BusinessOS" target="_blank" class="footer-link">
				<Github size={14} />
				<span>View on GitHub</span>
			</a>
		</div>
	</aside>

	<!-- Main Content -->
	<main class="help-content">
		<HelpGettingStarted />
		<HelpAIFeatures />
		<HelpDesktop />
		<HelpModules />
		<HelpKnowledge />
		<HelpClients />
		<HelpShortcuts />
		<HelpVoice />
		<HelpSearch />
		<HelpSettings />

		<footer class="help-footer">
			<div class="footer-content">
				<h3>Need More Help?</h3>
				<p>Can't find what you're looking for? Reach out to our support team.</p>
				<div class="footer-actions">
					<a href="mailto:roberto@osa.dev" class="footer-btn primary">
						<Mail size={16} />
						Contact Support
					</a>
					<a href="https://github.com/Miosa-osa/BusinessOS" target="_blank" class="footer-btn">
						<Github size={16} />
						GitHub
					</a>
				</div>
			</div>
		</footer>
	</main>
</div>

<style>
	.help-page {
		display: flex;
		height: 100vh;
		background: #F8FAFC;
		overflow: hidden;
	}

	.help-page.embedded {
		height: 100%;
	}

	/* ===== SIDEBAR ===== */
	.help-sidebar {
		width: 280px;
		background: white;
		border-right: 1px solid #E2E8F0;
		display: flex;
		flex-direction: column;
		position: sticky;
		top: 0;
		height: 100vh;
	}

	.embedded .help-sidebar {
		height: 100%;
	}

	.sidebar-header {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 20px;
		border-bottom: 1px solid #E2E8F0;
	}

	.header-icon {
		width: 40px;
		height: 40px;
		background: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%);
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
	}

	.header-text h2 {
		font-size: 16px;
		font-weight: 700;
		color: #1E293B;
		margin: 0;
	}

	.header-text .version {
		font-size: 11px;
		color: #94A3B8;
	}

	.sidebar-search {
		padding: 12px 16px;
		border-bottom: 1px solid #E2E8F0;
		position: relative;
	}

	.sidebar-search :global(.search-icon) {
		position: absolute;
		left: 28px;
		top: 50%;
		transform: translateY(-50%);
		color: #94A3B8;
	}

	.sidebar-search input {
		width: 100%;
		padding: 10px 12px 10px 36px;
		border: 1px solid #E2E8F0;
		border-radius: 8px;
		font-size: 13px;
		background: #F8FAFC;
		transition: all 0.2s;
	}

	.sidebar-search input:focus {
		outline: none;
		border-color: #6366F1;
		background: white;
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.sidebar-nav {
		flex: 1;
		padding: 12px;
		overflow-y: auto;
	}

	.nav-section {
		margin-bottom: 4px;
	}

	.nav-item {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		padding: 10px 12px;
		border: none;
		background: none;
		border-radius: 8px;
		cursor: pointer;
		text-align: left;
		transition: all 0.15s;
		color: #475569;
	}

	.nav-item:hover {
		background: #F1F5F9;
		color: #1E293B;
	}

	.nav-item.active {
		background: #EEF2FF;
		color: #4F46E5;
	}

	.nav-icon {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 24px;
		height: 24px;
		color: inherit;
	}

	.nav-text {
		flex: 1;
		font-size: 13px;
		font-weight: 500;
	}

	.nav-chevron {
		display: flex;
		color: #94A3B8;
	}

	.nav-subsections {
		margin-left: 36px;
		padding: 4px 0;
	}

	.nav-subitem {
		display: block;
		width: 100%;
		padding: 8px 12px;
		border: none;
		background: none;
		border-radius: 6px;
		cursor: pointer;
		text-align: left;
		font-size: 12px;
		color: #64748B;
		transition: all 0.15s;
	}

	.nav-subitem:hover {
		background: #F1F5F9;
		color: #1E293B;
	}

	.nav-subitem.active {
		background: #EEF2FF;
		color: #4F46E5;
		font-weight: 500;
	}

	.sidebar-footer {
		padding: 16px;
		border-top: 1px solid #E2E8F0;
	}

	.footer-link {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 12px;
		color: #64748B;
		text-decoration: none;
		font-size: 12px;
		border-radius: 6px;
		transition: all 0.15s;
	}

	.footer-link:hover {
		background: #F1F5F9;
		color: #4F46E5;
	}

	/* ===== MAIN CONTENT ===== */
	.help-content {
		flex: 1;
		padding: 40px 60px;
		max-width: 1000px;
		overflow-y: auto;
	}

	/* ===== SECTION SHARED STYLES (global so child components can use them) ===== */
	:global(.help-section) {
		margin-bottom: 80px;
	}

	:global(.section-header) {
		display: flex;
		align-items: flex-start;
		gap: 20px;
		margin-bottom: 32px;
	}

	:global(.section-icon) {
		width: 56px;
		height: 56px;
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		flex-shrink: 0;
	}

	:global(.gradient-blue) { background: linear-gradient(135deg, #3B82F6 0%, #1D4ED8 100%); }
	:global(.gradient-purple) { background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%); }
	:global(.gradient-green) { background: linear-gradient(135deg, #10B981 0%, #059669 100%); }
	:global(.gradient-orange) { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
	:global(.gradient-teal) { background: linear-gradient(135deg, #14B8A6 0%, #0D9488 100%); }
	:global(.gradient-indigo) { background: linear-gradient(135deg, #6366F1 0%, #4F46E5 100%); }
	:global(.gradient-gray) { background: linear-gradient(135deg, #64748B 0%, #475569 100%); }
	:global(.gradient-red) { background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%); }
	:global(.gradient-yellow) { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
	:global(.gradient-slate) { background: linear-gradient(135deg, #475569 0%, #334155 100%); }

	:global(.section-header h1) {
		font-size: 28px;
		font-weight: 700;
		color: #0F172A;
		margin: 0 0 4px;
	}

	:global(.section-subtitle) {
		font-size: 15px;
		color: #64748B;
		margin: 0;
	}

	:global(.subsection) {
		margin-bottom: 48px;
	}

	:global(.subsection h2) {
		font-size: 20px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 16px;
		padding-bottom: 12px;
		border-bottom: 1px solid #E2E8F0;
	}

	:global(.subsection-intro) {
		font-size: 14px;
		color: #64748B;
		line-height: 1.7;
		margin-bottom: 24px;
	}

	/* ===== CARDS & GRIDS ===== */
	:global(.intro-card) {
		background: linear-gradient(135deg, #EEF2FF 0%, #E0E7FF 100%);
		border: 1px solid #C7D2FE;
		border-radius: 12px;
		padding: 24px;
		margin-bottom: 40px;
	}

	:global(.intro-card p) {
		font-size: 15px;
		color: #3730A3;
		line-height: 1.7;
		margin: 0;
	}

	:global(.feature-grid) {
		display: grid;
		gap: 20px;
	}

	:global(.feature-grid.four-col) { grid-template-columns: repeat(4, 1fr); }
	:global(.feature-grid.three-col) { grid-template-columns: repeat(3, 1fr); }
	:global(.feature-grid.two-col) { grid-template-columns: repeat(2, 1fr); }

	:global(.feature-card) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
		transition: all 0.2s;
	}

	:global(.feature-card:hover) {
		border-color: #CBD5E1;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
	}

	:global(.feature-card.compact) {
		padding: 20px;
		text-align: center;
	}

	:global(.card-icon) {
		width: 44px;
		height: 44px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 16px;
	}

	:global(.card-icon.blue) { background: #DBEAFE; color: #2563EB; }
	:global(.card-icon.purple) { background: #EDE9FE; color: #7C3AED; }
	:global(.card-icon.green) { background: #D1FAE5; color: #059669; }
	:global(.card-icon.orange) { background: #FED7AA; color: #EA580C; }

	:global(.card-icon-inline) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	:global(.feature-card h3),
	:global(.feature-card h4) {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	:global(.feature-card p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== STEPS ===== */
	:global(.steps-list) {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	:global(.step-item) {
		display: flex;
		gap: 16px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 20px;
	}

	:global(.step-number) {
		width: 32px;
		height: 32px;
		background: #6366F1;
		color: white;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 14px;
		font-weight: 600;
		flex-shrink: 0;
	}

	:global(.step-content h4) {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	:global(.step-content p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== CONCEPT GRID ===== */
	:global(.concept-grid) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 16px;
	}

	:global(.concept-card) {
		background: #F8FAFC;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 20px;
	}

	:global(.concept-card h4) {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	:global(.concept-card p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== MODE CARDS ===== */
	:global(.mode-grid) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
	}

	:global(.mode-card) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
		border-top: 3px solid;
	}

	:global(.mode-card.research) { border-top-color: #3B82F6; }
	:global(.mode-card.analyze) { border-top-color: #8B5CF6; }
	:global(.mode-card.write) { border-top-color: #10B981; }
	:global(.mode-card.build) { border-top-color: #F59E0B; }

	:global(.mode-header) {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	:global(.mode-card.research .mode-header) { color: #3B82F6; }
	:global(.mode-card.analyze .mode-header) { color: #8B5CF6; }
	:global(.mode-card.write .mode-header) { color: #10B981; }
	:global(.mode-card.build .mode-header) { color: #F59E0B; }

	:global(.mode-header h3) {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0;
	}

	:global(.mode-card > p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0 0 16px;
	}

	:global(.mode-options) {
		background: #F8FAFC;
		border-radius: 8px;
		padding: 12px 16px;
		margin-bottom: 12px;
	}

	:global(.mode-options h5) {
		font-size: 11px;
		font-weight: 600;
		color: #94A3B8;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin: 0 0 8px;
	}

	:global(.mode-options ul) {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	:global(.mode-options li) {
		font-size: 12px;
		color: #475569;
		line-height: 1.8;
	}

	:global(.mode-example) {
		background: #FEF3C7;
		border-radius: 8px;
		padding: 12px;
		font-size: 12px;
		color: #92400E;
		font-style: italic;
	}

	:global(.example-label) {
		font-weight: 600;
		font-style: normal;
	}

	/* ===== INFO CARDS ===== */
	:global(.info-card) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	:global(.info-card.highlight) {
		background: linear-gradient(135deg, #F0FDF4 0%, #DCFCE7 100%);
		border-color: #86EFAC;
	}

	:global(.info-card h4) {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 12px;
	}

	:global(.info-card ul),
	:global(.info-card ol) {
		margin: 0;
		padding-left: 20px;
	}

	:global(.info-card li) {
		font-size: 13px;
		color: #475569;
		line-height: 1.8;
	}

	:global(.info-card code) {
		background: #F1F5F9;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		color: #6366F1;
	}

	/* ===== TIP CARD ===== */
	:global(.tip-card) {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: #FEF9C3;
		border: 1px solid #FDE047;
		border-radius: 10px;
		padding: 16px 20px;
		margin-top: 20px;
		color: #854D0E;
	}

	:global(.tip-card div) {
		font-size: 13px;
		line-height: 1.6;
	}

	/* ===== KEYBOARD HINTS ===== */
	:global(.keyboard-hint) {
		display: flex;
		align-items: center;
		gap: 10px;
		background: #F1F5F9;
		border-radius: 8px;
		padding: 12px 16px;
		margin-top: 16px;
		color: #475569;
		font-size: 13px;
	}

	:global(kbd) {
		display: inline-block;
		padding: 3px 8px;
		background: white;
		border: 1px solid #D1D5DB;
		border-radius: 5px;
		font-family: system-ui, -apple-system, sans-serif;
		font-size: 12px;
		font-weight: 500;
		color: #374151;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}

	/* ===== CUSTOMIZATION CARDS ===== */
	:global(.customization-card) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	:global(.customization-card > svg) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	:global(.customization-card h4) {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	:global(.customization-card > p) {
		font-size: 13px;
		color: #64748B;
		margin: 0 0 12px;
	}

	:global(.customization-card ul) {
		margin: 0;
		padding-left: 18px;
	}

	:global(.customization-card li) {
		font-size: 12px;
		color: #475569;
		line-height: 1.8;
	}

	/* ===== MODULES SHOWCASE ===== */
	:global(.modules-showcase) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
		margin-bottom: 40px;
	}

	:global(.module-detail) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	:global(.module-header) {
		display: flex;
		align-items: center;
		gap: 14px;
		margin-bottom: 14px;
	}

	:global(.module-icon) {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	:global(.module-header h3) {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0;
	}

	:global(.module-tagline) {
		font-size: 12px;
		color: #94A3B8;
	}

	:global(.module-detail > p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0 0 16px;
	}

	:global(.module-features) {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	:global(.module-features span) {
		padding: 4px 10px;
		background: #F1F5F9;
		border-radius: 100px;
		font-size: 11px;
		font-weight: 500;
		color: #475569;
	}

	:global(.more-modules h3) {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 16px;
	}

	:global(.mini-modules-grid) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 12px;
	}

	:global(.mini-module) {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 16px;
	}

	:global(.mini-icon) {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	:global(.mini-module h5) {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 2px;
	}

	:global(.mini-module p) {
		font-size: 12px;
		color: #64748B;
		line-height: 1.5;
		margin: 0;
	}

	/* ===== NODES ===== */
	:global(.node-types) {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
		margin-bottom: 24px;
	}

	:global(.node-type) {
		text-align: center;
		padding: 20px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
	}

	:global(.node-icon) {
		width: 48px;
		height: 48px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 12px;
	}

	:global(.node-icon.business) { background: #DBEAFE; color: #2563EB; }
	:global(.node-icon.project) { background: #EDE9FE; color: #7C3AED; }
	:global(.node-icon.learning) { background: #D1FAE5; color: #059669; }
	:global(.node-icon.operational) { background: #FED7AA; color: #EA580C; }

	:global(.node-type h4) {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	:global(.node-type p) {
		font-size: 12px;
		color: #64748B;
		margin: 0;
	}

	/* ===== PIPELINE ===== */
	:global(.pipeline-stages) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	:global(.stage) {
		text-align: center;
		flex: 1;
	}

	:global(.stage-dot) {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		margin: 0 auto 8px;
	}

	:global(.stage h5) {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	:global(.stage p) {
		font-size: 11px;
		color: #64748B;
		margin: 0;
	}

	:global(.stage-arrow) {
		color: #CBD5E1;
		font-size: 20px;
		padding: 0 8px;
	}

	/* ===== SHORTCUTS ===== */
	:global(.shortcuts-container) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 24px;
	}

	:global(.shortcut-category h3) {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 12px;
	}

	:global(.shortcut-list) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		overflow: hidden;
	}

	:global(.shortcut-row) {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #F1F5F9;
	}

	:global(.shortcut-row:last-child) {
		border-bottom: none;
	}

	:global(.shortcut-row .keys) {
		width: 180px;
		flex-shrink: 0;
	}

	:global(.shortcut-row .desc) {
		font-size: 13px;
		color: #64748B;
	}

	/* ===== PRIVACY NOTE ===== */
	:global(.privacy-note) {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: #F0FDF4;
		border: 1px solid #86EFAC;
		border-radius: 10px;
		padding: 16px 20px;
		margin-top: 24px;
		color: #166534;
	}

	:global(.privacy-note p) {
		font-size: 13px;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== SPOTLIGHT PREVIEW ===== */
	:global(.spotlight-preview) {
		margin-bottom: 24px;
	}

	:global(.spotlight-mock) {
		max-width: 500px;
		margin: 0 auto;
	}

	:global(.mock-search) {
		display: flex;
		align-items: center;
		gap: 12px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 16px 20px;
		box-shadow: 0 8px 30px rgba(0, 0, 0, 0.1);
		color: #94A3B8;
		font-size: 15px;
	}

	/* ===== COMMANDS ===== */
	:global(.commands-section) {
		margin-top: 24px;
	}

	:global(.commands-section h3) {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	:global(.commands-section > p) {
		font-size: 13px;
		color: #64748B;
		margin: 0 0 16px;
	}

	:global(.commands-grid) {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	:global(.command) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 8px;
		padding: 12px 16px;
		font-size: 13px;
		color: #475569;
	}

	:global(.command code) {
		background: #F1F5F9;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		color: #6366F1;
		margin-right: 6px;
	}

	/* ===== SETTINGS GRID ===== */
	:global(.settings-grid) {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
	}

	:global(.setting-card) {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	:global(.setting-card > svg) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	:global(.setting-card h4) {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	:global(.setting-card p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== FEATURE LIST ===== */
	:global(.feature-list) {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	:global(.feature-item) {
		display: flex;
		gap: 16px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 20px;
	}

	:global(.item-icon) {
		width: 40px;
		height: 40px;
		background: #EEF2FF;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: #6366F1;
		flex-shrink: 0;
	}

	:global(.item-content h4) {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	:global(.item-content p) {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== FOOTER ===== */
	.help-footer {
		margin-top: 60px;
		padding-top: 40px;
		border-top: 1px solid #E2E8F0;
	}

	.footer-content {
		text-align: center;
		max-width: 400px;
		margin: 0 auto;
	}

	.footer-content h3 {
		font-size: 18px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.footer-content > p {
		font-size: 14px;
		color: #64748B;
		margin: 0 0 24px;
	}

	.footer-actions {
		display: flex;
		justify-content: center;
		gap: 12px;
	}

	.footer-btn {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 12px 20px;
		border-radius: 10px;
		font-size: 14px;
		font-weight: 500;
		text-decoration: none;
		transition: all 0.2s;
		background: #F1F5F9;
		color: #475569;
	}

	.footer-btn:hover {
		background: #E2E8F0;
		color: #1E293B;
	}

	.footer-btn.primary {
		background: #6366F1;
		color: white;
	}

	.footer-btn.primary:hover {
		background: #4F46E5;
	}

	/* ===== RESPONSIVE ===== */
	@media (max-width: 1200px) {
		:global(.feature-grid.four-col) { grid-template-columns: repeat(2, 1fr); }
		:global(.node-types) { grid-template-columns: repeat(2, 1fr); }
		:global(.modules-showcase) { grid-template-columns: 1fr; }
	}

	@media (max-width: 900px) {
		.help-content {
			padding: 30px;
		}

		:global(.feature-grid.three-col),
		:global(.feature-grid.two-col),
		:global(.mode-grid),
		:global(.concept-grid),
		:global(.shortcuts-container),
		:global(.settings-grid),
		:global(.commands-grid),
		:global(.mini-modules-grid) {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.help-sidebar {
			display: none;
		}

		.help-content {
			padding: 20px;
		}

		:global(.section-header) {
			flex-direction: column;
			align-items: flex-start;
		}

		:global(.pipeline-stages) {
			flex-direction: column;
			gap: 16px;
		}

		:global(.stage-arrow) {
			transform: rotate(90deg);
		}
	}

	/* ===== DARK MODE ===== */
	:global(.dark) .help-page {
		background: #0f0f0f;
	}

	:global(.dark) .help-sidebar {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .sidebar-header {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .header-text h2 {
		color: #fff;
	}

	:global(.dark) .header-text .version {
		color: #666;
	}

	:global(.dark) .sidebar-search {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .sidebar-search input {
		background: #2a2a2a;
		border-color: rgba(255, 255, 255, 0.1);
		color: #fff;
	}

	:global(.dark) .sidebar-search input::placeholder {
		color: #666;
	}

	:global(.dark) .sidebar-search input:focus {
		border-color: #6366F1;
		background: #333;
	}

	:global(.dark) .sidebar-search :global(.search-icon) {
		color: #666;
	}

	:global(.dark) .sidebar-footer {
		border-color: rgba(255, 255, 255, 0.1);
		background: #1a1a1a;
	}

	:global(.dark) .footer-link {
		color: #888;
	}

	:global(.dark) .footer-link:hover {
		color: #6366F1;
	}

	:global(.dark) .nav-item {
		color: #aaa;
	}

	:global(.dark) .nav-item:hover {
		background: rgba(255, 255, 255, 0.05);
		color: #fff;
	}

	:global(.dark) .nav-item.active {
		background: rgba(99, 102, 241, 0.15);
		color: #818cf8;
	}

	:global(.dark) .nav-chevron {
		color: #666;
	}

	:global(.dark) .nav-subitem {
		color: #777;
	}

	:global(.dark) .nav-subitem:hover {
		color: #aaa;
	}

	:global(.dark) .nav-subitem.active {
		color: #818cf8;
	}

	:global(.dark) .help-footer {
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .footer-content h3 {
		color: #fff;
	}

	:global(.dark) .footer-content > p {
		color: #888;
	}

	:global(.dark) .footer-btn {
		background: #2a2a2a;
		color: #aaa;
	}

	:global(.dark) .footer-btn:hover {
		background: #333;
		color: #fff;
	}

	:global(.dark) .footer-btn.primary {
		background: #6366F1;
		color: white;
	}

	:global(.dark) .footer-btn.primary:hover {
		background: #4F46E5;
	}

	/* ===== DARK MODE - Shared components ===== */
	:global(.dark .section-header h1) { color: #fff; }
	:global(.dark .section-subtitle) { color: #888; }
	:global(.dark .subsection h2) { color: #fff; }
	:global(.dark .subsection-intro) { color: #aaa; }

	:global(.dark .intro-card) {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .intro-card p) { color: #bbb; }

	:global(.dark .feature-card) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .feature-card h3),
	:global(.dark .feature-card h4) { color: #fff; }
	:global(.dark .feature-card p) { color: #aaa; }
	:global(.dark .feature-card:hover) { border-color: rgba(99, 102, 241, 0.4); }

	:global(.dark .step-item) {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .step-number) { background: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%); }
	:global(.dark .step-content h4) { color: #fff; }
	:global(.dark .step-content p) { color: #aaa; }

	:global(.dark .concept-card) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .concept-card h4) { color: #fff; }
	:global(.dark .concept-card p) { color: #aaa; }

	:global(.dark .mode-card) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .mode-header h3) { color: #fff; }
	:global(.dark .mode-card > p) { color: #aaa; }
	:global(.dark .mode-options) { background: rgba(255, 255, 255, 0.03); }
	:global(.dark .mode-options h5) { color: #888; }
	:global(.dark .mode-options li) { color: #aaa; }
	:global(.dark .mode-example) {
		background: rgba(255, 255, 255, 0.05);
		color: #aaa;
	}

	:global(.dark .info-card) {
		background: rgba(99, 102, 241, 0.1);
		border-color: rgba(99, 102, 241, 0.2);
	}
	:global(.dark .info-card h4) { color: #818cf8; }
	:global(.dark .info-card li) { color: #aaa; }

	:global(.dark .tip-card) {
		background: rgba(16, 185, 129, 0.1);
		border-color: rgba(16, 185, 129, 0.2);
	}
	:global(.dark .tip-card div) { color: #aaa; }

	:global(.dark .keyboard-hint) {
		background: #333;
		color: #fff;
	}
	:global(.dark kbd) {
		background: #333;
		border-color: rgba(255, 255, 255, 0.15);
		color: #fff;
	}

	:global(.dark .customization-card) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .customization-card h4) { color: #fff; }
	:global(.dark .customization-card > p) { color: #aaa; }
	:global(.dark .customization-card li) { color: #888; }

	:global(.dark .module-detail) {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .module-header h3) { color: #fff; }
	:global(.dark .module-tagline) { color: #aaa; }
	:global(.dark .module-detail > p) { color: #aaa; }
	:global(.dark .module-features span) {
		background: rgba(99, 102, 241, 0.15);
		color: #818cf8;
	}
	:global(.dark .more-modules h3) { color: #fff; }

	:global(.dark .mini-module) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .mini-module h5) { color: #fff; }
	:global(.dark .mini-module p) { color: #888; }
	:global(.dark .mini-icon) {
		background: rgba(99, 102, 241, 0.15);
		color: #818cf8;
	}

	:global(.dark .node-type) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .node-type h4) { color: #fff; }
	:global(.dark .node-type p) { color: #888; }
	:global(.dark .node-icon.business) { background: rgba(59, 130, 246, 0.2); }
	:global(.dark .node-icon.project) { background: rgba(139, 92, 246, 0.2); }
	:global(.dark .node-icon.learning) { background: rgba(245, 158, 11, 0.2); }
	:global(.dark .node-icon.operational) { background: rgba(16, 185, 129, 0.2); }

	:global(.dark .pipeline-stages) {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .stage h5) { color: #fff; }
	:global(.dark .stage p) { color: #888; }
	:global(.dark .stage-dot) { background: #6366F1; }
	:global(.dark .stage-arrow) { color: #444; }

	:global(.dark .shortcut-category h3) { color: #fff; }
	:global(.dark .shortcut-list) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .shortcut-row) { border-color: rgba(255, 255, 255, 0.08); }
	:global(.dark .shortcut-row .desc) { color: #aaa; }

	:global(.dark .privacy-note) {
		background: rgba(16, 185, 129, 0.1);
		border-color: rgba(16, 185, 129, 0.2);
		color: #aaa;
	}

	:global(.dark .mock-search) {
		background: #333;
		border-color: rgba(255, 255, 255, 0.1);
		color: #888;
	}

	:global(.dark .commands-section h3) { color: #fff; }
	:global(.dark .commands-section > p) { color: #888; }
	:global(.dark .command) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
		color: #aaa;
	}
	:global(.dark .command code) {
		background: rgba(99, 102, 241, 0.2);
		color: #818cf8;
	}

	:global(.dark .setting-card) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .setting-card h4) { color: #fff; }
	:global(.dark .setting-card p) { color: #888; }
	:global(.dark .setting-card > svg) { color: #818cf8; }

	:global(.dark .feature-item) {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}
	:global(.dark .item-icon) { background: rgba(99, 102, 241, 0.15); }
	:global(.dark .item-content h4) { color: #fff; }
	:global(.dark .item-content p) { color: #888; }

	:global(.dark .card-icon.blue) {
		background: rgba(59, 130, 246, 0.2);
		color: #60a5fa;
	}
	:global(.dark .card-icon.purple) {
		background: rgba(139, 92, 246, 0.2);
		color: #a78bfa;
	}
	:global(.dark .card-icon.green) {
		background: rgba(16, 185, 129, 0.2);
		color: #34d399;
	}
	:global(.dark .card-icon.orange) {
		background: rgba(245, 158, 11, 0.2);
		color: #fbbf24;
	}
</style>
