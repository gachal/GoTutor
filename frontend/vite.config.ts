import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// `base: './'` is load-bearing: packaged Electron loads the SPA via
// file://, where absolute paths (/assets/foo.js) break. Relative paths
// work in both dev (vite serves /) and prod (file://).
export default defineConfig({
  plugins: [vue()],
  base: './',
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    // Code-split Monaco so the chapter list view doesn't pay its cost.
    rollupOptions: {
      output: {
        manualChunks: {
          monaco: ['monaco-editor', '@guolao/vue-monaco-editor'],
        },
      },
    },
  },
})
