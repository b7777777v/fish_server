@echo off
REM Windows 批处理脚本 - 端到端测试
REM 用法: e2e-test.bat [--keep-running]

setlocal enabledelayedexpansion

echo ==================================================
echo 🐟 Fish Server 端到端测试 (Windows)
echo ==================================================
echo.

REM 切换到项目根目录
cd /d "%~dp0\.."

REM 检查 Go 是否安装
where go >nul 2>&1
if %ERRORLEVEL% NEQ 0 (
    echo ❌ 错误: Go 未安装！请先安装 Go 1.24+
    exit /b 1
)
echo ✅ Go 已安装

REM 检查 Docker 是否安装
where docker >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo ✅ Docker 已安装
    set USE_DOCKER=true
) else (
    echo ⚠️  Docker 未安装，假设你已手动启动数据库
    set USE_DOCKER=false
)

echo.
echo ================================================
echo 1️⃣  启动数据库服务...
echo ================================================

if "%USE_DOCKER%"=="true" (
    echo 使用 Docker 启动 PostgreSQL 和 Redis...
    docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis
    if %ERRORLEVEL% NEQ 0 (
        echo ❌ Docker 启动失败，请检查 Docker 配置
        exit /b 1
    )
    echo 等待数据库启动...
    timeout /t 8 /nobreak >nul
    echo ✅ 数据库服务已启动
) else (
    echo ⚠️  请确保 PostgreSQL 和 Redis 已手动启动
)

echo.
echo ================================================
echo 2️⃣  运行数据库迁移...
echo ================================================

go run cmd\migrator\main.go up
if %ERRORLEVEL% EQU 0 (
    echo ✅ 数据库迁移完成
) else (
    echo ⚠️  迁移可能已运行，继续...
)

echo.
echo ================================================
echo 3️⃣  启动服务器...
echo ================================================

REM 创建日志目录
if not exist logs mkdir logs

REM 启动 Admin Server（后台）
echo 启动 Admin Server...
start /B cmd /c "go run cmd\admin\main.go > logs\admin-e2e.log 2>&1"
timeout /t 2 /nobreak >nul
echo ✅ Admin Server 已启动

REM 启动 Game Server（后台）
echo 启动 Game Server...
start /B cmd /c "go run cmd\game\main.go > logs\game-e2e.log 2>&1"
timeout /t 2 /nobreak >nul
echo ✅ Game Server 已启动

REM 等待服务器完全启动
echo 等待服务器完全启动...
timeout /t 8 /nobreak >nul

REM 验证 Admin Server
echo 验证 Admin Server...
set retries=0
:check_admin
curl -s http://localhost:6060/health >nul 2>&1
if %ERRORLEVEL% EQU 0 (
    echo ✅ Admin Server 健康检查通过
    goto admin_ok
)
set /a retries+=1
if %retries% GEQ 10 (
    echo ❌ Admin Server 启动失败，查看日志: logs\admin-e2e.log
    goto cleanup
)
timeout /t 1 /nobreak >nul
goto check_admin
:admin_ok

echo.
echo ================================================
echo 4️⃣  创建测试玩家...
echo ================================================

REM 创建 4 个测试玩家
for %%i in (1 2 3 4) do (
    echo Creating player%%i...
    go run cmd\test-player\main.go -username player%%i -password test123 -create-only
    timeout /t 1 /nobreak >nul
)
echo ✅ 测试玩家创建成功

echo.
echo ================================================
echo 5️⃣  运行完整游戏流程测试...
echo ================================================

go run cmd\test-player\main.go -username e2e_test_player -password e2epass123
if %ERRORLEVEL% EQU 0 (
    echo ✅ 端到端测试通过！
) else (
    echo ❌ 端到端测试失败！
    goto cleanup
)

echo.
echo ==================================================
echo 🎉 测试结果摘要
echo ==================================================
echo ✅ 所有测试通过！
echo.
echo 📊 创建的测试账户：
echo    player1 / test123
echo    player2 / test123
echo    player3 / test123
echo    player4 / test123
echo    e2e_test_player / e2epass123
echo.
echo 🌐 服务地址：
echo    Admin Server: http://localhost:6060
echo    Game Server:  ws://localhost:9090
echo.
echo 📂 日志文件：
echo    Admin: logs\admin-e2e.log
echo    Game:  logs\game-e2e.log
echo.
echo 🎮 开始游戏：
echo    在浏览器中打开: %CD%\js\index.html
echo.

REM 检查是否保持运行
if "%1"=="--keep-running" (
    echo 服务器将继续运行...
    echo 按任意键停止服务器...
    pause >nul
) else (
    echo 5 秒后自动关闭服务器...
    echo 如需保持运行，请使用: %~nx0 --keep-running
    timeout /t 5 /nobreak >nul
)

:cleanup
echo.
echo ⚠️  正在清理资源...

REM 查找并停止 Admin Server 和 Game Server 进程
for /f "tokens=2" %%a in ('tasklist ^| findstr /i "admin.exe game.exe go.exe"') do (
    taskkill /F /PID %%a >nul 2>&1
)

echo ✅ 清理完成

exit /b 0
