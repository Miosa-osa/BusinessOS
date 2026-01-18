# 🎤 Microphone Permission Fix

## The Problem:
macOS is blocking your browser from accessing the microphone.

## Quick Fix (5 steps):

### 1. Open System Settings
   - Click Apple menu () → System Settings
   - OR: Press ⌘ + Space and type "System Settings"

### 2. Go to Privacy & Security
   - In the left sidebar, click "Privacy & Security"
   - Scroll down to "Microphone" section
   - Click on "Microphone"

### 3. Enable Browser Permission
   - Look for your browser in the list:
     - ✓ Chrome
     - ✓ Arc
     - ✓ Brave  
     - ✓ Safari
   - Toggle it ON (green)
   - If your browser isn't in the list, you need to trigger it first

### 4. Check Microphone is Selected
   - Go to System Settings → Sound
   - Click "Input" tab
   - Make sure a microphone is selected (not "No Input Device")
   - If using external mic, make sure it's connected

### 5. Restart Browser
   - Close ALL browser windows completely
   - Quit the browser (⌘ + Q)
   - Open browser again
   - Go back to http://localhost:5173

## Then Try Again:
1. Navigate to voice interface
2. Click start voice session
3. Browser should now ask for microphone permission
4. Click "Allow"
5. Speak and test!

---

## Alternative: Trigger Permission Request

If your browser isn't showing in the Microphone list:

1. Open a new tab
2. Go to any website that uses microphone (like appear.in or discord.com)
3. Try to use voice there - this will trigger the permission request
4. Then go back to System Settings and enable it
5. Return to http://localhost:5173 and try again

---

## Verify It's Working:

Open Terminal and run:
```bash
# Check if any app has microphone access
sudo lsof | grep "AppleAudio"
```

Or test microphone directly in browser:
1. Go to: https://webcammictest.com/check-mic.html
2. Click "Check Mic"
3. If you see waveform moving = mic works
4. If not = still blocked in System Settings

