<!--
	RoundedInput.svelte
	iOS-style rounded input with glassmorphism

	Usage:
	<RoundedInput
		label="Email"
		type="email"
		placeholder="you@company.com"
		bind:value={email}
	/>
-->
<script lang="ts">
	interface Props {
		label?: string;
		type?: string;
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		required?: boolean;
		autocomplete?: string;
		error?: string;
		helperText?: string;
		id?: string;
		name?: string;
		class?: string;
	}

	let {
		label,
		type = 'text',
		value = $bindable(''),
		placeholder,
		disabled = false,
		required = false,
		autocomplete,
		error,
		helperText,
		id,
		name,
		class: className = ''
	}: Props = $props();

	const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
	const classes = `input-rounded ${className}`.trim();
</script>

<div class="flex flex-col gap-1.5">
	{#if label}
		<label for={inputId} class="text-sm font-medium text-gray-700 dark:text-gray-300">
			{label}
			{#if required}
				<span class="text-red-500">*</span>
			{/if}
		</label>
	{/if}

	<input
		id={inputId}
		{name}
		{type}
		bind:value
		{placeholder}
		{disabled}
		{required}
		{autocomplete}
		class={classes}
		class:border-red-500={error}
		aria-invalid={error ? 'true' : 'false'}
		aria-describedby={error ? `${inputId}-error` : helperText ? `${inputId}-helper` : undefined}
	/>

	{#if error}
		<p id="{inputId}-error" class="text-sm text-red-600 dark:text-red-400 animate-slide-down">
			{error}
		</p>
	{:else if helperText}
		<p id="{inputId}-helper" class="text-sm text-gray-500 dark:text-gray-400">
			{helperText}
		</p>
	{/if}
</div>
