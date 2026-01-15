# BuildProgress Component

A production-ready, real-time build progress component for OSA app generation with SSE streaming, auto-reconnection, collapsible phase logs, and live log updates.

## Features

- Real-time progress updates via Server-Sent Events (SSE)
- Animated progress bar with shimmer effect
- Visual phase indicators with icons (Planning, Building, Testing, Deploying)
- Collapsible log sections per build phase
- Terminal-style scrollable log viewer with syntax highlighting
- Auto-scroll to latest log entries
- Automatic reconnection with configurable attempts
- Estimated time remaining display
- Deployment URL display on success
- Retry button on build failure
- Dark mode optimized (Tailwind CSS)
- Memory-efficient log management (max 500 entries)
- Full TypeScript support with exported types

## Installation

The component uses `lucide-svelte` for icons. Ensure it's installed:

```bash
npm install lucide-svelte
```

## Usage

### Basic Example

```svelte
<script lang="ts">
  import BuildProgress from '$lib/components/osa/BuildProgress.svelte';
  import type { BuildResult } from '$lib/components/osa/BuildProgress.svelte';

  let buildId = 'build-abc123xyz';

  function handleComplete(result: BuildResult) {
    console.log('Build completed!', result);
    if (result.deploymentUrl) {
      console.log('App available at:', result.deploymentUrl);
    }
  }

  function handleError(error: Error) {
    console.error('Build failed:', error.message);
  }
</script>

<BuildProgress
  {buildId}
  onComplete={handleComplete}
  onError={handleError}
/>
```

### In a Modal/Dialog

```svelte
<script lang="ts">
  import BuildProgress from '$lib/components/osa/BuildProgress.svelte';
  import type { BuildResult } from '$lib/components/osa/BuildProgress.svelte';
  import { Dialog } from 'bits-ui';

  let showProgress = $state(false);
  let currentBuildId = $state('');

  async function startBuild() {
    const response = await fetch('/api/osa/build', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ appName: 'My App' })
    });

    const { buildId } = await response.json();
    currentBuildId = buildId;
    showProgress = true;
  }

  function handleBuildComplete(result: BuildResult) {
    console.log('Build completed!', result);
    // Optionally close after delay
    setTimeout(() => showProgress = false, 3000);
  }

  function handleBuildError(error: Error) {
    console.error('Build failed:', error.message);
    // Keep modal open for retry
  }
</script>

<button onclick={startBuild}>Generate App</button>

<Dialog.Root open={showProgress}>
  <Dialog.Portal>
    <Dialog.Overlay class="fixed inset-0 bg-black/50" />
    <Dialog.Content class="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full max-w-3xl">
      <Dialog.Title class="sr-only">Build Progress</Dialog.Title>
      <BuildProgress
        buildId={currentBuildId}
        onComplete={handleBuildComplete}
        onError={handleBuildError}
      />
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
```

### With Custom Container

```svelte
<div class="max-w-4xl mx-auto p-6">
  <BuildProgress buildId="build-123" />
</div>
```

## Props

| Prop | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `buildId` | `string` | Yes | - | The unique build ID to track |
| `onComplete` | `(result: BuildResult) => void` | No | `() => {}` | Callback fired when build completes successfully |
| `onError` | `(error: Error) => void` | No | `() => {}` | Callback fired when build fails |

## Exported Types

```typescript
export interface BuildResult {
  buildId: string;
  status: 'completed' | 'failed';
  deploymentUrl?: string;
  duration?: number;  // in seconds
  error?: string;
}

export interface LogEntry {
  timestamp: Date;
  message: string;
  phase: string;
  level: 'info' | 'warn' | 'error' | 'success';
}
```

## SSE Event Format

The component expects Server-Sent Events from `/api/osa/builds/{buildId}/stream` with the following JSON structure:

```typescript
interface BuildEvent {
  progress: number;                    // 0-100
  phase: string;                       // "Planning" | "Building" | "Testing" | "Deploying"
  log?: string;                        // Optional log line
  status: "in_progress" | "completed" | "failed";
  error?: string;                      // Error message if status is "failed"
  deploymentUrl?: string;              // URL when status is "completed"
  estimatedTimeRemaining?: number;     // Seconds remaining (optional)
  logLevel?: "info" | "warn" | "error" | "success";  // Log level for highlighting
}
```

### Example SSE Messages

```
data: {"progress": 5, "phase": "Planning", "log": "Analyzing requirements...", "estimatedTimeRemaining": 120}

data: {"progress": 15, "phase": "Planning", "log": "Generating project structure..."}

data: {"progress": 25, "phase": "Building", "log": "$ npm install", "logLevel": "info"}

data: {"progress": 40, "phase": "Building", "log": "Installing dependencies..."}

data: {"progress": 60, "phase": "Testing", "log": "Running test suite..."}

data: {"progress": 75, "phase": "Testing", "log": "All tests passed!", "logLevel": "success"}

data: {"progress": 90, "phase": "Deploying", "log": "Deploying to E2B sandbox..."}

data: {"progress": 100, "phase": "Deploying", "status": "completed", "log": "Deployment successful!", "deploymentUrl": "https://app-123.e2b.dev"}
```

### Error Event Example

```
data: {"progress": 45, "phase": "Building", "status": "failed", "error": "Build failed: npm install failed with exit code 1", "logLevel": "error"}
```

## Backend Implementation Example

```go
// Example SSE handler in Go
func (h *Handler) StreamBuild(w http.ResponseWriter, r *http.Request) {
    buildID := chi.URLParam(r, "buildId")

    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")

    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    // Send progress updates
    phases := []struct {
        name     string
        progress int
        logs     []string
    }{
        {"Planning", 20, []string{"Analyzing requirements...", "Generating structure..."}},
        {"Building", 50, []string{"$ npm install", "Installing dependencies...", "Compiling TypeScript..."}},
        {"Testing", 75, []string{"Running tests...", "All tests passed!"}},
        {"Deploying", 100, []string{"Deploying to E2B...", "Deployment successful!"}},
    }

    for _, phase := range phases {
        for i, log := range phase.logs {
            progress := phase.progress - (len(phase.logs)-i-1)*5
            event := BuildEvent{
                Progress: progress,
                Phase:    phase.name,
                Log:      log,
                Status:   "in_progress",
            }

            if phase.name == "Deploying" && i == len(phase.logs)-1 {
                event.Status = "completed"
                event.DeploymentUrl = "https://app-" + buildID + ".e2b.dev"
            }

            data, _ := json.Marshal(event)
            fmt.Fprintf(w, "data: %s\n\n", data)
            flusher.Flush()
            time.Sleep(time.Second)
        }
    }
}
```

## Features in Detail

### Status States

The component has four status states with distinct visuals:

1. **Connecting** - Pulsing loader animation, blue theme
2. **Building** (In Progress) - Active progress bar with shimmer, current phase highlighted
3. **Success** - Green checkmark, deployment URL with "Open App" button
4. **Error** - Red alert with error message, "Retry Build" button

### Phase Indicators

Visual stepper showing build progress through 4 phases:
1. **Planning** (Zap icon) - Requirements analysis, structure generation
2. **Building** (Package icon) - Code generation, dependency installation
3. **Testing** (TestTube icon) - Test suite execution
4. **Deploying** (Rocket icon) - E2B sandbox deployment

Each phase shows:
- Pending state (gray, dimmed)
- Active state (blue, pulsing ring)
- Completed state (green checkmark)

### Collapsible Phase Logs

Each phase has its own collapsible log section:
- Automatically expands when phase becomes active
- Automatically collapses when phase completes
- Shows log count per phase
- Click to toggle expand/collapse

### Log Syntax Highlighting

The terminal automatically highlights:
- **Commands** (starting with `$`, `>`, `npm`, `yarn`, etc.) - Cyan
- **Success messages** (containing "success", "complete", "passed") - Green
- **Warnings** (containing "warn", "warning") - Yellow
- **Errors** (containing "error", "fail", "failed") - Red

### Estimated Time Remaining

When the backend provides `estimatedTimeRemaining` in events, it displays formatted as:
- "~2m 30s remaining" (for times > 1 minute)
- "~45s remaining" (for times < 1 minute)

### Auto-Reconnection

The component automatically reconnects if the SSE connection drops:
- Maximum 5 reconnection attempts
- 2-second delay between attempts
- Shows "Reconnecting..." indicator during reconnection
- Calls `onError` if max attempts exceeded

### Retry Functionality

On build failure:
- "Retry Build" button appears
- Clicking resets all state and reconnects
- Note: This reconnects to the same build stream (backend must support retry)

## Memory Management

- Keeps only the last 500 log entries in main log view
- Each phase maintains its own bounded log list
- Clean event listener cleanup on component destroy

## Responsive Design

- Optimized for desktop (1024px+)
- Stacks appropriately on smaller screens
- Terminal height: 256px (16rem)
- Collapsible sections: max-height 160px

## Browser Support

- Modern browsers with EventSource support
- Chrome/Edge 80+
- Firefox 75+
- Safari 13+

## Styling

The component uses Tailwind CSS with a dark theme optimized for BusinessOS:
- `bg-gray-900` base background
- `border-gray-800` borders
- Color-coded status indicators
- Custom shimmer animation for progress bar
- Custom scrollbar styling for terminal

## Notes

- The component automatically starts SSE connection on mount
- Connection and event listeners are cleaned up on destroy
- The retry button reconnects to the same build ID
- `onComplete` receives the full `BuildResult` object
- `onError` receives a standard `Error` object
