import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'node:path'

// proxy /api to local backend
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    chunkSizeWarningLimit: 1500,
    rollupOptions: {
      output: {
        manualChunks: {
          vue: ['vue', 'vue-router', 'pinia'],
          element: ['element-plus', '@element-plus/icons-vue'],
          tiptap: [
            '@tiptap/vue-3',
            '@tiptap/starter-kit',
            '@tiptap/extension-text-style',
            '@tiptap/extension-color',
            '@tiptap/extension-image',
            '@tiptap/extension-table',
            '@tiptap/extension-table-row',
            '@tiptap/extension-table-cell',
            '@tiptap/extension-table-header'
          ]
        }
      }
    }
  }
})
