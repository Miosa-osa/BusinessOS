# BusinessOS Frontend Documentation

**Version:** 2.0.0
**Last Updated:** January 19, 2026
**Stack:** SvelteKit + TypeScript + Tailwind CSS + Svelte 5

Welcome to the comprehensive documentation for the BusinessOS frontend application. This is your central hub for all frontend-related documentation, guides, and references.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Documentation Index](#documentation-index)
3. [Features](#features)
4. [Architecture](#architecture)
5. [Components](#components)
6. [Development](#development)
7. [Team Resources](#team-resources)

---

## Quick Start

### Setup

See [setup/GETTING_STARTED_OSA.md](./setup/GETTING_STARTED_OSA.md) for complete setup instructions.

**Quick commands:**

```bash
cd /Users/rhl/Desktop/BusinessOS2/frontend
npm install
npm run dev
```

### Key Directories

```
frontend/
├── src/
│   ├── routes/              # SvelteKit routes (app pages)
│   ├── lib/
│   │   ├── components/      # Reusable components
│   │   ├── stores/          # Svelte stores (state)
│   │   └── utils/           # Utility functions
│   └── app.css              # Global styles + btn-pill system
├── static/                  # Static assets
└── docs/                    # THIS DOCUMENTATION
```

### Tech Stack

- **Framework:** SvelteKit (Svelte 5 with new Snippet API)
- **Language:** TypeScript (strict mode)
- **Styling:** Tailwind CSS + Custom CSS variables
- **State:** Svelte stores (reactive state management)
- **Forms:** SvelteKit form actions
- **Icons:** Lucide Svelte

---

## Documentation Index

### Features Documentation

**Major Features:**

- [Onboarding System](./features/onboarding/ONBOARDING_SYSTEM.md) - Complete onboarding flow (OAuth → Username → Gmail → AI Analysis → Starter Apps)
- [Onboarding Quick Reference](./features/onboarding/QUICK_REFERENCE.md) - Quick reference for onboarding implementation
- [Onboarding README](./features/onboarding/README_ONBOARDING.md) - Overview of onboarding system
- [Button System](./features/buttons/BUTTON_SYSTEM.md) - Unified btn-pill component system (100+ files migrated)
- [App Store System](./features/app-store/APP_STORE_SYSTEM.md) - Starter apps with social features
- [Notifications](./features/FRONTEND_NOTIFICATIONS_GUIDE.md) - Real-time notifications system
- [Workspace](./features/workspace/) - Workspace frontend integration and memory UI

**Feature Highlights:**

| Feature | Status | Documentation |
|---------|--------|---------------|
| Google OAuth Onboarding | ✅ Complete | [Onboarding System](./features/onboarding/ONBOARDING_SYSTEM.md) |
| Button Standardization (btn-pill) | ✅ Complete | [Button System](./features/buttons/BUTTON_SYSTEM.md) |
| App Store with Starter Apps | ✅ Complete | [App Store](./features/app-store/APP_STORE_SYSTEM.md) |
| 3D Desktop Environment | 🚧 In Progress | [3D Desktop](./architecture/3D_DESKTOP_ARCHITECTURE.md) |
| Gesture System | 📋 Planned | (archived in main docs) |
| Voice Integration | 📋 Planned | (archived in main docs) |

### Architecture Documentation

**System Design:**

- [iOS to Desktop Architecture](./architecture/IOS_TO_DESKTOP_ARCHITECTURE.md) - Complete migration from iOS to Desktop
- [3D Desktop Architecture](./architecture/3D_DESKTOP_ARCHITECTURE.md) - 3D desktop environment design
- [3D Desktop Feature](./architecture/3D_DESKTOP_FEATURE.md) - 3D desktop feature specifications

**Key Architectural Decisions:**

1. **Svelte 5 Snippet API** - Using new runes system (`$state`, `$derived`, `$effect`)
2. **Button Standardization** - Unified btn-pill system across 100+ files
3. **Onboarding Flow** - Multi-step OAuth with AI-powered personalization
4. **Desktop-First Design** - Optimized for 1440-1920px screens

### Component Library

**Core Components:**

- [Form Components Usage](./components/FORM_COMPONENTS_USAGE_GUIDE.md) - Form component patterns
- [Form Patterns Index](./components/FORM_PATTERNS_INDEX.md) - Index of form patterns
- **PillButton** - See [Button System](./features/buttons/BUTTON_SYSTEM.md)

**Component Index:**

| Component | Location | Documentation |
|-----------|----------|---------------|
| `PillButton` | `lib/components/osa/PillButton.svelte` | [Button System](./features/buttons/BUTTON_SYSTEM.md) |
| `BuildProgress` | `lib/components/osa/BuildProgress.svelte` | Inline docs |
| Form Components | `lib/components/forms/` | [Form Components](./components/FORM_COMPONENTS_USAGE_GUIDE.md) |
| Desktop Components | `lib/components/desktop/` | Inline docs |
| Dock | `lib/components/desktop/Dock.svelte` | Inline docs |

### Development Guides

- [Frontend Development Guide](./development/FRONTEND.md) - Complete frontend development guide
- [Getting Started with OSA](./setup/GETTING_STARTED_OSA.md) - Setup and onboarding

**Development Standards:**

```typescript
// TypeScript strict mode required
"strict": true

// Svelte 5 patterns
import { PillButton } from '$lib/components/osa';

let count = $state(0); // Reactive state
let doubled = $derived(count * 2); // Derived value

$effect(() => {
  console.log('Count changed:', count);
}); // Side effects
```

**CSS Standards:**

```css
/* Use CSS custom properties */
--text-primary: #1A1A1A;
--text-secondary: #666666;

/* Use btn-pill system for buttons */
<button class="btn-pill btn-pill-primary">Submit</button>

/* Use Tailwind utilities for layout */
<div class="flex gap-4 items-center">...</div>
```

---

## Features

### Onboarding Flow (Complete)

**Path:** `/routes/onboarding/`

**Screens:**

1. Welcome screen (`+page.svelte`)
2. Sign in with Google (`signin/+page.svelte`)
3. Username setup (`username/+page.svelte`)
4. Gmail integration (`gmail/+page.svelte`)
5. AI Analysis (3 screens: `analyzing/`, `analyzing-2/`, `analyzing-3/`)
6. Meet OSA (`meet-osa/+page.svelte`)
7. Starter Apps (`starter-apps/+page.svelte`)
8. Ready screen (`ready/+page.svelte`)

**Documentation:** [Onboarding System](./features/onboarding/ONBOARDING_SYSTEM.md)

### Button System (Complete)

**Standardized btn-pill system:**

- 100+ files migrated
- 585+ button instances
- 9 variants (primary, secondary, ghost, danger, success, warning, outline, soft, link)
- 5 sizes (xs, sm, md, lg, xl)
- Loading states, icon buttons, button groups

**Documentation:** [Button System](./features/buttons/BUTTON_SYSTEM.md)

### App Store (Complete)

**Starter apps with social features:**

- App discovery and browsing
- Like, comment, remix functionality
- Public/private app separation
- Builder interface (Chat, General, Icon, Prompts tabs)

**Documentation:** [App Store](./features/app-store/APP_STORE_SYSTEM.md)

### Desktop Environment (In Progress)

**3D desktop with window management:**

- Multi-window workflow
- Drag-and-drop between apps
- Window snapping and tiling
- Desktop icons with circular design

**Documentation:** [3D Desktop](./architecture/3D_DESKTOP_ARCHITECTURE.md)

---

## Architecture

### Tech Stack

```
Frontend Stack:
  ├── SvelteKit (SSR + Client-side routing)
  ├── TypeScript (Strict mode)
  ├── Tailwind CSS (Utility-first)
  ├── Svelte 5 (Runes API: $state, $derived, $effect)
  ├── Lucide Svelte (Icon library)
  └── Vite (Build tool)

State Management:
  ├── Svelte stores (reactive state)
  ├── Form actions (server-side mutations)
  └── Load functions (server-side data fetching)

Styling System:
  ├── Tailwind utilities (layout, spacing)
  ├── CSS custom properties (colors, shadows)
  └── btn-pill system (button standardization)
```

### Code Organization

```
src/
├── routes/
│   ├── (app)/               # Main app routes
│   │   ├── chat/
│   │   ├── dashboard/
│   │   ├── projects/
│   │   └── settings/
│   ├── onboarding/          # Onboarding flow
│   └── login/               # Auth routes
├── lib/
│   ├── components/
│   │   ├── osa/             # OSA-specific components
│   │   ├── desktop/         # Desktop environment
│   │   └── forms/           # Form components
│   ├── stores/              # Svelte stores
│   ├── utils/               # Utility functions
│   └── types/               # TypeScript types
└── app.css                  # Global styles
```

### Design System

**Colors:**

```css
--text-primary: #1A1A1A;
--text-secondary: #666666;
--bg-primary: #FFFFFF;
--bg-secondary: #F5F5F7;
```

**Typography:**

```css
--text-xs: 0.75rem;
--text-sm: 0.875rem;
--text-base: 1rem;
--text-lg: 1.125rem;
--text-xl: 1.25rem;
```

**Shadows:**

```css
--shadow-sm: 0 1px 2px 0 rgba(0, 0, 0, 0.05);
--shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
--shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
```

---

## Components

### Core Component Library

**OSA Components** (`lib/components/osa/`)

| Component | Purpose | Props |
|-----------|---------|-------|
| `PillButton` | Standardized button | `variant`, `size`, `loading`, `disabled` |
| `BuildProgress` | Build progress indicator | `progress`, `status` |

**Desktop Components** (`lib/components/desktop/`)

| Component | Purpose |
|-----------|---------|
| `Dock` | Application dock with app launcher |
| `Window` | Windowed app container |
| `DesktopIcon` | Desktop app icon |

**Form Components** (`lib/components/forms/`)

See [Form Components Usage](./components/FORM_COMPONENTS_USAGE_GUIDE.md) for complete documentation.

### Component Usage Patterns

**PillButton Examples:**

```svelte
<!-- Primary button -->
<PillButton variant="primary" onclick={handleSubmit}>
  Submit
</PillButton>

<!-- Loading state -->
<PillButton variant="primary" loading={isLoading}>
  {isLoading ? 'Saving...' : 'Save'}
</PillButton>

<!-- Different sizes -->
<PillButton variant="secondary" size="sm">Small</PillButton>
<PillButton variant="primary" size="lg">Large</PillButton>

<!-- Full width -->
<PillButton variant="primary" class="btn-pill-block">
  Full Width
</PillButton>
```

---

## Development

### Getting Started

1. **Clone the repository**

```bash
git clone <repo-url>
cd BusinessOS2/frontend
```

2. **Install dependencies**

```bash
npm install
```

3. **Run development server**

```bash
npm run dev
```

4. **Build for production**

```bash
npm run build
npm run preview
```

### Development Workflow

**Branch Strategy:**

- `main` - Production-ready code
- `feature/*` - New features
- `fix/*` - Bug fixes
- `docs/*` - Documentation updates

**Commit Conventions:**

```
feat: Add new onboarding screen
fix: Resolve button styling issue
docs: Update button system documentation
refactor: Simplify form component logic
style: Format code with prettier
```

### Code Standards

**TypeScript:**

```typescript
// ✅ Good - Explicit types
interface UserData {
  id: string;
  username: string;
  email: string;
}

function getUser(id: string): Promise<UserData> {
  return fetch(`/api/users/${id}`).then(r => r.json());
}

// ❌ Bad - Implicit any
function getUser(id) {
  return fetch(`/api/users/${id}`).then(r => r.json());
}
```

**Svelte Components:**

```svelte
<script lang="ts">
  import { PillButton } from '$lib/components/osa';

  // Reactive state with Svelte 5
  let count = $state(0);
  let doubled = $derived(count * 2);

  function increment() {
    count++;
  }
</script>

<div>
  <p>Count: {count} (Doubled: {doubled})</p>
  <PillButton variant="primary" onclick={increment}>
    Increment
  </PillButton>
</div>
```

**CSS/Tailwind:**

```svelte
<!-- ✅ Good - Use Tailwind utilities + btn-pill -->
<button class="btn-pill btn-pill-primary">
  Submit
</button>

<div class="flex gap-4 items-center">
  <span class="text-gray-600">Label</span>
  <PillButton variant="secondary">Action</PillButton>
</div>

<!-- ❌ Bad - Inline styles -->
<button style="background: blue; padding: 10px;">
  Submit
</button>
```

### Testing

**Unit Tests:**

```bash
npm run test
```

**E2E Tests:**

```bash
npm run test:e2e
```

### Building

**Production Build:**

```bash
npm run build
```

**Preview Production Build:**

```bash
npm run preview
```

---

## Team Resources

### Recent Changes

See [team-review/RECENT_FRONTEND_CHANGES.md](./team-review/RECENT_FRONTEND_CHANGES.md) for a summary of recent frontend changes.

**Major Changes (Q1 2026):**

1. ✅ Google OAuth onboarding flow (complete)
2. ✅ Button standardization (btn-pill system, 100+ files)
3. ✅ App Store with starter apps (complete)
4. 🚧 3D Desktop environment (in progress)

### Quick Links

**Documentation:**

- [Onboarding System](./features/onboarding/ONBOARDING_SYSTEM.md)
- [Button System](./features/buttons/BUTTON_SYSTEM.md)
- [App Store](./features/app-store/APP_STORE_SYSTEM.md)
- [Form Components](./components/FORM_COMPONENTS_USAGE_GUIDE.md)

**Architecture:**

- [iOS to Desktop Migration](./architecture/IOS_TO_DESKTOP_ARCHITECTURE.md)
- [3D Desktop Architecture](./architecture/3D_DESKTOP_ARCHITECTURE.md)

**Development:**

- [Frontend Development Guide](./development/FRONTEND.md)
- [Getting Started](./setup/GETTING_STARTED_OSA.md)

### Team Contacts

**Frontend Team:**

- Roberto - Architecture, coordination
- Nejd/Javaris - Frontend implementation, testing

**Related Teams:**

- Pedro - Backend (Go)
- Nick - Terminal integration, GCP deployment
- Abdul - E2B sandbox integration

---

## Contributing

### Adding New Features

1. Check existing documentation for similar patterns
2. Follow the button system for UI components
3. Use TypeScript strict mode
4. Add tests for new functionality
5. Update documentation

### Updating Documentation

1. Keep documentation in sync with code
2. Add examples for complex features
3. Update this README when adding new major features
4. Use clear, concise language

---

## Additional Resources

### External Documentation

- [SvelteKit Docs](https://kit.svelte.dev/)
- [Svelte 5 Docs](https://svelte.dev/)
- [Tailwind CSS Docs](https://tailwindcss.com/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)

### Internal Resources

- Main project docs: `/Users/rhl/Desktop/BusinessOS2/docs/`
- Backend docs: `/Users/rhl/Desktop/BusinessOS2/desktop/backend-go/docs/`
- API documentation: `/Users/rhl/Desktop/BusinessOS2/docs/api/`

---

**Last Updated:** January 19, 2026
**Maintained By:** BusinessOS Frontend Team
**Version:** 2.0.0
