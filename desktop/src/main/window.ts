import { BrowserWindow, shell, app } from 'electron';
import path from 'path';

// Global reference to main window
let mainWindow: BrowserWindow | null = null;

// Window state defaults
const DEFAULT_WIDTH = 1400;
const DEFAULT_HEIGHT = 900;
const MIN_WIDTH = 800;
const MIN_HEIGHT = 600;

/**
 * Get the main window instance
 */
export function getMainWindow(): BrowserWindow | null {
  return mainWindow;
}

/**
 * Create the main application window
 */
export async function createMainWindow(): Promise<BrowserWindow> {
  const isDev = !app.isPackaged;

  mainWindow = new BrowserWindow({
    width: DEFAULT_WIDTH,
    height: DEFAULT_HEIGHT,
    minWidth: MIN_WIDTH,
    minHeight: MIN_HEIGHT,
    title: 'BusinessOS',
    show: false, // Don't show until ready
    backgroundColor: '#f9fafb',
    titleBarStyle: process.platform === 'darwin' ? 'hiddenInset' : 'default',
    trafficLightPosition: { x: 20, y: 18 },
    webPreferences: {
      preload: path.join(__dirname, '../preload/index.js'),
      nodeIntegration: false,
      contextIsolation: true,
      sandbox: false,
      webSecurity: true,
      allowRunningInsecureContent: false,
      // Enable webview tag for embedding external apps
      webviewTag: true,
    },
  });

  // Load the app
  if (isDev) {
    // In development, load from the Vite dev server
    const devUrl = 'http://localhost:5173';
    console.log(`Loading from ${devUrl}`);
    await mainWindow.loadURL(devUrl);
  } else {
    // In production, use the custom app:// protocol to serve files
    // This allows SvelteKit's router to work correctly
    console.log('Loading production app via app:// protocol');
    await mainWindow.loadURL('app://localhost/');
  }

  // Show window when ready
  mainWindow.once('ready-to-show', () => {
    mainWindow?.show();
    // DevTools can be opened manually via View menu (Cmd+Option+I)
  });

  // Configure webview permissions for embedding external apps
  mainWindow.webContents.on('will-attach-webview', (event, webPreferences, params) => {
    // Allow webviews to load external content without security restrictions
    // This is needed for embedding apps like Linear, Notion, Claude, etc.
    webPreferences.webSecurity = false;
    webPreferences.allowRunningInsecureContent = true;

    // Webview attaching - security relaxed for external app embedding
  });

  // Handle external links
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    // Allow opening external URLs in the default browser
    if (url.startsWith('http://') || url.startsWith('https://')) {
      shell.openExternal(url);
      return { action: 'deny' };
    }
    return { action: 'allow' };
  });

  // Handle navigation - allow OAuth flows to stay within Electron
  // Without this, OAuth redirects would open in external browser
  mainWindow.webContents.on('will-navigate', (event, url) => {
    const appUrl = isDev ? 'http://localhost:5173' : `file://${__dirname}`;

    // URLs that should be allowed to navigate within Electron
    const allowedPatterns = [
      appUrl,
      'file://',
      'http://localhost:8001',  // Local backend OAuth
      'http://localhost:18080', // Alternative local backend port
      'https://businessos-api', // Cloud backend
      'https://accounts.google.com', // Google OAuth
      'https://www.google.com/accounts', // Google OAuth alternative
      'https://oauth', // Generic OAuth
    ];

    const isAllowed = allowedPatterns.some(pattern => url.startsWith(pattern));

    if (!isAllowed) {
      // Open truly external URLs in system browser
      event.preventDefault();
      shell.openExternal(url);
    }
  });

  // Handle window close
  mainWindow.on('close', (event) => {
    // On macOS, hide the window instead of quitting
    if (process.platform === 'darwin') {
      event.preventDefault();
      mainWindow?.hide();
    }
  });

  // Clean up reference when window is closed
  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  // Remember window state
  mainWindow.on('resize', saveWindowState);
  mainWindow.on('move', saveWindowState);

  // Restore window state
  restoreWindowState();

  return mainWindow;
}

/**
 * Save window state to localStorage (via IPC or electron-store)
 */
function saveWindowState(): void {
  if (!mainWindow) return;

  const bounds = mainWindow.getBounds();
  const state = {
    x: bounds.x,
    y: bounds.y,
    width: bounds.width,
    height: bounds.height,
    isMaximized: mainWindow.isMaximized(),
    isFullScreen: mainWindow.isFullScreen(),
  };

  // Store state (could use electron-store for persistence)
  // For now, we'll send it to the renderer to store in localStorage
  mainWindow.webContents.send('window:save-state', state);
}

/**
 * Restore window state from storage
 */
function restoreWindowState(): void {
  // Request stored state from renderer
  // The renderer will respond via IPC if it has stored state
}

/**
 * Focus or create the main window
 */
export function focusMainWindow(): void {
  if (mainWindow) {
    if (mainWindow.isMinimized()) {
      mainWindow.restore();
    }
    mainWindow.focus();
  } else {
    createMainWindow();
  }
}

/**
 * Send a message to the main window
 */
export function sendToMainWindow(channel: string, ...args: unknown[]): void {
  if (mainWindow && !mainWindow.isDestroyed()) {
    mainWindow.webContents.send(channel, ...args);
  }
}
