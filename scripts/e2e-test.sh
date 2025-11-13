#!/bin/bash
# 端到端测试脚本 - 自动化完整测试流程

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 函数：打印带颜色的消息
print_step() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 函数：清理资源
cleanup() {
    print_warning "清理资源..."
    if [ ! -z "$ADMIN_PID" ] && kill -0 $ADMIN_PID 2>/dev/null; then
        kill $ADMIN_PID
        print_success "已停止 Admin Server (PID: $ADMIN_PID)"
    fi
    if [ ! -z "$GAME_PID" ] && kill -0 $GAME_PID 2>/dev/null; then
        kill $GAME_PID
        print_success "已停止 Game Server (PID: $GAME_PID)"
    fi
}

# 捕获退出信号
trap cleanup EXIT

echo "=================================================="
echo "🐟 Fish Server 端到端测试"
echo "=================================================="
echo ""

# 切换到项目根目录
cd "$(dirname "$0")/.."

# 步骤 1: 检查前置条件
print_step "1️⃣ 检查前置条件..."

# 检查 Go
if ! command -v go &> /dev/null; then
    print_error "Go 未安装！请先安装 Go 1.24+"
    exit 1
fi
print_success "Go 已安装: $(go version)"

# 检查 PostgreSQL
if ! command -v psql &> /dev/null; then
    print_warning "psql 未安装，将无法验证数据库"
else
    print_success "PostgreSQL 客户端已安装"
fi

# 检查 Docker
if command -v docker &> /dev/null; then
    print_success "Docker 已安装"
    USE_DOCKER=true
else
    print_warning "Docker 未安装，将使用本地服务"
    USE_DOCKER=false
fi

echo ""

# 步骤 2: 启动数据库
print_step "2️⃣ 启动数据库服务..."

if [ "$USE_DOCKER" = true ]; then
    print_warning "使用 Docker 启动 PostgreSQL 和 Redis..."
    docker-compose -f deployments/docker-compose.dev.yml up -d postgres redis || {
        print_error "Docker 启动失败，请检查 Docker 配置"
        exit 1
    }
    print_success "等待数据库启动..."
    sleep 5
else
    print_warning "假设你已手动启动 PostgreSQL 和 Redis"
fi

# 验证数据库连接
print_warning "测试数据库连接..."
if PGPASSWORD=password psql -h localhost -U user -d fish_db -c "SELECT 1" &> /dev/null; then
    print_success "数据库连接成功"
else
    print_error "数据库连接失败，请检查 PostgreSQL 配置"
    exit 1
fi

echo ""

# 步骤 3: 运行数据库迁移
print_step "3️⃣ 运行数据库迁移..."

if make migrate-up; then
    print_success "数据库迁移完成"
else
    print_warning "迁移可能已运行，继续..."
fi

echo ""

# 步骤 4: 启动服务器
print_step "4️⃣ 启动服务器..."

# 创建日志目录
mkdir -p logs

# 启动 Admin Server
print_warning "启动 Admin Server..."
make run-admin > logs/admin-e2e.log 2>&1 &
ADMIN_PID=$!
print_success "Admin Server 已启动 (PID: $ADMIN_PID)"

# 启动 Game Server
print_warning "启动 Game Server..."
make run-game > logs/game-e2e.log 2>&1 &
GAME_PID=$!
print_success "Game Server 已启动 (PID: $GAME_PID)"

# 等待服务器启动
print_warning "等待服务器完全启动..."
sleep 8

# 验证服务器
print_warning "验证 Admin Server..."
for i in {1..10}; do
    if curl -s http://localhost:6060/health > /dev/null 2>&1; then
        print_success "Admin Server 健康检查通过"
        break
    fi
    if [ $i -eq 10 ]; then
        print_error "Admin Server 启动失败，查看日志: logs/admin-e2e.log"
        exit 1
    fi
    sleep 1
done

echo ""

# 步骤 5: 创建测试玩家
print_step "5️⃣ 创建测试玩家..."

if make create-test-players; then
    print_success "测试玩家创建成功"
else
    print_warning "部分玩家可能已存在，继续..."
fi

echo ""

# 步骤 6: 运行单个玩家完整测试
print_step "6️⃣ 运行完整游戏流程测试..."

if go run cmd/test-player/main.go -username e2e_test_player -password e2epass123; then
    print_success "端到端测试通过！"
else
    print_error "端到端测试失败！查看日志获取详细信息"
    exit 1
fi

echo ""

# 步骤 7: 显示结果
print_step "7️⃣ 测试结果摘要"
echo "=================================================="
print_success "所有测试通过！"
echo ""
echo "📊 创建的测试账户："
echo "   player1 / test123"
echo "   player2 / test123"
echo "   player3 / test123"
echo "   player4 / test123"
echo "   e2e_test_player / e2epass123"
echo ""
echo "🌐 服务地址："
echo "   Admin Server: http://localhost:6060"
echo "   Game Server:  ws://localhost:9090"
echo ""
echo "📂 日志文件："
echo "   Admin: logs/admin-e2e.log"
echo "   Game:  logs/game-e2e.log"
echo ""
echo "🎮 开始游戏："
echo "   在浏览器中打开: file://$(pwd)/js/index.html"
echo ""
echo "🛑 停止服务器："
echo "   kill $ADMIN_PID $GAME_PID"
echo ""
echo "=================================================="

# 如果提供了 --keep-running 参数，保持服务运行
if [ "$1" = "--keep-running" ]; then
    print_warning "服务器将继续运行..."
    print_warning "按 Ctrl+C 停止服务器"
    wait
else
    print_warning "5 秒后自动关闭服务器..."
    print_warning "如需保持运行，请使用: $0 --keep-running"
    sleep 5
fi
