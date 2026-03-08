<script lang="ts">
	import { Dialog, DropdownMenu } from 'bits-ui';

	interface Manager {
		id: string;
		name: string;
	}

	interface EditMember {
		id: string;
		name: string;
		email: string;
		role: string;
		managerId?: string;
		skills: string[];
		hourlyRate?: number;
	}

	interface Props {
		open?: boolean;
		managers?: Manager[];
		editMember?: EditMember | null;
		onClose?: () => void;
		onCreate?: (member: {
			name: string;
			email: string;
			role: string;
			managerId?: string;
			skills: string[];
			hourlyRate?: number;
		}) => void;
		onUpdate?: (id: string, member: {
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
		editMember = null,
		onClose,
		onCreate,
		onUpdate
	}: Props = $props();

	const isEditMode = $derived(!!editMember);

	let name = $state('');
	let email = $state('');
	let role = $state('');
	let managerId = $state<string>('');
	let skillInput = $state('');
	let skills = $state<string[]>([]);
	let hourlyRate = $state<string>('');

	// Populate form when editing
	$effect(() => {
		if (editMember && open) {
			name = editMember.name;
			email = editMember.email;
			role = editMember.role;
			managerId = editMember.managerId || '';
			skills = [...(editMember.skills || [])];
			hourlyRate = editMember.hourlyRate ? String(editMember.hourlyRate) : '';
		}
	});

	function handleSubmit() {
		if (!name.trim() || !email.trim() || !role.trim()) return;

		const data = {
			name,
			email,
			role,
			managerId: managerId || undefined,
			skills,
			hourlyRate: hourlyRate ? parseFloat(hourlyRate) : undefined
		};

		if (isEditMode && editMember) {
			onUpdate?.(editMember.id, data);
		} else {
			onCreate?.(data);
		}

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
		<Dialog.Overlay class="td-modal-overlay" />
		<Dialog.Content class="td-modal">
			<!-- Header -->
			<div class="td-modal__header">
				<Dialog.Title class="td-modal__title">{isEditMode ? 'Edit Team Member' : 'Add Team Member'}</Dialog.Title>
				<Dialog.Close
					class="td-modal__close"
					onclick={handleClose}
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
				</Dialog.Close>
			</div>

			<!-- Body -->
			<div class="td-modal__body">
				<!-- Full Name -->
				<div class="td-input-group">
					<label for="member-name" class="td-label">
						Full name <span class="td-label__req">*</span>
					</label>
					<input
						id="member-name"
						type="text"
						bind:value={name}
						placeholder="e.g., John Smith"
						class="td-input"
					/>
				</div>

				<!-- Email -->
				<div class="td-input-group">
					<label for="member-email" class="td-label">
						Email <span class="td-label__req">*</span>
					</label>
					<input
						id="member-email"
						type="email"
						bind:value={email}
						placeholder="john@company.com"
						class="td-input"
					/>
				</div>

				<!-- Role / Title -->
				<div class="td-input-group">
					<label for="member-role" class="td-label">
						Role / Title <span class="td-label__req">*</span>
					</label>
					<input
						id="member-role"
						type="text"
						bind:value={role}
						placeholder="e.g., Frontend Developer"
						class="td-input"
					/>
				</div>

				<!-- Reports To -->
				<div class="td-input-group">
					<label for="reports-to" class="td-label">Reports to</label>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger id="reports-to" class="td-select-trigger">
							{#if selectedManager}
								<span>{selectedManager.name}</span>
							{:else}
								<span style="color: var(--dt4, #bbb)">Select manager...</span>
							{/if}
							<svg class="w-4 h-4" style="color: var(--dt4, #bbb)" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
							</svg>
						</DropdownMenu.Trigger>
						<DropdownMenu.Portal>
							<DropdownMenu.Content class="td-dropdown" sideOffset={4}>
								<DropdownMenu.Item
									class="td-dropdown__item td-dropdown__item--muted"
									onclick={() => managerId = ''}
								>
									No manager
								</DropdownMenu.Item>
								{#each managers as manager}
									<DropdownMenu.Item
										class="td-dropdown__item"
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
				<div class="td-input-group">
					<label for="skill-input" class="td-label">Skills (optional)</label>
					<div class="td-skill-input-wrap">
						{#each skills as skill}
							<span class="td-skill-tag td-skill-tag--removable">
								{skill}
								<button onclick={() => removeSkill(skill)} class="td-skill-tag__remove" aria-label="Remove {skill}">
									<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
									</svg>
								</button>
							</span>
						{/each}
						<input
							id="skill-input"
							type="text"
							bind:value={skillInput}
							onkeydown={handleSkillKeydown}
							placeholder={skills.length === 0 ? '+ Add skills...' : ''}
							class="td-skill-input-wrap__input"
						/>
					</div>
				</div>

				<!-- Hourly Rate -->
				<div class="td-input-group">
					<label for="hourly-rate" class="td-label">
						Hourly rate (optional)
					</label>
					<div class="td-input-prefix-wrap">
						<span class="td-input-prefix-wrap__prefix">$</span>
						<input
							id="hourly-rate"
							type="number"
							bind:value={hourlyRate}
							placeholder="0.00"
							class="td-input td-input--prefix"
						/>
					</div>
				</div>
			</div>

			<!-- Footer -->
			<div class="td-modal__footer">
				<button
					onclick={handleClose}
					class="btn-pill btn-pill-ghost btn-pill-sm"
				>
					Cancel
				</button>
				<button
					onclick={handleSubmit}
					disabled={!name.trim() || !email.trim() || !role.trim()}
					class="btn-pill btn-pill-primary btn-pill-sm"
				>
					{isEditMode ? 'Update Member' : 'Add Member'}
				</button>
			</div>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.td-modal-overlay) {
		position: fixed;
		inset: 0;
		background: rgba(0,0,0,0.45);
		z-index: 50;
	}

	:global(.td-modal) {
		position: fixed;
		left: 50%;
		top: 50%;
		transform: translate(-50%, -50%);
		z-index: 50;
		width: 100%;
		max-width: 480px;
		background: var(--dbg, #fff);
		border-radius: 16px;
		box-shadow: 0 20px 60px rgba(0,0,0,0.18);
		border: 1px solid var(--dbd, #e0e0e0);
	}

	.td-modal__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px;
		border-bottom: 1px solid var(--dbd, #e0e0e0);
	}

	:global(.td-modal__title) {
		font-size: 15px;
		font-weight: 700;
		color: var(--dt, #111);
		letter-spacing: -0.01em;
	}

	:global(.td-modal__close) {
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 8px;
		color: var(--dt3, #888);
		background: transparent;
		border: none;
		cursor: pointer;
		transition: background 0.12s;
	}
	:global(.td-modal__close:hover) { background: var(--dbg2, #f5f5f5); }

	.td-modal__body {
		padding: 16px 20px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		max-height: 60vh;
		overflow-y: auto;
	}

	.td-modal__footer {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: 10px;
		padding: 14px 20px;
		border-top: 1px solid var(--dbd, #e0e0e0);
	}

	.td-input-group {
		display: flex;
		flex-direction: column;
	}

	.td-label {
		font-size: 12px;
		font-weight: 600;
		color: var(--dt2, #555);
		margin-bottom: 5px;
	}
	.td-label__req { color: #ef4444; }

	.td-input {
		width: 100%;
		padding: 9px 14px;
		font-size: 13px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		outline: none;
		transition: border-color 0.15s, box-shadow 0.15s;
	}
	.td-input::placeholder { color: var(--dt4, #bbb); }
	.td-input:focus {
		border-color: var(--dt, #111);
		box-shadow: 0 0 0 2px rgba(0,0,0,0.06);
	}
	.td-input--prefix { padding-left: 30px; }

	.td-input-prefix-wrap {
		position: relative;
	}
	.td-input-prefix-wrap__prefix {
		position: absolute;
		left: 14px;
		top: 50%;
		transform: translateY(-50%);
		color: var(--dt4, #bbb);
		font-size: 13px;
		pointer-events: none;
	}

	:global(.td-select-trigger) {
		width: 100%;
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 9px 14px;
		font-size: 13px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		background: var(--dbg, #fff);
		color: var(--dt, #111);
		cursor: pointer;
		text-align: left;
		transition: background 0.12s;
	}
	:global(.td-select-trigger:hover) { background: var(--dbg2, #f5f5f5); }

	:global(.td-dropdown) {
		z-index: 60;
		min-width: 200px;
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 12px;
		box-shadow: 0 8px 24px rgba(0,0,0,0.12);
		padding: 4px;
	}

	:global(.td-dropdown__item) {
		padding: 8px 12px;
		font-size: 13px;
		color: var(--dt, #111);
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.1s;
	}
	:global(.td-dropdown__item:hover) { background: var(--dbg2, #f5f5f5); }
	:global(.td-dropdown__item--muted) { color: var(--dt3, #888); }

	.td-skill-input-wrap {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 8px;
		border: 1px solid var(--dbd, #e0e0e0);
		border-radius: 10px;
		min-height: 42px;
		align-items: center;
	}
	.td-skill-input-wrap__input {
		flex: 1;
		min-width: 100px;
		padding: 4px 6px;
		font-size: 13px;
		border: none;
		background: transparent;
		color: var(--dt, #111);
		outline: none;
	}
	.td-skill-input-wrap__input::placeholder { color: var(--dt4, #bbb); }

	.td-skill-tag {
		display: inline-flex;
		align-items: center;
		height: 22px;
		padding: 0 10px;
		border-radius: 9999px;
		border: 1px solid var(--dbd2, #f0f0f0);
		background: var(--dbg2, #f5f5f5);
		font-size: 11px;
		font-weight: 600;
		color: var(--dt2, #555);
	}
	.td-skill-tag--removable { gap: 4px; padding-right: 6px; }
	.td-skill-tag__remove {
		display: flex;
		align-items: center;
		background: none;
		border: none;
		cursor: pointer;
		color: var(--dt3, #888);
		padding: 0;
	}
	.td-skill-tag__remove:hover { color: var(--dt, #111); }
</style>
