import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    hmr: {
      // Configure HMR to work through the Go proxy
      clientPort: 3000,
    },
  },
})
