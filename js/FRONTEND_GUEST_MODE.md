# 前端遊客模式說明

## 概述

前端已經整合遊客模式功能，允許玩家無需輸入任何信息即可快速開始遊戲。

## UI 更新

### 新增的 UI 元素

1. **遊客模式區塊（醒目的藍色區域）**
   - 位於頁面頂部控制區
   - 包含"🚀 遊客登入並開始遊戲"按鈕
   - 登入成功後顯示遊客暱稱

2. **傳統登入區塊（灰色區域）**
   - 保留原有的玩家ID輸入方式
   - 向後兼容舊的登入流程

3. **遊客信息顯示**
   - 登入成功後顯示遊客暱稱（例如：Guest_12345）
   - 淺藍色背景突顯遊客身份

## 功能流程

### 遊客登入流程

```
1. 點擊"遊客登入並開始遊戲"按鈕
   ↓
2. 前端調用 POST /guest-login API
   ↓
3. 後端創建遊客帳號並返回 JWT token
   ↓
4. 前端解析 token 獲取用戶信息
   ↓
5. 顯示遊客暱稱
   ↓
6. 自動使用 token 連接 WebSocket
   ↓
7. 開始遊戲
```

### 代碼實現

#### 1. 遊客登入函數

```javascript
async function guestLogin() {
    // 調用後端 API
    const response = await fetch(`${API_BASE_URL}/guest-login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        }
    });

    const data = await response.json();

    if (data.success && data.token) {
        authToken = data.token;
        isGuestMode = true;

        // 解析 token
        const tokenPayload = parseJWT(authToken);
        const nickname = `Guest_${tokenPayload.user_id}`;

        // 顯示遊客信息
        guestNickname.textContent = nickname;
        guestInfo.style.display = 'block';

        // 自動連接
        connectWithToken();
    }
}
```

#### 2. Token 連接函數

```javascript
function connectWithToken() {
    const url = `${WEBSOCKET_URL}?token=${encodeURIComponent(authToken)}`;
    socket = new WebSocket(url);
    socket.binaryType = "arraybuffer";
    setupWebSocketHandlers();
}
```

#### 3. JWT 解析（客戶端）

```javascript
function parseJWT(token) {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(function(c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
}
```

## 新增的變量

```javascript
const API_BASE_URL = 'http://localhost:9090';  // API 基礎 URL
let authToken = null;                          // JWT token
let isGuestMode = false;                       // 遊客模式標記
```

## DOM 元素

```javascript
const guestLoginBtn = document.getElementById('guestLoginBtn');      // 遊客登入按鈕
const guestInfo = document.getElementById('guestInfo');              // 遊客信息顯示區
const guestNickname = document.getElementById('guestNickname');      // 遊客暱稱顯示
```

## 狀態管理

### 遊客登入狀態

- **未登入**: 顯示"遊客登入並開始遊戲"按鈕
- **登入中**: 按鈕禁用，顯示"⏳ 正在登入..."
- **已連接**: 按鈕禁用，顯示遊客信息
- **已斷線**: 按鈕啟用，顯示"🔄 重新連接"

### 按鈕狀態控制

```javascript
// 連接時
guestLoginBtn.disabled = true;

// 斷線時
if (isGuestMode) {
    guestLoginBtn.disabled = false;
    guestLoginBtn.textContent = '🔄 重新連接';
}
```

## 遊戲渲染器整合

遊客模式完全整合到遊戲渲染器中：

```javascript
// 設置當前玩家 - 支持遊客模式
const currentPlayerId = isGuestMode
    ? (guestNickname ? guestNickname.textContent : 'Guest')
    : playerIdInput.value;
gameRenderer.setCurrentPlayer(currentPlayerId);
```

## 向後兼容

原有的玩家ID登入方式完全保留，不受影響：

1. 傳統 `player_id` 參數連接仍然可用
2. 遊客模式和傳統模式可以共存
3. UI 清晰區分兩種登入方式

## 測試方式

### 本地測試

1. 確保後端服務器運行在 `http://localhost:9090`
2. 在瀏覽器中打開 `js/index.html`
3. 點擊"遊客登入並開始遊戲"按鈕
4. 觀察控制台日誌和連接狀態

### 預期行為

1. 點擊按鈕後，按鈕變為"正在登入..."
2. 成功後顯示遊客暱稱（例如：Guest_1731423456789）
3. 自動連接到 WebSocket
4. 遊戲畫面和控制面板自動顯示
5. 可以正常進行遊戲操作

### 錯誤處理

- 網絡錯誤：顯示錯誤日誌，按鈕恢復可點擊狀態
- Token 解析失敗：顯示錯誤日誌，不進行連接
- WebSocket 連接失敗：按照正常的連接錯誤處理

## 日誌示例

成功的遊客登入日誌：

```
[14:30:15] 正在進行遊客登入...
[14:30:15] 遊客登入成功！暱稱: Guest_1731423456789
[14:30:15] 正在使用 token 連接到服務器...
[14:30:16] 成功連接到伺服器
```

## 安全注意事項

1. **Token 安全性**
   - Token 存儲在內存中（`authToken` 變量）
   - 頁面刷新後 token 會丟失
   - 不建議將 token 存儲在 localStorage（安全考慮）

2. **JWT 解析**
   - 客戶端僅解析 token 用於顯示
   - 不驗證簽名（由服務器驗證）
   - 僅提取 `user_id` 用於顯示暱稱

3. **HTTPS**
   - 生產環境應使用 HTTPS
   - WebSocket 應使用 WSS

## 未來改進

- [ ] 將 token 保存到 sessionStorage（可選）
- [ ] 添加 token 過期提示
- [ ] 支持遊客帳號轉換為正式帳號的 UI
- [ ] 添加遊客帳號的使用限制提示
- [ ] 美化遊客模式 UI
