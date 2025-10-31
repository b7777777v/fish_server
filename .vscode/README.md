# 🚀 VS Code 多環境開發配置

## 📋 概述

此配置為 Fish Server 項目提供了完整的多環境開發支持，包括 DEV、STAGING 和 PROD 環境的快速啟動、調試和測試功能。

## 🎯 主要功能

### 🟢 DEV 環境 (開發)
- ✅ **Pprof 已啟用** - 完整性能分析
- ✅ 詳細調試日誌
- ✅ Gin Debug 模式
- ✅ SQL 查詢日誌

### 🟡 STAGING 環境 (預發布)
- ❌ **Pprof 已關閉** - 降低資源使用
- ✅ 中等日誌級別
- ✅ 生產級別安全設置

### 🔴 PROD 環境 (生產)
- ❌ **Pprof 強制關閉** - 最佳性能
- ✅ 最高安全級別
- ✅ 最小日誌輸出

## 🎮 使用方法

### 1. 啟動配置 (F5 或 Ctrl+F5)

#### Admin Server 啟動選項
```
🟢 Admin Server - DEV (Pprof ON)      # 開發環境，Pprof 啟用
🟡 Admin Server - STAGING (Pprof OFF) # 預發布環境，Pprof 關閉  
🔴 Admin Server - PROD (Secure)       # 生產環境，最高安全性
⚡ Admin Server - Auto Environment    # 動態選擇環境
```

#### Game Server 啟動選項
```
🎮 Game Server - DEV                  # 開發環境
🎮 Game Server - STAGING              # 預發布環境
```

#### 複合啟動 (同時啟動多個服務)
```
🚀 DEV Environment - All Services     # 啟動所有開發環境服務
🏗️ STAGING Environment - All Services # 啟動所有預發布環境服務
```

#### 調試選項
```
🔍 Debug Admin with Delve             # 使用 Delve 調試器
🧪 Test Admin Service                 # 運行測試
```

### 2. 任務執行 (Ctrl+Shift+P → Tasks: Run Task)

#### 構建任務
```
🔨 Build Admin - DEV                  # 構建開發版本
🔨 Build Admin - STAGING              # 構建預發布版本 (優化)
🔨 Build Admin - PROD                 # 構建生產版本 (最優化)
```

#### Docker 任務
```
🐳 Docker Build - DEV                 # 構建開發環境 Docker 鏡像
🐳 Docker Build - STAGING             # 構建預發布環境 Docker 鏡像
```

#### 環境管理
```
🚀 Start DEV Environment              # 啟動開發環境 Docker 服務
🏗️ Start STAGING Environment          # 啟動預發布環境 Docker 服務
🛑 Stop All Environments              # 停止所有環境
```

#### 驗證任務
```
✅ Verify DEV Pprof                   # 驗證開發環境 Pprof 可用
❌ Verify STAGING Pprof Disabled      # 驗證預發布環境 Pprof 已關閉
📊 Check Environment Info             # 檢查所有環境信息
```

#### 測試任務
```
🧪 Test All                          # 運行所有測試
🧪 Test Admin Service                # 只測試 Admin 服務
🧪 Test with Coverage                # 運行測試並生成覆蓋率報告
```

#### 清理任務
```
🧹 Clean Build Artifacts             # 清理構建產物
🧹 Clean Docker Images               # 清理 Docker 鏡像
```

## 🔧 配置詳解

### launch.json 重點功能

#### 1. 環境變量自動設置
每個啟動配置都會自動設置對應的環境變量：

```json
"env": {
    "ENVIRONMENT": "dev",           // 環境標識
    "FISH_ENVIRONMENT": "dev",      // 備用環境標識
    "LOG_LEVEL": "debug",           // 日誌級別
    "GIN_MODE": "debug"             // Gin 模式
}
```

#### 2. 配置文件自動選擇
根據環境自動傳遞對應的配置文件：

```json
"args": ["./configs/config.dev.yaml"]        // DEV 環境
"args": ["./configs/config.staging.yaml"]    // STAGING 環境
"args": ["./configs/config.prod.yaml"]       // PROD 環境
```

#### 3. 動態環境選擇
`⚡ Admin Server - Auto Environment` 配置允許運行時選擇：
- 環境 (DEV/STAGING/PROD)
- 日誌級別 (debug/info/warn/error)
- Gin 模式 (debug/release)

### tasks.json 重點功能

#### 1. 構建優化
不同環境使用不同的構建參數：

```bash
# DEV 環境 - 包含調試信息
go build -tags dev

# STAGING 環境 - 部分優化
go build -tags staging -ldflags "-s -w"

# PROD 環境 - 完全優化
go build -tags prod -ldflags "-s -w"
```

#### 2. 依賴任務
所有構建任務都會自動執行 `wire-gen`：

```json
"dependsOn": "wire-gen"
```

### settings.json 重點功能

#### 1. Go 語言工具配置
- 自動更新工具
- 使用 golangci-lint 進行代碼檢查
- 使用 goimports 格式化代碼

#### 2. 測試環境變量
測試時自動設置：

```json
"go.testEnvVars": {
    "ENVIRONMENT": "test",
    "LOG_LEVEL": "debug",
    "GIN_MODE": "test"
}
```

## 🎯 快速操作指南

### 開發流程

1. **啟動開發環境**
   - 按 `F5` → 選擇 `🟢 Admin Server - DEV (Pprof ON)`
   - 訪問 http://localhost:6060/debug/pprof/ 進行性能分析

2. **測試預發布環境**
   - 按 `F5` → 選擇 `🟡 Admin Server - STAGING (Pprof OFF)`
   - 訪問 http://localhost:6060/debug/pprof/disabled 確認 Pprof 已關閉

3. **驗證生產配置**
   - 按 `F5` → 選擇 `🔴 Admin Server - PROD (Secure)`
   - 確認所有安全設置生效

### 調試流程

1. **設置斷點**
   - 在代碼中點擊行號左側設置斷點

2. **啟動調試**
   - 按 `F5` → 選擇 `🔍 Debug Admin with Delve`

3. **調試控制**
   - `F10` - 單步執行
   - `F11` - 進入函數
   - `Shift+F11` - 跳出函數
   - `F5` - 繼續執行

### 測試流程

1. **運行所有測試**
   - `Ctrl+Shift+P` → `Tasks: Run Task` → `🧪 Test All`

2. **運行覆蓋率測試**
   - `Ctrl+Shift+P` → `Tasks: Run Task` → `🧪 Test with Coverage`

3. **查看覆蓋率報告**
   ```bash
   go tool cover -html=coverage.out
   ```

## 🔍 Pprof 驗證

### DEV 環境 (應該可用)
```bash
# 啟動 DEV 環境後執行
curl http://localhost:6060/debug/pprof/          # 應該返回 pprof 首頁
curl http://localhost:6060/admin/env | jq       # pprof_enabled: true
```

### STAGING 環境 (應該關閉)
```bash
# 啟動 STAGING 環境後執行  
curl http://localhost:6060/debug/pprof/disabled  # 應該返回說明信息
curl http://localhost:6060/admin/env | jq       # pprof_enabled: false
```

## 🛠️ 自定義配置

### 添加新環境

1. **創建配置文件**
   ```bash
   cp configs/config.dev.yaml configs/config.custom.yaml
   ```

2. **添加啟動配置**
   在 `launch.json` 中添加：
   ```json
   {
       "name": "🔧 Admin Server - CUSTOM",
       "type": "go",
       "request": "launch",
       "mode": "auto",
       "program": "${workspaceFolder}/cmd/admin",
       "args": ["./configs/config.custom.yaml"],
       "env": {
           "ENVIRONMENT": "custom",
           "LOG_LEVEL": "info"
       }
   }
   ```

### 修改環境變量

在對應的啟動配置中修改 `env` 部分：

```json
"env": {
    "ENVIRONMENT": "dev",
    "LOG_LEVEL": "debug",
    "CUSTOM_VAR": "custom_value"
}
```

## 🚨 故障排除

### 常見問題

1. **Wire 生成失敗**
   ```bash
   go install github.com/google/wire/cmd/wire@latest
   go generate ./...
   ```

2. **Delve 調試器問題**
   ```bash
   go install github.com/go-delve/delve/cmd/dlv@latest
   ```

3. **環境變量未生效**
   - 檢查 `launch.json` 中的 `env` 配置
   - 重啟 VS Code
   - 檢查配置文件路徑

4. **端口衝突**
   - 修改配置文件中的端口設置
   - 或者停止其他占用端口的服務

### 調試技巧

1. **查看環境信息**
   - 訪問 `/admin/env` 端點確認當前環境配置

2. **查看日誌**
   - 在 VS Code 終端中查看詳細日誌輸出

3. **驗證配置**
   - 使用任務 `📊 Check Environment Info` 檢查所有環境

## 📚 相關文檔

- [環境配置指南](../deployments/ENVIRONMENT-GUIDE.md)
- [Docker 部署指南](../deployments/README.md) 
- [實現總結](../deployments/IMPLEMENTATION-SUMMARY.md)

---

**提示**: 這些配置確保了在開發階段可以充分利用 Pprof 進行性能分析，同時在預發布和生產環境中關閉這些功能以優化資源使用。