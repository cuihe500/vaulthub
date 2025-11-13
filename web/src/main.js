import { createApp } from 'vue'
// Element Plus组件通过unplugin按需自动导入
// 但Message等函数式API需要手动导入样式
import 'element-plus/es/components/message/style/css'
import 'element-plus/es/components/message-box/style/css'
import '@/assets/styles/variables.css'
import App from './App.vue'
import router from './router'
import store from './store'

const app = createApp(App)

app.use(router)
app.use(store)

app.mount('#app')
