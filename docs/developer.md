# VaultHub 开发指南

## 架构设计

### 整体架构

```
HTTP 请求 → Gin 路由 → Handler → Service → Model → 数据库
                         ↓
                    Middleware (认证/权限/日志)
```

### 分层职责

- **Handler**: HTTP 请求处理，参数验证
- **Service**: 业务逻辑，与 HTTP 无关
- **Model**: 数据访问，GORM 操作
- **Middleware**: 认证、权限、日志、限流

### 核心组件

**Manager** (`internal/app/manager.go`): 全局资源管理
- 数据库连接
- Redis 连接
- Casbin 权限执行器
- 配置管理器

**ConfigManager** (`internal/config/manager.go`): 配置管理
- 支持热更新
- 内存缓存
- 变更通知

**Scheduler** (`internal/app/scheduler.go`): 定时任务
- 基于 cron
- 密钥自动轮换

### 安全设计

**加密策略**:
- 算法: AES-256-GCM
- 信封加密 (Envelope Encryption)
  - 主密钥: 环境变量存储
  - DEK: 每个密钥独立生成

**认证授权**:
- 认证: JWT Token
- 授权: Casbin RBAC
- 密码: bcrypt (cost=10)

## 数据库设计

### 核心表

**users** - 用户表
```sql
id, uuid, username, email, password_hash, status
created_at, updated_at, deleted_at
```

**secrets** - 密钥表
```sql
id, uuid, user_id, name, type
encrypted_value, encrypted_dek, nonce
rotation_enabled, rotation_days
created_at, updated_at, deleted_at
```

**audit_logs** - 审计日志
```sql
id, uuid, user_id, action, resource_type, resource_id
ip_address, status, created_at
```

### 设计原则

1. **字符集**: utf8mb4
2. **时区**: 数据库存储 UTC，应用层使用 Asia/Shanghai
3. **软删除**: 所有表包含 deleted_at
4. **主键**: 自增 BIGINT id，对外暴露 UUID
5. **无外键**: 不使用数据库外键约束

## 开发规范

### 命名规范

```go
// 包名: 小写单词
package crypto

// 函数: 驼峰命名
func GetUserByID(id string) (*User, error)

// 常量: 驼峰命名
const DefaultTimeout = 30 * time.Second

// 错误: Err 前缀
var ErrNotFound = errors.New("not found")
```

### 错误处理

使用统一的错误包:

```go
import "github.com/cuihe500/vaulthub/pkg/errors"

// Good
if id == "" {
    return nil, errors.ErrInvalidParams
}

// 包装错误
if err := db.Find(&user).Error; err != nil {
    return nil, errors.Wrap(err, "查询用户失败")
}
```

### 日志规范

使用结构化日志:

```go
import "github.com/cuihe500/vaulthub/pkg/logger"

logger.Info("用户登录成功",
    zap.String("user_id", userID),
    zap.String("ip", clientIP),
)

logger.Error("数据库连接失败", zap.Error(err))
```

**禁止使用 fmt 打印**（CLI 工具除外）

### 并发规范

显式关闭 Goroutine:

```go
func processJobs(ctx context.Context) {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            // 显式退出
            return
        case <-ticker.C:
            processJob()
        }
    }
}
```

使用锁保护共享资源:

```go
type Cache struct {
    mu   sync.RWMutex  // 保护 data
    data map[string]string
}

func (c *Cache) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.data[key]
}
```

## 贡献指南

### 工作流程

```bash
# 1. Fork 仓库并克隆
git clone https://github.com/YOUR_USERNAME/vaulthub.git
cd vaulthub

# 2. 创建分支
git checkout -b feature/your-feature

# 3. 开发和测试
make fmt && make lint && make test

# 4. 提交代码
git commit -m "feat: add xxx feature"

# 5. 推送并创建 PR
git push origin feature/your-feature
```

### 提交规范

```
<type>: <subject>

<body>
```

**Type 类型**:
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `refactor`: 重构
- `test`: 测试相关

**示例**:
```
feat: 添加密钥批量导入功能

支持从 CSV 文件批量导入密钥
包括格式验证和重复检测

Closes #123
```

### 代码审查

PR 必须满足:
- [ ] 代码遵循编码规范
- [ ] 添加必要的测试
- [ ] 测试全部通过
- [ ] 更新相关文档

### 测试

```bash
# 运行测试
make test

# 测试覆盖率
make coverage

# 单个包测试
go test -v ./internal/service
```

## 开发环境

### 依赖安装

```bash
make deps
```

### 本地运行

```bash
# 启动数据库和 Redis (Docker)
docker-compose up -d mysql redis

# 运行迁移
make migrate-up

# 启动服务
make run
```

### 调试

使用 Delve:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/vaulthub/main.go
```

## 常用命令

```bash
make build          # 编译
make run            # 运行
make test           # 测试
make lint           # 代码检查
make fmt            # 格式化
make migrate-up     # 数据库迁移
make clean          # 清理
```

## 许可证

Apache 2.0
