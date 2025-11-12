# VaultHub

密钥管理系统，安全存储、管理和轮换加密密钥、API 密钥及敏感凭证。

## 快速开始

### 前置条件

- Go 1.25.1+
- MariaDB/MySQL 8.0+
- Redis 6.0+

### 安装运行

```bash
# 克隆仓库
git clone https://github.com/cuihe500/vaulthub.git
cd vaulthub

# 配置数据库
cp configs/config.toml.example configs/config.toml
vim configs/config.toml  # 修改数据库连接信息

# 初始化并运行
make migrate-up
make run
```

服务启动在 `http://localhost:8080`

验证: `curl http://localhost:8080/health`

## 常用命令

```bash
make run            # 运行服务
make build          # 编译
make test           # 测试
make migrate-up     # 数据库迁移
make help           # 查看所有命令
```

## 文档

- [用户指南](docs/guide.md) - 安装、配置、API 使用
- [开发指南](docs/developer.md) - 架构、数据库、贡献
- [GitHub Actions 使用指南](docs/github-actions.md) - 自动化构建和发布
- [变更日志](docs/CHANGELOG.md) - 版本记录
- [Swagger API](http://localhost:8080/swagger/index.html) - 启动后访问

## 许可证

Apache 2.0

## 作者

Changhe Cui - admin@thankseveryone.top

https://github.com/cuihe500/vaulthub
