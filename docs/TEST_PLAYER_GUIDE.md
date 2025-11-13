# 测试玩家创建和游戏流程验证指南

本指南说明如何创建测试玩家账户并验证完整的游戏流程。

## 📋 目录

- [工具概述](#工具概述)
- [前置要求](#前置要求)
- [快速开始](#快速开始)
- [详细使用说明](#详细使用说明)
- [测试流程说明](#测试流程说明)
- [故障排除](#故障排除)

## 🎯 工具概述

测试玩家工具（`cmd/test-player`）是一个综合性的测试工具，可以：

1. ✅ 创建新的测试玩家账户
2. ✅ 验证玩家登入功能
3. ✅ 获取玩家资料
4. ✅ 测试 WebSocket 连接
5. ✅ 验证游戏核心功能（房间列表、心跳、玩家信息等）

## 📦 前置要求

### 1. 启动所需服务

在创建测试玩家之前，确保以下服务已经运行：

```bash
# 方法1: 使用 Docker Compose（推荐）
docker-compose -f deployments/docker-compose.dev.yml up -d

# 方法2: 手动启动各服务
# 启动数据库
make run-dev

# 运行数据库迁移
make migrate-up

# 启动 Admin Server（新终端）
make run-admin

# 启动 Game Server（新终端）
make run-game
```

### 2. 验证服务状态

```bash
# 检查 Admin Server (端口 6060)
curl http://localhost:6060/health

# 检查 Game Server (端口 9090)
# Game Server 使用 WebSocket，可以通过浏览器连接测试
```

## 🚀 快速开始

### 使用脚本（推荐）

#### Linux/Mac

```bash
# 基本用法
./scripts/create-test-player.sh testplayer1

# 自定义密码
./scripts/create-test-player.sh testplayer1 mypassword123

# 启用详细输出
./scripts/create-test-player.sh testplayer1 mypassword123 -v

# 只创建账户，不测试游戏流程
./scripts/create-test-player.sh testplayer1 mypassword123 --create-only
```

#### Windows

```cmd
REM 基本用法
scripts\create-test-player.bat testplayer1

REM 自定义密码
scripts\create-test-player.bat testplayer1 mypassword123

REM 启用详细输出
scripts\create-test-player.bat testplayer1 mypassword123 -v
```

### 直接使用 Go 命令

```bash
# 进入项目根目录
cd fish_server

# 运行测试工具
go run cmd/test-player/main.go -username testplayer1 -password test123456

# 查看所有选项
go run cmd/test-player/main.go -h
```

## 📖 详细使用说明

### 命令行参数

| 参数 | 说明 | 默认值 | 必需 |
|------|------|--------|------|
| `-username` | 测试玩家的用户名 | 无 | ✅ |
| `-password` | 测试玩家的密码 | `test123456` | ❌ |
| `-admin` | Admin Server URL | `http://localhost:6060` | ❌ |
| `-game` | Game Server WebSocket URL | `ws://localhost:9090` | ❌ |
| `-create-only` | 只创建账户，不测试游戏流程 | `false` | ❌ |
| `-verbose` | 启用详细日志输出 | `false` | ❌ |

### 使用示例

#### 示例1: 创建基本测试玩家

```bash
go run cmd/test-player/main.go -username alice
```

**输出:**
```
🐟 鱼游戏测试工具
==================
Admin Server: http://localhost:6060
Game Server:  ws://localhost:9090
测试用户:     alice

✅ 玩家注册成功: alice
✅ 登入成功
   Token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOi...
   用户ID: 123
   昵称: alice

✅ 玩家资料验证成功

📡 连接到游戏服务器...
连接到: ws://localhost:9090?token=eyJhbGc...
✅ WebSocket连接成功
📋 等待欢迎消息...
   ✅ 等待欢迎消息成功
📋 获取房间列表...
   ✅ 获取房间列表成功
📋 发送心跳...
   ✅ 发送心跳成功
📋 获取玩家信息...
   ✅ 获取玩家信息成功

🎉 所有测试通过！
```

#### 示例2: 创建多个测试玩家

```bash
# 创建 4 个玩家用于测试多人游戏
for i in {1..4}; do
  go run cmd/test-player/main.go \
    -username "player$i" \
    -password "test123" \
    -create-only
done
```

#### 示例3: 详细调试模式

```bash
go run cmd/test-player/main.go \
  -username debugplayer \
  -password mypass \
  -verbose
```

**详细输出示例:**
```
✅ 玩家资料验证成功
   ID: 123
   用户名: debugplayer
   昵称: debugplayer
   头像:
   游客: false

📡 连接到游戏服务器...
   收到欢迎消息: Welcome to Fish Game Server!
   房间数量: 2
   房间1: ID=1, 玩家=2/4, 状态=WAITING
   房间2: ID=2, 玩家=0/4, 状态=WAITING
   服务器时间: 1699999999
   玩家ID: 123
   用户名: debugplayer
   余额: 1000
```

#### 示例4: 自定义服务器地址

```bash
# 测试远程服务器
go run cmd/test-player/main.go \
  -username testuser \
  -admin "http://192.168.1.100:6060" \
  -game "ws://192.168.1.100:9090"
```

## 🔍 测试流程说明

测试工具会按以下步骤验证整个游戏流程：

### 步骤 1: 玩家注册

- **API**: `POST /api/v1/auth/register`
- **功能**: 创建新的玩家账户
- **验证**:
  - ✅ 用户名唯一性
  - ✅ 密码加密存储
  - ✅ 初始金币（默认 1000）
  - ✅ JWT Token 生成

### 步骤 2: 玩家登入

- **API**: `POST /api/v1/auth/login`
- **功能**: 使用用户名和密码登入
- **验证**:
  - ✅ 凭证验证
  - ✅ Token 刷新
  - ✅ 用户资料返回

### 步骤 3: 获取玩家资料

- **API**: `GET /api/v1/user/profile`
- **功能**: 获取当前登入玩家的完整资料
- **验证**:
  - ✅ Token 认证
  - ✅ 资料完整性
  - ✅ 权限验证

### 步骤 4: WebSocket 连接

- **端点**: `ws://localhost:9090?token=<JWT>`
- **功能**: 建立游戏服务器的实时连接
- **验证**:
  - ✅ Token 验证
  - ✅ 连接建立
  - ✅ 接收欢迎消息

### 步骤 5: 游戏功能测试

#### 5.1 获取房间列表

- **消息类型**: `GET_ROOM_LIST`
- **验证**:
  - ✅ 房间信息正确
  - ✅ 玩家数量统计
  - ✅ 房间状态

#### 5.2 心跳保持

- **消息类型**: `HEARTBEAT`
- **验证**:
  - ✅ 连接保活
  - ✅ 服务器响应
  - ✅ 时间同步

#### 5.3 获取玩家信息

- **消息类型**: `GET_PLAYER_INFO`
- **验证**:
  - ✅ 玩家 ID
  - ✅ 余额信息
  - ✅ 游戏状态

## 🛠️ 故障排除

### 问题1: 注册失败 - 用户名已存在

**错误信息:**
```
❌ 注册失败（可能已存在）: 注册失败 [400]: username already exists
尝试直接登入...
✅ 登入成功
```

**解决方案:**
- 这是正常情况，工具会自动尝试登入
- 或者使用不同的用户名

### 问题2: 连接被拒绝

**错误信息:**
```
❌ HTTP请求失败: dial tcp [::1]:6060: connect: connection refused
```

**解决方案:**
```bash
# 1. 检查 Admin Server 是否运行
ps aux | grep admin

# 2. 检查端口占用
netstat -an | grep 6060

# 3. 启动 Admin Server
make run-admin
```

### 问题3: WebSocket 连接失败

**错误信息:**
```
❌ WebSocket连接失败: dial tcp [::1]:9090: connect: connection refused
```

**解决方案:**
```bash
# 1. 检查 Game Server 是否运行
ps aux | grep game

# 2. 检查端口占用
netstat -an | grep 9090

# 3. 启动 Game Server
make run-game
```

### 问题4: Token 认证失败

**错误信息:**
```
❌ 获取资料失败 [401]: unauthorized
```

**解决方案:**
1. 检查 JWT 配置（`configs/config.yaml`）
2. 确保 secret key 一致
3. 检查 token 是否过期

### 问题5: 数据库连接失败

**错误信息:**
```
❌ 登入失败 [500]: database connection failed
```

**解决方案:**
```bash
# 1. 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 2. 测试数据库连接
psql -h localhost -U user -d fish_db

# 3. 启动数据库
make run-dev
```

## 📊 测试场景示例

### 场景1: 多人游戏测试

创建 4 个测试玩家，模拟完整房间：

```bash
#!/bin/bash
# test-multiplayer.sh

for i in {1..4}; do
  echo "创建玩家 $i..."
  ./scripts/create-test-player.sh "player$i" "pass$i" --create-only
  sleep 1
done

echo "所有玩家创建完成！"
echo "现在可以使用前端客户端（js/index.html）进行多人测试"
```

### 场景2: 压力测试

快速创建大量玩家：

```bash
#!/bin/bash
# stress-test.sh

for i in {1..100}; do
  go run cmd/test-player/main.go \
    -username "stress_user_$i" \
    -password "test123" \
    -create-only &
done

wait
echo "创建了 100 个测试玩家"
```

### 场景3: 游客账户测试

使用 Admin API 创建游客账户：

```bash
curl -X POST http://localhost:6060/api/v1/auth/guest-login
```

## 📝 API 参考

### 认证相关 API

#### 注册新用户

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "testplayer",
  "password": "test123456"
}
```

**响应:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 123,
    "username": "testplayer",
    "nickname": "testplayer",
    "avatar_url": "",
    "is_guest": false
  }
}
```

#### 用户登入

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "testplayer",
  "password": "test123456"
}
```

#### 游客登入

```http
POST /api/v1/auth/guest-login
```

#### 获取用户资料

```http
GET /api/v1/user/profile
Authorization: Bearer <token>
```

### WebSocket 消息

#### 连接

```javascript
const ws = new WebSocket('ws://localhost:9090?token=' + authToken);
```

#### 获取房间列表

```protobuf
message GameMessage {
  MessageType type = 1;  // GET_ROOM_LIST
  GetRoomListRequest get_room_list = 2;
}
```

#### 心跳

```protobuf
message GameMessage {
  MessageType type = 1;  // HEARTBEAT
  HeartbeatRequest heartbeat = 2;
}
```

## 🎮 下一步

创建测试玩家后，你可以：

1. **使用前端客户端测试**
   ```bash
   # 打开浏览器访问
   open js/index.html
   ```

2. **使用 WebSocket 客户端测试**
   - Chrome DevTools
   - Postman
   - wscat

3. **查看玩家数据**
   ```sql
   -- 连接数据库
   psql -h localhost -U user -d fish_db

   -- 查询玩家
   SELECT * FROM users;
   SELECT * FROM wallets;
   ```

4. **监控服务器日志**
   ```bash
   # Game Server 日志
   tail -f logs/game-server.log

   # Admin Server 日志
   tail -f logs/admin-server.log
   ```

## 📚 相关文档

- [项目整体说明](../README.md)
- [VS Code 开发配置](../.vscode/README.md)
- [鱼群陣型指南](./FISH_FORMATION_GUIDE.md)
- [前端动画指南](./FRONTEND_FISH_DYNAMICS_GUIDE.md)

## 🤝 贡献

如果你发现问题或有改进建议，欢迎：

1. 提交 Issue
2. 发起 Pull Request
3. 更新文档

---

**Happy Testing! 🎮🐟**
