# 遊客模式使用指南

## 概述

遊客模式允許玩家無需註冊即可快速進入遊戲。遊客帳號會自動生成一個唯一的暱稱，並獲得一個 JWT token 用於後續的遊戲連接。

## 功能特點

- ✅ 快速進入遊戲，無需註冊
- ✅ 自動生成唯一的遊客暱稱（格式：`Guest_<timestamp>`）
- ✅ JWT token 認證支持
- ✅ 與正式用戶享有相同的遊戲功能
- ✅ 向後兼容舊的 `player_id` 參數

## API 使用說明

### 1. 遊客登入

**端點**: `POST /guest-login`

**請求範例**:
```bash
curl -X POST http://localhost:9090/guest-login
```

**響應範例**:
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Guest login successful"
}
```

### 2. 使用 Token 連接 WebSocket

獲取 token 後，可以通過以下兩種方式連接 WebSocket：

#### 方式 1: 使用查詢參數
```javascript
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...";
const ws = new WebSocket(`ws://localhost:9090/ws?token=${token}&room_id=101`);
```

#### 方式 2: 使用 Authorization Header
```javascript
const ws = new WebSocket('ws://localhost:9090/ws?room_id=101', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});
```

### 3. 向後兼容（舊方式）

仍然支持直接使用 `player_id` 參數連接（不推薦用於新實現）：
```javascript
const ws = new WebSocket('ws://localhost:9090/ws?player_id=test_player&room_id=101');
```

## 完整流程示例

```javascript
// 1. 遊客登入
async function guestLogin() {
  const response = await fetch('http://localhost:9090/guest-login', {
    method: 'POST'
  });
  const data = await response.json();
  return data.token;
}

// 2. 連接 WebSocket
async function connectToGame() {
  const token = await guestLogin();
  const ws = new WebSocket(`ws://localhost:9090/ws?token=${token}&room_id=101`);

  ws.onopen = () => {
    console.log('已連接到遊戲服務器');
  };

  ws.onmessage = (event) => {
    // 處理遊戲消息
    console.log('收到消息:', event.data);
  };

  ws.onerror = (error) => {
    console.error('WebSocket 錯誤:', error);
  };

  return ws;
}

// 使用
connectToGame();
```

## 技術實現細節

### 數據庫結構

遊客用戶存儲在 `users` 表中：
- `is_guest`: `true`
- `username`: `NULL`（遊客沒有用戶名）
- `password_hash`: `NULL`（遊客沒有密碼）
- `nickname`: 自動生成的暱稱（例如 `Guest_4294967296`）
- `coins`: 初始金幣（預設 1000）

### JWT Token 結構

Token 包含以下 claims：
```json
{
  "user_id": 12345,
  "is_guest": true,
  "iss": "fish_server",
  "exp": 1234567890,
  "nbf": 1234567890,
  "iat": 1234567890
}
```

### 認證流程

1. 客戶端調用 `/guest-login` 端點
2. 服務器創建一個新的遊客用戶記錄
3. 生成包含 `is_guest: true` 的 JWT token
4. 返回 token 給客戶端
5. 客戶端使用 token 連接 WebSocket
6. 服務器驗證 token，提取用戶信息
7. 使用用戶的 nickname 創建或獲取玩家記錄
8. 建立 WebSocket 連接，開始遊戲

## 限制與注意事項

1. **數據持久性**: 遊客數據會保存在數據庫中，但遊客如果遺失 token 將無法恢復帳號
2. **Token 過期**: JWT token 有過期時間（由配置文件 `jwt.expire` 設定）
3. **轉換為正式用戶**: 目前遊客無法直接轉換為正式用戶（需要額外實現）

## 安全考慮

- Token 使用 HS256 算法簽名
- Token 包含過期時間，防止長期濫用
- 遊客帳號與正式帳號在數據庫層面使用相同的結構，便於未來升級
- WebSocket 連接支持 CORS，但在生產環境中應該限制來源

## 測試

可以使用以下命令測試遊客登入功能：

```bash
# 測試遊客登入
curl -X POST http://localhost:9090/guest-login

# 測試健康檢查
curl http://localhost:9090/health

# 測試狀態端點
curl http://localhost:9090/status
```

## 未來改進

- [ ] 支持遊客帳號升級為正式帳號
- [ ] 實現遊客帳號自動清理機制（清理長期未使用的遊客）
- [ ] 添加遊客帳號使用限制（例如金幣上限）
- [ ] 支持遊客帳號綁定第三方登入
