import './style.css'
import App from './App.svelte'
import { initializeDevTools } from './tools/index.js'

// 初始化开发工具集
initializeDevTools()

const app = new App({
  target: document.getElementById('app')
})

export default app
