<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import EngineThemeToggle from '$lib/components/EngineThemeToggle.svelte';

	let { children } = $props();
</script>

<div class="shell">
	<header class="glass-header">
		<div class="header-inner">
			<a href="/" class="wordmark">
				<svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true">
					<circle cx="12" cy="12" r="10" stroke="var(--accent)" stroke-width="1.5" />
					<path d="M8 12h8M12 8v8" stroke="var(--accent)" stroke-width="1.5" stroke-linecap="round" />
				</svg>
				<span>Optimal Engine</span>
				<span class="badge">docs</span>
			</a>

			<nav class="header-nav">
				<a href="/quickstart" class="header-link">Quickstart</a>
				<a href="/concepts/three-tiers" class="header-link">Concepts</a>
				<a href="/api" class="header-link">API</a>
				<a href="/sdks/typescript" class="header-link">SDKs</a>
				<a href="/self-host" class="header-link">Self-host</a>
			</nav>

			<div class="header-actions">
				<a
					href="https://github.com/Miosa-osa/OptimalEngine"
					target="_blank"
					rel="noopener"
					class="pill pill-ghost"
					aria-label="GitHub"
				>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
						<path
							d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
						/>
					</svg>
					GitHub
				</a>
				<EngineThemeToggle />
			</div>
		</div>
	</header>

	<div class="body-row">
		<Sidebar />
		<main class="main-content">
			<div class="content-inner">
				{@render children()}
			</div>
		</main>
	</div>
</div>

<style>
	.shell {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	/* Glass header */
	.glass-header {
		position: sticky;
		top: 0;
		z-index: 100;
		height: var(--header-height);
		background: var(--bg-glass);
		backdrop-filter: blur(16px) saturate(1.4);
		-webkit-backdrop-filter: blur(16px) saturate(1.4);
		border-bottom: 1px solid var(--border);
	}

	.header-inner {
		max-width: 1400px;
		margin: 0 auto;
		padding: 0 1.5rem;
		height: 100%;
		display: flex;
		align-items: center;
		gap: 2rem;
	}

	.wordmark {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-weight: 600;
		font-size: 0.9375rem;
		color: var(--text);
		text-decoration: none;
		flex-shrink: 0;
	}

	.wordmark:hover {
		text-decoration: none;
		color: var(--text);
	}

	.badge {
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		background: var(--accent-dim);
		color: var(--accent);
		padding: 0.125rem 0.4rem;
		border-radius: var(--r-pill);
		border: 1px solid var(--border-accent);
	}

	.header-nav {
		display: flex;
		align-items: center;
		gap: 0.25rem;
		flex: 1;
	}

	.header-link {
		font-size: 0.875rem;
		color: var(--text-muted);
		padding: 0.375rem 0.625rem;
		border-radius: var(--r-sm);
		text-decoration: none;
		transition: color 0.12s, background 0.12s;
	}

	.header-link:hover {
		color: var(--text);
		background: var(--bg-hover);
		text-decoration: none;
	}

	.header-actions {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-left: auto;
	}

	/* Body */
	.body-row {
		display: flex;
		flex: 1;
		max-width: 1400px;
		margin: 0 auto;
		width: 100%;
	}

	.main-content {
		flex: 1;
		min-width: 0;
		padding: 2.5rem 3rem;
	}

	.content-inner {
		max-width: 780px;
	}

	@media (max-width: 768px) {
		.header-nav {
			display: none;
		}

		.main-content {
			padding: 1.5rem 1.25rem;
		}
	}
</style>
