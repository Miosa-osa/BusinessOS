<script lang="ts">
	// Voice settings tab — currently a navigation hub to the full voice notes page.
	// Future: expose audio input device selection, transcription quality, etc.
</script>

<div class="space-y-6">
	<div class="card text-center py-8">
		<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-green-100 dark:bg-green-900/30 flex items-center justify-center">
			<svg class="w-8 h-8 text-green-600 dark:text-green-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
				<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
				<line x1="12" y1="19" x2="12" y2="23"/>
				<line x1="8" y1="23" x2="16" y2="23"/>
			</svg>
		</div>
		<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-2">Voice Notes</h2>
		<p class="text-sm text-gray-500 dark:text-gray-400 mb-6 max-w-md mx-auto">
			View your voice transcription history, track speaking stats, and browse notes by date or project.
		</p>
		<a href="/voice-notes" class="inline-flex items-center gap-2 btn btn-primary">
			<svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M12 1a3 3 0 0 0-3 3v8a3 3 0 0 0 6 0V4a3 3 0 0 0-3-3z"/>
				<path d="M19 10v2a7 7 0 0 1-14 0v-2"/>
			</svg>
			View Voice Notes
		</a>
	</div>

	<div class="card">
		<h2 class="text-lg font-medium text-gray-900 dark:text-white mb-4">About Voice Notes</h2>
		<div class="space-y-4 text-sm text-gray-600 dark:text-gray-400">
			<div class="flex items-start gap-3">
				<div class="w-8 h-8 rounded-lg bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center flex-shrink-0">
					<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<polyline points="12 6 12 12 16 14"/>
					</svg>
				</div>
				<div>
					<p class="font-medium text-gray-900 dark:text-white">Track Your Speaking</p>
					<p>See duration, word count, and words per minute for each recording.</p>
				</div>
			</div>
			<div class="flex items-start gap-3">
				<div class="w-8 h-8 rounded-lg bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
					<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
						<line x1="16" y1="2" x2="16" y2="6"/>
						<line x1="8" y1="2" x2="8" y2="6"/>
						<line x1="3" y1="10" x2="21" y2="10"/>
					</svg>
				</div>
				<div>
					<p class="font-medium text-gray-900 dark:text-white">Browse by Date</p>
					<p>Notes organized by today, yesterday, and older dates for easy review.</p>
				</div>
			</div>
			<div class="flex items-start gap-3">
				<div class="w-8 h-8 rounded-lg bg-orange-100 dark:bg-orange-900/30 flex items-center justify-center flex-shrink-0">
					<svg class="w-4 h-4 text-orange-600 dark:text-orange-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
					</svg>
				</div>
				<div>
					<p class="font-medium text-gray-900 dark:text-white">Filter by Project</p>
					<p>View notes associated with specific projects and contexts.</p>
				</div>
			</div>
		</div>
	</div>
</div>
