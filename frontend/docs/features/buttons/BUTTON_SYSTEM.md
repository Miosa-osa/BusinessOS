# Button Component Standardization

## Overview

The BusinessOS frontend uses a unified **btn-pill component system** that provides consistent, accessible, and visually appealing buttons across the entire application. This system was implemented as a major migration effort, updating 100+ files to use the standardized button components.

## Table of Contents

- [Component Architecture](#component-architecture)
- [Available Variants](#available-variants)
- [Size Modifiers](#size-modifiers)
- [Special Modifiers](#special-modifiers)
- [Usage Examples](#usage-examples)
- [Migration from Old Buttons](#migration-from-old-buttons)
- [Where It's Used](#where-its-used)
- [Design System Integration](#design-system-integration)
- [CSS Class Reference](#css-class-reference)
- [Best Practices](#best-practices)

---

## Component Architecture

### PillButton Component

**Location:** `/frontend/src/lib/components/osa/PillButton.svelte`

The `PillButton` component is a Svelte component built using the Svelte 5 `Snippet` API, providing a type-safe, flexible button interface.

### TypeScript Interface

```typescript
type ButtonVariant = 'primary' | 'secondary' | 'ghost';
type ButtonSize = 'sm' | 'md' | 'lg';

interface Props {
  variant?: ButtonVariant;      // Button style variant
  size?: ButtonSize;            // Button size
  disabled?: boolean;           // Disabled state
  loading?: boolean;            // Loading state with spinner
  type?: 'button' | 'submit' | 'reset';
  onclick?: (e: MouseEvent) => void;
  children?: Snippet;           // Svelte 5 snippet for content
  class?: string;               // Additional CSS classes
}
```

### Default Props

- `variant`: `'primary'`
- `size`: `'md'`
- `disabled`: `false`
- `loading`: `false`
- `type`: `'button'`

---

## Available Variants

### 1. Primary (`btn-pill-primary`)

The main call-to-action button with a bold, dark gradient.

**Visual Characteristics:**
- Background: Linear gradient from `#1a1a1a` to `#000000`
- White text
- Prominent shadow and inset highlight
- Subtle lift on hover (2px translateY)
- Subtle shine effect via `::before` pseudo-element

**Usage:**
```svelte
<PillButton variant="primary" onclick={handleSubmit}>
  Submit
</PillButton>
```

**When to Use:**
- Primary actions (submit forms, confirm dialogs)
- Main navigation CTAs
- Important user actions

---

### 2. Secondary (`btn-pill-secondary`)

A lighter button for secondary actions.

**Visual Characteristics:**
- Background: `rgba(255, 255, 255, 0.9)` with backdrop blur
- Border: `1px solid rgba(0, 0, 0, 0.1)`
- Dark text
- Subtle shadow
- Glassmorphism effect

**Usage:**
```svelte
<PillButton variant="secondary" onclick={handleCancel}>
  Cancel
</PillButton>
```

**When to Use:**
- Secondary actions (cancel, skip)
- Alternative options
- Non-critical actions

**Dark Mode:**
- Background: `rgba(58, 58, 60, 0.8)`
- Border: `rgba(255, 255, 255, 0.1)`

---

### 3. Ghost (`btn-pill-ghost`)

A subtle, transparent button for minimal UI impact.

**Visual Characteristics:**
- Background: `rgba(255, 255, 255, 0.1)` with backdrop blur
- Border: `1px solid rgba(0, 0, 0, 0.08)`
- Gray text (`#666666`)
- Minimal shadow

**Usage:**
```svelte
<PillButton variant="ghost" onclick={handleClose}>
  Close
</PillButton>
```

**When to Use:**
- Tertiary actions
- Close buttons
- Actions that shouldn't draw attention

**Dark Mode:**
- Background: `rgba(255, 255, 255, 0.05)`
- Border: `rgba(255, 255, 255, 0.1)`
- Text: `#a1a1a6`

---

### 4. Danger (`btn-pill-danger`)

For destructive actions requiring user caution.

**Visual Characteristics:**
- Background: Linear gradient from `#dc2626` to `#b91c1c`
- White text
- Red-tinted shadow
- Hover: Brighter red gradient

**Usage:**
```svelte
<button class="btn-pill btn-pill-danger" onclick={handleDelete}>
  Delete Account
</button>
```

**When to Use:**
- Delete actions
- Destructive operations
- Actions that require caution

---

### 5. Success (`btn-pill-success`)

For positive, confirming actions.

**Visual Characteristics:**
- Background: Linear gradient from `#16a34a` to `#15803d`
- White text
- Green-tinted shadow
- Hover: Brighter green gradient

**Usage:**
```svelte
<button class="btn-pill btn-pill-success" onclick={handleConfirm}>
  Confirm
</button>
```

**When to Use:**
- Success confirmations
- Positive actions
- "Continue" or "Proceed" buttons

---

### 6. Additional Variants (CSS-only)

These variants are available via CSS classes only (not through the PillButton component props):

#### Outline (`btn-pill-outline`)
- Transparent background
- 2px border
- Fills with color on hover

```svelte
<button class="btn-pill btn-pill-outline">
  Outline Button
</button>
```

#### Soft (`btn-pill-soft`)
- Very subtle background tint
- Minimal shadow
- Gentle hover effect

```svelte
<button class="btn-pill btn-pill-soft">
  Soft Button
</button>
```

#### Warning (`btn-pill-warning`)
- Yellow/amber gradient
- Dark brown text
- Warning-tinted shadow

```svelte
<button class="btn-pill btn-pill-warning">
  Warning
</button>
```

#### Link (`btn-pill-link`)
- Transparent background
- Underlined text
- No border or shadow

```svelte
<button class="btn-pill btn-pill-link">
  Link Button
</button>
```

---

## Size Modifiers

### Extra Small (`btn-pill-xs`)
```svelte
<button class="btn-pill btn-pill-primary btn-pill-xs">
  Extra Small
</button>
```
- Padding: `0.375rem 0.875rem`
- Font size: `var(--text-xs)`
- Gap: `0.25rem`

### Small (`btn-pill-sm` or `size="sm"`)
```svelte
<PillButton variant="primary" size="sm">
  Small Button
</PillButton>
```
- Padding: `0.5rem 1.25rem`
- Font size: `var(--text-sm)`

### Medium (default)
```svelte
<PillButton variant="primary">
  Medium Button
</PillButton>
```
- Padding: `0.875rem 1.75rem`
- Font size: `var(--text-base)`

### Large (`btn-pill-lg` or `size="lg"`)
```svelte
<PillButton variant="primary" size="lg">
  Large Button
</PillButton>
```
- Padding: `1rem 2rem`
- Font size: `var(--text-lg)`

### Extra Large (`btn-pill-xl`)
```svelte
<button class="btn-pill btn-pill-primary btn-pill-xl">
  Extra Large
</button>
```
- Padding: `1.125rem 2.5rem`
- Font size: `var(--text-xl)`

---

## Special Modifiers

### Icon-Only Buttons (`btn-pill-icon`)

For buttons containing only an icon (no text).

```svelte
<button class="btn-pill btn-pill-primary btn-pill-icon">
  <svg>...</svg>
</button>
```

**Sizing:**
- Default: `2.5rem × 2.5rem` (padding: `0.625rem`)
- Small: `2rem × 2rem` (padding: `0.5rem`)
- Large: `3rem × 3rem` (padding: `0.75rem`)

### Full Width (`btn-pill-block`)

Makes button take up full width of container.

```svelte
<button class="btn-pill btn-pill-primary btn-pill-block">
  Full Width Button
</button>
```

- `width: 100%`
- `display: flex`

### Loading State

The `PillButton` component supports a `loading` prop that displays a spinner.

```svelte
<PillButton variant="primary" loading={isLoading} disabled={isLoading}>
  {isLoading ? 'Processing...' : 'Submit'}
</PillButton>
```

**Visual Behavior:**
- Displays animated spinner
- Automatically disables button
- Gap between spinner and text

### Button Groups (`btn-pill-group`)

For grouping multiple buttons together.

```svelte
<div class="btn-pill-group">
  <button class="btn-pill btn-pill-ghost">Option 1</button>
  <button class="btn-pill btn-pill-primary">Option 2</button>
  <button class="btn-pill btn-pill-ghost">Option 3</button>
</div>
```

**Visual Characteristics:**
- Wrapped in subtle background container
- 2px padding and gap between buttons
- Buttons have no shadow or border within group
- Active button stands out with primary style

---

## Usage Examples

### Basic Usage (Svelte Component)

```svelte
<script>
  import { PillButton } from '$lib/components/osa';

  function handleClick() {
    console.log('Button clicked!');
  }
</script>

<PillButton variant="primary" onclick={handleClick}>
  Click Me
</PillButton>
```

### Form Submission

```svelte
<script>
  import { PillButton } from '$lib/components/osa';

  let loading = false;

  async function handleSubmit() {
    loading = true;
    try {
      await submitForm();
    } finally {
      loading = false;
    }
  }
</script>

<form onsubmit|preventDefault={handleSubmit}>
  <PillButton type="submit" variant="primary" loading={loading}>
    {loading ? 'Submitting...' : 'Submit Form'}
  </PillButton>
</form>
```

### Action Buttons with Multiple Variants

```svelte
<script>
  import { PillButton } from '$lib/components/osa';
</script>

<div class="flex gap-4">
  <PillButton variant="primary" onclick={handleSave}>
    Save Changes
  </PillButton>

  <PillButton variant="secondary" onclick={handleCancel}>
    Cancel
  </PillButton>

  <PillButton variant="ghost" onclick={handleReset}>
    Reset
  </PillButton>
</div>
```

### Full Width CTA

```svelte
<script>
  import { PillButton } from '$lib/components/osa';
</script>

<div class="w-full max-w-md">
  <PillButton variant="primary" size="lg" class="btn-pill-block" onclick={handleEnter}>
    Enter Your OS
  </PillButton>
</div>
```

### Direct CSS Class Usage (Non-Component)

You can also use the btn-pill classes directly on any button element:

```svelte
<button class="btn-pill btn-pill-primary" onclick={handleClick}>
  Direct CSS Usage
</button>

<button class="btn-pill btn-pill-danger btn-pill-sm" onclick={handleDelete}>
  Delete
</button>

<button class="btn-pill btn-pill-ghost btn-pill-block" onclick={handleSkip}>
  Skip
</button>
```

---

## Migration from Old Buttons

### What Changed

The button standardization migration replaced various custom button implementations with the unified `btn-pill` system.

**Before:**
```svelte
<!-- Old custom button classes -->
<button class="w-full bg-blue-600 hover:bg-blue-700 rounded-xl px-8 py-3 text-white">
  Submit
</button>

<button class="bg-red-600 hover:bg-red-700 rounded-lg px-6 py-2.5 text-white">
  Delete
</button>

<button class="bg-gray-900 hover:bg-gray-800 rounded-lg px-4 py-2 text-white">
  Action
</button>
```

**After:**
```svelte
<!-- New unified btn-pill system -->
<button class="btn-pill btn-pill-primary btn-pill-block">
  Submit
</button>

<button class="btn-pill btn-pill-danger">
  Delete
</button>

<button class="btn-pill btn-pill-primary">
  Action
</button>
```

### Migration Script

**Location:** `/frontend/update-buttons.sh`

This automated script migrated 100+ files from old button patterns to the new btn-pill system.

**What it does:**
1. Finds all `.svelte` files with old button patterns
2. Creates backups (`.svelte.bak`)
3. Replaces old color/size classes with btn-pill variants
4. Removes redundant padding, rounding, and spacing classes
5. Logs changed files

**Pattern Replacements:**

| Old Pattern | New Pattern |
|-------------|-------------|
| `bg-blue-600 hover:bg-blue-700` | `btn-pill btn-pill-primary` |
| `bg-red-600 hover:bg-red-700` | `btn-pill btn-pill-danger` |
| `bg-green-600 hover:bg-green-700` | `btn-pill btn-pill-success` |
| `bg-purple-600 hover:bg-purple-700` | `btn-pill btn-pill-primary` |
| `bg-gray-900 hover:bg-gray-800` | `btn-pill btn-pill-primary` |
| `bg-gray-800 hover:bg-gray-700` | `btn-pill btn-pill-secondary` |

**Cleanup:**
- Removes `px-*` and `py-*` classes (btn-pill has its own padding)
- Removes `rounded-*` classes (btn-pill has its own border radius)

**Running the script:**

```bash
cd frontend
chmod +x update-buttons.sh
./update-buttons.sh
```

**Restoring from backups:**

```bash
# Restore all backups
find src -name '*.svelte.bak' -exec bash -c 'mv "$0" "${0%.bak}"' {} \;

# Delete all backups
find src -name '*.svelte.bak' -delete
```

---

## Where It's Used

The btn-pill system is used extensively across the BusinessOS frontend:

### Onboarding Flow (100% coverage)
- `/routes/onboarding/+page.svelte` - Welcome screen
- `/routes/onboarding/signin/+page.svelte` - Sign in
- `/routes/onboarding/username/+page.svelte` - Username setup
- `/routes/onboarding/gmail/+page.svelte` - Gmail integration
- `/routes/onboarding/analyzing/+page.svelte` - Analysis screens (1, 2, 3)
- `/routes/onboarding/meet-osa/+page.svelte` - Meet OSA
- `/routes/onboarding/starter-apps/+page.svelte` - Starter apps
- `/routes/onboarding/ready/+page.svelte` - Final ready screen

### Core Application Components
- **Dock**: `/lib/components/desktop/Dock.svelte` (15 occurrences)
- **Chat Interface**: `/routes/(app)/chat/+page.svelte` (35 occurrences)
- **Settings**: `/routes/(app)/settings/+page.svelte` (23 occurrences)
- **Project Management**: `/routes/(app)/projects/` (40 occurrences)
- **Tables**: `/routes/(app)/tables/` (18 occurrences)

### Modal Components
- `AddClientModal.svelte` (9 occurrences)
- `NewTaskModal.svelte` (2 occurrences)
- `DocumentUploadModal.svelte` (2 occurrences)
- `InviteMemberModal.svelte` (3 occurrences)
- `FilterModal.svelte` (2 occurrences)

### Authentication
- `/routes/login/+page.svelte`
- `/routes/register/+page.svelte`
- `/routes/forgot-password/+page.svelte`
- `/routes/reset-password/+page.svelte`

### Total Coverage
- **110 files** using btn-pill classes
- **585+ button instances** migrated
- **100% of onboarding flow** standardized

---

## Design System Integration

### OSA Component Library

The `PillButton` component is part of the OSA component library:

**Export Location:** `/lib/components/osa/index.ts`

```typescript
export { default as PillButton } from './PillButton.svelte';
export { default as BuildProgress } from './BuildProgress.svelte';
// ... other OSA components
```

### Tailwind Integration

The btn-pill system integrates seamlessly with Tailwind CSS:

```svelte
<PillButton variant="primary" class="mt-4">
  Custom Margin
</PillButton>

<div class="flex gap-4 items-center">
  <PillButton variant="secondary">Button 1</PillButton>
  <PillButton variant="primary">Button 2</PillButton>
</div>
```

### CSS Variables

The btn-pill system uses CSS custom properties defined in `/app.css`:

```css
:root {
  /* Text Colors */
  --text-primary: #1A1A1A;
  --text-secondary: #666666;

  /* Shadow System */
  --shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);

  /* Border Radius */
  --radius-full: 9999px;

  /* Font Sizes */
  --text-xs: 0.75rem;
  --text-sm: 0.875rem;
  --text-base: 1rem;
  --text-lg: 1.125rem;
  --text-xl: 1.25rem;
}
```

### Dark Mode Support

All btn-pill variants support dark mode via the `.dark` class:

```css
.dark .btn-pill-primary {
  background: linear-gradient(180deg, #3a3a3c 0%, #1c1c1e 100%);
  border-color: rgba(255, 255, 255, 0.15);
}

.dark .btn-pill-secondary {
  background: rgba(58, 58, 60, 0.8);
  border-color: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}
```

### iOS-Inspired Design

The btn-pill system follows iOS design principles:

- **Pill shape**: Full border radius (`border-radius: 9999px`)
- **Glassmorphism**: Backdrop blur effects on secondary/ghost variants
- **Subtle animations**: Smooth transforms and transitions
- **Haptic feedback**: Lift on hover, press on active
- **Accessibility**: Proper disabled states, loading indicators

---

## CSS Class Reference

### Complete Class List

```css
/* Base */
.btn-pill                    /* Base button styles */

/* Variants */
.btn-pill-primary            /* Main CTA (dark gradient) */
.btn-pill-secondary          /* Secondary actions (light, glass) */
.btn-pill-ghost              /* Subtle, transparent */
.btn-pill-danger             /* Destructive actions (red) */
.btn-pill-success            /* Positive actions (green) */
.btn-pill-warning            /* Warning actions (yellow) */
.btn-pill-outline            /* Bordered, transparent */
.btn-pill-soft               /* Very subtle background */
.btn-pill-link               /* Link-style, underlined */

/* Sizes */
.btn-pill-xs                 /* Extra small */
.btn-pill-sm                 /* Small */
/* (default)                    Medium (no class needed) */
.btn-pill-lg                 /* Large */
.btn-pill-xl                 /* Extra large */

/* Modifiers */
.btn-pill-icon               /* Icon-only button */
.btn-pill-block              /* Full width */
.btn-pill-loading            /* Loading state (with spinner) */

/* Grouping */
.btn-pill-group              /* Container for button groups */
```

### Combining Classes

Classes can be combined for complex button styles:

```svelte
<!-- Primary large full-width button -->
<button class="btn-pill btn-pill-primary btn-pill-lg btn-pill-block">
  Large Full Width Primary
</button>

<!-- Small danger icon button -->
<button class="btn-pill btn-pill-danger btn-pill-sm btn-pill-icon">
  <TrashIcon />
</button>

<!-- Ghost extra small button -->
<button class="btn-pill btn-pill-ghost btn-pill-xs">
  Tiny Ghost
</button>
```

---

## Best Practices

### 1. Use the PillButton Component When Possible

**Preferred:**
```svelte
<PillButton variant="primary" onclick={handleClick}>
  Submit
</PillButton>
```

**Also Valid (for more control):**
```svelte
<button class="btn-pill btn-pill-primary" onclick={handleClick}>
  Submit
</button>
```

### 2. Choose the Right Variant

- **Primary**: Main action, most important button on screen (1 per view)
- **Secondary**: Alternative actions, cancel buttons
- **Ghost**: Tertiary actions, close buttons, minimal impact
- **Danger**: Delete, destructive actions (use sparingly)
- **Success**: Confirmations, positive outcomes

### 3. Size Appropriately

- **xs/sm**: Compact UIs, toolbar buttons, inline actions
- **md**: Default for most use cases
- **lg/xl**: Hero CTAs, final onboarding steps, emphasis

### 4. Accessibility

Always provide meaningful text or aria-labels:

```svelte
<!-- Good: Descriptive text -->
<PillButton variant="primary" onclick={handleSave}>
  Save Changes
</PillButton>

<!-- Good: Icon with aria-label -->
<button class="btn-pill btn-pill-ghost btn-pill-icon" aria-label="Close dialog">
  <XIcon />
</button>

<!-- Bad: No context -->
<button class="btn-pill btn-pill-primary">
  <XIcon />
</button>
```

### 5. Loading States

Use the `loading` prop for async actions:

```svelte
<script>
  let saving = false;

  async function handleSave() {
    saving = true;
    try {
      await saveData();
    } finally {
      saving = false;
    }
  }
</script>

<PillButton variant="primary" loading={saving} disabled={saving}>
  {saving ? 'Saving...' : 'Save'}
</PillButton>
```

### 6. Grouping Related Actions

Use button groups for toggle-like or related actions:

```svelte
<div class="btn-pill-group">
  <button class="btn-pill btn-pill-ghost" class:btn-pill-primary={view === 'list'}>
    List
  </button>
  <button class="btn-pill btn-pill-ghost" class:btn-pill-primary={view === 'grid'}>
    Grid
  </button>
  <button class="btn-pill btn-pill-ghost" class:btn-pill-primary={view === 'gallery'}>
    Gallery
  </button>
</div>
```

### 7. Consistent Spacing

Use Tailwind's gap utilities for button layouts:

```svelte
<div class="flex gap-3">
  <PillButton variant="secondary">Cancel</PillButton>
  <PillButton variant="primary">Confirm</PillButton>
</div>
```

### 8. Form Buttons

Use proper `type` attributes:

```svelte
<form onsubmit|preventDefault={handleSubmit}>
  <!-- Submit button (default type="submit" in PillButton) -->
  <PillButton type="submit" variant="primary">
    Submit Form
  </PillButton>

  <!-- Reset button -->
  <PillButton type="reset" variant="secondary">
    Reset
  </PillButton>

  <!-- Regular action button -->
  <PillButton type="button" variant="ghost" onclick={handleCancel}>
    Cancel
  </PillButton>
</form>
```

### 9. Avoid Over-Using Primary

Only use `btn-pill-primary` for the most important action on a page:

```svelte
<!-- Good: One clear primary action -->
<div class="actions">
  <PillButton variant="ghost">Skip</PillButton>
  <PillButton variant="primary">Continue</PillButton>
</div>

<!-- Bad: Multiple primaries compete for attention -->
<div class="actions">
  <PillButton variant="primary">Action 1</PillButton>
  <PillButton variant="primary">Action 2</PillButton>
  <PillButton variant="primary">Action 3</PillButton>
</div>
```

### 10. Mobile Considerations

Use `btn-pill-block` for mobile-friendly full-width buttons:

```svelte
<div class="w-full sm:w-auto">
  <PillButton variant="primary" class="btn-pill-block sm:inline-flex">
    Responsive Button
  </PillButton>
</div>
```

---

## Common Patterns

### Confirm/Cancel Pair

```svelte
<div class="flex justify-end gap-3">
  <PillButton variant="secondary" onclick={handleCancel}>
    Cancel
  </PillButton>
  <PillButton variant="primary" onclick={handleConfirm}>
    Confirm
  </PillButton>
</div>
```

### Delete Confirmation

```svelte
<div class="flex flex-col gap-4">
  <p class="text-sm text-gray-600">
    Are you sure you want to delete this item? This action cannot be undone.
  </p>
  <div class="flex justify-end gap-3">
    <PillButton variant="ghost" onclick={handleCancel}>
      Cancel
    </PillButton>
    <PillButton variant="danger" onclick={handleDelete}>
      Delete Permanently
    </PillButton>
  </div>
</div>
```

### Icon + Text

```svelte
<PillButton variant="primary">
  <PlusIcon class="w-4 h-4" />
  Add New Item
</PillButton>
```

### Loading with State

```svelte
<script>
  let state: 'idle' | 'loading' | 'success' | 'error' = 'idle';

  async function handleAction() {
    state = 'loading';
    try {
      await performAction();
      state = 'success';
    } catch {
      state = 'error';
    }
  }
</script>

<PillButton
  variant={state === 'success' ? 'success' : 'primary'}
  loading={state === 'loading'}
  disabled={state === 'loading'}
  onclick={handleAction}
>
  {#if state === 'loading'}
    Processing...
  {:else if state === 'success'}
    Success!
  {:else if state === 'error'}
    Try Again
  {:else}
    Submit
  {/if}
</PillButton>
```

---

## Future Enhancements

Potential improvements to the btn-pill system:

1. **Icon Positioning**: Built-in support for left/right icon placement
2. **Badge/Notification**: Support for notification badges on buttons
3. **Tooltip Integration**: Built-in tooltip support
4. **Animation Variants**: Different hover/press animations
5. **Theme Variants**: Support for custom color themes
6. **Compound Components**: DropdownButton, SplitButton, etc.

---

## Troubleshooting

### Button Not Styling Correctly

**Issue**: Button doesn't have expected styles.

**Solution**: Ensure `btn-pill` base class is present:

```svelte
<!-- Wrong -->
<button class="btn-pill-primary">Button</button>

<!-- Correct -->
<button class="btn-pill btn-pill-primary">Button</button>
```

### Loading Spinner Not Showing

**Issue**: Loading prop doesn't show spinner.

**Solution**: Ensure you're using the `PillButton` component, not raw button:

```svelte
<!-- Won't work -->
<button class="btn-pill btn-pill-primary" loading={true}>Button</button>

<!-- Will work -->
<PillButton variant="primary" loading={true}>Button</PillButton>
```

### Dark Mode Not Working

**Issue**: Button doesn't change in dark mode.

**Solution**: Ensure parent element has `.dark` class:

```svelte
<div class="dark">
  <PillButton variant="primary">Dark Mode Button</PillButton>
</div>
```

### Button Group Styling Issues

**Issue**: Buttons in group look wrong.

**Solution**: Ensure wrapper has `btn-pill-group` class:

```svelte
<!-- Correct -->
<div class="btn-pill-group">
  <button class="btn-pill btn-pill-ghost">Option 1</button>
  <button class="btn-pill btn-pill-primary">Option 2</button>
</div>
```

---

## Resources

- **Component Source**: `/frontend/src/lib/components/osa/PillButton.svelte`
- **CSS Definitions**: `/frontend/src/app.css` (lines 947-1447)
- **Migration Script**: `/frontend/update-buttons.sh`
- **Example Usage**: `/frontend/src/routes/onboarding/ready/+page.svelte`

---

**Last Updated:** January 2026
**Version:** 1.0.0
**Maintainer:** BusinessOS Frontend Team
