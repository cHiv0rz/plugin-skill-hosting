import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import '@fontsource-variable/fraunces/wght.css'
import '@fontsource-variable/fraunces/wght-italic.css'
import '@fontsource-variable/jetbrains-mono/wght.css'
import './styles.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
