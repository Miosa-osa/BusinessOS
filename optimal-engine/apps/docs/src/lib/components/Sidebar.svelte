<script lang="ts">
	import { page } from '$app/stores';
	import { nav } from '$lib/data/nav.js';
</script>

<nav class="sidebar">
	<div class="sidebar-inner">
		{#each nav as section}
			{#if section.title}
				<div class="section-label">{section.title}</div>
			{/if}
			<ul>
				{#each section.items as item}
					<li>
						<a
							href={item.href}
							class="nav-link"
							class:active={$page.url.pathname === item.href ||
								($page.url.pathname.startsWith(item.href + '/') && item.href !== '/')}
						>
							{item.title}
						</a>
					</li>
				{/each}
			</ul>
		{/each}
	</div>
</nav>

<style>
	.sidebar {
		width: var(--sidebar-width);
		flex-shrink: 0;
		position: sticky;
		top: var(--header-height);
		height: calc(100vh - var(--header-height));
		overflow-y: auto;
		border-right: 1px solid var(--border);
		padding: 1.5rem 0;
	}

	.sidebar-inner {
		padding: 0 1rem;
	}

	.section-label {
		font-size: 0.6875rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: var(--text-subtle);
		margin-top: 1.5rem;
		margin-bottom: 0.375rem;
		padding: 0 0.5rem;
	}

	ul {
		list-style: none;
		padding: 0;
		margin: 0 0 0.25rem;
	}

	.nav-link {
		display: block;
		padding: 0.3125rem 0.5rem;
		border-radius: var(--r-sm);
		font-size: 0.875rem;
		color: var(--text-muted);
		text-decoration: none;
		transition:
			background 0.12s,
			color 0.12s;
		line-height: 1.4;
	}

	.nav-link:hover {
		background: var(--bg-hover);
		color: var(--text);
		text-decoration: none;
	}

	.nav-link.active {
		background: var(--accent-dim);
		color: var(--accent);
		font-weight: 500;
	}
</style>
