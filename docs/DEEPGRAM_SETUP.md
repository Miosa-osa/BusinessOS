# Deepgram Setup Guide

## 🎯 What is Deepgram?

Deepgram is the **industry-leading speech-to-text API** that provides real-time transcription with sub-300ms latency. We're using it to replace the broken local Whisper setup.

## Why Deepgram?

- ✅ **Sub-300ms latency** (perfect for voice commands)
- ✅ **$200 free credits** (plenty for beta testing)
- ✅ **Zero setup** (no FFmpeg, no Whisper binary, no models)
- ✅ **Production-ready** (used in telehealth, aviation)
- ✅ **Better accuracy** (~5% word error rate)
- ✅ **Easy to use** (just one API key)

## 🚀 Quick Setup (5 minutes)

### Step 1: Get Your Deepgram API Key

1. Go to https://console.deepgram.com/signup
2. Sign up for a free account
3. Verify your email
4. Go to https://console.deepgram.com/project/default/keys
5. Copy your API key (starts with something like `a1b2c3d4...`)

**Note:** You get **$200 in free credits** immediately!

### Step 2: Add API Key to Frontend .env

Open `/home/miosa/Desktop/BusinessOS/frontend/.env` and add:

```bash
# ===========================================
# DEEPGRAM (Speech-to-Text for Voice Commands)
# ===========================================
VITE_DEEPGRAM_API_KEY=paste_your_actual_api_key_here
```

**Example:**
```bash
VITE_DEEPGRAM_API_KEY=a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0
```

### Step 3: Restart Frontend Dev Server

```bash
cd /home/miosa/Desktop/BusinessOS/frontend
npm run dev
```

### Step 4: Test Voice Commands

1. Open http://localhost:5173
2. Enter 3D Desktop mode
3. Click the microphone button (bottom-right)
4. **Speak:** "Open chat"
5. **Watch the magic happen!** ✨

## 📊 What Changed?

### Before (Broken):
```
Microphone → WebM audio → Backend →
❌ FFmpeg (not installed) →
❌ Whisper (not installed) →
❌ Never works
```

### After (Working):
```
Microphone → WebSocket → Deepgram API →
✅ Transcription in <300ms →
✅ Voice commands execute
```

## 🔍 Debugging

### Issue: "VITE_DEEPGRAM_API_KEY not found"

**Solution:** Make sure you:
1. Added the key to `.env` (NOT `.env.example`)
2. Restarted the dev server (`npm run dev`)
3. Hard refresh browser (Cmd/Ctrl + Shift + R)

### Issue: "WebSocket connection failed"

**Solution:** Check your API key is valid:
1. Go to https://console.deepgram.com/project/default/keys
2. Make sure the key is active (not expired)
3. Try generating a new key

### Issue: "No transcription appears"

**Solution:** Check browser console for logs:
- Look for `[ActiveListening] ✅ Deepgram WebSocket connected!`
- Look for `[ActiveListening] 📝 Transcription received`
- If you see these, it's working!

## 💰 Cost Estimate

**Deepgram Pricing:** $0.0077/minute = $0.46/hour

**Example scenarios:**

| Scenario | Usage | Cost |
|----------|-------|------|
| Testing (10 hours) | 600 minutes | $4.62 |
| Beta user (30 min/month) | 30 minutes | $0.23 |
| 100 beta users (30 min each) | 3000 minutes | $23.10 |

**Your $200 credit covers:**
- ~26,000 minutes of transcription
- ~433 hours of voice usage
- Plenty for beta testing!

## 🎤 Voice Commands You Can Try

Once it's working, try these:

**Navigation:**
- "Open chat"
- "Open dashboard"
- "Switch to grid view"
- "Switch to orb view"

**Layout Management:**
- "Enter edit mode"
- "Save layout as workspace"
- "Exit edit mode"

**View Control:**
- "Zoom in"
- "Zoom out"
- "Next window"
- "Previous window"

**Conversation:**
- "Hello OSA"
- "What can you do?"
- "Tell me about this project"

## 🐛 Troubleshooting

### Console shows WebSocket errors

**Check:**
1. Is your API key correct?
2. Do you have internet connection?
3. Is your Deepgram account active?

### Console shows transcripts but commands don't execute

**Check:**
1. Look for `[Voice Debug] 🎯 Command detected:` in console
2. If it says "unknown", try the exact phrases above
3. OSA should speak a response if it worked

### No audio from OSA

**That's a different issue** - ElevenLabs TTS (separate from Deepgram STT):
1. Check backend has `ELEVENLABS_API_KEY` in `.env`
2. Check backend logs for ElevenLabs errors

## 📚 Additional Resources

- [Deepgram Documentation](https://developers.deepgram.com/)
- [Deepgram Pricing](https://deepgram.com/pricing)
- [Deepgram JavaScript SDK](https://github.com/deepgram/deepgram-js-sdk)
- [Live Transcription Guide](https://developers.deepgram.com/docs/getting-started-with-live-streaming-audio)

## ✅ Success Checklist

- [ ] Signed up for Deepgram account
- [ ] Copied API key
- [ ] Added `VITE_DEEPGRAM_API_KEY` to frontend `.env`
- [ ] Restarted dev server
- [ ] Clicked microphone button
- [ ] Saw "Deepgram WebSocket connected!" in console
- [ ] Spoke and saw transcription appear
- [ ] Voice command executed
- [ ] OSA responded with voice

If all checks pass, **you're good to go!** 🎉

---

**Questions?** Check the browser console - all logs start with `[ActiveListening]` for easy filtering.
