# P2: Terminal & Filesystem

> **Priority:** P2 - Nice to Have
> **Backend Status:** Complete (11 endpoints)
> **Frontend Status:** Not Started
> **Estimated Effort:** 1-2 sprints

---

## Overview

Developer-focused features for embedded terminal and file system access. Enables code execution, file management, and developer workflows within BusinessOS.

---

## Backend API Endpoints

### Terminal (PTY)
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/terminal/ws` | WebSocket for terminal session |
| GET | `/api/terminal/sessions` | List active sessions |
| DELETE | `/api/terminal/sessions/:id` | Close session |

### Filesystem
| Method | Endpoint | Purpose |
|--------|----------|---------|
| GET | `/api/filesystem/list` | List directory contents |
| GET | `/api/filesystem/read` | Read file contents |
| GET | `/api/filesystem/download` | Download file |
| GET | `/api/filesystem/info` | Get file info |
| GET | `/api/filesystem/quick-access` | Quick access paths |
| POST | `/api/filesystem/mkdir` | Create directory |
| POST | `/api/filesystem/upload` | Upload file |
| DELETE | `/api/filesystem/delete` | Delete file/directory |

---

## Frontend Implementation Tasks

### Phase 1: Terminal Component
- [ ] XTerm.js integration
- [ ] WebSocket connection to PTY
- [ ] Terminal tabs (multiple sessions)
- [ ] Copy/paste support
- [ ] Resizable terminal panel

### Phase 2: File Browser
- [ ] Tree view file explorer
- [ ] Breadcrumb navigation
- [ ] File icons by type
- [ ] Context menu (rename, delete, etc.)
- [ ] Drag-and-drop upload

### Phase 3: Code Viewer/Editor
- [ ] Monaco editor integration
- [ ] Syntax highlighting
- [ ] Basic editing capabilities
- [ ] Save file

---

## Linear Issues to Create

1. **[TERM-001]** Integrate XTerm.js terminal
2. **[TERM-002]** WebSocket PTY connection
3. **[TERM-003]** Build file browser component
4. **[TERM-004]** Add file viewer/editor
5. **[TERM-005]** API client implementation

---

## Notes

- Consider security implications of terminal access
- File access should be sandboxed appropriately
- May want to limit to specific project directories
