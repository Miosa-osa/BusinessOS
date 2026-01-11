# Form Patterns Analysis & Reusable Components

## Overview
Analysis of form patterns found in `frontend/src/routes/(app)/settings/ai/+page.svelte` with recommendations for creating reusable form components for agent-related forms.

---

## 1. Current Form Patterns Analysis

### 1.1 Form State Management

**Pattern**: Reactive state variables using Svelte 5 `$state` rune

```typescript
// Example from settings/ai page
let newCustomAgent = $state({
	name: '',
	display_name: '',
	description: '',
	system_prompt: '',
	model_preference: '',
	temperature: 0.7,
	category: 'custom'
});

let savingCustomAgent = $state(false);
let editingCustomAgent = $state<CustomAgent | null>(null);
```

**Key Characteristics**:
- Direct property binding using `bind:value` syntax
- Boolean flags for `loading`, `saving`, `editing` states
- Separate form state from UI state
- Optional/nullable state for edit forms (`editingCustomAgent = null` to reset)

**Best Practices Identified**:
- Always have a "saving" state separate from "loading" state
- Keep edit form state separate from display state
- Reset form values after successful submission
- Store selected item ID separately from full object (when applicable)

---

### 1.2 Validation Patterns

**Pattern**: Pre-submission validation with required field checks

```typescript
async function createCustomAgent() {
	// Validation checks BEFORE API call
	if (!newCustomAgent.name || !newCustomAgent.display_name || !newCustomAgent.system_prompt) {
		error = 'Name, display name, and system prompt are required';
		return;
	}

	savingCustomAgent = true;
	// ... API call
}
```

**Current Approach**:
- Simple required field checks (non-empty strings)
- No schema validation library (Zod, Valibot)
- Basic error messages
- Validation happens at submission time (not real-time)

**Validation Patterns Found**:

| Pattern | Example | When Used |
|---------|---------|-----------|
| **Required Check** | `if (!newCustomAgent.name)` | All text fields |
| **Trim Check** | `if (!key.trim())` | For API keys |
| **Confirmation** | `if (!confirm('Delete...?'))` | Destructive actions |
| **Format Check** | `name.toLowerCase().replace(/\s+/g, '-')` | Slug generation |

---

### 1.3 Submit Handlers

**Pattern**: Async functions with try-catch-finally structure

```typescript
async function createCustomAgent() {
	// 1. Validation
	if (!newCustomAgent.name || !newCustomAgent.display_name || !newCustomAgent.system_prompt) {
		error = 'Name, display name, and system prompt are required';
		return;
	}

	// 2. Set loading state
	savingCustomAgent = true;

	try {
		// 3. API call with request payload
		const res = await apiClient.post('/ai/custom-agents', newCustomAgent);

		// 4. Handle success
		if (res.ok) {
			const data = await res.json();
			customAgents = [...customAgents, data.agent];
			showNewCustomAgent = false;

			// 5. Reset form
			newCustomAgent = {
				name: '',
				display_name: '',
				description: '',
				system_prompt: '',
				model_preference: '',
				temperature: 0.7,
				category: 'custom'
			};

			// 6. Show success message
			saveStatus = 'Custom agent created!';
			setTimeout(() => saveStatus = '', 3000);
		} else {
			// 7. Handle error response
			const err = await res.json();
			error = err.error || 'Failed to create agent';
		}
	} catch (err) {
		// 8. Handle network/parse errors
		console.error('Failed to create custom agent:', err);
		error = 'Failed to create custom agent';
	} finally {
		// 9. Clear loading state
		savingCustomAgent = false;
	}
}
```

**Handler Structure**:
1. Validate input
2. Set loading state
3. Make API request
4. Parse response
5. Update state (list, form reset)
6. Show feedback (success message)
7. Handle errors with user message
8. Clear loading state
9. Auto-clear messages after timeout

---

### 1.4 Error Display Patterns

**Pattern 1: Global Error Toast**
```svelte
{#if error}
	<div class="error-alert">
		<span>{error}</span>
		<button onclick={() => error = ''}>×</button>
	</div>
{/if}
```

**Pattern 2: Success Toast**
```svelte
{#if saveStatus}
	<div class="save-toast">{saveStatus}</div>
{/if}
```

**Pattern 3: Auto-clearing Messages**
```typescript
saveStatus = 'Command updated successfully';
editingCommand = null;
// Auto-clear after 3 seconds
setTimeout(() => { saveStatus = ''; error = ''; }, 3000);
```

**Error Message Strategy**:
- User-friendly error messages from API responses: `err.error || 'Failed...'`
- Generic fallback messages for network errors
- Separate error and success feedback streams
- Auto-dismiss after 3000ms to reduce noise

---

### 1.5 Multi-form Patterns

**Pattern**: Multiple independent forms with toggle visibility

```typescript
// Form visibility states
let showNewCommand = $state(false);
let editingCommand = $state<CommandInfo | null>(null);

// Form data
let newCommand = $state<Partial<CommandInfo>>({
	name: '',
	display_name: '',
	description: '',
	icon: '✨',
	system_prompt: '',
	context_sources: []
});

// Usage
{#if showNewCommand}
	<!-- New form -->
{:else if editingCommand}
	<!-- Edit form -->
{:else}
	<!-- List view -->
{/if}
```

**Patterns**:
- Use `null` for "not editing" vs boolean for "form visible"
- Separate visibility from data
- Reset form state when closing

---

## 2. Reusable Form Component Pattern

### 2.1 FormBuilder Component

A generic form builder component for agent configuration forms:

```svelte
<!-- ReusableAgentForm.svelte -->
<script lang="ts">
	import FormField from '$lib/components/forms/FormField.svelte';
	import FormSection from '$lib/components/forms/FormSection.svelte';
	import type { AgentFormData, AgentFormField } from '$lib/types/forms';

	interface Props {
		data: AgentFormData;
		fields: AgentFormField[];
		loading?: boolean;
		error?: string;
		submitLabel?: string;
		onSubmit: (data: AgentFormData) => Promise<void>;
		onCancel?: () => void;
	}

	let { data = $bindable(), fields, loading = false, error = '', submitLabel = 'Save', onSubmit, onCancel }: Props = $props();

	let isSubmitting = $state(false);
	let localError = $state('');

	async function handleSubmit() {
		isSubmitting = true;
		try {
			await onSubmit(data);
		} catch (err) {
			localError = err instanceof Error ? err.message : 'An error occurred';
			setTimeout(() => { localError = ''; }, 3000);
		} finally {
			isSubmitting = false;
		}
	}
</script>

<form onsubmit|preventDefault={handleSubmit} class="space-y-6">
	{#if localError || error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-3">
			<p class="text-red-800 text-sm">{localError || error}</p>
		</div>
	{/if}

	{#each fields as field (field.name)}
		<FormField
			bind:value={data[field.name]}
			{...field}
		/>
	{/each}

	<div class="flex gap-2 justify-end">
		{#if onCancel}
			<button type="button" onclick={onCancel} disabled={isSubmitting} class="btn btn-secondary">
				Cancel
			</button>
		{/if}
		<button type="submit" disabled={isSubmitting} class="btn btn-primary">
			{#if isSubmitting}
				<span class="loading loading-spinner loading-sm"></span>
				Saving...
			{:else}
				{submitLabel}
			{/if}
		</button>
	</div>
</form>
```

---

## 3. Recommended Component Structure

### 3.1 FormField Component

```svelte
<!-- components/forms/FormField.svelte -->
<script lang="ts">
	import type { FormFieldConfig } from '$lib/types/forms';

	interface Props extends FormFieldConfig {
		value: string | number | string[];
		error?: string;
	}

	let { label, name, type = 'text', required = false, value = $bindable(), error = '', help = '', placeholder = '' }: Props = $props();
</script>

<div class="form-group">
	<label for={name} class="label">
		<span class="label-text font-medium">{label}</span>
		{#if required}
			<span class="text-red-500">*</span>
		{/if}
	</label>

	{#if type === 'textarea'}
		<textarea
			{name}
			{required}
			{placeholder}
			bind:value
			class="textarea textarea-bordered w-full {error ? 'textarea-error' : ''}"
		/>
	{:else if type === 'select'}
		<select bind:value class="select select-bordered w-full {error ? 'select-error' : ''}">
			<option disabled selected>Choose one</option>
			<!-- Options provided via context or prop -->
		</select>
	{:else}
		<input
			type={type}
			{name}
			{required}
			{placeholder}
			bind:value
			class="input input-bordered w-full {error ? 'input-error' : ''}"
		/>
	{/if}

	{#if error}
		<label class="label">
			<span class="label-text-alt text-red-600">{error}</span>
		</label>
	{/if}

	{#if help}
		<label class="label">
			<span class="label-text-alt text-gray-500">{help}</span>
		</label>
	{/if}
</div>
```

### 3.2 FormSection Component

```svelte
<!-- components/forms/FormSection.svelte -->
<script lang="ts">
	interface Props {
		title: string;
		description?: string;
	}

	let { title, description = '' }: Props = $props();
</script>

<div class="space-y-3">
	<div>
		<h3 class="text-lg font-semibold">{title}</h3>
		{#if description}
			<p class="text-sm text-gray-600">{description}</p>
		{/if}
	</div>
	<slot />
</div>

<style>
	/* Visual separation */
	div :global(+ *) {
		border-top: 1px solid var(--fallback-bc, oklch(var(--bc) / 0.1));
		padding-top: 1rem;
	}
</style>
```

---

## 4. Type Definitions for Forms

```typescript
// lib/types/forms.ts

export interface FormFieldConfig {
	name: string;
	label: string;
	type: 'text' | 'email' | 'password' | 'number' | 'textarea' | 'select' | 'checkbox' | 'radio';
	required?: boolean;
	placeholder?: string;
	help?: string;
	validation?: (value: any) => string | null; // Error message or null
	options?: Array<{ value: string; label: string }>; // For select/radio
	min?: number; // For number inputs
	max?: number;
}

export interface AgentFormData {
	name: string;
	display_name: string;
	description?: string;
	system_prompt: string;
	model_preference?: string;
	temperature?: number;
	max_tokens?: number;
	category?: string;
	avatar?: string;
	[key: string]: any;
}

export interface FormState<T> {
	data: T;
	isLoading: boolean;
	isSaving: boolean;
	isEditing: boolean;
	error: string;
	successMessage: string;
	fieldErrors: Record<string, string>;
}

export interface FormConfig<T> {
	title: string;
	description?: string;
	fields: FormFieldConfig[];
	onSubmit: (data: T) => Promise<void>;
	onCancel?: () => void;
	submitLabel?: string;
	cancelLabel?: string;
}
```

---

## 5. Usage Example for Agent Form

```svelte
<!-- routes/(app)/agents/create/+page.svelte -->
<script lang="ts">
	import ReusableAgentForm from '$lib/components/forms/ReusableAgentForm.svelte';
	import type { AgentFormData } from '$lib/types/forms';
	import { apiClient } from '$lib/api';

	let agentData: AgentFormData = $state({
		name: '',
		display_name: '',
		description: '',
		system_prompt: '',
		model_preference: '',
		temperature: 0.7,
		category: 'custom'
	});

	let error = $state('');

	async function handleCreateAgent(data: AgentFormData) {
		if (!data.name || !data.display_name || !data.system_prompt) {
			throw new Error('Name, display name, and system prompt are required');
		}

		const res = await apiClient.post('/ai/custom-agents', data);
		if (!res.ok) {
			const err = await res.json();
			throw new Error(err.error || 'Failed to create agent');
		}
	}

	const agentFormFields = [
		{
			name: 'name',
			label: 'Agent Name',
			type: 'text',
			required: true,
			placeholder: 'my-agent',
			help: 'Lowercase, use hyphens for spaces'
		},
		{
			name: 'display_name',
			label: 'Display Name',
			type: 'text',
			required: true,
			placeholder: 'My Agent'
		},
		{
			name: 'description',
			label: 'Description',
			type: 'textarea',
			placeholder: 'What does this agent do?'
		},
		{
			name: 'system_prompt',
			label: 'System Prompt',
			type: 'textarea',
			required: true,
			placeholder: 'You are a helpful assistant that...'
		},
		{
			name: 'category',
			label: 'Category',
			type: 'select',
			options: [
				{ value: 'general', label: 'General' },
				{ value: 'specialist', label: 'Specialist' },
				{ value: 'custom', label: 'Custom' }
			]
		},
		{
			name: 'temperature',
			label: 'Temperature',
			type: 'number',
			min: 0,
			max: 2,
			help: 'Higher = more creative, Lower = more deterministic'
		}
	];
</script>

<div class="page">
	<ReusableAgentForm
		bind:data={agentData}
		fields={agentFormFields}
		submitLabel="Create Agent"
		error={error}
		onSubmit={handleCreateAgent}
		onCancel={() => window.history.back()}
	/>
</div>
```

---

## 6. Advanced Patterns

### 6.1 Validation with Schema

```typescript
// For future use with Zod validation
import { z } from 'zod';

const agentSchema = z.object({
	name: z.string().min(1, 'Name is required').regex(/^[a-z0-9-]+$/, 'Name must be lowercase with hyphens'),
	display_name: z.string().min(1, 'Display name is required'),
	system_prompt: z.string().min(10, 'System prompt must be at least 10 characters'),
	temperature: z.number().min(0).max(2),
	category: z.enum(['general', 'specialist', 'custom'])
});

async function validateForm(data: AgentFormData) {
	try {
		await agentSchema.parseAsync(data);
		return { valid: true, errors: {} };
	} catch (err) {
		if (err instanceof z.ZodError) {
			const fieldErrors = Object.fromEntries(
				err.errors.map(e => [e.path[0], e.message])
			);
			return { valid: false, errors: fieldErrors };
		}
		return { valid: false, errors: { general: 'Validation failed' } };
	}
}
```

### 6.2 Context Source Multi-select Pattern

```typescript
// Pattern from settings/ai for multi-select handling
function toggleContextSource(sources: string[] | undefined, source: string): string[] {
	const current = sources || [];
	if (current.includes(source)) {
		return current.filter(s => s !== source); // Remove if exists
	}
	return [...current, source]; // Add if doesn't exist
}

// Usage in form
let contextSources = $state<string[]>([]);

function toggleSource(source: string) {
	contextSources = toggleContextSource(contextSources, source);
}
```

---

## 7. Implementation Checklist

For creating new agent-related forms:

- [ ] Define form data type (interface)
- [ ] Create field configuration array
- [ ] Implement validation function (pre-submit)
- [ ] Create submit handler with proper error handling
- [ ] Add loading state management
- [ ] Implement success/error message display
- [ ] Add form reset after successful submission
- [ ] Wire up cancel button
- [ ] Test with API responses (success and error cases)
- [ ] Add field-level error display (optional but recommended)
- [ ] Consider using ReusableAgentForm component

---

## 8. Current State Management Summary

```
Settings AI Page Form State Patterns:

┌─ Custom Agent Form
├─ State: newCustomAgent (form data)
├─ States: savingCustomAgent, editingCustomAgent, showNewCustomAgent
├─ Handlers: createCustomAgent(), updateCustomAgent(), deleteCustomAgent()
└─ Messages: error, saveStatus (auto-clear after 3s)

┌─ Command Form
├─ State: newCommand (form data)
├─ States: savingCommand, editingCommand, showNewCommand
├─ Handlers: saveNewCommand(), updateCommand(), deleteCommand()
└─ Messages: error, saveStatus (auto-clear after 3s)

┌─ API Key Form
├─ State: apiKeys (Record<string, string>)
├─ State: savingKey (current key being saved)
├─ Handlers: saveAPIKey()
└─ Messages: error, saveStatus

Pattern Consistency:
✓ All forms use try-catch-finally
✓ All forms validate before submission
✓ All forms reset after success
✓ All forms use 3-second auto-clear for messages
✓ Separate loading state for each async operation
```

---

## 9. Recommendations

### High Priority
1. **Create FormField component** - Extract common field logic
2. **Standardize validation** - Consider Zod for schema-based validation
3. **Error boundary** - Create error display component
4. **Loading indicators** - Consistent loading UI across forms

### Medium Priority
1. **Form helpers** - Utility functions for common patterns
2. **Type safety** - Stronger typing for form data
3. **Field-level errors** - Support displaying per-field error messages
4. **Auto-save** - Debounced save on field changes (for long forms)

### Low Priority
1. **Multi-step forms** - For complex agent configuration
2. **Form context** - useForm-like pattern with Svelte context
3. **Conditional fields** - Show/hide fields based on other values
4. **Custom validation** - Field-level validation rules

---

## 10. Code File Locations

| Component | Path | Status |
|-----------|------|--------|
| Existing FormInput | `frontend/src/lib/components/auth/FormInput.svelte` | ✓ Exists |
| Existing PasswordInput | `frontend/src/lib/components/auth/PasswordInput.svelte` | ✓ Exists |
| Settings AI Page | `frontend/src/routes/(app)/settings/ai/+page.svelte` | ✓ Exists |
| Recommended FormField | `frontend/src/lib/components/forms/FormField.svelte` | Needs creation |
| Recommended FormSection | `frontend/src/lib/components/forms/FormSection.svelte` | Needs creation |
| Recommended ReusableAgentForm | `frontend/src/lib/components/forms/ReusableAgentForm.svelte` | Needs creation |
| Types | `frontend/src/lib/types/forms.ts` | Needs creation |

