import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import fs from 'fs'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3000,
    host: '0.0.0.0',
    https: {
      key: fs.readFileSync(path.resolve(__dirname, '../backend/certs/server.key')),
      cert: fs.readFileSync(path.resolve(__dirname, '../backend/certs/server.crt')),
    },
    proxy: {
      '/api': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false, // Accept self-signed certs
      },
      '/v1': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false,
      },
      '/.well-known': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
