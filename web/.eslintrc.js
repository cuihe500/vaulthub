module.exports = {
  env: {
    browser: true,
    es2021: true,
    node: true
  },
  extends: [
    'eslint:recommended',
    'plugin:vue/vue3-recommended'
  ],
  parserOptions: {
    ecmaVersion: 'latest',
    sourceType: 'module'
  },
  plugins: ['vue'],
  rules: {
    // 禁止使用var
    'no-var': 'error',
    // 优先使用const
    'prefer-const': 'error',
    // 禁止console（开发环境warn，生产环境error）
    'no-console': process.env.NODE_ENV === 'production' ? 'error' : 'warn',
    // 禁止debugger（开发环境warn，生产环境error）
    'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'warn',
    // 组件名必须多个单词
    'vue/multi-word-component-names': 'off',
    // 限制组件最大属性数量
    'vue/max-attributes-per-line': ['error', {
      singleline: 3,
      multiline: 1
    }],
    // 强制props定义类型
    'vue/require-prop-types': 'error',
    // 强制v-for使用key
    'vue/require-v-for-key': 'error',
    // 禁止在v-html中使用用户输入
    'vue/no-v-html': 'warn'
  }
}
