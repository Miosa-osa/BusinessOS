import { ipcMain, app, shell, dialog, desktopCapturer, screen } from 'electron';
import { BackendManager } from '../backend/manager';
import { getMainWindow } from '../window';
import { setupDatabaseHandlers, initializeDatabaseSystem, startSync, stopSync } from './database';

// Re-export database functions for use in main process
export { initializeDatabaseSystem, startSync, stopSync };

/**
 * Set up all IPC handlers for communication with the renderer process
 */
export function setupIpcHandlers(backendManager: BackendManager | null): void {
  // Set up database IPC handlers
  setupDatabaseHandlers();
  // App info
  ipcMain.handle('app:get-version', () => {
    return app.getVersion();
  });

  ipcMain.handle('app:get-platform', () => {
    return {
      platform: process.platform,
      arch: process.arch,
      isPackaged: app.isPackaged,
    };
  });

  ipcMain.handle('app:get-path', (_, name: string) => {
    return app.getPath(name as any);
  });

  // Backend status
  ipcMain.handle('backend:get-status', () => {
    return {
      running: backendManager?.isRunning() ?? false,
      port: backendManager?.getPort() ?? 0,
      url: backendManager?.getUrl() ?? '',
    };
  });

  ipcMain.handle('backend:get-url', () => {
    return backendManager?.getUrl() ?? 'http://localhost:8000';
  });

  ipcMain.handle('backend:restart', async () => {
    if (backendManager) {
      await backendManager.restart();
      return true;
    }
    return false;
  });

  // Network status
  ipcMain.handle('network:get-status', () => {
    // Check if online by attempting to reach the remote server
    return {
      online: true, // Simplified - in production, implement actual check
    };
  });

  // Shell operations
  ipcMain.handle('shell:open-external', async (_, url: string) => {
    await shell.openExternal(url);
  });

  ipcMain.handle('shell:open-path', async (_, path: string) => {
    await shell.openPath(path);
  });

  // Dialog operations
  ipcMain.handle('dialog:show-open', async (_, options: Electron.OpenDialogOptions) => {
    const mainWindow = getMainWindow();
    if (!mainWindow) return { canceled: true, filePaths: [] };
    return dialog.showOpenDialog(mainWindow, options);
  });

  ipcMain.handle('dialog:show-save', async (_, options: Electron.SaveDialogOptions) => {
    const mainWindow = getMainWindow();
    if (!mainWindow) return { canceled: true, filePath: undefined };
    return dialog.showSaveDialog(mainWindow, options);
  });

  ipcMain.handle('dialog:show-message', async (_, options: Electron.MessageBoxOptions) => {
    const mainWindow = getMainWindow();
    if (!mainWindow) return { response: 0 };
    return dialog.showMessageBox(mainWindow, options);
  });

  // Window state persistence
  ipcMain.handle('window:get-state', () => {
    const mainWindow = getMainWindow();
    if (!mainWindow) return null;

    const bounds = mainWindow.getBounds();
    return {
      x: bounds.x,
      y: bounds.y,
      width: bounds.width,
      height: bounds.height,
      isMaximized: mainWindow.isMaximized(),
      isFullScreen: mainWindow.isFullScreen(),
    };
  });

  ipcMain.on('window:set-state', (_, state: { width: number; height: number; x?: number; y?: number }) => {
    const mainWindow = getMainWindow();
    if (!mainWindow) return;

    if (state.x !== undefined && state.y !== undefined) {
      mainWindow.setBounds({
        x: state.x,
        y: state.y,
        width: state.width,
        height: state.height,
      });
    } else {
      mainWindow.setSize(state.width, state.height);
    }
  });

  // Sync operations are now handled in database.ts via setupDatabaseHandlers()
  // Keeping only the stub status handler here for backwards compatibility
  ipcMain.handle('sync:get-status', () => {
    return {
      status: 'synced',
      lastSync: new Date().toISOString(),
      pendingChanges: 0,
    };
  });

  // NOTE: sync:trigger handler is registered in database.ts, not here

  // Update operations (to be implemented with auto-updater)
  ipcMain.handle('updates:check', async () => {
    // Check for updates
    return {
      available: false,
      version: null,
    };
  });

  ipcMain.handle('updates:download', async () => {
    // Download update
    return false;
  });

  ipcMain.handle('updates:install', async () => {
    // Install and restart
    return false;
  });

  // Screenshot capture
  ipcMain.handle('screenshot:capture', async () => {
    try {
      // Get all displays
      const displays = screen.getAllDisplays();
      const primaryDisplay = screen.getPrimaryDisplay();

      // Get desktop capturer sources
      const sources = await desktopCapturer.getSources({
        types: ['screen'],
        thumbnailSize: {
          width: primaryDisplay.size.width,
          height: primaryDisplay.size.height
        }
      });

      if (sources.length === 0) {
        return { success: false, error: 'No screen sources available' };
      }

      // Get the primary screen source
      const primarySource = sources.find(s => s.display_id === primaryDisplay.id.toString()) || sources[0];

      // Get the thumbnail as a data URL
      const thumbnail = primarySource.thumbnail;
      const dataUrl = thumbnail.toDataURL();

      return {
        success: true,
        dataUrl,
        size: {
          width: thumbnail.getSize().width,
          height: thumbnail.getSize().height
        }
      };
    } catch (error) {
      console.error('Screenshot capture failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Screenshot capture failed'
      };
    }
  });

  console.log('IPC handlers registered');
}
