# Agent Form Validation System

## Overview

The agent creation/editing forms now include comprehensive client-side validation with real-time feedback and character counters. This ensures data quality before submission and provides a better user experience.

## Features

### 1. Real-Time Validation
- Fields validate as the user types
- Instant feedback on errors
- Visual indicators (red borders, error messages)

### 2. Character Counters
All text fields display character counts with color-coded status:
- **Green/Gray**: Within limits
- **Orange**: Near limit (80% or more)
- **Red**: Over limit

### 3. Validation Summary
- Shows all validation errors at the top of the form
- Can be dismissed
- Updates as errors are fixed

### 4. Field-Specific Validation

#### Name (ID)
- **Required**: Yes
- **Min Length**: 2 characters
- **Max Length**: 50 characters
- **Pattern**: Lowercase alphanumeric with hyphens only (`[a-z0-9-]+`)
- **Examples**:
  - ✅ `my-agent`
  - ✅ `code-helper-v2`
  - ❌ `MyAgent` (uppercase)
  - ❌ `my_agent` (underscore)

#### Display Name
- **Required**: Yes
- **Min Length**: 2 characters
- **Max Length**: 100 characters
- **Pattern**: Any characters allowed
- **Character Counter**: Shows `X / 100 characters`

#### System Prompt
- **Required**: Yes
- **Min Length**: 10 characters
- **Max Length**: 5000 characters
- **Character Counter**: Shows `X / 5000 characters` with color coding
- **Note**: Previous limit was 10,000; reduced to 5,000 for better performance

#### Description
- **Required**: No
- **Max Length**: 500 characters
- **Character Counter**: Shows `X / 500 characters`

#### Welcome Message
- **Required**: No
- **Max Length**: 2000 characters
- **Character Counter**: Shows `X / 2000 characters` with remaining count when near limit

#### Suggested Prompts
- **Max Count**: 10 prompts
- **Per Prompt Max Length**: 500 characters
- **Validation**:
  - Cannot be empty (whitespace only)
  - Each prompt shows character count
  - "Add" button disabled when limit reached
  - Individual prompts show character counters

#### Category
- **Required**: No
- **Allowed Values**:
  - general
  - coding
  - writing
  - analysis
  - research
  - support
  - sales
  - marketing
  - specialist
  - productivity
  - creative
  - technical
  - custom

#### Temperature
- **Min**: 0.0
- **Max**: 2.0
- **Step**: 0.01
- **Default**: 0.7
- **Note**: Extended range from previous 0.0-1.0 to 0.0-2.0

#### Max Tokens
- **Min**: 100
- **Max**: 32,000
- **Default**: 4000
- **Step**: 100

#### Avatar URL
- **Required**: No
- **Validation**: Must be valid URL format
- **Preview**: Shows image preview if URL is valid

## Implementation Details

### Files Created/Modified

1. **New File**: `frontend/src/lib/utils/agentValidation.ts`
   - Contains validation logic
   - Exports validation functions and constants
   - Type-safe with TypeScript

2. **Modified**: `frontend/src/lib/components/agents/AgentBuilder.svelte`
   - Integrated validation utility
   - Added real-time validation
   - Added character counters
   - Enhanced error display

3. **Test File**: `frontend/src/lib/utils/agentValidation.test.ts`
   - Comprehensive unit tests
   - Covers all validation scenarios
   - Ensures validation logic correctness

### Key Functions

#### `validateAgentForm(agent: Partial<CustomAgent>): ValidationResult`
Main validation function that checks all fields and returns validation result with errors.

```typescript
const result = validateAgentForm(agentData);
if (!result.valid) {
  // Handle errors
  result.errors.forEach(error => {
    console.log(`${error.field}: ${error.message}`);
  });
}
```

#### `getCharacterCountStatus(current: number, max: number)`
Helper function for character count display with color coding.

```typescript
const status = getCharacterCountStatus(text.length, 500);
// Returns: { current, max, percentage, remaining, isNearLimit, isOverLimit, statusClass }
```

#### `validateField(field: keyof CustomAgent, value: any): string | null`
Validates a single field and returns error message or null.

```typescript
const error = validateField('temperature', 2.5);
if (error) {
  console.log(error); // "Temperature must be between 0.0 and 2.0"
}
```

### Constants

```typescript
export const VALIDATION_LIMITS = {
  NAME_MIN: 2,
  NAME_MAX: 50,
  DISPLAY_NAME_MIN: 2,
  DISPLAY_NAME_MAX: 100,
  SYSTEM_PROMPT_MIN: 10,
  SYSTEM_PROMPT_MAX: 5000,
  WELCOME_MESSAGE_MAX: 2000,
  DESCRIPTION_MAX: 500,
  SUGGESTED_PROMPTS_MAX: 10,
  SUGGESTED_PROMPT_MAX: 500,
  TEMPERATURE_MIN: 0.0,
  TEMPERATURE_MAX: 2.0,
  MAX_TOKENS_MIN: 100,
  MAX_TOKENS_MAX: 32000
};

export const ALLOWED_CATEGORIES = [
  'general', 'coding', 'writing', 'analysis',
  'research', 'support', 'sales', 'marketing',
  'specialist', 'productivity', 'creative',
  'technical', 'custom'
];
```

## User Experience

### Visual Feedback

1. **Input Borders**:
   - Default: Gray border
   - Error: Red border
   - Focus: Blue ring

2. **Character Counters**:
   - Normal: Gray text
   - Near limit (80%+): Orange text + remaining count
   - Over limit: Red text + "X over limit!" warning

3. **Error Messages**:
   - Displayed below field in red
   - Clear, actionable text
   - Appears immediately on validation

4. **Validation Summary**:
   - Red banner at top when errors exist
   - Lists all errors with field names
   - Dismissible
   - Auto-scrolls to first error on submit

### Form Behavior

1. **Prevent Invalid Submission**:
   - Submit button validates before calling API
   - Shows validation summary if errors
   - Scrolls to first error field
   - No API call until valid

2. **Real-Time Feedback**:
   - Validates on input/change
   - Updates error state immediately
   - Removes error when fixed

3. **Smart Disabling**:
   - "Add prompt" button disabled when max reached
   - Fields show "disabled" state appropriately

## Testing

### Run Tests
```bash
cd frontend
npm test -- agentValidation.test.ts
```

### Test Coverage
- Name validation (empty, too short, too long, invalid pattern)
- Display name validation
- System prompt validation
- Temperature validation (range)
- Max tokens validation (range)
- Suggested prompts validation (count, length, empty)
- Category validation (allowed values)
- Complete agent validation
- Multiple errors collection

## Future Enhancements

### Possible Additions
1. **Async Validation**: Check name uniqueness against backend
2. **Debounced Validation**: Reduce validation calls while typing
3. **Field Dependencies**: Validate based on other field values
4. **Custom Error Messages**: User-configurable validation messages
5. **Import/Export Validation**: Validate imported agent data
6. **Batch Validation**: Validate multiple agents at once

### Performance Optimizations
1. Memoize validation results
2. Debounce real-time validation (currently immediate)
3. Lazy load validation logic for large forms

## Accessibility

- All errors announced via `aria-live` regions
- Error messages associated with fields via `aria-describedby`
- Form validation summary has `role="alert"`
- Clear focus indicators
- Keyboard navigation preserved

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- ES2020+ JavaScript features
- Uses native HTML5 validation attributes where appropriate

## Related Files

- `frontend/src/lib/api/ai/types.ts` - Type definitions
- `frontend/src/lib/components/agents/AgentBuilder.svelte` - Main form component
- `frontend/src/routes/(app)/agents/new/+page.svelte` - Create agent page
- `frontend/src/routes/(app)/agents/[id]/edit/+page.svelte` - Edit agent page

## API Alignment

This client-side validation mirrors backend validation rules. However, backend should still validate for security:

- Never trust client-side validation alone
- Backend must enforce same rules
- Client validation improves UX, backend validation ensures security

## Known Issues

None currently. If you find validation bugs, please report them with:
1. Field being validated
2. Input value
3. Expected behavior
4. Actual behavior

## Changelog

### 2026-01-11
- Initial implementation of validation system
- Added character counters to all text fields
- Created validation utility module
- Added comprehensive unit tests
- Updated form UI with real-time feedback
- Extended temperature range to 0.0-2.0
- Reduced system prompt max from 10,000 to 5,000 characters
