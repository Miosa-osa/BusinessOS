<script lang="ts">
	import { desktopSettings } from '$lib/stores/desktopStore';

	interface Props {
		isOpen: boolean;
		onToggle: () => void;
		onAction: (action: string) => void;
	}

	let { isOpen, onToggle, onAction }: Props = $props();

	const enable3D = $derived($desktopSettings.enable3DDesktop);
</script>

<!-- Logo / Desktop Settings -->
<div class="menu-bar-item-wrapper">
	<button class="menu-bar-logo" onclick={onToggle} aria-label="Desktop menu" aria-haspopup="menu" aria-expanded={isOpen}>
		<svg class="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
			<path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" stroke-width="2" fill="none"/>
		</svg>
	</button>

	{#if isOpen}
		<div class="menu-dropdown" role="menu">
			<button class="menu-item" role="menuitem" onclick={() => onAction('desktop-settings')}>
				<span class="menu-item-check">
					<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
						<circle cx="12" cy="12" r="3"/>
						<path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
					</svg>
				</span>
				<span class="menu-item-label">Desktop Settings...</span>
			</button>

			<div class="menu-separator" role="separator"></div>

			<button class="menu-item experimental-3d" role="menuitem" onclick={() => { desktopSettings.toggle3DDesktop(); onAction(''); }}>
				<span class="menu-item-check">
					{#if enable3D}
						<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<polyline points="20 6 9 17 4 12"/>
						</svg>
					{:else}
						<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
							<path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
							<polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
							<line x1="12" y1="22.08" x2="12" y2="12"/>
						</svg>
					{/if}
				</span>
				<span class="menu-item-label">
					{enable3D ? 'Exit 3D Desktop' : 'Experimental 3D Desktop'}
				</span>
				<span class="menu-item-shortcut beta-badge">Beta</span>
			</button>

			<div class="menu-separator" role="separator"></div>

			<button class="menu-item" role="menuitem" onclick={() => onAction('open-terminal')}>
				<span class="menu-item-check">
					<svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
						<path d="M4 17l6-6-6-6M12 19h8"/>
					</svg>
				</span>
				<span class="menu-item-label">Open Terminal</span>
			</button>

			<div class="menu-separator" role="separator"></div>

			<button class="menu-item" role="menuitem" onclick={() => onAction('exit-desktop')}>
				<span class="menu-item-check"></span>
				<span class="menu-item-label">Exit Desktop View</span>
			</button>
		</div>
	{/if}
</div>

<style>
	.menu-bar-item-wrapper {
		position: relative;
	}

	.menu-bar-logo {
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 4px 10px;
		border-radius: 4px;
		background: none;
		border: none;
		cursor: pointer;
		color: #333;
	}

	.menu-bar-logo:hover {
		background: rgba(0, 0, 0, 0.08);
	}

	:global(.dark) .menu-bar-logo {
		color: #f5f5f7;
	}

	:global(.dark) .menu-bar-logo:hover {
		background: rgba(255, 255, 255, 0.1);
	}
</style>
