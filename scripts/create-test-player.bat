@echo off
REM Windows批处理脚本 - 创建测试玩家

setlocal enabledelayedexpansion

echo ============================================
echo 🐟 Fish Server - 测试玩家创建工具
echo ============================================
echo.

REM 检查参数
if "%1"=="" (
    echo 用法: %0 ^<用户名^> [密码]
    echo 示例: %0 testplayer1 mypassword
    echo.
    echo 选项:
    echo   -v          启用详细输出
    echo   --create-only  只创建账户，不测试游戏流程
    exit /b 1
)

set USERNAME=%1
set PASSWORD=%2
if "%PASSWORD%"=="" set PASSWORD=test123456

set ADMIN_URL=http://localhost:6060
set GAME_URL=ws://localhost:9090
set VERBOSE=
set CREATE_ONLY=

REM 解析额外参数
:parse_args
shift
shift
if "%1"=="" goto run_test
if "%1"=="-v" (
    set VERBOSE=-verbose
    goto parse_args
)
if "%1"=="--create-only" (
    set CREATE_ONLY=-create-only
    goto parse_args
)

:run_test
echo 正在创建测试玩家...
echo 用户名: %USERNAME%
echo 密码: %PASSWORD%
echo.

cd /d "%~dp0\.."
go run cmd/test-player/main.go ^
    -username %USERNAME% ^
    -password %PASSWORD% ^
    -admin %ADMIN_URL% ^
    -game %GAME_URL% ^
    %VERBOSE% ^
    %CREATE_ONLY%

echo.
echo ✅ 完成！
pause
