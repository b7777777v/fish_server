# 🚀 VS Code 多環境快速參考

## ⌨️ 快捷鍵速查

### 🎯 環境啟動
| 快捷鍵 | 功能 | 說明 |
|--------|------|------|
| `Ctrl+F1` | 🟢 Admin DEV | 啟動 Admin 開發版 (Pprof ON) |
| `Ctrl+F2` | 🟡 Admin STAGING | 啟動 Admin 預發布版 (Pprof OFF) |
| `Ctrl+F3` | 🔴 Admin PROD | 啟動 Admin 生產版 (Secure) |
| `Ctrl+Shift+F1` | 🚀 DEV 全服務 | 啟動所有 DEV 服務 |
| `Ctrl+Shift+F2` | 🏗️ STAGING 全服務 | 啟動所有 STAGING 服務 |

### 🔨 構建與測試
| 快捷鍵 | 功能 | 說明 |
|--------|------|------|
| `Ctrl+Shift+B` | 🔨 構建 Admin DEV | 僅構建 Admin 開發版本 |
| `Ctrl+Shift+T` | 🧪 運行所有測試 | 執行專案所有 `_test.go` |
| `Ctrl+Shift+D` | 🐳 Docker 構建 DEV | 構建 Admin 開發版 Docker 鏡像 |

### 🛠️ 代碼生成
| 快捷鍵 | 功能 | 說明 |
|--------|------|------|
| `Ctrl+Alt+W` | Wire 生成 | 生成 `wire_gen.go` 依賴注入 |
| `Ctrl+Alt+G` | Proto 生成 | 生成 `*.pb.go` gRPC/Protobuf |

### 🧹 環境管理與清理
| 快捷鍵 | 功能 | 說明 |
|--------|------|------|
| `Ctrl+Alt+1` | 🚀 啟動 DEV Docker | 啟動 `docker-compose.dev.yml` |
| `Ctrl+Alt+2` | 🏗️ 啟動 STAGING Docker | 啟動 `docker-compose.staging.yml` |
| `Ctrl+Alt+0` | 🛑 停止全部 Docker | 停止所有 docker-compose 環境 |
| `Ctrl+Alt+C` | 🧹 清理構建產物 | 刪除 `cmd/.../admin-*` 等文件 |
| `Ctrl+Alt+Shift+C` | 🧹 清理 Docker 鏡像 | 刪除 `fish-server-admin` 鏡像 |

## 🌍 環境對比

| 項目 | 🟢 DEV | 🟡 STAGING | 🔴 PROD |
|------|---------|------------|---------|
| **Pprof** | ✅ 啟用 | ❌ 關閉 | ❌ 關閉 |
| **端口** | 6060 | 6061 | 6062 |
| **日誌級別** | debug | info | warn |
| **Gin 模式** | debug | release | release |
| **優化** | 無 | `-s -w` | `-s -w` |
| **安全** | 低 | 中 | 高 |

## 🎮 常用工作流程

### 📝 日常開發 (DEV)
1. `Ctrl+F1` - 啟動 Admin DEV 環境。
2. 在代碼中設置斷點。
3. 按 `F5`，選擇 `🔍 Debug Admin with Delve` 開始偵錯。
4. `Ctrl+Alt+W` - 如有需要，更新依賴注入。
5. `Ctrl+Shift+T` - 運行單元測試。

### 🧪 預發布測試 (STAGING)
1. `Ctrl+F2` - 啟動 Admin STAGING 環境。
2. 測試生產級別功能。
3. `Ctrl+Alt+S` - 驗證 Pprof 已被關閉。

### 🚀 生產驗證 (PROD)
1. `Ctrl+F3` - 啟動 Admin PROD 環境。
2. 驗證所有安全設置與最終性能。

### 🐳 Docker 工作流程
1. `Ctrl+Shift+D` - 構建 DEV Docker 鏡像。
2. `Ctrl+Alt+1` - 啟動 DEV Docker 容器化環境。
3. `Ctrl+Alt+0` - 測試完畢後停止所有 Docker 服務。

### DEV 環境 (localhost:6060)
```
🏠 根頁面:          http://localhost:6060/
💚 健康檢查:        http://localhost:6060/ping
📊 系統狀態:        http://localhost:6060/admin/status
🌍 環境信息:        http://localhost:6060/admin/env
📈 系統指標:        http://localhost:6060/admin/metrics

🔍 Pprof 首頁:      http://localhost:6060/debug/pprof/
🧠 CPU 分析:        http://localhost:6060/debug/pprof/profile
💾 記憶體分析:      http://localhost:6060/debug/pprof/heap
🔄 Goroutine:      http://localhost:6060/debug/pprof/goroutine
ℹ️ Pprof 說明:      http://localhost:6060/debug/pprof/info
```

### STAGING 環境 (localhost:6061)
```
🏠 根頁面:          http://localhost:6061/
💛 健康檢查:        http://localhost:6061/ping
📊 系統狀態:        http://localhost:6061/admin/status
🌍 環境信息:        http://localhost:6061/admin/env

🚫 Pprof 說明:      http://localhost:6061/debug/pprof/disabled
```

## 🧪 快速驗證命令

### 開發環境 Pprof 檢查
```bash
# 應該返回 pprof 首頁 HTML
curl http://localhost:6060/debug/pprof/

# 檢查環境配置 (pprof_enabled: true)
curl http://localhost:6060/admin/env | jq '.features.pprof_enabled'
```

### 預發布環境 Pprof 檢查
```bash
# 應該返回說明信息
curl http://localhost:6061/debug/pprof/disabled

# 檢查環境配置 (pprof_enabled: false)  
curl http://localhost:6061/admin/env | jq '.features.pprof_enabled'
```

## 🎮 常用工作流程

### 📝 日常開發
1. `Ctrl+F1` - 啟動 DEV 環境
2. 設置斷點，開始調試
3. 訪問 http://localhost:6060/debug/pprof/ 進行性能分析
4. `Ctrl+Shift+T` - 運行測試

### 🧪 預發布測試
1. `Ctrl+F2` - 啟動 STAGING 環境
2. `Ctrl+Alt+S` - 驗證 Pprof 已關閉
3. 測試生產級別功能
4. 檢查資源使用是否降低

### 🚀 生產驗證
1. `Ctrl+F3` - 啟動 PROD 環境
2. `Ctrl+Alt+V` - 檢查環境信息
3. 驗證所有安全設置
4. 確認性能最佳化

### 🐳 Docker 工作流程
1. `Ctrl+Shift+D` - 構建 Docker 鏡像
2. `Ctrl+Alt+1` - 啟動 DEV Docker 環境
3. `Ctrl+Alt+2` - 啟動 STAGING Docker 環境
4. `Ctrl+Alt+V` - 驗證環境狀態
5. `Ctrl+Alt+0` - 清理所有環境

## 📁 文件結構速查

```
.vscode/
├── launch.json           # 🎯 啟動配置 (F5)
├── tasks.json            # 🔨 任務配置 (Ctrl+Shift+P)
├── settings.json         # ⚙️ 工作區設定
├── keybindings.json      # ⌨️ 快捷鍵配置
├── README.md             # 📚 詳細說明
└── QUICK-REFERENCE.md    # 📋 本快速參考

configs/
├── config.dev.yaml       # 🟢 DEV 配置 (Pprof ON)
├── config.staging.yaml   # 🟡 STAGING 配置 (Pprof OFF)
└── config.prod.yaml      # 🔴 PROD 配置 (最安全)

deployments/
├── docker-compose.dev.yml     # 🐳 DEV Docker 環境
├── docker-compose.staging.yml # 🐳 STAGING Docker 環境
├── run-environment.sh         # 🐧 Linux/Mac 管理腳本
└── run-environment.ps1        # 🪟 Windows 管理腳本
```

## 🆘 緊急故障排除

### Pprof 無法訪問
1. 確認環境: `curl http://localhost:6060/admin/env`
2. 重啟服務: `Ctrl+F1`
3. 檢查日誌: VS Code 終端輸出

### 環境變量未生效
1. 重啟 VS Code
2. 檢查 launch.json 中的 env 配置
3. 使用 `⚡ Auto Environment` 動態選擇

### 端口衝突
1. 檢查端口使用: `netstat -tulpn | grep :6060`
2. 修改配置文件中的端口
3. 或停止其他服務

### Wire 生成失敗
1. `Ctrl+Alt+W` - 重新生成
2. 或手動執行: `go generate ./...`
3. 檢查 wire.go 文件語法

## 💡 專業提示

1. **使用工作區文件**: 打開 `fish_server.code-workspace` 獲得最佳體驗
2. **環境標識**: 狀態欄顏色會根據環境變化
3. **自動完成**: 輸入配置時 VS Code 會提供智能提示
4. **多終端**: 可同時開啟多個環境的終端
5. **快速切換**: 使用 `⚡ Auto Environment` 快速測試不同配置

---

**記住**: DEV 環境啟用 Pprof 用於開發調試，STAGING 和 PROD 環境關閉 Pprof 以降低資源使用和提高安全性！