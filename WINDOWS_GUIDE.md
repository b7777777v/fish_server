# 🪟 Windows 使用指南

本指南專門為 Windows 用戶提供詳細的操作說明，無需使用 `make` 命令。

## 📋 前置要求

在開始之前，請確保已安裝：

1. **Go 1.24+**
   - 下載：https://golang.org/dl/
   - 安裝後確認：`go version`

2. **Docker Desktop for Windows**
   - 下載：https://www.docker.com/products/docker-desktop
   - 安裝後確認：`docker --version`

3. **Git for Windows** (可選)
   - 下載：https://git-scm.com/download/win

## 🚀 快速開始

### 1. 啟動數據庫

有兩種方式可以啟動數據庫：

#### 方式 A：使用批處理文件 (.bat)

```cmd
REM 在專案根目錄執行
scripts\start-database.bat
```

#### 方式 B：使用 PowerShell (.ps1)

```powershell
# 在專案根目錄執行
.\scripts\start-database.ps1
```

#### 方式 C：直接使用 Docker Compose

```cmd
docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis
```

### 2. 運行數據庫遷移

#### 使用批處理文件：

```cmd
REM 應用所有遷移
scripts\run-migration.bat up

REM 檢查版本
scripts\run-migration.bat version

REM 回滾最後一個遷移
scripts\run-migration.bat down
```

#### 使用 PowerShell：

```powershell
# 應用所有遷移
.\scripts\run-migration.ps1 up

# 檢查版本
.\scripts\run-migration.ps1 version

# 回滾最後一個遷移
.\scripts\run-migration.ps1 down
```

#### 直接使用 Go：

```cmd
REM 應用所有遷移
go run cmd\migrator\main.go up

REM 檢查版本
go run cmd\migrator\main.go version

REM 回滾最後一個遷移
go run cmd\migrator\main.go down
```

### 3. 修復 Dirty Migration

如果遇到 "Dirty database version 6" 錯誤：

#### 使用批處理文件：

```cmd
scripts\fix-dirty-migration.bat 5
```

#### 使用 PowerShell：

```powershell
.\scripts\fix-dirty-migration.ps1 -Version 5
```

#### 手動修復：

```cmd
REM 1. 強制設定版本為 5
go run cmd\migrator\main.go force 5

REM 2. 重新應用遷移
go run cmd\migrator\main.go up

REM 3. 驗證結果
go run cmd\migrator\main.go version
```

### 4. 停止數據庫

```cmd
REM 使用批處理文件
scripts\stop-database.bat

REM 或使用 Docker Compose
docker-compose -f deployments\docker-compose.dev.yml down
```

## 🔧 常用操作

### 編譯專案

```cmd
REM 建立 bin 目錄
mkdir bin

REM 編譯 Game Server
go build -o bin\game-server.exe cmd\game\main.go

REM 編譯 Admin Server
go build -o bin\admin-server.exe cmd\admin\main.go
```

### 生成代碼

```cmd
REM 生成 Protobuf 代碼 (需要先安裝 protoc)
.\scripts\proto-gen.sh

REM 生成 Wire 依賴注入代碼 (需要先安裝 wire)
.\scripts\wire-gen.sh
```

如果使用 Git Bash：
```bash
sh ./scripts/proto-gen.sh
sh ./scripts/wire-gen.sh
```

### 運行服務

```cmd
REM 運行 Game Server
.\bin\game-server.exe

REM 運行 Admin Server
.\bin\admin-server.exe
```

或直接使用 `go run`：

```cmd
REM 運行 Game Server
go run cmd\game\main.go

REM 運行 Admin Server
go run cmd\admin\main.go
```

### 運行測試

```cmd
REM 運行所有測試
go test -v -race -cover .\...

REM 運行特定包的測試
go test -v .\internal\biz\...
```

### 整理依賴

```cmd
go mod tidy
```

## 📝 可用腳本列表

| 腳本名稱 | 批處理 (.bat) | PowerShell (.ps1) | 說明 |
|---------|--------------|-------------------|------|
| 啟動數據庫 | `scripts\start-database.bat` | `.\scripts\start-database.ps1` | 啟動 PostgreSQL 和 Redis |
| 停止數據庫 | `scripts\stop-database.bat` | - | 停止所有數據庫服務 |
| **重置數據庫** | `scripts\reset-database.bat` | `.\scripts\reset-database.ps1` | **完全重置數據庫（推薦）** |
| 運行遷移 | `scripts\run-migration.bat [命令]` | `.\scripts\run-migration.ps1 [命令]` | 執行數據庫遷移 |

## 🔍 常見問題

### Q: PowerShell 腳本無法執行，提示安全錯誤

**A:** 需要修改執行策略。以管理員身份運行 PowerShell：

```powershell
# 查看當前策略
Get-ExecutionPolicy

# 設置為允許本地腳本執行
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# 或者只針對當前會話
Set-ExecutionPolicy Bypass -Scope Process
```

### Q: Docker 命令無法執行

**A:** 確保：
1. Docker Desktop 已經啟動
2. 在終端中執行 `docker info` 確認 Docker 正在運行
3. 如果使用 WSL2，確保 Docker Desktop 已啟用 WSL2 集成

### Q: 路徑分隔符問題

**A:** Windows 使用反斜線 `\` 作為路徑分隔符：

```cmd
REM 正確 ✓
scripts\start-database.bat
go run cmd\migrator\main.go

REM 錯誤 ✗ (Linux/Mac 風格)
scripts/start-database.bat
go run cmd/migrator/main.go
```

但在 Go 代碼中和 Git Bash 中可以使用正斜線 `/`。

### Q: 遇到任何遷移錯誤（包括 "already exists" 等）

**A:** 最簡單的解決方案是完全重置數據庫。

```cmd
REM 使用重置腳本（推薦）
scripts\reset-database.bat

REM 或使用 PowerShell
.\scripts\reset-database.ps1
```

這會刪除並重建整個數據庫，解決所有遷移問題。詳見 [DATABASE_MANAGEMENT.md](DATABASE_MANAGEMENT.md)

### Q: 資料庫連接失敗

**A:** 檢查以下事項：
1. Docker 容器是否正在運行：`docker ps`
2. PostgreSQL 是否準備就緒：
   ```cmd
   docker exec fish_server-postgres-1 pg_isready -U user -d fish_db
   ```
3. 端口是否被佔用：
   ```cmd
   netstat -ano | findstr :5432
   ```

### Q: Go 命令找不到

**A:** 確保 Go 已正確安裝並加入 PATH：

```cmd
REM 檢查 Go 版本
go version

REM 檢查 GOPATH
go env GOPATH

REM 如果找不到，需要將 Go 的 bin 目錄加入系統 PATH
REM 通常在：C:\Go\bin 或 C:\Program Files\Go\bin
```

### Q: 無法找到模組或依賴

**A:** 先下載依賴：

```cmd
REM 下載所有依賴
go mod download

REM 整理依賴
go mod tidy
```

## 💡 提示和技巧

### 1. 使用 Windows Terminal

建議安裝 [Windows Terminal](https://aka.ms/terminal)，它提供：
- 更好的顏色支持
- 多個標籤頁
- 支援 PowerShell、CMD、Git Bash 等多種 Shell

### 2. 使用 Git Bash

如果安裝了 Git for Windows，可以使用 Git Bash：
- 支援 Linux 風格的命令
- 可以直接運行 `.sh` 腳本
- 提供類似 Linux 的環境

### 3. 環境變數設定

在 PowerShell 中設定環境變數：

```powershell
# 臨時設定 (當前會話)
$env:LOG_LEVEL = "debug"

# 永久設定 (需要管理員權限)
[System.Environment]::SetEnvironmentVariable("LOG_LEVEL", "debug", "User")
```

在 CMD 中設定環境變數：

```cmd
REM 臨時設定
set LOG_LEVEL=debug

REM 查看環境變數
echo %LOG_LEVEL%
```

### 4. 查看日誌

```cmd
REM 查看所有容器日誌
docker-compose -f deployments\docker-compose.dev.yml logs

REM 持續監控日誌
docker-compose -f deployments\docker-compose.dev.yml logs -f

REM 只查看 PostgreSQL 日誌
docker-compose -f deployments\docker-compose.dev.yml logs postgres
```

### 5. 進入 PostgreSQL

```cmd
REM 使用 docker exec 進入 psql
docker exec -it fish_server-postgres-1 psql -U user -d fish_db

REM 在 psql 中執行常用命令：
REM \dt              - 列出所有表
REM \d users         - 查看 users 表結構
REM \di              - 列出所有索引
REM \q               - 退出 psql
```

## 🔗 相關資源

- [Go Windows 安裝指南](https://golang.org/doc/install)
- [Docker Desktop for Windows 文檔](https://docs.docker.com/desktop/windows/)
- [PowerShell 文檔](https://docs.microsoft.com/powershell/)
- [Windows Terminal 文檔](https://docs.microsoft.com/windows/terminal/)
- [Migration Fix Guide](MIGRATION_FIX_GUIDE.md) - 詳細的遷移修復指南
