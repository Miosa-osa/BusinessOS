# Agent Form Validation - Visual Examples

## Character Counter States

### Normal State (Within Limits)
```
Welcome Message
┌────────────────────────────────────────────────────┐
│ Hello! How can I help you today?                   │
└────────────────────────────────────────────────────┘
34 / 2000 characters
```

### Near Limit State (80%+)
```
Welcome Message
┌────────────────────────────────────────────────────┐
│ [Very long message approaching 2000 characters...] │
└────────────────────────────────────────────────────┘
🟠 1650 / 2000 characters (350 remaining)
```

### Over Limit State
```
Welcome Message
┌────────────────────────────────────────────────────┐
│ [Message exceeding the maximum character limit...] │
└────────────────────────────────────────────────────┘
🔴 2050 / 2000 characters (50 over limit!)
```

## Field Validation Examples

### Valid Name Field
```
Name (ID) *
┌────────────────────────────────────────────────────┐
│ my-awesome-agent                                    │
└────────────────────────────────────────────────────┘
16 / 50 characters - Lowercase letters, numbers, and hyphens only
```

### Invalid Name Field (Uppercase)
```
Name (ID) *
┌────────────────────────────────────────────────────┐
│ MyAwesomeAgent                                      │ ❌ Red border
└────────────────────────────────────────────────────┘
❌ Name must be lowercase alphanumeric with hyphens only
```

### Invalid Name Field (Too Short)
```
Name (ID) *
┌────────────────────────────────────────────────────┐
│ a                                                   │ ❌ Red border
└────────────────────────────────────────────────────┘
❌ Name must be at least 2 characters
```

## Suggested Prompts Counter

### Normal State
```
Suggested Prompts (Quick start options)                    3 / 10 prompts

┌────────────────────────────────────────────────────┐
│ Help me analyze data                            ✕ │
│ 21 / 500 characters                                │
└────────────────────────────────────────────────────┘
┌────────────────────────────────────────────────────┐
│ Explain a complex concept                       ✕ │
│ 26 / 500 characters                                │
└────────────────────────────────────────────────────┘
┌────────────────────────────────────────────────────┐
│ Write code for a specific task                  ✕ │
│ 30 / 500 characters                                │
└────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────┐
│ Type a prompt and press Enter                      │
└────────────────────────────────────────────────────┘
[Add]
```

### Maximum Reached
```
Suggested Prompts (Quick start options)                   🔴 10 / 10 prompts

[10 prompts listed above...]

┌────────────────────────────────────────────────────┐
│ Maximum prompts reached                             │ (Disabled)
└────────────────────────────────────────────────────┘
[Add] (Disabled)
```

## Temperature Slider

### Normal
```
Temperature                                              0.70

[━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━]
│
Precise (0.0)          Balanced (1.0)          Creative (2.0)

Higher values make output more random and creative. Range: 0.0 - 2.0
```

### Error State (Out of Range)
```
Temperature                                              2.50

[━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━]
│
Precise (0.0)          Balanced (1.0)          Creative (2.0)

❌ Temperature must be between 0.0 and 2.0
```

## Validation Summary Banner

### Multiple Errors
```
╔════════════════════════════════════════════════════════╗
║ ⚠️  Please fix 4 error(s) before saving:               ║
║                                                        ║
║ • Name: Name must be lowercase alphanumeric with      ║
║   hyphens only                                         ║
║ • Display Name: Display name must be at least 2       ║
║   characters                                           ║
║ • System Prompt: System prompt is required            ║
║ • Temperature: Temperature must be between 0.0 and    ║
║   2.0                                                  ║
╚════════════════════════════════════════════════════════╝
                                                      [✕]
```

## Real-Time Validation Flow

### Typing in Name Field
```
Step 1: User types "M"
Name (ID) *
┌────────────────────────────────────────────────────┐
│ M                                                   │ ❌ Red border
└────────────────────────────────────────────────────┘
❌ Name must be lowercase alphanumeric with hyphens only

Step 2: User deletes and types "my"
Name (ID) *
┌────────────────────────────────────────────────────┐
│ my                                                  │ ✅ Normal border
└────────────────────────────────────────────────────┘
2 / 50 characters - Lowercase letters, numbers, and hyphens only

Step 3: User continues typing "my-agent"
Name (ID) *
┌────────────────────────────────────────────────────┐
│ my-agent                                            │ ✅ Normal border
└────────────────────────────────────────────────────┘
8 / 50 characters - Lowercase letters, numbers, and hyphens only
```

## System Prompt with Character Counter

### Normal
```
System Prompt *                                    250 / 5000 characters

┌────────────────────────────────────────────────────┐
│ You are a helpful assistant that specializes in    │
│ data analysis. Provide clear, accurate insights    │
│ based on the data provided. Always explain your    │
│ reasoning and methodology.                         │
│                                                    │
│                                                    │
└────────────────────────────────────────────────────┘
```

### Near Limit (4500+ characters)
```
System Prompt *                        🟠 4650 / 5000 characters

┌────────────────────────────────────────────────────┐
│ [Long system prompt text approaching limit...]     │
│                                                    │
│                                                    │
└────────────────────────────────────────────────────┘
```

### Over Limit
```
System Prompt *                       🔴 5150 / 5000 characters

┌────────────────────────────────────────────────────┐
│ [System prompt text exceeding the limit...]        │ ❌ Red border
│                                                    │
│                                                    │
└────────────────────────────────────────────────────┘
❌ System prompt cannot exceed 5000 characters
```

## Avatar URL Validation

### Valid URL with Preview
```
Avatar URL (optional)
┌────────────────────────────────────────────┐ ┌────────┐
│ https://example.com/avatar.png             │ │  🖼️   │
└────────────────────────────────────────────┘ └────────┘
Provide a publicly accessible image URL
```

### Invalid URL
```
Avatar URL (optional)
┌────────────────────────────────────────────┐
│ not-a-valid-url                            │ ❌ Red border
└────────────────────────────────────────────┘
❌ Avatar must be a valid URL
```

### Failed to Load Image
```
Avatar URL (optional)
┌────────────────────────────────────────────┐
│ https://example.com/broken.png             │ ❌ Red border
└────────────────────────────────────────────┘
❌ Failed to load image from URL
```

## Category Dropdown

### Valid Selection
```
Category
┌────────────────────────────────────────────────────┐
│ General                                         ▼  │
└────────────────────────────────────────────────────┘

Options:
- General
- Coding
- Writing
- Analysis
- Research
- Support
- Sales
- Marketing
- Specialist
- Productivity
- Creative
- Technical
- Custom
```

## Complete Form States

### All Valid (Ready to Submit)
```
✅ All fields valid
✅ Character limits respected
✅ Proper formatting applied
✅ Optional fields filled correctly

[Cancel]  [Create Agent]  ← Enabled
```

### Has Errors (Cannot Submit)
```
❌ 3 validation errors
❌ Fix errors before submitting

[Cancel]  [Create Agent]  ← Shows validation summary on click
```

## Mobile Responsive Behavior

### Desktop (1440px+)
```
┌──────────────────────────────────────────────────────┐
│  Display Name                   Character Counter    │
│  ┌────────────────────────────────────────────────┐  │
│  │                                                │  │
│  └────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────┘
```

### Tablet (768px - 1024px)
```
┌────────────────────────────────────┐
│  Display Name                      │
│  Character Counter                 │
│  ┌──────────────────────────────┐  │
│  │                              │  │
│  └──────────────────────────────┘  │
└────────────────────────────────────┘
```

### Mobile (< 768px)
```
┌──────────────────────┐
│  Display Name        │
│  Character Counter   │
│  ┌────────────────┐  │
│  │                │  │
│  └────────────────┘  │
└──────────────────────┘
```

## Accessibility Features

### Screen Reader Announcements
```
Field: Name (ID)
Input: "MyAgent"
Announcement: "Error: Name must be lowercase alphanumeric with hyphens only"

Field: System Prompt
Input: [5500 characters]
Announcement: "Error: System prompt cannot exceed 5000 characters"
```

### Keyboard Navigation
```
Tab Order:
1. Display Name → 2. Name (ID) → 3. Description → 4. Avatar URL →
5. Category → 6. Welcome Message → 7. Suggested Prompts (Add) →
8. Model Preference → 9. Temperature → 10. Max Tokens →
11. System Prompt → 12. [Advanced sections...] → 13. Cancel → 14. Submit
```

### Focus Indicators
```
[Field with focus]
┌────────────────────────────────────────────────────┐
│ Text content                                       │ ← Blue ring
└────────────────────────────────────────────────────┘
                                                       2px blue outline
```

## Color Palette

### Character Count Colors
- **Normal**: `text-gray-500` / `#6B7280`
- **Near Limit**: `text-orange-600` / `#EA580C`
- **Over Limit**: `text-red-600` / `#DC2626`

### Border Colors
- **Normal**: `border-gray-300` / `#D1D5DB`
- **Focus**: `ring-blue-500` / `#3B82F6`
- **Error**: `border-red-500` / `#EF4444`

### Background Colors
- **Normal Input**: `bg-white` / `#FFFFFF`
- **Error Banner**: `bg-red-50` / `#FEF2F2`
- **Warning State**: `bg-orange-50` / `#FFF7ED`
