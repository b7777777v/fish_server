# 遊戲功能實現狀態報告

**日期**: 2025-11-17
**任務**: 實現遊戲核心功能（開火扣錢、擊殺贏錢、離開遊戲、遊戲紀錄）

---

## 📋 功能實現總覽

| 功能 | 業務層 (Usecase) | 處理層 (Handler) | Protobuf 定義 | 前端實現 | 狀態 |
|------|-----------------|-----------------|--------------|---------|------|
| 開火扣錢 | ✅ | ✅ | ✅ | ✅ | **完整** |
| 擊殺贏錢 | ✅ | ✅ | ✅ | ❌ | **後端完成** |
| 離開遊戲 | ✅ | ✅ | ✅ | ✅ | **完整** |
| 遊戲紀錄 | ✅ | ✅ | ✅ | N/A | **完整** |

---

## ✅ 已完成功能

### 1. 開火扣錢 ✅

**業務邏輯** (`internal/biz/game/usecase.go:218-283`)
```go
func (gu *GameUsecase) FireBullet(ctx context.Context, roomID string, playerID int64,
    direction float64, power int32, position Position) (*Bullet, error)
```
- ✅ 發射子彈
- ✅ 扣除玩家餘額 (Line 233-237)
- ✅ 創建錢包交易記錄 (Line 240-259)
- ✅ 記錄遊戲事件 (Line 263-277)

**處理層** (`internal/app/game/message_handler.go:72-150`)
```go
func (mh *MessageHandler) handleFireBullet(client *Client, message *pb.GameMessage)
```
- ✅ 接收 `FIRE_BULLET` 請求
- ✅ 驗證參數（威力 1-100）
- ✅ 調用業務邏輯
- ✅ 發送響應給客戶端
- ✅ 廣播 `BULLET_FIRED` 事件給房間

**前端** (`js/game-client.js:591-637`)
- ✅ 發送 FIRE_BULLET 請求
- ✅ 接收 FIRE_BULLET_RESPONSE
- ✅ 日誌記錄：`💥 成功開火！子彈ID: xxx, 消耗: xxx`

---

### 2. 離開遊戲 ✅

**業務邏輯** (`internal/biz/game/usecase.go:169-193`)
```go
func (gu *GameUsecase) LeaveRoom(ctx context.Context, roomID string, playerID int64) error
```
- ✅ 從房間管理器移除玩家
- ✅ 更新玩家狀態為 Idle (Line 177-179)
- ✅ 記錄離開事件 (Line 182-189)

**處理層** (`internal/app/game/message_handler.go:287-331`)
```go
func (mh *MessageHandler) handleLeaveRoom(client *Client, message *pb.GameMessage)
```
- ✅ 接收 `LEAVE_ROOM` 請求
- ✅ 調用業務邏輯
- ✅ 通知 Hub (Line 306-309)
- ✅ 清除客戶端房間ID
- ✅ 發送 LEAVE_ROOM_RESPONSE

**前端** (`js/game-client.js:640-645`)
- ✅ 發送 LEAVE_ROOM 請求
- ✅ 清除房間狀態

---

### 3. 遊戲紀錄 ✅

**業務邏輯** (`internal/biz/game/usecase.go`)
所有關鍵操作都會自動記錄事件：
- ✅ 創建房間 → `EventFishSpawn` (Line 114-122)
- ✅ 玩家加入 → `EventPlayerJoin` (Line 155-163)
- ✅ 玩家離開 → `EventPlayerLeave` (Line 182-189)
- ✅ 開火 → `EventBulletFire` (Line 264-277)
- ✅ 擊中魚 → `EventBulletHit` (Line 343-358)
- ✅ 魚死亡 → `EventFishDie` (Line 362-373)

**查詢介面**
```go
// 獲取遊戲事件記錄
func (gu *GameUsecase) GetGameEvents(ctx context.Context, roomID string, limit int) ([]*GameEvent, error)

// 獲取玩家統計
func (gu *GameUsecase) GetPlayerStatistics(ctx context.Context, playerID int64) (*GameStatistics, error)
```

---

## 🚧 部分完成功能

### 4. 擊殺贏錢 🟡 (後端完成，前端待實現)

**業務邏輯** (`internal/biz/game/usecase.go:285-379`) ✅
```go
func (gu *GameUsecase) HitFish(ctx context.Context, roomID string,
    bulletID int64, fishID int64) (*HitResult, error)
```
- ✅ 處理子彈命中
- ✅ 計算傷害和獎勵
- ✅ 更新玩家餘額 (Line 312-314)
- ✅ 創建錢包交易記錄 (Line 317-339)
- ✅ 記錄命中事件 (Line 343-358)
- ✅ 記錄魚死亡事件 (Line 361-373)

**處理層** (`internal/app/game/message_handler.go:333-418`) ✅ **本次新增**
```go
func (mh *MessageHandler) handleHitFish(client *Client, message *pb.GameMessage)
```
- ✅ 接收 `HIT_FISH` 請求
- ✅ 驗證參數
- ✅ 調用業務邏輯
- ✅ 發送 HIT_FISH_RESPONSE
- ✅ 廣播 FISH_DIED 事件
- ✅ 廣播 PLAYER_REWARD 事件

**Protobuf 定義** (`api/proto/v1/game.proto`) ✅ **本次新增**
```protobuf
// 消息類型
HIT_FISH = 9;
HIT_FISH_RESPONSE = 18;

// 請求消息
message HitFishRequest {
  int64 bullet_id = 1;
  int64 fish_id = 2;
}

// 響應消息
message HitFishResponse {
  bool success = 1;
  int64 bullet_id = 2;
  int64 fish_id = 3;
  int32 damage = 4;
  int64 reward = 5;
  bool is_killed = 6;
  bool is_critical = 7;
  double multiplier = 8;
  int64 timestamp = 9;
}
```

**前端實現** ❌ **待實現**
需要添加：
1. 碰撞檢測邏輯
2. 發送 HIT_FISH 請求
3. 處理 HIT_FISH_RESPONSE
4. 處理 FISH_DIED 和 PLAYER_REWARD 廣播

---

## 📝 本次修改文件清單

### 1. Protobuf 定義
**文件**: `api/proto/v1/game.proto`

添加內容：
- `HIT_FISH = 9` 消息類型
- `HIT_FISH_RESPONSE = 18` 消息類型
- `HitFishRequest` 消息定義
- `HitFishResponse` 消息定義
- GameMessage oneof 中添加相應字段

### 2. 後端處理器
**文件**: `internal/app/game/message_handler.go`

添加內容：
- HandleMessage switch 中添加 `HIT_FISH` case (Line 60-61)
- 新函數 `handleHitFish()` (Line 333-418)

實現細節：
```go
// 1. 接收並驗證請求
hitData := message.GetHitFish()
if hitData.GetBulletId() <= 0 || hitData.GetFishId() <= 0 {
    mh.sendErrorResponse(client, "Invalid bullet or fish ID")
    return
}

// 2. 調用業務邏輯
hitResult, err := mh.gameUsecase.HitFish(ctx, client.RoomID,
    hitData.GetBulletId(), hitData.GetFishId())

// 3. 發送響應
response := &pb.GameMessage{
    Type: pb.MessageType_HIT_FISH_RESPONSE,
    Data: &pb.GameMessage_HitFishResponse{ ... }
}

// 4. 如果擊殺，廣播事件
if hitResult.Reward > 0 {
    // 廣播 FISH_DIED
    // 廣播 PLAYER_REWARD
}
```

---

## ⚠️ 待辦事項

### 1. 安裝 Protobuf 編譯器 ⚡ **優先**

當前狀態：
```bash
$ make proto
 protoc is not installed. Please install protobuf compiler.
```

**解決方案**：
```bash
# Ubuntu/Debian
sudo apt-get install -y protobuf-compiler

# macOS
brew install protobuf

# 或者手動下載
# https://grpc.io/docs/protoc-installation/
```

**安裝後執行**：
```bash
# 生成 Go 代碼
make proto
# 或
sh ./scripts/proto-gen.sh

# 這會生成：
# - pkg/pb/v1/game.pb.go (Go protobuf)
# - js/generated/proto/v1/game_pb.js (JavaScript protobuf)
```

---

### 2. 前端碰撞檢測 ⚡ **優先**

需要在前端添加碰撞檢測邏輯。有兩種實現方案：

#### 方案 A：客戶端碰撞檢測（簡單，但可能被作弊）

**文件**: `js/game-client.js` 或 `js/game-renderer.js`

```javascript
// 在遊戲循環中檢測碰撞
function checkCollisions() {
    if (!gameRenderer || !gameRenderer.gameState) return;

    const bullets = gameRenderer.gameState.bullets || [];
    const fishes = gameRenderer.gameState.fishes || [];

    bullets.forEach(bullet => {
        fishes.forEach(fish => {
            if (isColliding(bullet, fish)) {
                // 發送 HIT_FISH 請求
                sendHitFishMessage(bullet.bulletId, fish.fishId);
            }
        });
    });
}

function isColliding(bullet, fish) {
    const distance = Math.sqrt(
        Math.pow(bullet.position.x - fish.position.x, 2) +
        Math.pow(bullet.position.y - fish.position.y, 2)
    );
    return distance < (fish.radius || 30); // 碰撞半徑
}

function sendHitFishMessage(bulletId, fishId) {
    const gameMessage = new proto.v1.GameMessage();
    gameMessage.setType(MessageType.HIT_FISH);
    const hitFishReq = new proto.v1.HitFishRequest();
    hitFishReq.setBulletId(bulletId);
    hitFishReq.setFishId(fishId);
    gameMessage.setHitFish(hitFishReq);
    sendMessage(gameMessage);
}

// 處理響應
case MessageType.HIT_FISH_RESPONSE:
    const hitFishResp = gameMessage.getHitFishResponse();
    if (hitFishResp.getSuccess()) {
        if (hitFishResp.getIsKilled()) {
            log(`🎯 擊殺！獲得獎勵: ${hitFishResp.getReward()}`);
        } else {
            log(`💥 命中！造成傷害: ${hitFishResp.getDamage()}`);
        }
    }
    break;

case MessageType.FISH_DIED:
    const fishDied = gameMessage.getFishDied();
    log(`🐟 魚死亡！玩家 ${fishDied.getPlayerId()} 獲得 ${fishDied.getReward()}`);
    // 更新UI，移除魚
    break;

case MessageType.PLAYER_REWARD:
    const reward = gameMessage.getPlayerReward();
    log(`💰 玩家 ${reward.getPlayerId()} 獲得獎勵 ${reward.getReward()}`);
    // 更新玩家餘額顯示
    break;
```

#### 方案 B：伺服器端碰撞檢測（安全，但需要更多工作）

在後端 RoomManager 或 GameUsecase 中添加定期碰撞檢測，自動處理擊殺。客戶端只負責渲染。

**優點**：
- 防止作弊
- 所有玩家看到一致的結果

**缺點**：
- 需要在後端實現物理碰撞檢測
- 可能有輕微延遲

---

### 3. 前端UI改進 🔧 **可選**

添加遊戲統計顯示：
```html
<!-- 在 index.html 添加 -->
<div id="gameStats">
    <h3>遊戲統計</h3>
    <p>總開火次數: <span id="totalShots">0</span></p>
    <p>總擊殺數: <span id="totalKills">0</span></p>
    <p>總獎勵: <span id="totalReward">0</span></p>
    <p>當前餘額: <span id="currentBalance">10000</span></p>
</div>
```

更新統計：
```javascript
let gameStats = {
    totalShots: 0,
    totalKills: 0,
    totalReward: 0,
    currentBalance: 10000
};

function updateGameStats(type, value) {
    switch(type) {
        case 'shot':
            gameStats.totalShots++;
            break;
        case 'kill':
            gameStats.totalKills++;
            gameStats.totalReward += value;
            break;
        case 'balance':
            gameStats.currentBalance = value;
            break;
    }

    // 更新 DOM
    document.getElementById('totalShots').textContent = gameStats.totalShots;
    document.getElementById('totalKills').textContent = gameStats.totalKills;
    document.getElementById('totalReward').textContent = gameStats.totalReward;
    document.getElementById('currentBalance').textContent = gameStats.currentBalance;
}
```

---

## 🎯 實現優先級

### 高優先級（必須完成）
1. ⚡ **安裝 protoc 並重新生成代碼**
   - 否則後端無法編譯

2. ⚡ **實現前端碰撞檢測**
   - 建議先用方案 A（客戶端檢測）快速驗證
   - 後續可升級到方案 B（伺服器檢測）

### 中優先級（建議完成）
3. 🔧 **添加前端UI統計**
   - 提升用戶體驗
   - 便於調試和測試

4. 🔧 **處理邊界情況**
   - 子彈已消失
   - 魚已死亡
   - 玩家已離開房間

### 低優先級（可選）
5. 📊 **遊戲統計查詢介面**
   - 實現查詢歷史記錄的前端頁面
   - 調用 `GetGameEvents` 和 `GetPlayerStatistics`

6. 🎨 **視覺特效**
   - 擊中特效
   - 擊殺動畫
   - 獎勵彈出

---

## 🧪 測試計劃

### 1. 單元測試
```bash
# 測試業務邏輯
go test ./internal/biz/game/... -v

# 測試 Handler
go test ./internal/app/game/... -v
```

### 2. 集成測試流程

**前置條件**：
1. ✅ 啟動資料庫
   ```bash
   docker-compose -f deployments/docker-compose.dev.yml up postgres redis -d
   ```

2. ✅ 執行遷移
   ```bash
   go run cmd/migrator/main.go up
   ```

3. ✅ 生成 Protobuf（需先安裝 protoc）
   ```bash
   make proto
   ```

4. ✅ 啟動 Game Server
   ```bash
   ENVIRONMENT=dev go run ./cmd/game/...
   ```

**測試步驟**：
1. 開啟 `js/index.html` 在瀏覽器
2. 點擊「遊客登入並開始遊戲」
3. 選擇座位
4. 測試開火 → 檢查日誌是否顯示扣錢
5. 測試擊殺 → 檢查是否獲得獎勵
6. 測試離開 → 檢查是否正常退出
7. 檢查遊戲記錄 → 查詢資料庫 `game_events` 表

---

## 📊 資料庫檢查

### 查看遊戲事件
```sql
SELECT * FROM game_events
ORDER BY timestamp DESC
LIMIT 20;
```

### 查看錢包交易
```sql
SELECT * FROM transactions
WHERE transaction_type IN ('game_bullet_cost', 'game_fish_reward')
ORDER BY created_at DESC
LIMIT 20;
```

### 查看玩家餘額變化
```sql
SELECT u.username, w.balance, w.updated_at
FROM wallets w
JOIN users u ON w.user_id = u.id
ORDER BY w.updated_at DESC;
```

---

## 📝 總結

### ✅ 已完成（本次工作）
1. ✅ 分析了當前功能實現狀態
2. ✅ 在 Protobuf 中添加 HIT_FISH 消息定義
3. ✅ 實現了 handleHitFish 處理器
4. ✅ 完善了擊殺贏錢的後端邏輯
5. ✅ 添加了 FISH_DIED 和 PLAYER_REWARD 廣播

### 🚧 待完成
1. ⚡ 安裝 protoc 並重新生成代碼
2. ⚡ 實現前端碰撞檢測和 HIT_FISH 發送
3. 🔧 添加前端UI統計
4. 🧪 完整測試所有流程

### 📌 其他發現
- ✅ 開火扣錢功能已經完整實現
- ✅ 離開遊戲功能已經完整實現
- ✅ 遊戲紀錄功能已經完整實現
- 🟡 擊殺贏錢功能後端完成，等待前端實現

---

**文檔版本**: 1.0
**完成日期**: 2025-11-17
**維護者**: Claude Code
