// Early crash handling - must be at the very top
process.on('uncaughtException', (error) => {
  console.error('UNCAUGHT EXCEPTION:', error);
  console.error('Stack:', error.stack);
});

process.on('unhandledRejection', (reason, promise) => {
  console.error('UNHANDLED REJECTION at:', promise, 'reason:', reason);
});

import { app, BrowserWindow, ipcMain, Menu, Tray, nativeImage, protocol, net, session } from 'electron';
import path from 'path';
import { createMainWindow, getMainWindow } from './window';
import { setupIpcHandlers, initializeDatabaseSystem, startSync, stopSync } from './ipc';
import { BackendManager } from './backend/manager';
import { setupAutoUpdater } from './updater/auto-update';
import { initializePopupSystem, cleanupPopupSystem } from './popup/chat-popup';
import { initializeMeetingRecorder } from './audio/meeting-recorder';
import { closeDatabase } from './database/sqlite';
import { pathToFileURL } from 'url';

console.log('Main process starting - imports complete');
console.log('Platform:', process.platform);
console.log('Is packaged:', app.isPackaged);

// Handle Squirrel events for Windows installer (only on Windows)
if (process.platform === 'win32') {
  console.log('Checking Squirrel startup...');
  try {
    if (require('electron-squirrel-startup')) {
      console.log('Squirrel startup - quitting');
      app.quit();
    }
  } catch {
    console.log('Squirrel startup not available');
    // electron-squirrel-startup not available, ignore on non-Windows
  }
}

// Single instance lock
console.log('Requesting single instance lock...');
const gotTheLock = app.requestSingleInstanceLock();
console.log('Got lock:', gotTheLock);

if (!gotTheLock) {
  console.log('Another instance running - quitting');
  app.quit();
} else {
  console.log('Single instance lock acquired');
  app.on('second-instance', () => {
    // Someone tried to run a second instance, focus our window
    const mainWindow = getMainWindow();
    if (mainWindow) {
      if (mainWindow.isMinimized()) mainWindow.restore();
      mainWindow.focus();
    }
  });
}

// Global references
let backendManager: BackendManager | null = null;

// App metadata
const isDev = !app.isPackaged;
const appPath = app.getAppPath();
const resourcesPath = isDev
  ? path.join(appPath, 'resources')
  : process.resourcesPath;

// Register custom protocol for serving app files
// This allows SvelteKit to work correctly with file:// URLs
protocol.registerSchemesAsPrivileged([
  {
    scheme: 'app',
    privileges: {
      standard: true,
      secure: true,
      supportFetchAPI: true,
      corsEnabled: true,
    },
  },
]);

/**
 * Create the native application menu
 */
function createAppMenu(): void {
  const isMac = process.platform === 'darwin';

  const template: Electron.MenuItemConstructorOptions[] = [
    // App menu (macOS only)
    ...(isMac ? [{
      label: app.name,
      submenu: [
        { role: 'about' as const },
        { type: 'separator' as const },
        {
          label: 'Preferences...',
          accelerator: 'CommandOrControl+,',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/profile');
            }
          }
        },
        { type: 'separator' as const },
        { role: 'services' as const },
        { type: 'separator' as const },
        { role: 'hide' as const },
        { role: 'hideOthers' as const },
        { role: 'unhide' as const },
        { type: 'separator' as const },
        { role: 'quit' as const }
      ] as Electron.MenuItemConstructorOptions[]
    }] : []),
    // File menu
    {
      label: 'File',
      submenu: [
        {
          label: 'New Task',
          accelerator: 'CommandOrControl+N',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('shortcut', 'new-task');
            }
          }
        },
        {
          label: 'New Project',
          accelerator: 'CommandOrControl+Shift+N',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('shortcut', 'new-project');
            }
          }
        },
        { type: 'separator' },
        isMac ? { role: 'close' } : { role: 'quit' }
      ] as Electron.MenuItemConstructorOptions[]
    },
    // Edit menu
    {
      label: 'Edit',
      submenu: [
        { role: 'undo' },
        { role: 'redo' },
        { type: 'separator' },
        { role: 'cut' },
        { role: 'copy' },
        { role: 'paste' },
        ...(isMac ? [
          { role: 'pasteAndMatchStyle' as const },
          { role: 'delete' as const },
          { role: 'selectAll' as const },
        ] : [
          { role: 'delete' as const },
          { type: 'separator' as const },
          { role: 'selectAll' as const }
        ])
      ] as Electron.MenuItemConstructorOptions[]
    },
    // View menu
    {
      label: 'View',
      submenu: [
        { role: 'reload' },
        { role: 'forceReload' },
        { role: 'toggleDevTools' },
        { type: 'separator' },
        { role: 'resetZoom' },
        { role: 'zoomIn' },
        { role: 'zoomOut' },
        { type: 'separator' },
        { role: 'togglefullscreen' }
      ] as Electron.MenuItemConstructorOptions[]
    },
    // Navigate menu
    {
      label: 'Navigate',
      submenu: [
        {
          label: 'Dashboard',
          accelerator: 'CommandOrControl+1',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/dashboard');
            }
          }
        },
        {
          label: 'Tasks',
          accelerator: 'CommandOrControl+2',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/tasks');
            }
          }
        },
        {
          label: 'Calendar',
          accelerator: 'CommandOrControl+3',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/calendar');
            }
          }
        },
        {
          label: 'Projects',
          accelerator: 'CommandOrControl+4',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/projects');
            }
          }
        },
        {
          label: 'Chat',
          accelerator: 'CommandOrControl+5',
          click: () => {
            const mainWindow = getMainWindow();
            if (mainWindow) {
              mainWindow.webContents.send('navigate', '/chat');
            }
          }
        },
      ]
    },
    // Window menu
    {
      label: 'Window',
      submenu: [
        { role: 'minimize' },
        { role: 'zoom' },
        ...(isMac ? [
          { type: 'separator' as const },
          { role: 'front' as const },
          { type: 'separator' as const },
          { role: 'window' as const }
        ] : [
          { role: 'close' as const }
        ])
      ] as Electron.MenuItemConstructorOptions[]
    },
    // Help menu
    {
      role: 'help',
      submenu: [
        {
          label: 'Documentation',
          click: async () => {
            const { shell } = require('electron');
            await shell.openExternal('https://businessos.app/docs');
          }
        },
        {
          label: 'Report Issue',
          click: async () => {
            const { shell } = require('electron');
            await shell.openExternal('https://github.com/your-org/businessos-desktop/issues');
          }
        },
        { type: 'separator' },
        {
          label: 'About BusinessOS',
          click: () => {
            const { dialog } = require('electron');
            dialog.showMessageBox({
              type: 'info',
              title: 'About BusinessOS',
              message: 'BusinessOS Desktop',
              detail: `Version: ${app.getVersion()}\nElectron: ${process.versions.electron}\nChrome: ${process.versions.chrome}\nNode.js: ${process.versions.node}`
            });
          }
        }
      ]
    }
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
}

/**
 * Configure session to persist cookies for authentication
 */
function configureSessionPersistence(): void {
  const ses = session.defaultSession;

  // Configure session to persist cookies
  // This ensures OAuth cookies survive app restarts
  ses.setUserAgent(ses.getUserAgent() + ' BusinessOS-Desktop');

  // Log session configuration for debugging
  console.log('Session persistence configured');
  console.log(`Session path: ${ses.getStoragePath()}`);

  // Get all cookies to verify persistence (async, will complete after init)
  ses.cookies.get({}).then(cookies => {
    console.log(`Found ${cookies.length} persisted cookies`);
  }).catch(err => {
    console.error('Error checking cookies:', err);
  });
}

/**
 * Initialize the application
 */
async function initialize(): Promise<void> {
  console.log('BusinessOS Desktop starting...');
  console.log(`Environment: ${isDev ? 'development' : 'production'}`);
  console.log(`App path: ${appPath}`);
  console.log(`Resources path: ${resourcesPath}`);

  // Configure session persistence BEFORE anything else
  // This ensures cookies from OAuth survive app restarts
  configureSessionPersistence();

  // Initialize local SQLite database and sync engine
  initializeDatabaseSystem();
  console.log('Local database initialized');

  // Register the app:// protocol handler for serving static files
  if (!isDev) {
    protocol.handle('app', (request) => {
      const url = new URL(request.url);
      // Map app://- to the renderer directory
      let filePath = url.pathname;

      // Default to index.html for root or SPA routes
      if (filePath === '/' || filePath === '') {
        filePath = '/index.html';
      }

      // Construct the full path to the file
      const basePath = path.join(__dirname, '../renderer/main_window');
      const fullPath = path.join(basePath, filePath);

      // Return the file
      return net.fetch(pathToFileURL(fullPath).href);
    });
    console.log('Registered app:// protocol handler');
  }

  // Start the Go backend sidecar
  // Skip BackendManager in dev mode - use external backend instead
  if (!isDev) {
    backendManager = new BackendManager(resourcesPath);

    try {
      await backendManager.start();
      console.log('Go backend started successfully');
    } catch (error) {
      console.error('Failed to start Go backend:', error);
      throw error;
    }
  } else {
    console.log('Dev mode: Skipping embedded backend (use external backend)');
    backendManager = null;
  }

  // Set up IPC handlers
  setupIpcHandlers(backendManager);

  // Create the application menu
  createAppMenu();

  // Create the main window
  await createMainWindow();

  // Initialize popup chat system (includes tray and global shortcuts)
  initializePopupSystem();

  // Initialize meeting recorder
  initializeMeetingRecorder();
  console.log('Meeting recorder initialized');

  // Set up auto-updater (production only)
  if (!isDev) {
    setupAutoUpdater();
  }

  // Start sync engine
  startSync();
  console.log('Sync engine started');
}

// App lifecycle events
app.whenReady().then(initialize).catch(console.error);

app.on('window-all-closed', () => {
  // On macOS, keep the app running in the background (tray)
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

app.on('activate', () => {
  // On macOS, re-create window when dock icon is clicked
  if (BrowserWindow.getAllWindows().length === 0) {
    createMainWindow();
  } else {
    const mainWindow = getMainWindow();
    if (mainWindow) {
      mainWindow.show();
      mainWindow.focus();
    }
  }
});

app.on('before-quit', async () => {
  console.log('BusinessOS Desktop shutting down...');

  // Stop sync engine
  stopSync();

  // Cleanup popup system (shortcuts, tray)
  cleanupPopupSystem();

  // Close SQLite database
  closeDatabase();

  // Stop the Go backend
  if (backendManager) {
    await backendManager.stop();
  }
});

// Handle deep links (businessos://)
app.on('open-url', (event, url) => {
  event.preventDefault();
  console.log('Deep link received:', url);

  const mainWindow = getMainWindow();
  if (mainWindow) {
    // Parse the URL and navigate accordingly
    const parsed = new URL(url);
    const path = parsed.pathname;
    mainWindow.webContents.send('navigate', path);
    mainWindow.show();
    mainWindow.focus();
  }
});

// Register deep link protocol
if (process.defaultApp) {
  if (process.argv.length >= 2) {
    app.setAsDefaultProtocolClient('businessos', process.execPath, [path.resolve(process.argv[1])]);
  }
} else {
  app.setAsDefaultProtocolClient('businessos');
}
