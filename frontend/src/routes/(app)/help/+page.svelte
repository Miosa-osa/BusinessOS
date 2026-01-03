<script lang="ts">
	import { windowStore } from '$lib/stores/windowStore';
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
		Terminal,
		HelpCircle,
		ChevronRight,
		ChevronDown,
		ExternalLink,
		Mail,
		Github,
		Zap,
		Layers,
		MousePointer,
		Maximize2,
		Move,
		FolderPlus,
		Palette,
		Anchor,
		ArrowUpRight,
		Clock,
		Play,
		Activity,
		Command,
		LayoutGrid,
		TreePine,
		List,
		Share2,
		Tag,
		Link,
		Globe,
		Bot,
		Wand2,
		FileSearch,
		PieChart,
		TrendingUp,
		Filter,
		SortAsc,
		Plus,
		Eye,
		EyeOff,
		Lock,
		Unlock,
		Volume2,
		VolumeX,
		Download,
		Upload,
		RefreshCw,
		AlertCircle,
		Info,
		Lightbulb,
		BookMarked,
		GraduationCap,
		Compass,
		Map,
		Phone
	} from 'lucide-svelte';

	const isEmbedded = $derived($page.url.searchParams.get('embed') === 'true');

	// Expanded sections with better organization
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

	// Filter sections based on search
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

		<!-- Search -->
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
			<a href="https://github.com/robertohluna/BusinessOS" target="_blank" class="footer-link">
				<Github size={14} />
				<span>View on GitHub</span>
			</a>
		</div>
	</aside>

	<!-- Main Content -->
	<main class="help-content">
		<!-- Getting Started -->
		<section id="getting-started" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-blue">
					<Rocket size={24} />
				</div>
				<div>
					<h1>Getting Started with Business OS</h1>
					<p class="section-subtitle">Your all-in-one operating system for managing your business</p>
				</div>
			</div>

			<div class="intro-card">
				<p>
					Business OS is an AI-powered workspace that combines project management, team collaboration,
					client management, and intelligent assistance in a beautiful desktop-like interface. Think of it
					as macOS meets Notion meets your personal AI assistant.
				</p>
			</div>

			<div id="overview" class="subsection">
				<h2>Platform Overview</h2>
				<div class="feature-grid four-col">
					<div class="feature-card">
						<div class="card-icon blue">
							<Monitor size={20} />
						</div>
						<h3>Desktop Environment</h3>
						<p>A familiar macOS-inspired interface with draggable windows, a dock for quick access,
						customizable backgrounds, and even an experimental 3D spatial mode.</p>
					</div>
					<div class="feature-card">
						<div class="card-icon purple">
							<Bot size={20} />
						</div>
						<h3>AI-First Design</h3>
						<p>Every module is enhanced with AI. Chat with specialized Focus Modes, get intelligent
						suggestions, and let AI help you analyze, write, research, and build.</p>
					</div>
					<div class="feature-card">
						<div class="card-icon green">
							<Layers size={20} />
						</div>
						<h3>Integrated Modules</h3>
						<p>Tasks, Projects, Team, Clients, Calendar, Documents - all connected and aware of each
						other. Create a task from chat, link it to a project, assign to team.</p>
					</div>
					<div class="feature-card">
						<div class="card-icon orange">
							<Brain size={20} />
						</div>
						<h3>Knowledge System</h3>
						<p>Organize your business knowledge in Nodes and Contexts. The AI understands your
						business structure and provides contextual assistance.</p>
					</div>
				</div>
			</div>

			<div id="first-steps" class="subsection">
				<h2>First Steps</h2>
				<div class="steps-list">
					<div class="step-item">
						<div class="step-number">1</div>
						<div class="step-content">
							<h4>Explore the Desktop</h4>
							<p>Click icons on the desktop or dock to open modules. Drag windows to move them,
							resize by dragging edges. Try <kbd>Cmd</kbd> + <kbd>Space</kbd> for Spotlight search.</p>
						</div>
					</div>
					<div class="step-item">
						<div class="step-number">2</div>
						<div class="step-content">
							<h4>Try the AI Chat</h4>
							<p>Open the Chat module or use the quick chat bar above the dock. Select a Focus Mode
							(Research, Analyze, Write, Build) for specialized assistance.</p>
						</div>
					</div>
					<div class="step-item">
						<div class="step-number">3</div>
						<div class="step-content">
							<h4>Create a Node</h4>
							<p>Nodes are the backbone of your business structure. Create a node for your main
							business, then add child nodes for departments, projects, or areas of focus.</p>
						</div>
					</div>
					<div class="step-item">
						<div class="step-number">4</div>
						<div class="step-content">
							<h4>Add a Project & Tasks</h4>
							<p>Create a project, then add tasks with priorities and due dates. Use the Kanban
							board to visualize progress or switch to list/calendar views.</p>
						</div>
					</div>
					<div class="step-item">
						<div class="step-number">5</div>
						<div class="step-content">
							<h4>Customize Your Space</h4>
							<p>Right-click the desktop for settings. Choose from 50+ backgrounds, 15 icon styles,
							adjust sizes, and make Business OS feel like home.</p>
						</div>
					</div>
				</div>
			</div>

			<div id="core-concepts" class="subsection">
				<h2>Core Concepts</h2>
				<div class="concept-grid">
					<div class="concept-card">
						<h4><Network size={16} /> Nodes</h4>
						<p>Hierarchical units representing areas of your business. Each node has a health status,
						purpose, and can contain sub-nodes. Set an "active" node to focus AI context.</p>
					</div>
					<div class="concept-card">
						<h4><FileText size={16} /> Contexts</h4>
						<p>Rich documents with a Notion-like block editor. Store business knowledge, SOPs,
						meeting notes - anything the AI should know about when helping you.</p>
					</div>
					<div class="concept-card">
						<h4><Target size={16} /> Focus Modes</h4>
						<p>Specialized AI agents for different tasks: Research (find info), Analyze (deep dive),
						Write (create content), Build (plan & strategy).</p>
					</div>
					<div class="concept-card">
						<h4><FolderKanban size={16} /> Projects & Tasks</h4>
						<p>Organize work into projects with status tracking. Tasks have priorities, due dates,
						assignees, and can be viewed as Kanban boards, lists, or calendar.</p>
					</div>
				</div>
			</div>
		</section>

		<!-- AI Features -->
		<section id="ai-features" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-purple">
					<Sparkles size={24} />
				</div>
				<div>
					<h1>AI Features</h1>
					<p class="section-subtitle">Intelligent assistance tailored to your business</p>
				</div>
			</div>

			<div id="focus-modes" class="subsection">
				<h2>Focus Modes</h2>
				<p class="subsection-intro">
					Focus Modes transform the AI into specialized agents optimized for different types of work.
					Each mode has unique prompts, capabilities, and configurable options.
				</p>

				<div class="mode-grid">
					<div class="mode-card research">
						<div class="mode-header">
							<FileSearch size={24} />
							<h3>Research Mode</h3>
						</div>
						<p>Search the web and your documents for information. Perfect for market research,
						competitive analysis, or finding specific facts.</p>
						<div class="mode-options">
							<h5>Options:</h5>
							<ul>
								<li><strong>Scope:</strong> Web Only, Documents Only, or All Sources</li>
								<li><strong>Depth:</strong> Quick Scan or Thorough Investigation</li>
								<li><strong>Output:</strong> Summary, Bullet Points, or Full Report</li>
							</ul>
						</div>
						<div class="mode-example">
							<span class="example-label">Example:</span>
							"Research the top 5 competitors in the CRM space and summarize their pricing models"
						</div>
					</div>

					<div class="mode-card analyze">
						<div class="mode-header">
							<BarChart3 size={24} />
							<h3>Analyze Mode</h3>
						</div>
						<p>Deep analysis of data, situations, or documents. Get insights, identify patterns,
						and make data-driven decisions.</p>
						<div class="mode-options">
							<h5>Options:</h5>
							<ul>
								<li><strong>Approach:</strong> Validate, Compare, or Forecast</li>
								<li><strong>Depth:</strong> Surface Level to Deep Dive</li>
								<li><strong>Output:</strong> Key Findings or Interactive Dashboard</li>
							</ul>
						</div>
						<div class="mode-example">
							<span class="example-label">Example:</span>
							"Analyze our Q4 sales data and identify which product categories are underperforming"
						</div>
					</div>

					<div class="mode-card write">
						<div class="mode-header">
							<PenTool size={24} />
							<h3>Write Mode</h3>
						</div>
						<p>Create professional content - documents, emails, proposals, blog posts,
						presentations, and more.</p>
						<div class="mode-options">
							<h5>Options:</h5>
							<ul>
								<li><strong>Format:</strong> Document, Email, Slides, or Spreadsheet</li>
								<li><strong>Tone:</strong> Professional, Casual, or Persuasive</li>
								<li><strong>Process:</strong> Step-by-Step or First Draft</li>
							</ul>
						</div>
						<div class="mode-example">
							<span class="example-label">Example:</span>
							"Write a professional proposal for our new consulting service targeting SMBs"
						</div>
					</div>

					<div class="mode-card build">
						<div class="mode-header">
							<Hammer size={24} />
							<h3>Build Mode</h3>
						</div>
						<p>Strategy, planning, and framework creation. Build SOPs, project plans,
						business frameworks, and operational procedures.</p>
						<div class="mode-options">
							<h5>Options:</h5>
							<ul>
								<li><strong>Type:</strong> Framework, SOP, or Plan</li>
								<li><strong>Detail:</strong> High-Level Outline or Detailed Steps</li>
								<li><strong>Format:</strong> Document, Checklist, or Diagram</li>
							</ul>
						</div>
						<div class="mode-example">
							<span class="example-label">Example:</span>
							"Build an SOP for our customer onboarding process with detailed checklists"
						</div>
					</div>
				</div>

				<div class="tip-card">
					<Lightbulb size={18} />
					<div>
						<strong>Pro Tip:</strong> Use "Do More" mode for general conversation without a specific focus.
						The AI adapts to your needs and can switch modes mid-conversation if needed.
					</div>
				</div>
			</div>

			<div id="ai-chat" class="subsection">
				<h2>AI Chat Interface</h2>
				<div class="feature-list">
					<div class="feature-item">
						<div class="item-icon"><MessageSquare size={18} /></div>
						<div class="item-content">
							<h4>Full Chat Module</h4>
							<p>Open the Chat module for extended conversations. Your chat history is saved,
							and you can create multiple conversation threads for different topics.</p>
						</div>
					</div>
					<div class="feature-item">
						<div class="item-icon"><Zap size={18} /></div>
						<div class="item-content">
							<h4>Quick Chat Bar</h4>
							<p>The chat bar above your dock provides instant access. Ask quick questions
							without leaving your current work. Responses appear inline.</p>
						</div>
					</div>
					<div class="feature-item">
						<div class="item-icon"><Mic size={18} /></div>
						<div class="item-content">
							<h4>Voice Input</h4>
							<p>Click the microphone to speak your message. Audio is transcribed in real-time
							and you can edit before sending.</p>
						</div>
					</div>
					<div class="feature-item">
						<div class="item-icon"><Wand2 size={18} /></div>
						<div class="item-content">
							<h4>AI Artifacts</h4>
							<p>When AI generates documents, code, or structured content, it appears as an
							"artifact" you can save directly to your Contexts.</p>
						</div>
					</div>
				</div>
			</div>

			<div id="ai-contexts" class="subsection">
				<h2>AI Context Awareness</h2>
				<p class="subsection-intro">
					The AI in Business OS isn't just a generic chatbot - it understands your business structure
					and can access your stored knowledge.
				</p>
				<div class="info-card">
					<h4>How Context Works</h4>
					<ul>
						<li><strong>Active Node:</strong> Set a node as "active" and AI conversations will be
						focused on that area of your business</li>
						<li><strong>Document Context:</strong> Reference specific Contexts in chat and AI will
						read and understand them</li>
						<li><strong>Project Awareness:</strong> AI knows about your projects, tasks, and team
						when providing suggestions</li>
						<li><strong>Conversation Memory:</strong> Within a chat session, AI remembers what
						you've discussed and builds on previous responses</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Desktop Environment -->
		<section id="desktop" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-green">
					<Monitor size={24} />
				</div>
				<div>
					<h1>Desktop Environment</h1>
					<p class="section-subtitle">A familiar, powerful workspace</p>
				</div>
			</div>

			<div id="windows" class="subsection">
				<h2>Window Management</h2>
				<div class="feature-grid three-col">
					<div class="feature-card compact">
						<Move size={20} class="card-icon-inline" />
						<h4>Move Windows</h4>
						<p>Drag the title bar to move windows anywhere on your desktop.</p>
					</div>
					<div class="feature-card compact">
						<Maximize2 size={20} class="card-icon-inline" />
						<h4>Resize Windows</h4>
						<p>Drag edges or corners to resize. Double-click title bar to maximize.</p>
					</div>
					<div class="feature-card compact">
						<Layers size={20} class="card-icon-inline" />
						<h4>Window Snapping</h4>
						<p>Drag to screen edges to snap windows to half or quarter positions.</p>
					</div>
				</div>

				<div class="keyboard-hint">
					<Keyboard size={16} />
					<span><kbd>Cmd</kbd> + <kbd>W</kbd> closes the current window,
					<kbd>Cmd</kbd> + <kbd>`</kbd> cycles between open windows</span>
				</div>
			</div>

			<div id="dock" class="subsection">
				<h2>Dock & Navigation</h2>
				<div class="info-card">
					<h4>Using the Dock</h4>
					<ul>
						<li><strong>Open Apps:</strong> Click any icon to open or focus that module</li>
						<li><strong>Pin to Dock:</strong> Right-click desktop icons → "Add to Dock"</li>
						<li><strong>Running Indicator:</strong> A dot appears under running applications</li>
						<li><strong>Quick Chat:</strong> The chat bar above the dock for instant AI access</li>
						<li><strong>Tooltips:</strong> Hover over icons to see module names</li>
					</ul>
				</div>
			</div>

			<div id="customization" class="subsection">
				<h2>Desktop Customization</h2>
				<p class="subsection-intro">
					Right-click anywhere on the desktop and select "Desktop Settings" to personalize your workspace.
				</p>

				<div class="feature-grid two-col">
					<div class="customization-card">
						<Palette size={20} />
						<h4>Backgrounds</h4>
						<p>Choose from 50+ options:</p>
						<ul>
							<li><strong>Solid Colors:</strong> Clean, minimal backgrounds</li>
							<li><strong>Gradients:</strong> Sunrise, Ocean, Aurora, Sunset, Forest</li>
							<li><strong>Patterns:</strong> Dots, Grid, Lines, Blueprint, Topography</li>
							<li><strong>Custom:</strong> Upload your own images</li>
						</ul>
					</div>
					<div class="customization-card">
						<Grid3X3 size={20} />
						<h4>Icon Styles</h4>
						<p>15 unique icon themes:</p>
						<ul>
							<li><strong>Modern:</strong> macOS, Glassmorphism, Minimal</li>
							<li><strong>Retro:</strong> Windows 95, Pixel Art, Classic</li>
							<li><strong>Creative:</strong> Neon, Gradient, Outlined</li>
							<li><strong>Size:</strong> Adjustable from 32px to 128px</li>
						</ul>
					</div>
				</div>

				<div class="tip-card">
					<Lightbulb size={18} />
					<div>
						<strong>Pro Tip:</strong> Enable the "Noise Texture" option for a subtle film grain
						effect that adds visual depth to gradient backgrounds.
					</div>
				</div>
			</div>

			<div id="3d-mode" class="subsection">
				<h2>3D Desktop Mode (Experimental)</h2>
				<p class="subsection-intro">
					Experience your workspace in a new dimension with the experimental 3D desktop mode.
				</p>

				<div class="info-card highlight">
					<h4>3D Mode Features</h4>
					<ul>
						<li><strong>Spatial Layout:</strong> Windows orbit around a central sphere in 3D space</li>
						<li><strong>Focus View:</strong> Click any window to bring it front and center</li>
						<li><strong>Orbit Controls:</strong> Click and drag to rotate the view</li>
						<li><strong>Navigation:</strong> Use arrow keys to move between windows when focused</li>
						<li><strong>Resize:</strong> Use +/- buttons or keyboard shortcuts to resize focused windows</li>
					</ul>
					<div class="keyboard-hint" style="margin-top: 12px;">
						<Keyboard size={16} />
						<span><kbd>Esc</kbd> to unfocus/exit, <kbd>Space</kbd> toggles view mode,
						<kbd>+</kbd>/<kbd>-</kbd> resize</span>
					</div>
				</div>
			</div>
		</section>

		<!-- Core Modules -->
		<section id="modules" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-orange">
					<Grid3X3 size={24} />
				</div>
				<div>
					<h1>Core Modules</h1>
					<p class="section-subtitle">Everything you need to run your business</p>
				</div>
			</div>

			<div class="modules-showcase">
				<div id="dashboard-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #E3F2FD; color: #1E88E5;">
							<PieChart size={24} />
						</div>
						<div>
							<h3>Dashboard</h3>
							<span class="module-tagline">Your daily command center</span>
						</div>
					</div>
					<p>Get a complete overview of your business at a glance. See today's tasks, upcoming
					deadlines, project progress, and key metrics. Quick actions let you create tasks,
					start a chat, or jump to any module.</p>
					<div class="module-features">
						<span>Today's Focus</span>
						<span>Task Summary</span>
						<span>Project Status</span>
						<span>Quick Actions</span>
					</div>
				</div>

				<div id="chat-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #E8F5E9; color: #43A047;">
							<MessageSquare size={24} />
						</div>
						<div>
							<h3>Chat</h3>
							<span class="module-tagline">AI-powered conversations</span>
						</div>
					</div>
					<p>Your intelligent assistant for any task. Use Focus Modes for specialized help,
					save conversation threads, and let AI access your business context for relevant answers.</p>
					<div class="module-features">
						<span>Focus Modes</span>
						<span>Voice Input</span>
						<span>Artifacts</span>
						<span>Context Awareness</span>
					</div>
				</div>

				<div id="tasks-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #FFF3E0; color: #FB8C00;">
							<CheckSquare size={24} />
						</div>
						<div>
							<h3>Tasks</h3>
							<span class="module-tagline">Get things done</span>
						</div>
					</div>
					<p>Powerful task management with multiple views. Kanban boards for visual workflow,
					list view for quick scanning, calendar view for deadline management. Set priorities,
					assignees, and due dates.</p>
					<div class="module-features">
						<span>Kanban Board</span>
						<span>List View</span>
						<span>Calendar View</span>
						<span>Priorities</span>
						<span>Filters</span>
					</div>
				</div>

				<div id="projects-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #F3E5F5; color: #8E24AA;">
							<FolderKanban size={24} />
						</div>
						<div>
							<h3>Projects</h3>
							<span class="module-tagline">Organize your work</span>
						</div>
					</div>
					<p>Group related tasks and track progress. Each project has a status, description,
					and linked tasks. See completion percentages and manage project-specific notes.</p>
					<div class="module-features">
						<span>Project Status</span>
						<span>Task Linking</span>
						<span>Progress Tracking</span>
						<span>Notes</span>
					</div>
				</div>

				<div id="team-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #E0F7FA; color: #00ACC1;">
							<Users size={24} />
						</div>
						<div>
							<h3>Team</h3>
							<span class="module-tagline">Manage your people</span>
						</div>
					</div>
					<p>Visualize your organization with an interactive org chart. Track team member
					capacity, assign tasks, and see workload distribution across your team.</p>
					<div class="module-features">
						<span>Org Chart</span>
						<span>Capacity Planning</span>
						<span>Workload View</span>
						<span>Team Directory</span>
					</div>
				</div>

				<div id="calendar-mod" class="module-detail">
					<div class="module-header">
						<div class="module-icon" style="background: #FCE4EC; color: #E91E63;">
							<Calendar size={24} />
						</div>
						<div>
							<h3>Calendar</h3>
							<span class="module-tagline">Time management</span>
						</div>
					</div>
					<p>Schedule events, meetings, and deadlines. Integrates with Google Calendar for
					sync. Multiple views including day, week, and month layouts.</p>
					<div class="module-features">
						<span>Event Creation</span>
						<span>Google Sync</span>
						<span>Multiple Views</span>
						<span>Reminders</span>
					</div>
				</div>
			</div>

			<div class="more-modules">
				<h3>Additional Modules</h3>
				<div class="mini-modules-grid">
					<div class="mini-module">
						<div class="mini-icon" style="background: #263238; color: #4CAF50;">
							<Terminal size={18} />
						</div>
						<div>
							<h5>Terminal</h5>
							<p>Command-line interface for power users and OS Agent interaction</p>
						</div>
					</div>
					<div class="mini-module">
						<div class="mini-icon" style="background: #EFEBE9; color: #795548;">
							<Activity size={18} />
						</div>
						<div>
							<h5>Daily Log</h5>
							<p>Track daily activities, patterns, and reflections over time</p>
						</div>
					</div>
					<div class="mini-module">
						<div class="mini-icon" style="background: #E0F2F1; color: #00897B;">
							<TrendingUp size={18} />
						</div>
						<div>
							<h5>Usage</h5>
							<p>Analytics on AI usage, tokens consumed, and costs by provider</p>
						</div>
					</div>
					<div class="mini-module">
						<div class="mini-icon" style="background: #F5F5F5; color: #616161;">
							<Settings size={18} />
						</div>
						<div>
							<h5>Settings</h5>
							<p>Configure AI models, API keys, integrations, and preferences</p>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Knowledge & Data -->
		<section id="knowledge" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-teal">
					<Brain size={24} />
				</div>
				<div>
					<h1>Knowledge & Data</h1>
					<p class="section-subtitle">Organize and leverage your business intelligence</p>
				</div>
			</div>

			<div id="nodes" class="subsection">
				<h2>Nodes System</h2>
				<p class="subsection-intro">
					Nodes are the cognitive backbone of Business OS - a hierarchical structure for organizing
					and managing different areas of your business.
				</p>

				<div class="node-types">
					<div class="node-type">
						<div class="node-icon business"><Briefcase size={20} /></div>
						<h4>Business Nodes</h4>
						<p>Top-level units like departments, divisions, or major business areas</p>
					</div>
					<div class="node-type">
						<div class="node-icon project"><FolderKanban size={20} /></div>
						<h4>Project Nodes</h4>
						<p>Specific initiatives, products, or campaigns within your business</p>
					</div>
					<div class="node-type">
						<div class="node-icon learning"><GraduationCap size={20} /></div>
						<h4>Learning Nodes</h4>
						<p>Knowledge areas, training programs, skill development tracks</p>
					</div>
					<div class="node-type">
						<div class="node-icon operational"><Settings size={20} /></div>
						<h4>Operational Nodes</h4>
						<p>Day-to-day processes, SOPs, and recurring operations</p>
					</div>
				</div>

				<div class="info-card">
					<h4>Node Properties</h4>
					<ul>
						<li><strong>Health Status:</strong> Healthy, Needs Attention, Critical, or Not Started</li>
						<li><strong>Purpose:</strong> What this node represents and its goals</li>
						<li><strong>Current Status:</strong> What's happening right now</li>
						<li><strong>Weekly Focus:</strong> This week's priorities and targets</li>
						<li><strong>Decision Queue:</strong> Pending decisions requiring attention</li>
						<li><strong>Delegation Ready:</strong> Tasks ready to hand off</li>
					</ul>
				</div>

				<div class="tip-card">
					<Target size={18} />
					<div>
						<strong>Active Node:</strong> Set a node as "active" to focus AI conversations on that
						area. The AI will use the node's context when providing assistance.
					</div>
				</div>
			</div>

			<div id="contexts" class="subsection">
				<h2>Contexts (Documents)</h2>
				<p class="subsection-intro">
					Rich documents with a Notion-like block editor. Store business knowledge that AI can
					reference in conversations.
				</p>

				<div class="feature-grid two-col">
					<div class="info-card">
						<h4>Block Types</h4>
						<ul>
							<li><strong>Text:</strong> Paragraphs, headings (H1, H2, H3)</li>
							<li><strong>Lists:</strong> Bullet, numbered, and to-do checkboxes</li>
							<li><strong>Media:</strong> Images, embeds, code blocks with syntax highlighting</li>
							<li><strong>Structure:</strong> Quotes, callouts, dividers, tables</li>
							<li><strong>AI:</strong> Artifact blocks for AI-generated content</li>
						</ul>
					</div>
					<div class="info-card">
						<h4>Document Features</h4>
						<ul>
							<li><strong>Slash Commands:</strong> Type <code>/</code> to insert any block type</li>
							<li><strong>Properties:</strong> Add custom fields (text, select, date, etc.)</li>
							<li><strong>Relations:</strong> Link to other contexts, projects, or clients</li>
							<li><strong>Sharing:</strong> Generate public links to share documents</li>
							<li><strong>Templates:</strong> Create and use document templates</li>
						</ul>
					</div>
				</div>
			</div>

			<div id="knowledge-graph" class="subsection">
				<h2>Knowledge Graph</h2>
				<p class="subsection-intro">
					Visualize connections between your business entities in an interactive 3D graph.
				</p>
				<div class="info-card">
					<h4>Graph Features</h4>
					<ul>
						<li>See relationships between nodes, contexts, projects, and clients</li>
						<li>Interactive 3D visualization with orbit controls</li>
						<li>Click nodes to view details and navigate</li>
						<li>Filter by entity type or connection strength</li>
						<li>Discover hidden relationships in your business data</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Clients & CRM -->
		<section id="clients" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-indigo">
					<Briefcase size={24} />
				</div>
				<div>
					<h1>Clients & CRM</h1>
					<p class="section-subtitle">Manage relationships and grow your business</p>
				</div>
			</div>

			<div id="client-profiles" class="subsection">
				<h2>Client Profiles</h2>
				<div class="feature-grid two-col">
					<div class="info-card">
						<h4>Profile Information</h4>
						<ul>
							<li>Company name and logo</li>
							<li>Industry and company size</li>
							<li>Contact information</li>
							<li>Custom fields and tags</li>
							<li>Notes and internal comments</li>
						</ul>
					</div>
					<div class="info-card">
						<h4>Status Tracking</h4>
						<ul>
							<li><strong>Lead:</strong> Initial contact, not yet qualified</li>
							<li><strong>Prospect:</strong> Qualified, actively pursuing</li>
							<li><strong>Active:</strong> Current paying client</li>
							<li><strong>Inactive:</strong> Paused or dormant relationship</li>
							<li><strong>Churned:</strong> Former client</li>
						</ul>
					</div>
				</div>
			</div>

			<div id="deals-pipeline" class="subsection">
				<h2>Deals Pipeline</h2>
				<p class="subsection-intro">
					Track sales opportunities through your pipeline with a visual Kanban board.
				</p>
				<div class="pipeline-stages">
					<div class="stage">
						<div class="stage-dot" style="background: #64B5F6;"></div>
						<h5>Qualification</h5>
						<p>Initial assessment of fit and potential</p>
					</div>
					<div class="stage-arrow">→</div>
					<div class="stage">
						<div class="stage-dot" style="background: #FFB74D;"></div>
						<h5>Proposal</h5>
						<p>Preparing and presenting your offer</p>
					</div>
					<div class="stage-arrow">→</div>
					<div class="stage">
						<div class="stage-dot" style="background: #BA68C8;"></div>
						<h5>Negotiation</h5>
						<p>Discussing terms and finalizing details</p>
					</div>
					<div class="stage-arrow">→</div>
					<div class="stage">
						<div class="stage-dot" style="background: #81C784;"></div>
						<h5>Closed</h5>
						<p>Deal won (or lost) - final outcome</p>
					</div>
				</div>
				<div class="info-card" style="margin-top: 20px;">
					<h4>Deal Properties</h4>
					<ul>
						<li><strong>Value:</strong> Expected deal amount</li>
						<li><strong>Probability:</strong> Likelihood of closing (affects forecasting)</li>
						<li><strong>Expected Close:</strong> Target closing date</li>
						<li><strong>Owner:</strong> Team member responsible</li>
					</ul>
				</div>
			</div>

			<div id="interactions" class="subsection">
				<h2>Interaction Tracking</h2>
				<div class="feature-grid three-col">
					<div class="feature-card compact">
						<Phone size={20} class="card-icon-inline" />
						<h4>Calls</h4>
						<p>Log phone calls with notes and outcomes</p>
					</div>
					<div class="feature-card compact">
						<Mail size={20} class="card-icon-inline" />
						<h4>Emails</h4>
						<p>Track email correspondence</p>
					</div>
					<div class="feature-card compact">
						<Users size={20} class="card-icon-inline" />
						<h4>Meetings</h4>
						<p>Record meeting notes and action items</p>
					</div>
				</div>
			</div>
		</section>

		<!-- Keyboard Shortcuts -->
		<section id="shortcuts" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-gray">
					<Keyboard size={24} />
				</div>
				<div>
					<h1>Keyboard Shortcuts</h1>
					<p class="section-subtitle">Work faster with these keyboard combinations</p>
				</div>
			</div>

			<div class="shortcuts-container">
				<div class="shortcut-category">
					<h3>Global</h3>
					<div class="shortcut-list">
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>Space</kbd></div>
							<div class="desc">Open Spotlight Search</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>Shift</kbd> + <kbd>Space</kbd></div>
							<div class="desc">Open Quick Chat popup</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>W</kbd></div>
							<div class="desc">Close current window</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>M</kbd></div>
							<div class="desc">Minimize current window</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>`</kbd></div>
							<div class="desc">Cycle between windows</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Esc</kbd></div>
							<div class="desc">Close modals / Deselect</div>
						</div>
					</div>
				</div>

				<div class="shortcut-category">
					<h3>Chat</h3>
					<div class="shortcut-list">
						<div class="shortcut-row">
							<div class="keys"><kbd>Enter</kbd></div>
							<div class="desc">Send message</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Shift</kbd> + <kbd>Enter</kbd></div>
							<div class="desc">New line in message</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Cmd</kbd> + <kbd>Enter</kbd></div>
							<div class="desc">Expand to full chat</div>
						</div>
					</div>
				</div>

				<div class="shortcut-category">
					<h3>Spotlight</h3>
					<div class="shortcut-list">
						<div class="shortcut-row">
							<div class="keys"><kbd>Tab</kbd></div>
							<div class="desc">Switch Search/Chat mode</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>↑</kbd> / <kbd>↓</kbd></div>
							<div class="desc">Navigate results</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Enter</kbd></div>
							<div class="desc">Select / Send</div>
						</div>
					</div>
				</div>

				<div class="shortcut-category">
					<h3>3D Desktop</h3>
					<div class="shortcut-list">
						<div class="shortcut-row">
							<div class="keys"><kbd>Esc</kbd></div>
							<div class="desc">Unfocus / Exit 3D mode</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>Space</kbd></div>
							<div class="desc">Toggle orb/grid view</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>←</kbd> / <kbd>→</kbd></div>
							<div class="desc">Navigate windows</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>+</kbd> / <kbd>-</kbd></div>
							<div class="desc">Resize focused window</div>
						</div>
						<div class="shortcut-row">
							<div class="keys"><kbd>1</kbd> - <kbd>9</kbd></div>
							<div class="desc">Focus window by index</div>
						</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Voice Features -->
		<section id="voice" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-red">
					<Mic size={24} />
				</div>
				<div>
					<h1>Voice Features</h1>
					<p class="section-subtitle">Hands-free interaction with Business OS</p>
				</div>
			</div>

			<div class="feature-grid two-col">
				<div class="info-card">
					<h4>How to Use Voice Input</h4>
					<ol>
						<li>Click the microphone icon in chat or Spotlight</li>
						<li>Speak your message clearly</li>
						<li>A waveform visualizes your audio</li>
						<li>Click stop when finished</li>
						<li>Review and edit the transcription</li>
						<li>Press Enter to send</li>
					</ol>
				</div>
				<div class="info-card">
					<h4>Tips for Best Results</h4>
					<ul>
						<li>Use a quiet environment when possible</li>
						<li>Speak at a natural pace</li>
						<li>Punctuation is added automatically</li>
						<li>You can always edit before sending</li>
						<li>Works in Quick Chat and full Chat module</li>
					</ul>
				</div>
			</div>

			<div class="privacy-note">
				<Lock size={16} />
				<p>
					<strong>Privacy:</strong> Audio is processed securely and not stored permanently.
					Transcription happens on secure servers with enterprise-grade encryption.
				</p>
			</div>
		</section>

		<!-- Spotlight Search -->
		<section id="search" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-yellow">
					<Search size={24} />
				</div>
				<div>
					<h1>Spotlight Search</h1>
					<p class="section-subtitle">Your command center for everything</p>
				</div>
			</div>

			<div class="spotlight-preview">
				<div class="spotlight-mock">
					<div class="mock-search">
						<Search size={18} />
						<span>Search or chat...</span>
					</div>
				</div>
			</div>

			<div class="feature-grid two-col">
				<div class="info-card">
					<h4>Search Mode</h4>
					<ul>
						<li>Find apps, projects, tasks, clients, documents</li>
						<li>Results grouped by category</li>
						<li>Use arrow keys to navigate</li>
						<li>Press Enter to open selected item</li>
						<li>Recent searches shown by default</li>
					</ul>
				</div>
				<div class="info-card">
					<h4>Chat Mode</h4>
					<ul>
						<li>Press Tab to switch to Chat</li>
						<li>Select a context for relevant answers</li>
						<li>Voice input available</li>
						<li>Responses stream in real-time</li>
						<li>Quick access to AI assistance</li>
					</ul>
				</div>
			</div>

			<div class="commands-section">
				<h3>Quick Commands</h3>
				<p>Type these in Spotlight for instant actions:</p>
				<div class="commands-grid">
					<div class="command"><code>/task</code> Create new task</div>
					<div class="command"><code>/note</code> Quick note</div>
					<div class="command"><code>/project</code> New project</div>
					<div class="command"><code>/meet</code> Schedule meeting</div>
					<div class="command"><code>/remind</code> Set reminder</div>
					<div class="command"><code>/client</code> New client</div>
				</div>
			</div>
		</section>

		<!-- Settings -->
		<section id="settings-help" class="help-section">
			<div class="section-header">
				<div class="section-icon gradient-slate">
					<Settings size={24} />
				</div>
				<div>
					<h1>Settings & Configuration</h1>
					<p class="section-subtitle">Customize Business OS to your needs</p>
				</div>
			</div>

			<div class="settings-grid">
				<div class="setting-card">
					<Bot size={20} />
					<h4>AI Configuration</h4>
					<p>Choose your AI provider (Anthropic, OpenAI, etc.), select models, and configure
					API keys for each service.</p>
				</div>
				<div class="setting-card">
					<Calendar size={20} />
					<h4>Integrations</h4>
					<p>Connect Google Calendar, set up webhooks, and configure third-party integrations.</p>
				</div>
				<div class="setting-card">
					<Palette size={20} />
					<h4>Appearance</h4>
					<p>Desktop backgrounds, icon styles, window behavior, and visual preferences.</p>
				</div>
				<div class="setting-card">
					<Users size={20} />
					<h4>Team Settings</h4>
					<p>Manage team members, roles, permissions, and workspace settings.</p>
				</div>
			</div>
		</section>

		<!-- Footer -->
		<footer class="help-footer">
			<div class="footer-content">
				<h3>Need More Help?</h3>
				<p>Can't find what you're looking for? Reach out to our support team.</p>
				<div class="footer-actions">
					<a href="mailto:roberto@osa.dev" class="footer-btn primary">
						<Mail size={16} />
						Contact Support
					</a>
					<a href="https://github.com/robertohluna/BusinessOS" target="_blank" class="footer-btn">
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

	.help-section {
		margin-bottom: 80px;
	}

	.section-header {
		display: flex;
		align-items: flex-start;
		gap: 20px;
		margin-bottom: 32px;
	}

	.section-icon {
		width: 56px;
		height: 56px;
		border-radius: 14px;
		display: flex;
		align-items: center;
		justify-content: center;
		color: white;
		flex-shrink: 0;
	}

	.gradient-blue { background: linear-gradient(135deg, #3B82F6 0%, #1D4ED8 100%); }
	.gradient-purple { background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%); }
	.gradient-green { background: linear-gradient(135deg, #10B981 0%, #059669 100%); }
	.gradient-orange { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
	.gradient-teal { background: linear-gradient(135deg, #14B8A6 0%, #0D9488 100%); }
	.gradient-indigo { background: linear-gradient(135deg, #6366F1 0%, #4F46E5 100%); }
	.gradient-gray { background: linear-gradient(135deg, #64748B 0%, #475569 100%); }
	.gradient-red { background: linear-gradient(135deg, #EF4444 0%, #DC2626 100%); }
	.gradient-yellow { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }
	.gradient-slate { background: linear-gradient(135deg, #475569 0%, #334155 100%); }

	.section-header h1 {
		font-size: 28px;
		font-weight: 700;
		color: #0F172A;
		margin: 0 0 4px;
	}

	.section-subtitle {
		font-size: 15px;
		color: #64748B;
		margin: 0;
	}

	.subsection {
		margin-bottom: 48px;
	}

	.subsection h2 {
		font-size: 20px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 16px;
		padding-bottom: 12px;
		border-bottom: 1px solid #E2E8F0;
	}

	.subsection-intro {
		font-size: 14px;
		color: #64748B;
		line-height: 1.7;
		margin-bottom: 24px;
	}

	/* ===== CARDS & GRIDS ===== */
	.intro-card {
		background: linear-gradient(135deg, #EEF2FF 0%, #E0E7FF 100%);
		border: 1px solid #C7D2FE;
		border-radius: 12px;
		padding: 24px;
		margin-bottom: 40px;
	}

	.intro-card p {
		font-size: 15px;
		color: #3730A3;
		line-height: 1.7;
		margin: 0;
	}

	.feature-grid {
		display: grid;
		gap: 20px;
	}

	.feature-grid.four-col { grid-template-columns: repeat(4, 1fr); }
	.feature-grid.three-col { grid-template-columns: repeat(3, 1fr); }
	.feature-grid.two-col { grid-template-columns: repeat(2, 1fr); }

	.feature-card {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
		transition: all 0.2s;
	}

	.feature-card:hover {
		border-color: #CBD5E1;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
	}

	.feature-card.compact {
		padding: 20px;
		text-align: center;
	}

	.card-icon {
		width: 44px;
		height: 44px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		margin-bottom: 16px;
	}

	.card-icon.blue { background: #DBEAFE; color: #2563EB; }
	.card-icon.purple { background: #EDE9FE; color: #7C3AED; }
	.card-icon.green { background: #D1FAE5; color: #059669; }
	.card-icon.orange { background: #FED7AA; color: #EA580C; }

	:global(.card-icon-inline) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	.feature-card h3, .feature-card h4 {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.feature-card p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== STEPS ===== */
	.steps-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.step-item {
		display: flex;
		gap: 16px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 20px;
	}

	.step-number {
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

	.step-content h4 {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	.step-content p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== CONCEPT GRID ===== */
	.concept-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 16px;
	}

	.concept-card {
		background: #F8FAFC;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 20px;
	}

	.concept-card h4 {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.concept-card p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== MODE CARDS ===== */
	.mode-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
	}

	.mode-card {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
		border-top: 3px solid;
	}

	.mode-card.research { border-top-color: #3B82F6; }
	.mode-card.analyze { border-top-color: #8B5CF6; }
	.mode-card.write { border-top-color: #10B981; }
	.mode-card.build { border-top-color: #F59E0B; }

	.mode-header {
		display: flex;
		align-items: center;
		gap: 12px;
		margin-bottom: 12px;
	}

	.mode-card.research .mode-header { color: #3B82F6; }
	.mode-card.analyze .mode-header { color: #8B5CF6; }
	.mode-card.write .mode-header { color: #10B981; }
	.mode-card.build .mode-header { color: #F59E0B; }

	.mode-header h3 {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0;
	}

	.mode-card > p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0 0 16px;
	}

	.mode-options {
		background: #F8FAFC;
		border-radius: 8px;
		padding: 12px 16px;
		margin-bottom: 12px;
	}

	.mode-options h5 {
		font-size: 11px;
		font-weight: 600;
		color: #94A3B8;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin: 0 0 8px;
	}

	.mode-options ul {
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.mode-options li {
		font-size: 12px;
		color: #475569;
		line-height: 1.8;
	}

	.mode-example {
		background: #FEF3C7;
		border-radius: 8px;
		padding: 12px;
		font-size: 12px;
		color: #92400E;
		font-style: italic;
	}

	.example-label {
		font-weight: 600;
		font-style: normal;
	}

	/* ===== INFO CARDS ===== */
	.info-card {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	.info-card.highlight {
		background: linear-gradient(135deg, #F0FDF4 0%, #DCFCE7 100%);
		border-color: #86EFAC;
	}

	.info-card h4 {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 12px;
	}

	.info-card ul, .info-card ol {
		margin: 0;
		padding-left: 20px;
	}

	.info-card li {
		font-size: 13px;
		color: #475569;
		line-height: 1.8;
	}

	.info-card code {
		background: #F1F5F9;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		color: #6366F1;
	}

	/* ===== TIP CARD ===== */
	.tip-card {
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

	.tip-card div {
		font-size: 13px;
		line-height: 1.6;
	}

	/* ===== KEYBOARD HINTS ===== */
	.keyboard-hint {
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

	kbd {
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
	.customization-card {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	.customization-card > :global(svg) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	.customization-card h4 {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.customization-card > p {
		font-size: 13px;
		color: #64748B;
		margin: 0 0 12px;
	}

	.customization-card ul {
		margin: 0;
		padding-left: 18px;
	}

	.customization-card li {
		font-size: 12px;
		color: #475569;
		line-height: 1.8;
	}

	/* ===== MODULES SHOWCASE ===== */
	.modules-showcase {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
		margin-bottom: 40px;
	}

	.module-detail {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	.module-header {
		display: flex;
		align-items: center;
		gap: 14px;
		margin-bottom: 14px;
	}

	.module-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.module-header h3 {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0;
	}

	.module-tagline {
		font-size: 12px;
		color: #94A3B8;
	}

	.module-detail > p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0 0 16px;
	}

	.module-features {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.module-features span {
		padding: 4px 10px;
		background: #F1F5F9;
		border-radius: 100px;
		font-size: 11px;
		font-weight: 500;
		color: #475569;
	}

	.more-modules h3 {
		font-size: 16px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 16px;
	}

	.mini-modules-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 12px;
	}

	.mini-module {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		padding: 16px;
	}

	.mini-icon {
		width: 36px;
		height: 36px;
		border-radius: 8px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.mini-module h5 {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 2px;
	}

	.mini-module p {
		font-size: 12px;
		color: #64748B;
		line-height: 1.5;
		margin: 0;
	}

	/* ===== NODES ===== */
	.node-types {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 16px;
		margin-bottom: 24px;
	}

	.node-type {
		text-align: center;
		padding: 20px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
	}

	.node-icon {
		width: 48px;
		height: 48px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto 12px;
	}

	.node-icon.business { background: #DBEAFE; color: #2563EB; }
	.node-icon.project { background: #EDE9FE; color: #7C3AED; }
	.node-icon.learning { background: #D1FAE5; color: #059669; }
	.node-icon.operational { background: #FED7AA; color: #EA580C; }

	.node-type h4 {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	.node-type p {
		font-size: 12px;
		color: #64748B;
		margin: 0;
	}

	/* ===== PIPELINE ===== */
	.pipeline-stages {
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	.stage {
		text-align: center;
		flex: 1;
	}

	.stage-dot {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		margin: 0 auto 8px;
	}

	.stage h5 {
		font-size: 13px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	.stage p {
		font-size: 11px;
		color: #64748B;
		margin: 0;
	}

	.stage-arrow {
		color: #CBD5E1;
		font-size: 20px;
		padding: 0 8px;
	}

	/* ===== SHORTCUTS ===== */
	.shortcuts-container {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 24px;
	}

	.shortcut-category h3 {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 12px;
	}

	.shortcut-list {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 10px;
		overflow: hidden;
	}

	.shortcut-row {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #F1F5F9;
	}

	.shortcut-row:last-child {
		border-bottom: none;
	}

	.shortcut-row .keys {
		width: 180px;
		flex-shrink: 0;
	}

	.shortcut-row .desc {
		font-size: 13px;
		color: #64748B;
	}

	/* ===== PRIVACY NOTE ===== */
	.privacy-note {
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

	.privacy-note p {
		font-size: 13px;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== SPOTLIGHT PREVIEW ===== */
	.spotlight-preview {
		margin-bottom: 24px;
	}

	.spotlight-mock {
		max-width: 500px;
		margin: 0 auto;
	}

	.mock-search {
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
	.commands-section {
		margin-top: 24px;
	}

	.commands-section h3 {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.commands-section > p {
		font-size: 13px;
		color: #64748B;
		margin: 0 0 16px;
	}

	.commands-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	.command {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 8px;
		padding: 12px 16px;
		font-size: 13px;
		color: #475569;
	}

	.command code {
		background: #F1F5F9;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		color: #6366F1;
		margin-right: 6px;
	}

	/* ===== SETTINGS GRID ===== */
	.settings-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
	}

	.setting-card {
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 24px;
	}

	.setting-card > :global(svg) {
		color: #6366F1;
		margin-bottom: 12px;
	}

	.setting-card h4 {
		font-size: 15px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 8px;
	}

	.setting-card p {
		font-size: 13px;
		color: #64748B;
		line-height: 1.6;
		margin: 0;
	}

	/* ===== FEATURE LIST ===== */
	.feature-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.feature-item {
		display: flex;
		gap: 16px;
		background: white;
		border: 1px solid #E2E8F0;
		border-radius: 12px;
		padding: 20px;
	}

	.item-icon {
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

	.item-content h4 {
		font-size: 14px;
		font-weight: 600;
		color: #1E293B;
		margin: 0 0 4px;
	}

	.item-content p {
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
		.feature-grid.four-col { grid-template-columns: repeat(2, 1fr); }
		.node-types { grid-template-columns: repeat(2, 1fr); }
		.modules-showcase { grid-template-columns: 1fr; }
	}

	@media (max-width: 900px) {
		.help-content {
			padding: 30px;
		}

		.feature-grid.three-col,
		.feature-grid.two-col,
		.mode-grid,
		.concept-grid,
		.shortcuts-container,
		.settings-grid,
		.commands-grid,
		.mini-modules-grid {
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

		.section-header {
			flex-direction: column;
			align-items: flex-start;
		}

		.pipeline-stages {
			flex-direction: column;
			gap: 16px;
		}

		.stage-arrow {
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

	:global(.dark) .sidebar-search {
		background: #2a2a2a;
		border-color: rgba(255, 255, 255, 0.1);
		color: #fff;
	}

	:global(.dark) .sidebar-search::placeholder {
		color: #888;
	}

	:global(.dark) .nav-section-title {
		color: #888;
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

	:global(.dark) .sub-item {
		color: #888;
	}

	:global(.dark) .sub-item:hover,
	:global(.dark) .sub-item.active {
		color: #818cf8;
	}

	:global(.dark) .help-content {
		background: #0f0f0f;
	}

	:global(.dark) .section {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .section-title {
		color: #fff;
	}

	:global(.dark) .section-desc {
		color: #888;
	}

	:global(.dark) .feature-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .feature-card:hover {
		border-color: rgba(99, 102, 241, 0.4);
	}

	:global(.dark) .feature-card h3 {
		color: #fff;
	}

	:global(.dark) .feature-card p {
		color: #aaa;
	}

	:global(.dark) .info-card {
		background: rgba(99, 102, 241, 0.1);
		border-color: rgba(99, 102, 241, 0.2);
	}

	:global(.dark) .info-card h4 {
		color: #818cf8;
	}

	:global(.dark) .info-card p {
		color: #aaa;
	}

	:global(.dark) .tip-card {
		background: rgba(16, 185, 129, 0.1);
		border-color: rgba(16, 185, 129, 0.2);
	}

	:global(.dark) .tip-card h4 {
		color: #34d399;
	}

	:global(.dark) .tip-card p,
	:global(.dark) .tip-card li {
		color: #aaa;
	}

	:global(.dark) .mode-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .mode-card h3 {
		color: #fff;
	}

	:global(.dark) .mode-card p {
		color: #aaa;
	}

	:global(.dark) .mode-card ul li {
		color: #888;
	}

	:global(.dark) .concept-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .concept-card h3 {
		color: #fff;
	}

	:global(.dark) .concept-card p {
		color: #aaa;
	}

	:global(.dark) .shortcut-category h3 {
		color: #fff;
	}

	:global(.dark) .shortcut {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .shortcut span {
		color: #aaa;
	}

	:global(.dark) .shortcut kbd {
		background: #333;
		border-color: rgba(255, 255, 255, 0.15);
		color: #fff;
	}

	:global(.dark) .module-showcase-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .module-showcase-card h3 {
		color: #fff;
	}

	:global(.dark) .module-showcase-card > p {
		color: #888;
	}

	:global(.dark) .showcase-features li {
		color: #aaa;
	}

	:global(.dark) .mini-module {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .mini-module h4 {
		color: #fff;
	}

	:global(.dark) .mini-module p {
		color: #888;
	}

	:global(.dark) .node-type {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .node-type h4 {
		color: #fff;
	}

	:global(.dark) .node-type p {
		color: #888;
	}

	:global(.dark) .pipeline-stage {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .stage-icon {
		background: rgba(99, 102, 241, 0.15);
	}

	:global(.dark) .pipeline-stage h4 {
		color: #fff;
	}

	:global(.dark) .pipeline-stage p {
		color: #888;
	}

	:global(.dark) .stage-arrow {
		color: #444;
	}

	:global(.dark) .command-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .command-card h4 {
		color: #fff;
	}

	:global(.dark) .command-card p {
		color: #888;
	}

	:global(.dark) .setting-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .setting-card h3 {
		color: #fff;
	}

	:global(.dark) .setting-card p {
		color: #888;
	}

	:global(.dark) .setting-option {
		background: #1a1a1a;
	}

	:global(.dark) .setting-option span {
		color: #aaa;
	}

	:global(.dark) .feature-item {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .item-icon {
		background: rgba(99, 102, 241, 0.15);
	}

	:global(.dark) .item-content h4 {
		color: #fff;
	}

	:global(.dark) .item-content p {
		color: #888;
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

	/* ===== DARK MODE - Additional Components ===== */

	/* Sidebar Header & Footer */
	:global(.dark) .sidebar-header {
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .header-text h2 {
		color: #fff;
	}

	:global(.dark) .header-text .version {
		color: #666;
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

	/* Navigation */
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

	/* Section Headers */
	:global(.dark) .section-header h1 {
		color: #fff;
	}

	:global(.dark) .section-subtitle {
		color: #888;
	}

	/* Card Icons */
	:global(.dark) .card-icon {
		background: rgba(99, 102, 241, 0.15);
	}

	:global(.dark) .card-icon.blue {
		background: rgba(59, 130, 246, 0.2);
		color: #60a5fa;
	}

	:global(.dark) .card-icon.purple {
		background: rgba(139, 92, 246, 0.2);
		color: #a78bfa;
	}

	:global(.dark) .card-icon.green {
		background: rgba(16, 185, 129, 0.2);
		color: #34d399;
	}

	:global(.dark) .card-icon.orange {
		background: rgba(245, 158, 11, 0.2);
		color: #fbbf24;
	}

	/* Subsections */
	:global(.dark) .subsection h2 {
		color: #fff;
	}

	:global(.dark) .subsection-intro {
		color: #aaa;
	}

	/* Intro Card */
	:global(.dark) .intro-card {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .intro-card p {
		color: #bbb;
	}

	/* Steps List */
	:global(.dark) .step-item {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .step-number {
		background: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%);
		color: white;
	}

	:global(.dark) .step-content h4 {
		color: #fff;
	}

	:global(.dark) .step-content p {
		color: #aaa;
	}

	:global(.dark) .step-content kbd {
		background: #333;
		border-color: rgba(255, 255, 255, 0.15);
		color: #fff;
	}

	/* Concept Cards */
	:global(.dark) .concept-card h4 {
		color: #fff;
	}

	/* Mode Cards */
	:global(.dark) .mode-header {
		color: #fff;
	}

	:global(.dark) .mode-header h3 {
		color: #fff;
	}

	:global(.dark) .mode-options {
		background: rgba(255, 255, 255, 0.03);
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .mode-options h5 {
		color: #888;
	}

	:global(.dark) .mode-options li {
		color: #aaa;
	}

	:global(.dark) .mode-options li strong {
		color: #ddd;
	}

	:global(.dark) .mode-example {
		background: rgba(255, 255, 255, 0.05);
		border-color: rgba(255, 255, 255, 0.1);
		color: #aaa;
	}

	:global(.dark) .example-label {
		color: #888;
	}

	/* Spotlight Preview */
	:global(.dark) .spotlight-preview {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .spotlight-mock {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.1);
	}

	:global(.dark) .mock-search {
		background: #333;
		border-color: rgba(255, 255, 255, 0.1);
		color: #888;
	}

	/* Modules Showcase */
	:global(.dark) .modules-showcase {
		background: transparent;
	}

	:global(.dark) .module-header {
		background: rgba(99, 102, 241, 0.1);
	}

	:global(.dark) .module-header h3 {
		color: #fff;
	}

	:global(.dark) .module-icon {
		background: rgba(99, 102, 241, 0.2);
	}

	:global(.dark) .module-tagline {
		color: #aaa;
	}

	:global(.dark) .module-detail {
		background: #1a1a1a;
	}

	:global(.dark) .module-detail > p {
		color: #aaa;
	}

	:global(.dark) .module-features span {
		background: rgba(99, 102, 241, 0.15);
		color: #818cf8;
	}

	:global(.dark) .more-modules h3 {
		color: #fff;
	}

	/* Mini Modules */
	:global(.dark) .mini-module h5 {
		color: #fff;
	}

	:global(.dark) .mini-icon {
		background: rgba(99, 102, 241, 0.15);
		color: #818cf8;
	}

	/* Node Types */
	:global(.dark) .node-icon {
		background: rgba(99, 102, 241, 0.15);
	}

	:global(.dark) .node-icon.business {
		background: rgba(59, 130, 246, 0.2);
	}

	:global(.dark) .node-icon.project {
		background: rgba(139, 92, 246, 0.2);
	}

	:global(.dark) .node-icon.operational {
		background: rgba(16, 185, 129, 0.2);
	}

	:global(.dark) .node-icon.learning {
		background: rgba(245, 158, 11, 0.2);
	}

	/* Pipeline Stages */
	:global(.dark) .stage {
		background: #1a1a1a;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .stage h5 {
		color: #fff;
	}

	:global(.dark) .stage p {
		color: #888;
	}

	:global(.dark) .stage-dot {
		background: #6366F1;
	}

	/* Commands */
	:global(.dark) .commands-section h3 {
		color: #fff;
	}

	:global(.dark) .commands-section > p {
		color: #888;
	}

	:global(.dark) .command {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
		color: #aaa;
	}

	:global(.dark) .command code {
		background: rgba(99, 102, 241, 0.2);
		color: #818cf8;
	}

	/* Customization Cards */
	:global(.dark) .customization-card {
		background: #242424;
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .customization-card h4 {
		color: #fff;
	}

	:global(.dark) .customization-card > p {
		color: #aaa;
	}

	:global(.dark) .customization-card li {
		color: #888;
	}

	/* Keyboard Hints */
	:global(.dark) .keyboard-hint {
		background: #333;
		border-color: rgba(255, 255, 255, 0.15);
		color: #fff;
	}

	/* Shortcut Rows */
	:global(.dark) .shortcut-row {
		border-color: rgba(255, 255, 255, 0.08);
	}

	:global(.dark) .shortcut-row .keys kbd {
		background: #333;
		border-color: rgba(255, 255, 255, 0.15);
		color: #fff;
	}

	:global(.dark) .shortcut-row .desc {
		color: #aaa;
	}

	/* Privacy Note */
	:global(.dark) .privacy-note {
		background: rgba(16, 185, 129, 0.1);
		border-color: rgba(16, 185, 129, 0.2);
	}

	:global(.dark) .privacy-note p {
		color: #aaa;
	}

	/* Setting Card Heading Fix */
	:global(.dark) .setting-card h4 {
		color: #fff;
	}

	:global(.dark) .setting-card > :global(svg) {
		color: #818cf8;
	}
</style>
