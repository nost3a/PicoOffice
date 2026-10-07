import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import './assets/main.css'
import { useNetStore } from '@/stores/net'
import { setupI18n } from '@/i18n'

const app = createApp(App)

// register all icons globally, no per-component imports
for (const [key, comp] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, comp)
}

const pinia = createPinia()
app.use(pinia)
app.use(router)
app.use(ElementPlus)
setupI18n(app)
app.mount('#app')

// offline state: listen to network + pending count
useNetStore(pinia).init()

// register hand-rolled service worker
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch(() => {
      // registration failure does not block normal use
    })
  })
}
