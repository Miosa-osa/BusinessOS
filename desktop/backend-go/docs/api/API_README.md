# BusinessOS Frontend API Patterns - Complete Documentation

## Executive Summary

This documentation provides a complete analysis of API function patterns used in BusinessOS frontend, with production-ready templates and best practices.

**Status:** Complete analysis based on production code
**Source:** `frontend/src/lib/api/ai/ai.ts` and related files
**Purpose:** Guide for implementing new API endpoints following established patterns

---

## What's Included

This documentation package contains **5 comprehensive guides**:

### 1. API_DOCUMENTATION_INDEX.md
**Purpose:** Navigation guide for all documentation
- Quick start by use case
- Key concepts summary
- File structure reference
- Implementation checklist

**Use when:** You need to navigate the documentation or understand what's available

---

### 2. API_PATTERNS_ANALYSIS.md (MAIN REFERENCE)
**Purpose:** Deep-dive analysis of all API patterns
- **Section 1:** Function naming conventions with 3 categories
  - Read operations: `get*`
  - Write operations: `save*`, `update*`
  - Async operations: `pull*`, `warmup*`

- **Section 2:** Error handling patterns with 3 approaches
  - Automatic error handling via `request<T>()`
  - Custom error handling for streams
  - Type-safe discriminated unions

- **Section 3:** Type safety strategies
  - Generic type parameters
  - Response type hierarchy
  - Request body patterns
  - Optional fields

- **Section 4:** Request/response structure
  - Standard GET/POST/PUT/DELETE flows
  - Query parameters
  - Headers and credentials
  - Streaming responses

- **Section 5:** API endpoint structure
  - `/ai/*` endpoint organization
  - `/mcp/*` tool endpoints
  - Endpoint naming patterns

- **Section 6:** Complete template for custom agents
  - CRUD operations with examples
  - Streaming operations
  - Batch operations

- **Section 7:** Usage examples
  - Creating and executing agents
  - Streaming responses
  - Batch operations
  - Validation before save

- **Section 8:** Best practices summary

**Use when:** You need detailed understanding or implementing complex endpoints

---

### 3. API_TEMPLATE_CUSTOM_AGENTS.ts (READY TO USE)
**Purpose:** Copy-paste ready TypeScript template

Contains 20+ production-ready functions:
- **CRUD:** `createCustomAgent`, `getCustomAgent`, `updateCustomAgent`, `deleteCustomAgent`
- **Actions:** `executeCustomAgent`, `testCustomAgent`, `cloneCustomAgent`
- **Streaming:** `executeCustomAgentStream`
- **Batch:** `executeCustomAgentsBatch`, `deleteCustomAgentsBatch`
- **Utilities:** `getCustomAgentStats`, `exportCustomAgent`, `importCustomAgent`

Each function includes:
- Full JSDoc documentation
- Detailed parameter descriptions
- Return type information
- Real-world usage example
- Error handling pattern

**Use when:** Implementing new custom agent endpoints

---

### 4. API_CHEATSHEET.md (QUICK REFERENCE)
**Purpose:** One-page reference for quick lookup

Contains:
- Function naming quick reference
- 8 copy-paste request patterns
- Type safety quick lookup table
- Error handling patterns
- 5 common use cases with code
- Common mistakes with corrections
- Endpoint structure reference
- Quick method reference table

**Use when:** You forgot a pattern or need quick code snippet

---

### 5. API_VISUAL_GUIDE.md (DIAGRAMS & FLOWS)
**Purpose:** Visual representation of patterns

Contains:
- Function flow diagram
- Function naming structure diagram
- Request pattern decision tree
- Error handling flow diagram
- Type safety flow
- Response type hierarchy
- Endpoint structure map
- Request construction template
- Create vs Update patterns
- Decision matrix for pattern selection
- Visual checklist for new endpoints

**Use when:** You learn better with diagrams or need overview

---

## Core Patterns at a Glance

### Pattern 1: Function Naming
```typescript
[verb][Resource]()
```

**Verbs:**
- `get` - Fetch data
- `create` - Create new
- `save` - Persist config
- `update` - Modify existing
- `delete` - Remove item
- `execute` - Run operation
- `pull` - Download/stream
- `test` - Validate
- `clone` - Duplicate

**Examples:**
```typescript
getCustomAgents()
createCustomAgent(agent)
executeCustomAgent(id, input)
deleteCustomAgent(id)
```

---

### Pattern 2: Request Function
```typescript
request<ResponseType>(endpoint, {
  method: 'GET|POST|PUT|DELETE',
  body: { /* payload */ }
})
```

**Features:**
- Automatically sets `Content-Type: application/json`
- Automatically includes `credentials: 'include'`
- Automatically stringifies body to JSON
- Automatically throws formatted errors
- Fully type-safe with generics

**Examples:**
```typescript
// GET request
request<CustomAgentsResponse>('/ai/custom-agents')

// POST request
request<CustomAgent>('/ai/custom-agents', {
  method: 'POST',
  body: agentData
})

// PUT request
request<CustomAgent>('/ai/custom-agents/123', {
  method: 'PUT',
  body: { name: 'New Name' }
})
```

---

### Pattern 3: Error Handling
```typescript
// Standard endpoints (automatic)
try {
  const result = await createCustomAgent(data);
} catch (error) {
  console.error(error.message);  // "Error (HTTP 400)"
}

// Streaming endpoints (custom)
const response = await fetch(url, options);
if (!response.ok) {
  const error = await response.json().catch(() => ({ detail: 'Failed' }));
  throw new Error(error.detail || `Failed (HTTP ${response.status})`);
}
```

---

### Pattern 4: Type Safety
```typescript
// Always specify generic type
const agents = await request<CustomAgentsResponse>('/ai/custom-agents');
// agents.agents is now properly typed

// Use discriminated unions
type Result =
  | { success: true; data: string }
  | { success: false; error: string };

const result = await validate();
if (result.success) {
  console.log(result.data);  // data available
}
```

---

### Pattern 5: Request Body Types
```typescript
// Creating new (omit auto-fields)
type CreateAgentRequest = Omit<CustomAgent,
  'id' | 'user_id' | 'created_at' | 'updated_at'>;

// Updating (partial changes)
type UpdateAgentRequest = Partial<Omit<CustomAgent,
  'id' | 'user_id' | 'created_at' | 'updated_at'>>;

// Custom actions
type ExecuteAgentRequest = {
  prompt: string;
  context?: Record<string, unknown>;
};
```

---

## Quick Start Scenarios

### Scenario 1: I need to add a new GET endpoint
**Steps:**
1. Read: `API_PATTERNS_ANALYSIS.md` - Section 4 (GET examples)
2. Copy: `API_TEMPLATE_CUSTOM_AGENTS.ts` - `getCustomAgent()` function
3. Adapt: Change endpoint path and response type
4. Reference: `API_CHEATSHEET.md` - Pattern 1-3

**Time:** 5 minutes

---

### Scenario 2: I need to implement a POST endpoint
**Steps:**
1. Read: `API_PATTERNS_ANALYSIS.md` - Section 4 (POST examples)
2. Copy: `API_TEMPLATE_CUSTOM_AGENTS.ts` - `createCustomAgent()` function
3. Define: Response type in `types.ts`
4. Reference: `API_CHEATSHEET.md` - Pattern 4

**Time:** 10 minutes

---

### Scenario 3: I need to add a streaming endpoint
**Steps:**
1. Read: `API_PATTERNS_ANALYSIS.md` - Section 4 (Streaming section)
2. Copy: `API_TEMPLATE_CUSTOM_AGENTS.ts` - `executeCustomAgentStream()` function
3. Reference: `API_VISUAL_GUIDE.md` - Error handling flow diagram

**Time:** 15 minutes

---

### Scenario 4: I'm debugging an API error
**Steps:**
1. Check: `API_CHEATSHEET.md` - Error handling section
2. Review: `API_PATTERNS_ANALYSIS.md` - Section 2
3. Look up: Error format and common causes

**Time:** 5 minutes

---

### Scenario 5: I forgot the exact syntax
**Steps:**
1. Check: `API_CHEATSHEET.md` - Request pattern cookbook
2. Copy: Exact code pattern needed
3. Adapt: To your endpoint

**Time:** 2 minutes

---

## File Map

```
documentation/
├── API_README.md (THIS FILE)
│   └─ Overview of entire documentation package
│
├── API_DOCUMENTATION_INDEX.md
│   └─ Navigation guide and quick start by use case
│
├── API_PATTERNS_ANALYSIS.md ⭐ MAIN REFERENCE
│   └─ Complete analysis with explanations
│
├── API_TEMPLATE_CUSTOM_AGENTS.ts
│   └─ Copy-paste ready TypeScript code
│
├── API_CHEATSHEET.md ⭐ QUICK LOOKUP
│   └─ One-page reference
│
└── API_VISUAL_GUIDE.md
    └─ Diagrams and visual flows

frontend/src/lib/api/
├── base.ts
│   └─ request<T>() wrapper and fetch utilities
├── ai/
│   ├── ai.ts
│   │   └─ Current implementations (patterns demonstrated here)
│   └── types.ts
│       └─ Type definitions
```

---

## Key Learnings

### 1. Consistency is Power
All API functions follow the same patterns:
- Same naming convention
- Same request structure
- Same error handling
- Same type safety approach

This makes it easy to learn once, then apply everywhere.

---

### 2. Type Safety Eliminates Bugs
```typescript
// Type-safe (compiler enforces)
const agent: CustomAgent = await getCustomAgent(id);
agent.system_prompt.toLowerCase();  // ✓ Safe

// Not type-safe (compiler doesn't check)
const agent: any = await getCustomAgent(id);
agent.nonexistent_field;  // ✗ No error caught
```

---

### 3. Error Format is Standardized
```
"[Error message from server] (HTTP [status code])"
```

All errors follow this format, making them predictable and easy to handle.

---

### 4. request<T>() Handles Most Cases
The `request<T>()` wrapper in `base.ts` handles:
- Setting headers
- Stringifying body
- Adding credentials
- Parsing response
- Formatting errors

Only use custom fetch for streaming endpoints.

---

### 5. Omit<> and Partial<> Pattern
```typescript
// Create: Omit auto-generated fields
Omit<T, 'id' | 'user_id' | 'created_at' | 'updated_at'>

// Update: Only changed fields
Partial<Omit<T, auto-generated fields>>
```

This ensures type safety and prevents accidents.

---

## Common Patterns Summary

| Need | Pattern | Location |
|------|---------|----------|
| GET single | `request<T>()` with `/path/:id` | Cheatsheet Pattern 2 |
| GET list | `request<T>()` with `/path` | Cheatsheet Pattern 1 |
| GET filtered | `request<T>()` with `?param=value` | Cheatsheet Pattern 3 |
| POST create | `request<T>(..., { method: 'POST', body })` | Cheatsheet Pattern 4 |
| PUT update | `request<T>(..., { method: 'PUT', body })` | Cheatsheet Pattern 5 |
| DELETE | `request<T>(..., { method: 'DELETE' })` | Cheatsheet Pattern 6 |
| Stream | Custom fetch + `response.body` | Cheatsheet Pattern 7 |
| Batch | `request<T[]>(..., { method: 'POST', body })` | Cheatsheet Pattern 8 |

---

## Implementation Checklist

Before implementing a new endpoint:

- [ ] **Naming**
  - [ ] Uses verb-noun pattern
  - [ ] Verb clearly indicates action
  - [ ] Resource name is clear

- [ ] **Types**
  - [ ] Return type defined
  - [ ] Request type defined (if applicable)
  - [ ] Uses `Omit<>` for create operations
  - [ ] Uses `Partial<>` for updates

- [ ] **Implementation**
  - [ ] Uses `request<T>()` or custom fetch appropriately
  - [ ] Sets correct HTTP method
  - [ ] Uses correct endpoint
  - [ ] Request body uses snake_case

- [ ] **Error Handling**
  - [ ] Errors are properly handled
  - [ ] Error messages include HTTP status

- [ ] **Type Safety**
  - [ ] Generic type specified: `request<T>()`
  - [ ] No `any` types
  - [ ] TypeScript strict mode enabled

- [ ] **Documentation**
  - [ ] JSDoc with description
  - [ ] Parameter types documented
  - [ ] Return type documented
  - [ ] Usage example provided

- [ ] **Testing**
  - [ ] Happy path tested
  - [ ] Error cases tested
  - [ ] Type checking verified
  - [ ] Error format correct

---

## Best Practices

### ✅ DO

1. **Use request<T>() for standard endpoints**
   ```typescript
   return request<CustomAgent>('/ai/custom-agents', {
     method: 'POST',
     body: agent
   });
   ```

2. **Use custom fetch for streaming**
   ```typescript
   const response = await fetch(url, options);
   if (!response.ok) throw new Error(...);
   return response.body;
   ```

3. **Always specify generic type**
   ```typescript
   request<CustomAgentsResponse>('/ai/custom-agents')
   ```

4. **Use snake_case in request body**
   ```typescript
   body: { api_key: key, user_id: userId }
   ```

5. **Omit auto-generated fields when creating**
   ```typescript
   Omit<CustomAgent, 'id' | 'user_id' | 'created_at' | 'updated_at'>
   ```

6. **Use discriminated unions for complex responses**
   ```typescript
   type Result =
     | { valid: true; warnings?: string[] }
     | { valid: false; errors: string[] };
   ```

---

### ❌ DON'T

1. **Don't use `any` types**
   ```typescript
   // ✗ Bad
   const agent: any = await getCustomAgent(id);

   // ✓ Good
   const agent: CustomAgent = await getCustomAgent(id);
   ```

2. **Don't skip generic type**
   ```typescript
   // ✗ Bad
   const result = request('/endpoint');

   // ✓ Good
   const result = request<ResponseType>('/endpoint');
   ```

3. **Don't use custom fetch for JSON responses**
   ```typescript
   // ✗ Bad - Use request<T>() instead
   const response = await fetch(url);
   const data = await response.json();

   // ✓ Good
   const data = request<T>(endpoint);
   ```

4. **Don't mix camelCase and snake_case in body**
   ```typescript
   // ✗ Bad
   body: { apiKey: '...', user_id: '...' }

   // ✓ Good
   body: { api_key: '...', user_id: '...' }
   ```

5. **Don't forget to check discriminated union**
   ```typescript
   // ✗ Bad - errors might not exist
   if (result.valid) {
     console.log(result.errors);
   }

   // ✓ Good
   if (!result.valid) {
     console.log(result.errors);
   }
   ```

---

## Troubleshooting

### Issue: TypeScript error "Type 'unknown' is not assignable to type..."
**Solution:** Add generic type parameter: `request<T>(endpoint)`

### Issue: Function returns wrong type
**Solution:** Check generic type parameter matches actual response

### Issue: Backend doesn't receive field
**Solution:** Check field is snake_case in request body

### Issue: Error message is too long
**Solution:** It includes HTTP status - this is expected

### Issue: Streaming endpoint hangs
**Solution:** Ensure you're using custom fetch, not `request<T>()`

### Issue: Type checking doesn't catch error
**Solution:** Enable TypeScript strict mode

---

## When to Update Documentation

Update these documents when:
- Adding new endpoint patterns
- Discovering better practices
- Backend API changes
- Response structure changes
- Type definitions change

Keep these files in sync with actual code in:
- `frontend/src/lib/api/ai/ai.ts`
- `frontend/src/lib/api/ai/types.ts`
- `frontend/src/lib/api/base.ts`

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-08 | Initial comprehensive analysis |

---

## How to Use This Package

### For Beginners
1. Start with **API_DOCUMENTATION_INDEX.md**
2. Read **API_PATTERNS_ANALYSIS.md** sections 1-3
3. Practice with **API_TEMPLATE_CUSTOM_AGENTS.ts**
4. Use **API_CHEATSHEET.md** as reference

### For Quick Reference
1. Use **API_CHEATSHEET.md** for patterns
2. Use **API_VISUAL_GUIDE.md** for diagrams
3. Reference **API_TEMPLATE_CUSTOM_AGENTS.ts** for code

### For Implementation
1. Check **API_DOCUMENTATION_INDEX.md** implementation checklist
2. Copy from **API_TEMPLATE_CUSTOM_AGENTS.ts**
3. Verify against **API_PATTERNS_ANALYSIS.md** patterns
4. Test using **API_CHEATSHEET.md** common use cases

### For Debugging
1. Check error handling in **API_PATTERNS_ANALYSIS.md** section 2
2. Reference error format in **API_CHEATSHEET.md**
3. Use decision tree in **API_VISUAL_GUIDE.md**

---

## Additional Resources

**Source Code:**
- `frontend/src/lib/api/ai/ai.ts` - Current implementations
- `frontend/src/lib/api/ai/types.ts` - Type definitions
- `frontend/src/lib/api/base.ts` - Base request wrapper

**Related Documentation:**
- CLAUDE.md - Project conventions and standards
- Backend Go API documentation

---

## Contact & Questions

For questions about API patterns:
1. Check **API_CHEATSHEET.md** first
2. Search **API_PATTERNS_ANALYSIS.md**
3. Review **API_TEMPLATE_CUSTOM_AGENTS.ts** examples

---

**Last Updated:** 2026-01-08
**Status:** Complete and ready for use
**Confidence Level:** High - based on production code analysis

