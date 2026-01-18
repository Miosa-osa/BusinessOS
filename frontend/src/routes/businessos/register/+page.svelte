<script lang="ts">
	import { goto } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { signUpWithEmail, initiateGoogleOAuth } from '$lib/auth-client';
	import { AuthLayout, FormInput, PasswordInput } from '$lib/components/auth';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let agreedToTerms = $state(false);
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';

		if (!agreedToTerms) {
			error = 'You must agree to the Terms of Service and Privacy Policy';
			return;
		}

		if (password.length < 8) {
			error = 'Password must be at least 8 characters';
			return;
		}

		loading = true;

		try {
			const result = await signUpWithEmail(email, password, name);
			if (result.error) {
				error = result.error.message || 'Registration failed';
				loading = false;
				return;
			}
			loading = false;
			goto('/onboarding');
		} catch (err) {
			error = (err as Error).message || 'Registration failed';
			loading = false;
		}
	}

	function handleGoogleSignUp() {
		initiateGoogleOAuth();
	}
</script>

<AuthLayout>
	<div in:fly={{ y: 20, duration: 400 }}>
		<!-- Header -->
		<div class="mb-8">
			<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Create account</h1>
			<p class="text-gray-500 text-sm font-mono">Initialize your workspace</p>
		</div>

		<!-- Form -->
		<form onsubmit={handleSubmit} class="space-y-5">
			{#if error}
				<div class="bg-red-50 border border-red-100 rounded-lg px-4 py-3 flex items-center gap-3" in:fly={{ y: -10, duration: 200 }}>
					<svg class="w-4 h-4 text-red-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					<p class="text-sm text-red-600 font-mono">{error}</p>
				</div>
			{/if}

			<FormInput
				id="name"
				label="Full name"
				type="text"
				bind:value={name}
				placeholder="John Smith"
				autocomplete="name"
				required
			/>

			<FormInput
				id="email"
				label="Email"
				type="email"
				bind:value={email}
				placeholder="you@company.com"
				autocomplete="email"
				required
			/>

			<PasswordInput
				id="password"
				label="Password"
				bind:value={password}
				autocomplete="new-password"
				required
				showStrength
			/>

			<!-- Terms Checkbox -->
			<label class="flex items-start gap-3 cursor-pointer group">
				<input
					type="checkbox"
					bind:checked={agreedToTerms}
					class="mt-0.5 h-4 w-4 rounded border-gray-300 text-black focus:ring-black cursor-pointer"
				/>
				<span class="text-sm text-gray-500 leading-relaxed font-mono">
					I agree to the
					<a href="/terms" class="text-gray-900 hover:underline">Terms</a>
					and
					<a href="/privacy" class="text-gray-900 hover:underline">Privacy Policy</a>
				</span>
			</label>

			<button
				type="submit"
				class="w-full h-12 bg-black text-white rounded-lg text-sm font-medium hover:bg-gray-800 transition-all flex items-center justify-center gap-2 font-mono"
				disabled={loading}
			>
				{#if loading}
					<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" fill="none" />
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
					</svg>
					Initializing...
				{:else}
					Create account
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14 5l7 7m0 0l-7 7m7-7H3" />
					</svg>
				{/if}
			</button>
		</form>

		<!-- Divider -->
		<div class="my-8 flex items-center gap-4">
			<div class="flex-1 h-px bg-gray-100"></div>
			<span class="text-xs text-gray-400 font-mono uppercase tracking-wider">or</span>
			<div class="flex-1 h-px bg-gray-100"></div>
		</div>

		<!-- Social Login -->
		<button
			type="button"
			onclick={handleGoogleSignUp}
			class="w-full h-12 bg-white border border-gray-200 text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50 hover:border-gray-300 transition-all flex items-center justify-center gap-3"
		>
			<svg class="w-5 h-5" viewBox="0 0 24 24">
				<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
				<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
				<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
				<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
			</svg>
			Continue with Google
		</button>

		<!-- Sign In Link -->
		<p class="mt-8 text-center text-sm text-gray-500 font-mono">
			Have an account?
			<a href="/login" class="text-gray-900 hover:underline ml-1">
				Sign in
			</a>
		</p>
	</div>
</AuthLayout>
