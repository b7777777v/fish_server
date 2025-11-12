# 遊客模式完整實現文檔

## 📋 概述

遊客模式允許玩家無需註冊即可快速進入遊戲。此功能已經在後端和前端完全實現。

## ✅ 已完成的功能

### 後端實現

1. **遊客登入 API** (`POST /guest-login`)
   - 自動創建遊客用戶記錄
   - 生成唯一的遊客暱稱（格式：`Guest_<timestamp>`）
   - 返回 JWT token（包含 `is_guest: true`）
   - 位置：`internal/app/game/app.go:224`

2. **JWT Token 認證**
   - Token 服務支持 `isGuest` claim
   - Token 包含用戶 ID 和遊客標識
   - 位置：`internal/pkg/token/token.go`

3. **WebSocket Token 驗證**
   - 支持通過查詢參數傳遞 token (`?token=...`)
   - 支持通過 Authorization header 傳遞 token
   - 自動驗證並提取用戶信息
   - 向後兼容 `player_id` 參數
   - 位置：`internal/app/game/websocket.go:156`

4. **數據庫整合**
   - 遊客用戶存儲在 `users` 表
   - `is_guest` 字段標記遊客帳號
   - `username` 和 `password_hash` 可為 NULL
   - 位置：`internal/data/postgres/account.go`

5. **依賴注入配置**
   - AccountUsecase 整合到 GameApp
   - WebSocketHandler 添加 token 支持
   - Wire 配置已更新
   - 位置：`cmd/game/wire_gen.go`

### 前端實現

1. **遊客登入 UI**
   - 醒目的藍色遊客模式區塊
   - 一鍵登入按鈕
   - 遊客暱稱顯示
   - 位置：`js/index.html:100-126`

2. **Token 認證邏輯**
   - `guestLogin()` 函數調用後端 API
   - `parseJWT()` 解析 token 獲取用戶信息
   - `connectWithToken()` 使用 token 連接 WebSocket
   - 位置：`js/game-client.js:128-231`

3. **遊戲整合**
   - 遊戲渲染器支持遊客模式
   - 自動顯示遊客暱稱
   - 所有遊戲功能可用
   - 位置：`js/game-client.js:257-260`

4. **狀態管理**
   - `isGuestMode` 標記
   - `authToken` 存儲
   - 按鈕狀態自動控制
   - 位置：`js/game-client.js:73-74`

## 🚀 使用方式

### 用戶角度

1. **啟動遊戲**
   ```bash
   # 啟動後端服務器
   ./bin/game
   ```

2. **打開前端**
   - 在瀏覽器中打開 `js/index.html`

3. **遊客登入**
   - 點擊"🚀 遊客登入並開始遊戲"按鈕
   - 等待登入完成
   - 自動連接到遊戲服務器
   - 開始遊戲！

### 開發者角度

#### 後端 API 調用

```bash
# 遊客登入
curl -X POST http://localhost:9090/guest-login

# 響應
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Guest login successful"
}
```

#### WebSocket 連接

```javascript
// 使用 token 連接
const ws = new WebSocket(`ws://localhost:9090/ws?token=${token}`);

// 或使用 Authorization header（需要支持）
const ws = new WebSocket('ws://localhost:9090/ws', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

#### 前端代碼示例

```javascript
// 遊客登入
async function guestLogin() {
  const response = await fetch('http://localhost:9090/guest-login', {
    method: 'POST'
  });
  const data = await response.json();
  const token = data.token;

  // 連接 WebSocket
  const ws = new WebSocket(`ws://localhost:9090/ws?token=${token}`);
}
```

## 📁 文件更改清單

### 後端文件

| 文件路徑 | 類型 | 說明 |
|---------|------|------|
| `internal/app/game/app.go` | 修改 | 添加 AccountUsecase，添加 /guest-login 端點 |
| `internal/app/game/websocket.go` | 修改 | 添加 token 認證支持 |
| `internal/biz/account/usecase.go` | 修改 | 改進 generateGuestID() 函數 |
| `cmd/game/wire_gen.go` | 修改 | 更新依賴注入配置 |
| `GUEST_MODE.md` | 新增 | 後端遊客模式文檔 |

### 前端文件

| 文件路徑 | 類型 | 說明 |
|---------|------|------|
| `js/index.html` | 修改 | 添加遊客模式 UI |
| `js/game-client.js` | 修改 | 添加遊客登入邏輯和 token 認證 |
| `js/FRONTEND_GUEST_MODE.md` | 新增 | 前端遊客模式文檔 |

### 文檔文件

| 文件路徑 | 類型 | 說明 |
|---------|------|------|
| `GUEST_MODE_COMPLETE.md` | 新增 | 完整實現總結文檔 |

## 🔍 技術細節

### JWT Token 結構

```json
{
  "user_id": 1731423456789,
  "is_guest": true,
  "iss": "fish_server",
  "exp": 1731510000,
  "nbf": 1731423400,
  "iat": 1731423400
}
```

### 數據庫記錄

```sql
-- 遊客用戶記錄示例
INSERT INTO users (
  username,        -- NULL
  password_hash,   -- NULL
  nickname,        -- 'Guest_1731423456789'
  is_guest,        -- true
  coins            -- 1000 (預設)
) VALUES (NULL, NULL, 'Guest_1731423456789', true, 1000);
```

### WebSocket 連接流程

```
客戶端                     服務器
  |                          |
  |--1. HTTP POST /guest-login-->|
  |                          |
  |<--2. 返回 JWT token -------|
  |                          |
  |--3. WS ?token=xxx ------->|
  |                          |
  |                    4. 驗證 token
  |                    5. 提取 user_id
  |                    6. 獲取用戶信息
  |                    7. 創建/獲取玩家
  |                          |
  |<--8. WebSocket 建立 ------|
  |                          |
  |<===== 遊戲通信 =========>|
```

## 🎯 測試方式

### 功能測試

1. **遊客登入測試**
   ```bash
   curl -X POST http://localhost:9090/guest-login
   ```
   - 預期：返回 token 和 success: true

2. **Token 連接測試**
   ```bash
   # 使用 websocat 或其他 WebSocket 客戶端
   websocat "ws://localhost:9090/ws?token=YOUR_TOKEN_HERE"
   ```
   - 預期：成功建立連接

3. **前端完整流程測試**
   - 打開 `js/index.html`
   - 點擊"遊客登入"按鈕
   - 觀察日誌和連接狀態
   - 預期：自動登入並進入遊戲

### 回歸測試

1. **傳統登入方式**
   - 使用玩家 ID 輸入框
   - 點擊"連接到伺服器"
   - 預期：正常連接

2. **API 健康檢查**
   ```bash
   curl http://localhost:9090/health
   curl http://localhost:9090/status
   ```

## 🔒 安全考慮

1. **Token 安全**
   - 使用 HS256 算法簽名
   - 包含過期時間
   - 僅在內存中存儲（前端）

2. **數據庫安全**
   - 遊客帳號與正式帳號分離標記
   - 使用相同的安全約束

3. **WebSocket 安全**
   - Token 驗證在建立連接時進行
   - 無效 token 會拒絕連接
   - 支持 CORS 配置

## 🚧 已知限制

1. **Token 持久性**
   - 前端刷新會丟失 token
   - 需要重新登入

2. **帳號升級**
   - 目前無法將遊客帳號升級為正式帳號
   - 需要額外實現

3. **帳號清理**
   - 沒有自動清理長期未使用的遊客帳號
   - 可能需要後台任務

## 📝 未來改進

- [ ] 支持將遊客帳號綁定到正式帳號
- [ ] 添加遊客帳號自動清理機制
- [ ] 實現 token 刷新功能
- [ ] 添加遊客帳號使用限制（金幣上限等）
- [ ] 支持 sessionStorage 存儲 token（可選）
- [ ] 添加遊客模式的使用統計

## 🎉 總結

遊客模式已經完全實現並整合到系統中：

✅ **後端**：完整的 API、JWT 認證、WebSocket 支持
✅ **前端**：一鍵登入、自動連接、完整遊戲體驗
✅ **向後兼容**：不影響現有登入方式
✅ **文檔完善**：包含使用說明和技術細節

玩家現在可以無障礙地快速開始遊戲，大大降低了進入門檻！🎮
