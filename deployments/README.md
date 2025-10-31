# Fish Server Admin Docker 部署指南

## 📦 Docker 鏡像概述

本目錄包含了 Fish Server Admin 服務的 Docker 化部署文件，能夠成功構建出一個優化的、安全的 Docker 鏡像。

## 🏗️ 鏡像特性

- **多階段構建**: 使用 Go 1.24 Alpine 進行構建，最終運行鏡像基於 Alpine 3.18
- **安全性**: 非 root 用戶運行，包含安全標頭
- **小體積**: 最終鏡像大小約 60MB
- **健康檢查**: 內建健康檢查機制
- **時區設置**: 預設為 Asia/Taipei

## 📁 文件說明

```
deployments/
├── Dockerfile.admin           # Admin 服務的 Dockerfile
├── .dockerignore             # Docker 構建忽略文件
├── build-admin.sh            # Linux/Mac 構建腳本
├── build-admin.ps1           # Windows PowerShell 構建腳本
├── docker-compose.test.yml   # 完整測試環境
├── config-docker.yaml        # Docker 環境配置
└── README.md                # 本說明文件
```

## 🚀 快速開始

### 方法一：使用構建腳本（推薦）

#### Linux/Mac:
```bash
# 賦予執行權限
chmod +x deployments/build-admin.sh

# 構建並測試鏡像
./deployments/build-admin.sh

# 只構建鏡像
./deployments/build-admin.sh build

# 只測試鏡像
./deployments/build-admin.sh test
```

#### Windows PowerShell:
```powershell
# 構建並測試鏡像
.\deployments\build-admin.ps1

# 只構建鏡像
.\deployments\build-admin.ps1 -Command build

# 只測試鏡像
.\deployments\build-admin.ps1 -Command test
```

### 方法二：手動構建

```bash
# 1. 構建鏡像
docker build -f deployments/Dockerfile.admin -t fish-server-admin:latest .

# 2. 運行測試容器
docker run --name fish-admin-test -p 6060:6060 -d fish-server-admin:latest

# 3. 檢查健康狀態
curl http://localhost:6060/ping
```

## 🐳 完整環境部署

使用 Docker Compose 啟動完整的測試環境（包括數據庫）：

```bash
# 啟動完整環境
docker-compose -f deployments/docker-compose.test.yml up -d

# 查看日誌
docker-compose -f deployments/docker-compose.test.yml logs -f admin

# 停止環境
docker-compose -f deployments/docker-compose.test.yml down
```

## 🌐 服務端點

Admin 服務啟動後，可以訪問以下端點：

### 基本端點
- **根頁面**: http://localhost:6060/
- **健康檢查**: http://localhost:6060/ping
- **API 健康檢查**: http://localhost:6060/admin/health

### 管理端點
- **服務器狀態**: http://localhost:6060/admin/status
- **系統指標**: http://localhost:6060/admin/metrics

### Pprof 性能分析端點
- **Pprof 首頁**: http://localhost:6060/debug/pprof/
- **CPU 分析**: http://localhost:6060/debug/pprof/profile
- **內存分析**: http://localhost:6060/debug/pprof/heap
- **Goroutine 分析**: http://localhost:6060/debug/pprof/goroutine
- **使用說明**: http://localhost:6060/debug/pprof/info

## 🔧 配置說明

### 環境變量
- `GIN_MODE`: Gin 運行模式 (release/debug)
- `LOG_LEVEL`: 日誌級別 (debug/info/warn/error)
- `CONFIG_PATH`: 配置文件路徑

### 配置文件
- 默認配置: `/app/configs/config.yaml`
- Docker 環境配置: `deployments/config-docker.yaml`

## 🛠️ 開發和調試

### 查看容器日誌
```bash
docker logs fish-admin-test -f
```

### 進入容器調試
```bash
docker exec -it fish-admin-test /bin/sh
```

### 性能分析
```bash
# CPU 分析
go tool pprof http://localhost:6060/debug/pprof/profile

# 內存分析
go tool pprof http://localhost:6060/debug/pprof/heap
```

## 📊 鏡像信息

```bash
# 查看鏡像大小
docker images fish-server-admin

# 查看鏡像構建歷史
docker history fish-server-admin:latest

# 查看鏡像詳細信息
docker inspect fish-server-admin:latest
```

## 🚧 故障排除

### 常見問題

1. **構建失敗 - Go 版本不匹配**
   ```
   錯誤: go.mod requires go >= 1.24.9
   解決: 確保 Dockerfile 中使用正確的 Go 版本
   ```

2. **容器無法啟動 - 數據庫連接失敗**
   ```
   錯誤: dial tcp 127.0.0.1:5432: connect: connection refused
   解決: 使用 docker-compose 或確保數據庫服務可訪問
   ```

3. **健康檢查失敗**
   ```bash
   # 檢查容器狀態
   docker ps -a
   
   # 查看詳細日誌
   docker logs container_name
   ```

### 清理命令

```bash
# 停止並移除測試容器
docker rm -f fish-admin-test

# 移除鏡像
docker rmi fish-server-admin:latest

# 清理 Docker Compose 環境
docker-compose -f deployments/docker-compose.test.yml down -v
```

## 🎯 生產部署建議

1. **安全性**
   - 修改默認的 JWT secret
   - 為 pprof 端點添加認證
   - 使用 HTTPS

2. **性能**
   - 根據需求調整資源限制
   - 配置適當的健康檢查間隔
   - 監控內存和 CPU 使用

3. **監控**
   - 集成日誌收集系統
   - 設置性能監控
   - 配置告警機制

## 📝 版本信息

- **鏡像版本**: 1.0.0
- **Go 版本**: 1.24
- **基礎鏡像**: Alpine 3.18
- **默認端口**: 6060

---

**注意**: 此 Docker 鏡像已經成功測試，可以正常構建和運行 Admin 服務。