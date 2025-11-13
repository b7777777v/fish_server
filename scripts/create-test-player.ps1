# PowerShell 脚本 - 创建测试玩家
# 用法: .\create-test-player.ps1 -Username alice [-Password mypass] [-Verbose] [-CreateOnly]

param(
    [Parameter(Mandatory=$true)]
    [string]$Username,

    [Parameter(Mandatory=$false)]
    [string]$Password = "test123456",

    [Parameter(Mandatory=$false)]
    [switch]$Verbose,

    [Parameter(Mandatory=$false)]
    [switch]$CreateOnly
)

# 设置错误处理
$ErrorActionPreference = "Stop"

# 颜色输出函数
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

Write-ColorOutput "============================================" "Cyan"
Write-ColorOutput "🐟 Fish Server - 测试玩家创建工具" "Cyan"
Write-ColorOutput "============================================" "Cyan"
Write-Host ""

Write-ColorOutput "正在创建测试玩家..." "Yellow"
Write-Host "用户名: $Username"
Write-Host "密码: $Password"
Write-Host ""

# 切换到项目根目录
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptPath
Set-Location $projectRoot

# 构建命令参数
$args = @(
    "run",
    "cmd/test-player/main.go",
    "-username", $Username,
    "-password", $Password
)

if ($Verbose) {
    $args += "-verbose"
}

if ($CreateOnly) {
    $args += "-create-only"
}

# 运行测试工具
try {
    & go @args
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-ColorOutput "✅ 完成！" "Green"
    } else {
        Write-ColorOutput "❌ 执行失败！" "Red"
        exit 1
    }
} catch {
    Write-ColorOutput "❌ 错误: $_" "Red"
    exit 1
}
