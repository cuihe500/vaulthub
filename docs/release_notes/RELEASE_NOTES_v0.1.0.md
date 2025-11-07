# VaultHub v0.1.0 - 首个公开版本

## Release Title
```
v0.1.0 - Initial Release: 密钥管理系统核心功能实现
```

## Release Notes

VaultHub v0.1.0 是项目的首个公开版本，实现了密钥管理系统的核心功能框架。

### 核心特性

**用户认证与权限管理**
- 完整的 JWT Token 认证机制
- 基于 Casbin 的 RBAC 权限控制系统
- 用户登录互踢机制（防止同一账户多设备登录）
- 登录限流保护（防暴力破解）

**加密存储系统**
- 三阶段加密存储实现
- 敏感数据全程加密
- 密钥派生与安全存储

**基础设施**
- 统一的日志系统（Zap）
- 统一的错误处理机制
- Redis 缓存支持
- 数据库自动迁移
- 配置热加载（TOML）

### 技术栈

- Go 1.25.1
- Gin Web 框架
- GORM ORM
- MariaDB/MySQL
- Redis
- Casbin (RBAC)
- JWT 认证
- Zap 日志

### 重要修复

- 修复多次登录导致失效 Token 仍有效的问题 (#2)
- 规范化数据库表命名
- 加强系统安全性检查

### 安全特性

- 最小权限原则实施
- 敏感数据加密存储
- UTC 时区统一管理
- 环境变量敏感配置支持

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/cuihe500/vaulthub.git
cd vaulthub

# 构建项目
make build

# 配置数据库和 Redis（编辑 configs/config.toml）
# 运行服务
./build/vaulthub serve
```

### 文档

- [项目文档](https://github.com/cuihe500/vaulthub/tree/main/docs)
- [API 文档](https://github.com/cuihe500/vaulthub/tree/main/docs/swagger)
- [CHANGELOG](https://github.com/cuihe500/vaulthub/blob/main/docs/CHANGELOG.md)

### 已知限制

- 本版本为初始开发版本（0.x），API 可能会有变动
- 部分高级功能尚在开发中
- 建议在测试环境使用，生产环境请谨慎评估

### 许可证

Apache License 2.0

### 贡献者

- Changhe Cui (@cuihe500)

---

**Full Changelog**: https://github.com/cuihe500/vaulthub/commits/v0.1.0
