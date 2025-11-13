#!/bin/bash
# 创建测试玩家脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🐟 Fish Server - 测试玩家创建工具${NC}"
echo "========================================"

# 检查参数
if [ -z "$1" ]; then
    echo "用法: $0 <用户名> [密码]"
    echo "示例: $0 testplayer1 mypassword"
    echo ""
    echo "选项:"
    echo "  -v, --verbose    启用详细输出"
    echo "  --create-only    只创建账户，不测试游戏流程"
    exit 1
fi

USERNAME=$1
PASSWORD=${2:-"test123456"}

# 默认参数
ADMIN_URL="http://localhost:6060"
GAME_URL="ws://localhost:9090"
VERBOSE=""
CREATE_ONLY=""

# 解析额外参数
shift
shift 2>/dev/null || true
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE="-verbose"
            shift
            ;;
        --create-only)
            CREATE_ONLY="-create-only"
            shift
            ;;
        *)
            echo "未知选项: $1"
            exit 1
            ;;
    esac
done

echo -e "${YELLOW}正在创建测试玩家...${NC}"
echo "用户名: $USERNAME"
echo "密码: $PASSWORD"
echo ""

# 运行测试工具
cd "$(dirname "$0")/.."
go run cmd/test-player/main.go \
    -username "$USERNAME" \
    -password "$PASSWORD" \
    -admin "$ADMIN_URL" \
    -game "$GAME_URL" \
    $VERBOSE \
    $CREATE_ONLY

echo ""
echo -e "${GREEN}✅ 完成！${NC}"
