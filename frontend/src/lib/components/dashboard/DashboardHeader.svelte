<script lang="ts">
	import { fly, fade } from 'svelte/transition';
	import { NotificationDropdown } from '$lib/components/notifications';

	interface Props {
		userName: string;
		energyLevel?: number | null;
		onEnergySet?: (level: number) => void;
	}

	let { userName, energyLevel = null, onEnergySet }: Props = $props();

	let showEnergyCheck = $state(energyLevel === null);
	let sliderValue = $state(energyLevel ?? 5);

	// Dynamic greeting based on time
	const getGreeting = () => {
		const hour = new Date().getHours();
		if (hour >= 5 && hour < 12) return 'Good morning';
		if (hour >= 12 && hour < 17) return 'Good afternoon';
		if (hour >= 17 && hour < 21) return 'Good evening';
		return 'Working late';
	};

	// Format date
	const formatDate = () => {
		return new Date().toLocaleDateString('en-US', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	};

	// Energy level labels
	const getEnergyLabel = (level: number): string => {
		if (level <= 2) return 'Low';
		if (level <= 4) return 'Below Average';
		if (level <= 6) return 'Average';
		if (level <= 8) return 'Good';
		return 'Excellent';
	};

	// Energy level colors
	const getEnergyColor = (level: number): string => {
		if (level <= 2) return 'bg-red-500';
		if (level <= 4) return 'bg-orange-500';
		if (level <= 6) return 'bg-yellow-500';
		if (level <= 8) return 'bg-lime-500';
		return 'bg-green-500';
	};

	function handleSliderChange(event: Event) {
		const target = event.target as HTMLInputElement;
		sliderValue = parseInt(target.value, 10);
	}

	function handleEnergySubmit() {
		onEnergySet?.(sliderValue);
		showEnergyCheck = false;
	}

	function dismissEnergyCheck() {
		showEnergyCheck = false;
	}
</script>

<header class="px-6 py-6" in:fade={{ duration: 300 }}>
	<div class="flex items-start justify-between">
		<div>
			<h1
				class="text-2xl font-semibold text-gray-900 dark:text-white"
				in:fly={{ y: -10, duration: 400, delay: 100 }}
			>
				{getGreeting()}, {userName}
			</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1" in:fly={{ y: -10, duration: 400, delay: 200 }}>
				{formatDate()}
			</p>
		</div>
		<div in:fly={{ x: 10, duration: 400, delay: 150 }}>
			<NotificationDropdown />
		</div>
	</div>

	{#if showEnergyCheck}
		<div
			class="mt-6 p-4 bg-gray-50 rounded-xl border border-gray-200"
			in:fly={{ y: 10, duration: 400, delay: 300 }}
		>
			<div class="flex items-center justify-between mb-4">
				<p class="text-sm font-medium text-gray-700">How's your energy today?</p>
				<button
					onclick={dismissEnergyCheck}
					class="text-gray-400 hover:text-gray-600 transition-colors"
					aria-label="Dismiss"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</button>
			</div>

			<!-- Energy Slider -->
			<div class="space-y-3">
				<div class="flex items-center gap-4">
					<span class="text-xs text-gray-500 w-8">1</span>
					<input
						type="range"
						min="1"
						max="10"
						value={sliderValue}
						oninput={handleSliderChange}
						class="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer accent-gray-900"
					/>
					<span class="text-xs text-gray-500 w-8 text-right">10</span>
				</div>

				<div class="flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="w-3 h-3 rounded-full {getEnergyColor(sliderValue)}"></div>
						<span class="text-sm font-medium text-gray-900">{sliderValue}</span>
						<span class="text-sm text-gray-500">- {getEnergyLabel(sliderValue)}</span>
					</div>
					<button
						onclick={handleEnergySubmit}
						class="px-3 py-1.5 bg-gray-900 text-white text-sm font-medium rounded-lg hover:bg-gray-800 transition-colors"
					>
						Log Energy
					</button>
				</div>
			</div>
		</div>
	{/if}
</header>
