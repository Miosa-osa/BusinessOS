# Visual Transcript Fix - Test Guide

## ✅ What Was Fixed

**Problem**: You could HEAR OSA's responses (audio was working), but couldn't SEE transcripts visually.

**Root Cause**: Backend was publishing OSA's transcript but NOT your transcript.

**Fix Applied**: Added code to publish user transcript with `speaker: "user"` after STT completes.

**File Changed**: `desktop/backend-go/internal/livekit/agent.go` lines 513-528

---

## 🎯 What You Should See Now

### Before Fix:
```
Console: ✅ Audio plays
Screen:  ❌ No visual feedback of what you said
         ❌ No visual feedback of what OSA said
```

### After Fix:
```
Console: ✅ Audio plays
Screen:  ✅ YOU see: "Not much. Can you hear me?" (what you said)
         ✅ OSA sees: "I'm OSA. Yeah, I can hear you." (what OSA said)
```

---

## 🧪 How to Test

### Step 1: Refresh Browser
```
Cmd + R (Mac) or Ctrl + R (Windows)
```

### Step 2: Open Browser Console
```
Cmd + Option + J (Mac) or F12 (Windows)
```

### Step 3: Start Backend Log Monitoring
In a terminal:
```bash
tail -f /private/tmp/backend-go-latest.log | grep -E "transcript|Transcript|LiveKit Agent"
```

### Step 4: Click Voice Orb

**Expected Browser Console Logs:**
```
[LiveKit] Connected to room: osa-voice-XXXXX
[LiveKit] 📦 Data received from: agent-osa size: XXXXX bytes
[LiveKit] ✅ Detected audio data (MP3 header found)
[LiveKit] 🔊 Playing agent audio response
[LiveKit] ▶️ Audio playback started
[LiveKit] 💬 Transcript: agent - [greeting text]
```

**Expected Backend Logs:**
```
[LiveKit Agent] Publishing user transcript... text="..."
[LiveKit Agent] ✅ User transcript published successfully
[LiveKit Agent] Publishing transcript data...
[LiveKit Agent] ✅ Transcript published successfully
```

**Expected On Screen:**
- LiveCaptions component should show OSA's greeting
- Visual confirmation that voice is active

### Step 5: Speak to OSA

Say clearly: **"Hello OSA, what's on my screen?"**

**Expected Browser Console:**
```
[LiveKit] 📦 Data received (user transcript)
[LiveKit] 💬 Transcript: user - Hello OSA, what's on my screen?
[LiveKit] 📦 Data received (audio)
[LiveKit] 🔊 Playing agent audio response
[LiveKit] 📦 Data received (agent transcript)
[LiveKit] 💬 Transcript: agent - [OSA's response]
```

**Expected Backend Logs:**
```
[LiveKit Agent] 🎤 Speech detected - starting recording rms=XXXX
[LiveKit Agent] 🔇 Silence detected after speech - processing utterance
[VoiceAgent] ✅ STT complete transcript: "Hello OSA, what's on my screen?"
[LiveKit Agent] Publishing user transcript... text="Hello OSA, what's on my screen?"
[LiveKit Agent] ✅ User transcript published successfully
[VoiceAgent] ✅ LLM complete response: "[OSA's response]"
[VoiceAgent] ✅ TTS complete audio_size: XXXXX
[LiveKit Agent] 🔊 Publishing TTS audio to room
[LiveKit Agent] Publishing transcript data...
[LiveKit Agent] ✅ Transcript published successfully
```

**Expected On Screen:**
```
LiveCaptions Component:

┌──────────────────────────────────────────────┐
│ 👤 YOU: Hello OSA, what's on my screen?      │
│                                              │
│ 🤖 OSA: You have the 3D desktop open in     │
│         orbit mode. [...]                    │
└──────────────────────────────────────────────┘
```

**Expected Audio:**
- You HEAR OSA's voice saying the response

---

## 🔍 Debugging If Transcripts Still Don't Show

### Problem 1: No User Transcript in Console

**Check backend logs for:**
```bash
tail -100 /private/tmp/backend-go-latest.log | grep "Publishing user transcript"
```

**If empty**:
- Backend not detecting speech properly
- Check VAD logs: `grep "Speech detected" /private/tmp/backend-go-latest.log`

**If present**:
- Backend sending it, frontend not receiving
- Check browser console for data received with `speaker: "user"`

### Problem 2: Transcripts in Console But Not On Screen

**Check**:
1. Is `LiveCaptions` component rendered?
   - Open React DevTools
   - Search for `LiveCaptions` component

2. Are `userMessage` and `osaMessage` props being set?
   - Check Desktop3D.svelte state variables (lines 73-74)
   - Check onTranscript callback (lines 183-190)

**Fix**:
- Verify `livekitVoice.onTranscript()` callback is firing
- Check console for: `console.log` statements in transcript handling

### Problem 3: Greeting Not Showing

**Check backend logs:**
```bash
grep "Greeting" /private/tmp/backend-go-latest.log | tail -20
```

**Should see:**
```
[LiveKit Agent] 👥 Participant connected
[LiveKit Agent] 🎙️  Sending greeting to new participant...
[VoiceAgent] Greeting generated greeting="..."
[LiveKit Agent] 🔊 Publishing greeting audio to participant
[LiveKit Agent] ✅ Greeting audio published to participant
[LiveKit Agent] ✅ Greeting transcript published to participant
```

**If missing**:
- Participant connected but no greeting sent
- Check ElevenLabs/Groq API keys in .env

---

## 📊 Complete Success Looks Like

### Browser Console:
```
[LiveKit] Connected ✅
[LiveKit] 📦 Data: agent transcript ✅
[LiveKit] 📦 Data: user transcript ✅
[LiveKit] 📦 Data: audio ✅
[LiveKit] Audio played ✅
```

### Backend Logs:
```
Speech detected ✅
STT complete ✅
User transcript published ✅
LLM complete ✅
TTS complete ✅
Agent transcript published ✅
Audio published ✅
```

### On Screen (LiveCaptions):
```
YOUR MESSAGE: [what you said] ✅
OSA'S RESPONSE: [what OSA said] ✅
```

### Audio:
```
You HEAR OSA speaking ✅
```

---

## 🎉 Summary

**What Changed**:
- Added 15 lines of code to publish user transcript
- Rebuilt backend
- Restarted server

**What Now Works**:
- ✅ User transcript sent to frontend with `speaker: "user"`
- ✅ OSA transcript sent to frontend with `speaker: "agent"` (was already working)
- ✅ LiveCaptions component should display BOTH transcripts
- ✅ Complete visual feedback of conversation

**Test Now**:
1. Refresh browser
2. Click voice orb
3. Speak: "Hello OSA, what's on my screen?"
4. LOOK at screen for visual transcripts
5. LISTEN for OSA's audio response

Both should work now!
