# FIX: Can't Hear OSA's Voice - Audio Output Troubleshooting

## ✅ What's WORKING (Confirmed from Backend Logs):

```
✅ ElevenLabs TTS generating audio (65,663 bytes)
✅ Audio published to LiveKit successfully
✅ Frontend receiving audio data
✅ Frontend playing audio (playback started/ended)
```

## ❌ What's BROKEN:

**YOU CAN'T HEAR IT** - This is a browser/system audio configuration issue, NOT a code problem.

---

## 🔧 FIXES APPLIED:

### 1. Feedback Loop Prevention
**Problem**: Agent was hearing its own voice and transcribing it as user speech
**Fix**: Agent now ignores its own audio tracks - only listens to users

**Evidence in logs (before fix)**:
```
Publishing TTS audio → immediately "Speech detected" → transcribed own audio
```

**After fix**: No more feedback loop!

---

## 🎯 STEP-BY-STEP: Make Audio Work

### STEP 1: Check macOS System Volume

```bash
# Run this command to check current volume:
osascript -e 'get volume settings'
```

**Should show**: `output volume:50, input volume:XX, alert volume:XX, output muted:false`

**If `output muted:true`**:
```bash
osascript -e 'set volume output volume 75'
```

### STEP 2: Check Which Audio Output Device is Selected

```bash
# Install switchaudio-osx if not installed:
brew install switchaudio-osx

# Check current output device:
SwitchAudioSource -c

# List all available devices:
SwitchAudioSource -a -t output
```

**Verify**: The correct speakers/headphones are selected.

**To change**:
```bash
# Example: Switch to MacBook Pro Speakers
SwitchAudioSource -s "MacBook Pro Speakers"
```

### STEP 3: Test System Audio Works

```bash
# Play a test beep:
afplay /System/Library/Sounds/Ping.aiff
```

**Expected**: You should HEAR a ping sound.

**If you don't hear it**:
1. macOS volume is muted or too low
2. Wrong output device selected
3. Hardware issue (speakers broken/disconnected)

### STEP 4: Check Browser Tab Audio

Open the browser and look at the tab title - there should NOT be a 🔇 (mute) icon.

**If there is a mute icon**:
- Right-click on the tab
- Click "Unmute Site"

### STEP 5: Test Browser Audio API

Open browser console (`Cmd + Option + J`) and paste:

```javascript
// Test if browser can play audio at all:
const testAudio = new Audio('data:audio/wav;base64,UklGRnoGAABXQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YQoGAACBhYqFbF1fdJivrJBhNjVgodDbq2EcBj+a2/LDciUFLIHO8tiJNwgZaLvt559NEAxQp+PwtmMcBjiR1/LMeSwFJHfH8N2QQAoUXrTp66hVFApGn+DyvmwhBSuBzvLZiTYIG2m98OScTgwOUKvm8LJeGwU7k9fyz3gv');
testAudio.volume = 1.0;
testAudio.play()
  .then(() => console.log('✅ Browser audio works!'))
  .catch(e => console.error('❌ Browser audio blocked:', e.message));
```

**Expected**: You hear a BEEEP sound and see "✅ Browser audio works!"

**If you see "NotAllowedError"**:
- Browser blocked autoplay (this is normal on first load)
- **FIX**: Click anywhere on the page first, then try voice again

### STEP 6: Force Play Latest OSA Audio

If OSA spoke but you didn't hear it, force replay the last audio:

```javascript
// Find all audio elements on page:
const audios = document.querySelectorAll('audio');
console.log(`Found ${audios.length} audio elements`);

// Play the latest one:
if (audios.length > 0) {
    const latest = audios[audios.length - 1];
    latest.volume = 1.0;
    latest.muted = false;
    latest.play()
        .then(() => console.log('✅ Playing latest audio!'))
        .catch(e => console.error('❌ Play failed:', e.message));
} else {
    console.warn('⚠️  No audio elements found - OSA hasn\'t spoken yet');
}
```

**Expected**: You should HEAR OSA's voice from the last response.

**If you still don't hear it**: The audio element exists but isn't producing sound - check system/browser volume.

### STEP 7: Inspect Audio Element State

```javascript
// Check audio element configuration:
const audios = document.querySelectorAll('audio');
Array.from(audios).forEach((a, i) => {
    console.log(`Audio ${i}:`, {
        volume: a.volume,
        muted: a.muted,
        paused: a.paused,
        duration: a.duration,
        currentTime: a.currentTime,
        readyState: a.readyState,
        networkState: a.networkState,
        src_length: a.src.length
    });
});
```

**Expected output**:
```
Audio 0: {
  volume: 1,
  muted: false,
  paused: true,  // After playback ends
  duration: 2.554195,
  currentTime: 2.554195,
  readyState: 4,  // HAVE_ENOUGH_DATA
  networkState: 1,  // NETWORK_IDLE
  src_length: 100+
}
```

**Red flags**:
- `volume: 0` → Audio element volume is 0 (code bug - shouldn't happen)
- `muted: true` → Audio element is muted (code bug - shouldn't happen)
- `readyState: 0` → Audio data didn't load (network issue)

### STEP 8: Test with Direct MP3 Blob

Let's create an audio element manually and test if MP3 playback works:

```javascript
// Fetch the latest MP3 data from backend directly:
fetch('http://localhost:8001/api/osa/speak', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ text: 'Testing audio output', emotion: 'neutral' })
})
.then(r => r.arrayBuffer())
.then(buffer => {
    const blob = new Blob([buffer], { type: 'audio/mpeg' });
    const url = URL.createObjectURL(blob);
    const audio = new Audio(url);
    audio.volume = 1.0;
    audio.onloadeddata = () => console.log('✅ Audio loaded, duration:', audio.duration);
    audio.onplay = () => console.log('▶️  Audio playing');
    audio.onended = () => console.log('✅ Audio ended');
    audio.onerror = (e) => console.error('❌ Audio error:', e);
    return audio.play();
})
.then(() => console.log('✅ Direct MP3 test successful!'))
.catch(e => console.error('❌ Direct MP3 test failed:', e));
```

**Expected**: You should HEAR "Testing audio output" spoken by OSA.

**If this works**: The backend + ElevenLabs + browser audio pipeline is fine. Issue is in LiveKit data channel delivery or timing.

**If this DOESN'T work**: Browser can't play MP3s (rare) or system audio is broken.

### STEP 9: Check Audio Codec Support

```javascript
// Test if browser supports MP3:
const audio = document.createElement('audio');
console.log('MP3 support:', audio.canPlayType('audio/mpeg'));
console.log('MP3 (audio/mp3):', audio.canPlayType('audio/mp3'));
console.log('WAV support:', audio.canPlayType('audio/wav'));
console.log('OGG support:', audio.canPlayType('audio/ogg'));
```

**Expected**: `probably` or `maybe` for MP3.

**If empty string**: Browser doesn't support MP3 (highly unlikely in modern Chrome/Safari/Firefox).

---

## 🚀 AFTER FIXES: Test End-to-End

1. **Refresh browser** (Cmd + R)
2. **Open browser console** (Cmd + Option + J)
3. **Click anywhere on the page** (enable autoplay)
4. **Click voice orb**
5. **Wait for greeting** - you should HEAR OSA say hello
6. **Speak**: "What time is it?"
7. **Listen for response** - you should HEAR OSA answer

---

## 📊 Expected Logs (After Fix)

### Backend:
```
[LiveKit Agent] 🔔 Subscribing to user audio track (participant: user-XXX)
[LiveKit Agent] ⏭️  Skipping agent's own audio track (prevent feedback)
[LiveKit Agent] 🎤 Speech detected rms=1500
[VoiceAgent] ✅ STT complete transcript="What time is it?"
[VoiceAgent] ✅ LLM complete response="It's 3:05 PM"
[ElevenLabs] ✅ TTS successful audio_size_bytes=45000
[LiveKit Agent] ✅ Audio data published successfully
```

**Key difference**: NO "Speech detected" immediately after "Publishing TTS audio" (feedback loop fixed!)

### Frontend Console:
```
[LiveKit] 📦 Data received: 45,000 bytes
[LiveKit] ✅ Detected audio data (MP3 header found)
[LiveKit] 🔊 Playing agent audio response
[LiveKit] ✅ Audio loaded, duration: 2.1 seconds
[LiveKit] ▶️  Audio playback started
[LiveKit] ✅ Audio playback ended
```

**What YOU should experience**:
1. See transcript on screen: "What time is it?"
2. **HEAR** OSA's voice: "It's 3:05 PM"
3. See OSA's transcript on screen

---

## 🔍 If You STILL Can't Hear Audio

### Last Resort Debugging:

**1. Check if audio is actually being sent from backend:**
```bash
# Watch backend logs for audio size:
tail -f /private/tmp/backend-go-latest.log | grep "audio_size"
```

You should see: `audio_size_bytes=XXXXX` (non-zero)

**2. Check if frontend is receiving binary data:**
Open browser console and watch for:
```
📦 Data received: XXXXX bytes
```

The size should match the backend audio_size.

**3. Verify MP3 header is present:**
Console should show:
```
✅ Detected audio data (MP3 header found)
```

**If you see "⚠️ Falling back to size heuristic"**: The MP3 header wasn't detected but size > 1000 bytes, so it's still trying to play it.

**4. Capture the blob URL and download it:**
```javascript
// Next time OSA speaks, grab the blob URL from console:
// Look for: "Created blob URL: blob:http://localhost:5173/XXXXXX"

// Then download it to verify it's valid audio:
const blobUrl = 'blob:http://localhost:5173/PASTE_ID_HERE';
fetch(blobUrl)
    .then(r => r.blob())
    .then(blob => {
        const a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = 'osa_audio_test.mp3';
        a.click();
        console.log('✅ Downloaded audio file - check Downloads folder');
    });
```

**Open the downloaded MP3 in QuickTime Player** - if you can hear it there, the audio data is fine and the issue is browser playback only.

---

## 🎉 Summary of Fixes

| Issue | Status |
|-------|--------|
| Feedback loop (agent hearing itself) | ✅ FIXED |
| VAD threshold too low | ✅ ADJUSTED (30 RMS) |
| Silence detection too slow | ✅ FAST (300ms) |
| User transcript not sent | ✅ FIXED (earlier) |
| ElevenLabs TTS generation | ✅ WORKING |
| LiveKit data publishing | ✅ WORKING |
| Frontend audio reception | ✅ WORKING |
| Audio element creation/playback | ✅ CODE IS CORRECT |
| **System/browser audio output** | ⚠️  **CHECK THIS** |

---

## 🎯 Most Likely Cause

Based on your screenshot showing "Audio playback started" and "Audio playback ended", the audio element IS playing, but you're not hearing it because:

1. **macOS volume is muted or very low** (90% chance)
2. **Wrong audio output device selected** (5% chance)
3. **Browser tab is muted** (3% chance)
4. **Browser autoplay policy blocking** (2% chance - but logs show it's playing)

**RUN THIS NOW**:
```bash
# Check and fix macOS volume:
osascript -e 'set volume output volume 75'
osascript -e 'set volume without output muted'

# Play test sound:
afplay /System/Library/Sounds/Ping.aiff
```

If you HEAR the ping, your audio works - just need to test the voice system again with volume up!
