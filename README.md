# VaultHub

VaultHub 是一个密钥管理系统，旨在安全地存储、管理和轮换加密密钥、API 密钥及其他敏感凭证。

## 功能特性

- 安全的密钥存储和管理
- 加密密钥、API 密钥及敏感凭证管理
- RESTful API 接口
- 数据库迁移管理
- 统一的错误处理机制
- 结构化日志记录
- 健康检查端点

## 快速开始

### 前置条件

- Go 1.25.1+
- MySQL/MariaDB 5.7+

### 安装

```bash
# 克隆代码库
git clone https://github.com/cuihe500/vaulthub.git
cd vaulthub/backend

# 安装依赖
make deps

# 配置环境
cp configs/.env.example configs/.env
# 编辑 configs/.env 填入数据库凭证
```

### 运行

```bash
# 直接构建并运行
make run

# 或者分步操作
make build
./build/vaulthub serve
```

服务将在 `http://localhost:8080` 启动

### 健康检查

```bash
curl http://localhost:8080/health
```

## 项目结构

```
.
├── cmd/vaulthub/           # 应用程序入口
├── internal/               # 私有应用代码
│   ├── api/                # HTTP 层
│   │   ├── handlers/       # 请求处理器
│   │   ├── middleware/     # 中间件
│   │   └── routes/         # 路由定义
│   ├── config/             # 配置管理
│   ├── database/           # 数据库层
│   │   ├── migrations/     # 数据库迁移文件
│   │   └── models/         # 数据模型
│   └── service/            # 业务逻辑
├── pkg/                    # 公共库
│   ├── crypto/             # 加密工具
│   ├── logger/             # 日志工具
│   └── response/           # HTTP 响应助手
├── configs/                # 配置文件
└── build/                  # 构建输出目录
```

## 配置

### 配置方式

配置可通过以下方式设置：
1. YAML 文件：`configs/config.yaml`
2. 环境变量（优先级更高）

### 环境变量说明

主要配置项：

```bash
# 服务器配置
SERVER_PORT=8080              # 服务端口
SERVER_MODE=debug             # 运行模式：debug/release

# 数据库配置
DB_HOST=localhost             # 数据库主机
DB_PORT=3306                  # 数据库端口
DB_USER=root                  # 数据库用户名
DB_PASSWORD=                  # 数据库密码
DB_NAME=vaulthub              # 数据库名称
DB_CHARSET=utf8mb4            # 字符集
DB_MAX_IDLE_CONNS=10          # 最大空闲连接数
DB_MAX_OPEN_CONNS=100         # 最大打开连接数

# 日志配置
LOG_LEVEL=info                # 日志级别：debug/info/warn/error
LOG_FILE=logs/vaulthub.log    # 日志文件路径
LOG_MAX_SIZE=100              # 单个日志文件最大大小（MB）
LOG_MAX_BACKUPS=3             # 保留的旧日志文件数量
LOG_MAX_AGE=28                # 日志文件保留天数

# 时区配置
TIMEZONE=Asia/Shanghai        # 时区设置
```

查看完整配置选项请参考 `configs/.env.example`

## 数据库迁移

VaultHub 使用 [golang-migrate](https://github.com/golang-migrate/migrate) 进行数据库架构管理。

### 启动时自动迁移

运行 `./vaulthub serve` 时会自动应用迁移。

### 手动迁移命令

```bash
# 应用所有待执行的迁移
./vaulthub migrate up

# 回滚最后一次迁移
./vaulthub migrate down

# 显示当前迁移版本
./vaulthub migrate version

# 应用 N 次迁移（正数向上，负数向下）
./vaulthub migrate steps -n 2    # 向上迁移 2 步
./vaulthub migrate steps -n -1   # 向下迁移 1 步

# 强制设置版本（谨慎使用，仅在脏状态时使用）
./vaulthub migrate force -v 1
```

### 创建新迁移

迁移文件必须遵循以下命名规范：
- `{version}_{name}.up.sql` - 正向迁移
- `{version}_{name}.down.sql` - 回滚迁移

### 最佳实践

1. **始终使用 `IF NOT EXISTS` / `IF EXISTS`** 确保幂等性
2. **在提交前测试正向和回滚迁移** 确保可逆性
3. **不要编辑已应用的迁移** 而是创建新的迁移
4. **版本号必须递增** 使用 6 位数字格式（000001, 000002, ...）
5. **保持迁移的原子性** 每次迁移只做一个逻辑变更

## API 文档

API 文档使用 Swagger 生成。

启动服务后访问：
- Swagger UI：`http://localhost:8080/swagger/index.html`

注意：所有接口使用 HTTP 200 状态码，实际错误状态通过响应体中的 `code` 字段表示。

## Makefile 命令参考

项目提供了完整的Makefile来简化开发工作流。所有开发任务都应使用Makefile命令，而不是直接使用go命令。

### 命令速查表

| 命令 | 说明 |
|------|------|
| `make help` | 显示所有可用命令 |
| `make deps` | 安装和整理依赖 |
| `make build` | 构建开发版本 |
| `make build-prod` | 构建生产版本（Linux/amd64） |
| `make run` | 构建并运行服务 |
| `make test` | 运行所有测试 |
| `make coverage` | 生成测试覆盖率报告 |
| `make fmt` | 格式化代码 |
| `make lint` | 运行代码检查 |
| `make clean` | 清理构建产物 |
| `make version` | 显示版本信息 |

### 查看帮助

```bash
# 查看所有可用命令及说明
make help
```

### 基本命令

```bash
# 安装依赖
make deps

# 构建项目
make build

# 构建并运行
make run

# 清理构建产物
make clean
```

### 开发命令

```bash
# 运行测试
make test

# 生成测试覆盖率报告（会生成 coverage.html）
make coverage

# 格式化代码
make fmt

# 代码检查
make lint
```

### 构建命令

```bash
# 开发构建（本地平台）
make build

# 生产构建（Linux/amd64，带版本信息）
make build-prod
```

### 版本信息

```bash
# 查看版本信息（包含Git提交、构建时间等）
make version
```

Makefile会自动注入以下版本信息到构建产物：
- **版本号**：从Git标签获取，无标签时显示 "dev"
- **Git提交哈希**：当前提交的短哈希
- **构建时间**：UTC时间戳
- **Go版本**：构建时使用的Go版本

### 典型开发工作流

```bash
# 1. 首次设置
make deps                    # 安装依赖

# 2. 开发循环
make fmt                     # 格式化代码
make lint                    # 检查代码
make test                    # 运行测试

# 3. 本地运行
make run                     # 构建并启动服务

# 4. 提交前检查
make fmt && make lint && make test && make build

# 5. 查看构建信息
make version                 # 确认版本信息

# 6. 清理
make clean                   # 清理构建产物
```

## 生产部署

### 构建生产版本

```bash
# 使用Makefile构建生产版本
make build-prod
```

生产构建特性：
- 静态链接（CGO_ENABLED=0）
- 目标平台：Linux/amd64
- 注入完整版本信息
- 优化二进制大小

### 构建 Docker 镜像

```bash
# 在项目根目录执行
docker build -t vaulthub:latest -f backend/Dockerfile .
```

### 部署建议

1. **使用生产模式**：设置 `SERVER_MODE=release`
2. **配置日志轮转**：合理设置日志文件大小和保留策略
3. **数据库连接池**：根据负载调整 `DB_MAX_OPEN_CONNS` 和 `DB_MAX_IDLE_CONNS`
4. **使用反向代理**：通过 Nginx 或其他反向代理暴露服务
5. **启用 HTTPS**：生产环境必须使用 TLS
6. **备份策略**：定期备份数据库和关键配置

### 安全建议

1. **敏感数据加密**：所有密钥和凭证在存储前必须加密
2. **环境变量保护**：不要在代码中硬编码敏感信息
3. **定期更新依赖**：及时更新依赖库以修复安全漏洞
4. **访问控制**：实施适当的身份验证和授权机制
5. **日志脱敏**：确保日志中不包含敏感信息

## 故障排查

### 常见问题

**问题：数据库连接失败**
```
解决：检查数据库配置是否正确，确认数据库服务已启动
```

**问题：迁移失败处于脏状态**
```
解决：使用 migrate force 命令修复，然后手动检查数据库状态
```

**问题：端口被占用**
```
解决：修改 SERVER_PORT 环境变量或关闭占用端口的进程
```

## 许可证

Apache 2.0 - 详见 [LICENSE](../LICENSE)

## 作者

Changhe Cui - admin@thankseveryone.top

## 代码库

https://github.com/cuihe500/vaulthub
