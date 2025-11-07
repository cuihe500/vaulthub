# Changelog

本文档记录 VaultHub 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [Unreleased]

## [0.1.0] - 2025-11-07

### Added

#### 核心功能
- 建立完整的应用基础架构（Gin + GORM + Viper + Cobra）
- 实现用户认证系统（JWT Token 机制）
- 实现基于 Casbin 的 RBAC 权限控制系统
- 实现用户 Profile 管理接口
- 实现三阶段加密存储功能
  - 第一阶段：基础加密框架
  - 第二阶段：密钥派生与存储优化
  - 第三阶段：完整的加密存储方案

#### 基础设施
- 统一日志系统（基于 Zap）
- 统一错误处理机制
- Redis 缓存支持
- 数据库迁移机制（golang-migrate）
- 配置管理系统（支持 TOML 格式）
- 数据验证框架（go-playground/validator）

#### 安全特性
- 用户登录互踢机制（防止同一用户多设备同时登录）
- 登录限流机制（防止暴力破解）
- 敏感数据加密存储
- JWT Token 自动刷新与过期管理

#### 文档与工具
- 项目文档库建设
- API 测试文件（api-test.http）
- Swagger 接口文档框架
- Makefile 构建脚本
- 中文 README 文档

### Fixed
- 修复用户多次登录时在 Redis 内重复产生记录导致失效 Token 仍然有效的问题（#2）
- 修复数据库表格不符合命名规范的问题
- 优化 README.md 表述和翻译

### Changed
- 重构日志系统，统一全项目日志输出格式
- 增强系统安全性检查机制

### Security
- 实施最小权限原则，严格控制权限提升接口
- 数据库时间字段统一使用 UTC 时区存储
- 所有敏感配置支持环境变量覆盖

## 技术栈

- **语言**: Go 1.25.1
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MariaDB/MySQL
- **缓存**: Redis
- **权限**: Casbin (RBAC)
- **认证**: JWT
- **日志**: Zap
- **配置**: Viper (TOML)
- **CLI**: Cobra
- **迁移**: golang-migrate

## 许可证

Apache 2.0

---

[Unreleased]: https://github.com/cuihe500/vaulthub/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/cuihe500/vaulthub/releases/tag/v0.1.0
