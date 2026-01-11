<!--
  ThemeToggle.svelte
  Light/dark mode toggle button
-->
<script lang="ts">
	import { SunIcon, MoonIcon } from './icons';

	interface Props {
		isDark?: boolean;
		onToggle?: (isDark: boolean) => void;
		class?: string;
	}

	let { isDark = $bindable(false), onToggle, class: className = '' }: Props = $props();

	function handleToggle() {
		isDark = !isDark;
		onToggle?.(isDark);

		// Toggle dark class on document
		if (typeof document !== 'undefined') {
			document.documentElement.classList.toggle('dark', isDark);
		}
	}
</script>

<button
	type="button"
	class="theme-toggle {className}"
	onclick={handleToggle}
	aria-label={isDark ? 'Switch to light mode' : 'Switch to dark mode'}
>
	{#if isDark}
		<SunIcon size={20} />
	{:else}
		<MoonIcon size={20} />
	{/if}
</button>

<style>
	.theme-toggle {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: none;
		background-color: var(--secondary, #f9fafb);
		color: var(--foreground, #1f2937);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.theme-toggle:hover {
		background-color: var(--accent, #f3f4f6);
	}

	:global(.dark) .theme-toggle {
		background-color: var(--secondary, #1a1a1a);
		color: var(--foreground, #f9fafb);
	}

	:global(.dark) .theme-toggle:hover {
		background-color: var(--accent, #2a2a2a);
	}
</style>
