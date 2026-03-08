import { app, dialog } from 'electron';
import { getMainWindow, sendToMainWindow } from '../window';

// Lazy-loaded auto-updater to avoid initialization errors
let autoUpdater: any = null;

function getAutoUpdater() {
  if (!autoUpdater) {
    // Only import when needed (after app is ready)
    const { autoUpdater: au } = require('electron-updater');
    autoUpdater = au;
    autoUpdater.autoDownload = false;
    autoUpdater.autoInstallOnAppQuit = true;
  }
  return autoUpdater;
}

/**
 * Set up the auto-updater with event handlers
 */
export function setupAutoUpdater(): void {
  const updater = getAutoUpdater();

  // Check for updates on startup (after a delay)
  setTimeout(() => {
    checkForUpdates();
  }, 10000);

  // Check for updates periodically (every 4 hours)
  setInterval(() => {
    checkForUpdates();
  }, 4 * 60 * 60 * 1000);

  // Event: Error occurred
  updater.on('error', (error: Error) => {
    console.error('Auto-update error:', error);
    sendToMainWindow('update:error', error.message);
  });

  // Event: Checking for updates
  updater.on('checking-for-update', () => {
    console.log('Checking for updates...');
    sendToMainWindow('update:checking');
  });

  // Event: Update available
  updater.on('update-available', (info: any) => {
    console.log('Update available:', info.version);
    sendToMainWindow('update:available', {
      version: info.version,
      releaseDate: info.releaseDate,
      releaseNotes: info.releaseNotes,
    });

    // Show dialog to user
    const mainWindow = getMainWindow();
    if (mainWindow) {
      dialog.showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Available',
        message: `A new version (${info.version}) is available.`,
        detail: 'Would you like to download it now?',
        buttons: ['Download', 'Later'],
        defaultId: 0,
      }).then(({ response }) => {
        if (response === 0) {
          downloadUpdate();
        }
      });
    }
  });

  // Event: Update not available
  updater.on('update-not-available', () => {
    console.log('No updates available');
    sendToMainWindow('update:not-available');
  });

  // Event: Download progress
  updater.on('download-progress', (progress: any) => {
    console.log(`Download progress: ${progress.percent.toFixed(1)}%`);
    sendToMainWindow('update:download-progress', {
      percent: progress.percent,
      bytesPerSecond: progress.bytesPerSecond,
      transferred: progress.transferred,
      total: progress.total,
    });
  });

  // Event: Update downloaded
  updater.on('update-downloaded', (info: any) => {
    console.log('Update downloaded:', info.version);
    sendToMainWindow('update:downloaded', {
      version: info.version,
    });

    // Show dialog to user
    const mainWindow = getMainWindow();
    if (mainWindow) {
      dialog.showMessageBox(mainWindow, {
        type: 'info',
        title: 'Update Ready',
        message: `Version ${info.version} has been downloaded.`,
        detail: 'The update will be installed when you restart the app. Would you like to restart now?',
        buttons: ['Restart Now', 'Later'],
        defaultId: 0,
      }).then(({ response }) => {
        if (response === 0) {
          installUpdate();
        }
      });
    }
  });

  console.log('Auto-updater initialized');
}

/**
 * Check for updates
 */
export async function checkForUpdates(): Promise<void> {
  try {
    const updater = getAutoUpdater();
    await updater.checkForUpdates();
  } catch (error) {
    console.error('Failed to check for updates:', error);
  }
}

/**
 * Download available update
 */
export async function downloadUpdate(): Promise<void> {
  try {
    const updater = getAutoUpdater();
    await updater.downloadUpdate();
  } catch (error) {
    console.error('Failed to download update:', error);
    sendToMainWindow('update:error', 'Failed to download update');
  }
}

/**
 * Install downloaded update (quit and install)
 */
export function installUpdate(): void {
  const updater = getAutoUpdater();
  updater.quitAndInstall(false, true);
}
