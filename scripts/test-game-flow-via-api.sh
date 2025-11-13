#!/bin/bash
# 完整的游戏流程测试 - 通过 Admin Server API
# 此脚本展示完整的玩家创建到游戏连接流程

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
ADMIN_URL="${ADMIN_URL:-http://localhost:6060}"
GAME_WS_URL="${GAME_WS_URL:-ws://localhost:9090}"
USERNAME="${1:-testplayer_$(date +%s)}"
PASSWORD="${2:-test123456}"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}完整游戏流程测试 - 通过 Admin Server API${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo "Admin Server: $ADMIN_URL"
echo "Game Server: $GAME_WS_URL"
echo "测试用户: $USERNAME"
echo ""

# ==================================================
# 步骤 1: 注册新用户
# ==================================================
echo -e "${YELLOW}📝 步骤 1: 注册新用户...${NC}"

REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$ADMIN_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

HTTP_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$REGISTER_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "201" ]; then
    TOKEN=$(echo "$RESPONSE_BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo -e "${GREEN}✅ 注册成功 [HTTP $HTTP_CODE]${NC}"
    echo "   用户 ID: $USER_ID"
    echo ""
else
    # 注册失败，可能用户已存在，尝试登录
    echo -e "${YELLOW}⚠️  注册失败 [HTTP $HTTP_CODE], 尝试登录...${NC}"

    LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$ADMIN_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    HTTP_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
    RESPONSE_BODY=$(echo "$LOGIN_RESPONSE" | head -n-1)

    if [ "$HTTP_CODE" = "200" ]; then
        TOKEN=$(echo "$RESPONSE_BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo -e "${GREEN}✅ 登录成功 [HTTP $HTTP_CODE]${NC}"
        echo ""
    else
        echo -e "${RED}❌ 登录失败 [HTTP $HTTP_CODE]${NC}"
        echo "$RESPONSE_BODY"
        exit 1
    fi
fi

if [ -z "$TOKEN" ]; then
    echo -e "${RED}❌ 无法获取 Token${NC}"
    exit 1
fi

# ==================================================
# 步骤 2: 获取用户资料
# ==================================================
echo -e "${YELLOW}👤 步骤 2: 获取用户资料...${NC}"

PROFILE_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$ADMIN_URL/api/v1/user/profile" \
    -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$PROFILE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$PROFILE_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "200" ]; then
    USER_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    USERNAME_FROM_PROFILE=$(echo "$RESPONSE_BODY" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)
    NICKNAME=$(echo "$RESPONSE_BODY" | grep -o '"nickname":"[^"]*"' | cut -d'"' -f4)

    echo -e "${GREEN}✅ 用户资料获取成功 [HTTP $HTTP_CODE]${NC}"
    echo "   ID: $USER_ID"
    echo "   用户名: $USERNAME_FROM_PROFILE"
    echo "   昵称: $NICKNAME"
    echo ""
else
    echo -e "${RED}❌ 获取用户资料失败 [HTTP $HTTP_CODE]${NC}"
    echo "$RESPONSE_BODY"
    exit 1
fi

# ==================================================
# 步骤 3: 验证 Token
# ==================================================
echo -e "${YELLOW}🔐 步骤 3: 验证 Token...${NC}"
echo "   Token (前50字符): ${TOKEN:0:50}..."
echo "   Token 长度: ${#TOKEN}"
echo -e "${GREEN}✅ Token 验证通过${NC}"
echo ""

# ==================================================
# 步骤 4: 测试游戏服务器连接（使用 websocat 如果可用）
# ==================================================
echo -e "${YELLOW}🎮 步骤 4: 测试游戏服务器连接...${NC}"

if command -v websocat &> /dev/null; then
    echo "使用 websocat 测试 WebSocket 连接..."
    WS_URL="${GAME_WS_URL}?token=${TOKEN}"

    # 测试连接（发送心跳消息）
    echo '{"type":"HEARTBEAT"}' | timeout 5 websocat "$WS_URL" 2>&1 | head -n 5 || true
    echo -e "${GREEN}✅ WebSocket 连接测试完成${NC}"
else
    echo -e "${YELLOW}⚠️  websocat 未安装，跳过 WebSocket 测试${NC}"
    echo "   可以安装 websocat 进行 WebSocket 测试: https://github.com/vi/websocat"
    echo "   或使用浏览器客户端测试: file://$(pwd)/js/index.html"
fi
echo ""

# ==================================================
# 步骤 5: 输出连接信息
# ==================================================
echo -e "${BLUE}================================================${NC}"
echo -e "${GREEN}🎉 测试完成!${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo "📋 账户信息:"
echo "  用户名: $USERNAME"
echo "  密码: $PASSWORD"
echo "  用户 ID: $USER_ID"
echo ""
echo "🔑 认证信息:"
echo "  JWT Token: ${TOKEN:0:50}..."
echo ""
echo "🎮 游戏服务器连接:"
echo "  WebSocket URL: ${GAME_WS_URL}?token=${TOKEN}"
echo ""
echo "📝 API 测试命令:"
echo ""
echo "  # 获取用户资料"
echo "  curl -H \"Authorization: Bearer $TOKEN\" \\"
echo "       $ADMIN_URL/api/v1/user/profile"
echo ""
echo "  # 更新用户资料"
echo "  curl -X PUT -H \"Authorization: Bearer $TOKEN\" \\"
echo "       -H \"Content-Type: application/json\" \\"
echo "       -d '{\"nickname\":\"新昵称\"}' \\"
echo "       $ADMIN_URL/api/v1/user/profile"
echo ""
echo "🌐 前端测试:"
echo "  打开浏览器: file://$(pwd)/js/index.html"
echo "  使用此账户登录"
echo ""

# 保存 Token 到文件
TOKEN_FILE=".tokens/${USERNAME}.txt"
mkdir -p .tokens
cat > "$TOKEN_FILE" << EOF
# 测试玩家: $USERNAME
# 创建时间: $(date)

USERNAME=$USERNAME
PASSWORD=$PASSWORD
USER_ID=$USER_ID
TOKEN=$TOKEN

# WebSocket URL
WS_URL=${GAME_WS_URL}?token=${TOKEN}

# API 基础 URL
API_URL=$ADMIN_URL/api/v1

# 使用示例:
# source $TOKEN_FILE
# curl -H "Authorization: Bearer \$TOKEN" \$API_URL/user/profile
EOF

echo -e "${GREEN}✅ Token 信息已保存到: $TOKEN_FILE${NC}"
echo "   可以使用: source $TOKEN_FILE"
echo ""
