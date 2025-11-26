# VaultHub 用户指南

## 快速开始

### 前置要求

- Go 1.25.1+
- MariaDB/MySQL 8.0+
- Redis 6.0+

### 5分钟快速部署

```bash
# 1. 克隆仓库
git clone https://github.com/cuihe500/vaulthub.git
cd vaulthub

# 2. 配置环境
cp configs/config.toml.example configs/config.toml
vim configs/config.toml  # 修改数据库和 Redis 连接信息

# 3. 初始化数据库
make migrate-up

# 4. 启动服务
make run
```

服务默认启动在 `http://localhost:8080`

### 验证安装

```bash
curl http://localhost:8080/health
# 返回: {"status":"ok"}
```

## 配置说明

### 配置分类

VaultHub 的配置分为两类：

**静态配置** (`configs/config.toml`): 启动时加载，不支持热更新
- 数据库连接
- Redis 连接
- 服务器端口
- JWT 密钥

**动态配置** (`system_config` 表): 支持热更新
- 密钥轮换周期
- 系统业务参数

### 核心配置项

```toml
[server]
host = "0.0.0.0"
port = 8080
mode = "release"  # debug / release / test

[database]
driver = "mysql"
host = "localhost"
port = 3306
name = "vaulthub"
user = "vaulthub"
password = "your_password"

[redis]
mode = "standalone"  # standalone / sentinel / cluster
host = "localhost"
port = 6379
password = ""
db = 0

[security]
jwt_secret = "your_jwt_secret_change_in_production"  # 必须修改为强随机字符串
jwt_expiration = 24  # JWT过期时间（小时）
encryption_key = "must-be-exactly-32-bytes-long!!"  # 必须是32字节
casbin_model_path = "./configs/rbac_model.conf"
admin_username = "admin"
admin_password = "Admin@123456"  # 首次启动时创建超级管理员

[logger]
level = "info"  # debug / info / warn / error / fatal
encoding = "console"  # console / json
output_paths = ["stdout"]
error_output_paths = ["stderr"]

```

**重要**: 生成强随机密钥

```bash
# 生成JWT密钥（任意长度）
openssl rand -base64 64

# 生成32字节加密密钥（必须恰好32字节）
openssl rand -base64 32
```

## API 使用

### 基础信息

- Base URL: `http://localhost:8080/api/v1`
- 认证: JWT Bearer Token
- 响应格式: JSON
- HTTP 状态码: 统一使用 200，错误码在响应体 `code` 字段

### 统一响应格式

成功:
```json
{
  "code": 0,
  "message": "success",
  "data": {...}
}
```

失败:
```json
{
  "code": 1001,
  "message": "invalid parameters",
  "data": null
}
```

### 认证流程

#### 1. 注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "SecureP@ssw0rd"
  }'
```

#### 2. 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecureP@ssw0rd"
  }'
```

返回:
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": 1640000000
  }
}
```

### 密钥管理

#### 创建密钥

```bash
curl -X POST http://localhost:8080/api/v1/secrets \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "aws-api-key",
    "type": "api_key",
    "value": "AKIAIOSFODNN7EXAMPLE",
    "description": "AWS S3 访问密钥",
    "rotation_enabled": true,
    "rotation_days": 90
  }'
```

#### 查询密钥列表

```bash
curl -X GET "http://localhost:8080/api/v1/secrets?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 获取密钥详情

```bash
curl -X GET http://localhost:8080/api/v1/secrets/{uuid} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 轮换密钥

```bash
curl -X POST http://localhost:8080/api/v1/secrets/{uuid}/rotate \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "new_value": "NEW_SECRET_VALUE"
  }'
```

### 常用错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 认证失败 |
| 1003 | 权限不足 |
| 1004 | 资源不存在 |
| 2001 | 数据库错误 |

### 更多问题

查看完整 [Swagger 文档](http://localhost:8080/swagger/index.html) 或提交 [Issue](https://github.com/cuihe500/vaulthub/issues)
