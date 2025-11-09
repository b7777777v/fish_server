# 陣型配置啟動載入指南

## 架構說明

陣型配置系統採用三層架構，確保配置的持久化和高性能訪問：

```
[PostgreSQL]  <-- 主存儲（Source of Truth）
     ↓
[Redis Cache] <-- 快取層（1小時過期）
     ↓
[Memory]      <-- Spawner 運行時配置
```

## 數據流向

### 1. 服務啟動
```
DB → Redis → Memory (Spawner)
```

### 2. 配置讀取
```
優先 Redis → 未命中則 DB → 回寫 Redis
```

### 3. 配置更新（熱更新）
```
Admin API → DB → Redis → Memory (Spawner)
```

## 數據庫遷移

首先運行數據庫遷移創建 `formation_configs` 表：

```bash
# 運行遷移
go run cmd/migrator/main.go up

# 檢查遷移狀態
go run cmd/migrator/main.go version
```

遷移會自動創建：
- `formation_configs` 表（JSONB 格式存儲配置）
- 索引（config_key, is_active, GIN 索引）
- 自動更新時間觸發器
- 默認配置數據

## 啟動時載入配置

### 方法 1: 在應用啟動時自動載入（推薦）

修改 `internal/app/game/app.go` 或服務啟動代碼：

```go
// NewGameApp 創建遊戲應用程序
func NewGameApp(
	gameUsecase *game.GameUsecase,
	formationConfigSvc *game.FormationConfigService, // 注入配置服務
	config *conf.Config,
	logger logger.Logger,
	hub *Hub,
	wsHandler *WebSocketHandler,
	messageHandler *MessageHandler,
) *GameApp {
	ctx, cancel := context.WithCancel(context.Background())

	app := &GameApp{
		hub:            hub,
		wsHandler:      wsHandler,
		messageHandler: messageHandler,
		gameUsecase:    gameUsecase,
		config:         config,
		logger:         logger.With("component", "game_app"),
		ctx:            ctx,
		cancel:         cancel,
	}

	// 設置 HTTP 服務器
	app.setupHTTPServer()

	// 啟動時從 DB 載入配置到 Redis
	go func() {
		if err := formationConfigSvc.LoadConfigFromDB(context.Background()); err != nil {
			app.logger.Errorf("Failed to load formation config from DB: %v", err)
		} else {
			app.logger.Info("Formation config loaded from DB to Redis successfully")

			// 應用配置到 Spawner
			config, _ := formationConfigSvc.LoadConfig(context.Background())
			if config != nil {
				gameUsecase.UpdateFormationConfig(*config)
				app.logger.Info("Formation config applied to Spawner")
			}
		}

		// 異步加載魚類數據到緩存
		if err := app.gameUsecase.LoadAndCacheFishTypes(context.Background()); err != nil {
			app.logger.Errorf("Failed to load and cache fish types: %v", err)
		}
	}()

	return app
}
```

### 方法 2: 在 main.go 中手動調用

```go
func main() {
	// ... 初始化代碼 ...

	// 創建服務
	app, cleanup, err := initApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	// 啟動時載入陣型配置
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := formationConfigService.LoadConfigFromDB(ctx); err != nil {
		log.Printf("Warning: Failed to load formation config: %v", err)
	}

	// 運行應用程序
	if err := app.Run(); err != nil {
		log.Fatalf("App failed to run: %v", err)
	}
}
```

## 依賴注入配置（Wire）

如果使用 Wire 進行依賴注入，需要添加以下提供者：

```go
// cmd/game/wire.go 或 cmd/admin/wire.go

//+build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/data"
	// ... 其他導入
)

// 添加 FormationConfigRepo 和 FormationConfigService 提供者
var dataSet = wire.NewSet(
	// ... 其他 data providers ...
	data.NewFormationConfigRepo,  // 新增
)

var bizSet = wire.NewSet(
	// ... 其他 biz providers ...
	game.NewFormationConfigService, // 新增
)
```

然後運行：

```bash
# 生成依賴注入代碼
cd cmd/game  # 或 cmd/admin
wire
```

## 配置查看和驗證

### 查看資料庫中的配置

```sql
-- 查看當前配置
SELECT
    id,
    config_key,
    config_data->>'enabled' as enabled,
    config_data->>'base_spawn_chance' as base_spawn_chance,
    is_active,
    created_at,
    updated_at
FROM formation_configs
WHERE is_active = true;

-- 查看配置的詳細 JSON
SELECT
    config_key,
    jsonb_pretty(config_data) as config
FROM formation_configs
WHERE config_key = 'default';
```

### 查看 Redis 中的快取

```bash
# 連接到 Redis
redis-cli

# 查看配置快取
GET game:formation:config:default

# 查看 TTL
TTL game:formation:config:default

# 手動清除快取（強制從 DB 重新載入）
DEL game:formation:config:default
```

### 通過 API 查看當前配置

```bash
# 獲取當前運行時配置
curl http://localhost:8081/admin/formations/config

# 獲取統計信息
curl http://localhost:8081/admin/formations/stats
```

## 配置更新流程

### 1. 通過 Admin API 更新（推薦）

```bash
# 更新配置（自動持久化到 DB + Redis + 熱更新 Spawner）
curl -X PUT http://localhost:8081/admin/formations/config \
  -H "Content-Type: application/json" \
  -d '{
    "base_spawn_chance": 0.4,
    "min_interval": 15,
    "max_interval": 45
  }'
```

### 2. 直接修改資料庫

```sql
-- 更新配置
UPDATE formation_configs
SET config_data = jsonb_set(
    config_data,
    '{base_spawn_chance}',
    '0.5'
)
WHERE config_key = 'default';

-- 更新後需要：
-- 1. 清除 Redis 快取
-- 2. 調用 API 重新載入或重啟服務
```

清除 Redis 快取：
```bash
redis-cli DEL game:formation:config:default
```

## 故障排查

### 問題 1: 配置未生效

**症狀**: 修改配置後遊戲行為沒有改變

**檢查步驟**:
```bash
# 1. 檢查 DB 是否已更新
psql -U postgres -d fish_game -c "SELECT updated_at FROM formation_configs WHERE config_key='default';"

# 2. 檢查 Redis 是否有快取
redis-cli GET game:formation:config:default

# 3. 檢查日誌
tail -f logs/game.log | grep formation
```

**解決方案**:
- 使用 Admin API 更新配置（自動同步）
- 或手動清除 Redis 快取並重啟服務

### 問題 2: Redis 快取過期

**症狀**: Redis 中沒有配置數據

**解決方案**:
```bash
# 調用 API 重新載入
curl -X POST http://localhost:8081/admin/formations/reload
```

或在代碼中：
```go
ctx := context.Background()
if err := formationConfigService.LoadConfigFromDB(ctx); err != nil {
    log.Printf("Failed to reload config: %v", err)
}
```

### 問題 3: 資料庫連接失敗

**症狀**: 啟動時報 "Failed to load formation config from DB"

**檢查**:
```bash
# 檢查資料庫連接
psql -U postgres -d fish_game -c "SELECT 1;"

# 檢查遷移狀態
go run cmd/migrator/main.go version
```

**解決方案**:
- 確保資料庫服務運行
- 檢查連接配置
- 運行遷移腳本

## 最佳實踐

1. **啟動順序**: DB遷移 → 服務啟動 → 載入配置
2. **配置更新**: 優先使用 Admin API（自動同步所有層）
3. **監控**: 定期檢查 Redis 快取命中率
4. **備份**: 定期備份 `formation_configs` 表
5. **日誌**: 啟用 debug 日誌查看配置載入過程

## 性能優化

### Redis 快取策略
- **過期時間**: 1小時（可根據需求調整）
- **預熱**: 服務啟動時自動載入
- **更新**: 配置變更時立即更新

### 資料庫查詢優化
- 使用索引查詢 `config_key`
- JSONB GIN 索引加速 JSON 查詢
- 查詢結果快取到 Redis

## 環境變量

```bash
# 資料庫配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=fish_game

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## 參考文檔

- [陣型配置 API 文檔](./formation_config_api.md)
- [資料庫遷移指南](../storage/migrations/README.md)
- [架構設計文檔](./architecture.md)
