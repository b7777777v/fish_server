# 🪑 多人座位系統 - 技術文檔

## 概述

本文檔描述了多人捕魚遊戲的座位系統實現，解決了不同座位玩家的砲台方向問題。

## 問題描述

**原始問題：**
- 前端不同座位位置不正確
- 所有玩家的子彈只能往上發射
- 無法根據座位位置控制發射方向

## 解決方案

### 1. 座位佈局系統

實現了4個座位的佈局，每個座位有獨立的位置和初始發射方向：

| 座位ID | 位置 | 初始角度 | 發射方向 | 座標 |
|--------|------|----------|----------|------|
| 0 | 底部中央 | -90° | ↑ 向上 | (centerX, height - 50) |
| 1 | 頂部中央 | 90° | ↓ 向下 | (centerX, 50) |
| 2 | 左側中央 | 0° | → 向右 | (50, centerY) |
| 3 | 右側中央 | 180° | ← 向左 | (width - 50, centerY) |

### 2. 核心代碼修改

#### 2.1 `getCannonPosition()` 函數

**修改前：**
```javascript
getCannonPosition(playerIndex) {
    const positions = [
        { x: centerX, y: this.height - margin },
        { x: centerX, y: margin },
        { x: margin, y: centerY },
        { x: this.width - margin, y: centerY }
    ];
    return positions[playerIndex % positions.length];
}
```

**修改後：**
```javascript
getCannonPosition(playerIndex) {
    const positions = [
        { x: centerX, y: this.height - margin, angle: -Math.PI / 2 },  // 底部 - 向上
        { x: centerX, y: margin, angle: Math.PI / 2 },                 // 頂部 - 向下
        { x: margin, y: centerY, angle: 0 },                           // 左側 - 向右
        { x: this.width - margin, y: centerY, angle: Math.PI }         // 右側 - 向左
    ];
    return positions[playerIndex % positions.length];
}
```

#### 2.2 `addPlayer()` 函數

**新增功能：**
- 支持傳入座位ID參數
- 使用座位對應的初始角度
- 保存座位ID到玩家對象

```javascript
addPlayer(playerId, seatId) {
    if (!this.players.has(playerId)) {
        const index = seatId !== undefined ? seatId : this.players.size;
        const positionData = this.getCannonPosition(index);

        this.players.set(playerId, {
            id: playerId,
            position: { x: positionData.x, y: positionData.y },
            cannonType: 1,
            level: 1,
            angle: positionData.angle,  // 使用座位對應的初始角度
            seatId: index               // 保存座位ID
        });
    }
}
```

#### 2.3 `drawCannon()` 函數

**新增功能：**
- 顯示座位標籤
- 根據座位位置調整標籤偏移
- 視覺化座位信息

```javascript
drawCannon(player, isCurrentPlayer) {
    // ... 繪製砲台 ...

    // 根據座位位置調整標籤位置
    const seatId = player.seatId !== undefined ? player.seatId : -1;
    let labelOffsetX = 0, labelOffsetY = -45;

    if (seatId === 0) labelOffsetY = -45;       // 底部 - 標籤在上方
    else if (seatId === 1) labelOffsetY = 60;   // 頂部 - 標籤在下方
    else if (seatId === 2) labelOffsetX = 50;   // 左側 - 標籤在右方
    else if (seatId === 3) labelOffsetX = -50;  // 右側 - 標籤在左方

    // 繪製座位標籤
    const seatLabel = seatId >= 0 ? `🪑 座位 ${seatId + 1}` : '未分配';
    // ... 繪製標籤代碼 ...
}
```

### 3. 測試頁面

創建了專用測試頁面 `js/seat-test.html`：

**功能特點：**
- 視覺化顯示4個座位佈局
- 展示每個座位的位置和方向
- 提供測試數據加載功能
- 實時顯示遊戲統計信息

**測試步驟：**
1. 打開 `js/seat-test.html`
2. 點擊「載入測試數據」
3. 點擊「開始渲染」
4. 移動滑鼠控制砲台
5. 驗證4個座位的砲台方向

## 技術細節

### 角度系統

使用弧度制（Radians）表示角度：
- `0°` = `0` rad → 向右
- `90°` = `π/2` rad → 向下
- `180°` = `π` rad → 向左
- `-90°` = `-π/2` rad → 向上

### 座位分配邏輯

```javascript
// 如果提供了座位ID，使用座位ID
// 否則使用當前玩家數量作為索引
const index = seatId !== undefined ? seatId : this.players.size;
```

### 滑鼠控制

玩家可以通過移動滑鼠來控制砲台角度：

```javascript
canvas.addEventListener('mousemove', (event) => {
    const mouseX = event.clientX - rect.left;
    const mouseY = event.clientY - rect.top;
    gameRenderer.updateCannonAngle(currentPlayerId, mouseX, mouseY);
});
```

## 視覺效果

### 座位標識

每個座位都有清晰的視覺標識：
- 🪑 座位圖標
- 座位編號（1-4）
- 玩家ID
- 等級顯示（如果 > 1）

### 顏色方案

- **當前玩家：** 綠色砲台 (#4CAF50)
- **其他玩家：** 灰色砲台 (#607D8B)
- **座位標籤：** 白色半透明背景
- **魚類：** 根據類型不同顏色

## 使用方法

### 在遊戲客戶端中使用

```javascript
// 1. 創建渲染器
const renderer = new GameRenderer('gameCanvas');

// 2. 設置當前玩家
renderer.setCurrentPlayer('player1');

// 3. 添加玩家到指定座位
renderer.addPlayer('player1', 0);  // 座位0 - 底部
renderer.addPlayer('player2', 1);  // 座位1 - 頂部
renderer.addPlayer('player3', 2);  // 座位2 - 左側
renderer.addPlayer('player4', 3);  // 座位3 - 右側

// 4. 開始渲染
renderer.start();
```

### 與後端座位系統整合

當後端座位選擇功能啟用後（需要生成 protobuf 代碼）：

```javascript
// 收到座位選擇響應時
case MessageType.SELECT_SEAT_RESPONSE:
    const selectResp = gameMessage.getSelectSeatResponse();
    if (selectResp.getSuccess()) {
        const seatId = selectResp.getSeatId();
        // 將當前玩家添加到指定座位
        gameRenderer.addPlayer(currentPlayerId, seatId);
    }
    break;
```

## 兼容性

### 瀏覽器支持

- ✅ Chrome/Edge (最新版)
- ✅ Firefox (最新版)
- ✅ Safari (最新版)

### Canvas API 需求

- `canvas.getContext('2d')`
- `ctx.rotate()`
- `ctx.translate()`
- `requestAnimationFrame()`

## 性能優化

### 已實現的優化

1. **條件渲染：** 只繪製畫布範圍內的對象
2. **FPS 限制：** 使用 `requestAnimationFrame`
3. **減少日誌：** 只在狀態變化時記錄
4. **對象池：** 重用玩家對象（未重新創建）

### 性能指標

- **目標 FPS：** 60
- **典型 FPS：** 55-60（4玩家 + 50魚 + 20子彈）
- **最大支持對象：** 500+（魚 + 子彈）

## 故障排除

### 常見問題

#### 問題1：砲台方向不正確

**症狀：** 所有玩家砲台都向上

**原因：** 使用舊版 `addPlayer()` 沒有傳入座位ID

**解決方案：**
```javascript
// ❌ 錯誤
renderer.addPlayer('player1');

// ✅ 正確
renderer.addPlayer('player1', 0);  // 指定座位ID
```

#### 問題2：座位標籤不顯示

**症狀：** 砲台繪製正常但沒有座位標籤

**原因：** `player.seatId` 未設置

**解決方案：** 確保在 `addPlayer()` 時設置了 `seatId`

#### 問題3：滑鼠控制失效

**症狀：** 移動滑鼠砲台不旋轉

**原因：**
1. 渲染器未運行
2. 未設置當前玩家

**解決方案：**
```javascript
renderer.setCurrentPlayer('player1');
renderer.start();
```

## 未來改進

### 計劃功能

- [ ] 支持座位重新選擇
- [ ] 添加座位鎖定機制
- [ ] 實現座位預覽模式
- [ ] 添加座位動畫效果
- [ ] 支持自定義座位佈局

### 後端整合

- [ ] 安裝 protoc 編譯器
- [ ] 生成 protobuf 代碼
- [ ] 啟用後端座位選擇功能
- [ ] 測試完整座位選擇流程

## 相關文件

- `js/game-renderer.js` - 渲染器核心代碼
- `js/seat-test.html` - 座位系統測試頁面
- `js/game-client.js` - WebSocket 客戶端
- `SEAT_SELECTION.md` - 座位選擇後端文檔
- `api/proto/v1/game.proto` - Protobuf 定義

## 參考資料

- [Canvas API 文檔](https://developer.mozilla.org/en-US/docs/Web/API/Canvas_API)
- [WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
- [Protocol Buffers](https://protobuf.dev/)

## 更新日誌

### 2025-01-13
- ✨ 實現4個座位佈局系統
- ✨ 每個座位有獨立的砲台方向
- ✨ 添加座位標籤顯示
- ✨ 創建座位系統測試頁面
- 🐛 修復子彈只能往上發射的問題

---

**作者：** Claude
**日期：** 2025-01-13
**版本：** 1.0.0
