# 🎤 Test Your Voice System NOW

## ✅ Setup Complete!

I just added your Deepgram API key to `.env`. Everything is ready to test!

---

## 🚀 How to Test (3 Steps)

### Step 1: Restart Dev Server

**In your terminal, press `Ctrl+C` to stop the current server, then:**

```bash
cd /home/miosa/Desktop/BusinessOS/frontend
npm run dev
```

### Step 2: Open 3D Desktop

1. Open http://localhost:5173
2. Log in if needed
3. Click **"Enter 3D Desktop"**

### Step 3: Test Voice Commands

1. **Click the microphone button** (bottom-right, circular button above dock)
2. **Allow microphone access** when browser asks
3. **Wait for button to turn BLUE** (means it's listening)
4. **Speak clearly:** "Open chat"
5. **Watch what happens!** ✨

---

## 🔍 How to Know It's Working

### **Open Browser Console** (F12 → Console Tab)

You should see these logs:

```
✅ [ActiveListening] 🔑 Deepgram API key found, initializing...
✅ [ActiveListening] 🌐 Connecting to Deepgram WebSocket...
✅ [ActiveListening] ✅ Deepgram WebSocket connected!
✅ [ActiveListening] 📤 Sending audio chunk to Deepgram: { size: 1234 }
✅ [ActiveListening] 📝 Transcription received: { transcript: "open chat", isFinal: true }
✅ [Voice Debug] 📝 Transcript (final): "open chat"
✅ [Voice Debug] 🎯 Command detected: focus_module
✅ [Voice Debug] 🔊 OSA responding: "Opening chat"
```

### **What You Should See/Hear:**

1. **Button turns BLUE** (listening)
2. **Live Captions show your words** as you speak
3. **Chat window opens** (command executed)
4. **OSA speaks:** "Opening chat" 🔊 (ElevenLabs TTS)
5. **Button pulses PURPLE** when OSA is speaking

---

## 🎤 Voice Commands to Try

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

**Conversation (anything else):**
- "Hello OSA"
- "What can you help me with?"
- "Tell me about this project"

---

## 🐛 Troubleshooting

### Button doesn't turn blue

**Fix:**
- Hard refresh browser (Cmd/Ctrl + Shift + R)
- Check console for errors
- Make sure you allowed microphone permission

### No transcription appears

**Check console for:**
- "Deepgram WebSocket connected!" - If you see this, speak louder
- "WebSocket connection failed" - API key issue (very unlikely, I just added it)

### Commands don't execute

**Make sure you're saying the EXACT phrases** from the list above.

Check console for:
```
[Voice Debug] 🎯 Command detected: unknown
```

If it says "unknown", try: "Open chat" or "Switch to grid view"

### OSA doesn't speak back

**That's a different issue** (ElevenLabs TTS):
- Check backend has ElevenLabs key in `.env`
- Check backend terminal for errors

---

## 🎯 What's Happening Under the Hood

```
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  YOU SPEAK: "Open chat"                                 │
│       ↓                                                 │
│  Microphone captures audio (250ms chunks)               │
│       ↓                                                 │
│  Deepgram WebSocket (HEARS YOU)                         │
│       ↓                                                 │
│  Transcript: "open chat" (<300ms latency!)              │
│       ↓                                                 │
│  Voice Command Parser detects: focus_module             │
│       ↓                                                 │
│  Chat window opens                                      │
│       ↓                                                 │
│  OSA prepares response: "Opening chat"                  │
│       ↓                                                 │
│  ElevenLabs TTS (OSA SPEAKS)                            │
│       ↓                                                 │
│  You hear: "Opening chat" 🔊                            │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## ✅ Success Checklist

- [ ] Restarted dev server
- [ ] Opened 3D Desktop
- [ ] Clicked microphone button
- [ ] Button turned BLUE
- [ ] Console shows "Deepgram WebSocket connected!"
- [ ] Spoke: "Open chat"
- [ ] Saw transcript appear
- [ ] Chat window opened
- [ ] OSA spoke: "Opening chat"

**If all checked: YOU'RE DONE!** 🎉

---

## 💡 Pro Tips

1. **Speak clearly and at normal volume** - no need to shout
2. **Wait for button to turn blue** before speaking
3. **Watch the console** - it tells you everything that's happening
4. **Use the Debug Panel** (🐛 button top-right) to see real-time status

---

**Ready to test? Restart that dev server and try it!** 🚀
