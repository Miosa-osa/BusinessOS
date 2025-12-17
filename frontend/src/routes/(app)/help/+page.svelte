<script lang="ts">
	import { windowStore } from '$lib/stores/windowStore';
	import { page } from '$app/stores';

	const isEmbedded = $derived($page.url.searchParams.get('embed') === 'true');

	// Sections for navigation
	const sections = [
		{ id: 'getting-started', title: 'Getting Started', icon: '🚀' },
		{ id: 'focus-modes', title: 'AI Focus Modes', icon: '🎯' },
		{ id: 'shortcuts', title: 'Keyboard Shortcuts', icon: '⌨️' },
		{ id: 'desktop', title: 'Desktop Features', icon: '🖥️' },
		{ id: 'clients', title: 'Clients & CRM', icon: '👔' },
		{ id: 'nodes', title: 'Nodes System', icon: '🧠' },
		{ id: 'documents', title: 'Documents', icon: '📄' },
		{ id: 'quick-chat', title: 'Quick Chat', icon: '💬' },
		{ id: 'spotlight', title: 'Spotlight Search', icon: '🔍' },
		{ id: 'modules', title: 'Modules', icon: '📦' },
		{ id: 'voice', title: 'Voice Commands', icon: '🎤' },
	];

	let activeSection = $state('getting-started');

	function scrollToSection(id: string) {
		activeSection = id;
		const element = document.getElementById(id);
		element?.scrollIntoView({ behavior: 'smooth' });
	}
</script>

<svelte:head>
	<title>Help - Business OS</title>
</svelte:head>

<div class="help-page" class:embedded={isEmbedded}>
	<!-- Sidebar Navigation -->
	<aside class="help-sidebar">
		<div class="sidebar-header">
			<h2>📚 Help Center</h2>
		</div>
		<nav class="sidebar-nav">
			{#each sections as section}
				<button
					class="nav-item"
					class:active={activeSection === section.id}
					onclick={() => scrollToSection(section.id)}
				>
					<span class="nav-icon">{section.icon}</span>
					<span class="nav-text">{section.title}</span>
				</button>
			{/each}
		</nav>
		<div class="sidebar-footer">
			<p>Business OS v1.0</p>
		</div>
	</aside>

	<!-- Main Content -->
	<main class="help-content">
		<!-- Getting Started -->
		<section id="getting-started" class="help-section">
			<h1>Getting Started with Business OS</h1>
			<p class="section-intro">
				Business OS is your all-in-one operating system for managing your business.
				It combines project management, AI assistance, team collaboration, and client
				management in a beautiful desktop-like interface.
			</p>

			<div class="feature-grid">
				<div class="feature-card">
					<span class="feature-icon">🖥️</span>
					<h3>Desktop Environment</h3>
					<p>A macOS-inspired desktop with windows, folders, and a dock for quick access to all your tools.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">🤖</span>
					<h3>AI Assistant</h3>
					<p>Built-in AI chat that understands your business context and helps with tasks, analysis, and planning.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">📊</span>
					<h3>Project Management</h3>
					<p>Track projects, tasks, and team progress with powerful visualization tools.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">👥</span>
					<h3>Team Collaboration</h3>
					<p>Manage your team, assign tasks, and track availability all in one place.</p>
				</div>
			</div>
		</section>

		<!-- AI Focus Modes -->
		<section id="focus-modes" class="help-section">
			<h1>AI Focus Modes</h1>
			<p class="section-intro">
				Focus Modes provide specialized AI assistance for different types of work.
				Each mode uses a tailored agent with specific prompts and capabilities.
			</p>

			<div class="feature-grid">
				<div class="feature-card">
					<span class="feature-icon">🔍</span>
					<h3>Research</h3>
					<p>Search the web and documents for information. Configure scope (Web/Docs/All), depth (Quick/Thorough), and output format (Summary/Report).</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">📊</span>
					<h3>Analyze</h3>
					<p>Deep analysis of data and situations. Choose approach (Validate/Compare/Forecast), depth level, and output (Findings/Dashboard).</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">✍️</span>
					<h3>Write</h3>
					<p>Create content and documents. Select format (Document/Slides/Spreadsheet) and mode (Step-by-step/First Draft).</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">🏗️</span>
					<h3>Build</h3>
					<p>Strategy and planning assistance. Create frameworks, SOPs, or plans with outline or detailed output.</p>
				</div>
			</div>

			<div class="feature-list" style="margin-top: 24px;">
				<div class="feature-item">
					<h3>💡 How to Use Focus Modes</h3>
					<ul>
						<li>Open the Chat module and look for the Focus Mode cards above the input</li>
						<li>Click a mode card to select it - options will expand below</li>
						<li>Configure the options for your specific task</li>
						<li>Type your message and send - the AI will use the specialized agent</li>
						<li>Use "Do More" mode for general chat without specific focus</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Keyboard Shortcuts -->
		<section id="shortcuts" class="help-section">
			<h1>Keyboard Shortcuts</h1>
			<p class="section-intro">Master these shortcuts to navigate Business OS like a pro.</p>

			<div class="shortcuts-table">
				<h3>Global Shortcuts</h3>
				<div class="shortcut-group">
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>Space</kbd>
						</div>
						<div class="shortcut-desc">Open Spotlight Search / Quick Chat</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>Shift</kbd> + <kbd>Space</kbd>
						</div>
						<div class="shortcut-desc">Open Quick Chat Popup (global)</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>W</kbd>
						</div>
						<div class="shortcut-desc">Close current window</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>M</kbd>
						</div>
						<div class="shortcut-desc">Minimize current window</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>`</kbd>
						</div>
						<div class="shortcut-desc">Cycle between open windows</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>Esc</kbd>
						</div>
						<div class="shortcut-desc">Close Spotlight / Deselect icons</div>
					</div>
				</div>

				<h3>Chat Shortcuts</h3>
				<div class="shortcut-group">
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>Enter</kbd>
						</div>
						<div class="shortcut-desc">Send message</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>Shift</kbd> + <kbd>Enter</kbd>
						</div>
						<div class="shortcut-desc">New line in message</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>⌘</kbd> + <kbd>Enter</kbd>
						</div>
						<div class="shortcut-desc">Open full chat from popup</div>
					</div>
				</div>

				<h3>Spotlight Commands</h3>
				<div class="shortcut-group">
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>Tab</kbd>
						</div>
						<div class="shortcut-desc">Switch between Search and Chat modes</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>↑</kbd> / <kbd>↓</kbd>
						</div>
						<div class="shortcut-desc">Navigate search results</div>
					</div>
					<div class="shortcut-row">
						<div class="shortcut-keys">
							<kbd>Enter</kbd>
						</div>
						<div class="shortcut-desc">Select result / Send message</div>
					</div>
				</div>
			</div>
		</section>

		<!-- Desktop Features -->
		<section id="desktop" class="help-section">
			<h1>Desktop Features</h1>
			<p class="section-intro">Your workspace is designed to feel familiar and powerful.</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>🖱️ Windows</h3>
					<ul>
						<li>Drag windows by their title bar to move them</li>
						<li>Drag edges or corners to resize</li>
						<li>Double-click title bar to maximize/restore</li>
						<li>Drag to screen edges to snap (left, right, corners)</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>📁 Folders</h3>
					<ul>
						<li>Right-click desktop → "New Folder" to create</li>
						<li>Drag icons onto folders to organize</li>
						<li>Right-click folders to rename or delete</li>
						<li>Pin folders to the dock for quick access</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🎨 Desktop Customization</h3>
					<ul>
						<li>Right-click desktop → "Desktop Settings"</li>
						<li><strong>50+ Backgrounds:</strong> Solid colors, gradients (Sunrise, Ocean, Aurora), patterns (Dots, Grid, Blueprint)</li>
						<li><strong>15 Icon Styles:</strong> macOS, Windows 95, Neon, Glassmorphism, Pixel, Minimal, and more</li>
						<li>Adjust icon size from 32px to 128px</li>
						<li>Upload custom background images</li>
						<li>Enable/disable noise texture overlay for film grain effect</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>⚓ Dock</h3>
					<ul>
						<li>Click icons to open or focus apps</li>
						<li>Drag icons from desktop to pin to dock</li>
						<li>Right-click icons → "Add to Dock"</li>
						<li>Hover for tooltip with app name</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Clients & CRM -->
		<section id="clients" class="help-section">
			<h1>Clients & CRM</h1>
			<p class="section-intro">
				Manage your client relationships with a full-featured CRM. Track companies,
				contacts, deals, and interactions all in one place.
			</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>👥 Client Profiles</h3>
					<ul>
						<li>Create profiles for companies or individuals</li>
						<li>Track status: Lead → Prospect → Active → Inactive → Churned</li>
						<li>Store contact info, industry, company size, and custom fields</li>
						<li>Add notes and tags for organization</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>📇 Contacts</h3>
					<ul>
						<li>Add multiple contacts per client</li>
						<li>Mark primary contacts for quick reference</li>
						<li>Store email, phone, and role information</li>
						<li>Track contact-specific notes</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>💰 Deals Pipeline</h3>
					<ul>
						<li>Create and track sales opportunities</li>
						<li>Kanban board with stages: Qualification → Proposal → Negotiation → Closed</li>
						<li>Set deal value and probability</li>
						<li>Drag deals between stages to update</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>📝 Interactions</h3>
					<ul>
						<li>Log calls, emails, meetings, and notes</li>
						<li>Track interaction outcomes</li>
						<li>View complete interaction history</li>
						<li>Associate interactions with contacts</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Nodes System -->
		<section id="nodes" class="help-section">
			<h1>Nodes System</h1>
			<p class="section-intro">
				Nodes are your cognitive operating system - a hierarchical structure for
				organizing and managing different areas of your business.
			</p>

			<div class="feature-grid">
				<div class="feature-card">
					<span class="feature-icon">🏢</span>
					<h3>Business</h3>
					<p>Top-level business units, departments, or major initiatives.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">📁</span>
					<h3>Project</h3>
					<p>Specific projects within your business structure.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">📚</span>
					<h3>Learning</h3>
					<p>Knowledge areas, training, and skill development.</p>
				</div>
				<div class="feature-card">
					<span class="feature-icon">⚙️</span>
					<h3>Operational</h3>
					<p>Day-to-day operations and processes.</p>
				</div>
			</div>

			<div class="feature-list" style="margin-top: 24px;">
				<div class="feature-item">
					<h3>🎯 Node Features</h3>
					<ul>
						<li><strong>Health Status:</strong> Healthy, Needs Attention, Critical, Not Started</li>
						<li><strong>Purpose:</strong> Define what this node is about</li>
						<li><strong>Current Status:</strong> What's happening now</li>
						<li><strong>Weekly Focus:</strong> This week's priorities</li>
						<li><strong>Decision Queue:</strong> Pending decisions</li>
						<li><strong>Delegation Ready:</strong> Tasks ready to delegate</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>⚡ Active Node</h3>
					<ul>
						<li>Set a node as "active" to focus your AI context on that area</li>
						<li>The active node provides context for chat conversations</li>
						<li>Only one node can be active at a time</li>
						<li>Click "Activate" on any node to set it as active</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🌳 Views</h3>
					<ul>
						<li><strong>Tree View:</strong> Hierarchical nested structure</li>
						<li><strong>List View:</strong> Flat list with filters</li>
						<li><strong>Grid View:</strong> Card-based overview</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Documents -->
		<section id="documents" class="help-section">
			<h1>Documents (Contexts)</h1>
			<p class="section-intro">
				Create rich documents with a Notion-like block editor. Documents can store
				business knowledge that AI can reference in conversations.
			</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>📝 Block Types</h3>
					<ul>
						<li><strong>Text:</strong> Paragraphs, headings (H1, H2, H3)</li>
						<li><strong>Lists:</strong> Bullet lists, numbered lists, to-do checkboxes</li>
						<li><strong>Media:</strong> Images, embeds, code blocks</li>
						<li><strong>Structure:</strong> Quotes, callouts, dividers, tables</li>
						<li><strong>AI:</strong> Artifact blocks for AI-generated content</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>⌨️ Slash Commands</h3>
					<ul>
						<li>Type <code>/</code> to open the block menu</li>
						<li>Search and select block types</li>
						<li>Use <code>/heading</code>, <code>/bullet</code>, <code>/code</code>, etc.</li>
						<li>Press <kbd>Enter</kbd> to insert the selected block</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🏷️ Properties</h3>
					<ul>
						<li>Add custom properties to documents</li>
						<li>Property types: Text, Select, Multi-select, Date, Number, Checkbox, URL, Email</li>
						<li>Create relations to other contexts, projects, or clients</li>
						<li>Define property schemas for consistent structure</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🔗 Sharing</h3>
					<ul>
						<li>Enable public sharing to generate a unique link</li>
						<li>Share documents with anyone without login required</li>
						<li>Disable sharing at any time</li>
						<li>Duplicate documents to create copies</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Quick Chat -->
		<section id="quick-chat" class="help-section">
			<h1>Quick Chat</h1>
			<p class="section-intro">
				The chat bar above your dock gives you instant access to AI assistance without
				leaving your current work.
			</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>💬 Basic Usage</h3>
					<ul>
						<li>Click the chat bar and type your question</li>
						<li>Press <kbd>Enter</kbd> to send</li>
						<li>AI response appears inline</li>
						<li>Click expand button to open full chat</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🎤 Voice Input</h3>
					<ul>
						<li>Click the microphone button to start recording</li>
						<li>Speak your message naturally</li>
						<li>Click stop or the mic button again to finish</li>
						<li>Your speech is transcribed and ready to send</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>⚡ Quick Tips</h3>
					<ul>
						<li>Press <kbd>Esc</kbd> to close and clear</li>
						<li>The chat expands as you type more</li>
						<li>Responses show with an AI badge</li>
						<li>Click the chat icon to open full chat module</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Spotlight Search -->
		<section id="spotlight" class="help-section">
			<h1>Spotlight Search</h1>
			<p class="section-intro">
				Press <kbd>⌘</kbd> + <kbd>Space</kbd> to open Spotlight - your command center for
				searching, launching apps, and chatting with AI.
			</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>🔍 Search Mode</h3>
					<ul>
						<li>Type to search for apps, projects, tasks, and more</li>
						<li>Use arrow keys to navigate results</li>
						<li>Press <kbd>Enter</kbd> to open selected item</li>
						<li>Results are grouped by category</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>💬 Chat Mode</h3>
					<ul>
						<li>Press <kbd>Tab</kbd> to switch to Chat mode</li>
						<li>Select a context (project) for relevant answers</li>
						<li>Use voice input for hands-free interaction</li>
						<li>Responses stream in real-time</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>⚡ Quick Commands</h3>
					<p>Type these commands for instant actions:</p>
					<div class="command-list">
						<code>/task</code> - Create a new task<br>
						<code>/note</code> - Create a quick note<br>
						<code>/project</code> - Create a new project<br>
						<code>/meet</code> - Schedule a meeting<br>
						<code>/remind</code> - Set a reminder
					</div>
				</div>
			</div>
		</section>

		<!-- Modules -->
		<section id="modules" class="help-section">
			<h1>Modules</h1>
			<p class="section-intro">Business OS includes powerful modules for every aspect of your business.</p>

			<div class="modules-grid">
				<div class="module-card">
					<div class="module-icon" style="background: #E3F2FD; color: #1E88E5;">📊</div>
					<h3>Dashboard</h3>
					<p>Overview of your business metrics, today's focus, and quick actions.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #E8F5E9; color: #43A047;">💬</div>
					<h3>Chat</h3>
					<p>AI-powered assistant with Focus Modes for specialized tasks.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #FFF3E0; color: #FB8C00;">✅</div>
					<h3>Tasks</h3>
					<p>Kanban boards, list views, and calendar views with priorities.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #F3E5F5; color: #8E24AA;">📁</div>
					<h3>Projects</h3>
					<p>Organize work into projects with status tracking and notes.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #E0F7FA; color: #00ACC1;">👥</div>
					<h3>Team</h3>
					<p>Org chart, capacity planning, and workload management.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #FCE4EC; color: #E91E63;">📅</div>
					<h3>Calendar</h3>
					<p>Events with Google Calendar sync and meeting management.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #EDE7F6; color: #5E35B1;">📄</div>
					<h3>Documents</h3>
					<p>Notion-like editor with blocks, properties, and sharing.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #FFF8E1; color: #FFA000;">👔</div>
					<h3>Clients</h3>
					<p>Full CRM with contacts, deals pipeline, and interactions.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #E8EAF6; color: #3F51B5;">🧠</div>
					<h3>Nodes</h3>
					<p>Hierarchical business structure and cognitive workspace.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #EFEBE9; color: #795548;">📝</div>
					<h3>Daily Log</h3>
					<p>Track your day, patterns, and reflections.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #FFEBEE; color: #E53935;">🎤</div>
					<h3>Voice Notes</h3>
					<p>Record and transcribe voice memos with AI.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #E0F2F1; color: #00897B;">📈</div>
					<h3>Usage</h3>
					<p>Analytics on AI usage, tokens, and costs by provider.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #F5F5F5; color: #616161;">⚙️</div>
					<h3>Settings</h3>
					<p>Configure AI models, providers, and integrations.</p>
				</div>

				<div class="module-card">
					<div class="module-icon" style="background: #1E1E1E; color: #00FF00;">⚡</div>
					<h3>Terminal</h3>
					<p>OS Agent - command-line interface for AI interaction.</p>
				</div>
			</div>
		</section>

		<!-- Voice Commands -->
		<section id="voice" class="help-section">
			<h1>Voice Commands</h1>
			<p class="section-intro">
				Use your voice to interact with Business OS hands-free. Voice input is available
				in the Quick Chat and Spotlight Search.
			</p>

			<div class="feature-list">
				<div class="feature-item">
					<h3>🎙️ How to Use</h3>
					<ul>
						<li>Click the microphone icon to start recording</li>
						<li>Speak clearly - a waveform shows your audio</li>
						<li>Click stop when finished</li>
						<li>Your speech is automatically transcribed</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>💡 Tips for Best Results</h3>
					<ul>
						<li>Speak in a quiet environment</li>
						<li>Use a clear, natural speaking pace</li>
						<li>Punctuation is added automatically</li>
						<li>You can edit the transcription before sending</li>
					</ul>
				</div>

				<div class="feature-item">
					<h3>🔒 Privacy</h3>
					<ul>
						<li>Audio is processed securely</li>
						<li>Recordings are not stored permanently</li>
						<li>Transcription happens on our secure servers</li>
					</ul>
				</div>
			</div>
		</section>

		<!-- Footer -->
		<footer class="help-footer">
			<p>Need more help? Contact support or check our documentation.</p>
			<div class="footer-links">
				<a href="mailto:roberto@osa.dev">Email Support</a>
				<span>•</span>
				<a href="/docs" target="_blank">Documentation</a>
				<span>•</span>
				<a href="https://github.com/robertohluna/BusinessOS" target="_blank">GitHub</a>
			</div>
		</footer>
	</main>
</div>

<style>
	.help-page {
		display: flex;
		min-height: 100vh;
		background: #FAFAFA;
	}

	.help-page.embedded {
		min-height: 100%;
	}

	/* Sidebar */
	.help-sidebar {
		width: 240px;
		background: white;
		border-right: 1px solid #E5E5E5;
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
		padding: 20px;
		border-bottom: 1px solid #E5E5E5;
	}

	.sidebar-header h2 {
		font-size: 16px;
		font-weight: 600;
		color: #333;
		margin: 0;
	}

	.sidebar-nav {
		flex: 1;
		padding: 12px;
		overflow-y: auto;
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
	}

	.nav-item:hover {
		background: #F5F5F5;
	}

	.nav-item.active {
		background: #EEF2FF;
		color: #4F46E5;
	}

	.nav-icon {
		font-size: 16px;
	}

	.nav-text {
		font-size: 13px;
		font-weight: 500;
	}

	.sidebar-footer {
		padding: 16px 20px;
		border-top: 1px solid #E5E5E5;
		font-size: 11px;
		color: #999;
	}

	/* Main Content */
	.help-content {
		flex: 1;
		padding: 40px;
		max-width: 900px;
		overflow-y: auto;
	}

	.help-section {
		margin-bottom: 60px;
	}

	.help-section h1 {
		font-size: 28px;
		font-weight: 700;
		color: #111;
		margin: 0 0 12px;
	}

	.section-intro {
		font-size: 16px;
		color: #666;
		line-height: 1.6;
		margin-bottom: 32px;
	}

	/* Feature Grid */
	.feature-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: 20px;
	}

	.feature-card {
		background: white;
		border: 1px solid #E5E5E5;
		border-radius: 12px;
		padding: 24px;
	}

	.feature-icon {
		font-size: 32px;
		display: block;
		margin-bottom: 12px;
	}

	.feature-card h3 {
		font-size: 16px;
		font-weight: 600;
		margin: 0 0 8px;
		color: #333;
	}

	.feature-card p {
		font-size: 13px;
		color: #666;
		line-height: 1.5;
		margin: 0;
	}

	/* Shortcuts Table */
	.shortcuts-table h3 {
		font-size: 14px;
		font-weight: 600;
		color: #333;
		margin: 24px 0 12px;
		padding-bottom: 8px;
		border-bottom: 1px solid #E5E5E5;
	}

	.shortcuts-table h3:first-child {
		margin-top: 0;
	}

	.shortcut-group {
		background: white;
		border: 1px solid #E5E5E5;
		border-radius: 8px;
		overflow: hidden;
	}

	.shortcut-row {
		display: flex;
		align-items: center;
		padding: 12px 16px;
		border-bottom: 1px solid #F0F0F0;
	}

	.shortcut-row:last-child {
		border-bottom: none;
	}

	.shortcut-keys {
		width: 200px;
		flex-shrink: 0;
	}

	.shortcut-keys kbd {
		display: inline-block;
		padding: 4px 8px;
		background: #F5F5F5;
		border: 1px solid #E0E0E0;
		border-radius: 4px;
		font-family: system-ui, -apple-system, sans-serif;
		font-size: 12px;
		font-weight: 500;
		color: #333;
		box-shadow: 0 1px 0 #D0D0D0;
	}

	.shortcut-desc {
		font-size: 13px;
		color: #666;
	}

	/* Feature List */
	.feature-list {
		display: flex;
		flex-direction: column;
		gap: 24px;
	}

	.feature-item {
		background: white;
		border: 1px solid #E5E5E5;
		border-radius: 12px;
		padding: 24px;
	}

	.feature-item h3 {
		font-size: 16px;
		font-weight: 600;
		margin: 0 0 12px;
		color: #333;
	}

	.feature-item ul {
		margin: 0;
		padding-left: 20px;
	}

	.feature-item li {
		font-size: 13px;
		color: #666;
		line-height: 1.8;
	}

	.feature-item p {
		font-size: 13px;
		color: #666;
		margin: 0 0 12px;
	}

	.command-list {
		background: #F8F8F8;
		padding: 12px 16px;
		border-radius: 6px;
		font-size: 13px;
		line-height: 1.8;
	}

	.command-list code {
		background: #E8E8E8;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, monospace;
		font-size: 12px;
		color: #4F46E5;
	}

	/* Modules Grid */
	.modules-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
	}

	.module-card {
		background: white;
		border: 1px solid #E5E5E5;
		border-radius: 12px;
		padding: 20px;
		text-align: center;
	}

	.module-icon {
		width: 48px;
		height: 48px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 24px;
		margin: 0 auto 12px;
	}

	.module-card h3 {
		font-size: 14px;
		font-weight: 600;
		margin: 0 0 8px;
		color: #333;
	}

	.module-card p {
		font-size: 12px;
		color: #666;
		line-height: 1.4;
		margin: 0;
	}

	/* Footer */
	.help-footer {
		margin-top: 60px;
		padding-top: 40px;
		border-top: 1px solid #E5E5E5;
		text-align: center;
	}

	.help-footer p {
		font-size: 14px;
		color: #666;
		margin: 0 0 16px;
	}

	.footer-links {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 12px;
		font-size: 13px;
	}

	.footer-links a {
		color: #4F46E5;
		text-decoration: none;
	}

	.footer-links a:hover {
		text-decoration: underline;
	}

	.footer-links span {
		color: #DDD;
	}

	/* Responsive */
	@media (max-width: 900px) {
		.feature-grid,
		.modules-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 700px) {
		.help-sidebar {
			display: none;
		}

		.help-content {
			padding: 24px;
		}
	}
</style>
