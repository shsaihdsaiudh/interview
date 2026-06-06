# ── 阶段一：构建 React 前端 ──
FROM node:22-alpine AS builder

WORKDIR /build

# 使用国内 npm 镜像
RUN npm config set registry https://registry.npmmirror.com

# 先复制依赖文件，利用 Docker 缓存层
COPY package.json package-lock.json ./
RUN npm ci

# 复制源码并构建
COPY . .
RUN npm run build

# ── 阶段二：Nginx 服务静态文件 ──
FROM nginx:alpine

# 复制构建产物
COPY --from=builder /build/dist /usr/share/nginx/html

# 复制 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
