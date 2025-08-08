import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  // Fix for crypto.hash not being a function error in Node.js 18.x
  optimizeDeps: {
    esbuildOptions: {
      nodePaths: [],
      define: {
        global: 'globalThis',
      },
    },
  },
  // Additional fix for crypto module issues in Node.js environments
  server: {
    fs: {
      allow: ['..'],
    },
  },
})
