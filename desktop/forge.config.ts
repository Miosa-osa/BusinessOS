import type { ForgeConfig } from '@electron-forge/shared-types';
import { MakerSquirrel } from '@electron-forge/maker-squirrel';
import { MakerZIP } from '@electron-forge/maker-zip';
import { MakerDeb } from '@electron-forge/maker-deb';
import { MakerRpm } from '@electron-forge/maker-rpm';
import { MakerDMG } from '@electron-forge/maker-dmg';
import { VitePlugin } from '@electron-forge/plugin-vite';
import { PublisherGithub } from '@electron-forge/publisher-github';
import { execSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';

const config: ForgeConfig = {
  packagerConfig: {
    // Disable asar to fix native module loading (better-sqlite3)
    // TODO: Re-enable with proper unpack config once working
    asar: false,
    icon: './resources/icons/icon',
    appBundleId: 'com.businessos.desktop',
    appCopyright: 'Copyright © 2025 BusinessOS',
    extraResource: [
      './resources/bin'
    ],
    // Merge additional Info.plist settings (for permission descriptions)
    extendInfo: './resources/Info.plist',
    // Code signing for macOS (configure via environment variables)
    osxSign: process.env.APPLE_ID ? {
      identity: process.env.APPLE_IDENTITY,
      // Use custom entitlements for required permissions
      entitlements: './resources/entitlements.mac.plist',
      'entitlements-inherit': './resources/entitlements.mac.plist',
      // Enable hardened runtime (required for notarization)
      hardenedRuntime: true,
      // Gate keeper will check these
      'gatekeeper-assess': false,
    } : undefined,
    osxNotarize: process.env.APPLE_ID ? {
      appleId: process.env.APPLE_ID,
      appleIdPassword: process.env.APPLE_PASSWORD,
      teamId: process.env.APPLE_TEAM_ID,
    } : undefined,
  },
  rebuildConfig: {},
  hooks: {
    postPackage: async (config, packageResult) => {
      // Copy native node_modules to the packaged app
      // This is needed because Vite externalizes native modules but forge doesn't copy them
      // better-sqlite3 needs 'bindings' to locate its .node file
      const nativeModules = ['better-sqlite3', 'bindings', 'file-uri-to-path'];

      for (const outputPath of packageResult.outputPaths) {
        const appPath = path.join(outputPath, 'BusinessOS.app', 'Contents', 'Resources', 'app');
        const nodeModulesPath = path.join(appPath, 'node_modules');

        // Create node_modules directory
        if (!fs.existsSync(nodeModulesPath)) {
          fs.mkdirSync(nodeModulesPath, { recursive: true });
        }

        // Copy each native module
        for (const moduleName of nativeModules) {
          const srcPath = path.join(__dirname, 'node_modules', moduleName);
          const destPath = path.join(nodeModulesPath, moduleName);

          if (fs.existsSync(srcPath)) {
            console.log(`Copying ${moduleName} to packaged app...`);
            execSync(`cp -R "${srcPath}" "${destPath}"`);
          }
        }
      }
    },
  },
  makers: [
    new MakerSquirrel({
      name: 'BusinessOS',
      setupIcon: './resources/icons/icon.ico',
    }),
    new MakerZIP({}, ['darwin']),
    new MakerDMG({
      icon: './resources/icons/icon.icns',
      background: './resources/dmg-background.png',
    }),
    new MakerRpm({
      options: {
        icon: './resources/icons/icon.png',
      },
    }),
    new MakerDeb({
      options: {
        icon: './resources/icons/icon.png',
        maintainer: 'BusinessOS Team',
        homepage: 'https://businessos.app',
      },
    }),
  ],
  plugins: [
    // AutoUnpackNativesPlugin removed - asar disabled for native module compatibility
    new VitePlugin({
      // `build` can specify multiple entry builds
      build: [
        {
          // Main process entry point
          entry: 'src/main/index.ts',
          config: 'vite.main.config.ts',
        },
        {
          // Preload script entry point
          entry: 'src/preload/index.ts',
          config: 'vite.preload.config.ts',
        },
      ],
      renderer: [
        {
          name: 'main_window',
          config: 'vite.renderer.config.ts',
        },
      ],
    }),
  ],
  publishers: [
    new PublisherGithub({
      repository: {
        owner: 'your-org',
        name: 'businessos-desktop',
      },
      prerelease: false,
      draft: true,
    }),
  ],
};

export default config;
