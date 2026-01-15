<!-- 
  OnboardingPage.svelte
  Main onboarding flow page with all steps
  Converted from Next.js app/onboarding/page.tsx
-->
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import FloatingChatScreen from './FloatingChatScreen.svelte';
  import SequentialTypewriter from './SequentialTypewriter.svelte';
  import ChatInput from './ChatInput.svelte';
  import Button from './Button.svelte';

  const dispatch = createEventDispatcher();

  type OnboardingStep = 'intro' | 'role' | 'problem' | 'deeper' | 'purpose' | 'processing';

  interface UserData {
    name: string;
    role: string;
    problem: string;
    deeper: string;
    purpose: 'internal' | 'external' | 'both' | '';
    processing: string;
  }

  const roles = ['Founder', 'Vibe Coder', 'Creator', 'Agency', 'Business', 'Other'];
  const purposes = [
    { value: 'internal', label: 'For Your Team' },
    { value: 'external', label: 'For Customers' },
    { value: 'both', label: 'Both' },
  ];

  let step: OnboardingStep = 'intro';
  let showInput = false;
  let currentInput = '';
  let userData: UserData = {
    name: '',
    role: '',
    problem: '',
    deeper: '',
    purpose: '',
    processing: '',
  };

  function handleSkip() {
    dispatch('skip');
  }

  function handleBack() {
    const stepOrder: OnboardingStep[] = ['intro', 'role', 'problem', 'deeper', 'purpose', 'processing'];
    const currentIndex = stepOrder.indexOf(step);
    if (currentIndex > 0) {
      step = stepOrder[currentIndex - 1];
      showInput = false;
      currentInput = '';
    } else {
      dispatch('back');
    }
  }

  function handleSubmit() {
    if (!currentInput.trim()) return;

    switch (step) {
      case 'intro':
        userData = { ...userData, name: currentInput };
        step = 'role';
        showInput = false;
        currentInput = '';
        break;
      case 'problem':
        userData = { ...userData, problem: currentInput };
        step = 'deeper';
        showInput = false;
        currentInput = '';
        break;
      case 'deeper':
        userData = { ...userData, deeper: currentInput };
        step = 'purpose';
        showInput = false;
        currentInput = '';
        break;
      case 'processing':
        userData = { ...userData, processing: currentInput };
        dispatch('complete', { userData });
        break;
    }
  }

  function handleRoleSelect(role: string) {
    userData = { ...userData, role };
    step = 'problem';
  }

  function handlePurposeSelect(purpose: 'internal' | 'external' | 'both') {
    userData = { ...userData, purpose };
    step = 'processing';
    showInput = false;
  }

  function handleInputChange(event: CustomEvent<{ value: string }>) {
    currentInput = event.detail.value;
  }

  function handleTypewriterComplete() {
    showInput = true;
  }

  function getProcessingLines(): string[] {
    switch (userData.purpose) {
      case 'internal':
        return [
          'Got it - building something for YOUR operations.',
          "What's frustrating you most about your current setup?",
        ];
      case 'external':
        return [
          'Perfect - creating something for YOUR customers to use.',
          'What would success look like in 90 days?',
        ];
      case 'both':
        return [
          'Interesting - you need both internal tools AND customer-facing features.',
          'What would success look like in 90 days?',
        ];
      default:
        return [];
    }
  }

  function getProcessingPlaceholder(): string {
    return userData.purpose === 'internal' ? 'Our biggest frustration is...' : 'Success means...';
  }
</script>

<div class="onboarding-page">
  <FloatingChatScreen on:back={handleBack} showBack={step !== 'intro'}>
    
    {#if step === 'intro'}
      <div class="step-content">
        <SequentialTypewriter
          lines={["I'm MIOSA, your OS Agent.", "Let's turn your idea into reality.", "What's your name?"]}
          on:complete={handleTypewriterComplete}
          className="typewriter-container"
        />
        {#if showInput}
          <div class="input-wrapper animate-fade-in">
            <ChatInput
              bind:value={currentInput}
              on:change={handleInputChange}
              on:submit={handleSubmit}
              placeholder="Type here..."
            />
          </div>
        {/if}
      </div>
    {/if}

    {#if step === 'role'}
      <div class="step-content">
        <SequentialTypewriter
          lines={[`Nice to meet you, ${userData.name}.`, 'Which best describes what you do?']}
          on:complete={handleTypewriterComplete}
          className="typewriter-container"
        />
        {#if showInput}
          <div class="button-grid animate-fade-in">
            {#each roles as role}
              <Button
                variant="outline"
                className="role-button"
                on:click={() => handleRoleSelect(role)}
              >
                {role}
              </Button>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    {#if step === 'problem'}
      <div class="step-content">
        <SequentialTypewriter
          lines={[
            `Alright ${userData.name}, let's dive in.`,
            "Is there a specific problem you're looking to solve or something you're thinking of building?",
          ]}
          on:complete={handleTypewriterComplete}
          className="typewriter-container"
        />
        {#if showInput}
          <div class="input-wrapper animate-fade-in">
            <ChatInput
              bind:value={currentInput}
              on:change={handleInputChange}
              on:submit={handleSubmit}
              placeholder="I want to build... / I have a problem with..."
              showMic={true}
            />
          </div>
        {/if}
      </div>
    {/if}

    {#if step === 'deeper'}
      <div class="step-content">
        <SequentialTypewriter
          lines={[
            `I see what you're dealing with, ${userData.name}.`,
            "Tell me more - who needs this solution? Is this something you've noticed others struggling with too?",
          ]}
          on:complete={handleTypewriterComplete}
          className="typewriter-container"
        />
        {#if showInput}
          <div class="input-wrapper animate-fade-in">
            <ChatInput
              bind:value={currentInput}
              on:change={handleInputChange}
              on:submit={handleSubmit}
              placeholder="My team needs... / I've noticed that..."
            />
          </div>
        {/if}
      </div>
    {/if}

    {#if step === 'purpose'}
      <div class="step-content">
        <SequentialTypewriter
          lines={[
            `Let me make sure I understand, ${userData.name}.`,
            'Is this for YOUR team to use internally, or for YOUR customers to use?',
          ]}
          on:complete={handleTypewriterComplete}
          className="typewriter-container"
        />
        {#if showInput}
          <div class="purpose-buttons animate-fade-in">
            {#each purposes as purpose}
              <Button
                variant="outline"
                className="purpose-button"
                on:click={() => handlePurposeSelect(purpose.value as 'internal' | 'external' | 'both')}
              >
                {purpose.label}
              </Button>
            {/each}
          </div>
        {/if}
      </div>
    {/if}

    {#if step === 'processing'}
      <div class="step-content">
        <SequentialTypewriter 
          lines={getProcessingLines()} 
          on:complete={handleTypewriterComplete} 
          className="typewriter-container" 
        />
        {#if showInput}
          <div class="input-wrapper animate-fade-in">
            <ChatInput
              bind:value={currentInput}
              on:change={handleInputChange}
              on:submit={handleSubmit}
              placeholder={getProcessingPlaceholder()}
            />
          </div>
        {/if}
      </div>
    {/if}

  </FloatingChatScreen>

  <Button
    variant="ghost"
    className="skip-button"
    on:click={handleSkip}
  >
    Skip for now →
  </Button>
</div>

<style>
  .onboarding-page {
    position: relative;
    min-height: 100vh;
  }

  .step-content {
    width: 100%;
    text-align: center;
  }

  :global(.typewriter-container) {
    margin-bottom: 2rem;
  }

  .input-wrapper {
    width: 100%;
  }

  .button-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.75rem;
    max-width: 28rem;
    margin: 0 auto;
  }

  :global(.role-button) {
    height: 3rem !important;
    background-color: transparent !important;
  }

  .purpose-buttons {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    max-width: 28rem;
    margin: 0 auto;
  }

  :global(.purpose-button) {
    height: 3rem !important;
    justify-content: flex-start !important;
    background-color: transparent !important;
  }

  :global(.skip-button) {
    position: fixed !important;
    bottom: 1.5rem;
    right: 1.5rem;
    color: var(--muted-foreground) !important;
  }

  :global(.skip-button:hover) {
    color: var(--foreground) !important;
  }

  /* Animation */
  .animate-fade-in {
    animation: fade-in 300ms ease-out forwards;
  }

  @keyframes fade-in {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
</style>
