# 資料庫讀寫分離配置指南

## 概述

Fish Server 支持資料庫讀寫分離架構，可以將讀操作和寫操作分別路由到不同的資料庫實例，以提升系統效能和可擴展性。

## 架構設計

### DBManager 架構

```
┌─────────────────────────────────────┐
│         DBManager                    │
├─────────────────────────────────────┤
│  - writeDB: *Client (主庫)          │
│  - readDB:  *Client (從庫)          │
├─────────────────────────────────────┤
│  + Write() *Client                   │
│  + Read()  *Client                   │
└─────────────────────────────────────┘
           │              │
           ▼              ▼
    ┌──────────┐   ┌──────────┐
    │ 主庫      │   │ 從庫      │
    │ (寫操作)  │   │ (讀操作)  │
    └──────────┘   └──────────┘
```

### 操作分類

| 操作類型 | 使用連接 | 範例 |
|---------|---------|------|
| SELECT 查詢 | Read DB (從庫) | `SELECT * FROM users WHERE id = $1` |
| INSERT 插入 | Write DB (主庫) | `INSERT INTO users (name) VALUES ($1)` |
| UPDATE 更新 | Write DB (主庫) | `UPDATE users SET name = $1 WHERE id = $2` |
| DELETE 刪除 | Write DB (主庫) | `DELETE FROM users WHERE id = $1` |
| 事務操作 | Write DB (主庫) | `BEGIN; ... COMMIT;` |

## 配置方式

### 方式一：主從共用（預設）

**適用場景**：
- 單機部署
- 暫時不需要讀寫分離
- 開發和測試環境

**配置範例**：
```yaml
data:
  master_database:
    host: "localhost"
    port: 5432
    user: "user"
    password: "password"
    dbname: "fish_db"
    sslmode: "disable"
```

**行為**：
- 讀寫操作都使用同一個資料庫連接池
- 系統會創建兩個獨立的連接池實例，但都指向同一個資料庫

### 方式二：讀寫分離（推薦用於生產環境）

**適用場景**：
- 生產環境
- 讀操作遠多於寫操作
- 需要水平擴展讀取能力

**配置範例**：
```yaml
data:
  # 主庫配置（用於寫操作）
  master_database:
    host: "master.db.example.com"
    port: 5432
    user: "fish_user"
    password: "secure-password"
    dbname: "fish_db"
    sslmode: "require"
    max_open_conns: 100
    max_idle_conns: 25

  # 從庫配置（用於讀操作）
  slave_database:
    host: "slave.db.example.com"
    port: 5432
    user: "fish_user"
    password: "secure-password"
    dbname: "fish_db"
    sslmode: "require"
    max_open_conns: 200  # 讀庫可配置更多連接
    max_idle_conns: 50
```

**行為**：
- 讀操作（SELECT）使用 `slave_database` 配置的從庫
- 寫操作（INSERT/UPDATE/DELETE/事務）使用 `master_database` 配置的主庫

## 程式碼範例

### Repository 層使用

```go
// 讀操作使用 Read DB
func (r *playerRepo) GetPlayer(ctx context.Context, playerID int64) (*Player, error) {
    query := `SELECT id, name, balance FROM players WHERE id = $1`
    // 使用讀庫
    err := r.data.DBManager().Read().QueryRow(ctx, query, playerID).Scan(...)
    return player, err
}

// 寫操作使用 Write DB
func (r *playerRepo) UpdateBalance(ctx context.Context, playerID int64, balance int64) error {
    query := `UPDATE players SET balance = $1 WHERE id = $2`
    // 使用寫庫
    _, err := r.data.DBManager().Write().Exec(ctx, query, balance, playerID)
    return err
}

// 事務操作使用 Write DB
func (r *walletRepo) Transfer(ctx context.Context, from, to int64, amount float64) error {
    // 事務必須使用寫庫
    tx, err := r.data.DBManager().Write().Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // 執行轉帳邏輯
    // ...

    return tx.Commit()
}
```

## 連接池配置建議

### 主庫（Write DB）

```yaml
master_database:
  max_open_conns: 100      # 最大連接數
  max_idle_conns: 25       # 最大空閒連接數
  conn_max_lifetime: "1h"  # 連接最大生命週期
```

**建議值**：
- **開發環境**：max_open_conns: 25, max_idle_conns: 10
- **測試環境**：max_open_conns: 50, max_idle_conns: 15
- **生產環境**：max_open_conns: 100-200, max_idle_conns: 25-50

### 從庫（Read DB）

```yaml
slave_database:
  max_open_conns: 200      # 讀庫可以設置更多連接
  max_idle_conns: 50       # 更多空閒連接
  conn_max_lifetime: "1h"
```

**建議值**：
- 從庫通常處理更多的讀請求，可以配置 1.5-2 倍的主庫連接數
- 根據實際流量調整，使用監控工具追蹤連接使用率

## 主從複製配置

### PostgreSQL 主從複製設置

1. **主庫配置** (`postgresql.conf`)：
```
wal_level = replica
max_wal_senders = 10
wal_keep_size = 1GB
```

2. **從庫配置**：
```bash
# 停止從庫
pg_ctl stop

# 從主庫複製資料
pg_basebackup -h master.db.example.com -D /var/lib/postgresql/data -U replicator -v -P

# 創建 standby.signal 文件
touch /var/lib/postgresql/data/standby.signal

# 配置 postgresql.auto.conf
primary_conninfo = 'host=master.db.example.com port=5432 user=replicator password=xxx'

# 啟動從庫
pg_ctl start
```

3. **驗證複製狀態**：
```sql
-- 在主庫執行
SELECT * FROM pg_stat_replication;

-- 在從庫執行
SELECT * FROM pg_stat_wal_receiver;
```

## 負載均衡

### 使用 PgBouncer

對於多個從庫，可以使用 PgBouncer 作為連接池和負載均衡器：

```ini
[databases]
fish_db = host=slave1.db.example.com,slave2.db.example.com port=5432 dbname=fish_db

[pgbouncer]
pool_mode = transaction
max_client_conn = 1000
default_pool_size = 25
```

配置中使用 PgBouncer 地址：
```yaml
slave_database:
  host: "pgbouncer.example.com"
  port: 6432
```

## 注意事項

### 1. 複製延遲

- **問題**：從庫可能存在複製延遲（通常 < 1 秒）
- **影響**：讀取可能獲得稍舊的資料
- **解決方案**：
  - 對於需要即時一致性的場景，使用主庫讀取
  - 監控複製延遲，設置告警閾值
  - 考慮使用同步複製（性能會下降）

### 2. 一致性考慮

**適合使用從庫的場景**：
- 查詢歷史資料
- 統計分析
- 列表查詢
- 非實時性要求的資料

**需要使用主庫的場景**：
- 寫入後立即讀取
- 需要讀取-修改-寫入（Read-Modify-Write）的操作
- 事務內的讀取

### 3. 監控指標

需要監控的關鍵指標：
- 主從複製延遲
- 連接池使用率
- 查詢執行時間
- 資料庫負載（CPU、記憶體、I/O）

## 測試驗證

### 1. 功能測試

```bash
# 編譯專案
go build ./...

# 執行測試
go test ./internal/data/... -v
```

### 2. 連接驗證

啟動服務後，查看日誌確認連接狀態：
```
INFO  Creating separate read database connection: slave.db.example.com:5432/fish_db
INFO  Successfully connected to write database: master.db.example.com:5432
INFO  Successfully connected to read database: slave.db.example.com:5432
```

### 3. 運行時驗證

```bash
# 查看 PostgreSQL 連接
SELECT datname, usename, application_name, client_addr, state
FROM pg_stat_activity
WHERE datname = 'fish_db';
```

## 故障處理

### 從庫故障

如果從庫故障，可以臨時切換到主庫讀取：

1. 修改配置文件，註釋掉 `slave_database` 配置
2. 重啟服務
3. 系統會自動使用主庫處理所有讀寫操作

### 主庫故障

1. 提升一個從庫為新的主庫
2. 更新配置文件中的 `master_database` 配置
3. 重啟服務

## 效能優化建議

1. **合理配置連接池**：根據實際負載調整連接數
2. **使用連接池復用**：避免頻繁創建/銷毀連接
3. **監控慢查詢**：優化查詢性能
4. **合理使用索引**：提升查詢速度
5. **定期維護**：執行 VACUUM 和分析統計信息

## 參考資源

- [PostgreSQL 複製文檔](https://www.postgresql.org/docs/current/runtime-config-replication.html)
- [pgx 連接池配置](https://github.com/jackc/pgx)
- [PgBouncer 文檔](https://www.pgbouncer.org/)
