<script lang="ts">
	import type { CreateClientData, ClientType, ClientStatus } from '$lib/api';

	interface Props {
		open: boolean;
		onCreate: (data: CreateClientData) => void;
	}

	let { open = $bindable(), onCreate }: Props = $props();

	// Form state
	let name = $state('');
	let type = $state<ClientType>('company');
	let email = $state('');
	let phone = $state('');
	let website = $state('');
	let industry = $state('');
	let companySize = $state('');
	let status = $state<ClientStatus>('lead');
	let source = $state('');
	let notes = $state('');
	let tagsInput = $state('');
	let submitting = $state(false);

	// Address fields
	let address = $state('');
	let city = $state('');
	let stateRegion = $state('');
	let zipCode = $state('');
	let country = $state('');

	// Toggle for showing more fields
	let showMoreFields = $state(false);

	const companySizes = [
		{ value: '', label: 'Select size' },
		{ value: '1-10', label: '1-10 employees' },
		{ value: '11-50', label: '11-50 employees' },
		{ value: '51-200', label: '51-200 employees' },
		{ value: '201-500', label: '201-500 employees' },
		{ value: '501-1000', label: '501-1000 employees' },
		{ value: '1000+', label: '1000+ employees' }
	];

	const sources = [
		{ value: '', label: 'Select source' },
		{ value: 'website', label: 'Website' },
		{ value: 'referral', label: 'Referral' },
		{ value: 'cold-call', label: 'Cold Call' },
		{ value: 'linkedin', label: 'LinkedIn' },
		{ value: 'conference', label: 'Conference' },
		{ value: 'advertising', label: 'Advertising' },
		{ value: 'other', label: 'Other' }
	];

	function resetForm() {
		name = '';
		type = 'company';
		email = '';
		phone = '';
		website = '';
		industry = '';
		companySize = '';
		status = 'lead';
		source = '';
		notes = '';
		tagsInput = '';
		address = '';
		city = '';
		stateRegion = '';
		zipCode = '';
		country = '';
		showMoreFields = false;
	}

	function handleClose() {
		open = false;
		resetForm();
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (!name.trim()) return;

		submitting = true;

		try {
			const tags = tagsInput
				.split(',')
				.map((t) => t.trim())
				.filter((t) => t);

			const data: CreateClientData = {
				name: name.trim(),
				type,
				email: email.trim() || undefined,
				phone: phone.trim() || undefined,
				website: website.trim() || undefined,
				industry: industry.trim() || undefined,
				company_size: companySize || undefined,
				status,
				source: source || undefined,
				notes: notes.trim() || undefined,
				tags: tags.length > 0 ? tags : undefined,
				address: address.trim() || undefined,
				city: city.trim() || undefined,
				state: stateRegion.trim() || undefined,
				zip_code: zipCode.trim() || undefined,
				country: country.trim() || undefined
			};

			await onCreate(data);
			handleClose();
		} catch (err) {
			console.error('Failed to create client:', err);
		} finally {
			submitting = false;
		}
	}
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<div class="absolute inset-0 bg-black/50" onclick={handleClose} role="presentation"></div>

		<!-- Modal -->
		<div class="relative bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-hidden flex flex-col">
			<!-- Header -->
			<div class="px-6 py-4 border-b border-gray-200">
				<h2 class="text-lg font-semibold text-gray-900">Add New Client</h2>
				<p class="text-sm text-gray-500">Create a new client record</p>
			</div>

			<!-- Form -->
			<form onsubmit={handleSubmit} class="flex-1 overflow-auto">
				<div class="px-6 py-4 space-y-4">
					<!-- Type Toggle -->
					<div class="btn-pill-group w-full">
						<button
							type="button"
							onclick={() => (type = 'company')}
							class="btn-pill btn-pill-sm flex-1 {type === 'company'
								? 'btn-pill-primary'
								: 'btn-pill-ghost'}"
						>
							Company
						</button>
						<button
							type="button"
							onclick={() => (type = 'individual')}
							class="btn-pill btn-pill-sm flex-1 {type === 'individual'
								? 'btn-pill-primary'
								: 'btn-pill-ghost'}"
						>
							Individual
						</button>
					</div>

					<!-- Name -->
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">
							{type === 'company' ? 'Company Name' : 'Full Name'} *
						</label>
						<input
							type="text"
							bind:value={name}
							required
							class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
							placeholder={type === 'company' ? 'Acme Inc.' : 'John Doe'}
						/>
					</div>

					<!-- Email & Phone -->
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
							<input
								type="email"
								bind:value={email}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
								placeholder="email@example.com"
							/>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
							<input
								type="tel"
								bind:value={phone}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
								placeholder="+1 (555) 123-4567"
							/>
						</div>
					</div>

					<!-- Status & Source -->
					<div class="grid grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
							<select
								bind:value={status}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
							>
								<option value="lead">Lead</option>
								<option value="prospect">Prospect</option>
								<option value="active">Active</option>
								<option value="inactive">Inactive</option>
							</select>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Source</label>
							<select
								bind:value={source}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
							>
								{#each sources as s}
									<option value={s.value}>{s.label}</option>
								{/each}
							</select>
						</div>
					</div>

					<!-- Company-specific fields -->
					{#if type === 'company'}
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Industry</label>
								<input
									type="text"
									bind:value={industry}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
									placeholder="Technology"
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Company Size</label>
								<select
									bind:value={companySize}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
								>
									{#each companySizes as size}
										<option value={size.value}>{size.label}</option>
									{/each}
								</select>
							</div>
						</div>

						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Website</label>
							<input
								type="url"
								bind:value={website}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
								placeholder="https://example.com"
							/>
						</div>
					{/if}

					<!-- Tags -->
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Tags</label>
						<input
							type="text"
							bind:value={tagsInput}
							class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
							placeholder="vip, enterprise, tech (comma separated)"
						/>
					</div>

					<!-- Show More Toggle -->
					<button
						type="button"
						onclick={() => (showMoreFields = !showMoreFields)}
						class="text-sm text-gray-600 hover:text-gray-900 flex items-center gap-1"
					>
						<svg
							class="w-4 h-4 transition-transform {showMoreFields ? 'rotate-180' : ''}"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M19 9l-7 7-7-7"
							/>
						</svg>
						{showMoreFields ? 'Show less' : 'Show more fields'}
					</button>

					{#if showMoreFields}
						<!-- Address -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
							<input
								type="text"
								bind:value={address}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
								placeholder="123 Main St"
							/>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">City</label>
								<input
									type="text"
									bind:value={city}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
									placeholder="San Francisco"
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">State/Region</label>
								<input
									type="text"
									bind:value={stateRegion}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
									placeholder="CA"
								/>
							</div>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">ZIP/Postal Code</label>
								<input
									type="text"
									bind:value={zipCode}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
									placeholder="94102"
								/>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 mb-1">Country</label>
								<input
									type="text"
									bind:value={country}
									class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900"
									placeholder="United States"
								/>
							</div>
						</div>

						<!-- Notes -->
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
							<textarea
								bind:value={notes}
								rows="3"
								class="w-full px-3 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-gray-900 resize-none"
								placeholder="Additional notes about this client..."
							></textarea>
						</div>
					{/if}
				</div>

				<!-- Footer -->
				<div class="px-6 py-4 border-t border-gray-200 flex justify-end gap-3">
					<button
						type="button"
						onclick={handleClose}
						class="btn-pill btn-pill-ghost"
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={submitting || !name.trim()}
						class="btn-pill btn-pill-primary {submitting ? 'btn-pill-loading' : ''} disabled:opacity-50 disabled:cursor-not-allowed"
					>
						{submitting ? 'Creating...' : 'Create Client'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
