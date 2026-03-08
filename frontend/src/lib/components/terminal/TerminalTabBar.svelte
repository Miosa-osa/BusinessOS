<script lang="ts">
	import type { TerminalTab } from '$lib/stores/terminal/terminalTypes';
	import { PROVIDER_CONFIGS } from '$lib/stores/terminal/terminalTypes';

	interface Props {
		tabs: TerminalTab[];
		activeTabId: string | null;
		onSwitchTab: (tabId: string) => void;
		onCloseTab: (tabId: string) => void;
		onNewTab: () => void;
	}

	let { tabs, activeTabId, onSwitchTab, onCloseTab, onNewTab }: Props = $props();

	function getProviderColor(tab: TerminalTab): string {
		return PROVIDER_CONFIGS.find(p => p.id === tab.provider)?.color ?? '#00ff00';
	}

	function handleMiddleClick(e: MouseEvent, tabId: string) {
		if (e.button === 1) {
			e.preventDefault();
			onCloseTab(tabId);
		}
	}
</script>

<div class="tab-bar" role="tablist">
	{#each tabs as tab, i (tab.id)}
		<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
		<div
			class="tab"
			class:active={tab.id === activeTabId}
			role="tab"
			aria-selected={tab.id === activeTabId}
			tabindex="0"
			onclick={() => onSwitchTab(tab.id)}
			onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSwitchTab(tab.id); } }}
			onauxclick={(e) => handleMiddleClick(e, tab.id)}
			title="{tab.title} ({i + 1})"
		>
			<span class="tab-dot" style="background: {getProviderColor(tab)}"></span>
			<span class="tab-title">{tab.title}</span>
			<button
				class="tab-close"
				class:disabled={tabs.length <= 1}
				onclick={(e) => { e.stopPropagation(); onCloseTab(tab.id); }}
				disabled={tabs.length <= 1}
				aria-label="Close tab {tab.title}"
			>
				&times;
			</button>
		</div>
	{/each}

	<button class="tab-add" onclick={onNewTab} aria-label="New tab">
		+
	</button>
</div>

<style>
	.tab-bar {
		display: flex;
		align-items: center;
		height: 32px;
		background: #111;
		border-bottom: 1px solid #222;
		overflow-x: auto;
		overflow-y: hidden;
		flex-shrink: 0;
	}

	.tab-bar::-webkit-scrollbar {
		height: 2px;
	}

	.tab-bar::-webkit-scrollbar-thumb {
		background: #333;
	}

	.tab {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 0 12px;
		height: 100%;
		background: transparent;
		border-right: 1px solid #1a1a1a;
		color: #888;
		font-family: 'SF Mono', monospace;
		font-size: 11px;
		cursor: pointer;
		white-space: nowrap;
		min-width: 80px;
		max-width: 160px;
		transition: background 0.1s, color 0.1s;
		outline: none;
		box-sizing: border-box;
	}

	.tab:hover {
		background: #1a1a1a;
		color: #ccc;
	}

	.tab.active {
		background: #1a1a1a;
		color: #fff;
		border-bottom: 2px solid var(--accent, #00ff00);
	}

	.tab-dot {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.tab-title {
		overflow: hidden;
		text-overflow: ellipsis;
		flex: 1;
		text-align: left;
	}

	.tab-close {
		background: transparent;
		border: none;
		color: #555;
		font-size: 14px;
		line-height: 1;
		padding: 0 2px;
		cursor: pointer;
		border-radius: 2px;
		flex-shrink: 0;
	}

	.tab-close:hover:not(:disabled) {
		background: #333;
		color: #ff5555;
	}

	.tab-close.disabled,
	.tab-close:disabled {
		opacity: 0.2;
		cursor: default;
	}

	.tab-add {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 100%;
		background: transparent;
		border: none;
		color: #555;
		font-size: 16px;
		cursor: pointer;
		flex-shrink: 0;
	}

	.tab-add:hover {
		background: #1a1a1a;
		color: #00ff00;
	}
</style>
