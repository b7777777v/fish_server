# 🐟 Fish Server - 多人捕魚遊戲

一個功能完整的即時多人捕魚遊戲專案，採用 Clean Architecture 設計，支持微服務架構和完整的前端測試客戶端。

## 🎯 專案概述

這是一個高效能、可擴展的多人線上捕魚遊戲，包含：

- **即時遊戲**：使用 WebSocket 實現低延遲通訊
- **微服務架構**：Game Server (遊戲邏輯) + Admin Server (後台管理)
- **完整前端**：基於 Canvas 的原生 JavaScript 遊戲客戶端
- **現代化開發環境**：完善的 VS Code 配置，支持一鍵啟動和多環境調試

## 🛠️ 技術棧

### 後端技術

- **語言**: Go 1.24+
- **Web 框架**: Gin (RESTful API)
- **RPC 框架**: gRPC + Protocol Buffers
- **資料庫**: PostgreSQL, Redis
- **ORM/資料庫工具**: `database/sql`, `pgx`
- **依賴注入**: Google Wire
- **容器化**: Docker, Docker Compose
- **資料庫遷移**: golang-migrate
- **日誌**: `log/slog` (結構化日誌)
- **配置管理**: Viper
- **測試**: testify
- **WebSocket**: gorilla/websocket
- **JWT**: golang-jwt

### 前端技術

- **語言**: JavaScript (原生)
- **渲染引擎**: HTML5 Canvas 2D
- **通訊協議**: WebSocket + Protocol Buffers
- **UI 框架**: 原生 HTML/CSS (無框架依賴)

## 🏗️ 專案架構

### 微服務架構

```
┌─────────────────┐     WebSocket      ┌──────────────┐
│  Game Client    │ ◄─────────────────► │ Game Server  │
│  (Browser)      │                     │  (:9090)     │
└─────────────────┘                     └──────────────┘
                                               │
                                               │ gRPC
┌─────────────────┐      RESTful       ┌──────▼───────┐
│  Admin Panel    │ ◄─────────────────► │ Admin Server │
│  (Browser)      │                     │  (:6060)     │
└─────────────────┘                     └──────────────┘
                                               │
                         ┌─────────────────────┴─────────────────┐
                         │                                       │
                  ┌──────▼───────┐                      ┌───────▼──────┐
                  │  PostgreSQL  │                      │    Redis     │
                  │  (持久化)     │                      │   (快取)     │
                  └──────────────┘                      └──────────────┘
```

### Clean Architecture 分層

```
┌─────────────────────────────────────────────────────────┐
│                    Presentation Layer                    │
│         (Handler: HTTP/WebSocket/gRPC Handlers)         │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                     Business Layer                       │
│  (Service/UseCase: 遊戲邏輯、房間管理、玩家系統)        │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                    Data Access Layer                     │
│        (Repository: 資料庫操作、快取管理)                │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────┐
│                   Infrastructure Layer                   │
│           (PostgreSQL, Redis, WebSocket)                │
└─────────────────────────────────────────────────────────┘
```

## 📁 專案結構

```
fish_server/
├── api/                        # API 定義
│   └── proto/                  # Protobuf 協議定義
│       └── v1/                 # 版本 1 協議
│           ├── game.proto      # 遊戲協議 (WebSocket 消息)
│           └── game_grpc.proto # gRPC 服務定義
│
├── cmd/                        # 應用程式入口
│   ├── admin/                  # Admin Server 主程式
│   │   └── main.go
│   ├── game/                   # Game Server 主程式
│   │   └── main.go
│   └── migrator/               # 資料庫遷移工具
│       └── main.go
│
├── internal/                   # 內部代碼 (不對外暴露)
│   ├── app/                    # 應用層：服務組織與啟動
│   │   ├── admin/              # Admin Server 應用
│   │   └── game/               # Game Server 應用
│   │
│   ├── biz/                    # 業務邏輯層
│   │   ├── domain/             # 領域模型
│   │   │   ├── bullet.go       # 子彈實體
│   │   │   ├── cannon.go       # 砲台實體
│   │   │   ├── fish.go         # 魚實體
│   │   │   ├── formation.go    # 魚群陣型
│   │   │   ├── player.go       # 玩家實體
│   │   │   ├── room.go         # 房間實體
│   │   │   └── wallet.go       # 錢包實體
│   │   │
│   │   ├── usecase/            # 業務用例
│   │   │   ├── player.go       # 玩家相關業務
│   │   │   ├── wallet.go       # 錢包相關業務
│   │   │   ├── game.go         # 遊戲核心邏輯
│   │   │   └── room.go         # 房間管理
│   │   │
│   │   └── service/            # 領域服務
│   │       ├── fish_spawner.go # 魚群生成服務
│   │       ├── formation.go    # 陣型計算服務
│   │       └── reward.go       # 獎勵計算服務
│   │
│   ├── data/                   # 資料存取層
│   │   ├── repo/               # Repository 實現
│   │   │   ├── player.go       # 玩家資料存取
│   │   │   ├── wallet.go       # 錢包資料存取
│   │   │   └── transaction.go  # 交易記錄存取
│   │   │
│   │   └── cache/              # 快取層
│   │       └── session.go      # 會話快取
│   │
│   ├── handler/                # 處理器層
│   │   ├── admin/              # Admin API 處理器
│   │   │   ├── health.go       # 健康檢查
│   │   │   ├── player.go       # 玩家管理 API
│   │   │   └── wallet.go       # 錢包管理 API
│   │   │
│   │   └── game/               # Game Server 處理器
│   │       ├── websocket.go    # WebSocket 連接管理
│   │       ├── message.go      # 消息路由與處理
│   │       └── broadcast.go    # 廣播服務
│   │
│   ├── middleware/             # HTTP 中介軟體
│   │   ├── cors.go             # CORS 處理
│   │   ├── logger.go           # 請求日誌
│   │   └── recovery.go         # Panic 恢復
│   │
│   ├── conf/                   # 配置結構
│   │   └── config.go           # 配置映射
│   │
│   └── pkg/                    # 內部共享工具
│       ├── jwt/                # JWT 工具
│       ├── errors/             # 錯誤處理
│       └── validator/          # 驗證工具
│
├── pkg/                        # 公共可重用代碼
│   ├── logger/                 # 日誌封裝
│   ├── database/               # 資料庫連接
│   └── redis/                  # Redis 客戶端
│
├── js/                         # 前端測試客戶端 ⭐
│   ├── index.html              # 遊戲客戶端主頁面
│   ├── game-client.js          # WebSocket 客戶端
│   ├── game-renderer.js        # Canvas 渲染引擎
│   └── generated/              # Protobuf 生成的 JS 代碼
│       └── game_pb.js
│
├── configs/                    # 配置檔案
│   └── config.yaml             # 主配置檔案
│
├── deployments/                # 部署相關
│   ├── docker-compose.dev.yml  # 開發環境 Docker Compose
│   ├── docker-compose.yml      # 生產環境 Docker Compose
│   └── .env.example            # 環境變數範例
│
├── storage/                    # 資料庫遷移
│   └── migrations/             # SQL 遷移檔案
│       ├── 000001_init.up.sql
│       └── 000001_init.down.sql
│
├── scripts/                    # 輔助腳本
│   ├── proto-gen.sh            # Protobuf 生成腳本
│   ├── wire-gen.sh             # Wire 生成腳本
│   └── start-database.bat      # Windows 資料庫啟動腳本
│
├── .vscode/                    # VS Code 配置 ⭐
│   ├── launch.json             # 調試配置
│   ├── tasks.json              # 任務配置
│   ├── settings.json           # 工作區設定
│   └── README.md               # VS Code 使用指南
│
├── docs/                       # 專案文檔
│   ├── FISH_FORMATION_GUIDE.md # 魚群陣型指南
│   └── FRONTEND_FISH_DYNAMICS_GUIDE.md # 前端動畫指南
│
├── Makefile                    # Make 命令
├── go.mod                      # Go 模組定義
├── go.sum                      # Go 依賴鎖定
└── README.md                   # 專案說明
```

## 📋 編碼規範

### Go 編碼標準

#### 1. Clean Code 原則

- **SOLID 原則**：遵循單一職責、開放封閉、里氏替換、介面隔離、依賴反轉
- **函數大小**：每個函數保持小而專注，單一職責
- **組合優於繼承**：使用組合而非繼承來擴展功能
- **自我文檔化**：使用清晰的命名，減少註釋需求

#### 2. 錯誤處理

```go
// ✅ 正確：明確處理錯誤
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ 錯誤：忽略錯誤
result, _ := doSomething()

// ✅ 正確：包裝錯誤以提供上下文
if err := processData(data); err != nil {
    return fmt.Errorf("process data for user %d: %w", userID, err)
}
```

#### 3. 命名慣例

```go
// 未導出 (private)
type playerService struct {}
func calculateReward() {}

// 導出 (public)
type PlayerService struct {}
func CalculateReward() {}

// 介面命名：通常使用 -er 後綴
type Reader interface {}
type Writer interface {}
type PlayerRepository interface {}

// 好的命名範例
getUserByID()     // 清楚描述功能
findActiveRooms() // 清楚描述功能

// 避免的命名
get()    // 太模糊
doStuff() // 不清楚
```

#### 4. 套件組織

```go
// ✅ 正確：套件專注且內聚
package player

type Player struct {}
func (p *Player) Join() {}
func (p *Player) Leave() {}

// ❌ 錯誤：循環依賴
// package A 引用 package B
// package B 引用 package A

// ✅ 正確：依賴方向
// domain (核心) ← service ← handler
// 內層不依賴外層
```

#### 5. 依賴注入 (Wire)

```go
// ✅ 在 wire.go 中定義 Provider
//go:build wireinject
// +build wireinject

package main

import "github.com/google/wire"

func InitializeGameServer() (*GameServer, error) {
    wire.Build(
        NewDatabase,
        NewPlayerRepository,
        NewPlayerService,
        NewGameHandler,
        NewGameServer,
    )
    return nil, nil
}

// ✅ 建構函數注入介面
func NewPlayerService(repo PlayerRepository) *PlayerService {
    return &PlayerService{repo: repo}
}
```

#### 6. 測試

```go
// ✅ 使用表格驅動測試
func TestCalculateReward(t *testing.T) {
    tests := []struct {
        name     string
        fishType FishType
        want     int
    }{
        {"small fish", FishTypeSmall, 10},
        {"medium fish", FishTypeMedium, 50},
        {"boss fish", FishTypeBoss, 1000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateReward(tt.fishType)
            assert.Equal(t, tt.want, got)
        })
    }
}

// ✅ 使用 testify 進行斷言和模擬
func TestPlayerService_GetPlayer(t *testing.T) {
    mockRepo := new(MockPlayerRepository)
    service := NewPlayerService(mockRepo)
    
    mockRepo.On("FindByID", 123).Return(&Player{ID: 123}, nil)
    
    player, err := service.GetPlayer(123)
    assert.NoError(t, err)
    assert.Equal(t, 123, player.ID)
    mockRepo.AssertExpectations(t)
}
```

#### 7. 日誌 (slog)

```go
// ✅ 使用結構化日誌
slog.Info("player joined room",
    "player_id", playerID,
    "room_id", roomID,
    "timestamp", time.Now(),
)

// ✅ 適當的日誌級別
slog.Debug("detailed debug info")  // 開發階段詳細資訊
slog.Info("user action")            // 一般資訊
slog.Warn("unusual condition")      // 警告但可處理
slog.Error("operation failed")      // 錯誤需要關注

// ❌ 避免：記錄敏感資料
slog.Info("user login", "password", password) // 絕對不要這樣做！
```

#### 8. 資料庫操作

```go
// ✅ 使用事務處理多步驟操作
func (r *WalletRepo) Transfer(ctx context.Context, from, to int64, amount int64) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    if err := r.deduct(ctx, tx, from, amount); err != nil {
        return err
    }
    if err := r.add(ctx, tx, to, amount); err != nil {
        return err
    }

    return tx.Commit()
}

// ✅ 使用參數化查詢
query := "SELECT * FROM players WHERE id = $1"
row := db.QueryRowContext(ctx, query, playerID)

// ❌ 避免 SQL 注入
query := fmt.Sprintf("SELECT * FROM players WHERE name = '%s'", name) // 危險！
```

### 前端 JavaScript 標準

#### 1. 現代 ES6+ 特性

```javascript
// ✅ 使用 const/let，避免 var
const MAX_PLAYERS = 4;
let currentScore = 0;

// ✅ 使用箭頭函數
const handleClick = (event) => {
    console.log('Clicked:', event);
};

// ✅ 解構賦值
const { playerId, roomId } = playerData;

// ✅ 模板字串
const message = `Player ${playerId} joined room ${roomId}`;
```

#### 2. Canvas 渲染

```javascript
// ✅ 使用 requestAnimationFrame
function gameLoop(timestamp) {
    update(timestamp);
    render();
    requestAnimationFrame(gameLoop);
}
requestAnimationFrame(gameLoop);

// ✅ 處理高 DPI 顯示器
const dpr = window.devicePixelRatio || 1;
canvas.width = width * dpr;
canvas.height = height * dpr;
ctx.scale(dpr, dpr);

// ✅ 高效清除與重繪
ctx.clearRect(0, 0, canvas.width, canvas.height);
```

#### 3. WebSocket 通訊

```javascript
// ✅ 實作重連邏輯
class GameWebSocket {
    connect() {
        this.ws = new WebSocket(this.url);
        this.ws.onclose = () => this.handleReconnect();
    }
    
    handleReconnect() {
        setTimeout(() => this.connect(), 5000);
    }
}

// ✅ 使用 Protobuf 序列化
const message = proto.GameMessage.create({
    type: proto.MessageType.FIRE_BULLET,
    fireBullet: { targetX: x, targetY: y }
});
const buffer = proto.GameMessage.encode(message).finish();
ws.send(buffer);
```

#### 4. 代碼組織

```javascript
// ✅ 模組化設計
class GameClient {
    constructor(wsUrl) {
        this.wsUrl = wsUrl;
        this.ws = null;
    }
    
    connect() { /* ... */ }
    sendMessage(message) { /* ... */ }
}

class GameRenderer {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
    }
    
    render(gameState) { /* ... */ }
    drawFish(fish) { /* ... */ }
}

// ✅ 分離關注點
// game-client.js - WebSocket 通訊
// game-renderer.js - Canvas 渲染
// game-state.js - 狀態管理
```

## 🎮 遊戲系統說明

### 核心系統

#### 1. 房間系統

- 支持多個獨立遊戲房間
- 每個房間最多 4 名玩家
- 房間狀態：等待中、遊戲中、已結束
- 自動魚群生成與管理

#### 2. 砲台系統

- 5 個等級的砲台 (Level 1-5)
- 不同等級消耗不同金幣
- 不同等級有不同的捕獲機率加成

#### 3. 魚群系統

- **魚的類型**：小魚、中魚、大魚、BOSS 魚
- **陣型系統**：V字型、直線型、圓形、三角形、菱形、波浪型、螺旋型
- **路線系統**：直線、曲線、Z字型、圓形、螺旋、波浪、三角巡邏、隨機
- 詳細資訊請參考 `FISH_FORMATION_GUIDE.md`

#### 4. 獎勵系統

- 根據魚的類型計算獎勵
- 砲台等級影響捕獲機率
- 獎勵直接添加到玩家錢包

### WebSocket 消息類型

#### 客戶端請求 (C → S)

```
FIRE_BULLET       - 開火射擊
SWITCH_CANNON     - 切換砲台
JOIN_ROOM         - 加入房間
LEAVE_ROOM        - 離開房間
GET_ROOM_LIST     - 獲取房間列表
GET_PLAYER_INFO   - 獲取玩家資訊
HEARTBEAT         - 心跳保持連接
```

#### 伺服器回應 (S → C)

```
*_RESPONSE        - 對應請求的回應
WELCOME           - 連接成功歡迎消息
ERROR             - 錯誤消息
```

#### 伺服器廣播 (S → All)

```
BULLET_FIRED      - 有玩家開火
CANNON_SWITCHED   - 有玩家切換砲台
FISH_SPAWNED      - 魚群生成
FISH_DIED         - 魚被捕獲
PLAYER_JOINED     - 玩家加入房間
PLAYER_LEFT       - 玩家離開房間
PLAYER_REWARD     - 玩家獲得獎勵
```

## 🔄 常用開發任務

### 本地開發

```bash
# 啟動資料庫服務 (PostgreSQL + Redis)
make run-dev
# Windows: scripts\start-database.bat

# 執行資料庫遷移
make migrate-up
# Windows: scripts\run-migration.bat up

# 啟動 Game Server
make run-game

# 啟動 Admin Server
make run-admin

# 同時啟動所有服務（推薦使用 VS Code）
# 打開 Run and Debug → 選擇 "🚀 DEV Environment - All Services"
```

### 程式碼生成

```bash
# 生成所有代碼
make gen

# 只生成 Protobuf (Go + JavaScript)
make proto
sh ./scripts/proto-gen.sh

# 只生成 Wire 依賴注入
make wire
sh ./scripts/wire-gen.sh
```

### 測試與檢查

```bash
# 執行所有測試
make test
go test ./...

# 測試覆蓋率
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 程式碼檢查 (Linter)
make lint
golangci-lint run
```

### 資料庫操作

```bash
# 創建新遷移
migrate create -ext sql -dir storage/migrations -seq add_new_table

# 執行遷移
make migrate-up
migrate -path storage/migrations -database "postgres://..." up

# 回滾遷移
make migrate-down
migrate -path storage/migrations -database "postgres://..." down 1

# 查看遷移狀態
migrate -path storage/migrations -database "postgres://..." version
```

### Docker 操作

```bash
# 啟動所有服務
docker-compose up --build

# 背景執行
docker-compose up -d

# 查看日誌
docker-compose logs -f game-server
docker-compose logs -f admin-server

# 停止服務
docker-compose down

# 重建特定服務
docker-compose up --build game-server
```

## 🔒 安全指引

### 重要安全實踐

1. **輸入驗證**
   - 所有用戶輸入必須驗證
   - 使用 validator 套件進行結構驗證
   - 檢查數值範圍和格式

2. **認證與授權**
   - 使用 JWT 進行玩家認證
   - Token 過期時間設置合理（默認 2 小時）
   - 敏感操作需要驗證 Token

3. **SQL 注入防護**
   - **永遠使用參數化查詢**
   - 絕對不要拼接 SQL 字串

   ```go
   // ✅ 安全
   db.QueryRow("SELECT * FROM players WHERE id = $1", playerID)
   
   // ❌ 危險
   db.QueryRow(fmt.Sprintf("SELECT * FROM players WHERE id = %d", playerID))
   ```

4. **敏感資料**
   - 絕不在日誌中記錄密碼、Token
   - 使用環境變數管理密鑰
   - `.env` 檔案不要提交到 Git

5. **速率限制**
   - 實作 API 速率限制
   - WebSocket 消息頻率限制
   - 防止暴力攻擊

6. **CORS 配置**
   - 生產環境正確配置 CORS
   - 不要使用 `Access-Control-Allow-Origin: *`

## 🐛 調試技巧

### Go 後端調試

```bash
# 使用 Delve 調試器
dlv debug cmd/game/main.go
dlv debug cmd/admin/main.go

# VS Code 調試（推薦）
# F5 啟動調試配置
# 支持斷點、變數檢查、步進執行
```

### 日誌分析

```bash
# 查看結構化日誌
tail -f logs/game-server.log | jq

# 過濾特定玩家的日誌
grep "player_id=123" logs/game-server.log

# 查看錯誤日誌
grep "level=error" logs/game-server.log
```

### 資料庫調試

```bash
# 使用 psql 連接
psql -h localhost -U user -d fish_db

# 查看連接狀態
SELECT * FROM pg_stat_activity;

# 查看慢查詢
SELECT * FROM pg_stat_statements ORDER BY mean_time DESC;
```

### Redis 調試

```bash
# 連接 Redis
redis-cli

# 查看所有鍵
KEYS *

# 監控命令
MONITOR

# 查看記憶體使用
INFO memory
```

### 前端調試

- 開啟瀏覽器開發者工具 (F12)
- **Console**: 查看 JavaScript 日誌和錯誤
- **Network**: 監控 WebSocket 連接
- **Application**: 查看 LocalStorage (如果使用)
- 遊戲客戶端內建消息日誌面板

## 💡 最佳實踐

### 開發工作流程

1. **開始新功能**

   ```bash
   git checkout -b feature/new-feature
   ```

2. **編寫代碼**
   - 遵循 Clean Code 原則
   - 保持函數小而專注
   - 編寫自我文檔化的代碼

3. **編寫測試**
   - 為核心業務邏輯編寫單元測試
   - 目標覆蓋率 > 80%

4. **提交前檢查**

   ```bash
   make lint  # 代碼檢查
   make test  # 執行測試
   ```

5. **提交代碼**

   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

6. **合併前整理**

   ```bash
   git rebase main
   ```

### Git 提交規範

使用語義化提交訊息：

```
feat: 新增功能
fix: 修復 Bug
docs: 文檔更新
style: 代碼格式調整
refactor: 代碼重構
test: 測試相關
chore: 雜項（構建、配置等）
```

範例：

```
feat: add fish formation system
fix: resolve room join race condition
docs: update API documentation
refactor: improve player service structure
```

## 📚 相關資源

### 官方文檔

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Gin Documentation](https://gin-gonic.com/docs/)
- [gRPC Go Quick Start](https://grpc.io/docs/languages/go/quickstart/)
- [Wire User Guide](https://github.com/google/wire/blob/main/docs/guide.md)

### 推薦書籍

- Clean Code (Robert C. Martin)
- Clean Architecture (Robert C. Martin)
- The Go Programming Language (Donovan & Kernighan)

### 專案文檔

- [魚群陣型系統](docs/FISH_FORMATION_GUIDE.md)
- [前端動畫指南](docs/FRONTEND_FISH_DYNAMICS_GUIDE.md)
- [Windows 開發指南](WINDOWS_GUIDE.md)
- [VS Code 配置說明](.vscode/README.md)

## 💡 使用 Claude Code 時的提示

當你使用 Claude Code 協助開發時，請提供以下上下文以獲得更好的協助：

### 開發上下文

- **指定分層**：明確說明你在哪一層工作（Handler/Service/Repository/Domain）
- **架構影響**：說明變更是否需要重新生成 Wire 或 Protobuf
- **資料庫變更**：如果涉及資料庫結構，說明是否需要建立遷移檔案
- **測試要求**：說明新功能是否需要測試覆蓋

### 常見開發場景

#### 添加新的遊戲功能

```
請在 Game Server 中添加 [功能名稱]：
- 需要在 domain/ 中定義新的實體嗎？
- Service 層需要什麼業務邏輯？
- WebSocket 需要新的消息類型嗎？（需要更新 proto）
- 需要資料庫持久化嗎？（需要 Repository 和遷移）
```

#### 修改現有功能

```
需要修改 [功能名稱]：
- 目前的實現在 [檔案路徑]
- 問題是 [描述問題]
- 期望的行為是 [描述期望]
- 是否影響其他模組？
```

#### 優化效能

```
[功能] 效能需要優化：
- 目前的瓶頸在 [位置]
- 預期的改善目標
- 是否可以使用快取？
- 是否需要調整資料庫查詢？
```

#### 修復 Bug

```
發現 Bug 在 [功能/位置]：
- 重現步驟
- 預期行為 vs 實際行為
- 相關日誌或錯誤訊息
- 可能影響的範圍
```

### 程式碼生成需求

**Protobuf 更新後**

```bash
# 提醒：需要重新生成 Protobuf 代碼
make proto
# 或
sh ./scripts/proto-gen.sh
```

**Wire 依賴更新後**

```bash
# 提醒：需要重新生成 Wire 代碼
make wire
# 或
sh ./scripts/wire-gen.sh
```

**資料庫結構變更**

```bash
# 提醒：需要建立遷移檔案
migrate create -ext sql -dir storage/migrations -seq [migration_name]
```

### 代碼審查重點

請協助檢查：

- ✅ 是否遵循 Clean Architecture 分層原則
- ✅ 錯誤處理是否完整
- ✅ 是否有潛在的併發問題
- ✅ 資料庫操作是否使用事務
- ✅ 是否有 SQL 注入風險
- ✅ 日誌是否包含敏感資訊
- ✅ 是否需要添加測試

## 🎯 專案目標與優先級

### 短期目標

- [ ] 完善遊戲核心邏輯
- [ ] 優化魚群生成演算法
- [ ] 提升 WebSocket 效能
- [ ] 增加單元測試覆蓋率

### 中期目標

- [ ] 實作排行榜系統
- [ ] 添加更多魚種和 BOSS
- [ ] 實作玩家成就系統
- [ ] 優化前端渲染效能

### 長期目標

- [ ] 支援水平擴展（多伺服器）
- [ ] 實作觀戰模式
- [ ] 添加錄像回放功能
- [ ] 實作 AI 玩家

## 🔧 疑難排解

### 常見問題

#### 1. 資料庫連接失敗

```bash
# 檢查 PostgreSQL 是否運行
docker ps | grep postgres

# 檢查連接字串
# 確保 .env.dev 中的設定正確

# 測試連接
psql -h localhost -U user -d fish_db
```

#### 2. Redis 連接失敗

```bash
# 檢查 Redis 是否運行
docker ps | grep redis

# 測試連接
redis-cli ping
```

#### 3. WebSocket 連接失敗

- 確保 Game Server 正在運行
- 檢查防火牆設定
- 確認 WebSocket URL 正確（ws:// 或 wss://）
- 查看瀏覽器 Console 錯誤訊息

#### 4. Protobuf 生成失敗

```bash
# 確保已安裝 protoc
protoc --version

# 確保已安裝 Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 重新生成
make proto
```

#### 5. Wire 生成失敗

```bash
# 確保已安裝 Wire
go install github.com/google/wire/cmd/wire@latest

# 檢查 wire.go 語法
# 重新生成
make wire
```

## 📝 變更日誌

### 最近更新

- ✨ 新增完整的前端遊戲客戶端
- ✨ 實作魚群陣型系統
- ✨ 添加多種魚群路線演算法
- 🔧 優化 WebSocket 訊息處理
- 📝 完善 VS Code 開發環境配置
- 🐛 修復房間加入的競爭條件問題

## 🤝 貢獻指南

### 如何貢獻

1. Fork 專案
2. 建立功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交變更 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 開啟 Pull Request

### Pull Request 檢查清單

- [ ] 代碼通過 `make lint` 檢查
- [ ] 所有測試通過 `make test`
- [ ] 添加了必要的測試
- [ ] 更新了相關文檔
- [ ] 遵循專案的編碼規範
- [ ] Commit 訊息符合規範
- [ ] 沒有提交敏感資訊（密碼、Token 等）

## 📄 授權

[請根據實際情況添加授權資訊]

## 🙏 致謝

感謝所有貢獻者對本專案的支持！

---

**Happy Coding! 🎮🐟**
