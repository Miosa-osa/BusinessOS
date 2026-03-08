<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		/** Optional accent color for left border strip (e.g., 'blue', 'green', 'purple', 'orange', 'pink') */
		accent?: 'blue' | 'green' | 'purple' | 'orange' | 'pink' | 'cyan' | 'none';
		/** Extra CSS classes to apply */
		class?: string;
	}

	let { children, accent = 'none', class: extraClass = '' }: Props = $props();
</script>

<div class="widget-card {accent !== 'none' ? `widget-accent-${accent}` : ''} {extraClass}">
	{@render children()}
</div>

<style>
	.widget-card {
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		border-radius: 0.75rem;
		border: 1px solid var(--dbd, #e0e0e0);
		padding: 1.25rem;
		box-shadow: var(--shadow-sm, 0 1px 2px rgba(0,0,0,0.05));
		transition: box-shadow 0.3s ease, border-color 0.3s ease;
	}

	.widget-card:hover {
		box-shadow: var(--shadow-md, 0 4px 6px rgba(0,0,0,0.07));
	}

	/* Accent left border variants */
	.widget-accent-blue {
		border-left: 3px solid #3b82f6;
	}
	.widget-accent-green {
		border-left: 3px solid #22c55e;
	}
	.widget-accent-purple {
		border-left: 3px solid #a855f7;
	}
	.widget-accent-orange {
		border-left: 3px solid #f97316;
	}
	.widget-accent-pink {
		border-left: 3px solid #ec4899;
	}
	.widget-accent-cyan {
		border-left: 3px solid #06b6d4;
	}

	/* Subtle glass effect — activates automatically in dark mode via token values */
	@supports (backdrop-filter: blur(1px)) {
		:global(.dark) .widget-card {
			backdrop-filter: blur(12px);
			background: rgba(26, 26, 26, 0.82);
			border-color: rgba(255, 255, 255, 0.07);
		}

		:global(.dark) .widget-card:hover {
			border-color: rgba(255, 255, 255, 0.12);
		}
	}
</style>
