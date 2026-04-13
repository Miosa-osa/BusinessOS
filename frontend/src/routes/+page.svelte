<script lang="ts">
	import { browser } from '$app/environment';

	// Instant redirect — no flash, no waiting for JS framework
	if (browser) {
		const hasSession = document.cookie.includes('better-auth.session_token');
		window.location.replace(hasSession ? '/dashboard' : '/landing');
	}
</script>

<div style="min-height:100vh;background:#0a0a0e;"></div>
