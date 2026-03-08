<script lang="ts">
	interface Profile {
		id: string;
		name: string;
		icon?: string | null;
		type?: string | null;
	}

	interface Props {
		show: boolean;
		selectedProfileForSave: string | null;
		availableProfiles: Profile[];
		savingArtifactToProfile: boolean;
		onClose: () => void;
		onSelectProfile: (profileId: string) => void;
		onSave: () => void;
	}

	let {
		show,
		selectedProfileForSave,
		availableProfiles,
		savingArtifactToProfile,
		onClose,
		onSelectProfile,
		onSave,
	}: Props = $props();
</script>

{#if show}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<button
			class="absolute inset-0 bg-black/50"
			onclick={onClose}
			aria-label="Close modal"
		></button>

		<!-- Modal -->
		<div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md mx-4 overflow-hidden">
			<!-- Header -->
			<div class="p-4 border-b border-gray-100">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-semibold text-gray-900">Save Artifact to Profile</h3>
					<button
						onclick={onClose}
						class="btn-pill btn-pill-ghost btn-pill-icon btn-pill-sm"
						aria-label="Close modal"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
				<p class="text-sm text-gray-500 mt-1">Select a context profile to save this artifact as a document</p>
			</div>

			<!-- Content -->
			<div class="p-4 max-h-80 overflow-y-auto">
				<div class="space-y-2">
					<!-- Loose Documents option -->
					<button
						onclick={() => onSelectProfile('loose')}
						class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedProfileForSave === 'loose' ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
					>
						<div class="w-10 h-10 rounded-lg bg-gray-100 text-gray-600 flex items-center justify-center flex-shrink-0 text-lg">
							📄
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-gray-900">Loose Documents</p>
							<p class="text-xs text-gray-500">Save without a parent profile</p>
						</div>
						{#if selectedProfileForSave === 'loose'}
							<svg class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
								<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
						{/if}
					</button>

					{#if availableProfiles.length > 0}
						<div class="text-xs text-gray-400 uppercase tracking-wider font-medium pt-2 pb-1">Profiles</div>
						{#each availableProfiles as profile (profile.id)}
							<button
								onclick={() => onSelectProfile(profile.id)}
								class="w-full flex items-center gap-3 p-3 rounded-xl border-2 transition-colors text-left {selectedProfileForSave === profile.id ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'}"
							>
								<div class="w-10 h-10 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center flex-shrink-0 text-lg">
									{profile.icon ?? '📁'}
								</div>
								<div class="flex-1 min-w-0">
									<p class="text-sm font-medium text-gray-900">{profile.name}</p>
									{#if profile.type}
										<p class="text-xs text-gray-500 capitalize">{profile.type ?? ''}</p>
									{/if}
								</div>
								{#if selectedProfileForSave === profile.id}
									<svg class="w-5 h-5 text-blue-500" fill="currentColor" viewBox="0 0 24 24">
										<path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
								{/if}
							</button>
						{/each}
					{/if}
				</div>
			</div>

			<!-- Footer -->
			<div class="p-4 border-t border-gray-100 flex gap-3">
				<button
					onclick={onClose}
					class="btn-pill btn-pill-soft btn-pill-sm flex-1"
				>
					Cancel
				</button>
				<button
					onclick={onSave}
					disabled={!selectedProfileForSave || savingArtifactToProfile}
					class="btn-pill btn-pill-primary btn-pill-sm flex-1 {savingArtifactToProfile ? 'btn-pill-loading' : ''}"
				>
					Save to Profile
				</button>
			</div>
		</div>
	</div>
{/if}
