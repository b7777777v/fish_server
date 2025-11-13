# 🪟 Windows 快速开始指南

专为 Windows 用户设计的 Fish Server 快速部署指南。

## 📋 前置要求

### 必需软件

1. **Go 1.24+**
   - 下载: https://go.dev/dl/
   - 安装后验证: `go version`

2. **PostgreSQL 16+**
   - 下载: https://www.postgresql.org/download/windows/
   - 或使用 Docker Desktop

3. **Redis**
   - 推荐使用 Docker Desktop
   - 或使用 Redis for Windows: https://github.com/tporadowski/redis/releases

4. **Docker Desktop (可选但推荐)**
   - 下载: https://www.docker.com/products/docker-desktop

### 可选软件

- **Git for Windows**: https://git-scm.com/download/win
- **VS Code**: https://code.visualstudio.com/
- **migrate CLI**: https://github.com/golang-migrate/migrate

## 🚀 三种启动方式

### 方式 1: PowerShell 自动化（最简单）⭐

使用现代化的 PowerShell 脚本，一键完成所有操作。

#### 步骤 1: 打开 PowerShell

```powershell
# 在项目根目录右键选择 "在终端中打开" 或
cd C:\path\to\fish_server
```

#### 步骤 2: 允许执行脚本（首次使用）

```powershell
# 以管理员身份运行 PowerShell，执行以下命令
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

#### 步骤 3: 运行端到端测试

```powershell
# 自动启动所有服务并创建测试玩家
.\scripts\e2e-test.ps1

# 保持服务运行（不自动关闭）
.\scripts\e2e-test.ps1 -KeepRunning
```

**完成！** 🎉 现在可以打开 `js\index.html` 开始游戏。

---

### 方式 2: 批处理脚本（传统方式）

使用经典的 .bat 批处理脚本。

#### 步骤 1: 双击运行

```cmd
# 直接双击文件
scripts\e2e-test.bat

# 或在 CMD 中运行
cd C:\path\to\fish_server
scripts\e2e-test.bat

# 保持服务运行
scripts\e2e-test.bat --keep-running
```

---

### 方式 3: 手动步骤（完全控制）

适合需要调试或自定义配置的开发者。

#### 步骤 1: 启动数据库

**使用 Docker Desktop:**

```powershell
# 启动 PostgreSQL 和 Redis
docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis

# 等待数据库启动
Start-Sleep -Seconds 5
```

**使用本地 PostgreSQL/Redis:**

确保服务已启动并运行在默认端口。

#### 步骤 2: 运行数据库迁移

```powershell
# 方法 A: 使用 Go
go run cmd\migrator\main.go up

# 方法 B: 使用批处理（如果有）
scripts\run-migration.bat up
```

#### 步骤 3: 启动服务器

**终端 1 - Admin Server:**

```powershell
go run cmd\admin\main.go
```

**终端 2 - Game Server:**

```powershell
go run cmd\game\main.go
```

#### 步骤 4: 创建测试玩家

**使用 PowerShell 脚本（推荐）:**

```powershell
# 创建单个玩家
.\scripts\create-test-player.ps1 -Username alice

# 创建并指定密码
.\scripts\create-test-player.ps1 -Username bob -Password mypass123

# 启用详细输出
.\scripts\create-test-player.ps1 -Username charlie -Verbose

# 只创建账户，不测试游戏流程
.\scripts\create-test-player.ps1 -Username dave -CreateOnly
```

**使用批处理脚本:**

```cmd
scripts\create-test-player.bat alice
scripts\create-test-player.bat bob mypass123
```

**使用 Go 命令:**

```powershell
# 创建单个玩家
go run cmd\test-player\main.go -username alice -password test123456

# 创建 4 个测试玩家
1..4 | ForEach-Object {
    go run cmd\test-player\main.go -username "player$_" -password "test123" -create-only
}
```

## 🎮 开始游戏

### 1. 打开游戏客户端

在文件资源管理器中，双击打开：

```
fish_server\js\index.html
```

或在浏览器中访问：

```
file:///C:/path/to/fish_server/js/index.html
```

### 2. 登入游戏

使用创建的测试账户登入：

- **用户名**: `player1`
- **密码**: `test123`

### 3. 多人游戏测试

打开多个浏览器窗口或标签页，使用不同账户：

- 窗口 1: `player1 / test123`
- 窗口 2: `player2 / test123`
- 窗口 3: `player3 / test123`
- 窗口 4: `player4 / test123`

## 📊 默认测试账户

| 用户名 | 密码 | 初始金币 | 用途 |
|--------|------|----------|------|
| player1 | test123 | 1000 | 多人测试 |
| player2 | test123 | 1000 | 多人测试 |
| player3 | test123 | 1000 | 多人测试 |
| player4 | test123 | 1000 | 多人测试 |
| e2e_test_player | e2epass123 | 1000 | 端到端测试 |

## 🛠️ 常用 PowerShell 命令

### 创建测试玩家

```powershell
# 基本用法
.\scripts\create-test-player.ps1 -Username alice

# 完整参数
.\scripts\create-test-player.ps1 `
    -Username bob `
    -Password mypass123 `
    -Verbose `
    -CreateOnly

# 批量创建（PowerShell 循环）
1..10 | ForEach-Object {
    $username = "testuser$_"
    .\scripts\create-test-player.ps1 -Username $username -CreateOnly
}
```

### 检查服务状态

```powershell
# 检查 Admin Server
Invoke-WebRequest -Uri "http://localhost:6060/health" -UseBasicParsing

# 检查进程
Get-Process | Where-Object { $_.ProcessName -like "*game*" -or $_.ProcessName -like "*admin*" }

# 检查端口占用
Get-NetTCPConnection -LocalPort 6060, 9090 | Format-Table
```

### 查看日志

```powershell
# 实时查看日志（PowerShell）
Get-Content logs\admin-server.log -Wait -Tail 50

# 搜索错误
Select-String -Path logs\*.log -Pattern "error" -CaseSensitive:$false
```

### 数据库操作

```powershell
# 连接数据库（需要 psql）
$env:PGPASSWORD = "password"
psql -h localhost -U user -d fish_db

# 查询玩家
psql -h localhost -U user -d fish_db -c "SELECT * FROM users;"

# 删除测试玩家
psql -h localhost -U user -d fish_db -c "DELETE FROM users WHERE username LIKE 'player%';"
```

### 停止服务

```powershell
# 停止所有 Go 进程
Get-Process | Where-Object { $_.ProcessName -eq "go" } | Stop-Process -Force

# 停止特定端口的进程
Get-NetTCPConnection -LocalPort 6060 | Select-Object -ExpandProperty OwningProcess | ForEach-Object { Stop-Process -Id $_ -Force }
Get-NetTCPConnection -LocalPort 9090 | Select-Object -ExpandProperty OwningProcess | ForEach-Object { Stop-Process -Id $_ -Force }
```

## 🐛 常见问题

### 问题 1: PowerShell 执行策略错误

**错误信息:**
```
.\scripts\e2e-test.ps1 : 无法加载文件，因为在此系统上禁止运行脚本
```

**解决方案:**
```powershell
# 以管理员身份运行 PowerShell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# 或临时允许
PowerShell -ExecutionPolicy Bypass -File .\scripts\e2e-test.ps1
```

### 问题 2: Docker Desktop 未启动

**错误信息:**
```
error during connect: ... Is the docker daemon running?
```

**解决方案:**
1. 打开 Docker Desktop 应用
2. 等待 Docker 启动完成（系统托盘图标变为绿色）
3. 重新运行脚本

### 问题 3: 端口已被占用

**错误信息:**
```
bind: address already in use
```

**解决方案:**
```powershell
# 查找占用端口的进程
Get-NetTCPConnection -LocalPort 6060
Get-NetTCPConnection -LocalPort 9090

# 停止进程
Stop-Process -Id <PID> -Force
```

### 问题 4: 数据库连接失败

**错误信息:**
```
connection refused
```

**解决方案:**
```powershell
# 检查 PostgreSQL 服务
Get-Service | Where-Object { $_.Name -like "*postgres*" }

# 启动服务
Start-Service postgresql-x64-16  # 服务名可能不同

# 或使用 Docker
docker-compose -f deployments\docker-compose.dev.yml up -d postgres
```

### 问题 5: Go 命令未找到

**错误信息:**
```
'go' 不是内部或外部命令
```

**解决方案:**
1. 确保已安装 Go
2. 将 Go 添加到系统 PATH:
   - 默认路径: `C:\Program Files\Go\bin`
   - 环境变量: `GOPATH\bin`
3. 重启 PowerShell/CMD

## 💡 开发技巧

### VS Code 配置

1. **打开项目**
   ```powershell
   code .
   ```

2. **使用调试配置**
   - 按 `F5` 启动调试
   - 选择 "🚀 DEV Environment - All Services"
   - 支持断点、变量检查等

3. **集成终端**
   - `` Ctrl+` `` 打开集成终端
   - 默认使用 PowerShell

### Git Bash（可选）

如果安装了 Git for Windows，可以使用 Git Bash 运行 Linux 脚本：

```bash
# 在 Git Bash 中
./scripts/create-test-player.sh alice
./scripts/e2e-test.sh
```

### 自动化部署脚本

创建一个 `deploy.ps1` 自动化脚本：

```powershell
# deploy.ps1
Write-Host "🚀 自动化部署 Fish Server" -ForegroundColor Cyan

# 1. 启动数据库
docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis

# 2. 运行迁移
go run cmd\migrator\main.go up

# 3. 构建服务
go build -o bin\admin.exe cmd\admin\main.go
go build -o bin\game.exe cmd\game\main.go

# 4. 启动服务（后台）
Start-Process -FilePath "bin\admin.exe" -WindowStyle Hidden
Start-Process -FilePath "bin\game.exe" -WindowStyle Hidden

Write-Host "✅ 部署完成！" -ForegroundColor Green
```

## 📚 相关文档

- [完整测试指南](./TEST_PLAYER_GUIDE.md)
- [项目说明](../README.md)
- [编码规范](../CLAUDE.md)
- [Linux/Mac 快速开始](../QUICKSTART.md)

## 🆘 获取帮助

### 命令帮助

```powershell
# Go 工具帮助
go run cmd\test-player\main.go -h

# PowerShell 脚本帮助
Get-Help .\scripts\create-test-player.ps1 -Full
```

### 查看日志

```powershell
# Admin Server
Get-Content logs\admin-e2e.log

# Game Server
Get-Content logs\game-e2e.log
```

### 调试模式

```powershell
# 启用详细输出
.\scripts\create-test-player.ps1 -Username alice -Verbose

# Go 运行时调试
$env:GODEBUG = "http2debug=1"
go run cmd\game\main.go
```

## 🎯 下一步

现在你已经成功设置了 Fish Server，可以：

1. **学习游戏机制**
   - 查看 [鱼群陣型指南](./FISH_FORMATION_GUIDE.md)
   - 了解 [前端动画系统](./FRONTEND_FISH_DYNAMICS_GUIDE.md)

2. **开发新功能**
   - 阅读 [编码规范](../CLAUDE.md)
   - 使用 VS Code 调试

3. **部署到生产环境**
   - 配置环境变量
   - 使用 Docker Compose 生产配置

---

**Happy Gaming on Windows! 🪟🎮🐟**
