# 🚀 Fish Server 快速开始指南

这份指南将帮助你在 5 分钟内启动鱼游戏服务器并创建测试玩家账户。

> **🪟 Windows 用户专属指南**: 请查看 [WINDOWS_QUICKSTART.md](docs/WINDOWS_QUICKSTART.md) 获取针对 Windows 优化的完整指南！

## 📋 前置要求

确保你已安装：

- Go 1.24+
- PostgreSQL 16+
- Redis 7+
- migrate CLI (可选，用于数据库迁移)

## 🎯 快速启动（3 步骤）

### 步骤 1: 启动数据库服务

**方法 A: 使用 Docker Compose（推荐）**

```bash
# 启动 PostgreSQL 和 Redis
docker-compose -f deployments/docker-compose.dev.yml up -d postgres redis

# 等待几秒让数据库完全启动
sleep 5
```

**方法 B: 手动启动（如果没有 Docker）**

```bash
# 启动 PostgreSQL（根据你的系统）
# Ubuntu/Debian:
sudo systemctl start postgresql

# macOS (使用 Homebrew):
brew services start postgresql

# 启动 Redis
# Ubuntu/Debian:
sudo systemctl start redis-server

# macOS (使用 Homebrew):
brew services start redis
```

### 步骤 2: 初始化数据库

```bash
# 运行数据库迁移
make migrate-up

# 你应该看到类似的输出：
# >> Applying database migrations...
# 1/u create_initial_tables (xxx.xxxs)
# 2/u add_fish_types (xxx.xxxs)
# ...
```

### 步骤 3: 启动服务器

**选项 A: 使用 VS Code（推荐，支持调试）**

1. 打开 VS Code
2. 按 `F5` 或点击 Run and Debug
3. 选择 "🚀 DEV Environment - All Services"

**选项 B: 使用终端（需要 2 个终端窗口）**

```bash
# 终端 1 - 启动 Admin Server
make run-admin

# 终端 2 - 启动 Game Server
make run-game
```

**选项 C: 后台运行（Linux/Mac）**

```bash
# 启动所有服务
make run-admin &
make run-game &

# 查看日志
tail -f logs/admin-server.log
tail -f logs/game-server.log
```

## 🎮 创建测试玩家

服务启动后，现在可以创建测试玩家了！

### 方法 1: 使用 Makefile（Linux/Mac）

```bash
# 创建单个玩家
make test-player USERNAME=alice

# 创建玩家并指定密码
make test-player USERNAME=bob PASSWORD=mypassword

# 只创建账户，不测试游戏流程
make test-player USERNAME=charlie CREATE_ONLY=1

# 启用详细输出
make test-player USERNAME=dave VERBOSE=1

# 创建 4 个测试玩家（用于多人游戏测试）
make create-test-players
```

### 方法 2: 使用脚本

**Linux/Mac:**
```bash
./scripts/create-test-player.sh alice
./scripts/create-test-player.sh bob mypassword
```

**Windows (PowerShell - 推荐):**
```powershell
.\scripts\create-test-player.ps1 -Username alice
.\scripts\create-test-player.ps1 -Username bob -Password mypassword
```

**Windows (批处理):**
```cmd
scripts\create-test-player.bat alice
scripts\create-test-player.bat bob mypassword
```

### 方法 3: 直接使用 Go

```bash
go run cmd/test-player/main.go -username alice -password test123
```

## ✅ 验证安装

### 1. 检查服务健康状态

```bash
# Admin Server
curl http://localhost:6060/health
# 应返回: {"status":"ok"}

# 你也可以在浏览器中访问
# http://localhost:6060/health
```

### 2. 查看创建的玩家

```bash
# 连接到数据库
psql -h localhost -U user -d fish_db

# 查询玩家
SELECT id, username, nickname, coins, created_at FROM users;

# 退出
\q
```

### 3. 使用前端客户端测试

```bash
# 在浏览器中打开前端客户端
open js/index.html
# 或直接双击 js/index.html 文件

# 使用刚创建的账户登入
# 用户名: alice
# 密码: test123456
```

## 🎯 完整示例：端到端测试

这个示例展示完整的游戏测试流程：

```bash
#!/bin/bash

echo "🐟 Fish Server 端到端测试"
echo "=========================="

# 1. 启动数据库
echo "1️⃣ 启动数据库..."
docker-compose -f deployments/docker-compose.dev.yml up -d postgres redis
sleep 5

# 2. 运行迁移
echo "2️⃣ 运行数据库迁移..."
make migrate-up

# 3. 启动服务器（后台）
echo "3️⃣ 启动服务器..."
make run-admin > logs/admin.log 2>&1 &
ADMIN_PID=$!
make run-game > logs/game.log 2>&1 &
GAME_PID=$!

# 等待服务器启动
echo "   等待服务器启动..."
sleep 5

# 4. 创建测试玩家
echo "4️⃣ 创建测试玩家..."
make create-test-players

# 5. 验证
echo "5️⃣ 验证安装..."
curl -s http://localhost:6060/health

echo ""
echo "✅ 完成！"
echo "   Admin Server PID: $ADMIN_PID"
echo "   Game Server PID: $GAME_PID"
echo ""
echo "测试玩家账户："
echo "   player1 / test123"
echo "   player2 / test123"
echo "   player3 / test123"
echo "   player4 / test123"
echo ""
echo "打开浏览器访问: file://$(pwd)/js/index.html"
echo ""
echo "停止服务器："
echo "   kill $ADMIN_PID $GAME_PID"
```

保存为 `scripts/e2e-test.sh`，然后运行：

```bash
chmod +x scripts/e2e-test.sh
./scripts/e2e-test.sh
```

## 🎮 开始游戏

### 使用前端客户端

1. **打开游戏客户端**
   ```bash
   # 浏览器中打开
   open js/index.html
   # 或直接双击文件
   ```

2. **输入测试账户**
   - 用户名: `player1`
   - 密码: `test123`
   - 点击 "Login"

3. **加入房间**
   - 查看可用房间列表
   - 点击 "Join Room"

4. **开始游戏**
   - 点击鱼发射子弹
   - 使用滚轮或按键切换炮台等级
   - 捕获鱼获得奖励

### 多人游戏测试

打开多个浏览器窗口（或不同浏览器），使用不同的测试账户登入：

- 窗口 1: player1 / test123
- 窗口 2: player2 / test123
- 窗口 3: player3 / test123
- 窗口 4: player4 / test123

所有玩家可以在同一个房间内一起游戏！

## 📊 测试数据一览

### 默认测试账户

使用 `make create-test-players` 创建的账户：

| 用户名 | 密码 | 初始金币 |
|--------|------|----------|
| player1 | test123 | 1000 |
| player2 | test123 | 1000 |
| player3 | test123 | 1000 |
| player4 | test123 | 1000 |

### 数据库连接信息

| 服务 | 地址 | 用户名 | 密码 | 数据库 |
|------|------|--------|------|--------|
| PostgreSQL | localhost:5432 | user | password | fish_db |
| Redis | localhost:6379 | - | - | db 0 |

### 服务端口

| 服务 | 端口 | 协议 | 用途 |
|------|------|------|------|
| Admin Server | 6060 | HTTP/REST | 用户管理、后台API |
| Game Server | 9090 | WebSocket | 游戏实时通信 |

## 🔧 常见问题

### Q: 数据库连接失败

**错误:**
```
failed to connect to database: connection refused
```

**解决方案:**
```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres
# 或
pg_isready -h localhost -p 5432

# 如果没运行，启动它
docker-compose -f deployments/docker-compose.dev.yml up -d postgres
```

### Q: 端口已被占用

**错误:**
```
bind: address already in use
```

**解决方案:**
```bash
# 查找占用端口的进程
lsof -i :6060  # Admin Server
lsof -i :9090  # Game Server

# 停止进程
kill -9 <PID>
```

### Q: 玩家注册失败 - 用户名已存在

这是正常的！如果用户已存在，测试工具会自动尝试登入。你可以：

1. 使用不同的用户名
2. 直接登入现有账户
3. 删除现有用户：
   ```sql
   psql -h localhost -U user -d fish_db
   DELETE FROM users WHERE username = 'alice';
   ```

### Q: 迁移失败 - 表已存在

**错误:**
```
error: relation "users" already exists
```

**解决方案:**
```bash
# 重置数据库
./scripts/reset-database.sh

# 重新运行迁移
make migrate-up
```

## 📝 下一步

恭喜！你已经成功设置了 Fish Server。现在你可以：

1. **开发新功能**
   - 阅读 [CLAUDE.md](./CLAUDE.md) 了解项目结构
   - 查看 [架构文档](./docs/)

2. **调试游戏**
   - 使用 VS Code 调试配置
   - 查看 [.vscode/README.md](./.vscode/README.md)

3. **部署到生产环境**
   - 使用 Docker Compose 生产配置
   - 配置环境变量

4. **深入学习**
   - [测试玩家详细指南](./docs/TEST_PLAYER_GUIDE.md)
   - [鱼群陣型系统](./docs/FISH_FORMATION_GUIDE.md)
   - [前端动画指南](./docs/FRONTEND_FISH_DYNAMICS_GUIDE.md)

## 🆘 获取帮助

- 查看所有 Make 命令: `make help`
- 查看测试工具帮助: `go run cmd/test-player/main.go -h`
- 查看项目文档: [docs/](./docs/)

---

**Happy Gaming! 🎮🐟**
