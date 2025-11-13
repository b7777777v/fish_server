# PowerShell 脚本 - 端到端测试
# 用法: .\e2e-test.ps1 [-KeepRunning]

param(
    [Parameter(Mandatory=$false)]
    [switch]$KeepRunning
)

$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-Step {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠️  $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

# 清理函数
$AdminProcess = $null
$GameProcess = $null

function Cleanup {
    Write-Warning "清理资源..."

    if ($AdminProcess -and !$AdminProcess.HasExited) {
        Stop-Process -Id $AdminProcess.Id -Force -ErrorAction SilentlyContinue
        Write-Success "已停止 Admin Server (PID: $($AdminProcess.Id))"
    }

    if ($GameProcess -and !$GameProcess.HasExited) {
        Stop-Process -Id $GameProcess.Id -Force -ErrorAction SilentlyContinue
        Write-Success "已停止 Game Server (PID: $($GameProcess.Id))"
    }
}

# 注册清理事件
Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Cleanup }

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "🐟 Fish Server 端到端测试 (PowerShell)" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# 切换到项目根目录
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptPath
Set-Location $projectRoot

# 步骤 1: 检查前置条件
Write-Step "1️⃣  检查前置条件..."

# 检查 Go
try {
    $goVersion = & go version
    Write-Success "Go 已安装: $goVersion"
} catch {
    Write-Error "Go 未安装！请先安装 Go 1.24+"
    exit 1
}

# 检查 Docker
$useDocker = $false
try {
    $dockerVersion = & docker --version
    Write-Success "Docker 已安装"
    $useDocker = $true
} catch {
    Write-Warning "Docker 未安装，将使用本地服务"
}

Write-Host ""

# 步骤 2: 启动数据库
Write-Step "2️⃣  启动数据库服务..."

if ($useDocker) {
    Write-Warning "使用 Docker 启动 PostgreSQL 和 Redis..."
    & docker-compose -f deployments/docker-compose.dev.yml up -d postgres redis
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Docker 启动失败，请检查 Docker 配置"
        exit 1
    }
    Write-Success "等待数据库启动..."
    Start-Sleep -Seconds 5
} else {
    Write-Warning "假设你已手动启动 PostgreSQL 和 Redis"
}

# 验证数据库连接（可选）
Write-Warning "测试数据库连接..."
$env:PGPASSWORD = "password"
try {
    $result = & psql -h localhost -U user -d fish_db -c "SELECT 1" 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Success "数据库连接成功"
    }
} catch {
    Write-Warning "无法验证数据库连接（psql 可能未安装），继续..."
}

Write-Host ""

# 步骤 3: 运行数据库迁移
Write-Step "3️⃣  运行数据库迁移..."

try {
    & go run cmd/migrator/main.go up
    Write-Success "数据库迁移完成"
} catch {
    Write-Warning "迁移可能已运行，继续..."
}

Write-Host ""

# 步骤 4: 启动服务器
Write-Step "4️⃣  启动服务器..."

# 创建日志目录
if (!(Test-Path "logs")) {
    New-Item -ItemType Directory -Path "logs" | Out-Null
}

# 启动 Admin Server
Write-Warning "启动 Admin Server..."
$AdminProcess = Start-Process -FilePath "go" -ArgumentList "run", "cmd/admin/main.go" -RedirectStandardOutput "logs/admin-e2e.log" -RedirectStandardError "logs/admin-e2e-err.log" -PassThru -NoNewWindow
Write-Success "Admin Server 已启动 (PID: $($AdminProcess.Id))"

# 启动 Game Server
Write-Warning "启动 Game Server..."
$GameProcess = Start-Process -FilePath "go" -ArgumentList "run", "cmd/game/main.go" -RedirectStandardOutput "logs/game-e2e.log" -RedirectStandardError "logs/game-e2e-err.log" -PassThru -NoNewWindow
Write-Success "Game Server 已启动 (PID: $($GameProcess.Id))"

# 等待服务器启动
Write-Warning "等待服务器完全启动..."
Start-Sleep -Seconds 8

# 验证服务器
Write-Warning "验证 Admin Server..."
$retries = 0
$maxRetries = 10
$adminOk = $false

while ($retries -lt $maxRetries) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:6060/health" -UseBasicParsing -ErrorAction SilentlyContinue
        if ($response.StatusCode -eq 200) {
            Write-Success "Admin Server 健康检查通过"
            $adminOk = $true
            break
        }
    } catch {
        # 继续重试
    }
    $retries++
    Start-Sleep -Seconds 1
}

if (!$adminOk) {
    Write-Error "Admin Server 启动失败，查看日志: logs/admin-e2e.log"
    Cleanup
    exit 1
}

Write-Host ""

# 步骤 5: 创建测试玩家
Write-Step "5️⃣  创建测试玩家..."

1..4 | ForEach-Object {
    $playerName = "player$_"
    Write-Host "Creating $playerName..."
    & go run cmd/test-player/main.go -username $playerName -password "test123" -create-only 2>&1 | Out-Null
    Start-Sleep -Seconds 1
}
Write-Success "测试玩家创建成功"

Write-Host ""

# 步骤 6: 运行完整测试
Write-Step "6️⃣  运行完整游戏流程测试..."

& go run cmd/test-player/main.go -username "e2e_test_player" -password "e2epass123"
if ($LASTEXITCODE -eq 0) {
    Write-Success "端到端测试通过！"
} else {
    Write-Error "端到端测试失败！查看日志获取详细信息"
    Cleanup
    exit 1
}

Write-Host ""

# 步骤 7: 显示结果
Write-Step "7️⃣  测试结果摘要"
Write-Host "==================================================" -ForegroundColor Cyan
Write-Success "所有测试通过！"
Write-Host ""
Write-Host "📊 创建的测试账户："
Write-Host "   player1 / test123"
Write-Host "   player2 / test123"
Write-Host "   player3 / test123"
Write-Host "   player4 / test123"
Write-Host "   e2e_test_player / e2epass123"
Write-Host ""
Write-Host "🌐 服务地址："
Write-Host "   Admin Server: http://localhost:6060"
Write-Host "   Game Server:  ws://localhost:9090"
Write-Host ""
Write-Host "📂 日志文件："
Write-Host "   Admin: logs/admin-e2e.log"
Write-Host "   Game:  logs/game-e2e.log"
Write-Host ""
Write-Host "🎮 开始游戏："
Write-Host "   在浏览器中打开: $PWD\js\index.html"
Write-Host ""
Write-Host "🛑 停止服务器："
Write-Host "   Admin PID: $($AdminProcess.Id)"
Write-Host "   Game PID: $($GameProcess.Id)"
Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan

# 保持运行或自动关闭
if ($KeepRunning) {
    Write-Warning "服务器将继续运行..."
    Write-Warning "按 Ctrl+C 停止服务器"

    # 等待进程
    Wait-Process -Id $AdminProcess.Id, $GameProcess.Id
} else {
    Write-Warning "5 秒后自动关闭服务器..."
    Write-Warning "如需保持运行，请使用: .\e2e-test.ps1 -KeepRunning"
    Start-Sleep -Seconds 5
    Cleanup
}
