#!/bin/bash
# 使用 Admin Server API 创建测试玩家账户
# 用法: ./create-player-via-api.sh <username> [password]

set -e

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 配置
ADMIN_URL="${ADMIN_URL:-http://localhost:6060}"
USERNAME="$1"
PASSWORD="${2:-test123456}"

# 检查参数
if [ -z "$USERNAME" ]; then
    echo -e "${RED}错误: 必须提供用户名${NC}"
    echo "用法: $0 <username> [password]"
    echo "示例: $0 player1 mypassword"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}通过 Admin Server API 创建测试玩家${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Admin Server: $ADMIN_URL"
echo "用户名: $USERNAME"
echo "密码: $PASSWORD"
echo ""

# 步骤 1: 注册新用户
echo -e "${YELLOW}步骤 1: 注册新用户...${NC}"

REGISTER_RESPONSE=$(curl -s -X POST "$ADMIN_URL/api/v1/auth/register" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# 检查是否有错误
if echo "$REGISTER_RESPONSE" | grep -q '"error"'; then
    ERROR_MSG=$(echo "$REGISTER_RESPONSE" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
    echo -e "${RED}❌ 注册失败: $ERROR_MSG${NC}"

    # 如果用户已存在，尝试登录
    if echo "$ERROR_MSG" | grep -qi "already exists\|duplicate"; then
        echo -e "${YELLOW}用户已存在，尝试登录...${NC}"
    else
        exit 1
    fi
else
    TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    USER_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    NICKNAME=$(echo "$REGISTER_RESPONSE" | grep -o '"nickname":"[^"]*"' | cut -d'"' -f4)

    echo -e "${GREEN}✅ 注册成功!${NC}"
    echo "   用户 ID: $USER_ID"
    echo "   昵称: $NICKNAME"
    echo "   Token: ${TOKEN:0:50}..."
    echo ""
fi

# 步骤 2: 登录（如果注册失败或需要刷新 token）
if [ -z "$TOKEN" ]; then
    echo -e "${YELLOW}步骤 2: 用户登录...${NC}"

    LOGIN_RESPONSE=$(curl -s -X POST "$ADMIN_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

    if echo "$LOGIN_RESPONSE" | grep -q '"error"'; then
        ERROR_MSG=$(echo "$LOGIN_RESPONSE" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
        echo -e "${RED}❌ 登录失败: $ERROR_MSG${NC}"
        exit 1
    fi

    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo -e "${GREEN}✅ 登录成功!${NC}"
    echo "   Token: ${TOKEN:0:50}..."
    echo ""
fi

# 步骤 3: 获取用户资料
echo -e "${YELLOW}步骤 3: 获取用户资料...${NC}"

PROFILE_RESPONSE=$(curl -s -X GET "$ADMIN_URL/api/v1/user/profile" \
    -H "Authorization: Bearer $TOKEN")

if echo "$PROFILE_RESPONSE" | grep -q '"error"'; then
    ERROR_MSG=$(echo "$PROFILE_RESPONSE" | grep -o '"error":"[^"]*"' | cut -d'"' -f4)
    echo -e "${RED}❌ 获取资料失败: $ERROR_MSG${NC}"
    exit 1
fi

USER_ID=$(echo "$PROFILE_RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
USERNAME_FROM_PROFILE=$(echo "$PROFILE_RESPONSE" | grep -o '"username":"[^"]*"' | cut -d'"' -f4)
NICKNAME=$(echo "$PROFILE_RESPONSE" | grep -o '"nickname":"[^"]*"' | cut -d'"' -f4)
IS_GUEST=$(echo "$PROFILE_RESPONSE" | grep -o '"is_guest":[a-z]*' | cut -d':' -f2)

echo -e "${GREEN}✅ 用户资料获取成功!${NC}"
echo "   ID: $USER_ID"
echo "   用户名: $USERNAME_FROM_PROFILE"
echo "   昵称: $NICKNAME"
echo "   游客: $IS_GUEST"
echo ""

# 步骤 4: 保存 Token 到文件（可选）
TOKEN_FILE=".tokens/$USERNAME.token"
mkdir -p .tokens
echo "$TOKEN" > "$TOKEN_FILE"
echo -e "${GREEN}✅ Token 已保存到: $TOKEN_FILE${NC}"
echo ""

# 总结
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✅ 测试玩家创建成功!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "账户信息:"
echo "  用户名: $USERNAME"
echo "  密码: $PASSWORD"
echo "  用户 ID: $USER_ID"
echo "  Token: ${TOKEN:0:50}..."
echo ""
echo "可以使用此 Token 连接到游戏服务器:"
echo "  ws://localhost:9090/ws?token=$TOKEN"
echo ""
echo "或使用 curl 测试:"
echo "  curl -H \"Authorization: Bearer $TOKEN\" $ADMIN_URL/api/v1/user/profile"
echo ""
