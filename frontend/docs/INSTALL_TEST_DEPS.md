# Installing Test Dependencies

Run these commands to install all required testing dependencies:

## Install Core Testing Libraries

```bash
npm install --save-dev vitest@^2.1.8
npm install --save-dev @vitest/coverage-v8@^2.1.8
npm install --save-dev jsdom@^25.0.1
npm install --save-dev @testing-library/svelte@^5.2.7
npm install --save-dev @testing-library/jest-dom@^6.6.5
```

## Or Install All at Once

```bash
npm install --save-dev vitest @vitest/coverage-v8 jsdom @testing-library/svelte @testing-library/jest-dom
```

## Add Test Scripts to package.json

Open `package.json` and add these scripts to the `"scripts"` section:

```json
{
  "scripts": {
    "dev": "vite dev",
    "build": "vite build",
    "preview": "vite preview",
    "prepare": "svelte-kit sync || echo ''",
    "check": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json",
    "check:watch": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json --watch",
    "test": "vitest",
    "test:ui": "vitest --ui",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage"
  }
}
```

## Verify Installation

After installing, verify the setup works:

```bash
npm run test:run
```

Expected output:
```
 ✓ src/lib/api/ai/customAgents.test.ts (XX tests)
 ✓ src/lib/stores/agents.test.ts (XX tests)
 ✓ src/lib/components/agents/AgentCard.test.ts (XX tests)
 ✓ src/lib/components/agents/AgentSandbox.test.ts (XX tests)

Test Files  4 passed (4)
     Tests  XXX passed (XXX)
```

## Files Created

The following files have been created:

1. **Test Files**:
   - `src/lib/api/ai/customAgents.test.ts` - API client tests
   - `src/lib/stores/agents.test.ts` - Store tests
   - `src/lib/components/agents/AgentCard.test.ts` - AgentCard component tests
   - `src/lib/components/agents/AgentSandbox.test.ts` - AgentSandbox component tests

2. **Configuration**:
   - `vitest.config.ts` - Vitest configuration
   - `src/test/setup.ts` - Test setup and global mocks

3. **Documentation**:
   - `TEST_README.md` - Comprehensive testing guide
   - `INSTALL_TEST_DEPS.md` - This file

## Next Steps

1. Install dependencies: `npm install`
2. Run tests: `npm test`
3. Generate coverage: `npm run test:coverage`
4. View coverage report: Open `coverage/index.html` in browser

## Troubleshooting

### If you get "Cannot find module" errors:

Make sure `vitest.config.ts` exists and has the correct path aliases.

### If tests timeout:

The default timeout is 10 seconds. For slower systems, increase it in `vitest.config.ts`:

```typescript
test: {
  testTimeout: 20000,
  hookTimeout: 20000
}
```

### If you get type errors:

Make sure TypeScript can find the testing types. Add to `tsconfig.json`:

```json
{
  "compilerOptions": {
    "types": ["vitest/globals", "@testing-library/jest-dom"]
  }
}
```
