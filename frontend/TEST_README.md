# Custom Agents Test Suite

Comprehensive unit test suite for the Custom Agents system in BusinessOS.

## Overview

This test suite provides thorough coverage of:
- API Client functions (Custom Agents API)
- Store logic (Agents store with Svelte stores)
- Component behavior (AgentCard and AgentSandbox)

## Test Files

### 1. API Client Tests
**File**: `src/lib/api/ai/customAgents.test.ts`

Tests all API functions:
- `getCustomAgents()` - Fetch all agents with filters
- `getCustomAgent()` - Fetch single agent
- `createCustomAgent()` - Create new agent
- `updateCustomAgent()` - Update existing agent
- `deleteCustomAgent()` - Delete agent
- `getAgentsByCategory()` - Filter by category
- `getAgentPresets()` - Fetch presets
- `getAgentPreset()` - Fetch single preset
- `createFromPreset()` - Create agent from preset
- `testAgent()` - Test agent with streaming
- `testSandbox()` - Test sandbox configuration

**Coverage**: API error handling, validation, streaming responses, URL encoding

### 2. Store Tests
**File**: `src/lib/stores/agents.test.ts`

Tests the agents Svelte store:
- `loadAgents()` - Load with filters (category, search, status)
- `loadAgent()` - Load single agent
- `createAgent()` - Create and add to store
- `updateAgent()` - Update in store and currentAgent
- `deleteAgent()` - Remove from store
- `setCurrentAgent()` / `clearCurrent()`
- `setFilters()` / `clearFilters()`
- `clearError()`
- Preset methods: `loadPresets()`, `loadPreset()`, `createFromPreset()`
- Test methods: `testAgent()`, `testSandbox()`

**Derived Stores**:
- `selectedAgent` - Current agent selection
- `agentsByCategory` - Agents grouped by category
- `activeAgents` - Only active agents

**Coverage**: State management, filtering logic, API integration, error handling

### 3. AgentCard Component Tests
**File**: `src/lib/components/agents/AgentCard.test.ts`

Tests the AgentCard UI component:
- Rendering: All agent information, badges, status indicators
- Avatar: Image vs initials generation
- Interactions: Click, keyboard (Enter/Space), Select button
- Menu: Edit/Delete actions, confirmation flow
- Event propagation: Proper click handling
- Variants: Default and compact modes
- Accessibility: ARIA attributes, keyboard navigation
- Edge cases: Long names, missing data, empty values

**Coverage**: Component rendering, user interactions, accessibility, edge cases

### 4. AgentSandbox Component Tests
**File**: `src/lib/components/agents/AgentSandbox.test.ts`

Tests the AgentSandbox testing interface:
- Rendering: Basic elements, advanced options
- Input: Message input, model override, temperature slider
- Testing modes: Agent ID vs Sandbox configuration
- SSE Streaming: Content streaming, metadata handling
- Loading states: Spinner, Stop button, disabled inputs
- Error handling: API errors, missing config, null streams
- History: Adding entries, toggling visibility, item limit (5)
- Callbacks: onTest callback with results
- Utilities: Duration/token formatting

**Coverage**: Real-time streaming, SSE parsing, error states, user interactions

## Setup

### Install Dependencies

```bash
npm install --save-dev vitest @testing-library/svelte @testing-library/jest-dom @vitest/coverage-v8 jsdom
```

### Required Dependencies

Add to `package.json`:

```json
{
  "devDependencies": {
    "vitest": "^2.1.8",
    "@testing-library/svelte": "^5.2.7",
    "@testing-library/jest-dom": "^6.6.5",
    "@vitest/coverage-v8": "^2.1.8",
    "jsdom": "^25.0.1"
  }
}
```

### Test Scripts

Add to `package.json`:

```json
{
  "scripts": {
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage"
  }
}
```

## Running Tests

### Run all tests in watch mode
```bash
npm test
```

### Run tests once (CI mode)
```bash
npm run test:run
```

### Run with UI
```bash
npm run test:ui
```

### Generate coverage report
```bash
npm run test:coverage
```

## Test Structure

Each test file follows this structure:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('Feature Name', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Sub-feature', () => {
    it('should do something specific', () => {
      // Arrange
      const mockData = { ... };

      // Act
      const result = functionUnderTest(mockData);

      // Assert
      expect(result).toBe(expected);
    });
  });
});
```

## Mocking Strategy

### API Mocking
```typescript
vi.mock('$lib/api/ai/ai', () => ({
  getCustomAgents: vi.fn(),
  createCustomAgent: vi.fn()
}));
```

### Store Mocking
```typescript
import { get } from 'svelte/store';
const state = get(agents);
```

### Component Mocking
```typescript
import { render, fireEvent } from '@testing-library/svelte';
const { container } = render(Component, { props: { ... } });
```

### SSE Stream Mocking
```typescript
function createMockSSEStream(events) {
  return new ReadableStream({
    start(controller) {
      for (const event of events) {
        controller.enqueue(encoder.encode(`data: ${JSON.stringify(event)}\n\n`));
      }
      controller.close();
    }
  });
}
```

## Coverage Goals

- **Lines**: 80%+
- **Functions**: 80%+
- **Branches**: 80%+
- **Statements**: 80%+

## Critical Test Scenarios

### API Tests
- ✅ Success responses
- ✅ Error responses (404, 500, network errors)
- ✅ Empty data
- ✅ Validation errors
- ✅ Streaming responses
- ✅ URL encoding

### Store Tests
- ✅ State updates
- ✅ Filter combinations
- ✅ Derived stores
- ✅ Error propagation
- ✅ Concurrent operations

### Component Tests
- ✅ User interactions
- ✅ Keyboard navigation
- ✅ Accessibility
- ✅ Edge cases (long text, missing data)
- ✅ Event propagation
- ✅ Loading states
- ✅ Error displays

## Best Practices

1. **Isolation**: Each test is independent and can run in any order
2. **Cleanup**: Use `beforeEach` and `afterEach` for setup/teardown
3. **Mocking**: Mock external dependencies, test internal logic
4. **Assertions**: Clear, specific expectations
5. **Naming**: Descriptive test names following "should..." pattern
6. **Coverage**: Test success paths, error paths, and edge cases

## Debugging Tests

### Run single test file
```bash
npm test -- customAgents.test.ts
```

### Run specific test
```bash
npm test -- -t "should fetch all custom agents"
```

### Enable verbose output
```bash
npm test -- --reporter=verbose
```

### Debug in VS Code
Add to `.vscode/launch.json`:
```json
{
  "type": "node",
  "request": "launch",
  "name": "Debug Vitest Tests",
  "runtimeExecutable": "npm",
  "runtimeArgs": ["run", "test"],
  "console": "integratedTerminal",
  "internalConsoleOptions": "neverOpen"
}
```

## Continuous Integration

### GitHub Actions Example
```yaml
- name: Run Tests
  run: npm run test:run

- name: Generate Coverage
  run: npm run test:coverage

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage/coverage-final.json
```

## Common Issues

### Issue: "Cannot find module '$lib/...'"
**Solution**: Check `vitest.config.ts` has correct path aliases

### Issue: "ReferenceError: ReadableStream is not defined"
**Solution**: Ensure `environment: 'jsdom'` in vitest.config.ts

### Issue: "TypeError: Cannot read property 'subscribe' of undefined"
**Solution**: Mock Svelte stores in test setup

### Issue: Tests timeout
**Solution**: Increase `testTimeout` in vitest.config.ts or use `waitFor` with longer timeout

## Contributing

When adding new features:
1. Write tests first (TDD)
2. Ensure 80%+ coverage
3. Test success, error, and edge cases
4. Update this README if needed

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [Testing Library Svelte](https://testing-library.com/docs/svelte-testing-library/intro/)
- [Testing Library Queries](https://testing-library.com/docs/queries/about)
- [Jest DOM Matchers](https://github.com/testing-library/jest-dom)
