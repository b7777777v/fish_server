# 🎮 Fish Server 遊戲核心模組

## 📋 概述

這個模組實現了魚類射擊遊戲的核心業務邏輯，包括房間管理、魚類生成、數學模型和遊戲流程控制。

## 🏗️ 核心組件

### 1. **Room (房間管理)**
- **`RoomManager`**: 房間管理器，負責房間的創建、銷毀和狀態管理
- **`Room`**: 房間實體，包含玩家、魚類、子彈等遊戲對象
- **多房間類型**: 新手房、中級房、高級房、VIP房，每種房間有不同的配置

#### 房間配置對比
| 房間類型 | 最小下注 | 最大下注 | 子彈倍數 | 最大魚數 |
|----------|----------|----------|----------|----------|
| 新手房   | 0.1元    | 1元      | 1.0x     | 20       |
| 中級房   | 1元      | 10元     | 2.0x     | 25       |
| 高級房   | 10元     | 100元    | 5.0x     | 30       |
| VIP房    | 100元    | 1000元   | 10.0x    | 35       |

### 2. **Spawner (魚類生成器)**
- **`FishSpawner`**: 負責魚類的生成和管理
- **13種魚類**: 從小丑魚到海王魚，涵蓋不同大小和稀有度
- **智能生成**: 基於稀有度的加權隨機生成
- **動態配置**: 支持生成率、位置、屬性的動態調整

#### 魚類分類
```
小型魚 (小丑魚, 熱帶魚, 銀魚)
├── 血量: 1
├── 獎勵: 5-10分
├── 出現率: 90%+
└── 命中率: 80-90%

中型魚 (石斑魚, 鯛魚, 比目魚)
├── 血量: 2-4
├── 獎勵: 25-40分
├── 出現率: 50-60%
└── 命中率: 60-70%

大型魚 (鯊魚, 鮪魚, 魔鬼魚)
├── 血量: 8-12
├── 獎勵: 100-150分
├── 出現率: 20-25%
└── 命中率: 40-50%

Boss級魚 (龍王魚, 金龍魚, 海王魚)
├── 血量: 30-80
├── 獎勵: 500-1000分
├── 出現率: 1-5%
└── 命中率: 10-20%
```

### 3. **MathModel (數學模型)**
- **簡化版實現**: 使用固定機率和基礎算法
- **莊家優勢**: 8%的莊家優勢確保遊戲平衡
- **命中計算**: 綜合考慮魚類大小、速度、稀有度等因素
- **獎勵計算**: 支持暴擊、倍數獎勵等機制

#### 核心參數
```yaml
基礎命中率: 70%
暴擊率: 10%
暴擊倍數: 2.0x
莊家優勢: 8%
最大賠付: 10倍
命中率範圍: 10%-95%
```

### 4. **GameUsecase (遊戲用例)**
- **業務邏輯封裝**: 將遊戲操作封裝為用例方法
- **數據持久化**: 與數據倉庫接口集成
- **事件記錄**: 記錄所有遊戲事件用於分析
- **統計功能**: 玩家遊戲統計和分析

## 🎯 主要功能

### 房間管理
```go
// 創建房間
room, err := gameUsecase.CreateRoom(ctx, RoomTypeNovice, 4)

// 玩家加入房間
err = gameUsecase.JoinRoom(ctx, roomID, playerID)

// 玩家離開房間
err = gameUsecase.LeaveRoom(ctx, roomID, playerID)
```

### 遊戲玩法
```go
// 玩家開火
bullet, err := gameUsecase.FireBullet(ctx, roomID, playerID, direction, power)

// 處理命中
hitResult, err := gameUsecase.HitFish(ctx, roomID, bulletID, fishID)

// 獲取房間狀態
roomState, err := gameUsecase.GetRoomState(ctx, roomID)
```

### 管理功能
```go
// 生成特殊魚類
fish, err := gameUsecase.SpawnSpecialFish(ctx, roomID, fishTypeID)

// 獲取遊戲統計
stats, err := gameUsecase.GetPlayerStatistics(ctx, playerID)

// 獲取數學模型統計
modelStats := gameUsecase.GetMathModelStats(ctx)
```

## 🔄 遊戲流程

### 1. 房間生命週期
```
創建房間 → 等待玩家 → 開始遊戲 → 遊戲循環 → 房間關閉
    ↓           ↓           ↓           ↓           ↓
  初始化魚類   玩家加入    生成新魚     更新狀態    清理資源
```

### 2. 遊戲循環 (10 FPS)
```
每100ms執行一次:
├── 更新魚類位置
├── 清理超時子彈  
├── 生成新魚類
└── 更新房間狀態
```

### 3. 命中計算流程
```
開火 → 計算基礎命中率 → 應用修正因子 → 判斷命中 → 計算傷害 → 計算獎勵
  ↓         ↓              ↓            ↓         ↓         ↓
威力檢查   魚類屬性      速度/稀有度    隨機判斷   威力影響   倍數計算
```

## 🧪 測試覆蓋

### 單元測試
- ✅ **MathModel**: 命中計算、獎勵計算
- ✅ **FishSpawner**: 魚類生成、類型管理
- ✅ **RoomManager**: 房間創建、玩家管理
- ✅ **GameUsecase**: 完整業務流程

### 集成測試
- ✅ **完整遊戲流程**: 創建房間 → 加入玩家 → 開火射擊 → 命中計算 → 離開房間
- ✅ **多玩家場景**: 多個玩家同時遊戲
- ✅ **邊界測試**: 異常情況處理

### 運行測試
```bash
# 運行所有測試
go test ./internal/biz/game/ -v

# 運行特定測試
go test ./internal/biz/game/ -v -run TestGameFlow

# 查看測試覆蓋率
go test ./internal/biz/game/ -v -cover
```

## 📊 性能特性

### 內存使用
- **房間數據**: 每個房間約 1-2KB
- **魚類對象**: 每條魚約 200-300 字節
- **子彈對象**: 每發子彈約 100-150 字節
- **總體估算**: 1000個活躍房間約 50-100MB

### 併發安全
- **讀寫鎖**: RoomManager 使用 RWMutex 保證併發安全
- **原子操作**: 關鍵計數器使用原子操作
- **無狀態組件**: MathModel 和 FishSpawner 設計為無狀態

### 擴展性
- **水平擴展**: 支持多實例部署
- **負載均衡**: 房間可以分散到不同實例
- **狀態分離**: 關鍵狀態可以外部化存儲

## 🔧 配置選項

### 房間配置
```yaml
min_bet: 10                    # 最小下注（分）
max_bet: 100                   # 最大下注（分）
bullet_cost_multiplier: 1.0    # 子彈成本倍數
fish_spawn_rate: 0.3           # 魚類生成率
max_fish_count: 20             # 最大魚數量
room_width: 1200               # 房間寬度
room_height: 800               # 房間高度
```

### 數學模型配置
```yaml
base_hit_rate: 0.7             # 基礎命中率
critical_rate: 0.1             # 暴擊率
critical_multiplier: 2.0       # 暴擊倍數
house_edge: 0.08               # 莊家優勢
max_payout: 10.0               # 最大賠付倍數
```

## 🚀 Wire 依賴注入

使用 Google Wire 進行依賴注入：

```go
// ProviderSet 包含所有遊戲組件
var ProviderSet = wire.NewSet(
    NewMathModelProvider,
    NewFishSpawnerProvider, 
    NewRoomManagerProvider,
    NewGameUsecaseProvider,
)
```

### 依賴關係
```
GameUsecase
├── GameRepo (由數據層提供)
├── PlayerRepo (由數據層提供)
├── RoomManager
│   ├── FishSpawner
│   └── MathModel
└── Logger (由基礎設施層提供)
```

## 🔄 後續擴展

### 短期規劃
- [ ] 添加更多魚類和特殊效果
- [ ] 實現技能系統和道具
- [ ] 添加房間主題和背景
- [ ] 優化數學模型算法

### 長期規劃  
- [ ] 機器學習驅動的難度調整
- [ ] 實時多人對戰模式
- [ ] 錦標賽和排行榜系統
- [ ] 虛擬現實支持

---

**注意**: 當前實現是簡化版，適合快速原型開發。生產環境需要根據實際需求進行優化和擴展。