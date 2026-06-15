import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// `base: './'` is load-bearing: packaged Electron loads the SPA via
// file://, where absolute paths (/assets/foo.js) break. Relative paths
// work in both dev (vite serves /) and prod (file://).
//
// No /api proxy: the frontend talks to the backend directly at
// http://localhost:8081 (absolute URL in src/api/client.ts).
export default defineConfig({
  plugins: [vue()],
  base: './',
  server: {
    port: 5173,
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          monaco: ['monaco-editor', '@guolao/vue-monaco-editor'],
        },
      },
    },
  },
})
