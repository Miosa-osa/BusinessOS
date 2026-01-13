<!--
  Onboarding Page
  Conversational AI onboarding flow powered by Grok
  See ONBOARDING_ARCHITECTURE.md for full specification
-->
<script lang="ts">
	import { goto } from '$app/navigation';
	import { ConversationalOnboarding } from '$lib/components/onboarding';

	interface ExtractedData {
		workspaceName?: string;
		businessType?: string;
		teamSize?: string;
		role?: string;
		challenge?: string;
		integrations?: string[];
	}

	/**
	 * Handle onboarding completion (fallback for local mode)
	 * In API mode, the component handles redirect itself
	 */
	function handleComplete(data: ExtractedData) {
		// Store onboarding data
		localStorage.setItem('onboarding_data', JSON.stringify(data));
		localStorage.setItem('onboarding_completed', 'true');

		// Redirect to main app
		goto('/window');
	}
</script>

<!-- 
	useApi=true (default): Uses backend API for session management
	useApi=false: Uses local mock mode for testing without backend
-->
<ConversationalOnboarding onComplete={handleComplete} useApi={true} />
