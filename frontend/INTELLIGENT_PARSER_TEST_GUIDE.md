# 🧠 Intelligent Voice Parser - Test Guide

**Status**: Phase 1 COMPLETE - Ready for Testing!
**Date**: January 14, 2026

---

## ✅ What Was Implemented

### NEW: 6-Layer Intelligence System

1. **LAYER 1: Wake Word Detection** ✅
   - Detects "OSA", "hey OSA", "ok OSA"
   - Strips wake word and continues parsing

2. **LAYER 2: Exact Pattern Matching** ✅
   - Tries all existing command patterns first
   - Highest confidence

3. **LAYER 3: Command Extraction** ✅
   - Extracts commands from conversational wrappers
   - "can you open terminal" → "open terminal"
   - "please show chat" → "show chat"

4. **LAYER 4: Fuzzy Module Detection** ✅
   - Detects module names even without action verbs
   - "terminal" → assumes "open terminal"
   - Confidence scoring (0.95 for exact, 0.7 for partial)

5. **LAYER 5: Help Intent** ✅
   - Short help queries (≤3 words)
   - "help", "commands", etc.

6. **LAYER 6: Conversation Routing** ✅
   - Only routes if > 7 words OR question WITHOUT module
   - Much less aggressive than before

---

## 🧪 Test Cases

Open browser console (F12) and test these commands:

### Test 1: Wake Word Support
```
Say: "OSA terminal"

Expected Logs:
[Parser] 🔍 ANALYZING: OSA terminal
[Parser] 👂 Wake word detected, stripped to: terminal
[Parser] 📝 Normalized: terminal
[Parser] 🔎 Module detection: { module: "terminal", confidence: 0.95 }
[Parser] ✅ FUZZY MODULE MATCH → open: terminal
[Voice] 📱 focus_module command for module: "terminal"
[Desktop3D Store] 🔵 openWindow() called for module: "terminal"

Expected Result: ✅ Terminal opens!
```

### Test 2: Conversational Wrapper Extraction
```
Say: "can you open the OSA terminal?"

Expected Logs:
[Parser] 🔍 ANALYZING: can you open the OSA terminal?
[Parser] 👂 Wake word detected, stripped to: can you open the terminal?
[Parser] 📝 Normalized: can you open the terminal?
[Parser] 🧠 Extracted command: { extracted: "open the terminal", confidence: 0.9 }
[Parser] ✅ EXTRACTED MATCH: focus_module
[Voice] 📱 focus_module command for module: "terminal"

Expected Result: ✅ Terminal opens!
```

### Test 3: Just Module Name (Fuzzy Match)
```
Say: "terminal"

Expected Logs:
[Parser] 🔍 ANALYZING: terminal
[Parser] 📝 Normalized: terminal
[Parser] 🔎 Module detection: { module: "terminal", confidence: 0.95 }
[Parser] ✅ FUZZY MODULE MATCH → open: terminal

Expected Result: ✅ Terminal opens!
```

### Test 4: Polite Natural Language
```
Say: "please show me the chat"

Expected Logs:
[Parser] 🔍 ANALYZING: please show me the chat
[Parser] 🧠 Extracted command: { extracted: "show me the chat", confidence: 0.9 }
[Parser] ✅ EXTRACTED MATCH: focus_module

Expected Result: ✅ Chat opens!
```

### Test 5: Still Routes Conversations Correctly
```
Say: "what can you help me with?"

Expected Logs:
[Parser] 🔍 ANALYZING: what can you help me with?
[Parser] 🤔 Routing decision: { wordCount: 5, isQuestion: true, isConversational: false, hasModule: false }
[Parser] 💬 ROUTING TO AI
[Voice] 💬 ROUTING TO AI: unknown

Expected Result: ✅ Routes to AI conversation
```

### Test 6: All Your Previous Commands Still Work
```
Say: "open chat"
Say: "expand"
Say: "zoom out"
Say: "next window"
Say: "close terminal"

Expected: ✅ All should work exactly as before
```

---

## 📊 Comparison: Before vs After

| Command | BEFORE (Broken) | AFTER (Intelligent) |
|---------|-----------------|---------------------|
| "OSA terminal" | ❌ UNKNOWN → AI (0 tokens) | ✅ Opens terminal (fuzzy match) |
| "can you open terminal" | ❌ UNKNOWN (conversational block) | ✅ Opens terminal (extracted) |
| "terminal" | ❌ UNKNOWN (no verb) | ✅ Opens terminal (fuzzy match) |
| "please show chat" | ❌ UNKNOWN (conversational block) | ✅ Opens chat (extracted) |
| "open chat" | ✅ Works | ✅ Still works (exact match) |
| "expand" | ✅ Works | ✅ Still works (exact match) |

---

## 🔍 How to Debug If Still Not Working

### Step 1: Check Console Logs

Look for the 6-layer flow:

```
[Parser] 🔍 ANALYZING: [your command]
[Parser] 👂 Wake word detected (if OSA used)
[Parser] 📝 Normalized: [normalized text]
[Parser] ✅ EXACT MATCH (Layer 2)
   OR
[Parser] 🧠 Extracted command (Layer 3)
[Parser] ✅ EXTRACTED MATCH
   OR
[Parser] 🔎 Module detection (Layer 4)
[Parser] ✅ FUZZY MODULE MATCH
```

### Step 2: Check Module List

If a module isn't being detected, check if it's in the list:
```typescript
const modules = [
    'dashboard', 'chat', 'tasks', 'projects', 'team', 'clients',
    'tables', 'communication', 'pages', 'nodes', 'daily',
    'settings', 'terminal', 'help', 'agents', 'crm',
    'integrations', 'knowledge-v2', 'notifications', 'profile',
    'voice-notes', 'usage'
];
```

### Step 3: Check Action Words

If a command is being flagged as conversational, check if the action verb is in the list:
```typescript
const actionWords = [
    'open', 'close', 'show', 'hide', 'zoom',
    'expand', 'contract', 'switch', 'go',
    'focus', 'load', 'save'
];
```

---

## 🎯 Success Criteria

### Must Pass All These:
- ✅ "OSA terminal" opens terminal
- ✅ "can you open terminal" opens terminal
- ✅ "terminal" opens terminal
- ✅ "please show me chat" opens chat
- ✅ Previous commands still work ("open chat", "expand", etc.)
- ✅ Conversations still route to AI correctly

---

## 🚨 Known Limitation: Backend AI Returns 0 Tokens

**Status**: Phase 2 (Not Yet Fixed)

**Symptom**: When commands route to AI conversation, backend returns:
```json
{"type":"done","data":{"output_tokens":0}}
```

**Impact**: Fuzzy commands that fail matching will route to AI but get no response.

**Workaround**: Use the intelligent parser! It should catch almost everything now.

**Next Step**: Investigate backend `chat_v2.go` handler (Phase 2)

---

## 📖 Quick Reference

### Natural Phrases Now Supported:
- "OSA [command]" → Strips "OSA" and executes
- "can you [command]" → Extracts and executes
- "please [command]" → Extracts and executes
- "could you [command]" → Extracts and executes
- "[module name only]" → Assumes "open [module]"

### Conversational Routing (Still Works):
- Questions with "?" → AI
- Long phrases (>7 words) WITHOUT module → AI
- Greetings without commands → AI

---

**Test now and report results!** 🎤
