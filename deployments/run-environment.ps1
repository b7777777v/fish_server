# ========================================
# 環境管理腳本 (PowerShell)
# ========================================

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("dev", "staging", "prod", "help")]
    [string]$Environment,
    
    [Parameter(Mandatory=$false)]
    [ValidateSet("build", "up", "start", "down", "stop", "restart", "logs", "status", "clean")]
    [string]$Command = "up"
)

# 函數：打印帶顏色的消息
function Write-Info { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Cyan }
function Write-Success { param([string]$Message) Write-Host "[SUCCESS] $Message" -ForegroundColor Green }
function Write-Warning { param([string]$Message) Write-Host "[WARNING] $Message" -ForegroundColor Yellow }
function Write-Error { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red }

# 顯示使用說明
function Show-Usage {
    Write-Host "用法: .\run-environment.ps1 -Environment <env> [-Command <cmd>]"
    Write-Host ""
    Write-Host "ENVIRONMENTS:"
    Write-Host "  dev       - 開發環境 (啟用 pprof, 詳細日誌)"
    Write-Host "  staging   - 預發布環境 (關閉 pprof, 中等安全性)"
    Write-Host "  prod      - 生產環境 (最高安全性, 最佳性能)"
    Write-Host ""
    Write-Host "COMMANDS:"
    Write-Host "  build     - 構建指定環境的鏡像"
    Write-Host "  up        - 啟動指定環境 [預設]"
    Write-Host "  down      - 停止指定環境"
    Write-Host "  restart   - 重啟指定環境"
    Write-Host "  logs      - 查看指定環境日誌"
    Write-Host "  status    - 查看指定環境狀態"
    Write-Host "  clean     - 清理指定環境"
    Write-Host ""
    Write-Host "範例:"
    Write-Host "  .\run-environment.ps1 -Environment dev -Command up"
    Write-Host "  .\run-environment.ps1 -Environment staging -Command build"
    Write-Host "  .\run-environment.ps1 -Environment prod -Command status"
}

# 檢查必要的文件
function Test-Prerequisites {
    param([string]$Env)
    
    Write-Info "檢查 $Env 環境的必要文件..."
    
    $files = @(
        "docker-compose.$Env.yml",
        "config-docker.$Env.yaml"
    )
    
    foreach ($file in $files) {
        if (-not (Test-Path $file)) {
            Write-Error "找不到 $file"
            exit 1
        }
    }
    
    if (-not (Test-Path ".env.$Env")) {
        Write-Warning "找不到 .env.$Env，將使用默認配置"
    }
    
    Write-Success "必要文件檢查完成"
}

# 構建鏡像
function Build-Image {
    param([string]$Env)
    
    Write-Info "構建 $Env 環境的 Docker 鏡像..."
    
    # 設置鏡像標籤
    $imageTag = "fish-server-admin:$Env"
    if ($Env -eq "dev") {
        $imageTag = "fish-server-admin:latest"
    }
    
    # 獲取構建時間
    $buildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    
    # 構建鏡像
    $buildArgs = @(
        "build"
        "-f", "Dockerfile.admin"
        "-t", $imageTag
        "--build-arg", "ENVIRONMENT=$Env"
        "--build-arg", "BUILDTIME=$buildTime"
        ".."
    )
    
    & docker @buildArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "鏡像構建完成: $imageTag"
    } else {
        Write-Error "鏡像構建失敗"
        exit 1
    }
}

# 啟動環境
function Start-Environment {
    param([string]$Env)
    
    Write-Info "啟動 $Env 環境..."
    
    # 檢查環境變量文件
    $envFile = ".env.$Env"
    $composeArgs = @(
        "-f", "docker-compose.$Env.yml"
    )
    
    if (Test-Path $envFile) {
        $composeArgs += "--env-file", $envFile
    }
    
    $composeArgs += "up", "-d"
    
    & docker-compose @composeArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "$Env 環境啟動成功"
        Show-Endpoints $Env
    } else {
        Write-Error "$Env 環境啟動失敗"
        exit 1
    }
}

# 顯示可用端點
function Show-Endpoints {
    param([string]$Env)
    
    Write-Info "可用端點:"
    
    switch ($Env) {
        "dev" {
            Write-Host "  - Admin API: http://localhost:6060"
            Write-Host "  - Environment Info: http://localhost:6060/admin/env"
            Write-Host "  - Pprof (已啟用): http://localhost:6060/debug/pprof/"
        }
        "staging" {
            Write-Host "  - Admin API: http://localhost:6061"
            Write-Host "  - Environment Info: http://localhost:6061/admin/env"
            Write-Host "  - Pprof: 已關閉 (訪問 /debug/pprof/disabled 查看原因)"
        }
        "prod" {
            Write-Host "  - Admin API: http://localhost:6062"
            Write-Host "  - Environment Info: http://localhost:6062/admin/env"
            Write-Host "  - Pprof: 已關閉 (生產環境)"
        }
    }
}

# 停止環境
function Stop-Environment {
    param([string]$Env)
    
    Write-Info "停止 $Env 環境..."
    
    & docker-compose -f "docker-compose.$Env.yml" down
    
    Write-Success "$Env 環境已停止"
}

# 重啟環境
function Restart-Environment {
    param([string]$Env)
    
    Write-Info "重啟 $Env 環境..."
    
    Stop-Environment $Env
    Start-Sleep -Seconds 2
    Start-Environment $Env
}

# 查看日誌
function Show-Logs {
    param([string]$Env)
    
    Write-Info "顯示 $Env 環境日誌..."
    
    & docker-compose -f "docker-compose.$Env.yml" logs -f admin
}

# 查看狀態
function Show-Status {
    param([string]$Env)
    
    Write-Info "$Env 環境狀態:"
    
    & docker-compose -f "docker-compose.$Env.yml" ps
    
    # 測試健康狀態
    $port = switch ($Env) {
        "dev" { 6060 }
        "staging" { 6061 }
        "prod" { 6062 }
    }
    
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:$port/ping" -TimeoutSec 5
        Write-Success "服務健康檢查通過"
    }
    catch {
        Write-Warning "服務健康檢查失敗: $_"
    }
}

# 清理環境
function Clean-Environment {
    param([string]$Env)
    
    Write-Info "清理 $Env 環境..."
    
    # 停止並移除容器
    & docker-compose -f "docker-compose.$Env.yml" down -v
    
    # 移除鏡像
    $imageTag = "fish-server-admin:$Env"
    if ($Env -eq "dev") {
        $imageTag = "fish-server-admin:latest"
    }
    
    try {
        & docker rmi $imageTag 2>$null
    }
    catch {
        # 忽略錯誤
    }
    
    Write-Success "$Env 環境清理完成"
}

# 主函數
function Main {
    if ($Environment -eq "help") {
        Show-Usage
        return
    }
    
    # 切換到腳本目錄
    $scriptDir = Split-Path -Parent $MyInvocation.ScriptName
    Set-Location $scriptDir
    
    # 檢查必要文件
    Test-Prerequisites $Environment
    
    # 執行命令
    switch ($Command.ToLower()) {
        "build" {
            Build-Image $Environment
        }
        { $_ -in @("up", "start") } {
            Start-Environment $Environment
        }
        { $_ -in @("down", "stop") } {
            Stop-Environment $Environment
        }
        "restart" {
            Restart-Environment $Environment
        }
        "logs" {
            Show-Logs $Environment
        }
        "status" {
            Show-Status $Environment
        }
        "clean" {
            Clean-Environment $Environment
        }
        default {
            Write-Error "未知命令: $Command"
            Show-Usage
            exit 1
        }
    }
}

# 執行主函數
Main