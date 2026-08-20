import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5175,
    proxy: {
      '/data': 'http://127.0.0.1:8642',
      '/events': 'http://127.0.0.1:8642',
    },
  },
  build: {
    outDir: path.resolve(__dirname, '../internal/server/dist'),
    emptyOutDir: true,
  },
})
