<!-- 
  FallbackForm.svelte
  Traditional form fallback when agent confidence is low
  Shows dropdown/input for specific field
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import Button from './Button.svelte';

  type FieldType = 'business_name' | 'industry' | 'team_size';

  export let field: FieldType = 'industry';
  export let value: string = "";

  const dispatch = createEventDispatcher();

  const industries = [
    'Marketing / Advertising',
    'Software / Technology',
    'Consulting / Professional Services',
    'E-commerce / Retail',
    'Healthcare',
    'Finance',
    'Education',
    'Other',
  ];

  const teamSizes = [
    '1-5 (Solo or Small Team)',
    '6-10',
    '11-25',
    '26-50',
    '51-100',
    '100+',
  ];

  let selectedValue = value;
  let otherValue = "";
  let showOther = false;

  function handleSelect(event: Event) {
    const target = event.target as HTMLSelectElement;
    selectedValue = target.value;
    showOther = selectedValue === 'Other';
    if (!showOther) {
      otherValue = "";
    }
  }

  function handleOtherInput(event: Event) {
    otherValue = (event.target as HTMLInputElement).value;
  }

  function handleTextInput(event: Event) {
    selectedValue = (event.target as HTMLInputElement).value;
  }

  function handleContinue() {
    const finalValue = showOther ? otherValue : selectedValue;
    if (finalValue.trim()) {
      dispatch('submit', { field, value: finalValue });
    }
  }

  function getOptions() {
    switch (field) {
      case 'industry':
        return industries;
      case 'team_size':
        return teamSizes;
      default:
        return [];
    }
  }

  function getLabel() {
    switch (field) {
      case 'business_name':
        return 'Business Name:';
      case 'industry':
        return 'Industry:';
      case 'team_size':
        return 'Team Size:';
      default:
        return '';
    }
  }

  function getPlaceholder() {
    switch (field) {
      case 'business_name':
        return 'Enter your business name';
      case 'industry':
        return 'Select your industry';
      case 'team_size':
        return 'Select your team size';
      default:
        return '';
    }
  }
</script>

<div class="fallback-form">
  <label class="form-label">{getLabel()}</label>
  
  {#if field === 'business_name'}
    <input
      type="text"
      value={selectedValue}
      on:input={handleTextInput}
      placeholder={getPlaceholder()}
      class="form-input"
    />
  {:else}
    <select
      value={selectedValue}
      on:change={handleSelect}
      class="form-select"
    >
      <option value="" disabled>{getPlaceholder()}</option>
      {#each getOptions() as option}
        <option value={option}>{option}</option>
      {/each}
    </select>

    {#if showOther}
      <input
        type="text"
        value={otherValue}
        on:input={handleOtherInput}
        placeholder="Please specify..."
        class="form-input other-input"
      />
    {/if}
  {/if}

  <Button 
    on:click={handleContinue}
    disabled={(!selectedValue && !otherValue) || (showOther && !otherValue.trim())}
    className="continue-btn"
  >
    Continue
  </Button>
</div>

<style>
  .fallback-form {
    width: 100%;
    max-width: 24rem;
    padding: 1.5rem;
    background-color: var(--card, #ffffff);
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  :global(.dark) .fallback-form {
    background-color: var(--card, #1a1a1a);
    border-color: var(--border, #2a2a2a);
  }

  .form-label {
    font-size: 0.9375rem;
    font-weight: 500;
    color: var(--foreground, #1f2937);
  }

  :global(.dark) .form-label {
    color: var(--foreground, #f9fafb);
  }

  .form-input,
  .form-select {
    width: 100%;
    padding: 0.625rem 0.75rem;
    border: 1px solid var(--border, #e5e7eb);
    border-radius: 0.375rem;
    background-color: var(--background, #ffffff);
    font-size: 0.9375rem;
    color: var(--foreground, #1f2937);
    outline: none;
    font-family: inherit;
    transition: border-color 150ms;
  }

  .form-input::placeholder {
    color: var(--muted-foreground, #6b7280);
  }

  .form-input:focus,
  .form-select:focus {
    border-color: var(--ring, #000000);
  }

  :global(.dark) .form-input,
  :global(.dark) .form-select {
    background-color: var(--background, #0a0a0a);
    border-color: var(--border, #2a2a2a);
    color: var(--foreground, #f9fafb);
  }

  .form-select {
    cursor: pointer;
    appearance: none;
    background-image: url("data:image/svg+xml,%3csvg xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 20 20'%3e%3cpath stroke='%236b7280' stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M6 8l4 4 4-4'/%3e%3c/svg%3e");
    background-position: right 0.5rem center;
    background-repeat: no-repeat;
    background-size: 1.5em 1.5em;
    padding-right: 2.5rem;
  }

  .other-input {
    margin-top: 0.5rem;
  }

  :global(.continue-btn) {
    margin-top: 0.5rem;
  }
</style>
