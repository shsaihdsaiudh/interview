#!/bin/bash
# ============================================================
# 面试互助平台 — CentOS 一键部署脚本
# 使用方法：./deploy.sh <服务器IP> <用户名>
# 示例：./deploy.sh 123.45.67.89 root
# ============================================================

set -e

SERVER_IP="${1:?请提供服务器 IP}"
SERVER_USER="${2:?请提供用户名（如 root）}"
PROJECT_DIR="/opt/interview-platform"

echo "============================================"
echo " 面试互助平台 — 部署到 ${SERVER_USER}@${SERVER_IP}"
echo "============================================"

# ── 1. 检查本地文件 ──
echo ""
echo "[1/5] 检查本地项目文件..."
if [ ! -f "docker-compose.prod.yml" ]; then
    echo "❌ 未找到 docker-compose.prod.yml，请在项目根目录运行此脚本"
    exit 1
fi
echo "✅ 项目文件完整"

# ── 2. 在服务器上安装 Docker（如未安装）──
echo ""
echo "[2/5] 检查服务器 Docker 环境..."
ssh "${SERVER_USER}@${SERVER_IP}" << 'EOF'
    # 安装 Docker（适用于 CentOS 7/8/9）
    if ! command -v docker &> /dev/null; then
        echo "正在安装 Docker..."
        curl -fsSL https://get.docker.com | bash
        systemctl enable docker
        systemctl start docker
        echo "✅ Docker 安装完成"
    else
        echo "✅ Docker 已安装: $(docker --version)"
    fi

    # 安装 Docker Compose 插件（如未安装）
    if ! docker compose version &> /dev/null; then
        echo "正在安装 Docker Compose 插件..."
        yum install -y docker-compose-plugin 2>/dev/null || \
        curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
        chmod +x /usr/local/bin/docker-compose
        echo "✅ Docker Compose 安装完成"
    else
        echo "✅ Docker Compose 已安装: $(docker compose version)"
    fi

    # 确保 Docker 正在运行
    systemctl is-active --quiet docker || systemctl start docker
EOF

# ── 3. 上传项目文件到服务器 ──
echo ""
echo "[3/5] 上传项目文件到 ${SERVER_IP}..."
ssh "${SERVER_USER}@${SERVER_IP}" "mkdir -p ${PROJECT_DIR}"
rsync -avz --progress \
    --exclude='.git' \
    --exclude='node_modules' \
    --exclude='web/dist' \
    --exclude='.idea' \
    --exclude='.reasonix' \
    --exclude='~/' \
    --exclude='coverage.out' \
    --exclude='.DS_Store' \
    --exclude='.env' \
    ./ "${SERVER_USER}@${SERVER_IP}:${PROJECT_DIR}/"

# ── 4. 创建 .env 文件（如不存在）──
echo ""
echo "[4/5] 检查 .env 配置文件..."
ssh "${SERVER_USER}@${SERVER_IP}" << EOF
    cd ${PROJECT_DIR}
    if [ ! -f .env ]; then
        cp .env.production.example .env
        echo "⚠️  已创建 .env 模板文件，请编辑填写真实的 SMTP 密码:"
        echo "   ssh ${SERVER_USER}@${SERVER_IP} 'vi ${PROJECT_DIR}/.env'"
    else
        echo "✅ .env 文件已存在"
    fi
EOF

# ── 5. 构建并启动服务 ──
echo ""
echo "[5/5] 构建并启动 Docker 服务..."
ssh "${SERVER_USER}@${SERVER_IP}" << EOF
    cd ${PROJECT_DIR}
    docker compose -f docker-compose.prod.yml up -d --build
    echo ""
    echo "等待服务启动..."
    sleep 5
    docker compose -f docker-compose.prod.yml ps
EOF

echo ""
echo "============================================"
echo " 🎉 部署完成！"
echo "============================================"
echo ""
echo "访问地址: http://${SERVER_IP}"
echo ""
echo "常用命令（SSH 到服务器后执行）:"
echo "  cd ${PROJECT_DIR}"
echo "  docker compose -f docker-compose.prod.yml ps           # 查看服务状态"
echo "  docker compose -f docker-compose.prod.yml logs -f      # 查看日志"
echo "  docker compose -f docker-compose.prod.yml restart      # 重启服务"
echo "  docker compose -f docker-compose.prod.yml down         # 停止服务"
echo "  docker compose -f docker-compose.prod.yml pull         # 更新镜像"
echo ""
