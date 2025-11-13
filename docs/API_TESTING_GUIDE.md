# Admin Server API 测试指南

本指南说明如何通过 Admin Server 的 REST API 创建测试玩家账户并验证完整的游戏流程。

## 📋 目录

- [API 概览](#api-概览)
- [快速开始](#快速开始)
- [详细 API 文档](#详细-api-文档)
- [完整测试流程](#完整测试流程)
- [使用脚本](#使用脚本)
- [故障排除](#故障排除)

## 🎯 API 概览

Admin Server 提供以下 REST API 端点用于用户管理：

### 认证相关 API（无需登录）

| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/v1/auth/register` | 注册新用户 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/guest-login` | 游客登录 |
| POST | `/api/v1/auth/oauth/callback` | OAuth 回调 |

### 用户相关 API（需要认证）

| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/v1/user/profile` | 获取用户资料 |
| PUT | `/api/v1/user/profile` | 更新用户资料 |

## 🚀 快速开始

### 前置要求

- Admin Server 运行在 `http://localhost:6060`
- 已安装 `curl` 命令行工具

### 3 步创建测试玩家

#### 步骤 1: 注册新用户

```bash
curl -X POST http://localhost:6060/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "player1",
    "password": "test123456"
  }'
```

**响应示例:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "player1",
    "nickname": "player1",
    "avatar_url": "",
    "is_guest": false
  }
}
```

#### 步骤 2: 保存 Token

```bash
# 将 Token 保存到环境变量
export TOKEN="<your_token_here>"
```

#### 步骤 3: 验证账户

```bash
curl -X GET http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

**完成！** 🎉 现在可以使用这个账户连接到游戏服务器。

## 📚 详细 API 文档

### 1. 注册新用户

创建一个新的用户账户。

**端点:** `POST /api/v1/auth/register`

**请求头:**
```
Content-Type: application/json
```

**请求体:**
```json
{
  "username": "string",  // 必需，用户名
  "password": "string",  // 必需，密码（最少6个字符）
  "nickname": "string"   // 可选，昵称（默认为用户名）
}
```

**成功响应:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "player1",
    "nickname": "player1",
    "avatar_url": "",
    "is_guest": false,
    "third_party_provider": "",
    "third_party_id": ""
  }
}
```

**错误响应:**
```json
{
  "error": "username already exists"
}
```

**curl 示例:**
```bash
curl -X POST http://localhost:6060/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testplayer",
    "password": "securepass123",
    "nickname": "测试玩家"
  }'
```

---

### 2. 用户登录

使用用户名和密码登录，获取 JWT Token。

**端点:** `POST /api/v1/auth/login`

**请求头:**
```
Content-Type: application/json
```

**请求体:**
```json
{
  "username": "string",  // 必需
  "password": "string"   // 必需
}
```

**成功响应:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**错误响应:** `401 Unauthorized`
```json
{
  "error": "invalid username or password"
}
```

**curl 示例:**
```bash
curl -X POST http://localhost:6060/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testplayer",
    "password": "securepass123"
  }'
```

---

### 3. 游客登录

创建并登录一个游客账户（无需用户名和密码）。

**端点:** `POST /api/v1/auth/guest-login`

**请求头:**
```
Content-Type: application/json
```

**请求体:** 无

**成功响应:** `200 OK`
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**curl 示例:**
```bash
curl -X POST http://localhost:6060/api/v1/auth/guest-login \
  -H "Content-Type: application/json"
```

---

### 4. 获取用户资料

获取当前登录用户的完整资料。

**端点:** `GET /api/v1/user/profile`

**请求头:**
```
Authorization: Bearer <token>
```

**成功响应:** `200 OK`
```json
{
  "id": 1,
  "username": "testplayer",
  "nickname": "测试玩家",
  "avatar_url": "https://example.com/avatar.jpg",
  "is_guest": false,
  "third_party_provider": "",
  "third_party_id": ""
}
```

**错误响应:** `401 Unauthorized`
```json
{
  "error": "unauthorized"
}
```

**curl 示例:**
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X GET http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN"
```

---

### 5. 更新用户资料

更新当前登录用户的昵称或头像。

**端点:** `PUT /api/v1/user/profile`

**请求头:**
```
Authorization: Bearer <token>
Content-Type: application/json
```

**请求体:**
```json
{
  "nickname": "string",   // 可选，新昵称
  "avatar_url": "string"  // 可选，新头像 URL
}
```

**成功响应:** `200 OK`
```json
{
  "message": "profile updated successfully"
}
```

**curl 示例:**
```bash
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X PUT http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nickname": "新昵称",
    "avatar_url": "https://example.com/new-avatar.jpg"
  }'
```

---

## 🔄 完整测试流程

以下是完整的测试流程，从创建账户到连接游戏服务器。

### 1. 准备环境变量

```bash
# Admin Server URL
export ADMIN_URL="http://localhost:6060"

# Game Server WebSocket URL
export GAME_WS_URL="ws://localhost:9090"

# 测试账户信息
export TEST_USERNAME="testplayer_$(date +%s)"
export TEST_PASSWORD="test123456"
```

### 2. 注册并获取 Token

```bash
# 注册新用户
RESPONSE=$(curl -s -X POST "$ADMIN_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")

# 提取 Token
TOKEN=$(echo "$RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
echo "Token: $TOKEN"

# 提取用户 ID
USER_ID=$(echo "$RESPONSE" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
echo "User ID: $USER_ID"
```

### 3. 验证账户信息

```bash
# 获取用户资料
curl -X GET "$ADMIN_URL/api/v1/user/profile" \
  -H "Authorization: Bearer $TOKEN" | jq
```

### 4. 更新用户资料

```bash
# 更新昵称
curl -X PUT "$ADMIN_URL/api/v1/user/profile" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nickname":"测试玩家VIP"}' | jq
```

### 5. 连接到游戏服务器

使用获取的 Token 连接到 Game Server：

**WebSocket URL:**
```
ws://localhost:9090?token=<your_token>
```

**使用 websocat 测试:**
```bash
# 安装 websocat: https://github.com/vi/websocat
echo '{"type":"HEARTBEAT"}' | websocat "${GAME_WS_URL}?token=${TOKEN}"
```

**使用浏览器测试:**
```javascript
// 在浏览器控制台中
const token = "your_token_here";
const ws = new WebSocket(`ws://localhost:9090?token=${token}`);

ws.onopen = () => console.log('Connected to game server');
ws.onmessage = (event) => console.log('Message:', event.data);
ws.onerror = (error) => console.error('Error:', error);
```

---

## 🛠️ 使用脚本

我们提供了便捷的脚本来自动化测试流程。

### 脚本 1: 创建测试玩家

创建单个测试玩家账户。

```bash
# 基本用法
./scripts/create-player-via-api.sh <username> [password]

# 示例
./scripts/create-player-via-api.sh player1
./scripts/create-player-via-api.sh player2 mypassword
```

**功能:**
- ✅ 注册新用户（如果失败则尝试登录）
- ✅ 获取并验证 Token
- ✅ 获取用户资料
- ✅ 保存 Token 到文件 (`.tokens/<username>.token`)

### 脚本 2: 完整游戏流程测试

测试从注册到游戏连接的完整流程。

```bash
# 基本用法
./scripts/test-game-flow-via-api.sh [username] [password]

# 使用默认值（自动生成用户名）
./scripts/test-game-flow-via-api.sh

# 指定用户名和密码
./scripts/test-game-flow-via-api.sh myplayer mypassword
```

**功能:**
- ✅ 注册/登录用户
- ✅ 获取用户资料
- ✅ 验证 Token
- ✅ 测试 WebSocket 连接（如果安装了 websocat）
- ✅ 输出完整的连接信息和测试命令
- ✅ 保存 Token 到文件

### 批量创建测试玩家

```bash
# 创建 10 个测试玩家
for i in {1..10}; do
  ./scripts/create-player-via-api.sh "player$i" "test123"
  sleep 1
done
```

---

## 📊 测试场景

### 场景 1: 单人游戏测试

```bash
# 1. 创建测试玩家
./scripts/test-game-flow-via-api.sh solo_player

# 2. 使用浏览器打开游戏
# file://path/to/fish_server/js/index.html

# 3. 使用创建的账户登录
# Username: solo_player
# Password: test123456
```

### 场景 2: 多人游戏测试

```bash
# 创建 4 个玩家
for i in {1..4}; do
  ./scripts/create-player-via-api.sh "player$i" "test123"
done

# 打开 4 个浏览器窗口
# 每个窗口使用不同的账户登录
```

### 场景 3: API 集成测试

```bash
# 测试完整的 API 流程
USERNAME="api_test_$(date +%s)"
PASSWORD="test123"

# 注册
REGISTER_RESP=$(curl -s -X POST http://localhost:6060/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

# 提取 Token
TOKEN=$(echo "$REGISTER_RESP" | jq -r '.token')

# 获取资料
curl -s -X GET http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" | jq

# 更新资料
curl -s -X PUT http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"nickname":"API测试用户"}' | jq

# 重新获取验证
curl -s -X GET http://localhost:6060/api/v1/user/profile \
  -H "Authorization: Bearer $TOKEN" | jq
```

---

## 🐛 故障排除

### 问题 1: 连接被拒绝

**错误:**
```
curl: (7) Failed to connect to localhost port 6060: Connection refused
```

**解决方案:**
```bash
# 检查 Admin Server 是否运行
ps aux | grep admin

# 检查端口占用
netstat -an | grep 6060

# 启动 Admin Server
go run cmd/admin/main.go
```

### 问题 2: 注册失败 - 用户已存在

**错误响应:**
```json
{
  "error": "username already exists"
}
```

**解决方案:**
```bash
# 方法 1: 使用不同的用户名
./scripts/create-player-via-api.sh player2

# 方法 2: 直接登录现有用户
curl -X POST http://localhost:6060/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"player1","password":"test123456"}'

# 方法 3: 删除现有用户（数据库操作）
psql -h localhost -U user -d fish_db -c "DELETE FROM users WHERE username='player1';"
```

### 问题 3: Token 无效

**错误响应:**
```json
{
  "error": "invalid token"
}
```

**解决方案:**
1. 检查 Token 是否正确复制（没有多余空格）
2. 检查 Token 是否过期（默认 2 小时）
3. 重新登录获取新 Token

```bash
# 重新登录
curl -X POST http://localhost:6060/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"your_username","password":"your_password"}'
```

### 问题 4: 数据库连接失败

**错误:**
```
database connection failed
```

**解决方案:**
```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 启动数据库
docker-compose -f deployments/docker-compose.dev.yml up -d postgres

# 测试数据库连接
psql -h localhost -U user -d fish_db -c "SELECT 1;"
```

### 问题 5: 密码太短

**错误响应:**
```json
{
  "error": "Key: 'RegisterRequest.Password' Error:Field validation for 'Password' failed on the 'min' tag"
}
```

**解决方案:**
使用至少 6 个字符的密码。

```bash
# ❌ 错误
curl -X POST http://localhost:6060/api/v1/auth/register \
  -d '{"username":"test","password":"123"}'

# ✅ 正确
curl -X POST http://localhost:6060/api/v1/auth/register \
  -d '{"username":"test","password":"123456"}'
```

---

## 📝 常见问题

### Q: 如何重置密码？

A: 目前 API 不支持密码重置。需要直接操作数据库：

```sql
-- 连接数据库
psql -h localhost -U user -d fish_db

-- 删除用户重新创建
DELETE FROM users WHERE username = 'player1';
```

### Q: Token 有效期多久？

A: 默认 2 小时（7200 秒），在 `configs/config.yaml` 中配置：

```yaml
jwt:
  expire: 7200  # 秒
```

### Q: 如何测试 WebSocket 连接？

A: 有几种方法：

1. **使用 websocat（推荐）:**
   ```bash
   echo '{"type":"HEARTBEAT"}' | websocat "ws://localhost:9090?token=$TOKEN"
   ```

2. **使用浏览器控制台:**
   ```javascript
   const ws = new WebSocket('ws://localhost:9090?token=your_token');
   ws.onopen = () => console.log('Connected');
   ws.onmessage = (e) => console.log('Message:', e.data);
   ```

3. **使用前端客户端:**
   打开 `js/index.html` 并使用测试账户登录。

### Q: 如何查看所有创建的测试玩家？

A: 连接数据库查询：

```bash
psql -h localhost -U user -d fish_db -c "
SELECT id, username, nickname, is_guest, created_at
FROM users
ORDER BY created_at DESC
LIMIT 20;"
```

---

## 📚 相关文档

- [项目整体说明](../README.md)
- [快速开始指南](../QUICKSTART.md)
- [编码规范](../CLAUDE.md)

---

## 🆘 获取帮助

如果遇到问题：

1. 查看服务器日志：
   ```bash
   tail -f logs/admin-server.log
   ```

2. 检查服务状态：
   ```bash
   curl http://localhost:6060/health
   ```

3. 查看数据库状态：
   ```bash
   psql -h localhost -U user -d fish_db -c "\dt"
   ```

---

**Happy Testing! 🎮🐟**
