import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import wails from '@wailsio/runtime/plugins/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), wails('./bindings')],
  build: {
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html'),
        timer: resolve(__dirname, 'timer.html'),
      },
    },
  },
})
