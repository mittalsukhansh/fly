import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/path': 'http://localhost:8080',
      '/subgraph': 'http://localhost:8080',
      '/nodes/all': 'http://localhost:8080',
      '/stats': 'http://localhost:8080',
      '/neuron': 'http://localhost:8080'
    }
  }
})
