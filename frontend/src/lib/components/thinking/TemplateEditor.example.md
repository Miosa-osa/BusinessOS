# TemplateEditor Component

A comprehensive form component for creating and editing reasoning templates with multi-step workflows.

## Usage

### Create Mode
```svelte
<script>
  import TemplateEditor from '$lib/components/thinking/TemplateEditor.svelte';
  import type { CreateTemplateData } from '$lib/api/thinking';

  function handleSave(data: CreateTemplateData) {
    // Handle template creation
    console.log('Creating template:', data);
  }

  function handleCancel() {
    // Handle cancel
    console.log('Cancelled');
  }
</script>

<TemplateEditor
  template={null}
  onSave={handleSave}
  onCancel={handleCancel}
/>
```

### Edit Mode
```svelte
<script>
  import TemplateEditor from '$lib/components/thinking/TemplateEditor.svelte';
  import type { ReasoningTemplate, UpdateTemplateData } from '$lib/api/thinking';

  let existingTemplate: ReasoningTemplate = {
    id: 'template-123',
    user_id: 'user-456',
    name: 'Problem Analysis',
    description: 'Framework for analyzing complex problems',
    steps: [
      {
        id: 'step-1',
        order: 0,
        type: 'exploration',
        prompt: 'What is the core problem we are trying to solve?'
      },
      {
        id: 'step-2',
        order: 1,
        type: 'analysis',
        prompt: 'What are the key factors and constraints?'
      }
    ],
    created_at: '2025-01-09T00:00:00Z',
    updated_at: '2025-01-09T00:00:00Z'
  };

  function handleSave(data: UpdateTemplateData) {
    // Handle template update
    console.log('Updating template:', data);
  }

  function handleCancel() {
    // Handle cancel
    console.log('Cancelled');
  }
</script>

<TemplateEditor
  template={existingTemplate}
  onSave={handleSave}
  onCancel={handleCancel}
/>
```

## Props

| Prop | Type | Required | Description |
|------|------|----------|-------------|
| `template` | `ReasoningTemplate \| null` | Yes | Template to edit (null for create mode) |
| `onSave` | `(data: CreateTemplateData \| UpdateTemplateData) => void` | Yes | Called when form is submitted |
| `onCancel` | `() => void` | Yes | Called when cancel is clicked |

## Features

### Form Fields
- **Name** (required): Template name (max 100 chars)
- **Description** (optional): Template description (max 500 chars)
- **Steps** (required): At least 1 step with:
  - Order number (auto-managed)
  - Type selector (exploration, analysis, conclusion, reflection)
  - Prompt textarea (max 2000 chars)

### Step Management
- **Add Step**: Add new step at the end
- **Remove Step**: Delete a step (with auto-reordering)
- **Move Up/Down**: Reorder steps with arrow buttons
- Auto-updates order numbers when steps are moved or removed

### Validation
- Name required and under 100 chars
- Description under 500 chars if provided
- At least 1 step required
- All steps must have:
  - Valid type selected
  - Non-empty prompt (max 2000 chars)
- Shows inline error messages for each field
- Scrolls to first error on submit

### Preview Section
Shows when steps exist:
- Template structure summary
- Total step count
- Estimated token usage (rough calculation)
- Step types being used

### State Management
- Form validation on submit
- Error state per field
- Loading state during submission
- Disabled states for invalid forms

## Step Types

| Type | Label | Description |
|------|-------|-------------|
| `exploration` | Exploration | Explore and understand the problem |
| `analysis` | Analysis | Analyze information and patterns |
| `conclusion` | Conclusion | Draw conclusions and synthesize |
| `reflection` | Reflection | Reflect on the process and results |

## Styling

Uses Tailwind CSS with:
- Card layout with border and shadow
- Clean form inputs with focus states
- Color-coded validation (red for errors)
- Responsive spacing
- Smooth transitions on interactive elements
- Icon buttons for step controls

## Example Integration

```svelte
<!-- In a page component -->
<script>
  import { goto } from '$app/navigation';
  import TemplateEditor from '$lib/components/thinking/TemplateEditor.svelte';
  import { api } from '$lib/api/thinking';

  let saving = false;
  let error = null;

  async function handleSave(data) {
    saving = true;
    error = null;

    try {
      const template = await api.createReasoningTemplate(data);
      await goto(`/thinking/templates/${template.id}`);
    } catch (err) {
      error = err.message;
      saving = false;
    }
  }

  function handleCancel() {
    goto('/thinking/templates');
  }
</script>

<div class="container mx-auto px-4 py-8">
  <h1 class="text-2xl font-bold mb-6">Create Template</h1>

  {#if error}
    <div class="mb-4 p-4 bg-red-50 text-red-600 rounded">
      {error}
    </div>
  {/if}

  <TemplateEditor
    template={null}
    onSave={handleSave}
    onCancel={handleCancel}
  />

  {#if saving}
    <div class="fixed inset-0 bg-black/50 flex items-center justify-center">
      <div class="bg-white p-6 rounded-lg">
        Saving template...
      </div>
    </div>
  {/if}
</div>
```

## Accessibility

- Proper label associations (for/id)
- Required field indicators (* suffix)
- Clear error messages
- Keyboard navigation support
- Disabled states for invalid actions
- Title attributes on icon buttons
- Semantic HTML structure

## Browser Compatibility

Works in all modern browsers that support:
- CSS Grid and Flexbox
- ES6+ JavaScript
- Svelte 5 runes
