<script lang="ts">
	import { goto } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import { signInWithEmail, initiateGoogleOAuth, cloudServerUrl, getSession } from '$lib/auth-client';
	import { AuthLayout, FormInput, PasswordInput } from '$lib/components/auth';
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);
	let rememberMe = $state(false);
	let rememberedUser = $state<{ email: string; name?: string } | null>(null);

	onMount(() => {
		// Check if there's a remembered user
		try {
			const stored = localStorage.getItem('rememberedUser');
			if (stored) {
				rememberedUser = JSON.parse(stored);
				if (rememberedUser?.email) {
					email = rememberedUser.email;
					rememberMe = true;
				}
			}
		} catch (e) {
			console.error('Error loading remembered user:', e);
		}
	});

	function clearRememberedUser() {
		localStorage.removeItem('rememberedUser');
		rememberedUser = null;
		email = '';
		rememberMe = false;
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = '';
		loading = true;

		try {
			const result = await signInWithEmail(email, password);
			if (result.error) {
				error = result.error.message || 'Invalid email or password';
				loading = false;
				return;
			}

			// Save remembered user if checkbox is checked
			if (rememberMe) {
				localStorage.setItem('rememberedUser', JSON.stringify({ email }));
			} else {
				localStorage.removeItem('rememberedUser');
			}

			loading = false;

			// Check if user has completed onboarding
			const session = await getSession();
			if (session.data?.user?.onboardingCompleted === false) {
				// User hasn't completed onboarding, redirect there
				goto('/onboarding');
			} else {
				// User has completed onboarding, go to main app
				goto('/window');
			}
		} catch (err) {
			error = (err as Error).message || 'Authentication failed';
			loading = false;
		}
	}

	function handleGoogleSignIn() {
		initiateGoogleOAuth();
	}
</script>

<AuthLayout>
	<div in:fly={{ y: 20, duration: 400 }}>
		<!-- Header -->
		<div class="mb-8">
			{#if rememberedUser}
				<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Welcome back</h1>
				<div class="flex items-center gap-3 mt-3 p-3 bg-gray-50 rounded-lg border border-gray-100">
					<div class="w-10 h-10 bg-black text-white rounded-full flex items-center justify-center font-mono text-sm font-bold">
						{rememberedUser.email.charAt(0).toUpperCase()}
					</div>
					<div class="flex-1 min-w-0">
						<p class="text-sm font-medium text-gray-900 truncate font-mono">{rememberedUser.email}</p>
						<button
							type="button"
							onclick={clearRememberedUser}
							class="text-xs text-gray-400 hover:text-gray-600 transition-colors font-mono"
						>
							Use a different account
						</button>
					</div>
				</div>
			{:else}
				<h1 class="text-2xl font-bold text-gray-900 mb-2 font-mono tracking-tight">Welcome back</h1>
				<p class="text-gray-500 text-sm font-mono">Sign in to continue</p>
			{/if}
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
				id="email"
				label="Email"
				type="email"
				bind:value={email}
				placeholder="you@company.com"
				autocomplete="email"
				required
			/>

			<div>
				<PasswordInput
					id="password"
					label="Password"
					bind:value={password}
					autocomplete="current-password"
					required
				/>
				<div class="mt-3 flex items-center justify-between">
					<label class="flex items-center gap-2 cursor-pointer">
						<input
							type="checkbox"
							bind:checked={rememberMe}
							class="w-4 h-4 rounded border-gray-300 text-black focus:ring-black focus:ring-offset-0"
						/>
						<span class="text-xs text-gray-500 font-mono">Remember me</span>
					</label>
					<a href="/forgot-password" class="text-xs text-gray-400 hover:text-gray-900 transition-colors font-mono">
						Forgot password?
					</a>
				</div>
			</div>

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
					Authenticating...
				{:else}
					Sign in
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
			onclick={handleGoogleSignIn}
			class="btn-pill btn-pill-secondary btn-pill-block flex items-center justify-center gap-3 h-12"
		>
			<svg class="w-5 h-5" viewBox="0 0 24 24">
				<path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
				<path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
				<path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
				<path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
			</svg>
			Continue with Google
		</button>

		<!-- Sign Up Link -->
		<p class="mt-8 text-center text-sm text-gray-500 font-mono">
			No account?
			<a href="/register" class="text-gray-900 hover:underline ml-1">
				Create one
			</a>
		</p>
	</div>
</AuthLayout>
