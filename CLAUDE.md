# VaultHub 开发章程

## 项目元数据（不可更改）

**作者**：Changhe Cui
**邮箱**：admin@thankseveryone.top
**代码库**：https://github.com/cuihe500/vaulthub
**许可证**：Apache 2.0

VaultHub 是一个密钥管理系统，旨在安全地存储、管理和轮换加密密钥、API 密钥及其他敏感凭证。

---

## 技术栈

### 后端技术栈

- **Go**: 1.25.1
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MariaDB/MySQL
- **配置管理**: Viper (TOML格式)
- **CLI**: Cobra
- **权限控制**: Casbin
- **认证**: JWT (golang-jwt/jwt/v5)
- **缓存**: Redis (go-redis/v9)
- **日志**: Zap
- **接口文档**: Swagger
- **任务调度**: robfig/cron/v3
- **数据库迁移**: golang-migrate/migrate/v4
- **数据验证**: go-playground/validator/v10
- **助记词生成**: tyler-smith/go-bip39

### 前端技术栈

- **框架**: Vue 3.x
- **UI库**: Element Plus
- **构建工具**: Vite
- **HTTP客户端**: Axios
- **路由**: Vue Router
- **状态管理**: Vuex
- **代码规范**: ESLint + Prettier
- **包管理器**: pnpm

---

## 项目结构

```
vaulthub/
├── cmd/
│   └── vaulthub/          # 应用程序入口
│       └── main.go        # 主程序
├── internal/              # 私有应用代码
│   ├── api/
│   │   ├── handlers/      # HTTP 请求处理器
│   │   ├── middleware/    # 中间件
│   │   └── routes/        # 路由定义
│   ├── app/               # 应用管理器
│   │   ├── manager.go     # 全局管理器（DB、Redis、Casbin等）
│   │   ├── init.go        # 应用初始化逻辑
│   │   └── scheduler.go   # 定时任务调度器
│   ├── config/            # 配置管理
│   │   ├── config.go      # 配置结构定义
│   │   └── manager.go     # 配置管理器（热更新、监控）
│   ├── database/          # 数据库连接和操作
│   │   ├── migrations/    # 数据库迁移
│   │   └── models/        # 数据模型
│   └── service/           # 业务逻辑层
├── pkg/                   # 可公开使用的库代码
│   ├── crypto/            # 加密工具
│   ├── email/             # 邮件通知与安全告警工具
│   ├── errors/            # 统一错误处理
│   ├── jwt/               # JWT工具
│   ├── logger/            # 日志工具
│   ├── redis/             # Redis工具
│   ├── response/          # 统一响应格式
│   ├── validator/         # 数据验证工具
│   └── version/           # 版本信息
├── web/                   # 前端资源
│   ├── public/            # 静态资源
│   ├── src/
│   │   ├── api/           # API调用层（唯一与后端交互的地方）
│   │   ├── assets/        # 静态资源（图片、字体等）
│   │   ├── components/    # 可复用组件
│   │   │   ├── common/    # 通用基础组件
│   │   │   └── business/  # 业务组件
│   │   ├── layouts/       # 布局组件
│   │   ├── router/        # 路由配置
│   │   ├── store/         # 状态管理
│   │   ├── utils/         # 工具函数
│   │   ├── views/         # 页面组件
│   │   ├── App.vue
│   │   └── main.js
│   └── vite.config.js
├── configs/               # 配置文件
│   ├── config.toml        # 主配置文件
│   ├── config.toml.example
│   ├── rbac_model.conf    # Casbin权限模型
│   └── .env.example
├── build/                 # 构建输出目录
├── docs/                  # 文档
│   └── swagger/           # Swagger接口文档
├── scripts/               # 构建和部署脚本
├── api-test.http          # HTTP 接口调试脚本
├── go.mod
└── go.sum
```

---

## 核心设计原则

### 1. 简洁至上

**数据结构决定一切**
- 单一职责：每个包、每个函数、每个组件只做一件事
- 避免过度抽象：不为"可能的需求"写代码
- 控制嵌套深度：函数/组件不超过3层缩进
- 扁平化数据结构，消除不必要的嵌套

**消除特殊情况**
- 好代码没有特殊情况
- 用数据驱动替代if/else分支
- 重新设计数据结构来消除条件判断

### 2. 清晰的分层架构

**后端分层**：
```
HTTP请求 → 路由 → 处理器 → 服务层 → 数据访问层 → 数据库
```
- **handlers**: 只负责 HTTP 协议相关（请求解析、响应格式）
- **service**: 核心业务逻辑，与 HTTP 无关
- **models/database**: 数据持久化，只关心数据

**前端分层**：
```
用户交互 → 页面组件(Views) → 业务组件(Components) → API层 → 后端
```
- **Views**: 页面布局、路由、组合业务组件
- **Components**: 可复用逻辑、UI交互、数据展示
- **API层**: 唯一与后端交互的地方，封装所有HTTP请求
- **Utils**: 纯函数工具，无副作用

### 3. 安全第一

- 所有敏感数据加密存储
- 输入验证在处理器层完成
- 防止常见漏洞：SQL 注入、XSS、CSRF
- 前端永远不可信，但必须做好第一道防线
- 遵循"最小适用"原则，反复检查安全隐患

### 4. 可测试性

- 业务逻辑与框架解耦
- 依赖注入优于全局变量
- 单元测试覆盖核心逻辑

---

## 后端开发规范

### 编码规范

#### 命名
- 包名：小写，单个单词
- 接口：简洁有力（如 `Reader`, `Writer`）
- 函数/变量：驼峰命名，见名知义
- 常量：大写+下划线或驼峰
- 日志、注释：中文优先且作为一等公民支持

#### 错误处理
```go
// Bad: 吞掉错误
_ = doSomething()

// Good: 明确处理
if err := doSomething(); err != nil {
    return fmt.Errorf("do something failed: %w", err)
}
```

#### 数据库操作
- 使用 GORM 的类型安全 API
- 避免原始 SQL（除非性能关键路径）
- 事务要有明确的边界

### 开发工作流

1. **新功能开发**
   - 从数据模型开始（`internal/database/models/`）
   - 实现业务逻辑（`internal/service/`）
   - 添加 HTTP 接口（`internal/api/handlers/`）
   - 注册路由（`internal/api/routes/`）

2. **代码审查标准**
   - 是否有不必要的复杂度？
   - 错误处理是否完整？
   - 是否有安全隐患？
   - 能否用更少的代码实现？

### 数据库约束

注意：下述约束不适用于特殊表格（如casbin_rule），特殊表格需要写清楚注释。

#### 一、表名命名规范

| 项目 | 规范说明 | 示例 |
|------|----------|------|
| **命名风格** | 小写字母 + 下划线分隔（snake_case） | `user_account` ✅ <br> `UserAccount` ❌ |
| **语义清晰** | 表名应为名词复数或明确实体 | `orders`, `product_inventory` ✅ |
| **前缀/后缀** | 可加后缀如 `_log`、`_history` 表示特殊用途 | `payment_log` ✅ |
| **避免保留字** | 不使用 SQL 关键字 | `app_user` ✅（替代 `user`） |
| **长度限制** | ≤ 64 字符，简洁明确 | `customer_address` ✅ |

#### 二、字段名命名规范

| 项目 | 规范说明 | 示例 |
|------|----------|------|
| **命名风格** | 小写 + 下划线（snake_case） | `created_at`, `order_amount` ✅ |
| **主键字段** | 统一命名为 `id` | `id` BIGINT PRIMARY KEY |
| **外键字段** | `<引用表名_singular>_id` | `user_id`, `product_id` |
| **布尔字段** | 以 `is_`, `has_`, `can_`, `enable_` 开头 | `is_active`, `has_children` |
| **时间字段** | 使用 `_at` 后缀 | `created_at`, `updated_at` |

#### 三、字段类型规范

| 数据类型 | 适用场景 | 推荐类型 |
|----------|--------|----------|
| **整数** | ID、计数、状态码 | `BIGINT`、`INT`、`TINYINT` |
| **字符串** | 短文本（≤255） | `VARCHAR(255)` |
| **长文本** | 内容、备注、JSON | `TEXT` / `JSON` |
| **布尔值** | 真/假状态 | `BOOLEAN` 或 `TINYINT(1)` |
| **金额** | 货币 | `DECIMAL(10,2)` |
| **日期时间** | 时间戳 | `TIMESTAMP` 或 `DATETIME` |

#### 四、通用设计原则

1. **主键**：每张表必须有主键，推荐自增 `BIGINT` 或 UUID
2. **非空约束**：重要字段加 `NOT NULL`
3. **默认值**：常用状态、时间字段设默认值
4. **索引命名**：`idx_<表名>_<字段>`
5. **注释**：每个表和字段添加 COMMENT 说明业务含义
6. **字符集**：统一使用 `utf8mb4`
7. **时区**：数据库存储统一使用 UTC；应用层使用 Asia/Shanghai
8. **外键约束**：不使用外键约束，使用关联表等方式
9. **时间戳**：每张表必须包含created_at、updated_at、deleted_at三个字段

---

## 前端开发规范

### 编码规范

#### 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 组件文件 | PascalCase | `UserProfile.vue` |
| 组件名 | PascalCase | `<UserProfile />` |
| 变量/函数 | camelCase | `getUserInfo()` |
| 常量 | UPPER_SNAKE_CASE | `API_BASE_URL` |
| CSS类名 | kebab-case | `user-profile` |
| 文件夹 | kebab-case | `user-management/` |

#### 组件设计

**单一职责**
```vue
<!-- Good: 拆分成独立组件 -->
<template>
  <div>
    <UserForm @submit="handleSubmit" />
    <UserList :users="users" />
  </div>
</template>
```

**Props定义必须明确类型**
```javascript
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
    if (code !== 0) {
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
```

### 错误处理

```javascript
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

2. **路由懒加载**
```javascript
const UserManagement = () => import('@/views/user/UserManagement.vue')
```

### 样式规范

**颜色系统 - 必须使用CSS变量**
```css
--color-primary: #667eea;
--color-success: #10b981;
--color-warning: #f59e0b;
--color-error: #ef4444;
```

**间距系统 - 8px基础单位**
```css
--spacing-xs: 4px;
--spacing-sm: 8px;
--spacing-md: 16px;
--spacing-lg: 24px;
```

**禁止硬编码颜色和魔法数字**

---

## 前后端协作规范

### 1. 接口契约

- 所有接口返回HTTP 200
- 响应格式统一为Base类型:
```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```
- 错误状态通过`code`字段判断（0为成功），不使用HTTP状态码

### 2. 时间处理规范

**统一时区规则**：
- 数据库存储：UTC时区
- 应用层处理：Asia/Shanghai时区
- 前后端传输：RFC3339格式（不含毫秒）

**前端发送时间到后端**：
```javascript
import { toRFC3339, getTodayStart, getTodayEnd } from '@/utils/date'

// 正确: RFC3339格式（不含毫秒）
const startTime = getTodayStart()  // "2025-11-09T16:00:00Z" ✅
const rfc3339Time = toRFC3339(new Date())  // ✅

// 错误: 带毫秒的格式会导致后端参数绑定失败
const badTime = new Date().toISOString()  // "2025-11-09T16:00:00.123Z" ❌
```

**前端显示后端返回的时间**：
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
- Token存储在localStorage或httpOnly Cookie
- 每次请求Header携带: `Authorization: Bearer <token>`
- Token过期后跳转登录页

---

## 通用开发规范（必须遵守）

### 核心约束

1. 所有构建放在build/文件夹内
2. 所有接口编写完毕后，必须完成详细的swagger文档，同时更新api-test.http
3. 不要添加任何非文字内容（如emoji、各种非ASCII符号）在日志、注释、文档等
4. 在做完任何测试之后，需要删除所有测试文件，保证工作空间干净整洁
5. 不要编写任何不必要的文档，除非显式指明
6. 所有接口的HTTP状态码必须是200，返回必须使用Base类型封装，错误状态码放在Code字段内，0代表成功
7. 维护错误枚举（错误状态码、错误类型），任何错误必须先补充枚举，从枚举中取值
8. 数据库时间字段存储使用UTC时区，应用层处理和展示使用Asia/Shanghai时区
9. 所有错误均需要使用统一的errors来处理
10. 需要编写简要、无歧义的注释，所有注释均使用中文编写

### 后端特定约束

11. 所有日志必须使用内部封装的日志接口，禁止随意打印，禁止使用fmt打印
12. 对于任何并发（多线程、goroutine），必须正确显式关闭，并且加入注释说明
13. 对于任何可能的竞争关系，必须加入锁机制，同时以注释说明
14. 添加、删除任何实体类或变更需要先创建幂等的up/down的sql文件到internal/database/migrations内
15. 各类连接必须从Manager中取出，禁止随意创建，特殊情况需注释说明
16. 对于一般新增接口，均需要权限认证。若无需权限认证，则需要添加注释说明
17. 权限验证必须基于现有的权限规则，禁止私自增加、修改、删除权限规则
18. 在命令执行时，先检索Makefile内的指令，优先使用Makefile指令
19. 若引入自带日志系统的框架，需要将日志系统替换为项目内封装的日志接口
20. 要尽可能注意安全问题，对于可以对用户提权的接口遵循"最小适用"原则
21. 一般业务，id(数字类型主键)不对外暴露，也不存储上下文内，只在查询时使用。对外暴露UUID
22. 若需要连接数据库来完成操作，请先检索toml内的配置，使用配置内的实际地址
23. 系统配置分为两类：静态配置（configs/config.toml）用于部署相关设置，启动时加载，不可热更新；动态配置（system_config表）用于业务相关参数，通过ConfigManager管理，支持热更新和变更监控
24. ServiceContainer和HandlerContainer必须放在internal/api/routes包内，禁止放在internal/app包内（会导致循环依赖）
25. 禁止在routes.go中手动创建服务和handler实例，必须统一通过ServiceContainer和HandlerContainer管理
26. 新增服务或handler时，必须先在对应容器中添加字段和构造逻辑，再在路由中使用
27. 容器内部构造顺序必须遵循依赖关系：基础服务优先，依赖其他服务的后创建
28. 容器只负责组装依赖关系，不允许包含任何业务逻辑、中间件逻辑或路由注册逻辑
29. 严格遵循分层架构，禁止出现循环依赖
30. ChainBuilder中间件链构建器必须放在internal/api/middleware包内，用于提供标准化的中间件组合
31. 禁止在routes.go中手动组装中间件链，必须使用ChainBuilder提供的标准方法
32. ScopeMiddleware用于统一处理基于用户角色的数据作用域控制，禁止在handler中重复编写角色判断代码
33. handler中需要获取作用域限制时，必须使用middleware.GetScopeUserUUID()方法
34. 所有需要基于角色限制数据访问范围的接口，必须使用AuthWithAuditAndScope中间件链
35. 权限验证逻辑必须在中间件层完成，handler只负责业务逻辑
36. 中间件链的顺序不能随意调整，必须遵循：RequestID -> Auth -> Audit -> Permission/Scope -> SecurityPIN的顺序
37. 在编码完成后，必须执行make fmt 和 make line等检查操作（单元测试默认不执行，除非显式指明），并且修复存在的问题。

### 前端特定约束

37. 在调用接口时，一定要先查看api-docs/下的接口文档，确认接口的定义
38. 组件不直接调用axios，必须通过API层
39. 能用Props/Emit解决的，不用Vuex/Pinia
40. 不要随意引入新依赖，需要先评估必要性
41. 不要在组件中直接操作DOM，用Vue的数据驱动
42. 不要把业务逻辑写在模板里，超过3个三元运算符就该提取成computed
43. 单个组件超过300行就该拆分
44. 不要用var，统一使用const/let
45. 不要在循环中使用index作为key，使用唯一标识符(如uuid)

### 通用工作流程约束

46. 任何需要现在时间的地方，都必须调用date命令获取真实的现在时间
47. 优先使用MCP能力（如搜索，第三方库检索），若MCP无法使用，再回退到原始模式
48. 在引用第三方库之前，必须完全了解该库的情况，包括实际能力、接口规范等
49. 所有文档应该正确放置在docs及其子文件夹下
50. 若该目录必须遵守其他相关的规定，则可以在该目录下新增README.md，该文档将会视为和章程一个优先级的规定
51. 若目录下存在README.md，必须先阅读并理解该文档内容，并且将其视为和章程一个优先级
52. 对于分页逻辑，若不带有分页参数，则视为全部导出

---

## 代码审查标准

提交代码前自问：

1. 这个功能/组件是否只做了一件事？
2. 数据结构是否合理，数据流是否清晰？
3. 有没有不必要的嵌套和条件判断？
4. 错误处理是否完整？
5. 有没有安全隐患？
6. 能否用更少的代码实现？

**如果答案有任何疑问，重构它。**

---

## 总结

记住核心原则：

- **数据结构决定代码质量** - 先想清楚数据怎么组织
- **消除特殊情况** - 不要用if/else打补丁
- **单一职责** - 每个包、函数、组件只做一件事
- **简洁胜过clever** - 代码是给人看的，不是炫技
- **安全永远第一** - 不信任任何用户输入
- **分层清晰** - 每层只关心自己的职责
