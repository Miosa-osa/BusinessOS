import { ipcMain } from 'electron';
import { BackendManager } from '../backend/manager';

/**
 * Set up all OSA (OptimalSystemAgent) IPC handlers
 */
export function setupOSAHandlers(backendManager: BackendManager | null): void {
  // osa:health → GET /api/osa/health
  ipcMain.handle('osa:health', async () => {
    try {
      const baseUrl = backendManager?.getUrl() ?? 'http://localhost:8000';
      const res = await fetch(`${baseUrl}/api/osa/health`);
      if (!res.ok) return { success: false, error: `HTTP ${res.status}` };
      const data = await res.json();
      return { success: true, data };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  });

  // osa:orchestrate → POST /api/osa/orchestrate
  // body: { input, user_id, workspace_id?, session_id? }
  ipcMain.handle(
    'osa:orchestrate',
    async (
      _,
      req: {
        input: string;
        user_id: string;
        workspace_id?: string;
        session_id?: string;
      }
    ) => {
      try {
        const baseUrl = backendManager?.getUrl() ?? 'http://localhost:8000';
        const res = await fetch(`${baseUrl}/api/osa/orchestrate`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(req),
        });
        if (!res.ok) return { success: false, error: `HTTP ${res.status}` };
        const data = await res.json();
        return { success: true, data };
      } catch (error) {
        return { success: false, error: String(error) };
      }
    }
  );

  // osa:classify → POST /api/osa/classify
  // body: { message, channel }
  ipcMain.handle('osa:classify', async (_, message: string, channel: string = 'http') => {
    try {
      const baseUrl = backendManager?.getUrl() ?? 'http://localhost:8000';
      const res = await fetch(`${baseUrl}/api/osa/classify`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message, channel }),
      });
      if (!res.ok) return { success: false, error: `HTTP ${res.status}` };
      const data = await res.json();
      return { success: true, data };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  });

  // osa:skills → GET /api/osa/skills
  ipcMain.handle('osa:skills', async () => {
    try {
      const baseUrl = backendManager?.getUrl() ?? 'http://localhost:8000';
      const res = await fetch(`${baseUrl}/api/osa/skills`);
      if (!res.ok) return { success: false, error: `HTTP ${res.status}` };
      const data = await res.json();
      return { success: true, data };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  });

  // osa:stream:start → starts listening to SSE from /api/osa/stream/:sessionID
  // and forwards events to the renderer via 'osa:stream:event' channel
  ipcMain.handle('osa:stream:start', async (event, sessionID: string) => {
    try {
      const baseUrl = backendManager?.getUrl() ?? 'http://localhost:8000';
      const url = `${baseUrl}/api/osa/stream/${sessionID}`;

      const controller = new AbortController();
      const res = await fetch(url, { signal: controller.signal });

      if (!res.ok || !res.body) {
        return { success: false, error: `HTTP ${res.status}` };
      }

      // Stream SSE events asynchronously to the renderer
      const sender = event.sender;
      (async () => {
        try {
          const reader = res.body!.getReader();
          const decoder = new TextDecoder();
          let buffer = '';

          while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop() ?? '';

            for (const line of lines) {
              if (line.startsWith('data: ')) {
                try {
                  const data = JSON.parse(line.slice(6));
                  sender.send('osa:stream:event', { sessionID, data });
                } catch {
                  // Skip malformed JSON
                }
              }
            }
          }

          sender.send('osa:stream:event', { sessionID, data: null, done: true });
        } catch {
          // Stream ended or was aborted
        }
      })();

      return { success: true };
    } catch (error) {
      return { success: false, error: String(error) };
    }
  });
}
