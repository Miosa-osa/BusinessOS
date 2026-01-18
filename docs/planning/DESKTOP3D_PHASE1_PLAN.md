# 3D Desktop - Phase 1 Implementation Plan

## 🎯 Goals

1. **Camera/Microphone Permissions** - Request access only when entering 3D Desktop mode
2. **Default Layout Preservation** - Keep current 5-ring geodesic layout as immutable default
3. **Custom Positioning** - Allow users to drag modules to custom positions
4. **Layout Persistence** - Save/load custom layouts to backend
5. **Layout Manager UI** - Simple interface to manage saved layouts

---

## 🏗️ Architecture Overview

### Permission System (3D Desktop Only)

```typescript
// Only active when user is in 3D Desktop route (/desktop3d or /window)
class Desktop3DPermissions {
  camera: MediaStream | null
  microphone: MediaStream | null

  // Request permissions when entering 3D Desktop
  async requestPermissions(): Promise<boolean>

  // Release when leaving 3D Desktop
  cleanup(): void

  // Check status
  hasCamera(): boolean
  hasMicrophone(): boolean
}
```

### Layout System

```typescript
// Layout types
type LayoutType = 'default' | 'custom'

interface Layout {
  id: string
  name: string
  type: LayoutType
  created_at: Date
  updated_at: Date
  is_active: boolean
  modules: ModulePosition[]
}

interface ModulePosition {
  module_id: ModuleId
  position: { x: number; y: number; z: number }
  rotation: { x: number; y: number; z: number }
  scale: number
}

// The DEFAULT layout (current 5-ring geodesic)
const DEFAULT_LAYOUT: Layout = {
  id: 'default',
  name: 'Default',
  type: 'default',
  // ... positions calculated by getRingLayout()
}
```

---

## 📋 Implementation Tasks

### Task 1: Camera/Microphone Permission System

**Files to create/modify:**
- `src/lib/services/desktop3dPermissions.ts` (new)
- `src/lib/stores/desktop3dStore.ts` (add permission state)
- `src/routes/window/+page.svelte` (trigger on mount/unmount)

**Implementation:**

#### 1.1 Create Permission Service

```typescript
// src/lib/services/desktop3dPermissions.ts

import { writable, get } from 'svelte/store';

export const cameraPermission = writable<'prompt' | 'granted' | 'denied'>('prompt');
export const microphonePermission = writable<'prompt' | 'granted' | 'denied'>('prompt');
export const cameraStream = writable<MediaStream | null>(null);
export const microphoneStream = writable<MediaStream | null>(null);

export class Desktop3DPermissions {
  private static instance: Desktop3DPermissions;

  static getInstance(): Desktop3DPermissions {
    if (!this.instance) {
      this.instance = new Desktop3DPermissions();
    }
    return this.instance;
  }

  async requestCamera(): Promise<boolean> {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: {
          width: { ideal: 1280 },
          height: { ideal: 720 },
          frameRate: { ideal: 30 }
        }
      });

      cameraStream.set(stream);
      cameraPermission.set('granted');
      console.log('[3D Desktop] Camera access granted');
      return true;
    } catch (err) {
      console.error('[3D Desktop] Camera access denied:', err);
      cameraPermission.set('denied');
      return false;
    }
  }

  async requestMicrophone(): Promise<boolean> {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      });

      microphoneStream.set(stream);
      microphonePermission.set('granted');
      console.log('[3D Desktop] Microphone access granted');
      return true;
    } catch (err) {
      console.error('[3D Desktop] Microphone access denied:', err);
      microphonePermission.set('denied');
      return false;
    }
  }

  async requestAll(): Promise<{ camera: boolean; microphone: boolean }> {
    const [camera, microphone] = await Promise.all([
      this.requestCamera(),
      this.requestMicrophone()
    ]);

    return { camera, microphone };
  }

  cleanup(): void {
    // Stop camera stream
    const camera = get(cameraStream);
    if (camera) {
      camera.getTracks().forEach(track => track.stop());
      cameraStream.set(null);
      console.log('[3D Desktop] Camera released');
    }

    // Stop microphone stream
    const microphone = get(microphoneStream);
    if (microphone) {
      microphone.getTracks().forEach(track => track.stop());
      microphoneStream.set(null);
      console.log('[3D Desktop] Microphone released');
    }
  }

  hasCamera(): boolean {
    return get(cameraPermission) === 'granted' && get(cameraStream) !== null;
  }

  hasMicrophone(): boolean {
    return get(microphonePermission) === 'granted' && get(microphoneStream) !== null;
  }
}

export const desktop3dPermissions = Desktop3DPermissions.getInstance();
```

#### 1.2 Add Permission UI Component

```svelte
<!-- src/lib/components/desktop3d/PermissionPrompt.svelte -->
<script lang="ts">
  import { desktop3dPermissions, cameraPermission, microphonePermission } from '$lib/services/desktop3dPermissions';

  let showPrompt = $state(true);
  let requesting = $state(false);

  async function handleRequestPermissions() {
    requesting = true;
    const result = await desktop3dPermissions.requestAll();
    requesting = false;

    if (result.camera && result.microphone) {
      showPrompt = false;
    }
  }

  function handleSkip() {
    showPrompt = false;
  }
</script>

{#if showPrompt && ($cameraPermission === 'prompt' || $microphonePermission === 'prompt')}
  <div class="permission-prompt">
    <div class="prompt-content">
      <div class="prompt-header">
        <svg class="prompt-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        <h3>Enable Advanced Controls</h3>
      </div>

      <p class="prompt-description">
        3D Desktop can use your camera and microphone for:
      </p>

      <ul class="feature-list">
        <li>🤚 Hand tracking and gesture control</li>
        <li>🎤 Voice commands</li>
        <li>👏 Clap and wave gestures</li>
        <li>🎯 Body pointing and presence detection</li>
      </ul>

      <p class="privacy-note">
        🔒 All processing happens locally on your device. No video or audio is sent to servers.
      </p>

      <div class="prompt-actions">
        <button
          onclick={handleSkip}
          class="btn-skip"
          disabled={requesting}
        >
          Skip for now
        </button>
        <button
          onclick={handleRequestPermissions}
          class="btn-enable"
          disabled={requesting}
        >
          {#if requesting}
            <span class="spinner"></span>
            Requesting...
          {:else}
            Enable Camera & Mic
          {/if}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .permission-prompt {
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 9999;
    background: rgba(0, 0, 0, 0.8);
    backdrop-filter: blur(10px);
    padding: 2rem;
    border-radius: 1rem;
    max-width: 500px;
    color: white;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .prompt-content {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .prompt-header {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .prompt-icon {
    width: 3rem;
    height: 3rem;
    color: #60A5FA;
  }

  .prompt-header h3 {
    font-size: 1.5rem;
    font-weight: 600;
    margin: 0;
  }

  .prompt-description {
    font-size: 1rem;
    color: #D1D5DB;
    margin: 0;
  }

  .feature-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .feature-list li {
    font-size: 0.95rem;
    color: #E5E7EB;
  }

  .privacy-note {
    font-size: 0.875rem;
    color: #9CA3AF;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 0.5rem;
    margin: 0;
  }

  .prompt-actions {
    display: flex;
    gap: 1rem;
    margin-top: 0.5rem;
  }

  .btn-skip, .btn-enable {
    flex: 1;
    padding: 0.75rem 1.5rem;
    border-radius: 0.5rem;
    font-weight: 500;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .btn-skip {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .btn-skip:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.15);
  }

  .btn-enable {
    background: #3B82F6;
    color: white;
  }

  .btn-enable:hover:not(:disabled) {
    background: #2563EB;
  }

  .btn-skip:disabled, .btn-enable:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
```

#### 1.3 Integrate into 3D Desktop Page

```svelte
<!-- src/routes/window/+page.svelte - Add this -->
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { desktop3dPermissions } from '$lib/services/desktop3dPermissions';
  import PermissionPrompt from '$lib/components/desktop3d/PermissionPrompt.svelte';

  // ... existing code ...

  onMount(() => {
    // Show permission prompt after a short delay (let user see the 3D Desktop first)
    setTimeout(() => {
      // Permission prompt will auto-show via component
    }, 2000);
  });

  onDestroy(() => {
    // IMPORTANT: Clean up camera/mic when leaving 3D Desktop
    desktop3dPermissions.cleanup();
  });
</script>

<!-- Add permission prompt -->
<PermissionPrompt />

<!-- Existing 3D Desktop content -->
<!-- ... -->
```

---

### Task 2: Layout System with Default Preservation

**Files to create/modify:**
- `src/lib/stores/desktop3dLayoutStore.ts` (new)
- `src/lib/api/desktop3d/layouts.ts` (new - backend API calls)
- `desktop/backend-go/internal/handlers/desktop3d_layouts.go` (new - backend)
- `desktop/backend-go/internal/database/migrations/046_desktop3d_layouts.sql` (new)

#### 2.1 Create Layout Store

```typescript
// src/lib/stores/desktop3dLayoutStore.ts

import { writable, derived, get } from 'svelte/store';
import type { ModuleId } from './desktop3dStore';
import { desktop3dStore } from './desktop3dStore';

export type LayoutType = 'default' | 'custom';

export interface ModulePosition {
  module_id: ModuleId;
  position: { x: number; y: number; z: number };
  rotation: { x: number; y: number; z: number };
  scale: number;
}

export interface Layout {
  id: string;
  name: string;
  type: LayoutType;
  created_at: Date;
  updated_at: Date;
  is_active: boolean;
  user_id: string;
  modules: ModulePosition[];
}

interface LayoutState {
  layouts: Layout[];
  activeLayoutId: string;
  editMode: boolean;
  loading: boolean;
}

const initialState: LayoutState = {
  layouts: [],
  activeLayoutId: 'default',
  editMode: false,
  loading: false
};

function createLayoutStore() {
  const { subscribe, set, update } = writable<LayoutState>(initialState);

  return {
    subscribe,

    // Get default layout (current 5-ring geodesic layout)
    getDefaultLayout: (): Layout => {
      const store = get(desktop3dStore);
      const modules: ModulePosition[] = store.windows.map(win => ({
        module_id: win.module,
        position: win.position,
        rotation: win.rotation || { x: 0, y: 0, z: 0 },
        scale: win.targetScale || 1
      }));

      return {
        id: 'default',
        name: 'Default',
        type: 'default',
        created_at: new Date(),
        updated_at: new Date(),
        is_active: true,
        user_id: '',
        modules
      };
    },

    // Load all layouts from backend
    loadLayouts: async () => {
      update(s => ({ ...s, loading: true }));

      try {
        const response = await fetch('/api/desktop3d/layouts', {
          credentials: 'include'
        });

        if (response.ok) {
          const layouts = await response.json();

          // Always include default layout
          const defaultLayout = get(desktop3dLayoutStore).getDefaultLayout();

          update(s => ({
            ...s,
            layouts: [defaultLayout, ...layouts],
            loading: false
          }));
        } else {
          throw new Error('Failed to load layouts');
        }
      } catch (err) {
        console.error('[Layout Store] Failed to load layouts:', err);

        // On error, just show default layout
        const defaultLayout = get(desktop3dLayoutStore).getDefaultLayout();
        update(s => ({
          ...s,
          layouts: [defaultLayout],
          loading: false
        }));
      }
    },

    // Save current positions as new custom layout
    saveLayout: async (name: string): Promise<boolean> => {
      const store = get(desktop3dStore);
      const modules: ModulePosition[] = store.windows.map(win => ({
        module_id: win.module,
        position: win.position,
        rotation: win.rotation || { x: 0, y: 0, z: 0 },
        scale: win.targetScale || 1
      }));

      try {
        const response = await fetch('/api/desktop3d/layouts', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({
            name,
            modules
          })
        });

        if (response.ok) {
          const newLayout = await response.json();

          update(s => ({
            ...s,
            layouts: [...s.layouts, newLayout],
            activeLayoutId: newLayout.id
          }));

          console.log(`[Layout Store] Saved layout "${name}"`);
          return true;
        } else {
          throw new Error('Failed to save layout');
        }
      } catch (err) {
        console.error('[Layout Store] Failed to save layout:', err);
        return false;
      }
    },

    // Load a specific layout
    loadLayout: async (layoutId: string) => {
      const state = get(desktop3dLayoutStore);
      const layout = state.layouts.find(l => l.id === layoutId);

      if (!layout) {
        console.error(`[Layout Store] Layout ${layoutId} not found`);
        return;
      }

      // Apply positions to desktop3dStore
      layout.modules.forEach(modulePos => {
        desktop3dStore.updateWindowPosition(
          modulePos.module_id,
          modulePos.position,
          modulePos.rotation,
          modulePos.scale
        );
      });

      // Mark as active in backend (if custom layout)
      if (layout.type === 'custom') {
        try {
          await fetch(`/api/desktop3d/layouts/${layoutId}/activate`, {
            method: 'POST',
            credentials: 'include'
          });
        } catch (err) {
          console.error('[Layout Store] Failed to activate layout:', err);
        }
      }

      update(s => ({ ...s, activeLayoutId: layoutId }));
      console.log(`[Layout Store] Loaded layout "${layout.name}"`);
    },

    // Delete a custom layout
    deleteLayout: async (layoutId: string): Promise<boolean> => {
      if (layoutId === 'default') {
        console.warn('[Layout Store] Cannot delete default layout');
        return false;
      }

      try {
        const response = await fetch(`/api/desktop3d/layouts/${layoutId}`, {
          method: 'DELETE',
          credentials: 'include'
        });

        if (response.ok) {
          update(s => {
            const newLayouts = s.layouts.filter(l => l.id !== layoutId);
            const newActiveId = s.activeLayoutId === layoutId ? 'default' : s.activeLayoutId;

            return {
              ...s,
              layouts: newLayouts,
              activeLayoutId: newActiveId
            };
          });

          console.log(`[Layout Store] Deleted layout ${layoutId}`);
          return true;
        } else {
          throw new Error('Failed to delete layout');
        }
      } catch (err) {
        console.error('[Layout Store] Failed to delete layout:', err);
        return false;
      }
    },

    // Toggle edit mode
    toggleEditMode: () => {
      update(s => ({ ...s, editMode: !s.editMode }));
    },

    // Enter edit mode
    enterEditMode: () => {
      update(s => ({ ...s, editMode: true }));
      console.log('[Layout Store] Entered edit mode');
    },

    // Exit edit mode
    exitEditMode: () => {
      update(s => ({ ...s, editMode: false }));
      console.log('[Layout Store] Exited edit mode');
    }
  };
}

export const desktop3dLayoutStore = createLayoutStore();

// Derived stores
export const activeLayout = derived(
  desktop3dLayoutStore,
  $store => $store.layouts.find(l => l.id === $store.activeLayoutId)
);

export const customLayouts = derived(
  desktop3dLayoutStore,
  $store => $store.layouts.filter(l => l.type === 'custom')
);

export const isEditMode = derived(
  desktop3dLayoutStore,
  $store => $store.editMode
);
```

#### 2.2 Backend API

**Database Migration:**

```sql
-- desktop/backend-go/internal/database/migrations/046_desktop3d_layouts.sql

CREATE TABLE IF NOT EXISTS desktop3d_layouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL DEFAULT 'custom', -- 'default' or 'custom'
    is_active BOOLEAN DEFAULT false,
    modules JSONB NOT NULL, -- Array of ModulePosition
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_desktop3d_layouts_user ON desktop3d_layouts(user_id);
CREATE INDEX idx_desktop3d_layouts_active ON desktop3d_layouts(user_id, is_active);

COMMENT ON TABLE desktop3d_layouts IS '3D Desktop custom layout storage';
COMMENT ON COLUMN desktop3d_layouts.modules IS 'JSON array of {module_id, position, rotation, scale}';
```

**Go Handler:**

```go
// desktop/backend-go/internal/handlers/desktop3d_layouts.go

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ModulePosition struct {
	ModuleID string             `json:"module_id"`
	Position map[string]float64 `json:"position"` // {x, y, z}
	Rotation map[string]float64 `json:"rotation"` // {x, y, z}
	Scale    float64            `json:"scale"`
}

type Desktop3DLayout struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	Name      string           `json:"name"`
	Type      string           `json:"type"` // "default" or "custom"
	IsActive  bool             `json:"is_active"`
	Modules   []ModulePosition `json:"modules"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// List all layouts for current user
func (h *Handlers) ListDesktop3DLayouts(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rows, err := h.pool.Query(c.Request.Context(), `
		SELECT id, user_id, name, type, is_active, modules, created_at, updated_at
		FROM desktop3d_layouts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch layouts"})
		return
	}
	defer rows.Close()

	layouts := []Desktop3DLayout{}
	for rows.Next() {
		var layout Desktop3DLayout
		var modulesJSON []byte

		err := rows.Scan(
			&layout.ID,
			&layout.UserID,
			&layout.Name,
			&layout.Type,
			&layout.IsActive,
			&modulesJSON,
			&layout.CreatedAt,
			&layout.UpdatedAt,
		)

		if err != nil {
			continue
		}

		json.Unmarshal(modulesJSON, &layout.Modules)
		layouts = append(layouts, layout)
	}

	c.JSON(http.StatusOK, layouts)
}

// Create new custom layout
func (h *Handlers) CreateDesktop3DLayout(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Name    string           `json:"name" binding:"required"`
		Modules []ModulePosition `json:"modules" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	modulesJSON, _ := json.Marshal(input.Modules)
	layoutID := uuid.New()

	_, err := h.pool.Exec(c.Request.Context(), `
		INSERT INTO desktop3d_layouts (id, user_id, name, type, modules)
		VALUES ($1, $2, $3, 'custom', $4)
	`, layoutID, user.ID, input.Name, modulesJSON)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save layout"})
		return
	}

	// Return created layout
	var layout Desktop3DLayout
	h.pool.QueryRow(c.Request.Context(), `
		SELECT id, user_id, name, type, is_active, modules, created_at, updated_at
		FROM desktop3d_layouts
		WHERE id = $1
	`, layoutID).Scan(
		&layout.ID,
		&layout.UserID,
		&layout.Name,
		&layout.Type,
		&layout.IsActive,
		&modulesJSON,
		&layout.CreatedAt,
		&layout.UpdatedAt,
	)

	json.Unmarshal(modulesJSON, &layout.Modules)
	c.JSON(http.StatusOK, layout)
}

// Activate a layout
func (h *Handlers) ActivateDesktop3DLayout(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	layoutID := c.Param("id")

	// Deactivate all other layouts
	_, err := h.pool.Exec(c.Request.Context(), `
		UPDATE desktop3d_layouts
		SET is_active = false
		WHERE user_id = $1
	`, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update layouts"})
		return
	}

	// Activate this layout
	_, err = h.pool.Exec(c.Request.Context(), `
		UPDATE desktop3d_layouts
		SET is_active = true, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, layoutID, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate layout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Layout activated"})
}

// Delete a layout
func (h *Handlers) DeleteDesktop3DLayout(c *gin.Context) {
	user := getUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	layoutID := c.Param("id")

	_, err := h.pool.Exec(c.Request.Context(), `
		DELETE FROM desktop3d_layouts
		WHERE id = $1 AND user_id = $2
	`, layoutID, user.ID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete layout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Layout deleted"})
}
```

**Add routes in main.go:**

```go
// Desktop 3D layouts
api.GET("/desktop3d/layouts", handlers.ListDesktop3DLayouts)
api.POST("/desktop3d/layouts", handlers.CreateDesktop3DLayout)
api.POST("/desktop3d/layouts/:id/activate", handlers.ActivateDesktop3DLayout)
api.DELETE("/desktop3d/layouts/:id", handlers.DeleteDesktop3DLayout)
```

---

### Task 3: Edit Mode & Drag-to-Position

**Files to modify:**
- `src/lib/stores/desktop3dStore.ts` (add updateWindowPosition method)
- `src/lib/components/desktop3d/Desktop3DWindow.svelte` (add drag handlers)
- `src/routes/window/+page.svelte` (add edit mode UI)

#### 3.1 Add Method to Update Window Position

```typescript
// src/lib/stores/desktop3dStore.ts - Add this method

updateWindowPosition: (
  moduleId: ModuleId,
  position: { x: number; y: number; z: number },
  rotation?: { x: number; y: number; z: number },
  scale?: number
) => {
  update((state) => {
    const windows = state.windows.map(w => {
      if (w.module === moduleId) {
        return {
          ...w,
          position,
          rotation: rotation || w.rotation || { x: 0, y: 0, z: 0 },
          targetScale: scale || w.targetScale || 1
        };
      }
      return w;
    });

    return { ...state, windows };
  });
}
```

#### 3.2 Make Windows Draggable in Edit Mode

This will use Threlte's TransformControls for 3D dragging. Will implement in actual code.

---

### Task 4: Layout Manager UI

**Files to create:**
- `src/lib/components/desktop3d/LayoutManager.svelte` (new)
- `src/lib/components/desktop3d/EditModeToolbar.svelte` (new)

#### 4.1 Edit Mode Toolbar

```svelte
<!-- src/lib/components/desktop3d/EditModeToolbar.svelte -->

<script lang="ts">
  import { desktop3dLayoutStore, isEditMode } from '$lib/stores/desktop3dLayoutStore';

  let showSaveDialog = $state(false);
  let layoutName = $state('');
  let saving = $state(false);

  async function handleSave() {
    if (!layoutName.trim()) return;

    saving = true;
    const success = await desktop3dLayoutStore.saveLayout(layoutName);
    saving = false;

    if (success) {
      layoutName = '';
      showSaveDialog = false;
      desktop3dLayoutStore.exitEditMode();
    }
  }

  function handleCancel() {
    desktop3dLayoutStore.exitEditMode();
    // TODO: Reload current active layout to discard changes
  }
</script>

{#if $isEditMode}
  <div class="edit-toolbar">
    <div class="toolbar-content">
      <div class="toolbar-info">
        <svg class="icon-edit" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
        </svg>
        <span>Edit Mode - Drag modules to reposition</span>
      </div>

      <div class="toolbar-actions">
        <button onclick={handleCancel} class="btn-cancel">
          Cancel
        </button>
        <button onclick={() => showSaveDialog = true} class="btn-save">
          Save Layout
        </button>
      </div>
    </div>
  </div>

  {#if showSaveDialog}
    <div class="save-dialog-overlay">
      <div class="save-dialog">
        <h3>Save Custom Layout</h3>
        <input
          type="text"
          placeholder="Enter layout name..."
          bind:value={layoutName}
          class="layout-name-input"
          autofocus
        />
        <div class="dialog-actions">
          <button onclick={() => showSaveDialog = false} class="btn-cancel">
            Cancel
          </button>
          <button
            onclick={handleSave}
            class="btn-save"
            disabled={!layoutName.trim() || saving}
          >
            {saving ? 'Saving...' : 'Save'}
          </button>
        </div>
      </div>
    </div>
  {/if}
{/if}

<style>
  /* Toolbar styles */
  .edit-toolbar {
    position: fixed;
    top: 1rem;
    left: 50%;
    transform: translateX(-50%);
    z-index: 1000;
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(10px);
    border-radius: 0.75rem;
    padding: 1rem 1.5rem;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.5);
  }

  .toolbar-content {
    display: flex;
    align-items: center;
    gap: 2rem;
  }

  .toolbar-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    color: white;
    font-weight: 500;
  }

  .icon-edit {
    width: 1.5rem;
    height: 1.5rem;
    color: #60A5FA;
  }

  .toolbar-actions {
    display: flex;
    gap: 0.75rem;
  }

  .btn-cancel, .btn-save {
    padding: 0.5rem 1.25rem;
    border-radius: 0.5rem;
    font-weight: 500;
    border: none;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-cancel {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }

  .btn-cancel:hover {
    background: rgba(255, 255, 255, 0.15);
  }

  .btn-save {
    background: #3B82F6;
    color: white;
  }

  .btn-save:hover:not(:disabled) {
    background: #2563EB;
  }

  .btn-save:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Save dialog styles */
  .save-dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(5px);
    z-index: 10000;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .save-dialog {
    background: #1F2937;
    border-radius: 1rem;
    padding: 2rem;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .save-dialog h3 {
    color: white;
    margin: 0 0 1.5rem 0;
    font-size: 1.25rem;
  }

  .layout-name-input {
    width: 100%;
    padding: 0.75rem;
    background: #374151;
    border: 2px solid #4B5563;
    border-radius: 0.5rem;
    color: white;
    font-size: 1rem;
    margin-bottom: 1.5rem;
  }

  .layout-name-input:focus {
    outline: none;
    border-color: #3B82F6;
  }

  .dialog-actions {
    display: flex;
    gap: 1rem;
  }

  .dialog-actions .btn-cancel,
  .dialog-actions .btn-save {
    flex: 1;
  }
</style>
```

---

## 📦 Implementation Steps (Sequential)

### Step 1: Camera/Mic Permissions (Day 1)
1. Create `desktop3dPermissions.ts` service
2. Create `PermissionPrompt.svelte` component
3. Integrate into `/window` route with onMount/onDestroy
4. Test: Permission prompt appears, camera/mic access works, cleanup on exit

### Step 2: Backend Layout API (Day 1-2)
1. Create database migration `046_desktop3d_layouts.sql`
2. Run migration
3. Create Go handlers in `desktop3d_layouts.go`
4. Add routes to main.go
5. Test: CRUD operations work via Postman/curl

### Step 3: Frontend Layout Store (Day 2)
1. Create `desktop3dLayoutStore.ts`
2. Add `updateWindowPosition` method to desktop3dStore
3. Test: Store loads default layout, can save/load/delete layouts

### Step 4: Edit Mode UI (Day 3)
1. Create `EditModeToolbar.svelte`
2. Add edit mode button to main 3D Desktop UI
3. Make windows draggable in edit mode (using Threlte TransformControls)
4. Test: Can enter edit mode, drag modules, save positions

### Step 5: Layout Manager (Day 3-4)
1. Create `LayoutManager.svelte` modal
2. Show list of saved layouts
3. Quick switch between layouts
4. Delete custom layouts
5. Export/import (bonus)

### Step 6: Polish & Testing (Day 4-5)
1. Add loading states
2. Add error handling
3. Add animations/transitions
4. Test all flows end-to-end
5. Fix bugs

---

## 🎯 Success Criteria

- [x] Permission prompt shows only in 3D Desktop mode
- [x] Camera/microphone cleanup when leaving 3D Desktop
- [x] Default 5-ring geodesic layout never changes
- [x] Can enter edit mode and drag modules
- [x] Can save custom layouts with names
- [x] Can load saved layouts
- [x] Can delete custom layouts
- [x] Active layout persists across sessions
- [x] All layouts stored in database

---

## 🚀 Ready to Start?

This plan gives us:
1. ✅ Camera/mic permissions (only in 3D Desktop)
2. ✅ Default layout preservation (5-ring geodesic)
3. ✅ Custom positioning with drag-and-drop
4. ✅ Layout persistence (backend + database)
5. ✅ Simple layout manager UI

**Estimated time:** 3-5 days

Should I start implementing? Which part would you like me to begin with?
