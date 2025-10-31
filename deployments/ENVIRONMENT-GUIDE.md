# 🌍 Fish Server 環境配置指南

## 📋 概述

本指南說明如何在不同環境中配置和運行 Fish Server Admin 服務，實現環境隔離和資源優化。

## 🏗️ 環境架構

### 開發環境 (DEV)
- **目的**: 本地開發和測試
- **特性**: 
  - ✅ **啟用 Pprof** - 性能分析和調試
  - ✅ 詳細調試日誌
  - ✅ Gin Debug 模式
  - ✅ SQL 查詢日誌
  - ❌ 不啟用安全限制
  - ❌ 不啟用 CORS 限制

### 預發布環境 (STAGING)
- **目的**: 生產前測試和驗證
- **特性**: 
  - ❌ **關閉 Pprof** - 降低資源使用
  - ✅ 中等日誌級別
  - ✅ 啟用認證和安全檢查
  - ✅ 啟用 CORS 限制
  - ✅ 啟用限流保護

### 生產環境 (PROD)
- **目的**: 正式服務運行
- **特性**: 
  - ❌ **強制關閉 Pprof** - 最佳性能和安全性
  - ✅ 最小日誌級別
  - ✅ 最高安全級別
  - ✅ 嚴格的 CORS 和認證
  - ✅ 嚴格的限流控制

## 📁 文件結構

```
deployments/
├── # Docker Compose 配置
├── docker-compose.dev.yml         # 開發環境
├── docker-compose.staging.yml     # 預發布環境
├── docker-compose.prod.yml        # 生產環境 (待創建)
│
├── # 應用配置
├── config-docker.dev.yaml         # 開發環境配置
├── config-docker.staging.yaml     # 預發布環境配置
├── config-docker.prod.yaml        # 生產環境配置 (待創建)
│
├── # 環境變量
├── .env.example                   # 環境變量範本
├── .env.dev                       # 開發環境變量
├── .env.staging                   # 預發布環境變量
├── .env.prod                      # 生產環境變量 (需創建)
│
├── # 管理腳本
├── run-environment.sh             # Linux/Mac 環境管理腳本
├── run-environment.ps1            # Windows 環境管理腳本
│
└── # 文檔
    ├── ENVIRONMENT-GUIDE.md       # 本文檔
    └── README.md                  # Docker 部署指南
```

## 🚀 快速開始

### 1. 選擇運行方式

#### 方式一：使用環境管理腳本 (推薦)

**Linux/Mac:**
```bash
# 賦予執行權限
chmod +x deployments/run-environment.sh

# 啟動開發環境
./deployments/run-environment.sh dev up

# 啟動預發布環境
./deployments/run-environment.sh staging up
```

**Windows PowerShell:**
```powershell
# 啟動開發環境
.\deployments\run-environment.ps1 -Environment dev -Command up

# 啟動預發布環境
.\deployments\run-environment.ps1 -Environment staging -Command up
```

#### 方式二：直接使用 Docker Compose

```bash
# 開發環境
docker-compose -f deployments/docker-compose.dev.yml --env-file deployments/.env.dev up -d

# 預發布環境
docker-compose -f deployments/docker-compose.staging.yml --env-file deployments/.env.staging up -d
```

### 2. 驗證環境

訪問對應的環境端點：

| 環境 | Admin API | 環境信息 | Pprof 狀態 |
|------|----------|----------|------------|
| DEV | http://localhost:6060 | http://localhost:6060/admin/env | ✅ 已啟用 |
| STAGING | http://localhost:6061 | http://localhost:6061/admin/env | ❌ 已關閉 |

## 🔧 環境配置詳解

### Pprof 配置對比

| 項目 | DEV | STAGING | PROD |
|------|-----|---------|------|
| `enable_pprof` | ✅ `true` | ❌ `false` | ❌ `false` |
| `pprof_auth` | ❌ `false` | ✅ `true` | ✅ `true` |
| 可用端點 | 全部 | 無 | 無 |
| 認證要求 | 無 | 需要密鑰 | 需要密鑰 |

### 資源使用對比

| 項目 | DEV | STAGING | PROD |
|------|-----|---------|------|
| 記憶體使用 | 高 (調試模式) | 中 | 低 (優化模式) |
| CPU 使用 | 高 (詳細日誌) | 中 | 低 |
| 網路連線 | 長時間 | 中等 | 短時間 |
| 日誌輸出 | 大量 | 適中 | 最少 |

## 🛠️ 管理命令

### 基本操作

```bash
# 構建特定環境的鏡像
./run-environment.sh <env> build

# 啟動環境
./run-environment.sh <env> up

# 停止環境
./run-environment.sh <env> down

# 重啟環境
./run-environment.sh <env> restart

# 查看日誌
./run-environment.sh <env> logs

# 查看狀態
./run-environment.sh <env> status

# 清理環境
./run-environment.sh <env> clean
```

### 高級操作

```bash
# 查看所有環境狀態
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 查看資源使用情況
docker stats

# 進入容器調試
docker exec -it <container_name> /bin/sh
```

## 🔍 Pprof 功能驗證

### 開發環境 (已啟用)

```bash
# 測試 Pprof 可用性
curl http://localhost:6060/debug/pprof/

# CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 記憶體分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 查看 Goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

### 預發布/生產環境 (已關閉)

```bash
# 驗證 Pprof 已關閉
curl http://localhost:6061/debug/pprof/disabled

# 回應範例:
{
  "message": "Pprof is disabled in this environment",
  "environment": "staging",
  "reason": "Performance profiling is disabled for security and resource optimization",
  "alternatives": {
    "metrics": "/admin/metrics",
    "status": "/admin/status",
    "health": "/admin/health"
  }
}
```

## 📊 環境監控

### 環境信息端點

訪問 `/admin/env` 查看當前環境配置：

```json
{
  "environment": "staging",
  "features": {
    "pprof_enabled": false,
    "pprof_auth": true,
    "gin_debug": false,
    "sql_debug": false,
    "rate_limit": true,
    "cors_enabled": true
  },
  "security": {
    "csrf_enabled": true,
    "secure_headers": true
  },
  "timestamp": "2024-10-31T10:00:00Z"
}
```

### 健康檢查

```bash
# 基本健康檢查
curl http://localhost:6060/ping

# 詳細健康檢查
curl http://localhost:6060/admin/health

# 系統狀態
curl http://localhost:6060/admin/status
```

## 🛡️ 安全考慮

### 開發環境
- ⚠️ Pprof 無認證 - 僅限本地使用
- ⚠️ 詳細錯誤信息 - 便於調試
- ⚠️ 允許所有 CORS - 開發便利

### 預發布環境
- ✅ Pprof 已關閉 - 降低攻擊面
- ✅ 限制 CORS 來源
- ✅ 啟用限流保護
- ✅ 隱藏詳細錯誤信息

### 生產環境
- ✅ 最高安全級別
- ✅ 強制 HTTPS (建議)
- ✅ 嚴格的認證和授權
- ✅ 完整的安全標頭

## 🐛 故障排除

### 常見問題

1. **Pprof 無法訪問**
   ```bash
   # 檢查環境配置
   curl http://localhost:6060/admin/env
   
   # 確認環境是否為 dev
   docker logs <container_name> | grep "Pprof"
   ```

2. **環境變量未生效**
   ```bash
   # 檢查環境變量文件
   cat deployments/.env.dev
   
   # 重新啟動環境
   ./run-environment.sh dev restart
   ```

3. **端口衝突**
   ```bash
   # 檢查端口使用情況
   netstat -tulpn | grep :6060
   
   # 修改 docker-compose 文件中的端口映射
   ```

### 性能對比測試

```bash
# 開發環境 (Pprof 啟用)
docker stats fish-dev-admin

# 預發布環境 (Pprof 關閉)
docker stats fish-staging-admin

# 比較記憶體和 CPU 使用差異
```

## 📈 最佳實踐

1. **環境隔離**: 不同環境使用不同的數據庫和 Redis
2. **資源優化**: 非開發環境關閉調試功能
3. **安全第一**: 生產環境強制關閉 Pprof
4. **監控**: 定期檢查各環境的資源使用情況
5. **文檔**: 保持環境配置文檔的更新

---

**注意**: 此環境配置已經過測試，確保在 DEV 環境中啟用 Pprof，在 STAGING 和 PROD 環境中關閉以降低資源使用。