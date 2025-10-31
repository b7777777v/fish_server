# ========================================
# Admin Service Docker 構建腳本 (PowerShell)
# ========================================

param(
    [string]$Command = "all",
    [string]$Tag = "latest",
    [int]$Port = 6060,
    [switch]$Help
)

# 配置
$ImageName = "fish-server-admin"
$ContainerName = "fish-admin-test"

# 函數：打印帶顏色的消息
function Write-ColorMessage {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor $Color
}

function Write-Info { param([string]$Message) Write-ColorMessage $Message "Cyan" }
function Write-Success { param([string]$Message) Write-ColorMessage $Message "Green" }
function Write-Warning { param([string]$Message) Write-ColorMessage $Message "Yellow" }
function Write-Error { param([string]$Message) Write-ColorMessage $Message "Red" }

# 顯示使用說明
function Show-Usage {
    Write-Host "用法: .\build-admin.ps1 [OPTIONS]"
    Write-Host ""
    Write-Host "參數:"
    Write-Host "  -Command    構建命令 (build, test, all, clean, logs, stop) [默認: all]"
    Write-Host "  -Tag        鏡像標籤 [默認: latest]"
    Write-Host "  -Port       測試端口 [默認: 6060]"
    Write-Host "  -Help       顯示此幫助信息"
    Write-Host ""
    Write-Host "範例:"
    Write-Host "  .\build-admin.ps1 -Command build"
    Write-Host "  .\build-admin.ps1 -Command test -Port 8080"
    Write-Host "  .\build-admin.ps1 -Tag v1.0.0"
}

# 檢查 Docker 是否運行
function Test-Docker {
    try {
        docker info | Out-Null
        return $true
    }
    catch {
        Write-Error "Docker 未運行或無法訪問"
        return $false
    }
}

# 清理函數
function Invoke-Cleanup {
    Write-Info "清理臨時容器..."
    try {
        docker rm -f $ContainerName 2>$null
    }
    catch {
        # 忽略錯誤
    }
}

# 構建鏡像
function Build-Image {
    Write-Info "開始構建 Admin Service Docker 鏡像..."
    
    # 切換到項目根目錄
    $scriptDir = Split-Path -Parent $MyInvocation.ScriptName
    $projectRoot = Split-Path -Parent $scriptDir
    Set-Location $projectRoot
    
    # 構建鏡像
    $buildTime = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    
    $buildArgs = @(
        "build"
        "-f", "deployments/Dockerfile.admin"
        "-t", "${ImageName}:${Tag}"
        "--build-arg", "BUILDTIME=$buildTime"
        "."
    )
    
    & docker @buildArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "鏡像構建成功！"
        
        # 顯示鏡像信息
        Write-Info "鏡像信息："
        docker images "${ImageName}:${Tag}"
        return $true
    }
    else {
        Write-Error "鏡像構建失敗！"
        return $false
    }
}

# 測試鏡像
function Test-Image {
    Write-Info "開始測試 Docker 鏡像..."
    
    # 檢查是否有同名容器在運行
    $existingContainer = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $ContainerName }
    if ($existingContainer) {
        Write-Warning "發現同名容器，正在移除..."
        docker rm -f $ContainerName
    }
    
    # 啟動測試容器
    Write-Info "啟動測試容器..."
    $runArgs = @(
        "run", "-d"
        "--name", $ContainerName
        "-p", "${Port}:6060"
        "-e", "LOG_LEVEL=debug"
        "${ImageName}:${Tag}"
    )
    
    & docker @runArgs
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "容器啟動失敗！"
        return $false
    }
    
    # 等待容器啟動
    Write-Info "等待容器啟動..."
    Start-Sleep -Seconds 5
    
    # 檢查容器狀態
    $runningContainer = docker ps --format "{{.Names}}" | Where-Object { $_ -eq $ContainerName }
    if (-not $runningContainer) {
        Write-Error "容器啟動失敗！"
        Write-Info "容器日誌："
        docker logs $ContainerName
        return $false
    }
    
    # 測試健康檢查端點
    Write-Info "測試健康檢查端點..."
    $healthCheckPassed = $false
    
    for ($i = 1; $i -le 10; $i++) {
        try {
            $response = Invoke-RestMethod -Uri "http://localhost:$Port/ping" -TimeoutSec 5
            Write-Success "健康檢查通過！"
            $healthCheckPassed = $true
            break
        }
        catch {
            if ($i -eq 10) {
                Write-Error "健康檢查失敗！"
                Write-Info "容器日誌："
                docker logs $ContainerName
                return $false
            }
            Write-Info "等待服務啟動... ($i/10)"
            Start-Sleep -Seconds 2
        }
    }
    
    if ($healthCheckPassed) {
        # 測試主要端點
        Write-Info "測試主要端點..."
        
        # 測試根端點
        try {
            $rootResponse = Invoke-RestMethod -Uri "http://localhost:$Port/" -TimeoutSec 5
            if ($rootResponse -match "Fish Server Admin API") {
                Write-Success "根端點測試通過"
            }
            else {
                Write-Warning "根端點測試失敗"
            }
        }
        catch {
            Write-Warning "根端點測試失敗: $_"
        }
        
        # 測試健康檢查端點
        try {
            $healthResponse = Invoke-RestMethod -Uri "http://localhost:$Port/admin/health" -TimeoutSec 5
            if ($healthResponse.status -eq "healthy") {
                Write-Success "健康檢查端點測試通過"
            }
            else {
                Write-Warning "健康檢查端點測試失敗"
            }
        }
        catch {
            Write-Warning "健康檢查端點測試失敗: $_"
        }
        
        # 顯示容器信息
        Write-Info "容器信息："
        docker ps --filter "name=$ContainerName" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
        
        Write-Success "所有測試通過！"
        return $true
    }
    
    return $false
}

# 清理所有
function Invoke-CleanAll {
    Write-Info "清理所有相關的容器和鏡像..."
    
    # 停止並移除容器
    try {
        docker rm -f $ContainerName 2>$null
    }
    catch {
        # 忽略錯誤
    }
    
    # 移除鏡像
    try {
        docker rmi "${ImageName}:${Tag}" 2>$null
    }
    catch {
        # 忽略錯誤
    }
    
    Write-Success "清理完成！"
}

# 查看日誌
function Show-Logs {
    $existingContainer = docker ps -a --format "{{.Names}}" | Where-Object { $_ -eq $ContainerName }
    if ($existingContainer) {
        Write-Info "顯示容器日誌："
        docker logs -f $ContainerName
    }
    else {
        Write-Error "測試容器不存在！"
        exit 1
    }
}

# 停止容器
function Stop-Container {
    Write-Info "停止測試容器..."
    try {
        docker stop $ContainerName 2>$null
        docker rm $ContainerName 2>$null
        Write-Success "容器已停止並移除"
    }
    catch {
        Write-Warning "容器可能不存在或已停止"
    }
}

# 主函數
function Main {
    if ($Help) {
        Show-Usage
        return
    }
    
    Write-Info "開始執行 Admin Service Docker 構建流程..."
    Write-Info "鏡像名稱: ${ImageName}:${Tag}"
    Write-Info "測試端口: $Port"
    
    # 檢查 Docker
    if (-not (Test-Docker)) {
        exit 1
    }
    
    # 設置清理
    Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Invoke-Cleanup }
    
    try {
        switch ($Command.ToLower()) {
            "build" {
                if (-not (Build-Image)) { exit 1 }
            }
            "test" {
                if (-not (Test-Image)) { exit 1 }
            }
            "all" {
                if (-not (Build-Image)) { exit 1 }
                if (-not (Test-Image)) { exit 1 }
            }
            "clean" {
                Invoke-CleanAll
            }
            "logs" {
                Show-Logs
            }
            "stop" {
                Stop-Container
            }
            default {
                Write-Error "未知命令: $Command"
                Show-Usage
                exit 1
            }
        }
        
        Write-Success "操作完成！"
    }
    finally {
        # 清理（如果需要）
        if ($Command.ToLower() -in @("test", "all")) {
            # 保留測試容器用於檢查
        }
    }
}

# 執行主函數
Main