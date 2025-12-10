<script lang="ts">
	import { goto } from '$app/navigation';
	import { fly, fade, scale } from 'svelte/transition';
	import { useSession } from '$lib/auth-client';
	import { OnboardingLayout, SelectionCard, MultiSelectCard } from '$lib/components/onboarding';

	const session = useSession();

	// Get user's first name
	const firstName = $derived($session.data?.user?.name?.split(' ')[0] || 'there');

	// Onboarding state
	let currentStep = $state(1);
	let direction = $state(1); // 1 = forward, -1 = back
	const totalSteps = 6;

	// Form data
	let businessType = $state('');
	let businessName = $state('');
	let role = $state('');
	let teamSize = $state('');
	let useCases = $state<string[]>([]);
	let workspaceName = $state('');
	let theme = $state('light');

	// SVG Icons as strings
	const icons = {
		building: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" /></svg>',
		rocket: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>',
		briefcase: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>',
		academic: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 14l9-5-9-5-9 5 9 5zm0 0l6.16-3.422a12.083 12.083 0 01.665 6.479A11.952 11.952 0 0012 20.055a11.952 11.952 0 00-6.824-2.998 12.078 12.078 0 01.665-6.479L12 14zm-4 6v-7.5l4-2.222" /></svg>',
		cube: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" /></svg>',
		sparkles: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" /></svg>',
		// Use case icons
		clipboard: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" /></svg>',
		users: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>',
		handshake: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" /></svg>',
		check: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>',
		chat: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" /></svg>',
		document: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>',
		book: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" /></svg>',
		clock: '<svg class="w-5 h-5 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>'
	};

	// Business type options with SVG icons
	const businessTypes = [
		{ icon: icons.building, label: 'Agency / Consultancy', value: 'agency' },
		{ icon: icons.rocket, label: 'Startup / Tech Company', value: 'startup' },
		{ icon: icons.briefcase, label: 'Freelancer / Solo', value: 'freelancer' },
		{ icon: icons.academic, label: 'Education / Coaching', value: 'education' },
		{ icon: icons.cube, label: 'E-commerce / Product', value: 'ecommerce' },
		{ icon: icons.sparkles, label: 'Other', value: 'other' }
	];

	// Role options
	const roles = [
		{ label: 'CEO / Founder', value: 'ceo_founder' },
		{ label: 'COO / Operations', value: 'coo_operations' },
		{ label: 'CTO / Technical', value: 'cto_technical' },
		{ label: 'Manager / Lead', value: 'manager_lead' },
		{ label: 'Team Member', value: 'team_member' },
		{ label: 'Other', value: 'other' }
	];

	// Team size options
	const teamSizes = [
		{ label: 'Just me', value: 'solo' },
		{ label: '2-5 people', value: '2-5' },
		{ label: '6-15 people', value: '6-15' },
		{ label: '16-50 people', value: '16-50' },
		{ label: '50+ people', value: '50+' }
	];

	// Use case options with SVG icons
	const useCaseOptions = [
		{ icon: icons.clipboard, label: 'Project Management', value: 'projects' },
		{ icon: icons.users, label: 'Client Management', value: 'clients' },
		{ icon: icons.handshake, label: 'Team Coordination', value: 'team' },
		{ icon: icons.check, label: 'Task Tracking', value: 'tasks' },
		{ icon: icons.chat, label: 'AI Assistant / Chat', value: 'ai_chat' },
		{ icon: icons.document, label: 'Document Creation', value: 'documents' },
		{ icon: icons.book, label: 'Knowledge Base', value: 'knowledge' },
		{ icon: icons.clock, label: 'Time Tracking', value: 'time' }
	];

	function goNext() {
		direction = 1;
		if (currentStep < totalSteps) {
			currentStep++;
		}
	}

	function goBack() {
		direction = -1;
		if (currentStep > 1) {
			currentStep--;
		}
	}

	function toggleUseCase(value: string) {
		if (useCases.includes(value)) {
			useCases = useCases.filter(v => v !== value);
		} else {
			useCases = [...useCases, value];
		}
	}

	async function skipOnboarding() {
		// Set defaults and redirect
		await saveOnboardingData({
			business_type: 'other',
			business_name: `${firstName}'s Workspace`,
			role: 'other',
			team_size: 'solo',
			use_cases: ['projects', 'tasks', 'ai_chat'],
			workspace_name: `${firstName}'s Workspace`,
			theme: 'light'
		});
		goto('/chat');
	}

	async function completeOnboarding() {
		await saveOnboardingData({
			business_type: businessType,
			business_name: businessName,
			role,
			team_size: teamSize,
			use_cases: useCases,
			workspace_name: workspaceName || `${businessName} HQ`,
			theme
		});
		goNext();
	}

	async function saveOnboardingData(data: Record<string, any>) {
		// Store in localStorage for now
		localStorage.setItem('onboarding_data', JSON.stringify(data));
		localStorage.setItem('onboarding_completed', 'true');

		// TODO: Also save to backend
		try {
			const response = await fetch('/api/onboarding', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(data)
			});
			if (!response.ok) {
				console.error('Failed to save onboarding data to backend');
			}
		} catch (error) {
			console.error('Error saving onboarding data:', error);
		}
	}

	function goToDashboard() {
		goto('/chat');
	}

	// Auto-fill workspace name when business name changes
	$effect(() => {
		if (businessName && !workspaceName) {
			workspaceName = `${businessName} HQ`;
		}
	});

	// Validation
	const canContinue = $derived(() => {
		switch (currentStep) {
			case 1: return true;
			case 2: return !!businessType;
			case 3: return !!businessName && !!role && !!teamSize;
			case 4: return useCases.length > 0;
			case 5: return !!workspaceName;
			default: return true;
		}
	});
</script>

{#if currentStep < totalSteps}
	<OnboardingLayout
		{currentStep}
		{totalSteps}
		showBack={currentStep > 1}
		showSkip={currentStep < totalSteps - 1}
		continueText={currentStep === 5 ? 'Create workspace' : currentStep === 1 ? "Let's go" : 'Continue'}
		continueDisabled={!canContinue()}
		onBack={goBack}
		onContinue={currentStep === 5 ? completeOnboarding : goNext}
		onSkip={skipOnboarding}
	>
		{#key currentStep}
			<div
				in:fly={{ x: direction * 100, duration: 300, delay: 150 }}
				out:fly={{ x: direction * -100, duration: 300 }}
			>
				<!-- Step 1: Welcome -->
				{#if currentStep === 1}
					<div class="text-center">
						<div class="w-20 h-20 bg-gray-100 rounded-2xl flex items-center justify-center mx-auto mb-6" in:scale={{ duration: 400, start: 0.5 }}>
							<svg class="w-10 h-10 text-gray-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11" />
							</svg>
						</div>
						<h1 class="text-3xl font-bold text-gray-900 mb-4">
							Welcome to Business OS, {firstName}!
						</h1>
						<p class="text-gray-600 text-lg max-w-md mx-auto">
							Let's set up your workspace in about 2 minutes. We'll ask a few questions to personalize your experience.
						</p>
					</div>

				<!-- Step 2: Business Type -->
				{:else if currentStep === 2}
					<div>
						<h1 class="text-2xl font-bold text-gray-900 mb-2 text-center">
							What type of business do you run?
						</h1>
						<p class="text-gray-500 text-center mb-8">
							This helps us customize your experience
						</p>
						<div class="space-y-3">
							{#each businessTypes as type, i}
								<div in:fly={{ y: 20, duration: 300, delay: i * 50 }}>
									<SelectionCard
										icon={type.icon}
										label={type.label}
										selected={businessType === type.value}
										onclick={() => businessType = type.value}
									/>
								</div>
							{/each}
						</div>
					</div>

				<!-- Step 3: Business Info -->
				{:else if currentStep === 3}
					<div>
						<h1 class="text-2xl font-bold text-gray-900 mb-2 text-center">
							Tell us about your business
						</h1>
						<p class="text-gray-500 text-center mb-8">
							Just the basics to get started
						</p>
						<div class="space-y-6">
							<div in:fly={{ y: 20, duration: 300, delay: 0 }}>
								<label for="businessName" class="block text-sm font-medium text-gray-700 mb-1.5">
									Business name
								</label>
								<input
									id="businessName"
									type="text"
									bind:value={businessName}
									placeholder="Your company name"
									class="input input-square w-full"
								/>
							</div>

							<div in:fly={{ y: 20, duration: 300, delay: 100 }}>
								<label for="role" class="block text-sm font-medium text-gray-700 mb-1.5">
									Your role
								</label>
								<select
									id="role"
									bind:value={role}
									class="input input-square w-full"
								>
									<option value="">Select your role</option>
									{#each roles as r}
										<option value={r.value}>{r.label}</option>
									{/each}
								</select>
							</div>

							<div in:fly={{ y: 20, duration: 300, delay: 200 }}>
								<label class="block text-sm font-medium text-gray-700 mb-3">
									Team size
								</label>
								<div class="space-y-2">
									{#each teamSizes as size}
										<label class="flex items-center gap-3 cursor-pointer p-3 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors h-[48px]
											{teamSize === size.value ? 'border-gray-900 bg-gray-50' : ''}">
											<input
												type="radio"
												name="teamSize"
												value={size.value}
												bind:group={teamSize}
												class="w-4 h-4 text-gray-900 border-gray-300 focus:ring-gray-900"
											/>
											<span class="text-gray-900">{size.label}</span>
										</label>
									{/each}
								</div>
							</div>
						</div>
					</div>

				<!-- Step 4: Use Cases -->
				{:else if currentStep === 4}
					<div>
						<h1 class="text-2xl font-bold text-gray-900 mb-2 text-center">
							What will you mainly use this for?
						</h1>
						<p class="text-gray-500 text-center mb-8">
							Select all that apply
						</p>
						<div class="grid grid-cols-2 gap-3">
							{#each useCaseOptions as useCase, i}
								<div in:fly={{ y: 20, duration: 300, delay: i * 50 }}>
									<MultiSelectCard
										icon={useCase.icon}
										label={useCase.label}
										selected={useCases.includes(useCase.value)}
										onclick={() => toggleUseCase(useCase.value)}
									/>
								</div>
							{/each}
						</div>
					</div>

				<!-- Step 5: Workspace Setup -->
				{:else if currentStep === 5}
					<div>
						<h1 class="text-2xl font-bold text-gray-900 mb-2 text-center">
							Let's set up your workspace
						</h1>
						<p class="text-gray-500 text-center mb-8">
							Almost done!
						</p>
						<div class="space-y-8">
							<div in:fly={{ y: 20, duration: 300, delay: 0 }}>
								<label for="workspaceName" class="block text-sm font-medium text-gray-700 mb-1.5">
									Workspace name
								</label>
								<input
									id="workspaceName"
									type="text"
									bind:value={workspaceName}
									placeholder="{businessName || 'Your'} HQ"
									class="input input-square w-full"
								/>
								<p class="text-sm text-gray-500 mt-1.5">
									This is what your team will see
								</p>
							</div>

							<div in:fly={{ y: 20, duration: 300, delay: 100 }}>
								<label class="block text-sm font-medium text-gray-700 mb-3">
									Theme
								</label>
								<div class="grid grid-cols-2 gap-4">
									<button
										type="button"
										onclick={() => theme = 'light'}
										class="p-4 rounded-xl border-2 transition-all h-[140px]
											{theme === 'light' ? 'border-gray-900' : 'border-gray-200 hover:border-gray-300'}"
									>
										<div class="h-16 bg-white border border-gray-200 rounded-lg mb-3 flex items-center justify-center">
											<div class="w-12 h-2 bg-gray-200 rounded"></div>
										</div>
										<div class="flex items-center justify-center gap-2">
											<span class="text-sm font-medium">Light</span>
											<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center
												{theme === 'light' ? 'border-gray-900' : 'border-gray-300'}">
												{#if theme === 'light'}
													<div class="w-2 h-2 rounded-full bg-gray-900"></div>
												{/if}
											</div>
										</div>
									</button>

									<button
										type="button"
										onclick={() => theme = 'dark'}
										class="p-4 rounded-xl border-2 transition-all h-[140px]
											{theme === 'dark' ? 'border-gray-900' : 'border-gray-200 hover:border-gray-300'}"
									>
										<div class="h-16 bg-gray-900 rounded-lg mb-3 flex items-center justify-center">
											<div class="w-12 h-2 bg-gray-700 rounded"></div>
										</div>
										<div class="flex items-center justify-center gap-2">
											<span class="text-sm font-medium">Dark</span>
											<div class="w-4 h-4 rounded-full border-2 flex items-center justify-center
												{theme === 'dark' ? 'border-gray-900' : 'border-gray-300'}">
												{#if theme === 'dark'}
													<div class="w-2 h-2 rounded-full bg-gray-900"></div>
												{/if}
											</div>
										</div>
									</button>
								</div>
							</div>
						</div>
					</div>
				{/if}
			</div>
		{/key}
	</OnboardingLayout>

{:else}
	<!-- Step 6: Completion -->
	<div class="min-h-screen bg-white flex flex-col items-center justify-center px-6">
		<div class="text-center max-w-md" in:fade={{ duration: 300 }}>
			<!-- Animated Checkmark -->
			<div class="mb-8" in:scale={{ duration: 400, start: 0.5 }}>
				<svg class="w-24 h-24 mx-auto" viewBox="0 0 52 52">
					<circle
						class="text-green-100"
						cx="26"
						cy="26"
						r="25"
						fill="currentColor"
					/>
					<circle
						class="text-green-500 checkmark-circle"
						cx="26"
						cy="26"
						r="25"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					/>
					<path
						class="text-green-500 checkmark-check"
						fill="none"
						stroke="currentColor"
						stroke-width="3"
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M14 27l7 7 16-16"
					/>
				</svg>
			</div>

			<h1 class="text-3xl font-bold text-gray-900 mb-4" in:fly={{ y: 20, duration: 400, delay: 200 }}>
				You're all set!
			</h1>
			<p class="text-gray-600 text-lg mb-8" in:fly={{ y: 20, duration: 400, delay: 300 }}>
				Your workspace is ready. Let's dive in.
			</p>

			<!-- Preview -->
			<div
				class="bg-gray-50 border border-gray-200 rounded-2xl p-6 mb-8"
				in:fly={{ y: 20, duration: 400, delay: 400 }}
			>
				<div class="flex items-center gap-3 mb-4">
					<div class="w-10 h-10 bg-gray-900 rounded-xl flex items-center justify-center">
						<span class="text-white font-bold">{workspaceName?.[0] || 'B'}</span>
					</div>
					<div class="text-left">
						<p class="font-semibold text-gray-900">{workspaceName || 'Your Workspace'}</p>
						<p class="text-sm text-gray-500">{businessName || 'Business OS'}</p>
					</div>
				</div>
				<div class="grid grid-cols-3 gap-2">
					{#each useCases.slice(0, 3) as uc}
						<div class="bg-white rounded-lg p-2 text-xs text-gray-600 text-center">
							{useCaseOptions.find(o => o.value === uc)?.label || uc}
						</div>
					{/each}
				</div>
			</div>

			<button
				type="button"
				onclick={goToDashboard}
				class="btn btn-primary w-full h-12 text-base flex items-center justify-center gap-2"
				in:fly={{ y: 20, duration: 400, delay: 500 }}
			>
				Go to Dashboard
				<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
				</svg>
			</button>
		</div>
	</div>
{/if}

<style>
	.checkmark-circle {
		stroke-dasharray: 166;
		stroke-dashoffset: 166;
		animation: stroke 0.6s ease forwards;
	}

	.checkmark-check {
		stroke-dasharray: 48;
		stroke-dashoffset: 48;
		animation: stroke 0.3s ease forwards 0.6s;
	}

	@keyframes stroke {
		to {
			stroke-dashoffset: 0;
		}
	}
</style>
