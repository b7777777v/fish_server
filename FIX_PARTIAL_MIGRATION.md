# 🔧 修復部分應用的 Migration 6

## 問題描述

當你看到這個錯誤時：
```
relation "idx_users_username" already exists
```

這表示 migration 6 (create_users_table) 部分執行了 - 創建了一些索引但沒有完成整個遷移。數據庫現在處於不一致的狀態。

## 🚀 快速修復步驟

### Windows 用戶

#### 步驟 1: 啟動數據庫

```cmd
REM 使用批處理腳本
scripts\start-database.bat

REM 或使用 Docker Compose
docker-compose -f deployments\docker-compose.dev.yml up -d postgres redis
```

等待 5-10 秒讓數據庫完全啟動。

#### 步驟 2: 清理部分應用的更改

```cmd
REM 使用批處理腳本
scripts\cleanup-migration-6.bat

REM 或使用 PowerShell
.\scripts\cleanup-migration-6.ps1
```

當提示確認時，輸入 `yes`。

#### 步驟 3: 強制版本到 5

```cmd
go run cmd\migrator\main.go force 5
```

#### 步驟 4: 重新應用所有遷移

```cmd
go run cmd\migrator\main.go up
```

#### 步驟 5: 驗證成功

```cmd
go run cmd\migrator\main.go version
```

應該顯示版本 9（或最新版本）且 dirty: false

### Linux/Mac 用戶

#### 步驟 1: 啟動數據庫

```bash
make run-dev

# 或使用 Docker Compose
docker-compose -f deployments/docker-compose.yml up -d postgres redis
```

#### 步驟 2: 清理部分應用的更改

```bash
./scripts/cleanup-migration-6.sh
```

當提示確認時，輸入 `yes`。

#### 步驟 3: 強制版本到 5

```bash
go run cmd/migrator/main.go force 5
```

#### 步驟 4: 重新應用所有遷移

```bash
go run cmd/migrator/main.go up
```

#### 步驟 5: 驗證成功

```bash
go run cmd/migrator/main.go version
```

## 🔍 手動修復（如果腳本無法運行）

如果自動腳本無法運行，可以手動執行以下步驟：

### 步驟 1: 連接到數據庫

**Windows:**
```cmd
docker exec -it fish_server-postgres-1 psql -U user -d fish_db
```

**Linux/Mac:**
```bash
PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db
```

### 步驟 2: 在 psql 中執行清理命令

```sql
-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_users_updated_at();

-- Drop all constraints
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_third_party;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS check_regular_user;

-- Drop all indexes (explicitly)
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_third_party;
DROP INDEX IF EXISTS idx_users_is_guest;
DROP INDEX IF EXISTS idx_users_created_at;

-- Drop the table
DROP TABLE IF EXISTS users;

-- Exit psql
\q
```

### 步驟 3: 強制版本並重新遷移

```bash
# 強制到版本 5
go run cmd/migrator/main.go force 5

# 重新應用遷移
go run cmd/migrator/main.go up

# 驗證
go run cmd/migrator/main.go version
```

## 🔍 檢查數據庫狀態

如果想在清理前檢查數據庫當前狀態：

```sql
-- 連接到數據庫
-- Windows: docker exec -it fish_server-postgres-1 psql -U user -d fish_db
-- Linux/Mac: PGPASSWORD=password psql -h localhost -p 5432 -U user -d fish_db

-- 檢查 users 表是否存在
\dt users

-- 檢查索引
\di idx_users*

-- 檢查 schema_migrations 表
SELECT * FROM schema_migrations;

-- 查看 users 表結構（如果存在）
\d users
```

## ⚠️ 常見問題

### Q: 為什麼會出現這個問題？

**A:** Migration 6 包含多個步驟（創建表、索引、約束、觸發器）。如果在執行過程中出現錯誤或中斷（如網絡問題、權限問題），只有部分步驟會成功執行，導致數據庫處於"dirty"狀態。

### Q: 清理腳本會刪除我的數據嗎？

**A:** 清理腳本只刪除 users 表及其相關對象。如果這是新設置的數據庫，不會影響其他數據。如果你有重要數據，請先備份：

```bash
# 備份整個數據庫
docker exec fish_server-postgres-1 pg_dump -U user fish_db > backup.sql

# 或只備份 users 表
docker exec fish_server-postgres-1 pg_dump -U user -t users fish_db > users_backup.sql
```

### Q: 腳本執行失敗怎麼辦？

**A:**
1. 確認數據庫正在運行：`docker ps | grep postgres`
2. 確認數據庫名稱正確：應該是 `fish_db`
3. 手動執行清理 SQL（見上面的手動修復部分）
4. 檢查 Docker 容器日誌：`docker logs fish_server-postgres-1`

### Q: 如何避免未來出現這種問題？

**A:**
1. 在應用遷移前備份數據庫
2. 確保數據庫連接穩定
3. 檢查遷移文件語法是否正確
4. 使用事務性遷移（golang-migrate 默認支持）
5. 在開發環境測試遷移後再應用到生產環境

## 📝 完整的一鍵修復命令

### Windows (PowerShell)

```powershell
# 一鍵執行所有步驟
.\scripts\start-database.ps1
Start-Sleep -Seconds 10
.\scripts\cleanup-migration-6.ps1
go run cmd\migrator\main.go force 5
go run cmd\migrator\main.go up
go run cmd\migrator\main.go version
```

### Windows (CMD)

```cmd
REM 執行每個步驟，確認成功後再執行下一步
scripts\start-database.bat
timeout /t 10 /nobreak
scripts\cleanup-migration-6.bat
go run cmd\migrator\main.go force 5
go run cmd\migrator\main.go up
go run cmd\migrator\main.go version
```

### Linux/Mac

```bash
# 一鍵執行所有步驟
make run-dev && sleep 10 && \
./scripts/cleanup-migration-6.sh && \
go run cmd/migrator/main.go force 5 && \
go run cmd/migrator/main.go up && \
go run cmd/migrator/main.go version
```

## 📚 相關文檔

- [MIGRATION_FIX_GUIDE.md](MIGRATION_FIX_GUIDE.md) - 通用的 dirty migration 修復指南
- [WINDOWS_GUIDE.md](WINDOWS_GUIDE.md) - Windows 完整使用指南
- [README.md](README.md) - 專案總體說明

## ✅ 成功標誌

修復成功後，你應該看到：

```
Current migration version: 9, dirty: false
```

或類似的輸出，其中 `dirty: false` 表示遷移狀態正常。
