# Voice Command Testing Guide

**Last Updated**: January 14, 2026

---

## 🧪 How to Test Voice Commands

### Open Browser Console
1. Navigate to 3D Desktop
2. Press `F12` to open DevTools
3. Go to Console tab

### Watch the Logs

When you speak, you'll see this flow:

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Voice] 🎤 HEARD: "zoom out"
[Parser] Original: zoom out
[Parser] Normalized: zoom out
[Parser] ✅ Matched: VIEW
[Voice] 🧠 PARSED: {
  "type": "zoom_out"
}
[Voice] 🔊 SPEAKING ACK: "Pulling back"
[Voice] ⚙️ EXECUTING: zoom_out
[Desktop3D Store] Adjusting sphere radius: 120 → 125
[Voice] ✅ SUCCESS: zoom_out
```

---

## 🔍 Diagnostic Checklist

### If Command Not Recognized:

**Check the logs for:**

1. **What was heard?**
   ```
   [Voice] 🎤 HEARD: "switch to terminal"
   ```
   - Is Deepgram transcribing correctly?
   - Are you saying "OSA" before commands? (Not needed!)

2. **How was it normalized?**
   ```
   [Parser] Normalized: "switch terminal"
   ```
   - Are filler words being removed correctly?

3. **Why didn't it match?**
   ```
   [Parser] ❌ No match - Word count: 3, Question: false, Conversational: false
   ```
   - Too many words? (>5 triggers conversation)
   - Contains conversational phrases?
   - Missing from patterns?

4. **Where did it route?**
   ```
   [Parser] → Routing to CONVERSATION
   ```
   - Or: `[Parser] → UNKNOWN command`

---

## 🎯 Test Commands by Category

### Navigation Commands
```
✅ "next window"
✅ "previous window"
✅ "next"
✅ "previous"
```

### Module Commands
```
✅ "open terminal"
✅ "open chat"
✅ "close terminal"
✅ "focus dashboard"
```

### Camera Commands
```
✅ "zoom in"
✅ "zoom out"
✅ "reset zoom"
✅ "toggle auto rotate"
```

### Window Commands
```
✅ "unfocus"
✅ "back to orb"
✅ "make wider"
✅ "make taller"
```

### View Commands
```
✅ "switch to grid"
✅ "switch to orb"
✅ "orb view"
✅ "grid view"
```

---

## 🐛 Common Issues

### Issue 1: "Switch from terminal to chat" Not Working

**Problem**: Too many words (6 words)
**What happens**: Routes to conversation AI instead of command

**Solutions**:
- Say: "open chat" (2 words) ✅
- Or: "close terminal" then "open chat"
- Or: "next window" to cycle

**Why**: Commands over 5 words are treated as conversations

---

### Issue 2: "Zoom out" Not Working

**Check console for:**
```
[Voice] ⚙️ EXECUTING: zoom_out
[Desktop3D Store] Adjusting sphere radius: 120 → 125
[Voice] ✅ SUCCESS: zoom_out
```

**If you see this**: Command executed successfully, camera should move

**If you see error**: Check error message

---

### Issue 3: Commands Going to AI Instead

**Example**: "Switch me from terminal to chat"

**What's happening**:
- 6 words detected
- Parser routes to conversation
- AI responds instead of executing

**Fix**: Use shorter commands (5 words or less)

---

## 📊 Testing Script (Copy to Console)

```javascript
// Test command parsing without speaking
import { voiceCommandParser } from '$lib/services/voiceCommands';

const testPhrases = [
  "zoom out",
  "open terminal",
  "switch to grid",
  "make wider",
  "next window",
  "switch me from terminal to chat" // Too long
];

testPhrases.forEach(phrase => {
  console.log('━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━');
  console.log('Testing:', phrase);
  const result = voiceCommandParser.parse(phrase);
  console.log('Result:', result);
});
```

---

## 🎤 Deepgram Transcription Issues

### If "OSA" is Transcribed Wrong:

**Check for:**
- "Elsa" → Should be "OSA"
- "Oza" → Should be "OSA"
- "Ossa" → Should be "OSA"

**Already Fixed**: Keyword boosting enabled
```typescript
keywords: ['OSA:2', 'BusinessOS:1.5']
```

**Still happening?**:
1. Check console for actual transcription
2. Report specific phrases

---

## 🔧 Quick Fixes

### Command Too Long?

❌ "Switch me from terminal to chat"
✅ "open chat"

❌ "Can you zoom out for me please"
✅ "zoom out"

### Command Not in List?

Check if pattern exists in `voiceCommands.ts`:
- Search for the command type
- Check if your phrasing matches patterns
- Add new pattern if needed

### Parser Routing Incorrectly?

Check `isConversational()` function:
- Might be detecting conversational markers
- Example: "can you", "please", "could you"

---

## 📝 Reporting Issues

When reporting command issues, provide:

1. **Exact phrase you said**
2. **Console log output** (the full flow)
3. **Expected behavior**
4. **Actual behavior**

Example:
```
Said: "zoom out"
Logs: [copy console output]
Expected: Camera zooms out
Actual: Nothing happened
```

---

## ✅ Success Indicators

You know it's working when you see:

```
[Voice] ✅ SUCCESS: zoom_out
[Desktop3D Store] Adjusting sphere radius: 120 → 125
```

And the camera actually moves.

---

**Enable verbose logging already active!** Just open console and speak commands.
