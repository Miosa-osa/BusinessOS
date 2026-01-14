# 🎤 Voice Commands Reference

**Last Updated**: January 14, 2026
**Status**: Production Ready

---

## 📊 All Available Commands

### Camera Control
| Command | Action | Details |
|---------|--------|---------|
| **"zoom in"** | Camera moves closer | True zoom - moves camera toward scene |
| **"zoom out"** | Camera moves farther | True zoom - moves camera away from scene |
| **"reset zoom"** | Reset camera distance | Returns to default camera position |

### Orb Control
| Command | Action | Details |
|---------|--------|---------|
| **"expand"** / **"expand orb"** | Increase sphere radius | Makes orb bigger, windows spread apart |
| **"contract"** / **"contract orb"** | Decrease sphere radius | Makes orb smaller, windows closer together |
| **"shrink"** / **"shrink orb"** | Same as contract | Alternative phrasing |

### Module Control
| Command | Action | Modules |
|---------|--------|---------|
| **"open [module]"** | Opens module | terminal, chat, tasks, projects, etc. |
| **"close [module]"** | Closes module | Any open module |
| **"focus [module]"** | Focuses existing module | Alternative to "open" |
| **"show [module]"** | Opens/focuses module | Alternative phrasing |

**Available Modules**:
- dashboard
- chat
- tasks
- projects
- team
- clients
- tables
- communication
- pages
- nodes
- daily
- settings
- terminal
- help
- agents
- crm
- integrations
- notifications
- profile

### Window Navigation
| Command | Action | Details |
|---------|--------|---------|
| **"next"** / **"next window"** | Focus next window | Cycles through open windows |
| **"previous"** / **"previous window"** | Focus previous window | Cycles backward |
| **"unfocus"** | Exit focus mode | Returns to orb view, shows all |
| **"back to orb"** | Same as unfocus | Alternative phrasing |

### Window Resizing
| Command | Action | Details |
|---------|--------|---------|
| **"make wider"** | Increase width | Expands focused window horizontally |
| **"make narrower"** | Decrease width | Contracts focused window horizontally |
| **"make taller"** | Increase height | Expands focused window vertically |
| **"make shorter"** | Decrease height | Contracts focused window vertically |

### View Control
| Command | Action | Details |
|---------|--------|---------|
| **"switch to grid"** | Grid view | 4-column flat layout |
| **"switch to orb"** | Orb view | 3D geodesic sphere layout |
| **"toggle auto rotate"** | Start/stop rotation | Orb rotates automatically |

### Layout Management
| Command | Action | Details |
|---------|--------|---------|
| **"manage layouts"** | Open layout manager | Modal for saving/loading layouts |
| **"enter edit mode"** | Start editing layout | Drag and position windows |
| **"exit edit mode"** | Stop editing | Lock window positions |
| **"save layout as [name]"** | Save current layout | Custom name required |
| **"load layout [name]"** | Load saved layout | Name must match existing |

---

## 🎯 Command Categories by Use Case

### Quick Window Access
```
"open terminal"
"open chat"
"close terminal"
"next window"
"previous window"
```

### Camera Positioning
```
"zoom in"         → Camera closer
"zoom out"        → Camera farther
"reset zoom"      → Default position
```

### Orb Sizing
```
"expand"          → Bigger orb
"contract"        → Smaller orb
"shrink"          → Smaller orb
```

### Window Management
```
"unfocus"         → Show all windows
"make wider"      → Resize current window
"make taller"     → Resize current window
```

### View Switching
```
"switch to grid"  → Flat 4-column layout
"switch to orb"   → 3D sphere layout
"toggle auto rotate" → Spin/stop
```

---

## 💡 Pro Tips

### Keep Commands Short (Under 5 Words)
- ✅ "open chat"
- ✅ "zoom out"
- ✅ "expand"
- ❌ "can you please open the chat for me"
- ❌ "switch me from terminal to chat"

### Be Direct
- ✅ "make wider"
- ❌ "can you make this wider"
- ✅ "contract"
- ❌ "could you contract the orb"

### Use Exact Phrases
These patterns are recognized:
- "open [module]" ✅
- "open the [module]" ❌ (extra word)
- "zoom out" ✅
- "zoom me out" ❌ (not in pattern)

---

## 🐛 Troubleshooting

### Command Not Working?

**1. Check Console Logs (F12)**
```
[Voice] 🎤 HEARD: "your command"
[Parser] ✅ Matched: CATEGORY
[Voice] ✅ SUCCESS: command_type
```

**2. Common Issues**:
- **Too many words** (>5) → Routes to AI conversation
- **Conversational phrases** ("can you", "please") → Routes to AI
- **Wrong module name** → Check spelling
- **Module not open** → Some commands require focused window

### Zoom/Expand Confusion?

| Say This | Effect |
|----------|--------|
| **"zoom in"** | Camera moves CLOSER to scene |
| **"zoom out"** | Camera moves FARTHER from scene |
| **"expand"** | Orb gets BIGGER (radius increases) |
| **"contract"** | Orb gets SMALLER (radius decreases) |

### Module Not Opening?

Check console:
```
[Parser] Checking modules for: open terminal
[Parser] ✅ Module matched: "terminal"
[Voice] ⚙️ EXECUTING: focus_module
```

If you see "CONVERSATION" instead of "MODULE", the command wasn't recognized.

---

## 📋 Quick Reference Card

### Most Used Commands
```
Modules:     "open [name]" | "close [name]"
Navigation:  "next" | "previous" | "unfocus"
Camera:      "zoom in" | "zoom out"
Orb:         "expand" | "contract"
Windows:     "make wider" | "make taller"
Views:       "switch to grid" | "switch to orb"
```

### Keyboard Shortcuts (Fallback)
If voice isn't working:
- `Tab`: Next window
- `Shift+Tab`: Previous window
- `Esc`: Unfocus
- `E`: Toggle edit mode

---

## ✅ System Status

- ✅ Voice Recognition: Deepgram Nova-2 with keyword boosting
- ✅ TTS: ElevenLabs OSA voice
- ✅ AI Conversation: Groq llama-3.3-70b
- ✅ Command Execution: Full error handling
- ✅ Audio Queue: Fixed with heartbeat mechanism
- ✅ Zoom Control: Camera + Orb separated
- ✅ Module Commands: All 20+ modules supported
- ✅ Logging: Comprehensive debug output

---

**Ready to use! Open browser console (F12) to see command execution flow.** 🎤
