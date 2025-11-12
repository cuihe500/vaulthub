# GitHub Actions 使用指南

## 概述

VaultHub 使用 GitHub Actions 实现自动化的持续集成和发布流程。所有构建命令复用项目的 Makefile,确保本地构建和 CI 构建完全一致。

## 工作流程

### 1. 持续集成 (CI)

**触发条件**:
- 推送代码到 `main` 或 `develop` 分支
- 创建或更新 Pull Request 到 `main` 或 `develop` 分支

**执行内容**:
```yaml
1. 代码格式检查 (go fmt)
2. 静态代码分析 (golangci-lint)
3. 前端构建验证
4. 单元测试 + 覆盖率统计
5. 后端构建验证
```

**配置文件**: `.github/workflows/ci.yml`

**查看结果**:
- GitHub 仓库页面 → Actions 标签 → CI workflow
- Pull Request 中会显示检查状态

### 2. 自动发布 (Release)

**触发条件**:
- 推送符合 `v*.*.*` 格式的 Git 标签

**执行内容**:
```yaml
1. 创建 GitHub Release
2. 构建多平台二进制文件:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
3. 构建并推送 Docker 多架构镜像:
   - linux/amd64
   - linux/arm64
4. 上传构建产物到 Release
5. 生成版本更新说明
```

**配置文件**: `.github/workflows/release.yml`

## 发布新版本

### 步骤 1: 更新版本号和变更日志

```bash
# 1. 更新 CHANGELOG.md
vim docs/CHANGELOG.md

# 添加新版本记录
## [0.2.0] - 2025-11-12
### Added
- 新功能描述
### Fixed
- 修复问题描述
```

### 步骤 2: 创建并推送标签

```bash
# 1. 创建带注释的标签
git tag -a v0.2.0 -m "Release v0.2.0"

# 2. 推送标签到远程仓库
git push origin v0.2.0
```

### 步骤 3: 等待自动构建

推送标签后,GitHub Actions 会自动:
1. 触发 Release 工作流
2. 并行构建 4 种平台二进制文件
3. 构建 Docker 多架构镜像
4. 创建 GitHub Release 并上传产物

**预计耗时**: 10-15 分钟

### 步骤 4: 验证发布结果

访问 GitHub 仓库的 **Releases** 页面,确认:
- ✅ Release 已创建
- ✅ 包含 4 个二进制压缩包:
  - `vaulthub_linux_amd64.tar.gz`
  - `vaulthub_linux_arm64.tar.gz`
  - `vaulthub_darwin_amd64.tar.gz`
  - `vaulthub_darwin_arm64.tar.gz`
- ✅ 每个压缩包都有对应的 `.sha256` 校验文件
- ✅ Docker 镜像已推送到 `ghcr.io/cuihe500/vaulthub`

### 步骤 5: 验证 Docker 镜像

```bash
# 拉取最新镜像
docker pull ghcr.io/cuihe500/vaulthub:v0.2.0

# 验证版本信息
docker run --rm ghcr.io/cuihe500/vaulthub:v0.2.0 version
```

## Docker 镜像标签策略

每次发布会生成多个标签:

| 标签 | 说明 | 示例 |
|------|------|------|
| `v0.2.0` | 完整版本号 | `ghcr.io/cuihe500/vaulthub:v0.2.0` |
| `0.2` | 主版本号.次版本号 | `ghcr.io/cuihe500/vaulthub:0.2` |
| `0` | 主版本号 | `ghcr.io/cuihe500/vaulthub:0` |
| `latest` | 最新稳定版 | `ghcr.io/cuihe500/vaulthub:latest` |

**推荐使用**:
- 生产环境: 固定完整版本号 `v0.2.0`
- 测试环境: 使用 `latest` 自动跟踪最新版本

## 使用发布的二进制文件

### Linux (amd64)

```bash
# 1. 下载
wget https://github.com/cuihe500/vaulthub/releases/download/v0.2.0/vaulthub_linux_amd64.tar.gz

# 2. 验证校验和
wget https://github.com/cuihe500/vaulthub/releases/download/v0.2.0/vaulthub_linux_amd64.tar.gz.sha256
sha256sum -c vaulthub_linux_amd64.tar.gz.sha256

# 3. 解压
tar -xzf vaulthub_linux_amd64.tar.gz

# 4. 安装
sudo mv vaulthub_linux_amd64 /usr/local/bin/vaulthub
sudo chmod +x /usr/local/bin/vaulthub

# 5. 验证
vaulthub version
```

### macOS (arm64/Apple Silicon)

```bash
# 1. 下载
curl -LO https://github.com/cuihe500/vaulthub/releases/download/v0.2.0/vaulthub_darwin_arm64.tar.gz

# 2. 验证校验和
curl -LO https://github.com/cuihe500/vaulthub/releases/download/v0.2.0/vaulthub_darwin_arm64.tar.gz.sha256
shasum -a 256 -c vaulthub_darwin_arm64.tar.gz.sha256

# 3. 解压
tar -xzf vaulthub_darwin_arm64.tar.gz

# 4. 安装
sudo mv vaulthub_darwin_arm64 /usr/local/bin/vaulthub
sudo chmod +x /usr/local/bin/vaulthub

# 5. 验证
vaulthub version
```

## 使用 Docker 镜像

### 快速启动

```bash
docker run -d \
  --name vaulthub \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  ghcr.io/cuihe500/vaulthub:latest
```

### 使用 Docker Compose

```yaml
version: '3.8'

services:
  vaulthub:
    image: ghcr.io/cuihe500/vaulthub:v0.2.0
    container_name: vaulthub
    ports:
      - "8080:8080"
    volumes:
      - ./configs:/app/configs:ro
    environment:
      - GIN_MODE=release
      - VAULTHUB_CONFIG=/app/configs/config.toml
    depends_on:
      - mysql
      - redis
    restart: unless-stopped

  mysql:
    image: mariadb:10.11
    environment:
      MYSQL_ROOT_PASSWORD: your_password
      MYSQL_DATABASE: vaulthub
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  mysql_data:
  redis_data:
```

## 故障排查

### CI 失败

**问题**: 测试失败
```bash
# 本地运行测试
make test

# 查看测试覆盖率
make coverage
```

**问题**: 格式检查失败
```bash
# 自动格式化代码
make fmt

# 运行 lint 检查
make lint
```

### Release 失败

**问题**: 标签已存在
```bash
# 删除本地标签
git tag -d v0.2.0

# 删除远程标签
git push origin :refs/tags/v0.2.0

# 重新创建标签
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

**问题**: Docker 镜像推送权限不足
- 确认 GitHub Actions 有 `packages: write` 权限
- 检查 GitHub Container Registry 设置: Settings → Packages

## 本地构建验证

在推送标签前,建议先本地验证构建:

```bash
# 1. 清理旧构建
make clean

# 2. 运行测试
make test

# 3. 构建前端
make build-frontend

# 4. 构建生产版本
make build-prod

# 5. 验证二进制文件
./build/vaulthub version

# 6. 构建 Docker 镜像
make docker-build

# 7. 测试 Docker 镜像
make docker-run
```

## 最佳实践

### 语义化版本

遵循 [Semantic Versioning](https://semver.org/) 规范:

- **主版本号** (MAJOR): 不兼容的 API 变更
- **次版本号** (MINOR): 向后兼容的功能新增
- **修订号** (PATCH): 向后兼容的问题修复

示例:
```
v1.0.0 - 首个稳定版本
v1.1.0 - 新增功能
v1.1.1 - 修复 bug
v2.0.0 - 重大变更
```

### 发布前检查清单

- [ ] 更新 `docs/CHANGELOG.md`
- [ ] 本地运行 `make test` 确保测试通过
- [ ] 本地运行 `make build-prod` 确保构建成功
- [ ] 审查待发布的代码变更
- [ ] 确认版本号符合语义化规范
- [ ] 推送标签到 GitHub
- [ ] 等待 CI 完成并验证产物
- [ ] 测试 Docker 镜像是否正常运行

### 标签命名规范

```bash
# 正式版本
v1.0.0, v1.1.0, v2.0.0

# 预发布版本
v1.0.0-alpha.1
v1.0.0-beta.1
v1.0.0-rc.1
```

## 配置 GitHub Secrets (可选)

如果需要推送到私有 Docker Registry,配置以下 Secrets:

1. 进入 GitHub 仓库 → Settings → Secrets and variables → Actions
2. 添加以下 Secrets:
   - `DOCKER_REGISTRY`: 私有镜像仓库地址
   - `DOCKER_USERNAME`: 镜像仓库用户名
   - `DOCKER_PASSWORD`: 镜像仓库密码

## 参考资料

- [GitHub Actions 文档](https://docs.github.com/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [语义化版本规范](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
