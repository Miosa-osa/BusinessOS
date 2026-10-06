<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { Building2, Check, Loader2, Mail, ShieldCheck, X } from 'lucide-svelte';
	import { initCSRF } from '$lib/api/base';
	import { acceptWorkspaceInvite, validateWorkspaceInvite } from '$lib/api/workspaces';
	import { useSession } from '$lib/auth-client';
	import { switchWorkspace } from '$lib/stores/workspaces';

	type Status = 'validating' | 'ready' | 'accepting' | 'success' | 'error';

	const session = useSession();
	const token = $derived($page.params.token);
	const returnPath = $derived(`/invite/${token}?accept=1`);
	const loginUrl = $derived(`/login?redirect=${encodeURIComponent(returnPath)}`);
	const isLoggedIn = $derived(!$session.isPending && Boolean($session.data?.user));

	let status = $state<Status>('validating');
	let workspaceName = $state('this workspace');
	let invitedEmail = $state('');
	let role = $state('member');
	let expiresAt = $state('');
	let errorMessage = $state('');
	let autoAcceptStarted = $state(false);
	let joinedWorkspaceId = $state('');

	onMount(async () => {
		if (!token) return showError('This invitation link is incomplete.');
		try {
			await initCSRF();
			const invite = await validateWorkspaceInvite(token);
			if (!invite.valid) return showError(invite.error || 'This invitation is no longer available.');
			workspaceName = invite.workspace_name || 'this workspace';
			invitedEmail = invite.email || '';
			role = invite.role || 'member';
			expiresAt = invite.expires_at || '';
			status = 'ready';
		} catch (error) {
			showError(error instanceof Error ? error.message : 'We could not verify this invitation.');
		}
	});

	$effect(() => {
		const shouldAccept = $page.url.searchParams.get('accept') === '1';
		if (status === 'ready' && isLoggedIn && shouldAccept && !autoAcceptStarted) {
			autoAcceptStarted = true;
			void handleAccept();
		}
	});

	function showError(message: string) {
		status = 'error';
		errorMessage = message;
	}

	async function handleAccept() {
		if (!isLoggedIn) return void goto(loginUrl);
		if (!token) return showError('This invitation link is incomplete.');
		status = 'accepting';
		errorMessage = '';
		try {
			await initCSRF();
			const result = await acceptWorkspaceInvite(token);
			if (result?.workspace_id) {
				joinedWorkspaceId = result.workspace_id;
				await switchWorkspace(result.workspace_id);
			}
			status = 'success';
		} catch (error) {
			showError(error instanceof Error ? error.message : 'We could not accept this invitation.');
		}
	}

	async function openWorkspace() {
		if (joinedWorkspaceId) await switchWorkspace(joinedWorkspaceId);
		await goto('/dashboard');
	}

	function formatRole(value: string) {
		return value.split(/[_-]/).filter(Boolean).map((part) => part[0].toUpperCase() + part.slice(1)).join(' ');
	}

	function formatDate(value: string) {
		if (!value) return '';
		const date = new Date(`${value}T12:00:00`);
		return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat('en-US', {
			month: 'long', day: 'numeric', year: 'numeric'
		}).format(date);
	}
</script>

<svelte:head>
	<title>Workspace invitation | BusinessOS</title>
</svelte:head>

<div class="page-shell">
	<header class="brand-bar">
		<a class="brand" href="/" aria-label="BusinessOS home">
			<span class="brand-mark"><Building2 size={17} /></span>
			<span>BUSINESS<span class="muted">OS</span></span>
		</a>
	</header>

	<main>
		<section class="invite-panel" aria-live="polite">
			{#if status === 'validating'}
				<div class="center-state">
					<Loader2 class="spinner" size={28} />
					<h1>Checking your invitation</h1>
					<p>One moment while BusinessOS verifies the link.</p>
				</div>
			{:else if status === 'error'}
				<div class="center-state">
					<span class="status-icon error"><X size={24} /></span>
					<p class="eyebrow">Workspace invitation</p>
					<h1>Invitation unavailable</h1>
					<p>{errorMessage}</p>
					<a class="secondary-button" href="/login">Go to BusinessOS sign in</a>
				</div>
			{:else if status === 'success'}
				<div class="center-state">
					<span class="status-icon success"><Check size={25} /></span>
					<p class="eyebrow">Invitation accepted</p>
					<h1>You joined {workspaceName}</h1>
					<p>Your BusinessOS workspace is ready.</p>
					<div class="success-details">
						<span><small>Workspace</small>{workspaceName}</span>
						<span><small>Signed in as</small>{$session.data?.user?.email}</span>
					</div>
					<button class="primary-button" onclick={openWorkspace}>Open {workspaceName}</button>
				</div>
			{:else}
				<div class="center-state">
					<span class="workspace-icon"><Building2 size={22} /></span>
					<p class="eyebrow">BusinessOS workspace invitation</p>
					<h1>Join {workspaceName}</h1>
					<p>You have been invited to work with the {workspaceName} team in BusinessOS.</p>
				</div>

				<div class="invite-details">
					<div><Mail size={17} /><span><small>Invited account</small>{invitedEmail || 'Your invited email address'}</span></div>
					<div><ShieldCheck size={17} /><span><small>Workspace role</small>{formatRole(role)}</span></div>
				</div>

				{#if status === 'accepting'}
					<button class="primary-button" disabled><Loader2 class="spinner" size={18} />Joining {workspaceName}</button>
				{:else if isLoggedIn}
					<p class="helper">Signed in as <strong>{$session.data?.user?.email}</strong></p>
					{#if invitedEmail && $session.data?.user?.email?.toLowerCase() !== invitedEmail.toLowerCase()}
						<p class="account-warning">This invitation is for {invitedEmail}. Sign in with that account to accept it.</p>
					{/if}
					<button class="primary-button" onclick={handleAccept}>Accept and join workspace</button>
				{:else}
					<a class="primary-button" href={loginUrl}>Sign in to accept</a>
					<p class="helper">Use {invitedEmail || 'the email address that received this invitation'} when you sign in or create your account.</p>
				{/if}
				{#if expiresAt}<p class="expiry">Invitation expires {formatDate(expiresAt)}</p>{/if}
			{/if}
		</section>
	</main>
	<footer>BusinessOS by MIOSA</footer>
</div>

<style>
	:global(body) { margin: 0; background: #f5f6f7; color: #111318; font-family: Inter, ui-sans-serif, system-ui, sans-serif; }
	.page-shell { min-height: 100vh; display: grid; grid-template-rows: auto 1fr auto; }
	.brand-bar { height: 64px; display: flex; align-items: center; border-bottom: 1px solid #dedfe2; background: white; padding: 0 32px; }
	.brand { display: inline-flex; align-items: center; gap: 10px; color: #111318; font-size: 14px; font-weight: 750; letter-spacing: 0; text-decoration: none; }
	.brand-mark, .workspace-icon, .status-icon { display: inline-grid; place-items: center; border-radius: 6px; background: #111318; color: white; }
	.brand-mark { width: 30px; height: 30px; }
	.muted { color: #8b8e95; font-weight: 650; }
	main { display: grid; place-items: center; padding: 48px 20px; }
	.invite-panel { width: min(100%, 560px); box-sizing: border-box; border: 1px solid #d9dadd; border-radius: 8px; background: white; padding: 40px; box-shadow: 0 16px 48px rgba(17,19,24,.08); }
	.center-state { text-align: center; }
	.workspace-icon, .status-icon { width: 48px; height: 48px; margin-bottom: 20px; }
	.status-icon.success { background: #e6f4ea; color: #217a3c; }
	.status-icon.error { background: #fce8e8; color: #b42318; }
	.eyebrow { margin: 0 0 10px; color: #6c7078; font-size: 12px; font-weight: 700; text-transform: uppercase; }
	h1 { margin: 0; font-size: clamp(26px, 5vw, 34px); line-height: 1.15; letter-spacing: 0; }
	.center-state > p:not(.eyebrow) { margin: 14px auto 0; max-width: 420px; color: #646871; font-size: 15px; line-height: 1.6; }
	.invite-details { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin: 30px 0 24px; }
	.invite-details > div { min-width: 0; display: flex; align-items: flex-start; gap: 10px; border: 1px solid #e2e3e6; border-radius: 6px; background: #fafafa; padding: 14px; font-size: 13px; font-weight: 600; }
	.invite-details svg { flex: 0 0 auto; color: #6c7078; }
	.invite-details span { min-width: 0; overflow-wrap: anywhere; }
	.invite-details small { display: block; margin-bottom: 4px; color: #7a7e86; font-size: 11px; font-weight: 600; }
	.success-details { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin: 28px 0 18px; text-align: left; }
	.success-details span { min-width: 0; border: 1px solid #e2e3e6; border-radius: 6px; background: #fafafa; padding: 13px 14px; font-size: 13px; font-weight: 650; overflow-wrap: anywhere; }
	.success-details small { display: block; margin-bottom: 4px; color: #7a7e86; font-size: 11px; font-weight: 600; }
	.primary-button, .secondary-button { width: 100%; min-height: 46px; box-sizing: border-box; display: flex; align-items: center; justify-content: center; gap: 9px; border: 1px solid #111318; border-radius: 6px; padding: 11px 16px; font: inherit; font-size: 14px; font-weight: 700; text-align: center; text-decoration: none; cursor: pointer; }
	.primary-button { background: #111318; color: white; }
	.primary-button:hover:not(:disabled) { background: #2b2e34; }
	.primary-button:disabled { cursor: wait; opacity: .72; }
	.secondary-button { width: auto; margin-top: 26px; background: white; color: #111318; }
	.helper, .expiry, .account-warning { margin: 12px 0; color: #6c7078; font-size: 12px; line-height: 1.5; text-align: center; overflow-wrap: anywhere; }
	.account-warning { border-left: 3px solid #bd7b00; background: #fff8e8; color: #694500; padding: 10px 12px; text-align: left; }
	.expiry { margin-top: 20px; }
	.spinner { animation: spin .8s linear infinite; }
	.center-state .spinner { margin-bottom: 18px; color: #555a63; }
	footer { padding: 0 20px 24px; color: #8b8e95; font-size: 11px; text-align: center; }
	@keyframes spin { to { transform: rotate(360deg); } }
	@media (max-width: 560px) {
		.brand-bar { padding: 0 18px; }
		main { align-items: start; padding: 24px 12px; }
		.invite-panel { padding: 28px 20px; }
		.invite-details { grid-template-columns: 1fr; }
		.success-details { grid-template-columns: 1fr; }
	}
</style>
