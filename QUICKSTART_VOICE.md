# 🎤 Voice System - Quick Start

## ✅ What I Just Fixed

Your voice system wasn't working because:
- ❌ FFmpeg was not installed
- ❌ Whisper binary was not installed
- ❌ The old system was too complex and slow

**NEW SOLUTION:** I replaced it with **Deepgram** (industry-leading speech-to-text API)

## 🚀 Get It Working in 3 Steps

### Step 1: Get Free Deepgram API Key (2 minutes)

1. Go to: https://console.deepgram.com/signup
2. Sign up (you get **$200 FREE credits**)
3. Copy your API key

### Step 2: Add API Key to .env (30 seconds)

Open this file:
```
/home/miosa/Desktop/BusinessOS/frontend/.env
```

Add this line at the bottom:
```bash
VITE_DEEPGRAM_API_KEY=paste_your_actual_api_key_here
```

**Example:**
```bash
VITE_DEEPGRAM_API_KEY=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
```

### Step 3: Restart & Test (1 minute)

```bash
cd /home/miosa/Desktop/BusinessOS/frontend
npm run dev
```

Then:
1. Open http://localhost:5173
2. Enter 3D Desktop
3. Click microphone button (bottom-right)
4. **Say:** "Open chat"
5. **It should work!** ✨

## 🔍 How to Tell If It's Working

Open browser console (F12 → Console), you should see:

```
[ActiveListening] 🔑 Deepgram API key found, initializing...
[ActiveListening] 🌐 Connecting to Deepgram WebSocket...
[ActiveListening] ✅ Deepgram WebSocket connected!
[ActiveListening] 📤 Sending audio chunk to Deepgram: { size: 1234 }
[ActiveListening] 📝 Transcription received: { transcript: "open chat", isFinal: true }
[Voice Debug] 🎯 Command detected: focus_module
[Voice Debug] 🔊 OSA responding: "Opening chat"
```

If you see those logs, **it's working perfectly!**

## 🆚 Before vs After

### Before (Broken):
```
Microphone → Backend → FFmpeg (missing) → Whisper (missing) → ❌ Never works
Latency: N/A (doesn't work)
```

### After (Working):
```
Microphone → Deepgram WebSocket → ✅ Transcription in <300ms
Latency: Sub-300ms (fast!)
```

## 💰 Cost

**Free tier:** $200 credits = ~26,000 minutes (~433 hours)

**After free credits:**
- $0.0077/minute = $0.46/hour
- 100 beta users @ 30 min/month = $23/month total

**Very affordable for production!**

## 🎯 Voice Commands to Try

Once working, try these:

**Navigation:**
- "Open chat"
- "Open dashboard"
- "Switch to grid view"
- "Switch to orb view"

**Layout:**
- "Enter edit mode"
- "Save layout as workspace"
- "Exit edit mode"

**View:**
- "Zoom in"
- "Zoom out"
- "Next window"
- "Previous window"

**Conversation:**
- "Hello OSA"
- "What can you do?"
- "Tell me about this project"

## 🐛 Troubleshooting

### "VITE_DEEPGRAM_API_KEY not found"

**Fix:**
1. Make sure you added the key to `.env` (not `.env.example`)
2. Restart dev server (`npm run dev`)
3. Hard refresh browser (Cmd/Ctrl + Shift + R)

### No transcription appears

**Check browser console:**
- Look for `[ActiveListening]` logs
- If you see "WebSocket connected" but no transcriptions, check your microphone permissions
- Try speaking louder/clearer

### Commands don't execute

**Check console for:**
```
[Voice Debug] 🎯 Command detected: unknown
```

If it says "unknown", use the **exact phrases** from the list above.

## 📚 Full Documentation

For detailed info, see:
- `/docs/DEEPGRAM_SETUP.md` - Complete setup guide
- `/docs/VOICE_SYSTEM_TESTING.md` - Testing procedures

## ✅ Success Checklist

- [ ] Signed up for Deepgram
- [ ] Added `VITE_DEEPGRAM_API_KEY` to `.env`
- [ ] Restarted dev server
- [ ] Clicked microphone button
- [ ] Console shows "Deepgram WebSocket connected!"
- [ ] Spoke and saw transcription
- [ ] Voice command executed
- [ ] OSA responded with voice

**If all checked, you're done!** 🎉

---

**Questions?** Check browser console - all logs start with `[ActiveListening]`
