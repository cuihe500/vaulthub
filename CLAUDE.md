# VaultHub 章程

## 项目元数据（不可更改）

**作者**：Changhe Cui
**邮箱**：admin@thankseveryone.top
**代码库**：https://github.com/cuihe500/vaulthub
**许可证**：Apache 2.0

VaultHub 是一个密钥管理系统，旨在安全地存储、管理和轮换加密密钥、API 密钥及其他敏感凭证。

## 技术栈

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
│   ├── errors/            # 统一错误处理
│   ├── jwt/               # JWT工具
│   ├── logger/            # 日志工具
│   ├── redis/             # Redis工具
│   ├── response/          # 统一响应格式
│   ├── validator/         # 数据验证工具
│   └── version/           # 版本信息
├── configs/               # 配置文件
│   ├── config.toml        # 主配置文件
│   ├── rbac_model.conf    # Casbin权限模型
│   └── .env.example       # 环境变量示例
├── build/                 # 构建输出目录
├── docs/                  # 文档
│   └── swagger/           # Swagger接口文档
├── scripts/               # 构建和部署脚本
├── web/                   # 前端资源
├── go.mod                 # Go 模块定义
└── go.sum                 # 依赖校验

```

## 设计原则

### 1. 简洁至上
- 单一职责：每个包、每个函数只做一件事
- 避免过度抽象：不为"可能的需求"写代码
- 控制嵌套深度：函数不超过3层缩进

### 2. 清晰的分层架构
```
HTTP请求 → 路由 → 处理器 → 服务层 → 数据访问层 → 数据库
```
- **handlers**: 只负责 HTTP 协议相关（请求解析、响应格式）
- **service**: 核心业务逻辑，与 HTTP 无关
- **models/database**: 数据持久化，只关心数据

**关键组件说明**：
- **ConfigManager**: 配置管理器，提供配置热更新、内存缓存、变更监控能力
- **Scheduler**: 定时任务调度器，基于cron管理密钥轮换等周期性任务
- **Manager**: 全局资源管理器，统一管理数据库、Redis、Casbin等连接

### 3. 安全第一
- 所有敏感数据加密存储
- 输入验证在处理器层完成
- 防止常见漏洞：SQL 注入、XSS、CSRF

### 4. 可测试性
- 业务逻辑与框架解耦
- 依赖注入优于全局变量
- 单元测试覆盖核心逻辑

## 编码规范

### 命名
- 包名：小写，单个单词
- 接口：简洁有力（如 `Reader`, `Writer`）
- 函数/变量：驼峰命名，见名知义
- 常量：大写+下划线或驼峰
- 日志、注释：中文优先且作为一等公民支持

### 错误处理
```go
// Bad: 吞掉错误
_ = doSomething()

// Good: 明确处理
if err := doSomething(); err != nil {
    return fmt.Errorf("do something failed: %w", err)
}
```

### 数据库操作
- 使用 GORM 的类型安全 API
- 避免原始 SQL（除非性能关键路径）
- 事务要有明确的边界

## 开发工作流

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

## 数据库约束

注意：下述约束不适用于特殊表格（如casbin_rule），特殊表格需要写清楚注释。

### 一、表名命名规范

| 项目 | 规范说明 | 示例 |
|------|----------|------|
| **命名风格** | 小写字母 + 下划线分隔（snake_case） | `user_account` ✅ <br> `UserAccount` ❌（避免驼峰/帕斯卡） |
| **语义清晰** | 表名应为名词复数或明确实体，表示存储内容 | `orders`, `product_inventory` ✅ <br> `data1`, `temp_table` ❌ |
| **前缀/后缀** | 一般不加前缀；可加后缀如 `_log`、`_history` 表示特殊用途 | `payment_log`, `employee_archive` ✅ |
| **避免保留字** | 不使用 SQL 关键字（如 `order`, `group`, `user`） | `app_user` ✅（替代 `user`） |
| **长度限制** | ≤ 64 字符，简洁明确 | `customer_address` ✅ |

> ✅ 推荐：`order_item`  
> ❌ 避免：`OrderItems`, `OI`, `123order`

### 二、字段名命名规范

| 项目 | 规范说明 | 示例 |
|------|----------|------|
| **命名风格** | 小写 + 下划线（snake_case） | `created_at`, `order_amount` ✅ |
| **语义明确** | 避免缩写歧义，优先完整英文 | `email_address` ✅ <br> `eml_addr` ❌（除非团队约定） |
| **主键字段** | 统一命名为 `id`（单列）或 `<table>_id`（复合主键） | `id` BIGINT PRIMARY KEY |
| **外键字段** | `<引用表名_singular>_id` | `user_id`, `product_id` |
| **布尔字段** | 以 `is_`, `has_`, `can_`, `enable_` 开头 | `is_active`, `has_children` |
| **时间字段** | 使用 `_at` 后缀，类型匹配含义 | `created_at` DATETIME, `updated_at` TIMESTAMP |
| **避免保留字** | 不使用 `desc`, `key`, `type`, `timestamp` 等 | `description` ✅（替代 `desc`） |
| **统一单位/格式** | 金额用最小货币单位（如分），时间用 UTC | `price_cents` INT, `order_time` TIMESTAMP |

> ✅ 示例字段：  
> `id`, `username`, `email`, `is_verified`, `created_at`, `updated_at`

### 三、字段类型规范（以常见类型为例）

| 数据类型 | 适用场景 | 推荐类型示例 |
|----------|--------|-------------|
| **整数** | ID、计数、状态码 | `BIGINT`（大表主键）、`INT`、`SMALLINT`、`TINYINT`（布尔用 `TINYINT(1)` 或 `BOOLEAN`） |
| **字符串** | 短文本（≤255） | `VARCHAR(255)` <br> 用户名、手机号、邮箱等 |
| **长文本** | 内容、备注、JSON | `TEXT` / `JSON`（MySQL 5.7+，PostgreSQL） |
| **布尔值** | 真/假状态 | `BOOLEAN` 或 `TINYINT(1)`（0=false, 1=true） |
| **金额/精确小数** | 货币、计算 | `DECIMAL(10,2)`（总10位，小数2位） |
| **日期时间** | 创建/更新时间 | `TIMESTAMP`（自动更新）或 `DATETIME`（MySQL）<br>`TIMESTAMPTZ`（PostgreSQL 带时区） |
| **浮点数** | 科学计算、允许误差 | `FLOAT` / `DOUBLE`（慎用于金额） |
| **二进制数据** | 文件、图片哈希等 | `BLOB` / `BYTEA`（PostgreSQL） |
| **枚举类型** | 有限固定值 | 优先用 `VARCHAR` + 应用层约束，或 `ENUM`（MySQL）/ 替代方案：`CHECK (status IN ('active','inactive'))` |

### 四、通用设计原则

1. **主键**：每张表必须有主键，推荐自增 `BIGINT` 或 UUID（需考虑性能）。
2. **非空约束**：重要字段加 `NOT NULL`，避免 NULL 语义混乱。
3. **默认值**：常用状态（如 `is_active = TRUE`）、时间字段设默认值。
4. **索引命名**：`idx_<表名>_<字段>`，如 `idx_user_email`
5. **注释**：每个表和字段添加 COMMENT 说明业务含义。
6. **字符集**：统一使用 `utf8mb4`（支持 emoji 和全 Unicode）。
7. **时区**：数据库存储统一使用 UTC；应用层使用 Asia/Shanghai，在读写时进行转换。
8. **外键约束**：不使用任何外键约束，而是使用关联表等其他方式。
9. **时间戳**：每张表必须包含created_at、updated_at、deleted_at三个字段，并且这三个字段优先由数据库层处理(deleted_at除外)，其次由应用层处理。

### 五、命名禁忌

| ❌ 错误做法 | ✅ 正确做法 |
|------------|------------|
| `User`（保留字） | `app_user` |
| `name`（模糊） | `full_name`, `nickname` |
| `time`（类型不明确） | `created_at`, `expire_time` |
| `data`（无意义） | `profile_data`, `extra_info` |
| 使用空格、中文、特殊符号 | 仅使用 `a-z`, `0-9`, `_` |

## 注意（必须遵守）
1. 所有的构建均放在build/文件夹内。
2. 所有接口编写完毕后，必须完成详细的swagger接口文档，同时更新api-test.http。
3. 不要添加任何非文字内容（如emoji、各种非ASCII符号）在日志、注释、文档等。
4. 在做完任何测试之后，需要删除所有测试文件，保证工作空间干净整洁。
5. 不要编写任何不必要的文档，除非显式指明。
6. 所有接口的必须使用200 HTTP状态码，所有接口的返回必须使用Base类型封装，错误状态码放置在Base基本类型的Code字段内。
7. 维护数个枚举（如错误状态码（用于返回值等）、错误类型（用于日志等）），任何错误都必须先补充枚举，所有错误类型都必须从枚举中取值。
8. 数据库时间字段存储使用UTC时区，应用层处理和展示使用Asia/Shanghai时区。
9. 所有错误均需要使用统一的errors来处理。
10. 需要编写简要、无歧义的注释，所有注释均使用中文编写。
11. 所有日志必须使用内部封装的日志接口，禁止随意打印，禁止使用fmt打印（特殊情况除外，需添加注释说明）。
12. 对于任何的并发（多线程、goroutine），必须正确显式关闭，并且加入注释说明。
13. 对于任何可能的竞争关系，必须加入锁机制，同时以注释的形式说明锁机制的意义。
14. 添加、删除任何实体类或者变更需要关联数据库变更，必须先创建幂等的up/down的sql文件到internal/database/migrations内。
15. 各类连接（如数据库连接、Redis连接等）必须从Manager中取出，禁止随意创建，若不存在，则必须在应用启动的时候创建连接，特殊情况请加注释说明。
16. 对于一般新增接口，均需要权限认证。若无需权限认证，则需要添加注释说明。
17. 权限验证必须基于现有的权限规则，禁止私自增加、修改、删除权限规则，若需要请询问并明确拿到肯定回复才可以，并且要添加注释说明。
18. 在命令执行的时候，先检索Makefile内的指令，若有可以替代的，优先使用Makefile内的指令而不是原生命令。
19. 在任何注释中不要提及CLAUDE.md文档的存在，该文档默认是对其他用户不可见的。
20. 若引入自带日志系统的框架，需要将日志系统替换为项目内封装的日志接口，保证全项目（包括三方框架）日志系统统一。
21. 要尽可能的注意安全问题，对于可以对用户提权的接口一定要遵循“最小适用”原则，要反复检查是否存在可能的安全问题，并且对该类接口标注好注释。
22. 一般业务，id(数字类型主键)不对外暴露，也不存储上下文内，只在查询的时候使用。对外暴露UUID或其他唯一标识符。
23. 若需要连接数据库来完成操作，请先检索toml内的配置，使用配置内的实际地址来完成操作
24. 各个配置需要区分清楚，将其放置在正确的位置内（配置文件/数据库），并且调用对应的方法完成配置变更。禁止在现有体系之外随意增加、删除、修改各类配置，若现有情况无法满足，则需要询问并且以注释的方式写出详细的原因。
25. 系统配置分为两类：静态配置（configs/config.toml）用于部署相关设置（如数据库连接、服务器端口），在启动时加载，不可热更新；动态配置（system_config表）用于业务相关参数（如密钥轮换周期、系统参数），通过ConfigManager管理，支持热更新和变更监控。
26. 任何需要现在时间的地方，都必须调用date命令获取真实的现在时间。