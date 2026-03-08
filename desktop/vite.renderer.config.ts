import { defineConfig } from 'vite';
import path from 'path';
import fs from 'fs';

// https://vitejs.dev/config
// This config serves the pre-built SvelteKit files without transformation
export default defineConfig({
  root: path.resolve(__dirname, 'src/renderer'),
  // Disable all processing - serve files as-is
  optimizeDeps: {
    exclude: ['**/*'],
  },
  server: {
    // Serve static files without transformation
    fs: {
      strict: false,
      allow: [path.resolve(__dirname, 'src/renderer')],
    },
  },
  build: {
    outDir: path.resolve(__dirname, '.vite/renderer/main_window'),
    emptyOutDir: true,
    // Minimal build - Vite will create index.html entry but we'll overwrite
    rollupOptions: {
      // Use index.html as entry to satisfy Vite
      input: path.resolve(__dirname, 'src/renderer/index.html'),
      // Disable code splitting which causes the ../chunks issue
      output: {
        manualChunks: undefined,
      }
    },
    // Disable minification to avoid transforming the already-built code
    minify: false,
    // Don't process CSS
    cssCodeSplit: false,
  },
  plugins: [
    {
      name: 'serve-static-assets',
      configureServer(server) {
        // Serve static files directly without processing
        server.middlewares.use((req, res, next) => {
          // Let Vite handle the request normally for static files
          next();
        });
      },
    },
    {
      name: 'copy-sveltekit-assets',
      closeBundle: async () => {
        // Copy the _app folder which contains all the pre-built SvelteKit assets
        const srcAppDir = path.resolve(__dirname, 'src/renderer/_app');
        const destAppDir = path.resolve(__dirname, '.vite/renderer/main_window/_app');

        const copyRecursive = (src: string, dest: string) => {
          if (!fs.existsSync(src)) return;

          if (fs.statSync(src).isDirectory()) {
            if (!fs.existsSync(dest)) {
              fs.mkdirSync(dest, { recursive: true });
            }
            for (const file of fs.readdirSync(src)) {
              copyRecursive(path.join(src, file), path.join(dest, file));
            }
          } else {
            fs.copyFileSync(src, dest);
          }
        };

        copyRecursive(srcAppDir, destAppDir);

        // Also copy other static assets
        const staticDirs = ['downloads'];
        for (const dir of staticDirs) {
          const srcDir = path.resolve(__dirname, 'src/renderer', dir);
          const destDir = path.resolve(__dirname, '.vite/renderer/main_window', dir);
          copyRecursive(srcDir, destDir);
        }

        // Copy root files
        const rootFiles = ['index.html', 'robots.txt', 'osa-logo.png'];
        for (const file of rootFiles) {
          const src = path.resolve(__dirname, 'src/renderer', file);
          const dest = path.resolve(__dirname, '.vite/renderer/main_window', file);
          if (fs.existsSync(src)) {
            fs.copyFileSync(src, dest);
          }
        }

        console.log('Copied SvelteKit assets to renderer output');
      }
    }
  ],
});
