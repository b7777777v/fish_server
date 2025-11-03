# 🐟 Fish Server - 多人捕魚遊戲後端服務

這是一個功能完整的、為多人捕魚遊戲設計的後端專案。它採用了現代化的 Go 語言、微服務架構和雲原生技術，旨在提供一個高效、可擴展且易於維護的後端解決方案。

此專案不僅包含了完整的業務邏輯，還提供了一套極其完善的本地開發環境 (`.vscode`)，讓開發者可以實現一鍵啟動、多環境偵錯和自動化任務，極大地提升了開發效率。

## ✨ 主要功能

- **即時多人遊戲**：使用 WebSocket 實現低延遲的即時客戶端/伺服器通訊。
- **遊戲房管理**：創建、加入、離開遊戲房間。
- **核心遊戲邏輯**：包括魚群生成、捕獲機率、分數計算等。
- **玩家系統**：玩家資料管理與認證。
- **錢包系統**：管理玩家的遊戲內貨幣。
- **後台管理**：提供一個獨立的 `admin` 服務，用於遊戲管理與監控。
- **多環境支援**：完整的 `DEV`, `STAGING`, `PROD` 環境隔離與配置。

## 🏛️ 架構概覽

專案採用了清晰的 **Clean Architecture** 設計理念，將業務邏輯與基礎設施分離。

- **微服務架構**：
  - `game-server`: 處理核心遊戲邏輯與玩家互動 (WebSocket)。
  - `admin-server`: 提供 RESTful API 用於後台管理。
- **通訊**：
  - 客戶端與 `game-server` 之間使用 **WebSocket**。
  - 服務內部或對外的 API 使用 **gRPC** 和 **RESTful (Gin)**。
- **資料庫**：
  - **PostgreSQL**: 用於持久化儲存核心業務資料 (玩家、錢包等)。
  - **Redis**: 用於快取、會話管理或發布/訂閱。
- **依賴注入**：使用 **Google Wire** 管理服務間的依賴關係，實現松耦合。
- **容器化**：所有服務都被設計為在 **Docker** 中運行。

## 🛠️ 技術棧

- **語言**: Go
- **Web 框架**: Gin (RESTful API)
- **RPC 框架**: gRPC
- **資料庫**: PostgreSQL, Redis
- **ORM/資料庫工具**: `database/sql`, `pgx`
- **依賴注入**: Google Wire
- **容器化**: Docker, Docker Compose
- **資料庫遷移**: golang-migrate
- **日誌**: `log/slog` (結構化日誌)
- **配置管理**: Viper

## 🚀 快速開始

### 1. 環境準備

在開始之前，請確保您已安裝以下工具：

- **Go**: 1.21 或更高版本
- **Docker** 和 **Docker Compose**
- **Make**
- **golang-migrate**: [安裝指南](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- **golangci-lint**: [安裝指南](https://golangci-lint.run/usage/install/)

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
├── api/                # Protobuf 定義
├── cmd/                # 主應用程式入口 (main.go)
│   ├── admin/          # 後台管理服務
│   └── game/           # 遊戲服務
├── configs/            # 環境配置文件 (config.yaml)
├── deployments/        # Docker 和部署相關文件
├── internal/           # 專案內部代碼 (不對外暴露)
│   ├── app/            # 應用層: 服務的啟動與組織
│   ├── biz/            # 業務邏輯層: 核心業務實體與用例
│   ├── conf/           # 配置映射結構
│   ├── data/           # 資料存取層: Repository 實現
│   └── pkg/            # 專案內部共享的工具包
├── pkg/                # 可供外部專案使用的共享代碼
├── scripts/            # 各類輔助腳本
└── storage/            # 資料庫遷移文件
```

## 🤝 貢獻
