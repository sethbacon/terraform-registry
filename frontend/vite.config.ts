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
      // Proxy the Swagger spec to the backend. /api-docs and /api-docs/ are React
      // Router SPA routes and must NOT be proxied — Vite's SPA fallback handles them.
      '/swagger.json': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false,
      },
      '/swagger.yaml': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false,
      },
      // Use /api/ (with slash) so this prefix only matches /api/v1/...
      // and does NOT accidentally match /api-docs which is a React Router route.
      '/api/': {
        target: 'https://localhost:443',
        changeOrigin: true,
        secure: false,
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
