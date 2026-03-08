<script lang="ts">
	import { page } from '$app/stores';
	import { Calendar, Mail, MessageSquare } from 'lucide-svelte';

	let { children } = $props();

	const tabs = [
		{ href: '/communication/calendar', label: 'Calendar', icon: Calendar },
		{ href: '/communication/email', label: 'Email', icon: Mail },
		{ href: '/communication/channels', label: 'Channels', icon: MessageSquare },
	];

	const isActiveTab = (tabHref: string) => {
		return $page.url.pathname.startsWith(tabHref);
	};
</script>

<div class="ch-layout">
	<!-- Header with Tabs -->
	<div class="ch-layout__header">
		<div class="ch-layout__header-inner">
			<h1 class="ch-layout__title">Communication Hub</h1>
			<nav class="ch-layout__tabs">
				{#each tabs as tab}
					<a
						href={tab.href}
						class="ch-layout__tab"
						class:ch-layout__tab--active={isActiveTab(tab.href)}
						aria-label="{tab.label} tab"
					>
						<tab.icon size={18} strokeWidth={1.5} />
						{tab.label}
					</a>
				{/each}
			</nav>
		</div>
	</div>

	<!-- Content -->
	<div class="ch-layout__content">
		{@render children()}
	</div>
</div>

<style>
	.ch-layout {
		height: 100%;
		display: flex;
		flex-direction: column;
	}

	.ch-layout__header {
		border-bottom: 1px solid var(--dbd);
		background: var(--dbg);
		flex-shrink: 0;
	}

	.ch-layout__header-inner {
		padding: var(--space-4) var(--space-6) 0;
	}

	.ch-layout__title {
		font-size: var(--text-2xl);
		font-weight: var(--font-bold);
		color: var(--dt);
		margin: 0 0 var(--space-4) 0;
	}

	.ch-layout__tabs {
		display: flex;
		gap: var(--space-1);
	}

	.ch-layout__tab {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-4);
		font-size: var(--text-sm);
		font-weight: var(--font-medium);
		border-bottom: 2px solid transparent;
		border-radius: var(--radius-sm) var(--radius-sm) 0 0;
		color: var(--dt3);
		text-decoration: none;
		transition: all 200ms ease;
	}

	.ch-layout__tab:hover {
		color: var(--dt2);
		background: var(--dbg2);
	}

	.ch-layout__tab--active {
		color: var(--accent-blue);
		border-bottom-color: var(--accent-blue);
		background: rgba(59, 130, 246, 0.06);
	}

	.ch-layout__content {
		flex: 1;
		overflow: auto;
	}
</style>
