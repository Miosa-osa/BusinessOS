<script lang="ts">
	import { Dialog, DropdownMenu } from 'bits-ui';

	interface Manager {
		id: string;
		name: string;
	}

	interface Props {
		open?: boolean;
		managers?: Manager[];
		onClose?: () => void;
		onCreate?: (member: {
			name: string;
			email: string;
			role: string;
			managerId?: string;
			skills: string[];
			hourlyRate?: number;
		}) => void;
	}

	let {
		open = $bindable(false),
		managers = [],
		onClose,
		onCreate
	}: Props = $props();

	let name = $state('');
	let email = $state('');
	let role = $state('');
	let managerId = $state<string>('');
	let skillInput = $state('');
	let skills = $state<string[]>([]);
	let hourlyRate = $state<string>('');

	function handleSubmit() {
		if (!name.trim() || !email.trim() || !role.trim()) return;

		onCreate?.({
			name,
			email,
			role,
			managerId: managerId || undefined,
			skills,
			hourlyRate: hourlyRate ? parseFloat(hourlyRate) : undefined
		});

		resetForm();
		open = false;
	}

	function resetForm() {
		name = '';
		email = '';
		role = '';
		managerId = '';
		skillInput = '';
		skills = [];
		hourlyRate = '';
	}

	function handleClose() {
		resetForm();
		open = false;
		onClose?.();
	}

	function addSkill() {
		if (skillInput.trim() && !skills.includes(skillInput.trim())) {
			skills = [...skills, skillInput.trim()];
			skillInput = '';
		}
	}

	function removeSkill(skill: string) {
		skills = skills.filter(s => s !== skill);
	}

	function handleSkillKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			addSkill();
		}
	}

	const selectedManager = $derived(managers.find(m => m.id === managerId));
</script>

<Dialog.Root bind:open>
	<Dialog.Portal>
		<Dialog.Overlay
			class="fixed inset-0 bg-black/50 z-50 animate-in fade-in-0"
		/>
		<Dialog.Content
			class="fixed left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 z-50 w-full max-w-lg bg-white rounded-2xl shadow-xl animate-in fade-in-0 zoom-in-95"
		>
			<!-- Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-100">
				<Dialog.Title class="text-lg font-semibold text-gray-900">Add Team Member</Dialog.Title>
				<Dialog.Close
					class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 transition-colors"
					onclick={handleClose}
				>
					<svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</Dialog.Close>
			</div>

			<!-- Body -->
			<div class="px-6 py-4 space-y-4 max-h-[60vh] overflow-y-auto">
				<!-- Full Name -->
				<div>
					<label for="member-name" class="block text-sm font-medium text-gray-700 mb-1">
						Full name <span class="text-red-500">*</span>
					</label>
					<input
						id="member-name"
						type="text"
						bind:value={name}
						placeholder="e.g., John Smith"
						class="w-full px-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
					/>
				</div>

				<!-- Email -->
				<div>
					<label for="member-email" class="block text-sm font-medium text-gray-700 mb-1">
						Email <span class="text-red-500">*</span>
					</label>
					<input
						id="member-email"
						type="email"
						bind:value={email}
						placeholder="john@company.com"
						class="w-full px-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
					/>
				</div>

				<!-- Role / Title -->
				<div>
					<label for="member-role" class="block text-sm font-medium text-gray-700 mb-1">
						Role / Title <span class="text-red-500">*</span>
					</label>
					<input
						id="member-role"
						type="text"
						bind:value={role}
						placeholder="e.g., Frontend Developer"
						class="w-full px-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
					/>
				</div>

				<!-- Reports To -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Reports to</label>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger
							class="w-full flex items-center justify-between px-4 py-2.5 text-sm border border-gray-200 rounded-xl hover:bg-gray-50 transition-colors text-left"
						>
							{#if selectedManager}
								<span>{selectedManager.name}</span>
							{:else}
								<span class="text-gray-400">Select manager...</span>
							{/if}
							<svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content
								class="z-[60] min-w-[200px] bg-white border border-gray-200 rounded-xl shadow-lg p-1 animate-in fade-in-0 zoom-in-95"
								sideOffset={4}
							>
								<DropdownMenu.Item
									class="px-3 py-2 text-sm text-gray-500 hover:bg-gray-100 rounded-lg cursor-pointer"
									onclick={() => managerId = ''}
								>
									No manager
								</DropdownMenu.Item>
								{#each managers as manager}
									<DropdownMenu.Item
										class="px-3 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg cursor-pointer"
										onclick={() => managerId = manager.id}
									>
										{manager.name}
									</DropdownMenu.Item>
								{/each}
							</DropdownMenu.Content>
						</DropdownMenu.Portal>
					</DropdownMenu.Root>
				</div>

				<!-- Skills -->
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Skills (optional)</label>
					<div class="flex flex-wrap gap-2 p-2 border border-gray-200 rounded-xl min-h-[44px]">
						{#each skills as skill}
							<span class="flex items-center gap-1 px-2 py-1 bg-gray-100 text-gray-700 text-sm rounded-lg">
								{skill}
								<button onclick={() => removeSkill(skill)} class="hover:text-gray-900">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</button>
							</span>
						{/each}
						<input
							type="text"
							bind:value={skillInput}
							onkeydown={handleSkillKeydown}
							placeholder={skills.length === 0 ? '+ Add skills...' : ''}
							class="flex-1 min-w-[100px] px-2 py-1 text-sm focus:outline-none"
						/>
					</div>
				</div>

				<!-- Hourly Rate -->
				<div>
					<label for="hourly-rate" class="block text-sm font-medium text-gray-700 mb-1">
						Hourly rate (optional)
					</label>
					<div class="relative">
						<span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">$</span>
						<input
							id="hourly-rate"
							type="number"
							bind:value={hourlyRate}
							placeholder="0.00"
							class="w-full pl-8 pr-4 py-2.5 text-sm border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-gray-900 focus:border-transparent transition-all"
						/>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100">
				<button
					onclick={handleClose}
					class="px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 rounded-lg transition-colors"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={!name.trim() || !email.trim() || !role.trim()}
					class="px-4 py-2 text-sm font-medium text-white bg-gray-900 hover:bg-gray-800 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
				>
					Add Member
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
