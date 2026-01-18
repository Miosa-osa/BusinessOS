/**
 * onboardingStore.ts
 * Manages the OSA Build onboarding flow state
 */

import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';

export interface OnboardingState {
	currentStep: number;
	totalSteps: number;
	completed: boolean;
	userData: {
		email?: string;
		username?: string;
		gmailConnected: boolean;
		interests?: string[];
		starterApps?: StarterApp[];
	};
	analysis: {
		message1?: string; // "No-code builder energy, big time"
		message2?: string; // "Design tools are your playground"
		message3?: string; // "AI-curious, testing new platforms"
	};
}

export interface StarterApp {
	id: string;
	title: string;
	description: string;
	iconUrl?: string;
	reason: string; // Why this app was generated for the user
}

// Default state
const defaultState: OnboardingState = {
	currentStep: 0,
	totalSteps: 13, // Total onboarding screens
	completed: false,
	userData: {
		gmailConnected: false
	},
	analysis: {}
};

// Load initial state from localStorage if available
function loadState(): OnboardingState {
	if (browser) {
		try {
			const stored = localStorage.getItem('osa_onboarding_state');
			if (stored) {
				return JSON.parse(stored);
			}
		} catch (e) {
			console.error('Error loading onboarding state:', e);
		}
	}
	return defaultState;
}

// Create the store
function createOnboardingStore() {
	const { subscribe, set, update } = writable<OnboardingState>(loadState());

	return {
		subscribe,

		// Navigate to next step
		nextStep: () => update(state => {
			const newState = { ...state, currentStep: Math.min(state.currentStep + 1, state.totalSteps - 1) };
			saveState(newState);
			return newState;
		}),

		// Navigate to previous step
		prevStep: () => update(state => {
			const newState = { ...state, currentStep: Math.max(state.currentStep - 1, 0) };
			saveState(newState);
			return newState;
		}),

		// Go to specific step
		goToStep: (step: number) => update(state => {
			const newState = { ...state, currentStep: Math.max(0, Math.min(step, state.totalSteps - 1)) };
			saveState(newState);
			return newState;
		}),

		// Update user data
		setUserData: (data: Partial<OnboardingState['userData']>) => update(state => {
			const newState = {
				...state,
				userData: { ...state.userData, ...data }
			};
			saveState(newState);
			return newState;
		}),

		// Set analysis messages
		setAnalysis: (analysis: Partial<OnboardingState['analysis']>) => update(state => {
			const newState = {
				...state,
				analysis: { ...state.analysis, ...analysis }
			};
			saveState(newState);
			return newState;
		}),

		// Set starter apps
		setStarterApps: (apps: StarterApp[]) => update(state => {
			const newState = {
				...state,
				userData: { ...state.userData, starterApps: apps }
			};
			saveState(newState);
			return newState;
		}),

		// Mark onboarding as completed
		complete: () => update(state => {
			const newState = { ...state, completed: true };
			saveState(newState);
			return newState;
		}),

		// Reset onboarding
		reset: () => {
			set(defaultState);
			if (browser) {
				localStorage.removeItem('osa_onboarding_state');
			}
		}
	};
}

// Save state to localStorage
function saveState(state: OnboardingState) {
	if (browser) {
		try {
			localStorage.setItem('osa_onboarding_state', JSON.stringify(state));
		} catch (e) {
			console.error('Error saving onboarding state:', e);
		}
	}
}

export const onboardingStore = createOnboardingStore();

// Derived store for progress percentage
export const onboardingProgress = derived(
	onboardingStore,
	$onboarding => {
		return Math.round(($onboarding.currentStep / ($onboarding.totalSteps - 1)) * 100);
	}
);

// Derived store to check if onboarding is complete
export const isOnboardingComplete = derived(
	onboardingStore,
	$onboarding => $onboarding.completed
);
