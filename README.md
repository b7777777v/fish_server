# 🐟 Fish Server - 多人捕魚遊戲全端專案

這是一個功能完整的多人捕魚遊戲專案，包含後端服務和前端測試客戶端。後端採用了現代化的 Go 語言、微服務架構和雲原生技術，前端提供了基於 WebSocket 的即時遊戲客戶端，旨在提供一個高效、可擴展且易於維護的完整解決方案。

此專案不僅包含了完整的業務邏輯和前端遊戲界面，還提供了一套極其完善的本地開發環境 (`.vscode`)，讓開發者可以實現一鍵啟動、多環境偵錯和自動化任務，極大地提升了開發效率。

> **最新更新**: 新增完整的前端遊戲客戶端，支持即時遊戲交互、房間管理和完整的遊戲功能展示！

## ✨ 主要功能

### 後端服務
- **即時多人遊戲**：使用 WebSocket 實現低延遲的即時客戶端/伺服器通訊。
- **遊戲房管理**：創建、加入、離開遊戲房間，支持多房間並行。
- **核心遊戲邏輯**：包括魚群生成、捕獲機率、分數計算、砲台系統等。
- **玩家系統**：完整的玩家資料管理與 JWT 認證。
- **錢包系統**：管理玩家的遊戲內貨幣，支持凍結/解凍、存提款等操作。
- **後台管理**：提供一個獨立的 `admin` 服務，用於遊戲管理與監控。
- **多環境支援**：完整的 `DEV`, `STAGING`, `PROD` 環境隔離與配置。

### 前端客戶端
- **完整遊戲界面**：基於 Canvas 的 2D 遊戲渲染引擎。
- **即時互動**：WebSocket 即時通訊，支持所有遊戲操作。
- **房間管理**：支持查看房間列表、加入/離開房間。
- **遊戲功能**：開火射擊、切換砲台、魚群動畫、分數顯示等。
- **調試工具**：完整的消息日誌和狀態監控，方便開發調試。

## 🏛️ 架構概覽

專案採用了清晰的 **Clean Architecture** 設計理念，將業務邏輯與基礎設施分離。

- **微服務架構**：
  - `game-server`: 處理核心遊戲邏輯與玩家互動 (WebSocket)。
  - `admin-server`: 提供 RESTful API 用於後台管理。
  - `migrator`: 用於資料庫遷移。
- **通訊**：
  - 客戶端與 `game-server` 之間使用 **WebSocket**。
  - 服務內部或對外的 API 使用 **gRPC** 和 **RESTful (Gin)**。
- **資料庫**：
  - **PostgreSQL**: 用於持久化儲存核心業務資料 (玩家、錢包等)。
  - **Redis**: 用於快取、會話管理或發布/訂閱。
- **依賴注入**：使用 **Google Wire** 管理服務間的依賴關係，實現松耦合。
- **容器化**：所有服務都被設計為在 **Docker** 中運行。

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

## 🚀 快速開始

### 1. 環境準備

在開始之前，請確保您已安裝以下工具：

- **Go**: 1.24 或更高版本
- **Docker** 和 **Docker Compose**
- **Make**
- **golang-migrate**: [安裝指南](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- **golangci-lint**: [安裝指南](https://golangci-lint.run/usage/install/)

> **可選**：現代瀏覽器用於訪問前端測試客戶端 (需支持 WebSocket 和 Canvas)

### 2. 專案設定

1. **Clone 專案**

    ```bash
    git clone <repository-url>
    cd fish_server
    ```

2. **設定環境變數**
    專案的 Docker 環境依賴 `.env` 檔案。

    ```bash
    # 從範本複製開發環境設定
    cp deployments/.env.example deployments/.env.dev
    ```

    > 您可以根據需要修改 `.env.dev` 中的資料庫密碼或其他設定。

3. **啟動基礎設施**
    使用 Docker Compose 啟動資料庫 (PostgreSQL, Redis) 等依賴服務。

    ```bash
    make run-dev
    ```

    > 這將會啟動 `deployments/docker-compose.yml` 中定義的服務。首次啟動會需要一些時間來下載鏡像。

4. **執行資料庫遷移**
    在另一個終端中，執行以下命令來初始化資料庫結構。

    ```bash
    # !! 注意 !!
    # Makefile 中的 DB_URL 可能需要根據您的 .env.dev 設定進行調整
    make migrate-up
    ```

### 3. 啟動應用程式

#### 方式一：使用 VS Code (推薦)

本專案提供了極其強大的 VS Code 多環境配置，強烈建議使用。

1. 在 VS Code 中打開 `fish_server.code-workspace` 工作區。
2. 打開 "Run and Debug" 側邊欄 (Ctrl+Shift+D)。
3. 從頂部的下拉列表中選擇一個啟動選項，例如：
    - `🚀 DEV Environment - All Services`: 一鍵啟動所有開發環境服務。
    - `🟢 Admin Server - DEV (Pprof ON)`: 單獨啟動 Admin 服務。
    - `🔍 Debug Admin with Delve`: 啟動並偵錯 Admin 服務。
4. 按下 `F5` 即可啟動。

> 關於 VS Code 環境的詳細用法，請參考 [.vscode/README.md](.vscode/README.md)。

#### 方式二：使用 `make` 命令

您也可以使用 `Makefile` 中的命令來手動編譯和執行。

```bash
# 編譯所有服務
make build

# 單獨運行 Game Server
make run-game

# 單獨運行 Admin Server
make run-admin
```

### 4. 訪問前端測試客戶端

前端測試客戶端已經集成到 `admin-server` 中，無需額外啟動服務。

#### 訪問方式

確保 Admin Server 已經啟動（默認端口：6060），然後在瀏覽器中訪問：

```
http://localhost:6060/test-client
```

> **提示**：前端測試客戶端是專門用來檢測和調試 Game Server 的工具，因此集成到 Admin Server 中最為合適。

#### 使用說明

1. **設定服務器地址**：在頁面頂部輸入遊戲服務器地址（默認：`ws://localhost:9090/ws`）
2. **輸入玩家信息**：輸入玩家 ID 和 Token（測試環境可使用任意值）
3. **連接服務器**：點擊「連接」按鈕建立 WebSocket 連接
4. **開始遊戲**：
   - 點擊「獲取房間列表」查看可用房間
   - 點擊「加入房間」進入遊戲
   - 使用「開火」按鈕或點擊遊戲畫面射擊
   - 使用「切換砲台」按鈕切換不同威力的砲台
5. **查看日誌**：所有 WebSocket 消息都會在右側日誌區域顯示，方便調試

> **提示**：確保遊戲服務器已經啟動並運行在 `localhost:9090`，否則無法連接。

## ⚙️ 配置

專案使用 `Viper` 來管理配置，支持從配置文件、環境變數等多種來源讀取配置。

### 配置文件

主要的配置文件是 `configs/config.yaml`，其中定義了所有可用的配置項及其默認值。

```yaml
# configs/config.yaml
log:
  level: "debug" # 可選值: debug, info, warn, error
  format: "json" # 可選值: json, console

server:
  game:
    port: "9090"
  admin:
    port: "6060"

data:
  database:
    driver: "postgres"
    host: "localhost"
    port: 5432
    user: "user"
    password: "password"
    dbname: "fish_db"
    sslmode: "disable"
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0

jwt:
  secret: "your-super-secret-key" # 務必修改成一個複雜的密鑰
  issuer: "fish_server" # token 發行者
  expire: 7200 # token 過期時間，單位為秒 (例如 7200 表示 2 小時)
```

### 環境變數

您可以使用環境變數來覆蓋配置文件中的值。環境變數的命名規則是 `[SECTION]_[KEY]`，例如：

```bash
export SERVER_GAME_PORT=9091
```

## 🐟 遊戲邏輯

### 魚群路線和陣型

遊戲中的魚群行為由路線和陣型系統控制。

#### 陣型類型

- **V字型** (`FormationTypeV`)
- **直線型** (`FormationTypeLine`)
- **圓形** (`FormationTypeCircle`)
- **三角形** (`FormationTypeTriangle`)
- **菱形** (`FormationTypeDiamond`)
- **波浪型** (`FormationTypeWave`)
- **螺旋型** (`FormationTypeSpiral`)

#### 路線類型

- **直線路線**
- **曲線路線**
- **Z字型路線**
- **圓形路線**
- **螺旋路線**
- **波浪路線**
- **三角巡邏**
- **隨機路線**

詳細信息請參考 [FISH_FORMATION_GUIDE.md](FISH_FORMATION_GUIDE.md)。

## 🎮 遊戲客戶端

### 前端數據推送

後端會通過 WebSocket 向前端推送遊戲狀態和事件。

#### 消息類型

- `ROOM_STATE_UPDATE`: 定期推送的完整房間狀態。
- `FORMATION_SPAWNED`: 魚群陣型生成事件。
- `FISH_SPAWNED`: 單個魚生成事件。
- `FISH_DIED`: 魚死亡事件。
- `BULLET_FIRED`: 子彈發射事件。

詳細信息請參考 [FRONTEND_FISH_DYNAMICS_GUIDE.md](FRONTEND_FISH_DYNAMICS_GUIDE.md)。

## 🔧 開發流程

### 代碼生成

專案使用 `go generate` 來自動生成 Protobuf 和 Wire 的代碼。

```bash
# 生成所有代碼
make gen

# 只生成 Protobuf 代碼
make proto

# 只生成 Wire 依賴注入代碼
make wire
```

> 在新增或修改 `.proto` 檔案或 `wire.go` 檔案後，需要執行這些命令。

### 測試

```bash
# 運行所有測試
make test

# 運行 Linter
make lint
```

### 清理

```bash
# 清理所有編譯產物和快取
make clean
```

## 📁 專案結構

```
.
├── api/                # Protobuf 定義 (遊戲協議、gRPC 服務)
├── cmd/                # 主應用程式入口 (main.go)
│   ├── admin/          # 後台管理服務
│   ├── game/           # 遊戲服務
│   └── migrator/       # 資料庫遷移工具
├── configs/            # 環境配置文件 (config.yaml)
├── deployments/        # Docker 和部署相關文件
├── docs/               # 專案文檔
├── internal/           # 專案內部代碼 (不對外暴露)
│   ├── app/            # 應用層: 服務的啟動與組織
│   ├── biz/            # 業務邏輯層: 核心業務實體與用例
│   ├── conf/           # 配置映射結構
│   ├── data/           # 資料存取層: Repository 實現
│   └── pkg/            # 專案內部共享的工具包
├── js/                 # 前端測試客戶端 ⭐ 新增
│   ├── index.html      # 遊戲客戶端主頁面
│   ├── game-client.js  # WebSocket 客戶端和遊戲邏輯
│   ├── game-renderer.js # Canvas 渲染引擎
│   └── generated/      # Protobuf 生成的 JavaScript 代碼
├── pkg/                # 可供外部專案使用的共享代碼
├── scripts/            # 各類輔助腳本 (proto-gen, wire-gen 等)
├── storage/            # 資料庫遷移文件
└── .vscode/            # VS Code 工作區配置
```

## 🎨 前端客戶端開發

前端測試客戶端已整合到 Admin Server 中，通過 `/test-client` 路由提供服務。客戶端採用模組化設計，主要包含以下組件：

### 主要模組

1. **game-client.js** - WebSocket 客戶端
   - 處理 WebSocket 連接和消息收發
   - Protobuf 消息的序列化/反序列化
   - 遊戲狀態管理

2. **game-renderer.js** - Canvas 渲染引擎
   - 遊戲畫面渲染（魚群、子彈、砲台等）
   - 動畫處理和碰撞檢測
   - UI 繪制和交互

3. **generated/** - Protobuf 生成代碼
   - 與後端共享的消息定義
   - 自動生成的 JavaScript Protobuf 代碼

### 開發指南

#### 訪問測試客戶端

1. 啟動 Admin Server：`make run-admin`
2. 在瀏覽器中訪問：`http://localhost:6060/test-client`
3. 確保 Game Server 也在運行（默認端口：9090）

#### 修改 Protobuf 定義後重新生成前端代碼

```bash
# 在項目根目錄執行
make proto

# 或手動執行生成腳本
sh ./scripts/proto-gen.sh
```

生成的 JavaScript Protobuf 代碼會自動放置在 `js/generated/` 目錄下。

#### 調試技巧

- 打開瀏覽器開發者工具查看 Console 日誌
- 使用頁面右側的消息日誌查看所有 WebSocket 通訊
- 通過 Network 面板監控 WebSocket 連接狀態
- Admin Server 會在啟動時顯示測試客戶端的訪問地址

## 🤝 貢獻

歡迎提交 Issue 和 Pull Request！

在提交代碼前，請確保：
- 代碼通過 `make lint` 檢查
- 所有測試通過 `make test`
- 更新相關文檔

## API 接口文檔

本文件概述了 `fish_server` 專案中的所有主要 API 接口，包括後台管理的 RESTful API、遊戲核心的 gRPC 服務以及客戶端與伺服器之間的 WebSocket 通訊協議。

### Admin API (RESTful)

此 API 主要用於後台管理、監控和數據查詢。所有端點都以 `/admin` 為前綴。

| 方法 (Method) | 路徑 (Path)                      | 描述 (Description)                               |
|---------------|----------------------------------|--------------------------------------------------|
| `GET`         | `/`                              | 顯示 API 根信息                                  |
| `GET`         | `/ping`                          | 簡單的 Ping-Pong 檢查                            |
| `GET`         | `/admin/health`                  | 一般健康檢查                                     |
| `GET`         | `/admin/health/live`             | Kubernetes 存活探測 (Liveness Probe)             |
| `GET`         | `/admin/health/ready`            | Kubernetes 就緒探測 (Readiness Probe)            |
| `GET`         | `/admin/status`                  | 獲取詳細的伺服器運行狀態 (記憶體、協程等)        |
| `GET`         | `/admin/metrics`                 | 獲取 Prometheus 格式的詳細指標                   |
| `GET`         | `/admin/env`                     | 獲取當前環境配置信息                           |
| `GET`         | `/admin/players/:id`             | 獲取指定 ID 的玩家信息                           |
| `GET`         | `/admin/players/:id/wallets`     | 獲取指定玩家的所有錢包                           |
| `GET`         | `/admin/wallets/:id`             | 獲取指定 ID 的錢包信息                           |
| `GET`         | `/admin/wallets/:id/transactions`| 獲取指定錢包的交易記錄                           |
| `POST`        | `/admin/wallets/:id/freeze`      | 凍結指定錢包                                     |
| `POST`        | `/admin/wallets/:id/unfreeze`    | 解凍指定錢包                                     |
| `POST`        | `/admin/wallets/:id/deposit`     | 向指定錢包存款 (增加餘額)                        |
| `POST`        | `/admin/wallets/:id/withdraw`    | 從指定錢包提款 (減少餘額)                        |
| `GET`         | `/debug/pprof/*`                 | (可選) Go pprof 性能分析端點                     |

### 遊戲服務 (gRPC)

此服務用於需要後端驗證或處理的遊戲相關操作，例如登入。

**服務名稱**: `v1.Game`

| RPC 方法 (Method) | 請求 (Request)         | 回應 (Response)        | 描述 (Description) |
|-------------------|------------------------|------------------------|--------------------|
| `Login`           | `v1.LoginRequest`      | `v1.LoginResponse`     | 玩家帳號密碼登入   |

### 遊戲客戶端通訊 (WebSocket)

客戶端與遊戲伺服器之間的主要通訊方式。所有消息都封裝在 `v1.GameMessage` 中，通過 `MessageType` 枚舉來區分。

#### 消息方向

*   **C -> S**: 客戶端發送到伺服器
*   **S -> C**: 伺服器發送到客戶端 (單播或廣播)

#### 消息類型 (`v1.MessageType`)

| 類型 (Type)                | 方向   | 數據結構 (Payload)             | 描述 (Description)                               |
|----------------------------|--------|--------------------------------|--------------------------------------------------|
| **客戶端請求**             |        |                                |                                                  |
| `FIRE_BULLET`              | C -> S | `v1.FireBulletRequest`         | 玩家請求開火                                     |
| `SWITCH_CANNON`            | C -> S | `v1.SwitchCannonRequest`       | 玩家請求切換砲台                                 |
| `JOIN_ROOM`                | C -> S | `v1.JoinRoomRequest`           | 玩家請求加入房間                                 |
| `LEAVE_ROOM`               | C -> S | `v1.LeaveRoomRequest`          | 玩家請求離開房間                                 |
| `HEARTBEAT`                | C -> S | `v1.HeartbeatMessage`          | 客戶端發送心跳以保持連接                         |
| `GET_ROOM_LIST`            | C -> S | `v1.GetRoomListRequest`        | 請求獲取當前可用的房間列表                       |
| `GET_PLAYER_INFO`          | C -> S | `v1.GetPlayerInfoRequest`      | 請求獲取當前玩家的詳細信息                       |
| **伺服器回應**             |        |                                |                                                  |
| `FIRE_BULLET_RESPONSE`     | S -> C | `v1.FireBulletResponse`        | 對開火請求的回應 (成功、子彈 ID、花費)           |
| `SWITCH_CANNON_RESPONSE`   | S -> C | `v1.SwitchCannonResponse`      | 對切換砲台請求的回應                             |
| `JOIN_ROOM_RESPONSE`       | S -> C | `v1.JoinRoomResponse`          | 對加入房間請求的回應                             |
| `LEAVE_ROOM_RESPONSE`      | S -> C | `v1.LeaveRoomResponse`         | 對離開房間請求的回應                             |
| `HEARTBEAT_RESPONSE`       | S -> C | `v1.HeartbeatResponse`         | 對心跳請求的回應                                 |
| `ROOM_LIST_RESPONSE`       | S -> C | `v1.RoomListResponse`          | 回應房間列表                                     |
| `PLAYER_INFO_RESPONSE`     | S -> C | `v1.PlayerInfoResponse`        | 回應玩家詳細信息                                 |
| **伺服器廣播事件**         |        |                                |                                                  |
| `BULLET_FIRED`             | S -> C | `v1.BulletFiredEvent`          | 廣播房間內有玩家開火                             |
| `CANNON_SWITCHED`          | S -> C | `v1.CannonSwitchedEvent`       | 廣播房間內有玩家切換砲台                         |
| `FISH_SPAWNED`             | S -> C | `v1.FishSpawnedEvent`          | 廣播場景中生成了新的魚群                         |
| `FISH_DIED`                | S -> C | `v1.FishDiedEvent`             | 廣播有魚被捕獲 (包含獎勵信息)                    |
| `PLAYER_REWARD`            | S -> C | `v1.PlayerRewardEvent`         | 廣播玩家獲得獎勵 (可用於非捕魚獎勵)              |
| `WELCOME`                  | S -> C | `v1.WelcomeMessage`            | 玩家成功連接後，伺服器發送的第一條歡迎消息       |
| `PLAYER_JOINED`            | S -> C | `v1.PlayerJoinedMessage`       | 廣播有新玩家加入房間                             |
| `PLAYER_LEFT`              | S -> C | `v1.PlayerLeftMessage`         | 廣播有玩家離開房間                               |
| **錯誤**                   |        |                                |                                                  |
| `ERROR`                    | S -> C | `v1.ErrorMessage`              | 當發生錯誤時，伺服器向客戶端發送錯誤信息         |
