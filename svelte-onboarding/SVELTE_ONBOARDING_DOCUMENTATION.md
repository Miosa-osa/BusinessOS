# BusinessOS Onboarding - Svelte Components Documentation

**Version:** 1.0  
**Last Updated:** January 10, 2026  
**Converted From:** Next.js/React to Svelte  

---

## Table of Contents

1. [Overview](#overview)
2. [File Structure](#file-structure)
3. [Installation & Setup](#installation--setup)
4. [Design System](#design-system)
5. [Components Reference](#components-reference)
6. [Icons](#icons)
7. [Animations & Effects](#animations--effects)
8. [Usage Examples](#usage-examples)
9. [Theming](#theming)

---

## Overview

This package contains all Svelte components converted from the original Next.js/React codebase for the BusinessOS conversational onboarding experience. The onboarding uses an **agent-first, conversational approach** powered by AI.

### Key Features

- **Purple Orb Icon** - Animated glowing orb with pulse and shimmer effects
- **Typewriter Effect** - Character-by-character text animation with blinking cursor
- **Sequential Typewriter** - Multiple lines typed one after another
- **Chat Interface** - Full conversational UI with message bubbles
- **Theme Support** - Light and dark mode with smooth transitions
- **Responsive Design** - Mobile-first approach

---

## File Structure

```
svelte-onboarding/
├── components/
│   ├── icons/
│   │   ├── ArrowLeftIcon.svelte      # Back navigation arrow
│   │   ├── CheckIcon.svelte          # Checkmark for completion
│   │   ├── MicIcon.svelte            # Microphone for voice input
│   │   ├── MoonIcon.svelte           # Dark theme icon
│   │   ├── SendIcon.svelte           # Send message arrow
│   │   └── SunIcon.svelte            # Light theme icon
│   │
│   ├── ActionButtons.svelte          # Quick action button group
│   ├── Button.svelte                 # Base button component
│   ├── ChatInput.svelte              # Chat-style input with send
│   ├── CompletionScreen.svelte       # Onboarding completion summary
│   ├── EmailInviteInput.svelte       # Team member email invites
│   ├── FallbackForm.svelte           # Traditional form fallback
│   ├── FileUpload.svelte             # Drag & drop file upload
│   ├── FloatingChatScreen.svelte     # Main layout container
│   ├── Input.svelte                  # Base input component
│   ├── IntegrationCard.svelte        # OAuth integration card
│   ├── MemberWelcome.svelte          # Team member welcome screen
│   ├── MessageBubble.svelte          # Chat message bubble
│   ├── OnboardingPage.svelte         # Main onboarding flow
│   ├── PurpleOrb.svelte              # Animated purple orb icon
│   ├── SequentialTypewriter.svelte   # Multi-line typewriter
│   ├── ThemeToggle.svelte            # Light/dark theme toggle
│   ├── ToolPicker.svelte             # Multi-select tool grid
│   ├── TypewriterText.svelte         # Single line typewriter
│   ├── TypingIndicator.svelte        # AI thinking indicator
│   └── index.ts                      # Component exports
│
└── styles/
    └── globals.css                   # Global styles & CSS variables
```

---

## Installation & Setup

### 1. Copy Files

Copy the `svelte-onboarding` folder to your Svelte/SvelteKit project.

### 2. Import Global Styles

In your main layout or app file:

```svelte
<script>
  import '../svelte-onboarding/styles/globals.css';
</script>
```

Or in your CSS:

```css
@import './svelte-onboarding/styles/globals.css';
```

### 3. Import Components

```svelte
<script>
  import { 
    PurpleOrb, 
    OnboardingPage, 
    FloatingChatScreen,
    SequentialTypewriter 
  } from './svelte-onboarding/components';
</script>
```

---

## Design System

### Color Palette

#### Light Mode
| Variable | Value | Usage |
|----------|-------|-------|
| `--background` | `#ffffff` | Page background |
| `--foreground` | `#1f2937` | Primary text |
| `--primary` | `#000000` | Buttons, accents |
| `--primary-foreground` | `#ffffff` | Text on primary |
| `--muted` | `#f9fafb` | Subtle backgrounds |
| `--muted-foreground` | `#6b7280` | Secondary text |
| `--border` | `#e5e7eb` | Borders, dividers |
| `--accent` | `#f3f4f6` | Hover states |
| `--success` | `#10b981` | Success states |
| `--error` | `#ef4444` | Error states |

#### Dark Mode
| Variable | Value | Usage |
|----------|-------|-------|
| `--background` | `#0a0a0a` | Page background |
| `--foreground` | `#f9fafb` | Primary text |
| `--primary` | `#ffffff` | Buttons, accents |
| `--primary-foreground` | `#000000` | Text on primary |
| `--muted` | `#1a1a1a` | Subtle backgrounds |
| `--muted-foreground` | `#9ca3af` | Secondary text |
| `--border` | `#2a2a2a` | Borders, dividers |
| `--accent` | `#2a2a2a` | Hover states |

#### Purple Orb Colors
| Variable | Value | Description |
|----------|-------|-------------|
| `--orb-gradient-1` | `#e0e7ff` | Lightest purple |
| `--orb-gradient-2` | `#a5b4fc` | Light purple |
| `--orb-gradient-3` | `#818cf8` | Medium purple |
| `--orb-gradient-4` | `#6366f1` | Dark purple |
| `--orb-gradient-5` | `#4f46e5` | Darkest purple |
| `--orb-glow` | `rgba(99, 102, 241, 0.4)` | Outer glow |
| `--orb-inner-glow` | `rgba(255, 255, 255, 0.3)` | Inner highlight |

### Typography

```css
--font-sans: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
--font-mono: "JetBrains Mono", "Fira Code", Consolas, monospace;
```

### Border Radius

```css
--radius-sm: 8px;
--radius-md: 8px;
--radius-lg: 12px;
--radius-xl: 16px;
--radius-full: 9999px;  /* Pill shape */
```

---

## Components Reference

### PurpleOrb

The animated purple orb icon - the main visual element of the onboarding.

```svelte
<script>
  import { PurpleOrb } from './components';
</script>

<PurpleOrb size={80} className="my-orb" />
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `size` | `number` | `80` | Width and height in pixels |
| `className` | `string` | `""` | Additional CSS classes |

#### CSS Details

```css
/* Gradient Background */
background: linear-gradient(
  135deg,
  #e0e7ff 0%,    /* Light indigo */
  #a5b4fc 25%,   /* Indigo 300 */
  #818cf8 50%,   /* Indigo 400 */
  #6366f1 75%,   /* Indigo 500 */
  #4f46e5 100%   /* Indigo 600 */
);

/* Glow Effect */
box-shadow: 
  0 0 40px rgba(99, 102, 241, 0.4),      /* Outer glow */
  inset 0 0 20px rgba(255, 255, 255, 0.3); /* Inner highlight */
```

---

### TypewriterText

Single-line typewriter text effect with optional blinking cursor.

```svelte
<script>
  import { TypewriterText } from './components';

  function handleComplete() {
    console.log('Typing finished!');
  }
</script>

<TypewriterText 
  text="Hello, I'm your AI assistant." 
  speed={50}
  showCursor={true}
  on:complete={handleComplete}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `text` | `string` | `""` | Text to type out |
| `speed` | `number` | `50` | Milliseconds between characters |
| `showCursor` | `boolean` | `false` | Show blinking cursor |
| `className` | `string` | `""` | Additional CSS classes |

#### Events
| Event | Detail | Description |
|-------|--------|-------------|
| `complete` | `void` | Fired when typing finishes |

---

### SequentialTypewriter

Types multiple lines one after another with pauses between.

```svelte
<script>
  import { SequentialTypewriter } from './components';

  const lines = [
    "I'm MIOSA, your OS Agent.",
    "Let's turn your idea into reality.",
    "What's your name?"
  ];

  function handleComplete() {
    showInput = true;
  }
</script>

<SequentialTypewriter 
  {lines}
  speed={50}
  pauseBetween={500}
  on:complete={handleComplete}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `lines` | `string[]` | `[]` | Array of lines to type |
| `speed` | `number` | `50` | Typing speed (ms/char) |
| `pauseBetween` | `number` | `500` | Pause between lines (ms) |
| `className` | `string` | `""` | Additional CSS classes |

#### Events
| Event | Detail | Description |
|-------|--------|-------------|
| `complete` | `void` | Fired when all lines finish |

---

### FloatingChatScreen

Main layout container with centered content, orb, and navigation.

```svelte
<script>
  import { FloatingChatScreen } from './components';

  function handleBack() {
    // Navigate back
  }
</script>

<FloatingChatScreen showBack={true} on:back={handleBack}>
  <h1>Your content here</h1>
</FloatingChatScreen>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `showBack` | `boolean` | `true` | Show back button |

#### Events
| Event | Detail | Description |
|-------|--------|-------------|
| `back` | `void` | Fired when back button clicked |

---

### ChatInput

Chat-style rounded input with send button and optional mic.

```svelte
<script>
  import { ChatInput } from './components';

  let value = "";

  function handleSubmit() {
    console.log('Submitted:', value);
    value = "";
  }
</script>

<ChatInput 
  bind:value={value}
  placeholder="Type here..."
  showMic={true}
  on:submit={handleSubmit}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `value` | `string` | `""` | Input value (two-way binding) |
| `placeholder` | `string` | `"Type here..."` | Placeholder text |
| `showMic` | `boolean` | `false` | Show microphone button |
| `disabled` | `boolean` | `false` | Disable input |
| `className` | `string` | `""` | Additional CSS classes |

#### Events
| Event | Detail | Description |
|-------|--------|-------------|
| `submit` | `void` | Fired on Enter or send click |
| `change` | `{ value: string }` | Fired on input change |
| `mic` | `void` | Fired on mic button click |

---

### Button

Versatile button component with multiple variants and sizes.

```svelte
<script>
  import { Button } from './components';
</script>

<Button variant="primary" size="lg" on:click={handleClick}>
  Get Started
</Button>

<Button variant="outline" className="rounded-full">
  Skip for now
</Button>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `variant` | `string` | `"default"` | `default`, `destructive`, `outline`, `secondary`, `ghost`, `link` |
| `size` | `string` | `"default"` | `default`, `sm`, `lg`, `icon`, `icon-sm`, `icon-lg` |
| `disabled` | `boolean` | `false` | Disable button |
| `type` | `string` | `"button"` | Button type attribute |
| `className` | `string` | `""` | Additional CSS classes |

---

### MessageBubble

Chat message bubble for agent or user messages.

```svelte
<script>
  import { MessageBubble } from './components';
</script>

<MessageBubble role="agent" content="Hello! How can I help you today?" />
<MessageBubble role="user" content="I want to set up my workspace." />
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `role` | `string` | `"agent"` | `"agent"` or `"user"` |
| `content` | `string` | `""` | Message content |
| `timestamp` | `string` | `""` | Optional timestamp |
| `showAvatar` | `boolean` | `true` | Show orb avatar for agent |

---

### TypingIndicator

Animated dots showing the AI is "thinking".

```svelte
<script>
  import { TypingIndicator } from './components';
</script>

<TypingIndicator message="MIOSA is thinking..." />
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `message` | `string` | `"MIOSA is thinking..."` | Status message |

---

### ActionButtons

Quick action buttons below agent messages.

```svelte
<script>
  import { ActionButtons } from './components';

  const buttons = [
    { label: 'Marketing', action: 'select:marketing' },
    { label: 'Software', action: 'select:software' },
    { label: 'Other...', action: 'other', variant: 'ghost' }
  ];

  function handleAction(event) {
    console.log('Action:', event.detail.action);
  }
</script>

<ActionButtons {buttons} layout="inline" on:action={handleAction} />
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `buttons` | `Array` | `[]` | Array of `{ label, action, variant? }` |
| `layout` | `string` | `"inline"` | `"inline"`, `"grid"`, `"stack"` |
| `className` | `string` | `""` | Additional CSS classes |

---

### IntegrationCard

Card for displaying OAuth integration options.

```svelte
<script>
  import { IntegrationCard } from './components';

  function handleConnect(event) {
    console.log('Connect:', event.detail.provider);
  }
</script>

<IntegrationCard
  provider="google"
  name="Google Workspace"
  description="Calendar • Gmail • Drive"
  connected={false}
  on:connect={handleConnect}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `provider` | `string` | `""` | Provider ID |
| `name` | `string` | `""` | Display name |
| `description` | `string` | `""` | Features description |
| `icon` | `string` | `""` | Icon URL (optional) |
| `connected` | `boolean` | `false` | Connection status |
| `connecting` | `boolean` | `false` | Loading state |

---

### ToolPicker

Multi-select grid for choosing business tools.

```svelte
<script>
  import { ToolPicker } from './components';

  let selectedTools = [];

  function handleContinue(event) {
    console.log('Selected:', event.detail.selectedTools);
  }
</script>

<ToolPicker 
  bind:selectedTools
  showOtherInput={true}
  on:continue={handleContinue}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `tools` | `Array` | Default 9 tools | Array of `{ id, name, icon? }` |
| `selectedTools` | `string[]` | `[]` | Selected tool IDs |
| `otherTools` | `string` | `""` | Other tools text |
| `showOtherInput` | `boolean` | `true` | Show "Other" text input |

---

### EmailInviteInput

Email input for inviting team members.

```svelte
<script>
  import { EmailInviteInput } from './components';

  let emails = [];

  function handleSend(event) {
    console.log('Send to:', event.detail.emails);
  }

  function handleSkip() {
    // Skip invites
  }
</script>

<EmailInviteInput 
  bind:emails
  on:send={handleSend}
  on:skip={handleSkip}
/>
```

---

### FallbackForm

Traditional form shown when agent confidence is low.

```svelte
<script>
  import { FallbackForm } from './components';

  function handleSubmit(event) {
    console.log('Field:', event.detail.field, 'Value:', event.detail.value);
  }
</script>

<FallbackForm 
  field="industry"
  on:submit={handleSubmit}
/>
```

#### Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `field` | `string` | `"industry"` | `"business_name"`, `"industry"`, `"team_size"` |
| `value` | `string` | `""` | Initial value |

---

### CompletionScreen

Final onboarding completion summary.

```svelte
<script>
  import { CompletionScreen } from './components';

  const workspace = {
    name: 'Acme Corporation',
    industry: 'Marketing',
    teamSize: '11-50 people'
  };

  const integrations = [
    { name: 'Google Workspace', connected: true },
    { name: 'Slack', connected: true }
  ];

  const invites = [
    { email: 'sarah@acme.com' },
    { email: 'mike@acme.com' }
  ];
</script>

<CompletionScreen 
  {workspace}
  {integrations}
  {invites}
  on:gotoDashboard={() => navigate('/dashboard')}
/>
```

---

### MemberWelcome

Simplified welcome for team members (not admin onboarding).

```svelte
<script>
  import { MemberWelcome } from './components';

  const workspace = {
    name: 'Acme Corporation',
    inviterEmail: 'admin@acme.com',
    integrations: ['Google Workspace', 'Slack'],
    memberCount: 5
  };

  function handleGetStarted(event) {
    console.log('Name:', event.detail.name, 'Role:', event.detail.role);
  }
</script>

<MemberWelcome {workspace} on:getStarted={handleGetStarted} />
```

---

### ThemeToggle

Light/dark theme toggle button.

```svelte
<script>
  import { ThemeToggle } from './components';
</script>

<ThemeToggle />
```

Theme state is automatically persisted to `localStorage`.

---

## Icons

All icons are SVG-based Svelte components.

```svelte
<script>
  import { 
    SendIcon, 
    MicIcon, 
    ArrowLeftIcon, 
    SunIcon, 
    MoonIcon, 
    CheckIcon 
  } from './components';
</script>

<SendIcon size={24} className="text-primary" />
<MicIcon size={20} />
<ArrowLeftIcon size={20} />
<CheckIcon size={16} />
```

### Icon Props
| Prop | Type | Default | Description |
|------|------|---------|-------------|
| `size` | `number` | `24` | Width and height |
| `className` | `string` | `""` | Additional CSS classes |

---

## Animations & Effects

### Keyframe Animations

#### 1. Pulse Orb
```css
@keyframes pulse-orb {
  0%, 100% {
    transform: scale(1);
    filter: brightness(1);
  }
  50% {
    transform: scale(1.05);
    filter: brightness(1.2);
  }
}
/* Duration: 2s, Timing: ease-in-out, Iteration: infinite */
```

#### 2. Shimmer
```css
@keyframes shimmer {
  0% {
    background-position: -1000px 0;
  }
  100% {
    background-position: 1000px 0;
  }
}
/* Duration: 3s, Timing: linear, Iteration: infinite */
```

#### 3. Fade In
```css
@keyframes fade-in {
  from { opacity: 0; }
  to { opacity: 1; }
}
/* Duration: 300ms, Timing: ease-out */
```

#### 4. Slide Up
```css
@keyframes slide-up {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
/* Duration: 300ms, Timing: ease-out */
```

#### 5. Blink Cursor
```css
@keyframes blink-cursor {
  0%, 49% { opacity: 1; }
  50%, 100% { opacity: 0; }
}
/* Duration: 1s, Timing: step-end, Iteration: infinite */
```

#### 6. Bounce (Typing Dots)
```css
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-4px); }
}
/* Duration: 600ms, Timing: ease-in-out, Iteration: infinite */
```

### Animation Classes

```css
.animate-pulse-orb   /* Orb breathing effect */
.animate-shimmer     /* Light sweep effect */
.animate-fade-in     /* Fade in entrance */
.animate-slide-up    /* Slide up entrance */
.animate-blink       /* Cursor blink */
.animate-bounce      /* Typing indicator dots */
.animate-scale-in    /* Scale up entrance */
.animate-spin        /* Loading spinner */
```

### Animation Delay Utilities

```css
.animation-delay-0    /* 0ms */
.animation-delay-150  /* 150ms */
.animation-delay-300  /* 300ms */
.animation-delay-500  /* 500ms */
```

---

## Usage Examples

### Basic Onboarding Flow

```svelte
<script>
  import { 
    FloatingChatScreen, 
    SequentialTypewriter, 
    ChatInput 
  } from './svelte-onboarding/components';

  let showInput = false;
  let userInput = '';

  function handleTypeComplete() {
    showInput = true;
  }

  function handleSubmit() {
    console.log('User said:', userInput);
    // Process and move to next step
  }
</script>

<FloatingChatScreen showBack={false}>
  <div class="text-center w-full">
    <SequentialTypewriter
      lines={[
        "Welcome to BusinessOS!",
        "I'll help you set up your workspace.",
        "What's your company name?"
      ]}
      on:complete={handleTypeComplete}
    />
    
    {#if showInput}
      <div class="animate-fade-in mt-8">
        <ChatInput
          bind:value={userInput}
          placeholder="Enter your company name..."
          on:submit={handleSubmit}
        />
      </div>
    {/if}
  </div>
</FloatingChatScreen>
```

### Chat Interface

```svelte
<script>
  import { 
    MessageBubble, 
    TypingIndicator, 
    ActionButtons 
  } from './svelte-onboarding/components';

  let messages = [
    { role: 'agent', content: 'What industry are you in?' },
    { role: 'user', content: "We're a marketing agency" },
    { role: 'agent', content: 'Great! How many people work at your company?' }
  ];

  let isTyping = false;

  const quickActions = [
    { label: '1-10', action: 'size:1-10' },
    { label: '11-50', action: 'size:11-50' },
    { label: '50+', action: 'size:50+' }
  ];
</script>

<div class="chat-messages">
  {#each messages as msg}
    <MessageBubble role={msg.role} content={msg.content} />
  {/each}
  
  {#if isTyping}
    <TypingIndicator />
  {/if}
</div>

<ActionButtons buttons={quickActions} on:action={handleAction} />
```

---

## Theming

### Enabling Dark Mode

Add the `dark` class to the `<html>` or `<body>` element:

```javascript
// Toggle dark mode
document.documentElement.classList.toggle('dark', isDark);
```

Or use the `ThemeToggle` component which handles this automatically.

### Custom Theme Colors

Override CSS variables in your app's stylesheet:

```css
:root {
  /* Custom primary color */
  --primary: #6366f1;  /* Indigo */
  --primary-foreground: #ffffff;
  
  /* Custom orb colors */
  --orb-gradient-3: #8b5cf6;  /* Purple */
  --orb-glow: rgba(139, 92, 246, 0.4);
}

.dark {
  --primary: #818cf8;
  --primary-foreground: #000000;
}
```

---

## Summary

This documentation covers all the Svelte components converted from the original Next.js codebase for the BusinessOS onboarding experience. The components maintain:

- ✅ Original design aesthetics
- ✅ Purple orb with pulse and shimmer animations
- ✅ Typewriter text effects
- ✅ Light/dark theme support
- ✅ Responsive design
- ✅ Accessibility features
- ✅ Full TypeScript support

For questions or issues, refer to the original architecture document: `ONBOARDING_ARCHITECTURE.md`

---

*Documentation generated for BusinessOS Svelte Onboarding Components*
