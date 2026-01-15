<!--
  ConversationalOnboarding.svelte
  Main conversational onboarding flow with AI agent
  Hybrid: Chips for quick-select + Chat input for open questions
  
  Supports two modes:
  1. API Mode (default): Uses backend API for session management
  2. Local Mode: Self-contained with mock responses (for testing/fallback)
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { fly } from 'svelte/transition';
	import PurpleOrb from './PurpleOrb.svelte';
	import SequentialTypewriter from './SequentialTypewriter.svelte';
	import IntegrationCard from './IntegrationCard.svelte';
	import { onboardingApi, type ExtractedOnboardingData } from '$lib/api/onboarding';

	type OnboardingPhase = 
		| 'loading'
		| 'intro'
		| 'conversation'
		| 'integrations';

	type QuestionType = 
		| 'company_name'      // chat input
		| 'business_type'     // chips
		| 'team_size'         // chips (skip if freelance)
		| 'role'              // chat input
		| 'challenge'         // chat input
		| 'integrations'
		| 'complete';

	interface Message {
		id: string;
		sender: 'agent' | 'user';
		content: string;
		timestamp: Date;
	}

	interface ExtractedData {
		workspaceName?: string;
		businessType?: string;
		teamSize?: string;
		role?: string;
		challenge?: string;
		integrations?: string[];
	}

	interface ChipOption {
		id: string;
		label: string;
		icon?: string;
	}

	interface Props {
		sessionId?: string;
		onComplete?: (data: ExtractedData) => void;
		useApi?: boolean; // Enable API integration
		class?: string;
	}

	let {
		sessionId: initialSessionId,
		onComplete,
		useApi = true, // Default to API mode
		class: className = ''
	}: Props = $props();

	// Session state (for API mode)
	let sessionId = $state<string | null>(initialSessionId || null);
	let apiError = $state<string | null>(null);
	let sessionExpired = $state(false);
	let lastErrorAction = $state<(() => void) | null>(null);

	// History tracking for Go Back
	interface HistoryEntry {
		question: QuestionType;
		data: ExtractedData;
		agentMessage: string;
	}
	let questionHistory = $state<HistoryEntry[]>([]);

	// State
	let phase = $state<OnboardingPhase>(useApi ? 'loading' : 'intro');
	let currentQuestion = $state<QuestionType>('company_name');
	let messages = $state<Message[]>([]);
	let isAgentTyping = $state(false);
	let extractedData = $state<ExtractedData>({});
	let introComplete = $state(false);
	let currentAgentMessage = $state('');

	// Can go back derived (after phase and isAgentTyping are declared)
	let canGoBack = $derived(questionHistory.length > 0 && phase === 'conversation' && !isAgentTyping);

	// Integration state
	let selectedIntegrations = $state<string[]>([]);
	let integrationStatuses = $state<Record<string, 'disconnected' | 'connecting' | 'connected' | 'error'>>({});
	let recommendedIntegrations = $state<string[]>([]);

	// Derived: check if all recommended are connected
	let allRecommendedConnected = $derived(
		recommendedIntegrations.length > 0 &&
		recommendedIntegrations.every(id => integrationStatuses[id] === 'connected')
	);

	// Input state
	let inputValue = $state('');
	let isRecording = $state(false);
	let inputRef = $state<HTMLInputElement | null>(null);

	// Resume state for welcome back flow
	let isResuming = $state(false);
	let resumeMessage = $state('');

	// OAuth error state for integration cards
	let oauthError = $state<string | null>(null);
	let failedIntegrationId = $state<string | null>(null);

	// Chip options
	const businessTypeOptions: ChipOption[] = [
		{ id: 'agency', label: 'Agency' },
		{ id: 'startup', label: 'Startup' },
		{ id: 'freelance', label: 'Freelance' },
		{ id: 'ecommerce', label: 'E-commerce' },
		{ id: 'consulting', label: 'Consulting' },
		{ id: 'other', label: 'Other' }
	];

	const teamSizeOptions: ChipOption[] = [
		{ id: 'solo', label: 'Just me' },
		{ id: '2-5', label: '2-5' },
		{ id: '6-15', label: '6-15' },
		{ id: '16-50', label: '16-50' },
		{ id: '50+', label: '50+' }
	];

	// Questions config - maps API step names to display
	const questions: Record<QuestionType, { message: string; inputType: 'chat' | 'chips'; chips?: ChipOption[] }> = {
		company_name: {
			message: "What's your company called?",
			inputType: 'chat'
		},
		business_type: {
			message: "What kind of work do you do?",
			inputType: 'chips',
			chips: businessTypeOptions
		},
		team_size: {
			message: "How big is your team?",
			inputType: 'chips',
			chips: teamSizeOptions
		},
		role: {
			message: "What's your role?",
			inputType: 'chat'
		},
		challenge: {
			message: "What's the biggest challenge you're hoping to solve?",
			inputType: 'chat'
		},
		integrations: {
			message: "Perfect! Let's connect your favorite tools.",
			inputType: 'chat'
		},
		complete: {
			message: "Perfect! Let's connect your favorite tools.",
			inputType: 'chat'
		}
	};

	// Get current question config
	let currentQuestionConfig = $derived(questions[currentQuestion] || questions.company_name);

	// Computed current step for progress indicator
	let currentStep = $derived(
		phase === 'loading' ? 0 :
		phase === 'intro' ? 1 : 
		phase === 'conversation' ? 2 : 
		3
	);

	// Simplified intro
	const introLines = [
		"Hi! I'm here to help set up your workspace.",
		"What's your company called?"
	];

	// Initialize API session on mount
	onMount(async () => {
		if (!useApi) return;

		try {
			// Check if user needs onboarding
			const status = await onboardingApi.checkStatus();
			
			if (!status.needs_onboarding) {
				// Already has workspace, redirect
				goto('/window');
				return;
			}

			// Check for resumeable session
			const resumeResult = await onboardingApi.getResumeableSession();
			
			if (resumeResult.has_session && resumeResult.session) {
				// Resume existing session
				sessionId = resumeResult.session.id;
				currentQuestion = (resumeResult.session.current_step as QuestionType) || 'company_name';
				
				// Convert API extracted data to local format
				if (resumeResult.session.extracted_data) {
					const apiData = resumeResult.session.extracted_data;
					extractedData = {
						workspaceName: apiData.workspace_name,
						businessType: apiData.business_type,
						teamSize: apiData.team_size,
						role: apiData.role,
						challenge: apiData.challenge,
						integrations: apiData.integrations
					};
				}
				
				// Build welcome back message with context
				if (currentQuestion !== 'company_name') {
					isResuming = true;
					const parts = ['Welcome back!'];
					if (extractedData.workspaceName) {
						parts.push(`Setting up ${extractedData.workspaceName}`);
						if (extractedData.businessType) {
							parts[1] += ` - ${extractedData.businessType}`;
						}
						parts[1] += '.';
					}
					resumeMessage = parts.join(' ');
				}
				
				currentAgentMessage = questions[currentQuestion]?.message || '';
				phase = currentQuestion === 'integrations' || currentQuestion === 'complete' ? 'integrations' : 'conversation';
			} else {
				// Create new session
				const { session, messages: apiMessages } = await onboardingApi.createSession();
				sessionId = session.id;
				phase = 'intro';
			}
		} catch (error) {
			console.error('Failed to initialize onboarding:', error);
			const errorMessage = error instanceof Error ? error.message : 'Failed to start onboarding';
			
			if (errorMessage.toLowerCase().includes('expired') || 
				errorMessage.toLowerCase().includes('invalid session') ||
				errorMessage.toLowerCase().includes('session not found')) {
				sessionExpired = true;
				apiError = 'Your session has expired. Starting fresh...';
				// Auto-create new session after a brief delay
				setTimeout(async () => {
					try {
						const { session } = await onboardingApi.createSession();
						sessionId = session.id;
						sessionExpired = false;
						apiError = null;
						phase = 'intro';
					} catch {
						apiError = 'Unable to start a new session. Please refresh the page.';
					}
				}, 2000);
			} else {
				apiError = errorMessage;
			}
			// Fall back to local mode
			phase = 'intro';
		}
	});

	// Available integrations
	const integrations = [
		{ id: 'google', name: 'Google Workspace', icon: '<svg viewBox="0 0 24 24"><path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/><path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/><path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/><path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/></svg>' },
		{ id: 'microsoft', name: 'Microsoft 365', icon: '<svg viewBox="0 0 24 24"><rect x="1" y="1" width="10" height="10" fill="#F25022"/><rect x="13" y="1" width="10" height="10" fill="#7FBA00"/><rect x="1" y="13" width="10" height="10" fill="#00A4EF"/><rect x="13" y="13" width="10" height="10" fill="#FFB900"/></svg>' },
		{ id: 'slack', name: 'Slack', icon: '<svg viewBox="0 0 24 24"><path fill="#E01E5A" d="M5.042 15.165a2.528 2.528 0 0 1-2.52 2.523A2.528 2.528 0 0 1 0 15.165a2.527 2.527 0 0 1 2.522-2.52h2.52v2.52z"/><path fill="#E01E5A" d="M6.313 15.165a2.527 2.527 0 0 1 2.521-2.52 2.527 2.527 0 0 1 2.521 2.52v6.313A2.528 2.528 0 0 1 8.834 24a2.528 2.528 0 0 1-2.521-2.522v-6.313z"/><path fill="#36C5F0" d="M8.834 5.042a2.528 2.528 0 0 1-2.521-2.52A2.528 2.528 0 0 1 8.834 0a2.528 2.528 0 0 1 2.521 2.522v2.52H8.834z"/><path fill="#36C5F0" d="M8.834 6.313a2.528 2.528 0 0 1 2.521 2.521 2.528 2.528 0 0 1-2.521 2.521H2.522A2.528 2.528 0 0 1 0 8.834a2.528 2.528 0 0 1 2.522-2.521h6.312z"/><path fill="#2EB67D" d="M18.958 8.834a2.528 2.528 0 0 1 2.52-2.521A2.528 2.528 0 0 1 24 8.834a2.528 2.528 0 0 1-2.522 2.521h-2.52V8.834z"/><path fill="#2EB67D" d="M17.687 8.834a2.528 2.528 0 0 1-2.521 2.521 2.528 2.528 0 0 1-2.521-2.521V2.522A2.528 2.528 0 0 1 15.166 0a2.528 2.528 0 0 1 2.521 2.522v6.312z"/><path fill="#ECB22E" d="M15.166 18.958a2.528 2.528 0 0 1 2.521 2.52A2.528 2.528 0 0 1 15.166 24a2.528 2.528 0 0 1-2.521-2.522v-2.52h2.521z"/><path fill="#ECB22E" d="M15.166 17.687a2.528 2.528 0 0 1-2.521-2.521 2.528 2.528 0 0 1 2.521-2.521h6.312A2.528 2.528 0 0 1 24 15.166a2.528 2.528 0 0 1-2.522 2.521h-6.312z"/></svg>' },
		{ id: 'notion', name: 'Notion', icon: '<svg viewBox="0 0 24 24"><path fill="currentColor" d="M4.459 4.208c.746.606 1.026.56 2.428.466l13.215-.793c.28 0 .047-.28-.046-.326L17.86 2.02c-.42-.327-.98-.7-2.053-.607L3.01 2.73c-.466.046-.56.28-.374.466zm.793 3.08v13.904c0 .747.373 1.027 1.213.98l14.523-.84c.84-.047.933-.56.933-1.167V6.354c0-.606-.233-.933-.746-.886l-15.177.887c-.56.046-.746.326-.746.933zm14.337.745c.093.42 0 .84-.42.888l-.7.14v10.264c-.608.327-1.168.514-1.635.514-.746 0-.932-.234-1.494-.933l-4.577-7.186v6.952L12.21 19s0 .84-1.168.84l-3.222.186c-.093-.186 0-.653.327-.746l.84-.233V9.854L7.822 9.76c-.094-.42.14-1.026.793-1.073l3.456-.233 4.764 7.279v-6.44l-1.215-.14c-.093-.514.28-.886.747-.933zM2.571 1.027l13.636-1c1.634-.14 2.053-.047 3.082.7l4.249 2.986c.7.513.933.653.933 1.213v16.378c0 1.026-.373 1.634-1.68 1.726l-15.458.934c-.98.047-1.447-.093-1.96-.747l-3.128-4.053c-.56-.747-.793-1.306-.793-1.96V2.667c0-.839.374-1.54 1.119-1.64z"/></svg>' },
		{ id: 'linear', name: 'Linear', icon: '<svg viewBox="0 0 24 24"><path fill="#5E6AD2" d="M3.357 3.357a12.166 12.166 0 0 0-1.473 2.195l16.564 16.564a12.166 12.166 0 0 0 2.195-1.473L3.357 3.357zm-1.947 3.49A12.147 12.147 0 0 0 .69 9.75L14.25 23.31a12.147 12.147 0 0 0 2.903-.72L1.41 6.847zm-.655 4.18a12.244 12.244 0 0 0-.065 1.973l7.31 7.31c.656.03 1.317-.003 1.973-.065L.755 11.027zm.025 3.528 5.618 5.618a12.198 12.198 0 0 1-5.618-5.618zm19.863-3.198L14.027.755a12.244 12.244 0 0 0-1.973.065l9.218 9.218c.062-.656.095-1.317.065-1.973zm.047-1.555a12.147 12.147 0 0 0-.72-2.903L6.847 1.41a12.147 12.147 0 0 0-2.903.72L19.688 17.87zm1.06 3.485a12.166 12.166 0 0 0 1.473-2.195L5.959 2.528a12.166 12.166 0 0 0-2.195 1.473l17.286 17.286z"/></svg>' },
		{ id: 'hubspot', name: 'HubSpot', icon: '<svg viewBox="0 0 24 24"><path fill="#FF7A59" d="M18.164 7.93V5.084a2.198 2.198 0 0 0 1.267-1.984v-.066A2.198 2.198 0 0 0 17.233.836h-.066a2.198 2.198 0 0 0-2.198 2.198v.066c0 .867.503 1.617 1.232 1.974v2.862a6.175 6.175 0 0 0-2.868 1.165l-7.62-5.937a2.636 2.636 0 1 0-1.168 1.49l7.322 5.705a6.205 6.205 0 0 0 .035 6.103l-2.22 2.22a2.092 2.092 0 1 0 1.131 1.131l2.234-2.234a6.2 6.2 0 1 0 5.087-9.649zm-1.031 9.544a3.343 3.343 0 1 1 0-6.687 3.343 3.343 0 0 1 0 6.687z"/></svg>' },
		{ id: 'airtable', name: 'Airtable', icon: '<svg viewBox="0 0 24 24"><path fill="#FCB400" d="M11.992 2 2 6.462v11.078L11.992 22l9.992-4.462V6.462L11.992 2z"/><path fill="#18BFFF" d="M12 2.016 2.016 6.462v.082l9.976 4.362 9.992-4.362-.008-.082L12 2.016z"/><path fill="#F82B60" d="m2 6.544-.008 11.002L11.984 22V10.924L2 6.544z"/><path fill="#666" d="M22 6.544 12.016 22V10.924L22 6.544z" opacity=".25"/></svg>' },
		{ id: 'clickup', name: 'ClickUp', icon: '<svg viewBox="0 0 24 24"><path fill="#7B68EE" d="m3.986 14.446 2.824-2.173a5.568 5.568 0 0 0 5.19 3.452 5.568 5.568 0 0 0 5.19-3.452l2.824 2.173A9.141 9.141 0 0 1 12 19.725a9.141 9.141 0 0 1-8.014-5.28z"/><path fill="#49CCF9" d="m4 9.247 2.778 2.135 5.222-4.28 5.222 4.28L20 9.247 12 3 4 9.247z"/></svg>' },
		{ id: 'fathom', name: 'Fathom', icon: '<svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="10" fill="#7C3AED"/><path fill="white" d="M8 12a4 4 0 0 1 8 0v4H8v-4z"/></svg>' }
	];

	// Skip to integrations
	async function skipToIntegrations() {
		phase = 'integrations';
		
		// Try to get recommendations from API if available
		if (useApi && sessionId) {
			try {
				const recs = await onboardingApi.getRecommendations(sessionId);
				if (recs?.length) {
					recommendedIntegrations = recs;
					return;
				}
			} catch (e) {
				console.warn('Failed to get recommendations from API, using local fallback');
			}
		}
		
		// Fallback to local computation
		computeRecommendedIntegrations();
	}

	// Auto-focus input when showing chat input questions
	$effect(() => {
		if (phase === 'conversation' && currentQuestionConfig.inputType === 'chat' && inputRef && !isAgentTyping) {
			setTimeout(() => inputRef?.focus(), 100);
		}
	});

	// Handle form submit for chat input questions
	function handleFormSubmit(e: Event) {
		e.preventDefault();
		if (inputValue.trim() && !isAgentTyping) {
			handleChatAnswer(inputValue.trim());
			inputValue = '';
		}
	}

	// Handle voice input toggle
	function toggleVoiceInput() {
		if (isRecording) {
			isRecording = false;
		} else {
			isRecording = true;
		}
	}

	// Handle chip selection
	async function handleChipSelect(chipId: string) {
		if (useApi && sessionId) {
			// API mode: send to backend
			isAgentTyping = true;
			// Save current state before advancing
			saveToHistory();
			// Store retry action
			lastErrorAction = () => handleChipSelect(chipId);
			try {
				const response = await onboardingApi.sendMessage(sessionId, chipId);
				lastErrorAction = null; // Clear on success
				
				// Update local state from API response
				const apiData = response.extracted_data;
				extractedData = {
					workspaceName: apiData.workspace_name,
					businessType: apiData.business_type,
					teamSize: apiData.team_size,
					role: apiData.role,
					challenge: apiData.challenge,
					integrations: apiData.integrations
				};
				
				// Get the AI message BEFORE changing currentQuestion
				const aiMessage = response.message?.content || '';
				const nextStep = response.next_step as QuestionType;
				
				if (response.recommended_integrations?.length) {
					recommendedIntegrations = response.recommended_integrations;
				}
				
				// Move to integrations if complete - show transition message first
				if (nextStep === 'integrations' || nextStep === 'complete') {
					currentAgentMessage = aiMessage || "Perfect! Let me recommend some tools for you.";
					setTimeout(() => {
						phase = 'integrations';
					}, 2000);
				} else {
					// Normal flow - advance to next question
					currentQuestion = nextStep;
					currentAgentMessage = aiMessage || questions[nextStep]?.message || '';
				}
			} catch (error) {
				console.error('Failed to process chip selection:', error);
				apiError = error instanceof Error ? error.message : 'Failed to process selection';
			} finally {
				isAgentTyping = false;
			}
		} else {
			// Local mode: use deterministic logic
			if (currentQuestion === 'business_type') {
				extractedData = { ...extractedData, businessType: chipId };
				
				// Branching: If freelance, skip team size
				if (chipId === 'freelance') {
					extractedData = { ...extractedData, teamSize: 'solo' };
					advanceToQuestion('role');
				} else {
					advanceToQuestion('team_size');
				}
			} else if (currentQuestion === 'team_size') {
				extractedData = { ...extractedData, teamSize: chipId };
				advanceToQuestion('role');
			}
		}
	}

	// Handle chat answer
	async function handleChatAnswer(answer: string) {
		if (useApi && sessionId) {
			// API mode: send to backend
			isAgentTyping = true;
			// Save current state before advancing
			saveToHistory();
			// Store retry action
			lastErrorAction = () => handleChatAnswer(answer);
			try {
				const response = await onboardingApi.sendMessage(sessionId, answer);
				lastErrorAction = null; // Clear on success
				
				// Update local state from API response
				const apiData = response.extracted_data;
				extractedData = {
					workspaceName: apiData.workspace_name,
					businessType: apiData.business_type,
					teamSize: apiData.team_size,
					role: apiData.role,
					challenge: apiData.challenge,
					integrations: apiData.integrations
				};
				
				// Get the AI message BEFORE changing currentQuestion
				const aiMessage = response.message?.content || '';
				const nextStep = response.next_step as QuestionType;
				
				if (response.recommended_integrations?.length) {
					recommendedIntegrations = response.recommended_integrations;
				}
				
				// Move to integrations if complete - show the transition message first!
				if (nextStep === 'integrations' || nextStep === 'complete') {
					// Show the AI's transition message (e.g., "I hear you! Based on what you've shared...")
					currentAgentMessage = aiMessage || "Perfect! Based on what you've shared, let me recommend some tools.";
					// Don't change currentQuestion yet - keep showing conversation
					
					// Wait for user to read, then transition
					setTimeout(() => {
						phase = 'integrations';
					}, 2500);
				} else {
					// Normal flow - advance to next question
					currentQuestion = nextStep;
					currentAgentMessage = aiMessage || questions[nextStep]?.message || '';
				}
			} catch (error) {
				console.error('Failed to process message:', error);
				apiError = error instanceof Error ? error.message : 'Failed to process message';
			} finally {
				isAgentTyping = false;
			}
		} else {
			// Local mode: use deterministic logic
			if (currentQuestion === 'company_name') {
				extractedData = { ...extractedData, workspaceName: answer };
				advanceToQuestion('business_type');
			} else if (currentQuestion === 'role') {
				extractedData = { ...extractedData, role: answer };
				advanceToQuestion('challenge');
			} else if (currentQuestion === 'challenge') {
				extractedData = { ...extractedData, challenge: answer };
				advanceToQuestion('complete');
			}
		}
	}

	// Advance to next question with animation
	function advanceToQuestion(nextQuestion: QuestionType) {
		isAgentTyping = true;
		
		setTimeout(() => {
			currentQuestion = nextQuestion;
			currentAgentMessage = questions[nextQuestion]?.message || '';
			isAgentTyping = false;
			
			// If complete, go to integrations
			if (nextQuestion === 'complete' || nextQuestion === 'integrations') {
				computeRecommendedIntegrations();
				setTimeout(() => {
					phase = 'integrations';
				}, 1500);
			}
		}, 800);
	}

	// Compute recommended integrations based on answers
	function computeRecommendedIntegrations() {
		const challenge = extractedData.challenge?.toLowerCase() || '';
		const businessType = extractedData.businessType || '';
		
		// Challenge-based recommendations
		if (challenge.includes('organiz') || challenge.includes('chaos') || challenge.includes('mess')) {
			recommendedIntegrations = ['notion', 'google', 'linear'];
		} else if (challenge.includes('scale') || challenge.includes('grow') || challenge.includes('automat')) {
			recommendedIntegrations = ['linear', 'slack', 'airtable'];
		} else if (challenge.includes('client') || challenge.includes('customer') || challenge.includes('crm')) {
			recommendedIntegrations = ['hubspot', 'slack', 'google'];
		} else if (challenge.includes('team') || challenge.includes('collaborat') || challenge.includes('communic')) {
			recommendedIntegrations = ['slack', 'notion', 'linear'];
		} else if (challenge.includes('time') || challenge.includes('busy') || challenge.includes('meeting')) {
			recommendedIntegrations = ['google', 'fathom', 'slack'];
		} else {
			// Default by business type
			if (businessType === 'agency' || businessType === 'consulting') {
				recommendedIntegrations = ['hubspot', 'slack', 'notion'];
			} else if (businessType === 'startup') {
				recommendedIntegrations = ['linear', 'slack', 'notion'];
			} else if (businessType === 'freelance') {
				recommendedIntegrations = ['google', 'notion', 'fathom'];
			} else {
				recommendedIntegrations = ['google', 'slack', 'notion'];
			}
		}
	}

	// Handle intro completion
	function handleIntroComplete() {
		introComplete = true;
		currentAgentMessage = questions.company_name.message;
		setTimeout(() => {
			phase = 'conversation';
		}, 300);
	}

	// Generate unique ID
	function generateId(): string {
		return Math.random().toString(36).substring(2, 9);
	}

	// Add message to chat
	function addMessage(sender: 'agent' | 'user', content: string) {
		messages = [...messages, {
			id: generateId(),
			sender,
			content,
			timestamp: new Date()
		}];
	}

	// Dismiss error banner
	function dismissError() {
		apiError = null;
		oauthError = null;
		lastErrorAction = null;
	}

	// Retry last failed action
	function retryLastAction() {
		if (lastErrorAction) {
			apiError = null;
			oauthError = null;
			lastErrorAction();
			lastErrorAction = null;
		} else if (failedIntegrationId) {
			// Retry the OAuth connection
			oauthError = null;
			handleIntegrationConnect(failedIntegrationId);
			failedIntegrationId = null;
		}
	}

	// Go back to previous question
	function goBack() {
		if (questionHistory.length === 0 || phase !== 'conversation') return;
		
		const lastEntry = questionHistory[questionHistory.length - 1];
		questionHistory = questionHistory.slice(0, -1);
		
		// Restore the previous state
		currentQuestion = lastEntry.question;
		extractedData = lastEntry.data;
		currentAgentMessage = lastEntry.agentMessage;
	}

	// Save current state to history before advancing
	function saveToHistory() {
		questionHistory = [...questionHistory, {
			question: currentQuestion,
			data: { ...extractedData },
			agentMessage: currentAgentMessage
		}];
	}

	// Handle integration connect via OAuth popup
	async function handleIntegrationConnect(integrationId: string) {
		integrationStatuses[integrationId] = 'connecting';
		
		try {
			// Get the API base URL
			const apiBase = getApiBase();
			
			// Map integration ID to provider path
			const providerMap: Record<string, string> = {
				'google': 'google',
				'microsoft': 'microsoft',
				'slack': 'slack',
				'notion': 'notion',
				'linear': 'linear',
				'hubspot': 'hubspot',
				'airtable': 'airtable',
				'clickup': 'clickup',
				'fathom': 'fathom'
			};
			
			const provider = providerMap[integrationId];
			if (!provider) {
				throw new Error(`Unknown integration: ${integrationId}`);
			}
			
			// Store onboarding context in localStorage for callback
			localStorage.setItem('onboarding_oauth_provider', integrationId);
			localStorage.setItem('onboarding_session_id', sessionId || '');
			
			// Open OAuth flow in popup window
			const width = 600;
			const height = 700;
			const left = window.screenX + (window.outerWidth - width) / 2;
			const top = window.screenY + (window.outerHeight - height) / 2;
			
			const authUrl = `${apiBase}/integrations/${provider}/auth`;
			const popup = window.open(
				authUrl,
				'oauth_popup',
				`width=${width},height=${height},left=${left},top=${top},scrollbars=yes,resizable=yes`
			);
			
			// Poll for popup close or success
			const pollInterval = setInterval(() => {
				if (!popup || popup.closed) {
					clearInterval(pollInterval);
					// Check if connection succeeded
					checkIntegrationStatus(integrationId);
				}
			}, 500);
			
			// Also listen for postMessage from callback page
			const handleMessage = (event: MessageEvent) => {
				if (event.data?.type === 'oauth_callback' && event.data?.provider === integrationId) {
					clearInterval(pollInterval);
					window.removeEventListener('message', handleMessage);
					
					if (event.data.success) {
						integrationStatuses[integrationId] = 'connected';
						oauthError = null;
						failedIntegrationId = null;
						if (!selectedIntegrations.includes(integrationId)) {
							selectedIntegrations = [...selectedIntegrations, integrationId];
						}
					} else {
						integrationStatuses[integrationId] = 'error';
						failedIntegrationId = integrationId;
						oauthError = event.data.error || `Failed to connect ${integrations.find(i => i.id === integrationId)?.name || integrationId}`;
					}
				}
			};
			window.addEventListener('message', handleMessage);
			
			// Cleanup after 5 minutes max
			setTimeout(() => {
				clearInterval(pollInterval);
				window.removeEventListener('message', handleMessage);
				if (integrationStatuses[integrationId] === 'connecting') {
					integrationStatuses[integrationId] = 'disconnected';
					oauthError = 'Connection timed out. Please try again.';
					failedIntegrationId = integrationId;
				}
			}, 5 * 60 * 1000);
			
		} catch (error) {
			console.error('Failed to connect integration:', error);
			integrationStatuses[integrationId] = 'error';
			failedIntegrationId = integrationId;
			oauthError = error instanceof Error ? error.message : 'Failed to start OAuth connection';
		}
	}
	
	// Check integration status via API
	async function checkIntegrationStatus(integrationId: string) {
		try {
			const apiBase = getApiBase();
			const response = await fetch(`${apiBase}/integrations/${integrationId}/status`, {
				credentials: 'include'
			});
			
			if (response.ok) {
				const data = await response.json();
				if (data.connected) {
					integrationStatuses[integrationId] = 'connected';
					if (!selectedIntegrations.includes(integrationId)) {
						selectedIntegrations = [...selectedIntegrations, integrationId];
					}
				} else {
					integrationStatuses[integrationId] = 'disconnected';
				}
			}
		} catch (error) {
			console.error('Failed to check integration status:', error);
			if (integrationStatuses[integrationId] === 'connecting') {
				integrationStatuses[integrationId] = 'disconnected';
			}
		}
	}
	
	// Helper to get API base URL
	function getApiBase(): string {
		if (typeof window === 'undefined') {
			return import.meta.env.VITE_API_URL || '/api';
		}
		const isElectron = 'electron' in window;
		if (isElectron) {
			const mode = localStorage.getItem('businessos_mode');
			const cloudUrl = localStorage.getItem('businessos_cloud_url');
			if (mode === 'cloud' && cloudUrl) return `${cloudUrl}/api`;
			if (mode === 'local') return 'http://localhost:18080/api';
			return 'http://localhost:8001/api';
		}
		return import.meta.env.VITE_API_URL || '/api';
	}

	// Handle integration disconnect
	function handleIntegrationDisconnect(integrationId: string) {
		integrationStatuses[integrationId] = 'disconnected';
		selectedIntegrations = selectedIntegrations.filter(id => id !== integrationId);
	}

	// Connect all recommended integrations
	function connectAllRecommended() {
		for (const id of recommendedIntegrations) {
			if (integrationStatuses[id] !== 'connected' && integrationStatuses[id] !== 'connecting') {
				handleIntegrationConnect(id);
			}
		}
	}

	// Complete integrations and finish onboarding
	async function completeIntegrations() {
		// Add integrations to extracted data
		extractedData = {
			...extractedData,
			integrations: selectedIntegrations
		};
		
		if (useApi && sessionId) {
			// API mode: complete via backend
			try {
				const result = await onboardingApi.completeOnboarding(sessionId, selectedIntegrations);
				
				// Store completion data
				localStorage.setItem('onboarding_completed', 'true');
				localStorage.setItem('workspace_id', result.workspace_id);
				
				// Redirect to dashboard
				goto(result.redirect_url || '/window');
			} catch (error) {
				console.error('Failed to complete onboarding:', error);
				apiError = error instanceof Error ? error.message : 'Failed to complete';
				
				// Store what we have and redirect anyway - don't leave user stuck
				localStorage.setItem('onboarding_completed', 'true');
				localStorage.setItem('onboarding_data', JSON.stringify(extractedData));
				
				// Fall back to client-side completion with redirect
				if (onComplete) {
					onComplete(extractedData);
				} else {
					// Direct redirect as final fallback
					goto('/window');
				}
			}
		} else {
			// Local mode: call callback or redirect directly
			if (onComplete) {
				onComplete(extractedData);
			} else {
				goto('/window');
			}
		}
	}
</script>

<div class="onboarding-screen {className}">
	<!-- Progress indicator -->
	<div class="progress-dots">
		<span class="dot" class:active={currentStep >= 1} class:current={currentStep === 1}></span>
		<span class="dot" class:active={currentStep >= 2} class:current={currentStep === 2}></span>
		<span class="dot" class:active={currentStep >= 3} class:current={currentStep === 3}></span>
	</div>

	<!-- Error Banner -->
	{#if apiError || oauthError}
		<div class="error-banner" transition:fly={{ y: -20, duration: 300 }}>
			<div class="error-content">
				<svg class="error-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"/>
					<line x1="12" y1="8" x2="12" y2="12"/>
					<line x1="12" y1="16" x2="12.01" y2="16"/>
				</svg>
				<span class="error-message">{apiError || oauthError}</span>
			</div>
			<div class="error-actions">
				{#if lastErrorAction || failedIntegrationId}
					<button class="error-btn retry" onclick={retryLastAction}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="23 4 23 10 17 10"/>
							<path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
						</svg>
						Retry
					</button>
				{/if}
				<button class="error-btn dismiss" onclick={dismissError} aria-label="Dismiss error">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"/>
						<line x1="6" y1="6" x2="18" y2="18"/>
					</svg>
				</button>
			</div>
		</div>
	{/if}

	{#if phase === 'loading'}
		<!-- Loading state -->
		<div class="centered-layout">
			<div class="orb-section">
				<PurpleOrb size="lg" isThinking={true} />
			</div>
			<div class="text-section">
				<p class="agent-text">Setting things up...</p>
			</div>
		</div>
	{:else if phase === 'intro' || phase === 'conversation'}
		<!-- Skip button - only show during conversation, not intro -->
		{#if phase === 'conversation' && !isResuming}
			<button class="skip-btn" onclick={skipToIntegrations} title="Skip questions and go to integrations">
				Skip for now
			</button>
		{/if}

		<!-- Centered layout for intro and conversation -->
		<div class="centered-layout">
			<div class="orb-section">
				<PurpleOrb size="lg" isThinking={isAgentTyping} />
			</div>

			<div class="text-section">
				{#if phase === 'intro'}
					<SequentialTypewriter
						lines={introLines}
						speed={30}
						lineDelay={600}
						onComplete={handleIntroComplete}
					/>
				{:else}
					<!-- Show welcome back message if resuming -->
					{#if isResuming}
						<p class="agent-text resume-message">{resumeMessage}</p>
						<button class="continue-resume-btn" onclick={() => { isResuming = false; }}>
							Continue
						</button>
					{:else if isAgentTyping}
						<!-- Current question text -->
						<div class="agent-text typing">
							<span class="dot"></span>
							<span class="dot"></span>
							<span class="dot"></span>
						</div>
					{:else}
						<p class="agent-text">
							{currentAgentMessage}
						</p>
					{/if}
				{/if}
			</div>

			{#if phase === 'conversation' && !isAgentTyping && !isResuming}
				<div class="input-section">
					<!-- Chips for businessType and teamSize -->
					{#if currentQuestionConfig.inputType === 'chips' && currentQuestionConfig.chips}
						<div class="chips-container">
							{#each currentQuestionConfig.chips as chip (chip.id)}
								<button 
									class="chip"
									onclick={() => handleChipSelect(chip.id)}
								>
									{chip.label}
								</button>
							{/each}
						</div>
					{:else}
						<!-- Chat input for other questions -->
						<form class="minimal-input" onsubmit={handleFormSubmit}>
							<button 
								type="button" 
								class="voice-btn" 
								class:recording={isRecording}
								onclick={toggleVoiceInput}
								disabled={isAgentTyping}
								aria-label={isRecording ? 'Stop recording' : 'Start voice input'}
							>
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<path d="M12 2a3 3 0 0 0-3 3v7a3 3 0 0 0 6 0V5a3 3 0 0 0-3-3Z"/>
									<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
									<line x1="12" x2="12" y1="19" y2="22"/>
								</svg>
							</button>
							<input
								type="text"
								bind:this={inputRef}
								bind:value={inputValue}
								placeholder={isRecording ? 'Listening...' : 'Type here...'}
								disabled={isAgentTyping || isRecording}
								autocomplete="off"
							/>
							<button type="submit" class="send-btn" disabled={isAgentTyping || !inputValue.trim()} aria-label="Send">
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
									<path d="m5 12 7-7 7 7"/>
									<path d="M12 19V5"/>
								</svg>
							</button>
						</form>
					{/if}

					<!-- Go Back button -->
					{#if canGoBack}
						<button class="go-back-btn" onclick={goBack}>
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
								<polyline points="15 18 9 12 15 6"/>
							</svg>
							Go back
						</button>
					{/if}
				</div>
			{/if}
		</div>
	{:else if phase === 'integrations'}
		<!-- Integrations phase -->
		<div class="integrations-layout">
			<div class="integrations-container">
				<h2 class="section-title">Connect your tools</h2>
				<p class="section-subtitle">
					Connect your favorite tools and we'll sync your data automatically.
				</p>

				<!-- Recommended section -->
				{#if recommendedIntegrations.length > 0}
					<div class="integrations-section">
						<div class="section-header">
							<h3 class="section-label">Recommended for you</h3>
							<button 
								class="connect-all-btn"
								onclick={connectAllRecommended}
								disabled={allRecommendedConnected}
							>
								{allRecommendedConnected ? 'All connected' : 'Connect all'}
							</button>
						</div>
						<div class="integrations-grid">
							{#each integrations.filter(i => recommendedIntegrations.includes(i.id)) as integration (integration.id)}
								<div class="integration-wrapper recommended">
									<IntegrationCard
										name={integration.name}
										icon={integration.icon}
										status={integrationStatuses[integration.id] || 'disconnected'}
										onConnect={() => handleIntegrationConnect(integration.id)}
										onDisconnect={() => handleIntegrationDisconnect(integration.id)}
									/>
								</div>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Other integrations section -->
				<div class="integrations-section">
					<h3 class="section-label">{recommendedIntegrations.length > 0 ? 'Other integrations' : 'Available integrations'}</h3>
					<div class="integrations-grid">
						{#each integrations.filter(i => !recommendedIntegrations.includes(i.id)) as integration (integration.id)}
							<div class="integration-wrapper">
								<IntegrationCard
									name={integration.name}
									icon={integration.icon}
									status={integrationStatuses[integration.id] || 'disconnected'}
									onConnect={() => handleIntegrationConnect(integration.id)}
									onDisconnect={() => handleIntegrationDisconnect(integration.id)}
								/>
							</div>
						{/each}
					</div>
				</div>
				<button class="continue-btn" onclick={completeIntegrations}>
					{selectedIntegrations.length > 0 ? 'Continue' : "I'll do this later"}
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.onboarding-screen {
		min-height: 100vh;
		background-color: var(--background, #ffffff);
		color: var(--foreground, #1f2937);
		position: relative;
	}

	/* Progress dots */
	.progress-dots {
		position: fixed;
		top: 24px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		gap: 8px;
		z-index: 10;
	}

	.progress-dots .dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background-color: var(--border, #e5e7eb);
		transition: all 0.3s ease;
	}

	.progress-dots .dot.active {
		background-color: var(--primary, #6366f1);
	}

	.progress-dots .dot.current {
		transform: scale(1.25);
	}

	/* Skip button */
	.skip-btn {
		position: fixed;
		top: 20px;
		right: 24px;
		padding: 8px 16px;
		font-size: 14px;
		font-weight: 500;
		color: var(--muted-foreground, #6b7280);
		background: transparent;
		border: none;
		cursor: pointer;
		transition: color 0.2s;
		z-index: 10;
	}

	.skip-btn:hover {
		color: var(--foreground, #1f2937);
	}

	/* Error Banner */
	.error-banner {
		position: fixed;
		top: 56px;
		left: 50%;
		transform: translateX(-50%);
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 16px;
		padding: 12px 16px;
		background: var(--destructive-bg, #fef2f2);
		border: 1px solid var(--destructive-border, #fecaca);
		border-radius: 10px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		max-width: 90%;
		min-width: 320px;
		z-index: 100;
	}

	.error-content {
		display: flex;
		align-items: center;
		gap: 10px;
		flex: 1;
	}

	.error-icon {
		width: 20px;
		height: 20px;
		color: var(--destructive, #ef4444);
		flex-shrink: 0;
	}

	.error-message {
		font-size: 14px;
		color: var(--destructive-text, #991b1b);
		line-height: 1.4;
	}

	.error-actions {
		display: flex;
		align-items: center;
		gap: 8px;
		flex-shrink: 0;
	}

	.error-btn {
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 6px 12px;
		font-size: 13px;
		font-weight: 500;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		transition: background 0.2s, transform 0.1s;
	}

	.error-btn svg {
		width: 14px;
		height: 14px;
	}

	.error-btn.retry {
		color: var(--destructive, #ef4444);
		background: var(--destructive-btn-bg, #fee2e2);
	}

	.error-btn.retry:hover {
		background: var(--destructive-btn-hover, #fecaca);
	}

	.error-btn.dismiss {
		padding: 6px;
		color: var(--muted-foreground, #6b7280);
		background: transparent;
	}

	.error-btn.dismiss:hover {
		background: var(--muted, #f3f4f6);
	}

	/* Resume/Welcome back styles */
	.resume-message {
		font-size: 16px;
		color: var(--muted-foreground, #6b7280);
		margin-bottom: 8px;
	}

	.continue-resume-btn {
		margin-top: 16px;
		padding: 12px 32px;
		font-size: 15px;
		font-weight: 500;
		color: white;
		background: var(--primary, #6366f1);
		border: none;
		border-radius: 8px;
		cursor: pointer;
		transition: background 0.2s, transform 0.1s;
	}

	.continue-resume-btn:hover {
		background: var(--primary-dark, #4f46e5);
		transform: translateY(-1px);
	}

	/* Centered layout for intro/conversation */
	.centered-layout {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
		gap: 32px;
	}

	.orb-section {
		display: flex;
		justify-content: center;
	}

	.text-section {
		text-align: center;
		max-width: 400px;
	}

	.text-section :global(.sequential-typewriter) {
		font-size: 18px;
		line-height: 1.6;
		color: var(--foreground, #1f2937);
	}

	.agent-text {
		font-size: 18px;
		line-height: 1.6;
		color: var(--foreground, #1f2937);
		margin: 0;
	}

	/* Typing dots */
	.agent-text.typing {
		display: flex;
		justify-content: center;
		gap: 6px;
	}

	.dot {
		width: 8px;
		height: 8px;
		background-color: var(--muted-foreground, #9ca3af);
		border-radius: 50%;
		animation: bounce 1.4s infinite ease-in-out both;
	}

	.dot:nth-child(1) { animation-delay: -0.32s; }
	.dot:nth-child(2) { animation-delay: -0.16s; }
	.dot:nth-child(3) { animation-delay: 0s; }

	@keyframes bounce {
		0%, 80%, 100% { transform: scale(0.8); opacity: 0.5; }
		40% { transform: scale(1); opacity: 1; }
	}

	/* Minimal input */
	.input-section {
		margin-top: 16px;
	}

	/* Go Back button */
	.go-back-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		margin-top: 16px;
		padding: 8px 16px;
		font-size: 13px;
		font-weight: 500;
		color: var(--muted-foreground, #6b7280);
		background: transparent;
		border: none;
		cursor: pointer;
		transition: color 0.2s, transform 0.1s;
	}

	.go-back-btn:hover {
		color: var(--foreground, #1f2937);
		transform: translateX(-2px);
	}

	.go-back-btn svg {
		transition: transform 0.2s;
	}

	.go-back-btn:hover svg {
		transform: translateX(-2px);
	}

	/* Chips for quick selection */
	.chips-container {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
		justify-content: center;
		max-width: 400px;
	}

	.chip {
		padding: 10px 20px;
		font-size: 14px;
		font-weight: 500;
		border: 1px solid var(--border, #e5e7eb);
		border-radius: 20px;
		background: var(--card, #ffffff);
		color: var(--foreground, #1f2937);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.chip:hover {
		border-color: var(--primary, #6366f1);
		background: var(--primary, #6366f1);
		color: white;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(99, 102, 241, 0.3);
	}

	.minimal-input {
		display: flex;
		align-items: center;
		gap: 4px;
		background: var(--card, #f9fafb);
		border: 1px solid var(--border, #e5e7eb);
		border-radius: 24px;
		padding: 4px;
		transition: border-color 0.2s, box-shadow 0.2s;
	}

	.minimal-input:focus-within {
		border-color: var(--primary, #6366f1);
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.minimal-input input {
		border: none;
		outline: none;
		background: transparent;
		font-size: 15px;
		color: var(--foreground, #1f2937);
		width: 160px;
		padding: 8px 12px;
	}

	.minimal-input input::placeholder {
		color: var(--muted-foreground, #9ca3af);
	}

	/* Voice button */
	.voice-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: none;
		background: transparent;
		color: var(--muted-foreground, #6b7280);
		cursor: pointer;
		transition: all 0.2s;
	}

	.voice-btn:hover:not(:disabled) {
		background: var(--accent, #f3f4f6);
		color: var(--foreground, #1f2937);
	}

	.voice-btn.recording {
		background: #ef4444;
		color: white;
		animation: pulse-recording 1.5s infinite;
	}

	@keyframes pulse-recording {
		0%, 100% { transform: scale(1); }
		50% { transform: scale(1.1); }
	}

	.voice-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	/* Send button */
	.send-btn {
		display: flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: none;
		background: var(--primary, #6366f1);
		color: white;
		cursor: pointer;
		transition: opacity 0.2s, transform 0.2s;
	}

	.send-btn:hover:not(:disabled) {
		opacity: 0.9;
		transform: scale(1.05);
	}

	.send-btn:disabled {
		opacity: 0.4;
		cursor: not-allowed;
	}

	/* Integrations layout */
	.integrations-layout {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: 48px 24px;
	}

	.integrations-container {
		display: flex;
		flex-direction: column;
		gap: 24px;
		max-width: 500px;
		width: 100%;
	}

	.section-title {
		font-size: 24px;
		font-weight: 600;
		color: var(--foreground, #1f2937);
		margin: 0;
		text-align: center;
	}

	.section-subtitle {
		font-size: 15px;
		color: var(--muted-foreground, #6b7280);
		margin: 0;
		text-align: center;
	}

	.integrations-section {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.section-label {
		font-size: 13px;
		font-weight: 600;
		color: var(--muted-foreground, #6b7280);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		margin: 0;
	}

	.connect-all-btn {
		padding: 6px 12px;
		font-size: 12px;
		font-weight: 500;
		border: 1px solid var(--primary, #6366f1);
		border-radius: 16px;
		background: transparent;
		color: var(--primary, #6366f1);
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.connect-all-btn:hover:not(:disabled) {
		background: var(--primary, #6366f1);
		color: white;
	}

	.connect-all-btn:disabled {
		opacity: 0.5;
		cursor: default;
		border-color: var(--success, #10b981);
		color: var(--success, #10b981);
	}

	.integrations-grid {
		display: flex;
		flex-direction: column;
		gap: 12px;
	}

	.integration-wrapper {
		position: relative;
		transition: transform 0.2s ease, box-shadow 0.2s ease;
	}

	.integration-wrapper:hover {
		transform: translateY(-2px);
	}

	.integration-wrapper.recommended {
		order: -1;
	}

	.recommended-badge {
		position: absolute;
		top: -6px;
		right: 12px;
		background: var(--primary, #6366f1);
		color: white;
		font-size: 10px;
		font-weight: 600;
		padding: 2px 8px;
		border-radius: 10px;
		z-index: 1;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.continue-btn {
		margin-top: 8px;
		padding: 14px 28px;
		font-size: 15px;
		font-weight: 500;
		border: none;
		border-radius: 24px;
		background-color: var(--primary, #6366f1);
		color: white;
		cursor: pointer;
		transition: opacity 0.2s, transform 0.2s;
		align-self: center;
	}

	.continue-btn:hover {
		opacity: 0.9;
		transform: translateY(-1px);
	}

	/* Dark mode */
	:global(.dark) .text-section :global(.sequential-typewriter) {
		color: var(--foreground, #f9fafb);
	}

	:global(.dark) .agent-text {
		color: var(--foreground, #f9fafb);
	}

	:global(.dark) .minimal-input {
		background: var(--card, #1f2937);
		border-color: var(--border, #374151);
	}

	:global(.dark) .minimal-input input {
		color: var(--foreground, #f9fafb);
	}

	:global(.dark) .section-title {
		color: var(--foreground, #f9fafb);
	}
</style>
