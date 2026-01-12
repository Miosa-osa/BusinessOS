# Form Components Usage Guide

Complete guide for using the reusable form components created for BusinessOS agent configuration.

---

## Quick Start

### 1. Basic Agent Form

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
		const res = await apiClient.post('/ai/custom-agents', data);
		if (!res.ok) {
			const err = await res.json();
			throw new Error(err.error || 'Failed to create agent');
		}
	}

	const fields = [
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
			name: 'system_prompt',
			label: 'System Prompt',
			type: 'textarea',
			required: true,
			placeholder: 'You are a helpful assistant...'
		}
	];
</script>

<div class="page">
	<ReusableAgentForm
		bind:data={agentData}
		{fields}
		submitLabel="Create Agent"
		onSubmit={handleCreateAgent}
		onCancel={() => window.history.back()}
	/>
</div>
```

---

## Component Reference

### ReusableAgentForm

Complete form component for agent configuration.

#### Props

```typescript
interface Props {
	data: AgentFormData;              // Form data object (bindable)
	fields: FormFieldConfig[];        // Array of field configurations
	title?: string;                   // Form title
	description?: string;             // Form description
	loading?: boolean;                // Loading state
	error?: string;                   // Error message to display
	successMessage?: string;          // Success message to display
	submitLabel?: string;             // Submit button text (default: 'Save Agent')
	cancelLabel?: string;             // Cancel button text (default: 'Cancel')
	onSubmit: (data: AgentFormData) => Promise<void>;  // Submit handler
	onCancel?: () => void;            // Cancel handler
}
```

#### Features

- Automatic field grouping into sections
- Field validation with error display
- Loading state management
- Success/error message display
- Auto-clearing messages after 3 seconds
- Form validation before submission

---

### FormField

Individual form input component.

#### Props

```typescript
interface Props extends Partial<FormFieldConfig> {
	value: string | number | boolean | string[];
	error?: string;
}
```

#### Supported Field Types

| Type | Description | Example |
|------|-------------|---------|
| `text` | Text input | Name, email |
| `email` | Email input | Email address |
| `password` | Password input | Password fields |
| `number` | Number input | Temperature, counts |
| `textarea` | Multi-line text | System prompt, description |
| `select` | Dropdown select | Category, model |
| `checkbox` | Single checkbox | Enable feature |
| `radio` | Radio buttons | Select one option |

#### Example Usage

```svelte
<FormField
	name="temperature"
	label="Temperature"
	type="number"
	value={data.temperature}
	min={0}
	max={2}
	help="Higher = more creative, Lower = more deterministic"
	bind:value={data.temperature}
/>
```

---

### FormSection

Group related form fields into sections.

#### Props

```typescript
interface Props {
	title: string;           // Section title
	description?: string;    // Optional section description
}
```

#### Example Usage

```svelte
<FormSection title="Basic Information" description="Agent identity">
	<FormField name="name" label="Name" bind:value={data.name} />
	<FormField name="display_name" label="Display Name" bind:value={data.display_name} />
</FormSection>
```

---

## Form Helpers

Utility functions in `$lib/utils/form-helpers.ts`:

### validateField

Validate a single field value.

```typescript
import { validateField } from '$lib/utils/form-helpers';

const error = validateField('', 'Name', {
	required: true,
	minLength: 1,
	maxLength: 100
});

if (error) {
	console.log(error); // "Name is required"
}
```

### validateFields

Validate all form fields.

```typescript
import { validateFields } from '$lib/utils/form-helpers';

const data = { name: '', temperature: 0.7 };
const rules = {
	name: { required: true, minLength: 1 },
	temperature: { min: 0, max: 2 }
};

const result = validateFields(data, rules);
if (!result.valid) {
	console.log(result.errors); // { name: "name is required" }
}
```

### resetForm

Reset form to initial state.

```typescript
import { resetForm } from '$lib/utils/form-helpers';

let formData = {
	name: 'John',
	email: 'john@example.com',
	tags: ['tag1']
};

formData = resetForm(formData);
// Now: { name: '', email: '', tags: [] }
```

### hasFormChanged

Detect if form data has changed.

```typescript
import { hasFormChanged } from '$lib/utils/form-helpers';

const original = { name: 'Agent', temperature: 0.7 };
const current = { name: 'Agent', temperature: 0.8 };

const changed = hasFormChanged(original, current);
// true - temperature changed
```

### toggleArrayItem

Toggle item in array (for multi-select).

```typescript
import { toggleArrayItem } from '$lib/utils/form-helpers';

let tags = ['coding', 'writing'];

tags = toggleArrayItem(tags, 'coding');
// Now: ['writing']

tags = toggleArrayItem(tags, 'coding');
// Now: ['writing', 'coding']
```

### createAutoSaveDebounce

Debounce form auto-save.

```typescript
import { createAutoSaveDebounce } from '$lib/utils/form-helpers';

const autoSave = createAutoSaveDebounce(async (data) => {
	await apiClient.put('/agents/1', data);
}, 1500); // Save after 1.5 seconds of inactivity

// In form change handler:
function onFieldChange(field: string, value: any) {
	data[field] = value;
	autoSave(data);
}
```

---

## Common Patterns

### Pattern 1: Create Form

```svelte
<script lang="ts">
	import ReusableAgentForm from '$lib/components/forms/ReusableAgentForm.svelte';
	import { apiClient } from '$lib/api';

	let agentData = $state({ name: '', display_name: '', system_prompt: '' });

	async function handleCreate(data) {
		const res = await apiClient.post('/ai/agents', data);
		if (!res.ok) throw new Error('Failed to create');
	}
</script>

<ReusableAgentForm
	bind:data={agentData}
	fields={fields}
	onSubmit={handleCreate}
	submitLabel="Create Agent"
/>
```

### Pattern 2: Edit Form

```svelte
<script lang="ts">
	import { page } from '$app/stores';

	let agentData = $state<AgentFormData | null>(null);

	onMount(async () => {
		const res = await apiClient.get(`/ai/agents/${page.params.id}`);
		const data = await res.json();
		agentData = data.agent;
	});

	async function handleUpdate(data) {
		const res = await apiClient.put(`/ai/agents/${page.params.id}`, data);
		if (!res.ok) throw new Error('Failed to update');
	}
</script>

{#if agentData}
	<ReusableAgentForm
		bind:data={agentData}
		fields={fields}
		title="Edit Agent"
		submitLabel="Update Agent"
		onSubmit={handleUpdate}
	/>
{/if}
```

### Pattern 3: Multi-field Validation

```typescript
import { ValidationRules, createFormSubmitHandler } from '$lib/utils/form-helpers';

const handleSubmit = createFormSubmitHandler({
	onValidate: (data) => {
		if (!data.name) return 'Name is required';
		if (!data.system_prompt) return 'System prompt is required';
		if (data.temperature < 0 || data.temperature > 2) {
			return 'Temperature must be between 0 and 2';
		}
		return null; // Valid
	},
	onSubmit: async (data) => {
		const res = await apiClient.post('/agents', data);
		if (!res.ok) throw new Error('Failed to save');
	},
	onSuccess: () => {
		console.log('Agent saved!');
	},
	onError: (error) => {
		console.error(error);
	}
});
```

### Pattern 4: Conditional Fields

```svelte
<script lang="ts">
	let agentData = $state({
		name: '',
		useCustomModel: false,
		customModel: ''
	});
</script>

<FormField
	name="useCustomModel"
	type="checkbox"
	label="Use custom model"
	bind:value={agentData.useCustomModel}
/>

{#if agentData.useCustomModel}
	<FormField
		name="customModel"
		type="text"
		label="Model name"
		required={agentData.useCustomModel}
		bind:value={agentData.customModel}
	/>
{/if}
```

### Pattern 5: Array Fields (Multi-select)

```svelte
<script lang="ts">
	import { toggleArrayItem } from '$lib/utils/form-helpers';

	let agentData = $state({
		capabilities: ['coding']
	});

	function toggleCapability(cap: string) {
		agentData.capabilities = toggleArrayItem(agentData.capabilities, cap);
	}
</script>

<FormSection title="Capabilities">
	{#each ['coding', 'writing', 'analysis'] as capability}
		<label class="label cursor-pointer">
			<input
				type="checkbox"
				checked={agentData.capabilities.includes(capability)}
				onchange={() => toggleCapability(capability)}
				class="checkbox"
			/>
			<span class="label-text">{capability}</span>
		</label>
	{/each}
</FormSection>
```

---

## Field Configuration Examples

### Basic Text Field

```typescript
{
	name: 'name',
	label: 'Agent Name',
	type: 'text',
	required: true,
	placeholder: 'Enter agent name',
	minLength: 1,
	maxLength: 100
}
```

### Select Field

```typescript
{
	name: 'category',
	label: 'Category',
	type: 'select',
	required: true,
	options: [
		{ value: 'general', label: 'General' },
		{ value: 'specialist', label: 'Specialist' },
		{ value: 'custom', label: 'Custom' }
	]
}
```

### Textarea Field

```typescript
{
	name: 'system_prompt',
	label: 'System Prompt',
	type: 'textarea',
	required: true,
	placeholder: 'Define agent behavior...',
	rows: 6,
	minLength: 10,
	maxLength: 5000,
	help: 'Be specific about the agent\'s role and behavior'
}
```

### Number Field with Range

```typescript
{
	name: 'temperature',
	label: 'Temperature',
	type: 'number',
	required: true,
	min: 0,
	max: 2,
	help: 'Controls randomness: 0 = deterministic, 2 = creative'
}
```

### Custom Validation

```typescript
{
	name: 'email',
	label: 'Email',
	type: 'email',
	required: true,
	validation: (value) => {
		if (!value.includes('@')) return 'Invalid email';
		return null;
	}
}
```

---

## Error Handling

### Field-level Errors

Errors are automatically displayed below each field:

```svelte
<FormField
	name="name"
	label="Name"
	error={fieldErrors['name']}
	bind:value={data.name}
/>
```

### Form-level Errors

Display at the top of the form:

```svelte
{#if error}
	<div class="alert alert-error">
		{error}
	</div>
{/if}
```

### API Error Handling

```typescript
async function handleSubmit(data) {
	try {
		const res = await apiClient.post('/agents', data);
		if (!res.ok) {
			const errorData = await res.json();
			throw new Error(errorData.error || 'Failed to save');
		}
	} catch (err) {
		error = err instanceof Error ? err.message : 'An error occurred';
		setTimeout(() => { error = ''; }, 4000);
	}
}
```

---

## Styling & Customization

### DaisyUI Classes Used

The components use DaisyUI utility classes:

- `.btn` - Button
- `.input` - Text input
- `.textarea` - Text area
- `.select` - Select dropdown
- `.checkbox` - Checkbox
- `.radio` - Radio button
- `.label` - Label wrapper
- `.alert` - Alert box
- `.form-control` - Form field wrapper

### Custom Styling

Override styles using CSS modules or Tailwind classes:

```svelte
<style>
	:global(.form-control) {
		@apply mb-6;
	}

	:global(.label) {
		@apply font-semibold;
	}
</style>
```

---

## Migration Guide: From Inline Forms

If you have inline forms like in `settings/ai`, migrate to reusable components:

### Before (Inline)

```svelte
<script>
	let newAgent = $state({ name: '', display_name: '' });
	let savingAgent = $state(false);
	let error = $state('');

	async function createAgent() {
		if (!newAgent.name) {
			error = 'Name required';
			return;
		}
		savingAgent = true;
		try {
			const res = await apiClient.post('/agents', newAgent);
			// ... handle response
		} finally {
			savingAgent = false;
		}
	}
</script>

<form onsubmit|preventDefault={createAgent}>
	<input bind:value={newAgent.name} />
	<button disabled={savingAgent}>Save</button>
</form>
```

### After (Reusable)

```svelte
<script>
	import ReusableAgentForm from '$lib/components/forms/ReusableAgentForm.svelte';

	let agentData = $state({ name: '', display_name: '' });

	async function handleCreate(data) {
		const res = await apiClient.post('/agents', data);
		if (!res.ok) throw new Error('Failed to create');
	}
</script>

<ReusableAgentForm
	bind:data={agentData}
	fields={[{ name: 'name', label: 'Name', required: true }]}
	onSubmit={handleCreate}
/>
```

---

## Best Practices

1. **Always use bindable data** - Use `bind:data` for two-way binding
2. **Handle async operations** - Wrap API calls in try-catch
3. **Validate before submit** - Use validation props on fields
4. **Show feedback** - Display success/error messages
5. **Reset on success** - Clear form after successful submission
6. **Group related fields** - Use FormSection for organization
7. **Provide helpful text** - Use `help` and `placeholder` props
8. **Test edge cases** - Empty, very long, special character inputs

---

## Troubleshooting

### Form not updating

Make sure you're using `bind:value` or `bind:group`:

```svelte
<!-- ❌ Wrong -->
<FormField value={data.name} />

<!-- ✅ Correct -->
<FormField bind:value={data.name} />
```

### Validation not running

Check that validation rules are properly defined:

```typescript
// ❌ Missing validation function
{ name: 'email', type: 'email' }

// ✅ With validation
{
	name: 'email',
	type: 'email',
	validation: (v) => !v.includes('@') ? 'Invalid email' : null
}
```

### Messages not clearing

Ensure setTimeout is used:

```typescript
// Messages should auto-clear after 3-4 seconds
setTimeout(() => { error = ''; }, 3000);
```

---

## File Structure

```
frontend/
├── src/
│   ├── lib/
│   │   ├── components/
│   │   │   └── forms/
│   │   │       ├── FormField.svelte
│   │   │       ├── FormSection.svelte
│   │   │       └── ReusableAgentForm.svelte
│   │   ├── types/
│   │   │   └── forms.ts
│   │   └── utils/
│   │       └── form-helpers.ts
│   └── routes/
│       └── (app)/
│           └── agents/
│               ├── create/
│               │   └── +page.svelte
│               └── [id]/
│                   └── edit/
│                       └── +page.svelte
└── docs/
    └── FORM_COMPONENTS_USAGE_GUIDE.md
```

