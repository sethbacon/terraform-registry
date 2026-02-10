import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'
import fs from 'fs'

// Only read certs when they exist (skipped during Docker build)
const certPath = path.resolve(__dirname, '../backend/certs/server.crt')
const keyPath = path.resolve(__dirname, '../backend/certs/server.key')
const certsExist = fs.existsSync(certPath) && fs.existsSync(keyPath)

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
    host: 'registry.local',
    ...(certsExist ? {
      https: {
        key: fs.readFileSync(keyPath),
        cert: fs.readFileSync(certPath),
      },
    } : {}),
    proxy: {
      '/api': {
        target: 'https://registry.local:443',
        changeOrigin: true,
        secure: false, // Accept self-signed certs
      },
      '/v1': {
        target: 'https://registry.local:443',
        changeOrigin: true,
        secure: false,
      },
      '/.well-known': {
        target: 'https://registry.local:443',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
