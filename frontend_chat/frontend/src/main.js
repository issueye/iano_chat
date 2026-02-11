import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'

// Check if running in Wails environment
if (window.go && window.go.main && window.go.main.App) {
  console.log('Running in Wails environment')
} else {
  console.log('Running in browser mode')
}

const app = createApp(App)
app.use(createPinia())
app.mount('#app')
