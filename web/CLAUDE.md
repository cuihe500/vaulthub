# VaultHub 前端开发章程

## 技术栈

- **框架**: Vue 3.x
- **UI库**: Element Plus（Vue 3）
- **构建工具**: Vite
- **HTTP客户端**: Axios
- **路由**: Vue Router
- **状态管理**: Vuex
- **代码规范**: ESLint + Prettier
- **包管理器**: pnpm

## 项目结构

```
web/
├── public/               # 静态资源
│   └── index.html
├── src/
│   ├── api/             # API调用层（唯一与后端交互的地方）
│   │   ├── request.js   # Axios封装
│   │   ├── auth.js      # 认证相关API
│   │   ├── vault.js     # 密钥管理API
│   │   └── user.js      # 用户管理API
│   ├── assets/          # 静态资源（图片、字体等）
│   ├── components/      # 可复用组件
│   │   ├── common/      # 通用基础组件
│   │   └── business/    # 业务组件
│   ├── layouts/         # 布局组件
│   ├── router/          # 路由配置
│   │   └── index.js
│   ├── store/           # 状态管理（仅在必要时使用）
│   │   └── index.js
│   ├── utils/           # 工具函数
│   │   ├── crypto.js    # 前端加密工具
│   │   ├── validate.js  # 表单验证
│   │   └── storage.js   # 本地存储封装
│   ├── views/           # 页面组件
│   │   ├── login/
│   │   ├── vault/
│   │   └── user/
│   ├── App.vue
│   └── main.js
├── .env.development     # 开发环境配置
├── .env.production      # 生产环境配置
├── .eslintrc.js         # ESLint配置
├── .prettierrc          # Prettier配置
├── vite.config.js       # Vite配置
└── package.json
```

## 设计原则

### 1. 简洁至上

**数据结构决定一切**
```javascript
// Bad: 复杂的嵌套状态
state: {
  user: {
    profile: {
      info: {
        data: { name: 'xxx' }
      }
    }
  }
}

// Good: 扁平化数据
state: {
  userId: null,
  userName: '',
  userEmail: ''
}
```

**消除特殊情况**
```javascript
// Bad: 到处都是条件判断
if (type === 'create') { /*...*/ }
else if (type === 'edit') { /*...*/ }
else if (type === 'view') { /*...*/ }

// Good: 数据驱动，统一处理
const MODE_CONFIG = {
  create: { title: '新建', readonly: false },
  edit: { title: '编辑', readonly: false },
  view: { title: '查看', readonly: true }
}
const config = MODE_CONFIG[mode]
```

**组件嵌套不超过3层**
```
页面(View) -> 业务组件(Business) -> 基础组件(Common)
```
超过3层说明设计有问题，重新拆分。

### 2. 清晰的职责分层

```
用户交互 -> 页面组件(Views) -> 业务组件(Components) -> API层 -> 后端
```

- **Views**: 页面布局、路由、组合业务组件
- **Components**: 可复用逻辑、UI交互、数据展示
- **API层**: 唯一与后端交互的地方，封装所有HTTP请求
- **Utils**: 纯函数工具，无副作用

**铁律**: 组件不直接调用axios，必须通过API层。

### 3. 状态管理原则

**能用Props/Emit解决的，不用Vuex/Pinia**

状态管理只用于：
- 跨页面共享的用户信息（userId、token等）
- 全局配置（主题、语言等）
- 复杂的跨组件通信

**不要把所有数据都塞进Store**，大部分组件内部状态用`data`即可。

### 4. 安全第一

**前端永远不可信**，但必须做好第一道防线：

1. **输入验证**: 所有用户输入必须验证（防XSS）
2. **敏感数据**: Token存储使用httpOnly Cookie或加密localStorage
3. **密码显示**: 密码输入框type="password"，提供显示/隐藏切换
4. **HTTPS**: 生产环境强制HTTPS
5. **CSP**: 配置Content-Security-Policy

```javascript
// Bad: 直接渲染用户输入
<div v-html="userInput"></div>

// Good: 使用文本插值或DOMPurify清洗
<div>{{ userInput }}</div>
```

## 编码规范

### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserProfile.vue` |
| 组件名 | PascalCase | `<UserProfile />` |
| 变量/函数 | camelCase | `getUserInfo()` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| CSS类名 | kebab-case | `user-profile` |
| 文件夹 | kebab-case | `user-management/` |

### 组件设计

**单一职责**
```vue
<!-- Bad: 一个组件干太多事 -->
<template>
  <div>
    <user-form />
    <user-list />
    <user-stats />
  </div>
</template>

<!-- Good: 拆分成独立组件 -->
<!-- UserManagement.vue -->
<template>
  <div>
    <UserForm @submit="handleSubmit" />
    <UserList :users="users" />
  </div>
</template>
```

**Props定义必须明确类型**
```javascript
// Bad
props: ['user', 'type']

// Good
props: {
  user: {
    type: Object,
    required: true
  },
  type: {
    type: String,
    default: 'view',
    validator: v => ['view', 'edit'].includes(v)
  }
}
```

**事件命名使用kebab-case**
```javascript
// Bad
this.$emit('updateUser', data)

// Good
this.$emit('update-user', data)
```

### API调用规范

**统一封装Axios**
```javascript
// api/request.js
import axios from 'axios'
import { getToken } from '@/utils/storage'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 10000
})

// 请求拦截: 添加Token
request.interceptors.request.use(config => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截: 统一处理错误
request.interceptors.response.use(
  response => {
    const { code, data, message } = response.data
    if (code !== 200) {
      // 根据后端Base类型的Code字段处理错误
      console.error(message)
      return Promise.reject(new Error(message))
    }
    return data
  },
  error => {
    console.error('请求失败:', error.message)
    return Promise.reject(error)
  }
)

export default request
```

**API按模块组织**
```javascript
// api/vault.js
import request from './request'

export const getVaultList = (params) => {
  return request.get('/api/v1/vaults', { params })
}

export const createVault = (data) => {
  return request.post('/api/v1/vaults', data)
}

export const updateVault = (uuid, data) => {
  return request.put(`/api/v1/vaults/${uuid}`, data)
}

export const deleteVault = (uuid) => {
  return request.delete(`/api/v1/vaults/${uuid}`)
}
```

**组件中使用**
```vue
<script>
import { getVaultList, createVault } from '@/api/vault'

export default {
  methods: {
    async loadVaults() {
      try {
        this.loading = true
        this.vaults = await getVaultList({ page: 1, size: 10 })
      } catch (error) {
        this.$message.error('加载失败')
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
```

### 错误处理

**必须处理所有异步操作的错误**
```javascript
// Bad: 吞掉错误
async loadData() {
  const data = await fetchData()
}

// Good: 明确处理
async loadData() {
  try {
    this.data = await fetchData()
  } catch (error) {
    console.error('加载数据失败:', error)
    this.$message.error('加载失败，请重试')
  }
}
```

### 性能优化

1. **列表渲染必须加key**
```vue
<div v-for="item in list" :key="item.id">
  {{ item.name }}
</div>
```

2. **大列表使用虚拟滚动**（如el-table的虚拟滚动模式）

3. **图片懒加载**
```vue
<el-image lazy :src="imageUrl" />
```

4. **路由懒加载**
```javascript
const UserManagement = () => import('@/views/user/UserManagement.vue')
```

## 开发工作流

### 1. 新增页面

1. 在`views/`下创建页面组件
2. 在`router/index.js`注册路由
3. 在`api/`下添加对应的API调用
4. 根据需要拆分业务组件到`components/business/`

### 2. 新增API

1. 查看后端Swagger文档，确认接口契约
2. 在对应的API模块文件中添加方法
3. 使用TypeScript的话，定义请求/响应类型

### 3. 调试

1. 使用Vue DevTools查看组件状态
2. 使用浏览器Network面板检查API调用
3. 不要用`console.log`调试生产代码（提交前删除）

### 4. 提交代码

1. ESLint检查通过: `npm run lint`
2. 本地构建成功: `npm run build`
3. 提交前检查是否有敏感信息（Token、密码等）

## 与后端对接规范

### 1. 接口契约

- 所有接口返回HTTP 200
- 响应格式统一为Base类型:
```json
{
  "code": 200,
  "message": "success",
  "data": {}
}
```
- 错误状态通过`code`字段判断，不使用HTTP状态码

### 2. 时间处理

- 后端存储UTC时间，前端展示时转换为`Asia/Shanghai`
- 使用Day.js或date-fns处理时间格式化

```javascript
import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

dayjs.extend(utc)
dayjs.extend(timezone)

// 后端返回UTC时间，前端显示本地时间
const displayTime = dayjs.utc(utcTime).tz('Asia/Shanghai').format('YYYY-MM-DD HH:mm:ss')
```

### 3. 认证机制

- 使用JWT认证
- Token存储在localStorage或Cookie
- 每次请求在Header中携带: `Authorization: Bearer <token>`
- Token过期后跳转登录页

## 禁止事项

1. **不要随意引入新依赖**: 需要先评估必要性
2. **不要在组件中直接操作DOM**: 用Vue的数据驱动
3. **不要把业务逻辑写在模板里**: 超过3个三元运算符就该提取成computed
4. **不要把所有代码写在一个文件**: 单个组件超过300行就该拆分
5. **不要用var**: 统一使用const/let
6. **不要在循环中使用index作为key**: 使用唯一标识符(如uuid)

## 代码审查标准

提交代码前自问：

1. 这个组件是否只做了一件事？
2. 数据流是否清晰（谁拥有、谁修改）？
3. 有没有不必要的嵌套和条件判断？
4. 错误处理是否完整？
5. 有没有安全隐患（XSS、敏感数据泄露）？
6. 能否用更少的代码实现？

**如果答案有任何疑问，重构它。**

## 样式规范

### 设计哲学

**简约大气 = 少即是多**

- 用最少的颜色表达最丰富的语义
- 用统一的间距建立视觉秩序
- 用克制的动画提升体验
- 消除所有魔法数字，一切可追溯

### 颜色系统

**所有颜色必须使用CSS变量，禁止硬编码颜色值**

```css
/* 主色调 - 深蓝紫渐变系（科技、专业、可信） */
--color-primary: #667eea;         /* 主色 */
--color-primary-light: #7c8df0;   /* 主色-浅 */
--color-primary-dark: #5568d3;    /* 主色-深 */
--color-primary-bg: #f0f2ff;      /* 主色-背景 */

/* 辅助色 - 紫色系 */
--color-secondary: #764ba2;       /* 辅助色 */

/* 功能色 - 语义明确 */
--color-success: #10b981;         /* 成功/安全 */
--color-warning: #f59e0b;         /* 警告 */
--color-error: #ef4444;           /* 错误/危险 */
--color-info: #3b82f6;            /* 信息 */

/* 中性色 - 灰度系统（只用5个层级） */
--color-text-primary: #1f2937;    /* 主要文字 */
--color-text-secondary: #6b7280;  /* 次要文字 */
--color-text-disabled: #9ca3af;   /* 禁用文字 */
--color-border: #e5e7eb;          /* 边框 */
--color-bg: #f9fafb;              /* 背景 */
--color-white: #ffffff;           /* 纯白 */

/* 阴影色 */
--color-shadow: rgba(0, 0, 0, 0.1);
```

**使用规则**：
```css
/* Bad: 硬编码颜色 */
.button {
  background: #667eea;
  color: #ffffff;
}

/* Good: 使用变量 */
.button {
  background: var(--color-primary);
  color: var(--color-white);
}
```

### 字体系统

**字体族**：优先使用系统字体，零加载时间
```css
--font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto,
               'Helvetica Neue', Arial, 'Noto Sans', sans-serif,
               'Apple Color Emoji', 'Segoe UI Emoji';

--font-family-mono: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono',
                    Consolas, monospace;
```

**字号系统**：基于4px倍数，只用5个层级
```css
--font-size-xs: 12px;    /* 辅助文字 */
--font-size-sm: 14px;    /* 次要文字 */
--font-size-base: 16px;  /* 正文（基准） */
--font-size-lg: 18px;    /* 小标题 */
--font-size-xl: 24px;    /* 大标题 */
```

**行高**：统一比例
```css
--line-height-tight: 1.25;   /* 标题 */
--line-height-normal: 1.5;   /* 正文 */
--line-height-loose: 1.75;   /* 松散文本 */
```

**字重**：只用3个
```css
--font-weight-normal: 400;
--font-weight-medium: 500;
--font-weight-bold: 600;
```

### 间距系统

**8px基础单位，只用倍数，消除随意间距**

```css
--spacing-xs: 4px;     /* 0.5倍 */
--spacing-sm: 8px;     /* 1倍 - 基础单位 */
--spacing-md: 16px;    /* 2倍 */
--spacing-lg: 24px;    /* 3倍 */
--spacing-xl: 32px;    /* 4倍 */
--spacing-2xl: 48px;   /* 6倍 */
--spacing-3xl: 64px;   /* 8倍 */
```

**使用规则**：
```css
/* Bad: 随意的数字 */
.card {
  padding: 18px 22px;
  margin-bottom: 15px;
}

/* Good: 使用间距系统 */
.card {
  padding: var(--spacing-md) var(--spacing-lg);
  margin-bottom: var(--spacing-md);
}
```

### 圆角系统

```css
--radius-sm: 4px;      /* 小圆角：按钮、输入框 */
--radius-md: 8px;      /* 中圆角：卡片 */
--radius-lg: 12px;     /* 大圆角：弹窗 */
--radius-full: 9999px; /* 全圆角：标签、头像 */
```

### 阴影系统

**分3层，表达层级关系**

```css
--shadow-sm: 0 1px 2px 0 var(--color-shadow);              /* 轻微抬起 */
--shadow-md: 0 4px 6px -1px var(--color-shadow);           /* 卡片 */
--shadow-lg: 0 10px 15px -3px var(--color-shadow);         /* 弹窗/抽屉 */
```

### 动画系统

**克制的动画，只用于反馈交互**

```css
--duration-fast: 150ms;      /* 快速反馈：按钮hover */
--duration-base: 250ms;      /* 标准过渡：展开收起 */
--duration-slow: 350ms;      /* 慢速过渡：弹窗进出 */

--easing: cubic-bezier(0.4, 0, 0.2, 1);  /* 标准缓动 */
```

**使用规则**：
```css
/* 统一的过渡 */
.button {
  transition: all var(--duration-fast) var(--easing);
}

.modal {
  transition: opacity var(--duration-slow) var(--easing);
}
```

### 布局原则

**响应式断点**：基于主流设备
```css
--breakpoint-sm: 640px;    /* 手机 */
--breakpoint-md: 768px;    /* 平板 */
--breakpoint-lg: 1024px;   /* 笔记本 */
--breakpoint-xl: 1280px;   /* 桌面 */
```

**容器宽度**：
```css
--container-sm: 640px;
--container-md: 768px;
--container-lg: 1024px;
--container-xl: 1280px;
```

**层级管理**：统一z-index
```css
--z-dropdown: 1000;   /* 下拉菜单 */
--z-sticky: 1020;     /* 吸顶元素 */
--z-modal: 1040;      /* 弹窗 */
--z-popover: 1060;    /* 气泡 */
--z-toast: 1080;      /* 提示 */
```

### 全局样式文件

创建 `src/assets/styles/variables.css` 统一管理：

```css
:root {
  /* 颜色 */
  --color-primary: #667eea;
  --color-primary-light: #7c8df0;
  --color-primary-dark: #5568d3;
  --color-primary-bg: #f0f2ff;
  --color-secondary: #764ba2;

  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-error: #ef4444;
  --color-info: #3b82f6;

  --color-text-primary: #1f2937;
  --color-text-secondary: #6b7280;
  --color-text-disabled: #9ca3af;
  --color-border: #e5e7eb;
  --color-bg: #f9fafb;
  --color-white: #ffffff;
  --color-shadow: rgba(0, 0, 0, 0.1);

  /* 字体 */
  --font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  --font-family-mono: 'SF Mono', Monaco, Consolas, monospace;

  --font-size-xs: 12px;
  --font-size-sm: 14px;
  --font-size-base: 16px;
  --font-size-lg: 18px;
  --font-size-xl: 24px;

  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-bold: 600;

  --line-height-tight: 1.25;
  --line-height-normal: 1.5;
  --line-height-loose: 1.75;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
  --spacing-2xl: 48px;
  --spacing-3xl: 64px;

  /* 圆角 */
  --radius-sm: 4px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-full: 9999px;

  /* 阴影 */
  --shadow-sm: 0 1px 2px 0 var(--color-shadow);
  --shadow-md: 0 4px 6px -1px var(--color-shadow);
  --shadow-lg: 0 10px 15px -3px var(--color-shadow);

  /* 动画 */
  --duration-fast: 150ms;
  --duration-base: 250ms;
  --duration-slow: 350ms;
  --easing: cubic-bezier(0.4, 0, 0.2, 1);

  /* 层级 */
  --z-dropdown: 1000;
  --z-sticky: 1020;
  --z-modal: 1040;
  --z-popover: 1060;
  --z-toast: 1080;
}

/* 全局基础样式 */
* {
  box-sizing: border-box;
}

body {
  font-family: var(--font-family);
  font-size: var(--font-size-base);
  line-height: var(--line-height-normal);
  color: var(--color-text-primary);
  background-color: var(--color-bg);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
```

### 组件样式规范

**1. 使用scoped避免污染**
```vue
<style scoped>
.card { /* ... */ }
</style>
```

**2. BEM命名法（组件内部）**
```css
/* Block */
.user-card { }

/* Element */
.user-card__header { }
.user-card__body { }

/* Modifier */
.user-card--highlighted { }
```

**3. 禁止深度选择器**（除非必要覆盖Element Plus）
```css
/* Bad: 破坏封装 */
::v-deep .el-button { }

/* Good: 通过props传递类名 */
<el-button :class="customClass" />
```

### 样式审查标准

提交前检查：

1. **是否使用了魔法数字？** 所有数字应该来自变量系统
2. **是否硬编码颜色？** 必须使用CSS变量
3. **间距是否是8的倍数？** 必须符合间距系统
4. **是否过度使用动画？** 只在必要时添加
5. **响应式是否考虑？** 大于768px的设计需要适配

**记住：每个CSS规则都应该有明确的理由，否则删掉它。**

## 总结

记住：

- **数据结构决定代码质量** - 先想清楚数据怎么组织
- **消除特殊情况** - 不要用if/else打补丁
- **组件职责单一** - 一个组件只做一件事
- **简洁胜过clever** - 代码是给人看的，不是炫技
- **安全永远第一** - 不信任任何用户输入

## 注意（必须遵守）
1. 在调度接口的时候，一定要先查看api-docs/下的接口文档，确认接口的定义。
2. 任何需要现在时间的地方，都必须调用date命令获取真实的现在时间。
3. 优先使用MCP能力（如搜索，第三方库检索）等，若MCP无法使用，再回退到原始模式。
4. 在引用第三方库之前，必须完全了解该库的情况，包括但不限于实际能力、接口规范等。
5. 若该目录必须遵守其他相关的规定，则可以在该目录下新增README.md，该文档将会视为和章程一个优先级的规定。
6. 若目录下存在README.md，必须先阅读并理解该文档内容，并且将其视为和章程一个优先级。