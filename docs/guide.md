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
[app]
port = 8080
mode = "release"  # debug / release
timezone = "Asia/Shanghai"

[database]
host = "localhost"
port = 3306
database = "vaulthub"
username = "vaulthub"
password = "your_password"
max_open_conns = 100

[redis]
host = "localhost"
port = 6379
password = ""

[jwt]
secret = "your_jwt_secret"  # 必须修改
expire = 3600  # 秒
```

**重要**: 生成强随机密钥

```bash
openssl rand -base64 64
```

### 主密钥配置

主密钥通过环境变量设置，不要写入配置文件：

```bash
export MASTER_KEY="your_master_encryption_key"
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

## 部署

### 单机部署

#### 1. 安装依赖

```bash
# Ubuntu/Debian
sudo apt install -y mariadb-server redis-server

# 创建数据库
sudo mysql -u root -p << EOF
CREATE DATABASE vaulthub CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'vaulthub'@'localhost' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON vaulthub.* TO 'vaulthub'@'localhost';
FLUSH PRIVILEGES;
EOF
```

#### 2. 部署应用

```bash
# 编译
make build

# 配置 systemd
sudo cp scripts/vaulthub.service /etc/systemd/system/
sudo systemctl enable vaulthub
sudo systemctl start vaulthub
```

#### 3. 配置 Nginx

```nginx
upstream vaulthub {
    server 127.0.0.1:8080;
}

server {
    listen 443 ssl http2;
    server_name vault.example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://vaulthub;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 备份

#### 数据库备份

```bash
#!/bin/bash
mysqldump -u vaulthub -p vaulthub \
  --single-transaction \
  | gzip > backup-$(date +%Y%m%d).sql.gz
```

#### 主密钥备份

主密钥必须安全存储，建议：
- 密钥分片存储（Shamir's Secret Sharing）
- 离线保存在保险柜
- 异地备份

## 故障排查

### 无法启动

```bash
# 查看日志
sudo journalctl -u vaulthub -n 50

# 检查配置
./build/vaulthub --config configs/config.toml

# 检查端口
sudo lsof -i :8080
```

### 数据库连接失败

```bash
# 测试连接
mysql -h localhost -u vaulthub -p vaulthub

# 检查数据库状态
sudo systemctl status mariadb
```

### 密钥解密失败

**原因**: 主密钥不正确或数据损坏

**解决**:
- 确认 MASTER_KEY 环境变量正确
- 从备份恢复数据

### API 响应缓慢

```bash
# 检查数据库慢查询
mysql -u vaulthub -p -e "SHOW PROCESSLIST;"

# 检查 Redis
redis-cli --latency

# 优化连接池
# configs/config.toml
[database]
max_open_conns = 200
```

### 更多问题

查看完整 [Swagger 文档](http://localhost:8080/swagger/index.html) 或提交 [Issue](https://github.com/cuihe500/vaulthub/issues)
