<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { getApiBaseUrl } from '$lib/api/base';

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
	let desktopStreamAuth = $state('');
	let stoppingRuntime = $state<string | null>(null);

	// ── Derived ────────────────────────────────────────────────────────────────

	let ramPercent = $derived(
		metrics && metrics.ram_total_gb > 0
			? Math.round((metrics.ram_used_gb / metrics.ram_total_gb) * 100)
			: 0
	);
	let storagePercent = $derived(
		metrics && metrics.storage_total_gb > 0
			? Math.round((metrics.storage_used_gb / metrics.storage_total_gb) * 100)
			: 0
	);
	let creditsPercent = $derived(
		subscription && subscription.credits_total > 0
			? Math.round(((subscription.credits_total - subscription.credits_used) / subscription.credits_total) * 100)
			: 0
	);

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

	function formatUptime(seconds: number): string {
		if (seconds < 60) return `${seconds}s`;
		if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
		const h = Math.floor(seconds / 3600);
		const m = Math.floor((seconds % 3600) / 60);
		return m > 0 ? `${h}h ${m}m` : `${h}h`;
	}

	function formatPrice(cents: number): string {
		return `$${(cents / 100).toFixed(0)}`;
	}

	function formatBillingDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });
		} catch {
			return iso;
		}
	}

	function formatDate(iso: string): string {
		try {
			return new Date(iso).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' });
		} catch {
			return iso;
		}
	}

	function calcUptime(createdAt: string): string {
		try {
			const seconds = Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000);
			if (seconds < 60) return `${seconds}s`;
			if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
			const h = Math.floor(seconds / 3600);
			const m = Math.floor((seconds % 3600) / 60);
			if (h < 24) return m > 0 ? `${h}h ${m}m` : `${h}h`;
			const d = Math.floor(h / 24);
			const rh = h % 24;
			return rh > 0 ? `${d}d ${rh}h` : `${d}d`;
		} catch {
			return '—';
		}
	}

	function extractSlug(domain: string): string {
		try {
			// Extract subdomain or first path segment as the slug
			const url = domain.startsWith('http') ? new URL(domain) : new URL(`https://${domain}`);
			return url.hostname.split('.')[0];
		} catch {
			return domain.split('.')[0] ?? domain;
		}
	}

	function creditsColor(pct: number): string {
		if (pct > 50) return '#22c55e';
		if (pct > 25) return '#f59e0b';
		return '#ef4444';
	}

	function statusLabel(status: string): string {
		return (
			({ running: 'Running', hibernating: 'Hibernating', stopped: 'Stopped', provisioning: 'Provisioning', error: 'Error' } as Record<string, string>)[status] ?? status
		);
	}

	function runtimeStatusLabel(status: string): string {
		return (
			({ active: 'Active', idle: 'Idle', stopped: 'Stopped' } as Record<string, string>)[status] ?? status
		);
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

	// Keepalive — ping the computer every 2 minutes to prevent auto-stop
	let keepaliveTimer: ReturnType<typeof setInterval> | null = null;

	// Metrics auto-refresh — every 30s when dashboard is active and computer is running
	let metricsInterval: ReturnType<typeof setInterval> | null = null;

	function startKeepalive() {
		if (keepaliveTimer) return;
		keepaliveTimer = setInterval(async () => {
			if (!computer || computer.status !== 'running') return;
			try {
				await apiFetch('/computer');
			} catch { /* silent */ }
		}, 120_000); // 2 minutes
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
		}, 30_000); // 30 seconds
	}

	onDestroy(() => {
		if (keepaliveTimer) clearInterval(keepaliveTimer);
		if (metricsInterval) clearInterval(metricsInterval);
	});

	onMount(async () => {
		// API returns wrapped objects — extract the inner data
		const [compResp, metResp, rtResp, subResp, plResp] = await Promise.all([
			apiFetch<{ computer: Computer | null }>('/computer'),
			apiFetch<{ ram_used_mb: number; ram_total_mb: number; cpu_percent: number; storage_used_gb: number; storage_total_gb: number }>('/computer/metrics'),
			apiFetch<{ runtimes: Array<{ name: string; status: string; memory_mb: number; uptime_seconds: number }> }>('/computer/runtimes'),
			apiFetch<{ subscription: { plan: string; price_monthly: number; credits_total: number; credits_used: number; seats_used: number; next_billing_at: string; status: string } }>('/billing/subscription'),
			apiFetch<{ plans: Array<{ id: string; name: string; price_monthly: number; ram_gb: number; cpus: number; storage_gb: number; credits_included: number; features: string[] }> }>('/billing/plans'),
		]);

		// Extract computer (null = no cloud computer)
		computer = compResp?.computer ?? null;

		// Map metrics (backend returns MB, frontend expects GB)
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

		// Map runtimes
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

		// Map subscription
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

		// Map plans
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

		// Mark current plan
		if (subscription && plans.length > 0) {
			plans = plans.map((p) => ({
				...p,
				is_current: p.name.toLowerCase() === subscription!.plan_name.toLowerCase(),
			}));
		}

		loading = false;

		// Start keepalive, metrics refresh, and pre-load desktop URL if computer is running
		if (computer && (computer.status === 'running' || computer.status === 'active')) {
			startKeepalive();
			startMetricsRefresh();
			loadDesktopStream(); // Pre-load so Desktop tab is instant
		}
	});

	// ── Actions ────────────────────────────────────────────────────────────────

	let provisioning = $state(false);
	let provisionError = $state<string | null>(null);
	let startingComputer = $state(false);

	// Desktop iframe URL — loaded from the backend desktop-stream endpoint
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
			console.log('[Desktop] response:', data);
			if (data?.mode === 'cloud' && data.desktop_url) {
				desktopSrc = data.desktop_url;
				desktopStreamAuth = 'loaded';
				console.log('[Desktop] iframe src set:', desktopSrc.substring(0, 80));
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
			// Call MIOSA start via our backend terminal-session endpoint (it auto-wakes)
			const res = await fetch(`${getApiBaseUrl()}/computer/terminal-session`, {
				credentials: 'include',
				signal: AbortSignal.timeout(90000), // 90s — VM boot can take a while
			});
			if (res.ok) {
				const data = await res.json();
				if (data.mode === 'cloud') {
					// Computer is now running — reload all data
					computer = { ...computer!, status: 'running' };
				}
			}
			// Reload computer status regardless
			const statusRes = await apiFetch<{ computer: Computer | null }>('/computer');
			if (statusRes?.computer) {
				computer = statusRes.computer;
			}
		} catch {
			// Timeout is expected if VM takes long — reload status
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
			// 1. Subscribe to the plan
			const subRes = await fetch(`${getApiBaseUrl()}/billing/subscribe`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				signal: AbortSignal.timeout(10000),
				body: JSON.stringify({ plan: planId }),
			});
			if (!subRes.ok) throw new Error(`Subscribe failed: ${subRes.status}`);
			const subData = await subRes.json();

			// Update subscription state
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

			// 2. Create the computer
			const compRes = await fetch(`${getApiBaseUrl()}/computer`, {
				method: 'POST',
				credentials: 'include',
				headers: { 'Content-Type': 'application/json' },
				signal: AbortSignal.timeout(15000),
				body: JSON.stringify({ plan: planId }),
			});
			if (!compRes.ok) throw new Error(`Create computer failed: ${compRes.status}`);
			const compData = await compRes.json();

			// Update computer state
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

			// 3. Load metrics + runtimes for the new computer
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

			// Mark current plan
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
				signal: AbortSignal.timeout(10000),
			});
			runtimes = runtimes.map((r) => (r.id === id ? { ...r, status: 'stopped' } : r));
		} catch {
			// no-op — gracefully ignore (includes timeout)
		} finally {
			stoppingRuntime = null;
		}
	}
</script>

<!-- ── Pricing Modal ─────────────────────────────────────────────────────── -->

{#if showPricingModal}
	<div
		class="cp-modal-backdrop"
		onclick={() => (showPricingModal = false)}
		role="dialog"
		aria-modal="true"
		aria-label="Choose a plan"
		in:fade={{ duration: 180 }}
		out:fade={{ duration: 150 }}
	>
		<div
			class="cp-modal"
			onclick={(e) => e.stopPropagation()}
			in:fly={{ y: 24, duration: 220 }}
		>
			<div class="cp-modal-header">
				<div>
					<h2 class="cp-modal-title">Choose Your Plan</h2>
					<p class="cp-modal-subtitle">Scale your cloud computer as your team grows</p>
				</div>
				<button
					class="cp-modal-close"
					onclick={() => (showPricingModal = false)}
					aria-label="Close pricing modal"
				>
					<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
						<path d="M18 6L6 18M6 6l12 12"/>
					</svg>
				</button>
			</div>

			<div class="cp-plans-grid">
				{#each plans as plan}
					<div class="cp-plan-card {plan.is_current ? 'cp-plan-card--current' : ''} {plan.is_popular ? 'cp-plan-card--popular' : ''}">
						{#if plan.is_popular}
							<div class="cp-plan-badge cp-plan-badge--popular">Popular</div>
						{/if}
						{#if plan.is_current}
							<div class="cp-plan-badge cp-plan-badge--current">Current Plan</div>
						{/if}

						<div class="cp-plan-header">
							<h3 class="cp-plan-name">{plan.name}</h3>
							<div class="cp-plan-price">
								<span class="cp-plan-price-amount">{formatPrice(plan.price_cents)}</span>
								<span class="cp-plan-price-period">/{plan.interval}</span>
							</div>
						</div>

						<ul class="cp-plan-specs">
							<li class="cp-plan-spec">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<rect x="2" y="3" width="20" height="14" rx="2"/>
									<path d="M8 21h8M12 17v4"/>
								</svg>
								{plan.ram_gb} GB RAM
							</li>
							<li class="cp-plan-spec">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<rect x="4" y="4" width="16" height="16" rx="2"/>
									<rect x="9" y="9" width="6" height="6"/>
									<path d="M15 2v2M9 2v2M2 9h2M2 15h2M22 9h-2M22 15h-2M15 22v-2M9 22v-2"/>
								</svg>
								{plan.cpu_cores} vCPUs
							</li>
							<li class="cp-plan-spec">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<ellipse cx="12" cy="5" rx="9" ry="3"/>
									<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
									<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
								</svg>
								{plan.storage_gb} GB SSD
							</li>
							<li class="cp-plan-spec">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<circle cx="12" cy="12" r="10"/>
									<path d="M12 6v6l4 2"/>
								</svg>
								{plan.credits >= 1000 ? `${plan.credits / 1000}K` : plan.credits} credits/mo
							</li>
							<li class="cp-plan-spec">
								<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
									<circle cx="9" cy="7" r="4"/>
									<path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
								</svg>
								{plan.seats} seats
							</li>
							{#each plan.features as feature}
								<li class="cp-plan-spec">
									<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
										<polyline points="20 6 9 17 4 12"/>
									</svg>
									{feature}
								</li>
							{/each}
						</ul>

						{#if plan.is_current}
							<button class="cp-plan-btn cp-plan-btn--current" disabled aria-disabled="true">
								Current Plan
							</button>
						{:else}
							<button
								class="cp-plan-btn cp-plan-btn--upgrade"
								onclick={() => selectPlan(plan.id)}
								disabled={provisioning}
							>
								{provisioning ? 'Creating...' : subscription ? 'Upgrade' : 'Get Started'}
							</button>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	</div>
{/if}

<!-- ── Page ──────────────────────────────────────────────────────────────── -->

<div class="cp-page">

	{#if loading}
		<!-- Loading skeletons -->
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
		<!-- ── Provisioning State ──────────────────────────────────────────────── -->
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
		<!-- ── Free / No Computer State ────────────────────────────────────────── -->
		<div class="cp-empty" in:fade={{ duration: 250 }}>
			<div class="cp-empty-card" in:fly={{ y: 16, duration: 300 }}>
				{#if provisionError}
					<div class="cp-error-banner" role="alert">
						<span>{provisionError}</span>
						<button onclick={() => (provisionError = null)}>Dismiss</button>
					</div>
				{/if}
				<div class="cp-empty-icon" aria-hidden="true">
					<svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" stroke-linejoin="round">
						<rect x="2" y="3" width="20" height="14" rx="2"/>
						<path d="M8 21h8M12 17v4"/>
					</svg>
				</div>
				<h1 class="cp-empty-title">Create Your Computer</h1>
				<p class="cp-empty-subtitle">
					Access BusinessOS from anywhere. Invite your team. Run AI agents in the cloud.
				</p>

				<!-- Mini plan comparison -->
				<div class="cp-empty-plans">
					{#each plans.slice(0, 3) as plan}
						<div class="cp-empty-plan">
							<div class="cp-empty-plan-name">{plan.name}</div>
							<div class="cp-empty-plan-price">{formatPrice(plan.price_cents)}<span>/mo</span></div>
							<ul class="cp-empty-plan-specs">
								<li>{plan.ram_gb} GB RAM</li>
								<li>{plan.cpu_cores} vCPUs</li>
								<li>{plan.storage_gb} GB SSD</li>
							</ul>
							{#if plan.is_popular}
								<div class="cp-empty-plan-badge">Popular</div>
							{/if}
						</div>
					{/each}
				</div>

				<button class="cp-btn cp-btn--primary cp-btn--lg" onclick={() => (showPricingModal = true)}>
					Get Started
				</button>
			</div>
		</div>

	{:else}
		<!-- ── Section 1: Computer Status Hero ──────────────────────────────────── -->
		<section class="cp-hero" aria-label="Computer status" in:fly={{ y: -8, duration: 280 }}>
			<div class="cp-hero-left">
				<div class="cp-status-indicator cp-status-indicator--{computer.status}" aria-label="Status: {statusLabel(computer.status)}"></div>
				<div class="cp-hero-info">
					<div class="cp-hero-title-row">
						<h1 class="cp-hero-title">Your Computer</h1>
						<span class="cp-status-badge cp-status-badge--{computer.status}">{statusLabel(computer.status)}</span>
					</div>
					<p class="cp-hero-domain">{computer.domain}</p>
				</div>
			</div>
			<div class="cp-hero-actions">
				{#if computer.status === 'stopped' || computer.status === 'hibernating'}
					<button
						class="cp-btn cp-btn--primary"
						disabled={startingComputer}
						onclick={handleStartComputer}
					>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<polygon points="5 3 19 12 5 21 5 3"/>
						</svg>
						{startingComputer ? 'Starting...' : 'Start Computer'}
					</button>
				{:else}
					<a
						href="/terminal"
						class="cp-btn cp-btn--outline"
						aria-label="Open terminal"
					>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<polyline points="4 17 10 11 4 5"/>
							<line x1="12" y1="19" x2="20" y2="19"/>
						</svg>
						Open Terminal
					</a>
					<button
						class="cp-btn cp-btn--outline"
						aria-label="Open desktop"
						onclick={switchToDesktop}
					>
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<rect x="2" y="3" width="20" height="14" rx="2"/>
							<path d="M8 21h8M12 17v4"/>
						</svg>
						Open Desktop
					</button>
					<button class="cp-btn cp-btn--outline" aria-label="Restart computer" disabled title="Coming soon">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<polyline points="23 4 23 10 17 10"/>
							<path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
						</svg>
						Restart
					</button>
					<button class="cp-btn cp-btn--outline" aria-label="Hibernate computer" disabled title="Coming soon">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
						</svg>
						Hibernate
					</button>
				{/if}
			</div>
		</section>

		<!-- View Toggle: Dashboard / Desktop -->
		{#if computer.status === 'running'}
			<div class="cp-view-tabs">
				<button class="cp-view-tab" class:active={activeView === 'dashboard'} onclick={() => (activeView = 'dashboard')}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/></svg>
					Dashboard
				</button>
				<button class="cp-view-tab" class:active={activeView === 'desktop'} onclick={switchToDesktop}>
					<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8M12 17v4"/></svg>
					Desktop
				</button>
			</div>
		{/if}

		{#if activeView === 'desktop' && computer.status === 'running'}
			<!-- Desktop VNC Stream -->
			<section class="cp-desktop-container">
				{#if desktopSrc}
					<iframe
						src={desktopSrc}
						title="BusinessOS Desktop"
						class="cp-desktop-iframe"
						allow="clipboard-read; clipboard-write; autoplay"
					></iframe>
				{:else}
					<div class="cp-desktop-loading">
						<div class="cp-desktop-spinner"></div>
						<p>Connecting to desktop...</p>
						<button class="cp-btn cp-btn--outline" style="margin-top: 12px;" onclick={loadDesktopStream}>
							Retry
						</button>
						{#if computer?.domain}
							<a href={computer.domain} target="_blank" class="cp-btn cp-btn--outline" style="margin-top: 8px;">
								Open in new tab
							</a>
						{/if}
					</div>
				{/if}
			</section>
		{:else}

		<!-- ── Section 2: Resource Meters ───────────────────────────────────────── -->
		<section class="cp-metrics-grid" aria-label="Resource usage" in:fly={{ y: 12, duration: 300, delay: 50 }}>

			<!-- RAM -->
			<div class="cp-metric-card">
				<div class="cp-metric-header">
					<div class="cp-metric-icon cp-metric-icon--ram" aria-hidden="true">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
							<rect x="2" y="6" width="20" height="12" rx="2"/>
							<path d="M6 10v4M10 10v4M14 10v4M18 10v4M2 12h2M20 12h2"/>
						</svg>
					</div>
					<span class="cp-metric-label">Memory</span>
				</div>
				<div class="cp-metric-value">
					{metrics ? `${metrics.ram_used_gb.toFixed(1)} / ${metrics.ram_total_gb} GB` : '— GB'}
				</div>
				<div class="cp-progress-track" role="progressbar" aria-valuenow={ramPercent} aria-valuemin={0} aria-valuemax={100} aria-label="RAM usage {ramPercent}%">
					<div class="cp-progress-fill" style="width: {ramPercent}%"></div>
				</div>
				<div class="cp-metric-pct">{ramPercent}%</div>
			</div>

			<!-- CPU -->
			<div class="cp-metric-card">
				<div class="cp-metric-header">
					<div class="cp-metric-icon cp-metric-icon--cpu" aria-hidden="true">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
							<rect x="4" y="4" width="16" height="16" rx="2"/>
							<rect x="9" y="9" width="6" height="6"/>
							<path d="M15 2v2M9 2v2M2 9h2M2 15h2M22 9h-2M22 15h-2M15 22v-2M9 22v-2"/>
						</svg>
					</div>
					<span class="cp-metric-label">CPU</span>
				</div>
				<div class="cp-metric-value">
					{metrics ? `${metrics.cpu_percent}%` : '—%'}
				</div>
				<div class="cp-progress-track" role="progressbar" aria-valuenow={metrics?.cpu_percent ?? 0} aria-valuemin={0} aria-valuemax={100} aria-label="CPU usage {metrics?.cpu_percent ?? 0}%">
					<div class="cp-progress-fill" style="width: {metrics?.cpu_percent ?? 0}%"></div>
				</div>
				<div class="cp-metric-pct cp-metric-sub">{metrics ? `${metrics.cpu_cores} cores` : '—'}</div>
			</div>

			<!-- Storage -->
			<div class="cp-metric-card">
				<div class="cp-metric-header">
					<div class="cp-metric-icon cp-metric-icon--storage" aria-hidden="true">
						<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
							<ellipse cx="12" cy="5" rx="9" ry="3"/>
							<path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/>
							<path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>
						</svg>
					</div>
					<span class="cp-metric-label">Storage</span>
				</div>
				<div class="cp-metric-value">
					{metrics ? `${metrics.storage_used_gb.toFixed(1)} / ${metrics.storage_total_gb} GB` : '— GB'}
				</div>
				<div class="cp-progress-track" role="progressbar" aria-valuenow={storagePercent} aria-valuemin={0} aria-valuemax={100} aria-label="Storage usage {storagePercent}%">
					<div class="cp-progress-fill" style="width: {storagePercent}%"></div>
				</div>
				<div class="cp-metric-pct">{storagePercent}%</div>
			</div>
		</section>

		<!-- ── Section 2b: Computer Info ───────────────────────────────────────── -->
		<section class="cp-info-card" aria-label="Computer information" in:fly={{ y: 12, duration: 300, delay: 75 }}>
			<div class="cp-info-header">
				<h2 class="cp-panel-title">Computer Info</h2>
			</div>
			<div class="cp-info-grid">
				<div class="cp-info-row">
					<span class="cp-info-label">Name</span>
					<span class="cp-info-value">Your Computer</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Slug</span>
					<span class="cp-info-value cp-info-value--mono">{extractSlug(computer.domain)}</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Template</span>
					<span class="cp-info-value">MIOSA Desktop</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Region</span>
					<span class="cp-info-value">{computer.region}</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Created</span>
					<span class="cp-info-value">{formatDate(computer.created_at)}</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Uptime</span>
					<span class="cp-info-value">{calcUptime(computer.created_at)}</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">IP Address</span>
					<span class="cp-info-value">Cloud (MIOSA)</span>
				</div>
				<div class="cp-info-row">
					<span class="cp-info-label">Desktop URL</span>
					<a
						href={computer.domain.startsWith('http') ? computer.domain : `https://${computer.domain}`}
						target="_blank"
						rel="noopener noreferrer"
						class="cp-info-link"
					>
						{computer.domain}
						<svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
							<polyline points="15 3 21 3 21 9"/>
							<line x1="10" y1="14" x2="21" y2="3"/>
						</svg>
					</a>
				</div>
			</div>
		</section>

		<!-- ── Section 3: Runtimes + Plan ───────────────────────────────────────── -->
		<section class="cp-bottom-grid" in:fly={{ y: 12, duration: 300, delay: 100 }}>

			<!-- Runtimes column -->
			<div class="cp-panel">
				<div class="cp-panel-header">
					<h2 class="cp-panel-title">Installed Tools</h2>
					<span class="cp-panel-count">{runtimes.filter((r) => r.status !== 'stopped').length + 5} installed</span>
				</div>

				<div class="cp-runtimes-list">
					<!-- Hardcoded tools present on every BusinessOS computer -->
					{#each [
						{ id: 'cc', name: 'Claude Code', status: 'active', label: 'AI Agent' },
						{ id: 'codex', name: 'Codex', status: 'active', label: 'AI Agent' },
						{ id: 'pg', name: 'PostgreSQL', status: 'active', label: 'Database' },
						{ id: 'redis', name: 'Redis', status: 'active', label: 'Cache' },
						{ id: 'go', name: 'Go', status: 'idle', label: 'Runtime' },
					] as tool}
						<div class="cp-runtime-row">
							<div class="cp-runtime-left">
								<div class="cp-runtime-status-dot cp-runtime-status-dot--{tool.status}" aria-hidden="true"></div>
								<div class="cp-runtime-info">
									<span class="cp-runtime-name">{tool.name}</span>
									<span class="cp-runtime-meta">
										<span class="cp-runtime-status-badge cp-runtime-status-badge--{tool.status}">{tool.status === 'active' ? 'Active' : 'Installed'}</span>
										<span class="cp-runtime-stat">{tool.label}</span>
									</span>
								</div>
							</div>
						</div>
					{/each}

					<!-- Dynamic runtimes from API -->
					{#each runtimes as runtime (runtime.id)}
						<div class="cp-runtime-row" in:fade={{ duration: 150 }}>
							<div class="cp-runtime-left">
								<div class="cp-runtime-status-dot cp-runtime-status-dot--{runtime.status}" aria-hidden="true"></div>
								<div class="cp-runtime-info">
									<span class="cp-runtime-name">{runtime.name}</span>
									<span class="cp-runtime-meta">
										<span class="cp-runtime-status-badge cp-runtime-status-badge--{runtime.status}">{runtimeStatusLabel(runtime.status)}</span>
										<span class="cp-runtime-stat">{runtime.memory_gb.toFixed(1)} GB</span>
										<span class="cp-runtime-sep" aria-hidden="true"></span>
										<span class="cp-runtime-stat">{formatUptime(runtime.uptime_seconds)}</span>
									</span>
								</div>
							</div>
							<button
								class="cp-runtime-stop-btn"
								onclick={() => stopRuntime(runtime.id)}
								disabled={stoppingRuntime === runtime.id || runtime.status === 'stopped'}
								aria-label="Stop {runtime.name}"
							>
								{stoppingRuntime === runtime.id ? 'Stopping…' : 'Stop'}
							</button>
						</div>
					{/each}

					{#if runtimes.length === 0}
						<p class="cp-empty-list-text">No runtimes running.</p>
					{/if}
				</div>

				<div class="cp-panel-footer">
					<button class="cp-btn cp-btn--outline cp-btn--full">
						<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
							<circle cx="12" cy="12" r="10"/>
							<line x1="12" y1="8" x2="12" y2="16"/>
							<line x1="8" y1="12" x2="16" y2="12"/>
						</svg>
						Start Runtime
					</button>
				</div>
			</div>

			<!-- Plan & Credits column -->
			<div class="cp-panel">
				{#if subscription}
					<div class="cp-panel-header">
						<h2 class="cp-panel-title">Plan & Credits</h2>
					</div>

					<div class="cp-plan-summary">
						<div class="cp-plan-summary-row">
							<span class="cp-plan-summary-name">{subscription.plan_name}</span>
							<span class="cp-plan-summary-price">
								{formatPrice(subscription.price_cents)}<span class="cp-plan-summary-interval">/{subscription.interval}</span>
							</span>
						</div>

						<!-- Credits meter -->
						<div class="cp-credits-section">
							<div class="cp-credits-header">
								<span class="cp-credits-label">Credits</span>
								<span class="cp-credits-value">
									{(subscription.credits_total - subscription.credits_used).toLocaleString()} / {subscription.credits_total.toLocaleString()} remaining
								</span>
							</div>
							<div
								class="cp-progress-track"
								role="progressbar"
								aria-valuenow={creditsPercent}
								aria-valuemin={0}
								aria-valuemax={100}
								aria-label="Credits: {creditsPercent}% remaining"
							>
								<div
									class="cp-progress-fill cp-progress-fill--solid"
									style="width: {creditsPercent}%; background: {creditsColor(creditsPercent)}"
								></div>
							</div>
						</div>

						<div class="cp-plan-details">
							<div class="cp-plan-detail">
								<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
									<circle cx="9" cy="7" r="4"/>
									<path d="M23 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75"/>
								</svg>
								{subscription.seats} {subscription.seats === 1 ? 'member' : 'members'}
							</div>
							<div class="cp-plan-detail">
								<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
									<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
									<line x1="16" y1="2" x2="16" y2="6"/>
									<line x1="8" y1="2" x2="8" y2="6"/>
									<line x1="3" y1="10" x2="21" y2="10"/>
								</svg>
								Next billing {formatBillingDate(subscription.next_billing_date)}
							</div>
						</div>
					</div>

					<div class="cp-plan-actions">
						<button class="cp-btn cp-btn--primary" onclick={() => (showPricingModal = true)}>
							Upgrade Plan
						</button>
						<button class="cp-btn cp-btn--outline">
							Buy Credits
						</button>
					</div>

				{:else}
					<!-- No subscription CTA -->
					<div class="cp-no-sub-cta">
						<div class="cp-no-sub-icon" aria-hidden="true">
							<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
								<rect x="2" y="3" width="20" height="14" rx="2"/>
								<path d="M8 21h8M12 17v4"/>
							</svg>
						</div>
						<h3 class="cp-no-sub-title">Get a Cloud Computer</h3>
						<p class="cp-no-sub-body">Access BusinessOS from anywhere. Invite your team. Run AI agents.</p>
						<button class="cp-btn cp-btn--primary" onclick={() => (showPricingModal = true)}>
							Create Computer
						</button>
					</div>
				{/if}
			</div>

		</section>
	{/if} <!-- end desktop/dashboard toggle -->
	{/if} <!-- end main computer view -->

</div>

<style>
	/* ── View Tabs ─────────────────────────────────────────────────────────── */
	.cp-view-tabs {
		display: flex;
		gap: 2px;
		background: var(--dbg, #fff);
		border: 1px solid var(--dbd, #e5e7eb);
		border-radius: 8px;
		padding: 3px;
		width: fit-content;
	}

	.cp-view-tab {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 16px;
		border-radius: 8px;
		border: none;
		background: transparent;
		color: var(--dt3, #6b7280);
		font-size: 13px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.15s ease;
	}

	.cp-view-tab:hover { background: var(--dbg2, #f3f4f6); color: var(--dt, #111); }
	.cp-view-tab.active { background: var(--dt, #111); color: #fff; }
	:global(.dark) .cp-view-tab.active { background: #3b82f6; }

	/* ── Desktop Stream ───────────────────────────────────────────────────── */
	.cp-desktop-container {
		border-radius: 8px;
		overflow: hidden;
		border: 1px solid var(--dbd, #e5e7eb);
		background: #000;
		position: relative;
		flex: 1;
		min-height: 0;
	}

	.cp-desktop-iframe {
		width: 100%;
		height: 100%;
		border: none;
		display: block;
	}

	.cp-desktop-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		height: 500px;
		color: #888;
		gap: 12px;
	}

	.cp-desktop-spinner {
		width: 32px;
		height: 32px;
		border: 3px solid rgba(255,255,255,0.1);
		border-top-color: #7c3aed;
		border-radius: 50%;
		animation: cp-dspin 0.8s linear infinite;
	}

	@keyframes cp-dspin {
		to { transform: rotate(360deg); }
	}

	/* ── Page layout ───────────────────────────────────────────────────────── */
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

	/* ── Status Hero ───────────────────────────────────────────────────────── */
	.cp-hero {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		padding: 10px 16px;
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 8px;
		flex-wrap: wrap;
	}

	:global(.dark) .cp-hero {
		background: var(--dbg, rgba(255,255,255,0.03));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}

	.cp-hero-left {
		display: flex;
		align-items: center;
		gap: 16px;
	}

	.cp-status-indicator {
		width: 12px;
		height: 12px;
		border-radius: 50%;
		flex-shrink: 0;
	}

	.cp-status-indicator--running {
		background: #22c55e;
		box-shadow: 0 0 0 3px rgba(34, 197, 94, 0.2);
	}

	.cp-status-indicator--hibernating {
		background: #f59e0b;
		box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.2);
	}

	.cp-status-indicator--stopped {
		background: #9ca3af;
	}

	.cp-hero-info {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.cp-hero-title-row {
		display: flex;
		align-items: center;
		gap: 10px;
	}

	.cp-hero-title {
		font-size: 18px;
		font-weight: 600;
		color: var(--dt, #111827);
		margin: 0;
		letter-spacing: -0.02em;
	}

	:global(.dark) .cp-hero-title { color: var(--dt, #f3f4f6); }

	.cp-hero-domain {
		font-size: 13px;
		color: var(--dt3, #9ca3af);
		margin: 0;
		font-family: ui-monospace, 'SF Mono', 'Fira Code', monospace;
	}

	.cp-status-badge {
		font-size: 11px;
		font-weight: 600;
		padding: 3px 8px;
		border-radius: 20px;
		letter-spacing: 0.02em;
	}

	.cp-status-badge--running {
		background: rgba(34, 197, 94, 0.12);
		color: #16a34a;
	}
	:global(.dark) .cp-status-badge--running {
		background: rgba(34, 197, 94, 0.15);
		color: #4ade80;
	}

	.cp-status-badge--hibernating {
		background: rgba(245, 158, 11, 0.12);
		color: #d97706;
	}
	:global(.dark) .cp-status-badge--hibernating {
		background: rgba(245, 158, 11, 0.15);
		color: #fbbf24;
	}

	.cp-status-badge--stopped {
		background: rgba(156, 163, 175, 0.15);
		color: #6b7280;
	}
	:global(.dark) .cp-status-badge--stopped {
		background: rgba(156, 163, 175, 0.1);
		color: #9ca3af;
	}

	.cp-hero-actions {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}

	/* ── Buttons ───────────────────────────────────────────────────────────── */
	.cp-btn {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 500;
		padding: 8px 14px;
		border-radius: 8px;
		cursor: pointer;
		border: none;
		transition: background 140ms ease-out, color 140ms ease-out, opacity 140ms ease-out;
		text-decoration: none;
		line-height: 1;
		white-space: nowrap;
	}

	.cp-btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.cp-btn--primary {
		background: #3b82f6;
		color: #ffffff;
	}
	.cp-btn--primary:hover:not(:disabled) { background: #2563eb; }

	.cp-btn--outline {
		background: transparent;
		color: var(--dt2, #374151);
		border: 1px solid var(--dbd, rgba(0,0,0,0.1));
	}
	.cp-btn--outline:hover:not(:disabled) {
		background: var(--dbg3, rgba(0,0,0,0.04));
	}
	:global(.dark) .cp-btn--outline {
		color: var(--dt2, rgba(255,255,255,0.7));
		border-color: var(--dbd, rgba(255,255,255,0.1));
	}
	:global(.dark) .cp-btn--outline:hover:not(:disabled) {
		background: rgba(255,255,255,0.06);
	}

	.cp-btn--lg { padding: 12px 24px; font-size: 14px; }
	.cp-btn--full { width: 100%; justify-content: center; }

	/* ── Metrics grid ──────────────────────────────────────────────────────── */
	.cp-metrics-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 12px;
	}

	@media (max-width: 900px) {
		.cp-metrics-grid { grid-template-columns: 1fr; }
	}

	.cp-metric-card {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 12px;
		padding: 20px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		transition: border-color 140ms ease-out;
	}
	.cp-metric-card:hover { border-color: rgba(59, 130, 246, 0.2); }

	:global(.dark) .cp-metric-card {
		background: var(--dbg, rgba(255,255,255,0.03));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}
	:global(.dark) .cp-metric-card:hover {
		border-color: rgba(59, 130, 246, 0.25);
	}

	.cp-metric-header {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.cp-metric-icon {
		width: 28px;
		height: 28px;
		border-radius: 7px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.cp-metric-icon--ram { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
	.cp-metric-icon--cpu { background: rgba(139, 92, 246, 0.1); color: #8b5cf6; }
	.cp-metric-icon--storage { background: rgba(16, 185, 129, 0.1); color: #10b981; }

	.cp-metric-label {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt3, #9ca3af);
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.cp-metric-value {
		font-size: 22px;
		font-weight: 600;
		color: var(--dt, #111827);
		letter-spacing: -0.02em;
		line-height: 1;
	}
	:global(.dark) .cp-metric-value { color: var(--dt, #f3f4f6); }

	.cp-metric-pct {
		font-size: 12px;
		color: var(--dt3, #9ca3af);
		font-weight: 500;
	}

	.cp-metric-sub {
		font-size: 12px;
		color: var(--dt3, #9ca3af);
	}

	/* ── Progress bars ─────────────────────────────────────────────────────── */
	.cp-progress-track {
		height: 8px;
		background: var(--dbg2, rgba(0,0,0,0.06));
		border-radius: 100px;
		overflow: hidden;
	}

	:global(.dark) .cp-progress-track {
		background: rgba(255,255,255,0.08);
	}

	.cp-progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #8b5cf6);
		border-radius: 100px;
		transition: width 600ms cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	.cp-progress-fill--solid {
		/* color applied inline from $derived creditsColor */
		background: none;
		transition: width 600ms cubic-bezier(0.34, 1.56, 0.64, 1);
	}

	/* ── Bottom two-column grid ────────────────────────────────────────────── */
	.cp-bottom-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
		align-items: start;
	}

	@media (max-width: 900px) {
		.cp-bottom-grid { grid-template-columns: 1fr; }
	}

	/* ── Panel shared ──────────────────────────────────────────────────────── */
	.cp-panel {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 12px;
		overflow: hidden;
	}

	:global(.dark) .cp-panel {
		background: var(--dbg, rgba(255,255,255,0.03));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}

	.cp-panel-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 16px 20px 0;
	}

	.cp-panel-title {
		font-size: 13px;
		font-weight: 600;
		color: var(--dt2, #374151);
		text-transform: uppercase;
		letter-spacing: 0.06em;
		margin: 0;
	}
	:global(.dark) .cp-panel-title { color: var(--dt2, rgba(255,255,255,0.5)); }

	.cp-panel-count {
		font-size: 11px;
		color: var(--dt3, #9ca3af);
		font-weight: 500;
	}

	.cp-panel-footer {
		padding: 12px 20px 20px;
		border-top: 1px solid var(--dbd, rgba(0,0,0,0.06));
		margin-top: 4px;
	}
	:global(.dark) .cp-panel-footer {
		border-color: rgba(255,255,255,0.06);
	}

	/* ── Runtimes list ─────────────────────────────────────────────────────── */
	.cp-runtimes-list {
		padding: 12px 20px;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.cp-runtime-row {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 10px 12px;
		border-radius: 8px;
		transition: background 120ms ease-out;
	}
	.cp-runtime-row:hover { background: var(--dbg2, rgba(0,0,0,0.03)); }
	:global(.dark) .cp-runtime-row:hover { background: rgba(255,255,255,0.04); }

	.cp-runtime-left {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
	}

	.cp-runtime-status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex-shrink: 0;
	}
	.cp-runtime-status-dot--active { background: #22c55e; }
	.cp-runtime-status-dot--idle { background: #f59e0b; }
	.cp-runtime-status-dot--stopped { background: #9ca3af; }

	.cp-runtime-info {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.cp-runtime-name {
		font-size: 13px;
		font-weight: 500;
		color: var(--dt, #111827);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	:global(.dark) .cp-runtime-name { color: var(--dt, #f3f4f6); }

	.cp-runtime-meta {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.cp-runtime-status-badge {
		font-size: 10px;
		font-weight: 600;
		padding: 1px 6px;
		border-radius: 20px;
	}
	.cp-runtime-status-badge--active {
		background: rgba(34, 197, 94, 0.12);
		color: #16a34a;
	}
	:global(.dark) .cp-runtime-status-badge--active {
		background: rgba(34, 197, 94, 0.15);
		color: #4ade80;
	}
	.cp-runtime-status-badge--idle {
		background: rgba(245, 158, 11, 0.12);
		color: #d97706;
	}
	:global(.dark) .cp-runtime-status-badge--idle {
		background: rgba(245, 158, 11, 0.15);
		color: #fbbf24;
	}
	.cp-runtime-status-badge--stopped {
		background: rgba(156, 163, 175, 0.12);
		color: #6b7280;
	}

	.cp-runtime-stat {
		font-size: 11px;
		color: var(--dt3, #9ca3af);
	}

	.cp-runtime-sep {
		display: inline-block;
		width: 1px;
		height: 10px;
		background: var(--dbd, rgba(0,0,0,0.1));
	}
	:global(.dark) .cp-runtime-sep {
		background: rgba(255,255,255,0.1);
	}

	.cp-runtime-stop-btn {
		font-size: 11px;
		font-weight: 500;
		color: var(--dt3, #9ca3af);
		background: transparent;
		border: 1px solid var(--dbd, rgba(0,0,0,0.1));
		border-radius: 6px;
		padding: 4px 10px;
		cursor: pointer;
		transition: color 120ms, border-color 120ms, background 120ms;
		white-space: nowrap;
		flex-shrink: 0;
	}
	.cp-runtime-stop-btn:hover:not(:disabled) {
		color: #ef4444;
		border-color: rgba(239, 68, 68, 0.3);
		background: rgba(239, 68, 68, 0.05);
	}
	.cp-runtime-stop-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}
	:global(.dark) .cp-runtime-stop-btn {
		color: var(--dt3, rgba(255,255,255,0.35));
		border-color: rgba(255,255,255,0.08);
	}
	:global(.dark) .cp-runtime-stop-btn:hover:not(:disabled) {
		color: #f87171;
		border-color: rgba(248, 113, 113, 0.3);
		background: rgba(248, 113, 113, 0.06);
	}

	.cp-empty-list-text {
		font-size: 13px;
		color: var(--dt3, #9ca3af);
		text-align: center;
		padding: 24px 0;
		margin: 0;
	}

	/* ── Plan summary ──────────────────────────────────────────────────────── */
	.cp-plan-summary {
		padding: 16px 20px;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.cp-plan-summary-row {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 8px;
	}

	.cp-plan-summary-name {
		font-size: 18px;
		font-weight: 700;
		color: var(--dt, #111827);
		letter-spacing: -0.02em;
	}
	:global(.dark) .cp-plan-summary-name { color: var(--dt, #f3f4f6); }

	.cp-plan-summary-price {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111827);
		letter-spacing: -0.02em;
	}
	:global(.dark) .cp-plan-summary-price { color: var(--dt, #f3f4f6); }

	.cp-plan-summary-interval {
		font-size: 13px;
		font-weight: 400;
		color: var(--dt3, #9ca3af);
	}

	.cp-credits-section {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.cp-credits-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.cp-credits-label {
		font-size: 11px;
		font-weight: 600;
		color: var(--dt3, #9ca3af);
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.cp-credits-value {
		font-size: 12px;
		color: var(--dt2, #6b7280);
		font-weight: 500;
	}
	:global(.dark) .cp-credits-value { color: var(--dt2, rgba(255,255,255,0.5)); }

	.cp-plan-details {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}

	.cp-plan-detail {
		display: flex;
		align-items: center;
		gap: 7px;
		font-size: 12px;
		color: var(--dt3, #6b7280);
	}
	:global(.dark) .cp-plan-detail { color: var(--dt3, rgba(255,255,255,0.4)); }

	.cp-plan-actions {
		padding: 12px 20px 20px;
		border-top: 1px solid var(--dbd, rgba(0,0,0,0.06));
		display: flex;
		gap: 8px;
	}
	:global(.dark) .cp-plan-actions { border-color: rgba(255,255,255,0.06); }

	/* ── No subscription CTA ───────────────────────────────────────────────── */
	.cp-no-sub-cta {
		padding: 32px 24px;
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
		gap: 12px;
	}

	.cp-no-sub-icon {
		width: 56px;
		height: 56px;
		border-radius: 14px;
		background: rgba(59, 130, 246, 0.1);
		color: #3b82f6;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.cp-no-sub-title {
		font-size: 16px;
		font-weight: 600;
		color: var(--dt, #111827);
		margin: 0;
	}
	:global(.dark) .cp-no-sub-title { color: var(--dt, #f3f4f6); }

	.cp-no-sub-body {
		font-size: 13px;
		color: var(--dt3, #6b7280);
		margin: 0;
		max-width: 260px;
		line-height: 1.5;
	}
	:global(.dark) .cp-no-sub-body { color: var(--dt3, rgba(255,255,255,0.4)); }

	/* ── Empty / Free State ────────────────────────────────────────────────── */
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

	.cp-error-banner {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		width: 100%;
		padding: 10px 14px;
		border-radius: 8px;
		background: rgba(239, 68, 68, 0.1);
		border: 1px solid rgba(239, 68, 68, 0.25);
		color: #dc2626;
		font-size: 13px;
		margin-bottom: 12px;
	}

	.cp-error-banner button {
		background: none;
		border: none;
		color: #dc2626;
		font-weight: 600;
		cursor: pointer;
		font-size: 12px;
		padding: 2px 8px;
		border-radius: 4px;
	}

	.cp-error-banner button:hover {
		background: rgba(239, 68, 68, 0.15);
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

	.cp-empty-plans {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 10px;
		width: 100%;
		margin: 8px 0;
	}

	@media (max-width: 600px) {
		.cp-empty-plans { grid-template-columns: 1fr; }
	}

	.cp-empty-plan {
		background: var(--dbg2, rgba(0,0,0,0.03));
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 10px;
		padding: 14px 12px;
		display: flex;
		flex-direction: column;
		gap: 4px;
		position: relative;
	}
	:global(.dark) .cp-empty-plan {
		background: rgba(255,255,255,0.04);
		border-color: rgba(255,255,255,0.08);
	}

	.cp-empty-plan-name {
		font-size: 12px;
		font-weight: 600;
		color: var(--dt2, #374151);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	:global(.dark) .cp-empty-plan-name { color: var(--dt2, rgba(255,255,255,0.6)); }

	.cp-empty-plan-price {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111827);
		letter-spacing: -0.02em;
		margin-bottom: 6px;
	}
	:global(.dark) .cp-empty-plan-price { color: var(--dt, #f3f4f6); }

	.cp-empty-plan-price span {
		font-size: 12px;
		font-weight: 400;
		color: var(--dt3, #9ca3af);
	}

	.cp-empty-plan-specs {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.cp-empty-plan-specs li {
		font-size: 11px;
		color: var(--dt3, #6b7280);
	}
	:global(.dark) .cp-empty-plan-specs li { color: var(--dt3, rgba(255,255,255,0.4)); }

	.cp-empty-plan-badge {
		position: absolute;
		top: -8px;
		left: 50%;
		transform: translateX(-50%);
		font-size: 10px;
		font-weight: 700;
		padding: 2px 8px;
		border-radius: 20px;
		background: linear-gradient(90deg, #3b82f6, #8b5cf6);
		color: #ffffff;
		white-space: nowrap;
	}

	/* ── Pricing Modal ─────────────────────────────────────────────────────── */
	.cp-modal-backdrop {
		position: fixed;
		inset: 0;
		z-index: 100;
		background: rgba(0, 0, 0, 0.55);
		backdrop-filter: blur(6px);
		-webkit-backdrop-filter: blur(6px);
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 24px;
	}

	.cp-modal {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.1));
		border-radius: 16px;
		width: 100%;
		max-width: 860px;
		max-height: calc(100vh - 48px);
		overflow-y: auto;
		box-shadow: 0 24px 80px rgba(0, 0, 0, 0.18);
	}
	:global(.dark) .cp-modal {
		background: #1a1a1f;
		border-color: rgba(255,255,255,0.1);
		box-shadow: 0 24px 80px rgba(0, 0, 0, 0.5);
	}

	.cp-modal-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 16px;
		padding: 28px 28px 20px;
		border-bottom: 1px solid var(--dbd, rgba(0,0,0,0.07));
	}
	:global(.dark) .cp-modal-header { border-color: rgba(255,255,255,0.07); }

	.cp-modal-title {
		font-size: 20px;
		font-weight: 700;
		color: var(--dt, #111827);
		margin: 0 0 4px;
		letter-spacing: -0.02em;
	}
	:global(.dark) .cp-modal-title { color: var(--dt, #f3f4f6); }

	.cp-modal-subtitle {
		font-size: 13px;
		color: var(--dt3, #9ca3af);
		margin: 0;
	}

	.cp-modal-close {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: transparent;
		border: 1px solid var(--dbd, rgba(0,0,0,0.08));
		color: var(--dt3, #9ca3af);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		transition: background 120ms, color 120ms;
	}
	.cp-modal-close:hover {
		background: var(--dbg2, rgba(0,0,0,0.05));
		color: var(--dt, #111827);
	}
	:global(.dark) .cp-modal-close {
		border-color: rgba(255,255,255,0.08);
		color: rgba(255,255,255,0.4);
	}
	:global(.dark) .cp-modal-close:hover {
		background: rgba(255,255,255,0.06);
		color: #f3f4f6;
	}

	/* ── Plan cards ────────────────────────────────────────────────────────── */
	.cp-plans-grid {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 16px;
		padding: 24px 28px 28px;
	}

	@media (max-width: 700px) {
		.cp-plans-grid { grid-template-columns: 1fr; }
	}

	.cp-plan-card {
		background: var(--dbg2, rgba(0,0,0,0.03));
		border: 1px solid var(--dbd, rgba(0,0,0,0.08));
		border-radius: 12px;
		padding: 24px 20px 20px;
		display: flex;
		flex-direction: column;
		gap: 16px;
		position: relative;
		transition: border-color 140ms ease-out, box-shadow 140ms ease-out;
	}
	:global(.dark) .cp-plan-card {
		background: rgba(255,255,255,0.04);
		border-color: rgba(255,255,255,0.08);
	}

	.cp-plan-card--popular {
		border-color: #3b82f6;
		box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2);
	}
	:global(.dark) .cp-plan-card--popular {
		border-color: #3b82f6;
		box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.3);
	}

	.cp-plan-card--current {
		border-color: #22c55e;
		box-shadow: 0 0 0 1px rgba(34, 197, 94, 0.2);
	}

	.cp-plan-badge {
		position: absolute;
		top: -11px;
		left: 50%;
		transform: translateX(-50%);
		font-size: 10px;
		font-weight: 700;
		padding: 3px 10px;
		border-radius: 20px;
		white-space: nowrap;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.cp-plan-badge--popular {
		background: linear-gradient(90deg, #3b82f6, #8b5cf6);
		color: #ffffff;
	}

	.cp-plan-badge--current {
		background: rgba(34, 197, 94, 0.15);
		color: #16a34a;
		border: 1px solid rgba(34, 197, 94, 0.3);
	}
	:global(.dark) .cp-plan-badge--current {
		background: rgba(34, 197, 94, 0.2);
		color: #4ade80;
	}

	.cp-plan-header {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.cp-plan-name {
		font-size: 14px;
		font-weight: 700;
		color: var(--dt, #111827);
		margin: 0;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	:global(.dark) .cp-plan-name { color: var(--dt, #f3f4f6); }

	.cp-plan-price {
		display: flex;
		align-items: baseline;
		gap: 2px;
	}

	.cp-plan-price-amount {
		font-size: 28px;
		font-weight: 800;
		color: var(--dt, #111827);
		letter-spacing: -0.03em;
	}
	:global(.dark) .cp-plan-price-amount { color: var(--dt, #f3f4f6); }

	.cp-plan-price-period {
		font-size: 13px;
		color: var(--dt3, #9ca3af);
		font-weight: 400;
	}

	.cp-plan-specs {
		list-style: none;
		padding: 0;
		margin: 0;
		display: flex;
		flex-direction: column;
		gap: 8px;
		flex: 1;
	}

	.cp-plan-spec {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12.5px;
		color: var(--dt2, #4b5563);
	}
	:global(.dark) .cp-plan-spec { color: var(--dt2, rgba(255,255,255,0.6)); }

	.cp-plan-spec svg {
		flex-shrink: 0;
		opacity: 0.6;
	}

	.cp-plan-btn {
		width: 100%;
		padding: 10px;
		border-radius: 8px;
		font-size: 13px;
		font-weight: 600;
		cursor: pointer;
		border: none;
		transition: background 140ms, opacity 140ms;
	}

	.cp-plan-btn--upgrade {
		background: linear-gradient(135deg, #3b82f6, #8b5cf6);
		color: #ffffff;
	}
	.cp-plan-btn--upgrade:hover { opacity: 0.9; }

	.cp-plan-btn--current {
		background: rgba(34, 197, 94, 0.1);
		color: #16a34a;
		cursor: default;
		border: 1px solid rgba(34, 197, 94, 0.2);
	}
	:global(.dark) .cp-plan-btn--current {
		background: rgba(34, 197, 94, 0.12);
		color: #4ade80;
	}

	/* ── Computer Info card ────────────────────────────────────────────────── */
	.cp-info-card {
		background: var(--dbg, #ffffff);
		border: 1px solid var(--dbd, rgba(0,0,0,0.07));
		border-radius: 12px;
		overflow: hidden;
	}

	:global(.dark) .cp-info-card {
		background: var(--dbg, rgba(255,255,255,0.03));
		border-color: var(--dbd, rgba(255,255,255,0.07));
	}

	.cp-info-header {
		padding: 16px 20px 0;
	}

	.cp-info-grid {
		display: grid;
		grid-template-columns: repeat(4, 1fr);
		gap: 0;
		padding: 12px 20px 16px;
	}

	@media (max-width: 1100px) {
		.cp-info-grid { grid-template-columns: repeat(2, 1fr); }
	}

	@media (max-width: 600px) {
		.cp-info-grid { grid-template-columns: 1fr; }
	}

	.cp-info-row {
		display: flex;
		flex-direction: column;
		gap: 3px;
		padding: 10px 12px;
	}

	.cp-info-label {
		font-size: 10px;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.07em;
		color: var(--dt3, #9ca3af);
	}

	.cp-info-value {
		font-size: 13px;
		font-weight: 500;
		color: var(--dt, #111827);
	}
	:global(.dark) .cp-info-value { color: var(--dt, #f3f4f6); }

	.cp-info-value--mono {
		font-family: ui-monospace, 'SF Mono', 'Fira Code', monospace;
		font-size: 12px;
	}

	.cp-info-link {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		font-weight: 500;
		color: #3b82f6;
		text-decoration: none;
		font-family: ui-monospace, 'SF Mono', 'Fira Code', monospace;
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.cp-info-link:hover { text-decoration: underline; }
	:global(.dark) .cp-info-link { color: #60a5fa; }
</style>
