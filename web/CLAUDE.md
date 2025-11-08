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

## 总结

记住：

- **数据结构决定代码质量** - 先想清楚数据怎么组织
- **消除特殊情况** - 不要用if/else打补丁
- **组件职责单一** - 一个组件只做一件事
- **简洁胜过clever** - 代码是给人看的，不是炫技
- **安全永远第一** - 不信任任何用户输入

## 注意（必须遵守）
1. 在调度接口的时候，一定要先查看api-docs/下的接口文档，确认接口的定义
2. 任何需要现在时间的地方，都必须调用date命令获取真实的现在时间。
3. 优先使用MCP能力（如搜索，第三方库检索）等，若MCP无法使用，再回退到原始模式。
4. 在引用第三方库之前，必须完全了解该库的情况，包括但不限于实际能力、接口规范等。