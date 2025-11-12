# 座位選擇流程實現文檔

## ⚠️ 當前狀態

**座位選擇功能已設計和編碼完成，但需要 protobuf 代碼重新生成才能啟用。**

由於開發環境無法安裝 protoc 編譯器，後端座位選擇代碼已暫時注釋（帶有 TODO 標記）。

**啟用步驟：**
1. 安裝 protoc 編譯器
2. 運行 `make proto` 重新生成 protobuf 代碼
3. 取消注釋 `internal/app/game/websocket.go` 和 `internal/app/game/room_manager.go` 中帶有 "TODO: Uncomment after running `make proto`" 標記的代碼
4. 重新編譯項目

## 概述

實現了進入房間後必須先選擇座位才能開火的流程，提升遊戲體驗和座位管理。

## 功能特點

- ✅ 玩家進入房間後必須先選擇座位
- ✅ 未選擇座位時無法開火
- ✅ 座位狀態實時顯示（可用/已佔用）
- ✅ 防止重複選擇已佔用的座位
- ✅ 支持 4 個座位（0-3）

## 後端實現

### 1. Protobuf 定義更新

**文件**: `api/proto/v1/game.proto`

添加了新的消息類型：

```protobuf
// 消息類型枚舉
enum MessageType {
  SELECT_SEAT = 8;              // 選擇座位請求
  SELECT_SEAT_RESPONSE = 17;    // 選擇座位響應
}

// 選擇座位請求
message SelectSeatRequest {
  int32 seat_id = 1;  // 座位ID (0-3)
}

// 選擇座位響應
message SelectSeatResponse {
  bool success = 1;
  int32 seat_id = 2;
  string message = 3;
  int64 timestamp = 4;
}

// 玩家信息響應（添加座位ID字段）
message PlayerInfoResponse {
  // ... 其他字段 ...
  int32 seat_id = 8;  // 當前座位ID，-1 表示未選擇
}
```

### 2. WebSocket 處理器

**文件**: `internal/app/game/websocket.go`

添加了座位選擇處理器：

```go
// handleSelectSeat 處理選擇座位請求
func (c *Client) handleSelectSeat(msg *pb.GameMessage) {
    if c.RoomID == "" {
        c.sendErrorPB("Not in any room")
        return
    }

    // 轉發到房間處理
    c.hub.gameAction <- &GameActionMessage{
        Client:    c,
        RoomID:    c.RoomID,
        Action:    "select_seat",
        Data:      msg,
        Timestamp: time.Now(),
    }
}
```

### 3. 房間管理器

**文件**: `internal/app/game/room_manager.go`

#### 座位選擇處理

```go
// handleSelectSeat 處理選擇座位操作
func (rm *RoomManager) handleSelectSeat(action *GameActionMessage) {
    client := action.Client

    // 檢查玩家是否在房間中
    playerInfo, exists := rm.gameState.Players[client.ID]
    if !exists {
        client.sendError("Player not in game")
        return
    }

    // 獲取選擇的座位 ID
    selectData := gameMsg.GetSelectSeat()
    requestedSeatID := selectData.SeatId

    // 驗證座位 ID 範圍 (0-3)
    if requestedSeatID < 0 || requestedSeatID > 3 {
        client.sendError("Invalid seat ID, must be between 0 and 3")
        return
    }

    // 檢查座位是否已被佔用
    for _, p := range rm.gameState.Players {
        if p.SeatID == int(requestedSeatID) && p.PlayerID != client.ID {
            client.sendError("Seat already taken")
            return
        }
    }

    // 分配座位
    playerInfo.SeatID = int(requestedSeatID)

    // 發送響應並廣播狀態更新
    // ...
}
```

#### 開火驗證

在 `handleFireBullet` 中添加座位檢查：

```go
// 檢查玩家是否已選擇座位
if playerInfo.SeatID == -1 {
    client.sendError("Please select a seat first")
    return
}
```

## 前端實現

### 1. UI 組件

**文件**: `js/index.html`

添加了座位選擇面板：

```html
<!-- 座位選擇面板 -->
<div id="seatSelectionPanel" style="display: none;">
    <h3>🪑 選擇座位</h3>
    <p>請選擇一個座位開始遊戲</p>
    <div style="display: grid; grid-template-columns: repeat(4, 1fr);">
        <button class="seat-btn" data-seat="0">座位 1</button>
        <button class="seat-btn" data-seat="1">座位 2</button>
        <button class="seat-btn" data-seat="2">座位 3</button>
        <button class="seat-btn" data-seat="3">座位 4</button>
    </div>
</div>
```

開火按鈕初始狀態為禁用：

```html
<button id="fireBulletBtn" disabled>🔫 開火</button>
```

### 2. 狀態管理

**文件**: `js/game-client.js`

添加座位狀態變量：

```javascript
// 座位選擇相關
let currentSeat = -1;         // 當前選擇的座位，-1 表示未選擇
let hasSelectedSeat = false;  // 是否已選擇座位
```

### 3. 座位選擇邏輯

```javascript
// 座位選擇函數
function selectSeat(seatId) {
    const gameMessage = new proto.v1.GameMessage();
    gameMessage.setType(MessageType.SELECT_SEAT);

    const selectSeatReq = new proto.v1.SelectSeatRequest();
    selectSeatReq.setSeatId(seatId);
    gameMessage.setSelectSeat(selectSeatReq);

    sendMessage(gameMessage);
    log(`正在選擇座位 ${seatId + 1}...`, 'system');
}

// 綁定座位按鈕事件
seatButtons.forEach(btn => {
    btn.addEventListener('click', () => {
        const seatId = parseInt(btn.dataset.seat);
        selectSeat(seatId);
    });
});
```

### 4. 響應處理

```javascript
// 處理選擇座位響應
case MessageType.SELECT_SEAT_RESPONSE:
    const selectResp = gameMessage.getSelectSeatResponse();
    if (selectResp.getSuccess()) {
        currentSeat = selectResp.getSeatId();
        hasSelectedSeat = true;

        // 啟用開火按鈕
        fireBulletBtn.disabled = false;

        // 隱藏警告，顯示提示
        fireWarning.style.display = 'none';
        fireTip.style.display = 'block';

        // 更新座位信息顯示
        currentSeatInfo.style.display = 'block';
        currentSeatId.textContent = `座位 ${currentSeat + 1}`;

        log(`座位選擇成功：座位 ${currentSeat + 1}`, 'system');
    } else {
        log(`座位選擇失敗：${selectResp.getMessage()}`, 'error');
    }
    break;
```

### 5. 連接流程

```javascript
socket.onopen = () => {
    // ... 連接成功處理 ...

    // 顯示座位選擇面板
    if (seatSelectionPanel) {
        seatSelectionPanel.style.display = 'block';
    }

    // 禁用開火按鈕直到選擇座位
    fireBulletBtn.disabled = true;
};
```

## 完整流程

```
1. 玩家登入（遊客模式或傳統模式）
   ↓
2. 連接 WebSocket
   ↓
3. 加入房間
   ↓
4. 顯示座位選擇面板
   ↓
5. 玩家選擇座位（座位 1-4）
   ↓
6. 發送 SELECT_SEAT 請求到服務器
   ↓
7. 服務器驗證座位可用性
   ↓
8. 返回 SELECT_SEAT_RESPONSE
   ↓
9. 啟用開火按鈕
   ↓
10. 玩家可以開始遊戲
```

## 驗證邏輯

### 後端驗證

1. **房間檢查**: 玩家必須在房間中
2. **座位範圍**: 座位 ID 必須在 0-3 之間
3. **座位可用性**: 座位不能已被其他玩家佔用
4. **開火驗證**: 開火時檢查 `SeatID != -1`

### 前端驗證

1. **按鈕狀態**: 未選座位時開火按鈕禁用
2. **視覺提示**: 顯示警告信息提醒選擇座位
3. **座位狀態**: 實時更新座位佔用情況
4. **防止重複**: 已選座位後更新 UI 狀態

## 錯誤處理

### 常見錯誤

1. **未加入房間**: `"Not in any room"`
2. **無效座位ID**: `"Invalid seat ID, must be between 0 and 3"`
3. **座位已佔用**: `"Seat already taken"`
4. **未選座位開火**: `"Please select a seat first"`

### 錯誤顯示

- 後端錯誤通過 `sendErrorPB()` 發送
- 前端在日誌中顯示紅色錯誤消息
- 提供友好的用戶提示

## 測試方式

### 手動測試

1. **正常流程測試**:
   ```
   - 遊客登入
   - 加入房間 101
   - 選擇座位 1
   - 嘗試開火 ✓ 應該成功
   ```

2. **未選座位測試**:
   ```
   - 遊客登入
   - 加入房間 101
   - 不選座位直接嘗試開火
   - ✓ 應該被阻止，顯示錯誤
   ```

3. **座位佔用測試**:
   ```
   - 玩家A選擇座位1
   - 玩家B嘗試選擇座位1
   - ✓ 應該被拒絕，顯示"Seat already taken"
   ```

## 待生成Protobuf

**重要**: 需要重新生成 protobuf 代碼才能編譯和運行座位選擇功能。

### 當前實現狀態

- ✅ Protobuf 定義已完成（`api/proto/v1/game.proto`）
- ✅ 後端處理邏輯已編寫（已暫時注釋）
- ✅ 前端 UI 已完成（`js/index.html`）
- ⏸️ 後端代碼已注釋，等待 protobuf 生成後啟用

### 生成 Protobuf 代碼

```bash
# 方法 1: 使用 Makefile
make proto

# 方法 2: 手動運行 protoc
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --js_out=import_style=browser,binary:js/generated \
       api/proto/v1/*.proto
```

### 啟用後端代碼

生成 protobuf 代碼後，在以下文件中取消注釋帶有 `TODO: Uncomment after running 'make proto'` 標記的代碼：

1. `internal/app/game/websocket.go:463-465` - SELECT_SEAT case statement
2. `internal/app/game/websocket.go:655-671` - handleSelectSeat function
3. `internal/app/game/room_manager.go:350-352` - "select_seat" case statement
4. `internal/app/game/room_manager.go:546-611` - handleSelectSeat function
5. `internal/app/game/room_manager.go:370-375` - Seat selection validation in handleFireBullet

## 未來改進

- [ ] 添加座位視覺化顯示（畫布上顯示玩家位置）
- [ ] 支持座位重新選擇
- [ ] 添加座位預覽功能
- [ ] 實現座位預留機制（斷線重連）
- [ ] 添加座位使用統計

## 相關文件

- `api/proto/v1/game.proto` - Protobuf 定義
- `internal/app/game/websocket.go` - WebSocket 處理
- `internal/app/game/room_manager.go` - 房間管理
- `js/index.html` - 前端 UI
- `js/game-client.js` - 前端邏輯
