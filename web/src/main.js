import { createApp } from 'vue'
// Element Plus组件和样式现在通过unplugin按需自动导入,无需手动引入
import '@/assets/styles/variables.css'
import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

app.use(router)
app.use(store)

app.mount('#app')
