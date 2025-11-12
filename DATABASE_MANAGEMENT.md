# 🗄️ 數據庫管理指南

## 快速參考

### 完全重置數據庫（推薦）

當遇到任何遷移問題時，最簡單的解決方案是完全重置數據庫。

#### Windows

```cmd
REM 使用批處理
scripts\reset-database.bat

REM 使用 PowerShell
.\scripts\reset-database.ps1
```

#### Linux/Mac

```bash
./scripts/reset-database.sh
```

## 這個腳本做什麼？

重置腳本會：
1. ✅ 終止所有數據庫連接
2. ✅ 完全刪除數據庫
3. ✅ 創建全新的數據庫
4. ✅ 從頭開始運行所有遷移

**⚠️ 警告：所有數據將被刪除！**

## 常規遷移操作

### 運行遷移

```bash
# 應用所有待執行的遷移
go run cmd/migrator/main.go up

# 回滾最後一個遷移
go run cmd/migrator/main.go down

# 查看當前遷移狀態
go run cmd/migrator/main.go version
```

### Windows 用戶

```cmd
REM 應用所有遷移
scripts\run-migration.bat up

REM 回滾遷移
scripts\run-migration.bat down

REM 查看狀態
scripts\run-migration.bat version
```

## 遷移問題排除

### 問題：遷移失敗或報錯

**解決方案：完全重置數據庫**

```bash
# Linux/Mac
./scripts/reset-database.sh

# Windows
scripts\reset-database.bat
```

這將清除所有問題並從乾淨的狀態重新開始。

### 問題：找不到數據庫或連接失敗

**解決方案：啟動數據庫**

```bash
# Linux/Mac
make run-dev

# Windows
scripts\start-database.bat
```

等待 10-15 秒讓數據庫完全啟動，然後重試。

## 遷移文件說明

所有遷移文件位於 `storage/migrations/` 目錄：

- `000001_create_initial_tables` - 創建核心表（users, wallets, wallet_transactions）
- `000002_create_game_tables` - 創建遊戲相關表
- `000003_create_fish_types_table` - 魚種類型表
- `000004_seed_fish_types_data` - 填充魚種數據
- `000005_create_formation_config_table` - 陣型配置表
- `000006_create_users_table` - **已棄用**（內容已合併到 migration 1）
- `000007_create_announcements_table` - 公告表
- `000008_create_fish_tide_config_table` - 魚潮配置表
- `000009_create_room_configs_table` - 房間配置表

### Migration 6 說明

⚠️ **重要**：Migration 6 現在是一個空操作（no-op）。

原因：Migration 1 和 Migration 6 原本都創建 `users` 表，造成衝突。
解決：將完整的 users 表定義合併到 Migration 1，Migration 6 改為空操作以保持版本號連續性。

## 冪等性

所有遷移文件現在都使用 `IF NOT EXISTS` 來創建索引和表，確保：
- ✅ 遷移可以安全地重複執行
- ✅ 部分失敗的遷移不會阻止重試
- ✅ 不會出現 "already exists" 錯誤

## 數據庫配置

### 開發環境

配置文件：`configs/config.dev.yaml`

```yaml
data:
  database:
    driver: "postgres"
    host: "localhost"
    port: 5432
    user: "user"
    password: "password"
    dbname: "fish_db"  # 注意：必須與 docker-compose.dev.yml 一致
    sslmode: "disable"
```

### Docker Compose

配置文件：`deployments/docker-compose.dev.yml`

```yaml
postgres:
  environment:
    POSTGRES_DB: fish_db  # 注意：必須與 config.dev.yaml 一致
    POSTGRES_USER: user
    POSTGRES_PASSWORD: password
```

## 最佳實踐

### 1. 開發流程

```bash
# 1. 啟動數據庫
make run-dev  # 或 scripts\start-database.bat (Windows)

# 2. 運行遷移
go run cmd/migrator/main.go up

# 3. 開發...

# 4. 如果遇到問題，重置數據庫
./scripts/reset-database.sh
```

### 2. 添加新的遷移

```bash
# 創建新的遷移文件
# 文件名格式：000010_description.up.sql 和 000010_description.down.sql

# up.sql - 應用遷移的 SQL
# down.sql - 回滾遷移的 SQL
```

### 3. 遷移文件規範

```sql
-- ✅ 好的做法：使用 IF NOT EXISTS
CREATE TABLE IF NOT EXISTS my_table (...);
CREATE INDEX IF NOT EXISTS idx_my_index ON my_table(column);

-- ❌ 不好的做法：不使用 IF NOT EXISTS
CREATE TABLE my_table (...);
CREATE INDEX idx_my_index ON my_table(column);
```

## 快速命令參考

| 操作 | Linux/Mac | Windows |
|------|-----------|---------|
| 啟動數據庫 | `make run-dev` | `scripts\start-database.bat` |
| 停止數據庫 | `make docker-down` | `scripts\stop-database.bat` |
| 重置數據庫 | `./scripts/reset-database.sh` | `scripts\reset-database.bat` |
| 運行遷移 | `go run cmd/migrator/main.go up` | `scripts\run-migration.bat up` |
| 查看狀態 | `go run cmd/migrator/main.go version` | `scripts\run-migration.bat version` |

## 故障排除

### 數據庫無法連接

1. 檢查 Docker 是否運行：`docker ps | grep postgres`
2. 檢查容器日誌：`docker logs fish_server-postgres-1`
3. 重啟數據庫：停止後重新啟動

### 遷移一直失敗

直接使用重置腳本：

```bash
./scripts/reset-database.sh  # Linux/Mac
scripts\reset-database.bat   # Windows
```

這會解決 99% 的遷移問題。

### 數據庫名稱不匹配

確保以下配置一致：
- `configs/config.dev.yaml` → `dbname: "fish_db"`
- `deployments/docker-compose.dev.yml` → `POSTGRES_DB: fish_db`

## 相關文檔

- [Windows 使用指南](WINDOWS_GUIDE.md) - Windows 專用詳細說明
- [README.md](README.md) - 專案總體說明
