<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { contexts } from '$lib/stores/contexts';
	import { editor, wordCount, type EditorBlock, type BlockType, blockTypes, createEmptyBlock } from '$lib/stores/editor';
	import type { Context, Block, VoiceNote } from '$lib/api';
	import { api } from '$lib/api';
	import BlockComponent from '$lib/components/editor/Block.svelte';
	import BlockMenu from '$lib/components/editor/BlockMenu.svelte';
	import ChatInput from '$lib/components/chat/ChatInput.svelte';
	import AssistantMessage from '$lib/components/chat/AssistantMessage.svelte';
	import UserMessage from '$lib/components/chat/UserMessage.svelte';
	import TypingIndicator from '$lib/components/chat/TypingIndicator.svelte';
	import { desktopBackgrounds, getBackgroundCSS } from '$lib/stores/desktopStore';

	// Check if we're in embed mode to propagate to links
	const embedSuffix = $derived($page.url.searchParams.get('embed') === 'true' ? '?embed=true' : '');

	let context: Context | null = $state(null);
	let parentContext: Context | null = $state(null);
	let loading = $state(true);
	let error = $state<string | null>(null);

	// Profile selector state
	let availableProfiles: Context[] = $state([]);
	let showProfileSelector = $state(false);
	let loadingProfiles = $state(false);
	let titleInput: HTMLInputElement | null = $state(null);
	let title = $state('');
	let icon = $state<string | null>(null);
	let coverImage = $state<string | null>(null);
	let showCoverInput = $state(false);
	let showCoverPicker = $state(false);
	let coverInputValue = $state('');
	let coverTab = $state<'presets' | 'upload' | 'url'>('presets');
	let fileInput: HTMLInputElement | null = $state(null);
	let uploadingCover = $state(false);

	// Cover preset backgrounds (subset of desktop backgrounds suitable for covers)
	const coverPresets = desktopBackgrounds.filter(bg =>
		bg.type === 'gradient' || (bg.type === 'solid' && !bg.id.includes('dark'))
	).slice(0, 20);
	let autoSaveTimer: ReturnType<typeof setTimeout>;
	let showShareMenu = $state(false);
	let shareUrl = $state('');

	// AI Panel state
	interface AIMessage {
		id: string;
		role: 'user' | 'assistant';
		content: string;
		timestamp: string;
	}
	let aiMessages = $state<AIMessage[]>([]);
	let aiInput = $state('');
	let isAIStreaming = $state(false);
	let aiMessagesContainer: HTMLDivElement | undefined = $state(undefined);

	// Voice Notes state
	let showVoiceNotesPanel = $state(false);
	let voiceNotes = $state<VoiceNote[]>([]);
	let loadingVoiceNotes = $state(false);
	let isRecording = $state(false);
	let recordingTime = $state(0);
	let recordingTimer: ReturnType<typeof setInterval> | null = null;
	let mediaRecorder: MediaRecorder | null = null;
	let audioChunks: Blob[] = [];
	let isUploading = $state(false);
	let playingNoteId = $state<string | null>(null);
	let audioElement: HTMLAudioElement | null = null;

	const contextId = $derived($page.params.id);

	// Document icons for picker - SVG paths
	interface DocIcon {
		id: string;
		label: string;
		path: string;
	}

	const documentIcons: DocIcon[] = [
		// Documents
		{ id: 'document', label: 'Document', path: 'M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z' },
		{ id: 'document-text', label: 'Text Document', path: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
		{ id: 'clipboard', label: 'Clipboard', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2' },
		{ id: 'clipboard-list', label: 'Checklist', path: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
		{ id: 'document-duplicate', label: 'Documents', path: 'M8 7v8a2 2 0 002 2h6M8 7V5a2 2 0 012-2h4.586a1 1 0 01.707.293l4.414 4.414a1 1 0 01.293.707V15a2 2 0 01-2 2h-2M8 7H6a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2v-2' },
		// Folders
		{ id: 'folder', label: 'Folder', path: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' },
		{ id: 'folder-open', label: 'Open Folder', path: 'M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z' },
		{ id: 'archive', label: 'Archive', path: 'M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4' },
		// Writing & Notes
		{ id: 'pencil', label: 'Pencil', path: 'M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z' },
		{ id: 'pencil-alt', label: 'Edit', path: 'M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z' },
		{ id: 'annotation', label: 'Annotation', path: 'M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z' },
		{ id: 'book-open', label: 'Book', path: 'M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253' },
		// Data & Charts
		{ id: 'chart-bar', label: 'Bar Chart', path: 'M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z' },
		{ id: 'chart-pie', label: 'Pie Chart', path: 'M11 3.055A9.001 9.001 0 1020.945 13H11V3.055z M20.488 9H15V3.512A9.025 9.025 0 0120.488 9z' },
		{ id: 'trending-up', label: 'Trending Up', path: 'M13 7h8m0 0v8m0-8l-8 8-4-4-6 6' },
		{ id: 'table', label: 'Table', path: 'M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z' },
		// Goals & Stars
		{ id: 'flag', label: 'Flag', path: 'M3 21v-4m0 0V5a2 2 0 012-2h6.5l1 1H21l-3 6 3 6h-8.5l-1-1H5a2 2 0 00-2 2zm9-13.5V9' },
		{ id: 'star', label: 'Star', path: 'M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z' },
		{ id: 'sparkles', label: 'Sparkles', path: 'M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z' },
		{ id: 'fire', label: 'Fire', path: 'M17.657 18.657A8 8 0 016.343 7.343S7 9 9 10c0-2 .5-5 2.986-7C14 5 16.09 5.777 17.656 7.343A7.975 7.975 0 0120 13a7.975 7.975 0 01-2.343 5.657z M9.879 16.121A3 3 0 1012.015 11L11 14H9c0 .768.293 1.536.879 2.121z' },
		// Ideas & Creativity
		{ id: 'lightbulb', label: 'Idea', path: 'M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z' },
		{ id: 'puzzle', label: 'Puzzle', path: 'M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z' },
		{ id: 'color-swatch', label: 'Design', path: 'M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01' },
		// Technology
		{ id: 'code', label: 'Code', path: 'M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4' },
		{ id: 'terminal', label: 'Terminal', path: 'M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'database', label: 'Database', path: 'M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4' },
		{ id: 'server', label: 'Server', path: 'M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01' },
		{ id: 'chip', label: 'Chip', path: 'M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z' },
		// Media
		{ id: 'photograph', label: 'Image', path: 'M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'film', label: 'Video', path: 'M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z' },
		{ id: 'music-note', label: 'Music', path: 'M9 19V6l12-3v13M9 19c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zm12-3c0 1.105-1.343 2-3 2s-3-.895-3-2 1.343-2 3-2 3 .895 3 2zM9 10l12-3' },
		{ id: 'microphone', label: 'Audio', path: 'M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z' },
		// Business
		{ id: 'briefcase', label: 'Briefcase', path: 'M21 13.255A23.931 23.931 0 0112 15c-3.183 0-6.22-.62-9-1.745M16 6V4a2 2 0 00-2-2h-4a2 2 0 00-2 2v2m4 6h.01M5 20h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z' },
		{ id: 'currency-dollar', label: 'Finance', path: 'M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'receipt-tax', label: 'Receipt', path: 'M9 14l6-6m-5.5.5h.01m4.99 5h.01M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16l3.5-2 3.5 2 3.5-2 3.5 2zM10 8.5a.5.5 0 11-1 0 .5.5 0 011 0zm5 5a.5.5 0 11-1 0 .5.5 0 011 0z' },
		{ id: 'presentation-chart-line', label: 'Presentation', path: 'M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z' },
		// Security
		{ id: 'lock-closed', label: 'Lock', path: 'M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z' },
		{ id: 'key', label: 'Key', path: 'M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z' },
		{ id: 'shield-check', label: 'Shield', path: 'M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z' },
		// People
		{ id: 'user', label: 'User', path: 'M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z' },
		{ id: 'users', label: 'Team', path: 'M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z' },
		{ id: 'user-group', label: 'Group', path: 'M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z' },
		// Objects
		{ id: 'cube', label: 'Cube', path: 'M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4' },
		{ id: 'globe', label: 'Globe', path: 'M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9' },
		{ id: 'heart', label: 'Heart', path: 'M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z' },
		{ id: 'calendar', label: 'Calendar', path: 'M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z' },
		{ id: 'clock', label: 'Clock', path: 'M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'location-marker', label: 'Location', path: 'M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z M15 11a3 3 0 11-6 0 3 3 0 016 0z' },
		{ id: 'home', label: 'Home', path: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
		{ id: 'office-building', label: 'Office', path: 'M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4' },
		// Actions
		{ id: 'check-circle', label: 'Complete', path: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'exclamation-circle', label: 'Important', path: 'M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'question-mark-circle', label: 'Question', path: 'M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z' },
		{ id: 'lightning-bolt', label: 'Quick', path: 'M13 10V3L4 14h7v7l9-11h-7z' },
		{ id: 'rocket', label: 'Launch', path: 'M15.59 14.37a6 6 0 01-5.84 7.38v-4.8m5.84-2.58a14.98 14.98 0 006.16-12.12A14.98 14.98 0 009.631 8.41m5.96 5.96a14.926 14.926 0 01-5.841 2.58m-.119-8.54a6 6 0 00-7.381 5.84h4.8m2.581-5.84a14.927 14.927 0 00-2.58 5.84m2.699 2.7c-.103.021-.207.041-.311.06a15.09 15.09 0 01-2.448-2.448 14.9 14.9 0 01.06-.312m-2.24 2.39a4.493 4.493 0 00-1.757 4.306 4.493 4.493 0 004.306-1.758M16.5 9a1.5 1.5 0 11-3 0 1.5 1.5 0 013 0z' },
	];

	// Helper to get icon path by ID
	function getDocIconPath(iconId: string | null): string {
		if (!iconId) return documentIcons[0].path;
		const docIcon = documentIcons.find(i => i.id === iconId);
		return docIcon?.path || documentIcons[0].path;
	}

	let showIconPicker = $state(false);

	onMount(async () => {
		if (!contextId) {
			error = 'No document ID provided';
			loading = false;
			return;
		}
		try {
			const ctx = await contexts.loadContext(contextId);
			context = ctx;
			title = ctx.name;
			icon = ctx.icon;
			coverImage = ctx.cover_image;
			editor.initialize(ctx.blocks);

			// Load parent context if exists
			if (ctx.parent_id) {
				try {
					parentContext = await contexts.loadContext(ctx.parent_id);
				} catch (e) {
					console.error('Failed to load parent context:', e);
				}
			}

			loading = false;
		} catch (e) {
			error = 'Failed to load document';
			loading = false;
		}
	});

	onDestroy(() => {
		if (autoSaveTimer) clearTimeout(autoSaveTimer);
		if (recordingTimer) clearInterval(recordingTimer);
		if (audioElement) {
			audioElement.pause();
			audioElement = null;
		}
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
		}
		editor.reset();
	});

	// Auto-save with debounce
	$effect(() => {
		if ($editor.isDirty && context) {
			if (autoSaveTimer) clearTimeout(autoSaveTimer);
			autoSaveTimer = setTimeout(async () => {
				await saveDocument();
			}, 1500);
		}
	});

	async function saveDocument() {
		if (!context || $editor.isSaving) return;
		editor.setSaving(true);
		try {
			await contexts.updateBlocks(context.id, $editor.blocks as Block[], $wordCount);
			editor.markSaved();
		} catch (e) {
			console.error('Failed to save:', e);
			editor.setSaving(false);
		}
	}

	async function updateTitle() {
		if (!context || title === context.name) return;
		try {
			await contexts.updateContext(context.id, { name: title });
		} catch (e) {
			console.error('Failed to update title:', e);
		}
	}

	async function updateIcon(newIcon: string) {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { icon: newIcon || undefined });
			icon = newIcon || null;
		} catch (e) {
			console.error('Failed to update icon:', e);
		}
	}

	async function updateCoverImage() {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { cover_image: coverInputValue || undefined });
			coverImage = coverInputValue || null;
			showCoverInput = false;
			coverInputValue = '';
		} catch (e) {
			console.error('Failed to update cover:', e);
		}
	}

	async function removeCoverImage() {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { cover_image: undefined });
			coverImage = null;
			showCoverPicker = false;
		} catch (e) {
			console.error('Failed to remove cover:', e);
		}
	}

	async function selectCoverPreset(bgId: string) {
		if (!context) return;
		const bg = desktopBackgrounds.find(b => b.id === bgId);
		if (!bg) return;

		// For presets, we store the background CSS directly
		const coverValue = `preset:${bgId}`;
		try {
			await contexts.updateContext(context.id, { cover_image: coverValue });
			coverImage = coverValue;
			showCoverPicker = false;
		} catch (e) {
			console.error('Failed to set cover preset:', e);
		}
	}

	function getCoverStyle(cover: string | null): string {
		if (!cover) return '';
		if (cover.startsWith('preset:')) {
			const bgId = cover.replace('preset:', '');
			const cssObj = getBackgroundCSS(bgId);
			let style = `background: ${cssObj.background};`;
			if (cssObj.backgroundSize) {
				style += ` background-size: ${cssObj.backgroundSize};`;
			}
			return style;
		}
		return `background-image: url(${cover}); background-size: cover; background-position: center;`;
	}

	// Helper to convert background object to style string for presets
	function getPresetStyle(bgId: string): string {
		const cssObj = getBackgroundCSS(bgId);
		let style = `background: ${cssObj.background};`;
		if (cssObj.backgroundSize) {
			style += ` background-size: ${cssObj.backgroundSize};`;
		}
		return style;
	}

	// Handle file upload for cover image
	async function handleCoverUpload(event: Event) {
		const input = event.target as HTMLInputElement;
		const file = input.files?.[0];
		if (!file || !context) return;

		// Validate file type
		if (!file.type.startsWith('image/')) {
			alert('Please select an image file');
			return;
		}

		// Validate file size (max 5MB)
		if (file.size > 5 * 1024 * 1024) {
			alert('Image must be less than 5MB');
			return;
		}

		uploadingCover = true;
		try {
			// Convert to base64 for now (in production, upload to storage)
			const reader = new FileReader();
			reader.onload = async (e) => {
				const dataUrl = e.target?.result as string;
				await contexts.updateContext(context!.id, { cover_image: dataUrl });
				coverImage = dataUrl;
				showCoverPicker = false;
				uploadingCover = false;
			};
			reader.onerror = () => {
				alert('Failed to read image file');
				uploadingCover = false;
			};
			reader.readAsDataURL(file);
		} catch (error) {
			console.error('Failed to upload cover:', error);
			alert('Failed to upload cover image');
			uploadingCover = false;
		}
	}

	async function toggleShare() {
		if (!context) return;
		try {
			if (context.is_public) {
				await contexts.disableSharing(context.id);
				showShareMenu = false;
			} else {
				const response = await contexts.enableSharing(context.id);
				shareUrl = response.share_url;
			}
		} catch (e) {
			console.error('Failed to toggle sharing:', e);
		}
	}

	async function copyShareLink() {
		if (shareUrl) {
			await navigator.clipboard.writeText(shareUrl);
		}
	}

	async function duplicateDoc() {
		if (!context) return;
		try {
			const newContext = await contexts.duplicateContext(context.id);
			goto(`/knowledge/${newContext.id}${embedSuffix}`);
		} catch (e) {
			console.error('Failed to duplicate:', e);
		}
	}

	async function archiveDoc() {
		if (!context) return;
		try {
			await contexts.archiveContext(context.id);
			goto('/knowledge' + embedSuffix);
		} catch (e) {
			console.error('Failed to archive:', e);
		}
	}

	async function deleteDoc() {
		if (!context) return;
		if (!confirm('Are you sure you want to delete this document? This cannot be undone.')) return;
		try {
			await contexts.deleteContext(context.id);
			goto('/knowledge' + embedSuffix);
		} catch (e) {
			console.error('Failed to delete:', e);
		}
	}

	async function loadAvailableProfiles() {
		if (loadingProfiles) return;
		loadingProfiles = true;
		try {
			// Load all contexts and wait for them
			await contexts.loadContexts();
			// Get profiles from the store (non-document contexts) excluding current
			availableProfiles = $contexts.contexts.filter(
				(c) => c.type !== 'document' && c.id !== contextId
			) as Context[];
		} catch (e) {
			console.error('Failed to load profiles:', e);
		} finally {
			loadingProfiles = false;
		}
	}

	async function updateParentProfile(profileId: string | null) {
		if (!context) return;
		try {
			await contexts.updateContext(context.id, { parent_id: profileId });
			if (profileId) {
				parentContext = availableProfiles.find(p => p.id === profileId) || null;
			} else {
				parentContext = null;
			}
			showProfileSelector = false;
		} catch (e) {
			console.error('Failed to update parent profile:', e);
		}
	}

	function handleTitleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			updateTitle();
			// Focus first block
			const firstBlock = document.querySelector('[data-block-id]') as HTMLElement;
			firstBlock?.focus();
		}
	}

	function handleBlockKeydown(e: KeyboardEvent) {
		// Space on empty block opens AI panel
		if (e.key === ' ' && !e.ctrlKey && !e.metaKey) {
			const target = e.target as HTMLElement;
			const blockId = target.getAttribute('data-block-id');
			if (blockId) {
				const block = $editor.blocks.find((b) => b.id === blockId);
				if (block && block.content === '') {
					e.preventDefault();
					editor.showAIPanel();
				}
			}
		}
	}

	function getTypeIcon(type: BlockType) {
		return blockTypes.find(bt => bt.type === type)?.icon || 'T';
	}

	function addNewBlockAtEnd() {
		// Add a new empty block at the end
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];
		if (lastBlock) {
			const newBlockId = editor.addBlockAfter(lastBlock.id);
			// Focus the new block after render
			setTimeout(() => {
				const blockEl = document.querySelector(`[data-block-id="${newBlockId}"]`) as HTMLElement;
				blockEl?.focus();
			}, 10);
		}
	}

	// Close icon picker and profile selector when clicking outside
	function handleClickOutside(e: MouseEvent) {
		const target = e.target as HTMLElement;
		if (showIconPicker && !target.closest('.relative.inline-block')) {
			showIconPicker = false;
		}
		if (showProfileSelector && !target.closest('.profile-selector-container')) {
			showProfileSelector = false;
		}
	}

	// AI Panel functions
	async function handleAISend(message: string) {
		if (!message.trim() || isAIStreaming) return;

		// Add user message
		const userMessage: AIMessage = {
			id: crypto.randomUUID(),
			role: 'user',
			content: message,
			timestamp: new Date().toISOString()
		};
		aiMessages = [...aiMessages, userMessage];

		// Scroll to bottom
		setTimeout(() => {
			if (aiMessagesContainer) {
				aiMessagesContainer.scrollTop = aiMessagesContainer.scrollHeight;
			}
		}, 10);

		isAIStreaming = true;

		try {
			// Get current document content for context
			const documentContent = $editor.blocks.map(b => b.content).join('\n');

			// Call AI endpoint
			const response = await fetch('/api/chat/ai/document', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({
					message,
					context: {
						documentTitle: title,
						documentContent,
						contextType: context?.type
					}
				})
			});

			if (!response.ok) {
				throw new Error('AI request failed');
			}

			const data = await response.json();

			// Add assistant message
			const assistantMessage: AIMessage = {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: data.response || 'I apologize, but I was unable to generate a response.',
				timestamp: new Date().toISOString()
			};
			aiMessages = [...aiMessages, assistantMessage];

		} catch (error) {
			console.error('AI error:', error);
			// Add error message
			const errorMessage: AIMessage = {
				id: crypto.randomUUID(),
				role: 'assistant',
				content: 'Sorry, I encountered an error. Please try again.',
				timestamp: new Date().toISOString()
			};
			aiMessages = [...aiMessages, errorMessage];
		} finally {
			isAIStreaming = false;
			// Scroll to bottom
			setTimeout(() => {
				if (aiMessagesContainer) {
					aiMessagesContainer.scrollTop = aiMessagesContainer.scrollHeight;
				}
			}, 10);
		}
	}

	function handleAIStop() {
		isAIStreaming = false;
	}

	function insertAIContent(content: string) {
		// Insert AI-generated content as new blocks
		const lines = content.split('\n').filter(line => line.trim());
		const lastBlock = $editor.blocks[$editor.blocks.length - 1];

		if (lastBlock) {
			let currentBlockId = lastBlock.id;
			for (const line of lines) {
				currentBlockId = editor.addBlockAfter(currentBlockId, 'paragraph');
				editor.updateBlock(currentBlockId, line);
			}
		}

		// Close AI panel after inserting
		editor.hideAIPanel();
	}

	function clearAIChat() {
		aiMessages = [];
	}

	// Voice Notes functions
	async function loadVoiceNotes() {
		loadingVoiceNotes = true;
		try {
			voiceNotes = await api.getVoiceNotes(contextId);
		} catch (e) {
			console.error('Failed to load voice notes:', e);
		} finally {
			loadingVoiceNotes = false;
		}
	}

	async function startRecording() {
		try {
			const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
			mediaRecorder = new MediaRecorder(stream, { mimeType: 'audio/webm' });
			audioChunks = [];

			mediaRecorder.ondataavailable = (e) => {
				if (e.data.size > 0) {
					audioChunks.push(e.data);
				}
			};

			mediaRecorder.onstop = async () => {
				stream.getTracks().forEach(track => track.stop());
				const audioBlob = new Blob(audioChunks, { type: 'audio/webm' });
				await uploadVoiceNote(audioBlob);
			};

			mediaRecorder.start(1000);
			isRecording = true;
			recordingTime = 0;
			recordingTimer = setInterval(() => {
				recordingTime++;
			}, 1000);
		} catch (e) {
			console.error('Failed to start recording:', e);
			alert('Could not access microphone. Please check permissions.');
		}
	}

	function stopRecording() {
		if (mediaRecorder && mediaRecorder.state !== 'inactive') {
			mediaRecorder.stop();
		}
		isRecording = false;
		if (recordingTimer) {
			clearInterval(recordingTimer);
			recordingTimer = null;
		}
	}

	async function uploadVoiceNote(audioBlob: Blob) {
		isUploading = true;
		try {
			const note = await api.uploadVoiceNote(audioBlob, contextId);
			voiceNotes = [note, ...voiceNotes];
		} catch (e) {
			console.error('Failed to upload voice note:', e);
			alert('Failed to save voice note');
		} finally {
			isUploading = false;
		}
	}

	async function playVoiceNote(noteId: string) {
		if (playingNoteId === noteId) {
			// Stop playing
			if (audioElement) {
				audioElement.pause();
				audioElement = null;
			}
			playingNoteId = null;
			return;
		}

		try {
			const blob = await api.getVoiceNoteAudio(noteId);
			const url = URL.createObjectURL(blob);

			if (audioElement) {
				audioElement.pause();
			}

			audioElement = new Audio(url);
			audioElement.onended = () => {
				playingNoteId = null;
				URL.revokeObjectURL(url);
			};
			audioElement.play();
			playingNoteId = noteId;
		} catch (e) {
			console.error('Failed to play voice note:', e);
		}
	}

	async function deleteVoiceNote(noteId: string) {
		if (!confirm('Delete this voice note?')) return;
		try {
			await api.deleteVoiceNote(noteId);
			voiceNotes = voiceNotes.filter(n => n.id !== noteId);
		} catch (e) {
			console.error('Failed to delete voice note:', e);
		}
	}

	function formatDuration(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = Math.floor(seconds % 60);
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function formatTimeAgo(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 7) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	}

	function openVoiceNotesPanel() {
		showVoiceNotesPanel = true;
		loadVoiceNotes();
	}
</script>

<svelte:head>
	<title>{title || 'Untitled'} - BusinessOS</title>
</svelte:head>

<svelte:window onclick={handleClickOutside} />

{#if loading}
	<div class="h-full flex items-center justify-center bg-white dark:bg-[#1c1c1e]">
		<div class="animate-spin h-8 w-8 border-2 border-gray-400 dark:border-gray-600 border-t-gray-700 dark:border-t-gray-300 rounded-full"></div>
	</div>
{:else if error}
	<div class="h-full flex items-center justify-center bg-white dark:bg-[#1c1c1e]">
		<div class="text-center">
			<p class="text-red-500 mb-4">{error}</p>
			<a href="/knowledge{embedSuffix}" class="px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors">Back to Knowledge</a>
		</div>
	</div>
{:else if context}
	<div class="h-full flex flex-col bg-white dark:bg-[#1c1c1e]">
		<!-- Top toolbar -->
		<div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700/50 flex items-center justify-between">
			<div class="flex items-center gap-2">
				<a href="/knowledge{embedSuffix}" class="p-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400" title="Back to Knowledge">
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
					</svg>
				</a>

				<!-- Profile Selector -->
				<div class="relative profile-selector-container">
					<button
						onclick={() => {
							loadAvailableProfiles();
							showProfileSelector = !showProfileSelector;
						}}
						class="flex items-center gap-1.5 px-2 py-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-sm transition-colors {parentContext ? 'text-gray-700 dark:text-gray-200 bg-gray-100 dark:bg-gray-700/50' : 'text-gray-500 dark:text-gray-400'}"
						title={parentContext ? `Linked to ${parentContext.name}` : 'Add to a profile'}
					>
						{#if parentContext}
							<span class="text-base">{parentContext.icon || '📁'}</span>
							<span class="font-medium">{parentContext.name}</span>
							<span class="text-xs text-gray-500 ml-0.5">(linked)</span>
						{:else}
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
							</svg>
							<span>Add to profile</span>
						{/if}
						<svg class="w-3 h-3 text-gray-500 ml-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
						</svg>
					</button>

					{#if showProfileSelector}
						<div class="absolute left-0 top-full mt-1 w-64 bg-white dark:bg-[#2c2c2e] rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 py-1 z-50 max-h-80 overflow-y-auto">
							<div class="px-3 py-2 border-b border-gray-100 dark:border-gray-700/50">
								<span class="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Link to Profile</span>
							</div>

							{#if loadingProfiles}
								<div class="px-3 py-4 text-center">
									<div class="animate-spin h-5 w-5 border-2 border-gray-300 dark:border-gray-600 border-t-gray-500 dark:border-t-gray-300 rounded-full mx-auto"></div>
								</div>
							{:else if availableProfiles.length === 0}
								<div class="px-3 py-4 text-center text-sm text-gray-500">
									No profiles available
								</div>
							{:else}
								{#if parentContext}
									<button
										onclick={() => updateParentProfile(null)}
										class="w-full px-3 py-2 text-left text-sm text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 flex items-center gap-2"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
										</svg>
										Remove link
									</button>
									<hr class="my-1 border-gray-100 dark:border-gray-700/50" />
								{/if}

								{#each availableProfiles as profile}
									<button
										onclick={() => updateParentProfile(profile.id)}
										class="w-full px-3 py-2 text-left text-sm hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center gap-2 {parentContext?.id === profile.id ? 'bg-blue-50 dark:bg-blue-900/40 text-blue-600 dark:text-blue-300' : 'text-gray-700 dark:text-gray-200'}"
									>
										<span class="text-base">{profile.icon || '📁'}</span>
										<span class="truncate flex-1">{profile.name}</span>
										{#if parentContext?.id === profile.id}
											<svg class="w-4 h-4 text-blue-500 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
											</svg>
										{/if}
									</button>
								{/each}
							{/if}
						</div>
					{/if}
				</div>

				{#if parentContext}
					<span class="text-sm text-gray-400 dark:text-gray-500">/</span>
				{/if}
				<span class="text-sm text-gray-700 dark:text-gray-300">{title || 'Untitled'}</span>
			</div>

			<div class="flex items-center gap-2">
				<!-- Save status -->
				<div class="text-xs text-gray-500 mr-2">
					{#if $editor.isDirty}
						<span class="text-amber-500 dark:text-amber-400">Unsaved</span>
					{:else if $editor.isSaving}
						<span>Saving...</span>
					{:else if $editor.lastSavedAt}
						<span class="text-gray-500 dark:text-gray-400">Saved</span>
					{/if}
				</div>

				<!-- Voice Notes button -->
				<button
					onclick={openVoiceNotesPanel}
					class="p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors relative"
					title="Voice notes"
				>
					<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
					</svg>
					{#if voiceNotes.length > 0}
						<span class="absolute -top-0.5 -right-0.5 w-4 h-4 bg-blue-500 text-white text-[10px] rounded-full flex items-center justify-center font-medium">
							{voiceNotes.length > 9 ? '9+' : voiceNotes.length}
						</span>
					{/if}
				</button>

				<!-- Share button -->
				<div class="relative">
					<button
						onclick={() => showShareMenu = !showShareMenu}
						class="px-3 py-1.5 text-sm flex items-center gap-1.5 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 transition-colors"
					>
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
						</svg>
						Share
					</button>

					{#if showShareMenu}
						<div class="absolute right-0 top-full mt-2 w-72 bg-white dark:bg-[#2c2c2e] rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 p-4 z-50">
							<div class="flex items-center justify-between mb-3">
								<span class="text-sm font-medium text-gray-900 dark:text-gray-100">Share to web</span>
								<button
									onclick={toggleShare}
									class="relative w-10 h-6 rounded-full transition-colors {context.is_public ? 'bg-blue-500' : 'bg-gray-300 dark:bg-gray-600'}"
								>
									<span
										class="absolute top-1 w-4 h-4 rounded-full bg-white shadow transition-transform {context.is_public ? 'left-5' : 'left-1'}"
									></span>
								</button>
							</div>
							{#if context.is_public}
								<div class="space-y-2">
									<p class="text-xs text-gray-500 dark:text-gray-400">Anyone with the link can view</p>
									<div class="flex gap-2">
										<input
											type="text"
											value={shareUrl}
											readonly
											class="flex-1 px-3 py-1.5 text-xs bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-gray-700 dark:text-gray-200"
										/>
										<button onclick={copyShareLink} class="btn-pill btn-pill-primary btn-pill-xs">Copy</button>
									</div>
								</div>
							{:else}
								<p class="text-xs text-gray-500 dark:text-gray-400">Enable sharing to get a public link</p>
							{/if}
						</div>
					{/if}
				</div>

				<!-- More options -->
				<div class="relative group">
					<button class="p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
						</svg>
					</button>
					<div class="absolute right-0 top-full mt-1 w-48 bg-white dark:bg-[#2c2c2e] rounded-lg shadow-xl border border-gray-200 dark:border-gray-700 py-1 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all z-50">
						<button onclick={duplicateDoc} class="btn-pill btn-pill-ghost w-full justify-start">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							Duplicate
						</button>
						<button onclick={archiveDoc} class="btn-pill btn-pill-ghost w-full justify-start">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
							</svg>
							Archive
						</button>
						<hr class="my-1 border-gray-100 dark:border-gray-700/50" />
						<button onclick={deleteDoc} class="btn-pill btn-pill-danger btn-pill-ghost w-full justify-start">
							<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
							</svg>
							Delete
						</button>
					</div>
				</div>
			</div>
		</div>

		<!-- Document content -->
		<div class="flex-1 overflow-y-auto">
			<!-- Cover Image - Full width outside max-w container -->
			{#if coverImage}
				<div class="relative w-full h-52 group">
					{#if coverImage.startsWith('preset:')}
						<div class="w-full h-full" style={getCoverStyle(coverImage)}></div>
					{:else}
						<img src={coverImage} alt="Cover" class="w-full h-full object-cover" />
					{/if}
					<!-- Hover buttons - top right corner only -->
					<div class="absolute top-3 right-3 flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
						<button
							onclick={() => showCoverPicker = true}
							class="px-2.5 py-1 text-xs font-medium bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm text-gray-700 dark:text-gray-200 rounded-md hover:bg-white dark:hover:bg-gray-700 shadow-sm transition-colors"
						>
							Change cover
						</button>
						<button
							onclick={removeCoverImage}
							class="p-1.5 bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm text-gray-500 dark:text-gray-400 rounded-md hover:bg-white dark:hover:bg-gray-700 hover:text-red-500 dark:hover:text-red-400 shadow-sm transition-colors"
							title="Remove cover"
						>
							<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				</div>
			{/if}

			<div class="max-w-3xl mx-auto px-8 {coverImage ? 'pt-8' : 'pt-12'} pb-12">
				<!-- Icon and Title - with hover actions -->
				<div class="mb-6 group/title">
					<!-- Icon - Only show if icon is set -->
					{#if icon}
						<div class="relative inline-block mb-2">
							<button
								onclick={() => showIconPicker = !showIconPicker}
								class="w-16 h-16 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded-xl transition-colors"
							>
								<svg class="w-10 h-10 text-gray-700 dark:text-gray-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={getDocIconPath(icon)} />
								</svg>
							</button>
							{#if showIconPicker}
								<div class="absolute top-full left-0 mt-1 bg-white dark:bg-[#2c2c2e] rounded-xl shadow-xl border border-gray-200 dark:border-gray-700 p-3 z-20 w-72">
									<div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">Choose icon</div>
									<div class="grid grid-cols-6 gap-1.5 max-h-64 overflow-y-auto">
										{#each documentIcons as docIcon}
											<button
												onclick={() => { updateIcon(docIcon.id); showIconPicker = false; }}
												class="w-9 h-9 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors {icon === docIcon.id ? 'bg-blue-100 dark:bg-blue-900/50 ring-2 ring-blue-500' : ''}"
												title={docIcon.label}
											>
												<svg class="w-5 h-5 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
													<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={docIcon.path} />
												</svg>
											</button>
										{/each}
									</div>
									<hr class="my-2 border-gray-200 dark:border-gray-700/50" />
									<button
										onclick={() => { updateIcon(''); showIconPicker = false; }}
										class="w-full px-2 py-1.5 text-left text-sm text-red-500 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/30 rounded flex items-center gap-2"
									>
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
										</svg>
										Remove icon
									</button>
								</div>
							{/if}
						</div>
					{/if}

					<!-- Icon Picker Modal (when no icon set) -->
					{#if showIconPicker && !icon}
						<div class="fixed inset-0 z-50 flex items-start justify-center pt-32">
							<button class="absolute inset-0 bg-black/20 dark:bg-black/40" onclick={() => showIconPicker = false}></button>
							<div class="relative bg-white dark:bg-[#2c2c2e] rounded-xl shadow-xl border border-gray-200 dark:border-gray-700 p-4 w-80">
								<div class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-3">Choose an icon</div>
								<div class="grid grid-cols-6 gap-1.5 max-h-64 overflow-y-auto">
									{#each documentIcons as docIcon}
										<button
											onclick={() => { updateIcon(docIcon.id); showIconPicker = false; }}
											class="w-10 h-10 flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
											title={docIcon.label}
										>
											<svg class="w-6 h-6 text-gray-600 dark:text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d={docIcon.path} />
											</svg>
										</button>
									{/each}
								</div>
							</div>
						</div>
					{/if}

					<!-- Action bar - appears on hover over title area -->
					{#if !icon || !coverImage}
						<div class="flex items-center gap-2 mb-2 opacity-0 group-hover/title:opacity-100 transition-opacity">
							{#if !icon}
								<button
									onclick={() => showIconPicker = true}
									class="text-sm text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md px-2 py-1 flex items-center gap-1.5 transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									Add icon
								</button>
							{/if}
							{#if !coverImage}
								<button
									onclick={() => showCoverPicker = true}
									class="text-sm text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-md px-2 py-1 flex items-center gap-1.5 transition-colors"
								>
									<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
									</svg>
									Add cover
								</button>
							{/if}
						</div>
					{/if}

					<!-- Title -->
					<input
						bind:this={titleInput}
						bind:value={title}
						onblur={updateTitle}
						onkeydown={handleTitleKeydown}
						onfocus={(e) => { if (title === 'Untitled') { title = ''; (e.target as HTMLInputElement).select(); } }}
						placeholder="Untitled"
						class="w-full text-4xl font-bold border-none outline-none bg-transparent {title === 'Untitled' || !title ? 'text-gray-400 dark:text-gray-500' : 'text-gray-900 dark:text-gray-100'} placeholder:text-gray-400 dark:placeholder:text-gray-500 caret-gray-900 dark:caret-gray-100"
					/>
				</div>


				<!-- Blocks -->
				<div
					class="blocks-container"
					onkeydown={handleBlockKeydown}
					role="textbox"
					tabindex="-1"
				>
					{#each $editor.blocks as block, index (block.id)}
						<BlockComponent {block} {index} readonly={false} parentContextId={contextId} />
					{/each}
				</div>

				<!-- Click area to add new blocks -->
				<button
					onclick={addNewBlockAtEnd}
					class="w-full min-h-32 mt-4 text-left cursor-text group"
				>
					<span class="text-gray-500 opacity-0 group-hover:opacity-100 transition-opacity text-sm">
						Click to add a block, or press / for commands
					</span>
				</button>
			</div>
		</div>

		<!-- Status Bar -->
		<div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700/50 flex items-center justify-between text-xs text-gray-500 dark:text-gray-500">
			<div class="flex items-center gap-4">
				<span>{$wordCount} words</span>
				<span>{$editor.blocks.length} blocks</span>
			</div>
			<div class="flex items-center gap-2">
				<button onclick={saveDocument} class="hover:text-gray-700 dark:hover:text-gray-300" disabled={!$editor.isDirty}>
					Save now
				</button>
			</div>
		</div>

		<!-- Slash Command Menu (global) -->
		{#if $editor.showSlashMenu && $editor.slashMenuPosition}
			<BlockMenu />
		{/if}

		<!-- AI Panel -->
		{#if $editor.showAIPanel}
			<div class="fixed inset-y-0 right-0 w-[420px] bg-white dark:bg-[#1c1c1e] border-l border-gray-200 dark:border-gray-700 shadow-xl z-50 flex flex-col">
				<!-- Header -->
				<div class="p-4 border-b border-gray-200 dark:border-gray-700/50 flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center">
							<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
							</svg>
						</div>
						<div>
							<h3 class="font-medium text-gray-900 dark:text-gray-100">AI Assistant</h3>
							<p class="text-xs text-gray-500">Help with writing & editing</p>
						</div>
					</div>
					<div class="flex items-center gap-1">
						{#if aiMessages.length > 0}
							<button
								onclick={clearAIChat}
								class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
								title="Clear chat"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
							</button>
						{/if}
						<button
							onclick={() => editor.hideAIPanel()}
							class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
							title="Close"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>
				</div>

				<!-- Quick Actions -->
				<div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700/50 flex flex-wrap gap-2">
					<button
						onclick={() => handleAISend('Help me write a summary of this document')}
						class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-full transition-colors"
					>
						Summarize
					</button>
					<button
						onclick={() => handleAISend('Help me improve the writing in this document')}
						class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-full transition-colors"
					>
						Improve writing
					</button>
					<button
						onclick={() => handleAISend('Check this document for grammar and spelling errors')}
						class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-full transition-colors"
					>
						Check grammar
					</button>
					<button
						onclick={() => handleAISend('Make this document shorter and more concise')}
						class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-full transition-colors"
					>
						Make shorter
					</button>
					<button
						onclick={() => handleAISend('Expand on the ideas in this document with more detail')}
						class="px-3 py-1.5 text-xs bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-full transition-colors"
					>
						Expand
					</button>
				</div>

				<!-- Messages Area -->
				<div
					bind:this={aiMessagesContainer}
					class="flex-1 overflow-y-auto p-4 space-y-4"
				>
					{#if aiMessages.length === 0}
						<div class="text-center py-8">
							<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
								<svg class="w-8 h-8 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
								</svg>
							</div>
							<h4 class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">How can I help?</h4>
							<p class="text-xs text-gray-500 max-w-[240px] mx-auto">
								Ask me to write, edit, summarize, or improve your document content.
							</p>
						</div>
					{:else}
						{#each aiMessages as message (message.id)}
							{#if message.role === 'user'}
								<UserMessage
									content={message.content}
									timestamp={message.timestamp}
								/>
							{:else}
								<div>
									<AssistantMessage
										content={message.content}
										timestamp={message.timestamp}
										isStreaming={false}
										onCopy={() => navigator.clipboard.writeText(message.content)}
									/>
									<!-- Insert to document button -->
									<div class="ml-9 mt-1">
										<button
											onclick={() => insertAIContent(message.content)}
											class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-blue-300 hover:text-blue-200 bg-blue-900/40 hover:bg-blue-900/60 rounded-lg transition-colors"
										>
											<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
											</svg>
											Insert into document
										</button>
									</div>
								</div>
							{/if}
						{/each}
						{#if isAIStreaming}
							<TypingIndicator />
						{/if}
					{/if}
				</div>

				<!-- Input Area -->
				<ChatInput
					bind:value={aiInput}
					placeholder="Ask AI to help with your document..."
					streaming={isAIStreaming}
					contextName={title || 'Document'}
					modelName="AI Assistant"
					onSend={handleAISend}
					onStop={handleAIStop}
				/>
			</div>
		{/if}

		<!-- Voice Notes Panel -->
		{#if showVoiceNotesPanel}
			<div class="fixed inset-y-0 right-0 w-[380px] bg-white dark:bg-[#1c1c1e] border-l border-gray-200 dark:border-gray-700 shadow-xl z-50 flex flex-col">
				<!-- Header -->
				<div class="p-4 border-b border-gray-200 dark:border-gray-700/50 flex items-center justify-between">
					<div class="flex items-center gap-2">
						<div class="w-8 h-8 rounded-lg bg-gradient-to-br from-red-500 to-orange-500 flex items-center justify-center">
							<svg class="w-4 h-4 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
							</svg>
						</div>
						<div>
							<h3 class="font-medium text-gray-900 dark:text-gray-100">Voice Notes</h3>
							<p class="text-xs text-gray-500">{voiceNotes.length} recording{voiceNotes.length !== 1 ? 's' : ''}</p>
						</div>
					</div>
					<button
						onclick={() => showVoiceNotesPanel = false}
						class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 transition-colors"
						title="Close"
					>
						<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
						</svg>
					</button>
				</div>

				<!-- Recording Section -->
				<div class="p-4 border-b border-gray-200 dark:border-gray-700/50">
					{#if isRecording}
						<div class="flex items-center gap-4">
							<button
								onclick={stopRecording}
								class="w-14 h-14 rounded-full btn-pill btn-pill-danger flex items-center justify-center text-white shadow-lg transition-all animate-pulse"
							>
								<svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24">
									<rect x="6" y="6" width="12" height="12" rx="2" />
								</svg>
							</button>
							<div class="flex-1">
								<div class="text-sm font-medium text-gray-900 dark:text-gray-100">Recording...</div>
								<div class="text-2xl font-mono text-red-500 dark:text-red-400">{formatDuration(recordingTime)}</div>
							</div>
							<div class="flex gap-1">
								{#each Array(5) as _, i}
									<div
										class="w-1 bg-red-500 rounded-full animate-pulse"
										style="height: {8 + Math.random() * 24}px; animation-delay: {i * 0.1}s"
									></div>
								{/each}
							</div>
						</div>
					{:else if isUploading}
						<div class="flex items-center gap-4">
							<div class="w-14 h-14 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
								<div class="animate-spin h-6 w-6 border-2 border-gray-400 dark:border-gray-500 border-t-gray-600 dark:border-t-gray-300 rounded-full"></div>
							</div>
							<div>
								<div class="text-sm font-medium text-gray-900 dark:text-gray-100">Processing...</div>
								<div class="text-xs text-gray-500">Transcribing audio</div>
							</div>
						</div>
					{:else}
						<div class="flex items-center gap-4">
							<button
								onclick={startRecording}
								class="w-14 h-14 rounded-full btn-pill btn-pill-danger flex items-center justify-center text-white shadow-lg transition-all hover:scale-105"
							>
								<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
								</svg>
							</button>
							<div>
								<div class="text-sm font-medium text-gray-900 dark:text-gray-100">Record a note</div>
								<div class="text-xs text-gray-500">Click to start recording</div>
							</div>
						</div>
					{/if}
				</div>

				<!-- Voice Notes List -->
				<div class="flex-1 overflow-y-auto">
					{#if loadingVoiceNotes}
						<div class="p-8 text-center">
							<div class="animate-spin h-6 w-6 border-2 border-gray-400 dark:border-gray-600 border-t-gray-600 dark:border-t-gray-300 rounded-full mx-auto"></div>
						</div>
					{:else if voiceNotes.length === 0}
						<div class="p-8 text-center">
							<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
								<svg class="w-8 h-8 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
								</svg>
							</div>
							<h4 class="text-sm font-medium text-gray-700 dark:text-gray-200 mb-1">No voice notes yet</h4>
							<p class="text-xs text-gray-500">Record your first voice note above</p>
						</div>
					{:else}
						<div class="divide-y divide-gray-200 dark:divide-gray-700/50">
							{#each voiceNotes as note (note.id)}
								<div class="p-4 hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors group">
									<div class="flex items-start gap-3">
										<!-- Play button -->
										<button
											onclick={() => playVoiceNote(note.id)}
											class="w-10 h-10 rounded-full flex-shrink-0 flex items-center justify-center transition-all {playingNoteId === note.id ? 'bg-red-500 text-white' : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}"
										>
											{#if playingNoteId === note.id}
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
													<rect x="6" y="5" width="4" height="14" rx="1" />
													<rect x="14" y="5" width="4" height="14" rx="1" />
												</svg>
											{:else}
												<svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
													<path d="M8 5v14l11-7z" />
												</svg>
											{/if}
										</button>

										<div class="flex-1 min-w-0">
											<!-- Duration and time -->
											<div class="flex items-center gap-2 mb-1">
												<span class="text-xs font-medium text-gray-900 dark:text-gray-100">
													{formatDuration(note.duration || 0)}
												</span>
												<span class="text-xs text-gray-500">
													{formatTimeAgo(note.created_at)}
												</span>
											</div>

											<!-- Transcript -->
											{#if note.transcript}
												<p class="text-sm text-gray-700 dark:text-gray-300 line-clamp-3">{note.transcript}</p>
											{:else}
												<p class="text-sm text-gray-500 italic">No transcript available</p>
											{/if}
										</div>

										<!-- Delete button -->
										<button
											onclick={() => deleteVoiceNote(note.id)}
											class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-gray-500 hover:text-red-500 dark:hover:text-red-400 opacity-0 group-hover:opacity-100 transition-all"
											title="Delete"
										>
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
												<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
											</svg>
										</button>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{/if}

		<!-- Cover Picker Modal -->
		{#if showCoverPicker}
			<!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
			<div
				class="fixed inset-0 bg-black/50 dark:bg-black/70 z-50 flex items-center justify-center"
				onclick={(e) => { if (e.target === e.currentTarget) showCoverPicker = false; }}
				onkeydown={(e) => { if (e.key === 'Escape') showCoverPicker = false; }}
				role="dialog"
				aria-modal="true"
				aria-label="Choose cover image"
				tabindex="-1"
			>
				<div class="bg-white dark:bg-[#2c2c2e] rounded-xl shadow-2xl w-[480px] max-h-[80vh] flex flex-col">
					<!-- Header -->
					<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-700/50 flex items-center justify-between">
						<h3 class="font-semibold text-gray-900 dark:text-gray-100">Choose cover</h3>
						<button
							onclick={() => showCoverPicker = false}
							class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
							aria-label="Close"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
							</svg>
						</button>
					</div>

					<!-- Tabs -->
					<div class="px-5 pt-3 flex gap-1">
						<button
							onclick={() => coverTab = 'presets'}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {coverTab === 'presets' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}"
						>
							Gradients
						</button>
						<button
							onclick={() => coverTab = 'upload'}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {coverTab === 'upload' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}"
						>
							Upload
						</button>
						<button
							onclick={() => coverTab = 'url'}
							class="px-4 py-2 text-sm font-medium rounded-lg transition-colors {coverTab === 'url' ? 'bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-gray-100' : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}"
						>
							Link
						</button>
					</div>

					<!-- Content -->
					<div class="flex-1 overflow-y-auto p-5">
						{#if coverTab === 'presets'}
							<div class="grid grid-cols-4 gap-2">
								{#each coverPresets as bg}
									<button
										onclick={() => selectCoverPreset(bg.id)}
										class="aspect-video rounded-lg transition-all hover:scale-105 hover:shadow-lg ring-2 ring-transparent hover:ring-blue-500 {coverImage === `preset:${bg.id}` ? 'ring-blue-500' : ''}"
										style={getPresetStyle(bg.id)}
										title={bg.name}
									></button>
								{/each}
							</div>
						{:else if coverTab === 'upload'}
							<div class="space-y-4">
								<p class="text-sm text-gray-500 dark:text-gray-400">Upload an image from your computer</p>

								<!-- Hidden file input -->
								<input
									bind:this={fileInput}
									type="file"
									accept="image/*"
									onchange={handleCoverUpload}
									class="hidden"
								/>

								<!-- Upload area -->
								<button
									onclick={() => fileInput?.click()}
									disabled={uploadingCover}
									class="w-full aspect-video rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-600 hover:border-blue-500 dark:hover:border-blue-400 transition-colors flex flex-col items-center justify-center gap-3 bg-gray-50 dark:bg-gray-800"
								>
									{#if uploadingCover}
										<div class="animate-spin h-8 w-8 border-2 border-gray-400 border-t-blue-500 rounded-full"></div>
										<span class="text-sm text-gray-500">Uploading...</span>
									{:else}
										<svg class="w-10 h-10 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
											<path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
										</svg>
										<span class="text-sm text-gray-600 dark:text-gray-300 font-medium">Click to upload</span>
										<span class="text-xs text-gray-400">PNG, JPG, GIF up to 5MB</span>
									{/if}
								</button>
							</div>
						{:else}
							<div class="space-y-4">
								<p class="text-sm text-gray-500 dark:text-gray-400">Paste an image URL to use as your cover</p>
								<div class="flex gap-2">
									<input
										type="text"
										bind:value={coverInputValue}
										placeholder="https://example.com/image.jpg"
										class="flex-1 px-3 py-2 text-sm bg-gray-50 dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg text-gray-900 dark:text-gray-200 placeholder:text-gray-400 dark:placeholder:text-gray-500"
										onkeydown={(e) => e.key === 'Enter' && updateCoverImage()}
									/>
									<button
										onclick={updateCoverImage}
										class="px-4 py-2 text-sm bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-colors"
										disabled={!coverInputValue.trim()}
									>
										Apply
									</button>
								</div>
								{#if coverInputValue}
									<div class="mt-4">
										<p class="text-xs text-gray-500 mb-2">Preview:</p>
										<div class="aspect-video rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-700">
											<img
												src={coverInputValue}
												alt="Preview"
												class="w-full h-full object-cover"
												onerror={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
											/>
										</div>
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Footer -->
					{#if coverImage}
						<div class="px-5 py-3 border-t border-gray-200 dark:border-gray-700/50">
							<button
								onclick={removeCoverImage}
								class="text-sm text-red-500 dark:text-red-400 hover:text-red-600 dark:hover:text-red-300 flex items-center gap-1.5"
							>
								<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
								</svg>
								Remove cover
							</button>
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
{/if}

<style>
	/* Ensure proper focus styles for contenteditable elements */
	.blocks-container :global([contenteditable]:focus) {
		outline: none;
	}

	/* Light mode block styling (default) */
	.blocks-container :global(.block-wrapper) {
		color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]) {
		color: #1f2937;
		caret-color: #1f2937;
	}

	.blocks-container :global(.block-wrapper [contenteditable]:empty::before) {
		color: #9ca3af;
	}

	/* Headings */
	.blocks-container :global(.block-wrapper h1),
	.blocks-container :global(.block-wrapper h2),
	.blocks-container :global(.block-wrapper h3) {
		color: #111827;
	}

	/* Quote */
	.blocks-container :global(.block-wrapper blockquote) {
		border-left-color: #d1d5db;
		color: #6b7280;
	}

	/* Code */
	.blocks-container :global(.block-wrapper pre) {
		background-color: #f3f4f6;
		border-color: #e5e7eb;
	}

	.blocks-container :global(.block-wrapper code) {
		color: #374151;
	}

	/* Divider */
	.blocks-container :global(.block-wrapper hr) {
		border-color: #e5e7eb;
	}

	/* Drag handle */
	.blocks-container :global(.block-wrapper .text-gray-400) {
		color: #9ca3af;
	}

	.blocks-container :global(.block-wrapper .hover\:bg-gray-100:hover) {
		background-color: #f3f4f6;
	}

	/* Todo checkbox */
	.blocks-container :global(.block-wrapper input[type="checkbox"]) {
		background-color: #ffffff;
		border-color: #d1d5db;
	}

	/* Dark mode block styling */
	:global(.dark) .blocks-container :global(.block-wrapper) {
		color: #f5f5f7;
	}

	:global(.dark) .blocks-container :global(.block-wrapper [contenteditable]) {
		color: #f5f5f7;
		caret-color: #f5f5f7;
	}

	:global(.dark) .blocks-container :global(.block-wrapper [contenteditable]:empty::before) {
		color: #6b7280;
	}

	/* Dark mode headings */
	:global(.dark) .blocks-container :global(.block-wrapper h1),
	:global(.dark) .blocks-container :global(.block-wrapper h2),
	:global(.dark) .blocks-container :global(.block-wrapper h3) {
		color: #ffffff;
	}

	/* Dark mode quote */
	:global(.dark) .blocks-container :global(.block-wrapper blockquote) {
		border-left-color: #4b5563;
		color: #9ca3af;
	}

	/* Dark mode code */
	:global(.dark) .blocks-container :global(.block-wrapper pre) {
		background-color: #0d0d0d;
		border-color: #374151;
	}

	:global(.dark) .blocks-container :global(.block-wrapper code) {
		color: #e5e7eb;
	}

	/* Dark mode divider */
	:global(.dark) .blocks-container :global(.block-wrapper hr) {
		border-color: #374151;
	}

	/* Dark mode drag handle */
	:global(.dark) .blocks-container :global(.block-wrapper .text-gray-400) {
		color: #6b7280;
	}

	:global(.dark) .blocks-container :global(.block-wrapper .hover\:bg-gray-100:hover) {
		background-color: #374151;
	}

	/* Dark mode todo checkbox */
	:global(.dark) .blocks-container :global(.block-wrapper input[type="checkbox"]) {
		background-color: #374151;
		border-color: #4b5563;
	}
</style>
