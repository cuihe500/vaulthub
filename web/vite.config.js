import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { fileURLToPath } from 'url'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { visualizer } from 'rollup-plugin-visualizer'

const __dirname = fileURLToPath(new URL('.', import.meta.url))

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // Element Plus按需自动导入
    AutoImport({
      imports: ['vue', 'vue-router'],
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts'
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts'
    }),
    // Bundle分析工具,构建后生成stats.html
    visualizer({
      open: false,
      gzipSize: true,
      brotliSize: true,
      filename: 'dist/stats.html'
    })
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    // 使用esbuild压缩(Vite默认,速度更快)
    minify: 'esbuild',
    esbuild: {
      drop: ['console', 'debugger']
    },
    // 提高chunk大小警告阈值至900KB (ECharts懒加载chunk约876KB,不影响首屏)
    chunkSizeWarningLimit: 1000,
    rollupOptions: {
      output: {
        // 手动拆分vendor chunk
        manualChunks(id) {
          if (id.includes('node_modules')) {
            // Element Plus单独打包
            if (id.includes('element-plus')) {
              return 'chunk-element-plus'
            }
            // ECharts单独打包
            if (id.includes('echarts')) {
              return 'chunk-echarts'
            }
            // Vue核心库单独打包
            if (id.includes('vue') || id.includes('vue-router') || id.includes('vuex')) {
              return 'chunk-vue'
            }
            // 其他第三方库打包为通用vendor
            return 'chunk-vendor'
          }
        },
        // 自定义chunk文件名
        chunkFileNames: 'assets/js/[name]-[hash].js',
        entryFileNames: 'assets/js/[name]-[hash].js',
        assetFileNames: 'assets/[ext]/[name]-[hash].[ext]'
      }
    }
  }
})
