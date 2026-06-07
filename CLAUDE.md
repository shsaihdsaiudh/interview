# 面试互助平台 - 项目笔记

## 服务器信息

- **地址**: ggggs.icu (47.109.128.53)
- **用户**: root
- **项目路径**: /opt/interview-platform
- **内存**: 1.6GB（低配，不能在线构建）

## 部署流程

服务器只有 **1.6GB 内存**，无法在服务器上编译 Go 和前端（会 OOM）。
必须 **本地交叉编译后上传预编译产物**，再通过 Docker 打包运行。

### 一键部署步骤

```bash
# 1. 本地编译 Go（交叉编译到 Linux amd64）
cd server
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o interview-server ./cmd/main.go

# 2. 本地编译前端
cd ../web
npm run build

# 3. 上传项目文件（排除源码和 node_modules）
cd ..
rsync -avz --progress \
    --exclude='.git' --exclude='node_modules' --exclude='web/dist' \
    --exclude='.idea' --exclude='.reasonix' --exclude='~/' \
    --exclude='coverage.out' --exclude='.DS_Store' --exclude='.env' \
    ./ root@ggggs.icu:/opt/interview-platform/

# 4. 上传编译产物
scp server/interview-server root@ggggs.icu:/opt/interview-platform/server/
scp -r web/dist root@ggggs.icu:/opt/interview-platform/web/

# 5. 在服务器上重建并启动
ssh root@ggggs.icu "cd /opt/interview-platform && docker compose -f docker-compose.prod.yml up -d --build"
```

### Dockerfile 说明

- `server/Dockerfile` — 完整多阶段构建（在服务器上编译 Go），**不要用**
- `server/Dockerfile.prod` — 只用预编译的 `interview-server` 二进制打包镜像，**生产用这个**
- `web/Dockerfile` — 完整多阶段构建（在服务器上 npm build），**不要用**
- `web/Dockerfile.prod` — 只用预构建的 `dist/` 打包 nginx 镜像，**生产用这个**

### .env 生产配置

服务器上的 `.env` 必须包含以下变量（docker-compose.prod.yml 使用 `${VAR:?error}` 强制检查）：

```bash
POSTGRES_PASSWORD=<数据库密码>
JWT_SECRET=<随机密钥>
ADMIN_EMAIL=<你的管理员邮箱>@std.uestc.edu.cn
SMTP_HOST=smtp.std.uestc.edu.cn
SMTP_PORT=465
SMTP_USER=<学号>@std.uestc.edu.cn
SMTP_PASS=<客户端专用密码>
SMTP_TLS_SERVERNAME=icoremail.net
```

- `ADMIN_EMAIL` 配置的管理员邮箱对应的用户注册后，启动时将自动提升为管理员角色
- 管理员后台入口：`/admin`（仅管理员角色可访问，普通用户访问会跳回首页）

生成随机 JWT_SECRET：`openssl rand -base64 32`

## 安全注意事项

- JWT 密钥从 `JWT_SECRET` 环境变量读取，代码中只有开发默认值（带警告日志）
- 数据库密码从 `POSTGRES_PASSWORD` 环境变量读取
- 认证接口已添加 IP 速率限制（`interfaces/middleware/ratelimit.go`）
- `.env` 文件已在 `.gitignore` 中，不会被提交到 GitHub
- 生产 PostgreSQL 不暴露端口到宿主机（仅 Docker 内网可访问）

## 技术栈

- 后端: Go 1.26 + Gin + pgx + JWT
- 前端: React + TypeScript + Vite + Tailwind CSS
- 数据库: PostgreSQL 17
- 部署: Docker Compose（Nginx 反向代理 + Go 后端 + PostgreSQL）
- HTTPS: Let's Encrypt 证书，自动续期 + Docker nginx reload hook

## HTTPS / TLS
- 证书由 Let's Encrypt (certbot) 管理，位于 `/etc/letsencrypt/live/ggggs.icu/`
- Docker nginx 通过 volumes 挂载证书文件（只读）
- certbot 每 12 小时检查一次续期（systemd timer）
- 续期后自动执行 `/etc/letsencrypt/renewal-hooks/deploy/reload-docker-nginx.sh` 重载 Docker nginx
- 当前证书有效期至 2026-08-28

## 已知问题 & 踩坑记录

### 部署时端口 80 冲突
服务器上有一个 host-level 的 nginx（可能是 Certbot/Let's Encrypt 装的），占用 80 端口。
Docker 前端容器需要先停掉它：
```bash
ssh root@ggggs.icu "systemctl stop nginx; systemctl disable nginx"
```
然后再 `docker compose up -d`。如果容器已经创建但端口绑定损坏（`invalid IP`），需要 `docker rm` 后重建。

### Docker Hub 镜像拉取慢
国内服务器拉 Docker Hub 镜像可能极慢（下载 20MB nginx 镜像要 7+ 分钟）。
已配置了 registry mirrors，但稳定性看运气。如果遇到超时，多试几次 `docker pull`。
