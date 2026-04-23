<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { getApiBaseUrl } from '$lib/api/base';
	import PricingModal from '$lib/components/computer/PricingModal.svelte';
	import ComputerDashboard from '$lib/components/computer/ComputerDashboard.svelte';
	import NoComputerState from '$lib/components/computer/NoComputerState.svelte';

	// ── Types ──────────────────────────────────────────────────────────────────

	interface Computer {
		id: string;
		status: 'running' | 'hibernating' | 'stopped';
		domain: string;
		region: string;
		created_at: string;
	}

	interface Metrics {
		ram_used_gb: number;
		ram_total_gb: number;
		cpu_percent: number;
		cpu_cores: number;
		storage_used_gb: number;
		storage_total_gb: number;
	}

	interface Runtime {
		id: string;
		name: string;
		status: 'active' | 'idle' | 'stopped';
		memory_gb: number;
		uptime_seconds: number;
	}

	interface Subscription {
		plan_name: string;
		price_cents: number;
		interval: 'month' | 'year';
		credits_used: number;
		credits_total: number;
		seats: number;
		next_billing_date: string;
	}

	interface Plan {
		id: string;
		name: string;
		price_cents: number;
		interval: 'month' | 'year';
		ram_gb: number;
		cpu_cores: number;
		storage_gb: number;
		credits: number;
		seats: string;
		features: string[];
		is_popular?: boolean;
		is_current?: boolean;
	}

	// ── State ──────────────────────────────────────────────────────────────────

	let computer = $state<Computer | null>(null);
	let metrics = $state<Metrics | null>(null);
	let runtimes = $state<Runtime[]>([]);
	let subscription = $state<Subscription | null>(null);
	let plans = $state<Plan[]>([]);
	let loading = $state(true);
	let showPricingModal = $state(false);
	let activeView = $state<'dashboard' | 'desktop'>('dashboard');
	let stoppingRuntime = $state<string | null>(null);

	// ── Helpers ────────────────────────────────────────────────────────────────

	function buildHeaders(): Record<string, string> {
		return {};
	}

	async function apiFetch<T>(path: string): Promise<T | null> {
		try {
			const res = await fetch(`${getApiBaseUrl()}${path}`, {
				credentials: 'include',
				headers: buildHeaders(),
				signal: AbortSignal.timeout(8000),
			});
			if (!res.ok) return null;
			return res.json() as Promise<T>;
		} catch {
			return null;
		}
	}

	// ── Fallback plans (used when /api/billing/plans is unavailable) ───────────

	const FALLBACK_PLANS: Plan[] = [
		{
			id: 'pro',
			name: 'Pro',
			price_cents: 4000,
			interval: 'month',
			ram_gb: 4,
			cpu_cores: 2,
			storage_gb: 10,
			credits: 500,
			seats: 'Unlimited',
			features: [],
		},
		{
			id: 'growth',
			name: 'Growth',
			price_cents: 10000,
			interval: 'month',
			ram_gb: 8,
			cpu_cores: 4,
			storage_gb: 50,
			credits: 2000,
			seats: 'Unlimited',
			features: ['Priority support'],
			is_popular: true,
		},
		{
			id: 'business',
			name: 'Business',
			price_cents: 20000,
			interval: 'month',
			ram_gb: 16,
			cpu_cores: 8,
			storage_gb: 100,
			credits: 5000,
			seats: 'Unlimited',
			features: ['SSO', 'Audit logs', 'Dedicated support'],
		},
	];

	// ── Fallback runtimes (mock when API unavailable) ──────────────────────────

	const FALLBACK_RUNTIMES: Runtime[] = [
		{ id: 'r1', name: 'Claude Code', status: 'active', memory_gb: 2.1, uptime_seconds: 4980 },
		{ id: 'r2', name: 'Ollama', status: 'idle', memory_gb: 1.4, uptime_seconds: 12600 },
	];

	// ── Lifecycle ──────────────────────────────────────────────────────────────

	let keepaliveTimer: ReturnType<typeof setInterval> | null = null;
	let metricsInterval: ReturnType<typeof setInterval> | null = null;

	function startKeepalive() {
		if (keepaliveTimer) return;
		keepaliveTimer = setInterval(async () => {
			if (!computer || computer.status !== 'running') return;
			try {
				await apiFetch('/computer');
			} catch { /* silent */ }
		}, 120_000);
	}

	function startMetricsRefresh() {
		if (metricsInterval) return;
		metricsInterval = setInterval(async () => {
			if (activeView !== 'dashboard' || !computer || computer.status !== 'running') return;
			const metResp = await apiFetch<{ ram_used_mb: number; ram_total_mb: number; cpu_percent: number; storage_used_gb: number; storage_total_gb: number }>('/computer/metrics');
			if (metResp) {
				metrics = {
					ram_used_gb: (metResp.ram_used_mb ?? 0) / 1024,
					ram_total_gb: (metResp.ram_total_mb ?? 0) / 1024,
					cpu_percent: metResp.cpu_percent ?? 0,
					cpu_cores: metrics?.cpu_cores ?? 2,
					storage_used_gb: metResp.storage_used_gb ?? 0,
					storage_total_gb: metResp.storage_total_gb ?? 0,
				};
			}
		}, 30_000);
	}

	onDestroy(() => {
		if (keepaliveTimer) clearInterval(keepaliveTimer);
		if (metricsInterval) clearInterval(metricsInterval);
	});

	onMount(async () => {
		const [compResp, metResp, rtResp, subResp, plResp] = await Promise.all([
			apiFetch<{ computer: Computer | null }>('/computer'),
			apiFetch<{ ram_used_mb: number; ram_total_mb: number; cpu_percent: number; storage_used_gb: number; storage_total_gb: number }>('/computer/metrics'),
			apiFetch<{ runtimes: Array<{ name: string; status: string; memory_mb: number; uptime_seconds: number }> }>('/computer/runtimes'),
			apiFetch<{ subscription: { plan: string; price_monthly: number; credits_total: number; credits_used: number; seats_used: number; next_billing_at: string; status: string } }>('/billing/subscription'),
			apiFetch<{ plans: Array<{ id: string; name: string; price_monthly: number; ram_gb: number; cpus: number; storage_gb: number; credits_included: number; features: string[] }> }>('/billing/plans'),
		]);

		computer = compResp?.computer ?? null;

		if (metResp) {
			metrics = {
				ram_used_gb: (metResp.ram_used_mb ?? 0) / 1024,
				ram_total_gb: (metResp.ram_total_mb ?? 0) / 1024,
				cpu_percent: metResp.cpu_percent ?? 0,
				cpu_cores: 2,
				storage_used_gb: metResp.storage_used_gb ?? 0,
				storage_total_gb: metResp.storage_total_gb ?? 0,
			};
		}

		if (rtResp?.runtimes) {
			runtimes = rtResp.runtimes.map((r, i) => ({
				id: `r${i}`,
				name: r.name,
				status: r.status as Runtime['status'],
				memory_gb: (r.memory_mb ?? 0) / 1024,
				uptime_seconds: r.uptime_seconds,
			}));
		} else {
			runtimes = FALLBACK_RUNTIMES;
		}

		if (subResp?.subscription) {
			const s = subResp.subscription;
			subscription = {
				plan_name: s.plan,
				price_cents: s.price_monthly,
				interval: 'month',
				credits_used: s.credits_used,
				credits_total: s.credits_total,
				seats: s.seats_used,
				next_billing_date: s.next_billing_at,
			};
		}

		if (plResp?.plans && plResp.plans.length > 0) {
			plans = plResp.plans.map((p) => ({
				id: p.id,
				name: p.name,
				price_cents: p.price_monthly,
				interval: 'month' as const,
				ram_gb: p.ram_gb,
				cpu_cores: p.cpus,
				storage_gb: p.storage_gb,
				credits: p.credits_included,
				seats: 'Unlimited',
				features: p.features ?? [],
				is_popular: p.id === 'growth',
			}));
		} else {
			plans = FALLBACK_PLANS;
		}

		if (subscription && plans.length > 0) {
			plans = plans.map((p) => ({
				...p,
				is_current: p.name.toLowerCase() === subscription!.plan_name.toLowerCase(),
			}));
		}

		loading = false;

		if (computer && computer.status === 'running') {
			startKeepalive();
			startMetricsRefresh();
			loadDesktopStream();
		}
	});

	// ── Actions ────────────────────────────────────────────────────────────────

	let provisioning = $state(false);
	let provisionError = $state<string | null>(null);
	let startingComputer = $state(false);

	let desktopSrc = $state('');

	async function loadDesktopStream() {
		if (!computer) return;
		try {
			const res = await fetch(`${getApiBaseUrl()}/computer/desktop-stream`, {
				credentials: 'include',
				headers: buildHeaders(),
				signal: AbortSignal.timeout(15000),
			});
			if (!res.ok) {
				console.error('[Desktop] fetch failed:', res.status);
				return;
			}
			const data = await res.json();
			if (data?.mode === 'cloud' && data.desktop_url) {
				desktopSrc = data.desktop_url;
			}
		} catch (err) {
			console.error('[Desktop] error:', err);
		}
	}

	async function switchToDesktop() {
		activeView = 'desktop';
		if (!desktopSrc) {
			await loadDesktopStream();
		}
	}

	async function handleStartComputer() {
		if (!computer) return;
		startingComputer = true;
		try {
			const res = await fetch(`${getApiBaseUrl()}/computer/terminal-session`, {
				credentials: 'include',
				signal: AbortSignal.timeout(90000),
			});
			if (res.ok) {
				const data = await res.json();
				if (data.mode === 'cloud') {
					computer = { ...computer!, status: 'running' };
				}
			}
			const statusRes = await apiFetch<{ computer: Computer | null }>('/computer');
			if (statusRes?.computer) {
				computer = statusRes.computer;
			}
		} catch {
			const statusRes = await apiFetch<{ computer: Computer | null }>('/computer');
			if (statusRes?.computer) {
				computer = statusRes.computer;
			}
		} finally {
			startingComputer = false;
		}
	}

	async function selectPlan(planId: string) {
		provisioning = true;
		provisionError = null;
		showPricingModal = false;

		try {
			const subRes = await fetch(`${getApiBaseUrl()}/billing/subscribe`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				signal: AbortSignal.timeout(10000),
				body: JSON.stringify({ plan: planId }),
			});
			if (!subRes.ok) throw new Error(`Subscribe failed: ${subRes.status}`);
			const subData = await subRes.json();

			if (subData?.subscription) {
				const s = subData.subscription;
				subscription = {
					plan_name: s.plan,
					price_cents: s.price_monthly,
					interval: 'month',
					credits_used: s.credits_used ?? 0,
					credits_total: s.credits_total ?? 500,
					seats: s.seats_used ?? 1,
					next_billing_date: s.next_billing_at ?? '',
				};
			}

			const compRes = await fetch(`${getApiBaseUrl()}/computer`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				signal: AbortSignal.timeout(15000),
				body: JSON.stringify({ plan: planId }),
			});
			if (!compRes.ok) throw new Error(`Create computer failed: ${compRes.status}`);
			const compData = await compRes.json();

			if (compData?.computer) {
				const c = compData.computer;
				computer = {
					id: c.id,
					status: c.status === 'provisioning' ? 'running' : c.status,
					domain: c.domain ?? `${planId}.app.businessos.com`,
					region: 'us-east-1',
					created_at: c.created_at ?? new Date().toISOString(),
				};
			}

			const [metResp, rtResp] = await Promise.all([
				apiFetch<{ ram_used_mb: number; ram_total_mb: number; cpu_percent: number; storage_used_gb: number; storage_total_gb: number }>('/computer/metrics'),
				apiFetch<{ runtimes: Array<{ name: string; status: string; memory_mb: number; uptime_seconds: number }> }>('/computer/runtimes'),
			]);

			if (metResp) {
				metrics = {
					ram_used_gb: (metResp.ram_used_mb ?? 0) / 1024,
					ram_total_gb: (metResp.ram_total_mb ?? 0) / 1024,
					cpu_percent: metResp.cpu_percent ?? 0,
					cpu_cores: planId === 'business' ? 8 : planId === 'growth' ? 4 : 2,
					storage_used_gb: metResp.storage_used_gb ?? 0,
					storage_total_gb: metResp.storage_total_gb ?? 10,
				};
			}

			if (rtResp?.runtimes) {
				runtimes = rtResp.runtimes.map((r, i) => ({
					id: `r${i}`,
					name: r.name,
					status: r.status as Runtime['status'],
					memory_gb: (r.memory_mb ?? 0) / 1024,
					uptime_seconds: r.uptime_seconds,
				}));
			}

			plans = plans.map((p) => ({ ...p, is_current: p.id === planId }));

		} catch (err) {
			provisionError = err instanceof Error ? err.message : 'Failed to create computer';
		} finally {
			provisioning = false;
		}
	}

	async function stopRuntime(id: string) {
		stoppingRuntime = id;
		try {
			await fetch(`${getApiBaseUrl()}/computer/runtimes/${id}/stop`, {
				method: 'POST',
				credentials: 'include',
			});
			runtimes = runtimes.map((r) => (r.id === id ? { ...r, status: 'stopped' } : r));
		} catch {
			// no-op — gracefully ignore
		} finally {
			stoppingRuntime = null;
		}
	}
</script>

{#if showPricingModal}
	<PricingModal
		{plans}
		{provisioning}
		hasSubscription={subscription !== null}
		onClose={() => (showPricingModal = false)}
		onSelectPlan={selectPlan}
	/>
{/if}

<div class="cp-page">

	{#if loading}
		<div class="cp-skeletons" in:fade>
			<div class="cp-skeleton cp-skeleton--hero"></div>
			<div class="cp-metrics-grid">
				<div class="cp-skeleton cp-skeleton--metric"></div>
				<div class="cp-skeleton cp-skeleton--metric"></div>
				<div class="cp-skeleton cp-skeleton--metric"></div>
			</div>
			<div class="cp-bottom-grid">
				<div class="cp-skeleton cp-skeleton--panel"></div>
				<div class="cp-skeleton cp-skeleton--panel"></div>
			</div>
		</div>

	{:else if provisioning}
		<div class="cp-empty" in:fade={{ duration: 250 }}>
			<div class="cp-empty-card" in:fly={{ y: 16, duration: 300 }}>
				<div class="cp-empty-icon cp-empty-icon--spin" aria-hidden="true">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
						<path d="M21 12a9 9 0 1 1-6.219-8.56"/>
					</svg>
				</div>
				<h1 class="cp-empty-title">Creating Your Computer...</h1>
				<p class="cp-empty-subtitle">
					Setting up your cloud environment. This takes about 60 seconds.
				</p>
			</div>
		</div>

	{:else if computer === null}
		<NoComputerState
			{plans}
			{provisionError}
			onGetStarted={() => (showPricingModal = true)}
			onDismissError={() => (provisionError = null)}
		/>

	{:else}
		<ComputerDashboard
			{computer}
			{metrics}
			{runtimes}
			{subscription}
			{stoppingRuntime}
			{startingComputer}
			{activeView}
			{desktopSrc}
			onStartComputer={handleStartComputer}
			onSwitchToDashboard={() => (activeView = 'dashboard')}
			onSwitchToDesktop={switchToDesktop}
			onLoadDesktopStream={loadDesktopStream}
			onStopRuntime={stopRuntime}
			onOpenPricing={() => (showPricingModal = true)}
		/>
	{/if}

</div>

<style>
	.cp-page {
		padding: 0 16px 16px;
		max-width: 100%;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		gap: 8px;
		min-height: 100%;
		height: 100vh;
		overflow: hidden;
		background: var(--dbg2, #f5f5f7);
	}

	:global(.dark) .cp-page {
		background: var(--dbg2, #0f0f11);
	}

	/* ── Skeleton loading ──────────────────────────────────────────────────── */
	.cp-skeletons {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.cp-skeleton {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 12px;
		animation: cp-shimmer 1.6s ease-in-out infinite;
	}

	:global(.dark) .cp-skeleton {
		background: var(--dbg, rgba(255,255,255,0.04));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}

	.cp-skeleton--hero { height: 96px; }
	.cp-skeleton--metric { height: 120px; }
	.cp-skeleton--panel { height: 300px; }

	@keyframes cp-shimmer {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.5; }
	}

	/* ── Skeleton grid helpers (mirror dashboard layout for accurate skeletons) */
	.cp-metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	@media (max-width: 900px) {
		.cp-metrics-grid { grid-template-columns: 1fr; }
	}

	.cp-bottom-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}

	@media (max-width: 900px) {
		.cp-bottom-grid { grid-template-columns: 1fr; }
	}

	/* ── Provisioning / empty state ────────────────────────────────────────── */
	.cp-empty {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: calc(100vh - 160px);
	}

	.cp-empty-card {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 16px;
		padding: 48px 40px;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 16px;
		max-width: 640px;
		width: 100%;
	}
	:global(.dark) .cp-empty-card {
		background: var(--dbg, rgba(255,255,255,0.03));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}

	.cp-empty-icon {
		width: 72px;
		height: 72px;
		border-radius: 18px;
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.12), rgba(139, 92, 246, 0.12));
		color: #3b82f6;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cp-empty-icon--spin {
		animation: cp-spin 1s linear infinite;
	}

	@keyframes cp-spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.cp-empty-title {
		font-size: 26px;
		font-weight: 700;
		color: var(--dt, #111827);
		margin: 0;
		letter-spacing: -0.03em;
	}
	:global(.dark) .cp-empty-title { color: var(--dt, #f3f4f6); }

	.cp-empty-subtitle {
		font-size: 14px;
		color: var(--dt3, #6b7280);
		margin: 0;
		max-width: 400px;
		line-height: 1.6;
	}
	:global(.dark) .cp-empty-subtitle { color: var(--dt3, rgba(255,255,255,0.4)); }
</style>
