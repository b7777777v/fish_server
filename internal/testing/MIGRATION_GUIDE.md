# 测试迁移指南

本指南说明如何将现有测试从旧的 Mock 实现迁移到新的统一测试框架。

## 📋 迁移概览

### 旧架构 vs 新架构

| 方面 | 旧架构 | 新架构 |
|------|--------|--------|
| Mock 位置 | 测试文件内部 | `internal/testing/mocks/` |
| Mock 类型 | 手写 struct | `testify/mock` |
| 验证 | 无法验证调用 | 可验证期望 |
| 测试设置 | 长函数手动设置 | `testhelper.NewGameTestEnv()` |
| 测试数据 | 内联创建 | Fixtures 工厂函数 |
| 可复用性 | 低 | 高 |

## 🔄 迁移步骤

### 步骤 1：导入新包

**旧代码：**
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)
```

**新代码：**
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/b7777777v/fish_server/internal/testing/testhelper"
    "github.com/stretchr/testify/mock"
)
```

### 步骤 2：替换 Mock 定义

**旧代码：**
```go
type MockGameRepo struct{}

func (m *MockGameRepo) SaveRoom(ctx context.Context, room *Room) error {
    return nil
}
func (m *MockGameRepo) GetRoom(ctx context.Context, roomID string) (*Room, error) {
    return nil, nil
}
// ... 更多方法
```

**新代码：**
```go
// ✨ 不需要定义 Mock，直接使用 testhelper 提供的
// Mocks 已在 internal/testing/mocks/ 中定义
```

### 步骤 3：简化测试环境设置

**旧代码：**
```go
func setupTestEnvironment(t *testing.T) *testEnvironment {
    log := logger.New(os.Stdout, "debug", "console")
    gameRepo := &MockGameRepo{}
    playerRepo := &MockPlayerRepo{}
    walletRepo := &MockWalletRepo{}
    inventoryRepo := NewMockInventoryRepo()

    walletUC := wallet.NewWalletUsecase(walletRepo, log)

    testRoomConfig := RoomConfig{
        MinBet:               1,
        MaxBet:               100,
        BulletCostMultiplier: 1.0,
        FishSpawnRate:        0.3,
        MaxFishCount:         20,
        RoomWidth:            1200,
        RoomHeight:           800,
        TargetRTP:            0.96,
    }

    spawner := NewFishSpawner(log, testRoomConfig)
    mathModel := NewMathModel(log)
    inventoryManager, err := NewInventoryManager(inventoryRepo, log)
    assert.NoError(t, err)

    rtpController := NewRTPController(inventoryManager, log)
    roomManager := NewRoomManager(log, spawner, mathModel, inventoryManager, rtpController)
    gameUsecase := NewGameUsecase(gameRepo, playerRepo, walletUC, roomManager, spawner, mathModel, inventoryManager, rtpController, log)

    return &testEnvironment{
        ctx:              context.Background(),
        log:              log,
        gameRepo:         gameRepo,
        playerRepo:       playerRepo,
        inventoryRepo:    inventoryRepo,
        spawner:          spawner,
        mathModel:        mathModel,
        inventoryManager: inventoryManager,
        rtpController:    rtpController,
        roomManager:      roomManager,
        gameUsecase:      gameUsecase,
    }
}
```

**新代码：**
```go
// ✨ 一行代码完成所有设置！
func setupTestEnvironment(t *testing.T) *testhelper.GameTestEnv {
    return testhelper.NewGameTestEnv(t, nil)
}

// 或者直接在测试中使用
func TestExample(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)
    // 测试代码...
}
```

### 步骤 4：使用 Fixtures 替代内联测试数据

**旧代码：**
```go
func TestPlayerJoin(t *testing.T) {
    te := setupTestEnvironment(t)

    // 内联创建测试数据
    player := &Player{
        ID:       1,
        UserID:   1,
        Nickname: "TestPlayer",
        Balance:  100000,
        WalletID: 1,
        Status:   PlayerStatusIdle,
    }

    // 测试代码...
}
```

**新代码：**
```go
func TestPlayerJoin(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 使用 Fixture 创建测试数据
    player := testhelper.NewTestPlayer(1)

    // 或者自定义余额
    richPlayer := testhelper.NewTestPlayerWithBalance(2, 500000)

    // 测试代码...
}
```

### 步骤 5：添加 Mock 期望验证

**旧代码：**
```go
func TestSaveRoom(t *testing.T) {
    te := setupTestEnvironment(t)

    room, err := te.roomManager.CreateRoom(RoomTypeNovice, 1)
    assert.NoError(t, err)

    // ❌ 无法验证 SaveRoom 是否被调用
}
```

**新代码：**
```go
func TestSaveRoom(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t) // ✅ 验证所有期望

    // 设置期望：SaveRoom 应该被调用一次
    env.GameRepo.On("SaveRoom", env.Ctx, mock.AnythingOfType("*game.Room")).
        Return(nil).Once()

    room, err := env.RoomManager.CreateRoom(RoomTypeNovice, 1)
    assert.NoError(t, err)

    // AssertExpectations 会验证 SaveRoom 是否被调用
}
```

## 📝 实际迁移示例

### 示例 1：简单测试迁移

**旧代码 (game_test.go):**
```go
func TestInventoryManager(t *testing.T) {
    te := setupTestEnvironment(t)

    roomType := RoomTypeNovice
    te.inventoryManager.AddBet(roomType, 100)
    te.inventoryManager.AddWin(roomType, 50)

    inv := te.inventoryManager.GetInventory(roomType)
    assert.Equal(t, int64(100), inv.TotalIn)
    assert.Equal(t, int64(50), inv.TotalOut)
    assert.Equal(t, 0.5, inv.CurrentRTP)
}
```

**新代码 (game_refactored_test.go):**
```go
func TestInventoryManager_Refactored(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    roomType := RoomTypeNovice

    // 设置 Mock 期望
    initialInventory := testhelper.NewTestInventory(string(roomType), 0, 0)
    env.InventoryRepo.On("GetInventory", env.Ctx, string(roomType)).
        Return(initialInventory, nil).Maybe()
    env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
        Return(nil).Maybe()

    // 测试逻辑不变
    env.InventoryManager.AddBet(roomType, 100)
    env.InventoryManager.AddWin(roomType, 50)

    inv := env.InventoryManager.GetInventory(roomType)
    assert.Equal(t, int64(100), inv.TotalIn)
    assert.Equal(t, int64(50), inv.TotalOut)
    assert.Equal(t, 0.5, inv.CurrentRTP)
}
```

**改进点：**
- ✅ 使用 `testhelper.NewGameTestEnv` 简化设置
- ✅ 使用 `testhelper.NewTestInventory` 创建测试数据
- ✅ 添加 Mock 期望验证
- ✅ 使用 `defer env.AssertExpectations(t)` 自动验证

### 示例 2：复杂测试迁移

**旧代码:**
```go
func TestGameFlowWithRTP(t *testing.T) {
    te := setupTestEnvironment(t)

    // 1. Create Room & Player
    room, err := te.gameUsecase.CreateRoom(te.ctx, RoomTypeNovice, 1)
    assert.NoError(t, err)

    playerID := int64(1)
    err = te.gameUsecase.JoinRoom(te.ctx, room.ID, playerID)
    assert.NoError(t, err)

    // ... 更多测试代码
}
```

**新代码:**
```go
func TestGameFlow_Refactored(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 使用 Fixture 创建测试数据
    playerID := int64(1)
    testPlayer := testhelper.NewTestPlayerWithBalance(playerID, 100000)
    env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)

    // 设置库存 Mock
    inventory := testhelper.NewTestInventory("novice", 0, 0)
    env.InventoryRepo.On("GetInventory", env.Ctx, string(RoomTypeNovice)).
        Return(inventory, nil).Maybe()
    env.InventoryRepo.On("SaveInventory", env.Ctx, mock.AnythingOfType("*game.Inventory")).
        Return(nil).Maybe()

    // 1. Create Room
    room, err := env.GameUsecase.CreateRoom(env.Ctx, RoomTypeNovice, 1)
    assert.NoError(t, err)

    // 2. Join Room
    err = env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)
    assert.NoError(t, err)

    // ... 更多测试代码
}
```

**改进点：**
- ✅ 明确的 Mock 期望设置
- ✅ 使用 Fixtures 提高可读性
- ✅ 可验证的测试行为

## 🎯 迁移检查清单

在迁移每个测试时，确保：

- [ ] 使用 `testhelper.NewGameTestEnv` 创建测试环境
- [ ] 添加 `defer env.AssertExpectations(t)` 验证 Mock
- [ ] 使用 Fixtures 替代内联测试数据
- [ ] 为关键操作设置明确的 Mock 期望
- [ ] 测试名称清晰描述测试场景
- [ ] 使用子测试组织相关测试用例
- [ ] 删除旧的 Mock 定义（如果已迁移完成）

## 📊 迁移优先级

**高优先级（立即迁移）：**
- 核心业务逻辑测试
- 经常修改的测试
- 发现 Bug 需要修复的测试

**中优先级（逐步迁移）：**
- 稳定的功能测试
- 集成测试

**低优先级（可选迁移）：**
- 即将废弃的功能测试
- 性能测试（使用真实实现）

## 💡 迁移技巧

### 1. 逐步迁移

不需要一次性迁移所有测试：
- 新测试直接使用新框架
- 旧测试逐步迁移
- 两种方式可以共存

### 2. 保留原测试作为参考

```go
// 原测试（保留作为参考）
func TestRTPController(t *testing.T) {
    // ... 旧代码
}

// 重构后的测试
func TestRTPController_Refactored(t *testing.T) {
    // ... 新代码
}
```

验证通过后删除旧测试。

### 3. 使用 Table-Driven Tests

**旧代码：**
```go
func TestCalculateReward_SmallFish(t *testing.T) { /* ... */ }
func TestCalculateReward_MediumFish(t *testing.T) { /* ... */ }
func TestCalculateReward_LargeFish(t *testing.T) { /* ... */ }
```

**新代码：**
```go
func TestCalculateReward(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    fixtures := testhelper.NewFishTypeFixtures()
    tests := []struct {
        name     string
        fishType *game.FishType
        expected int64
    }{
        {"small fish", fixtures.SmallFish, 10},
        {"medium fish", fixtures.MediumFish, 50},
        {"large fish", fixtures.LargeFish, 200},
        {"boss fish", fixtures.BossFish, 1000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            reward := calculateReward(tt.fishType)
            assert.Equal(t, tt.expected, reward)
        })
    }
}
```

## 🚀 下一步

1. **阅读完整文档**：`internal/testing/README.md`
2. **查看示例**：`internal/biz/game/game_refactored_test.go`
3. **开始迁移**：选择一个简单的测试开始
4. **逐步推进**：每次迁移一个测试文件

## ❓ 常见问题

### Q: 是否必须迁移所有旧测试？
A: 不是。新测试使用新框架，旧测试可以保留并逐步迁移。

### Q: 新框架是否支持集成测试？
A: 是的。可以通过 `SkipDefaultMocks: true` 使用真实实现。

### Q: 如何测试错误场景？
A: 设置 Mock 返回错误：
```go
env.PlayerRepo.On("GetPlayer", env.Ctx, int64(999)).
    Return(nil, errors.New("player not found"))
```

### Q: 迁移后测试变慢了？
A: 不应该。如果变慢，检查：
- 是否有不必要的 Mock 期望
- 日志级别是否设置为 "error"
- 是否有意外的真实 I/O 操作

## 📞 需要帮助？

如遇到迁移问题：
1. 查阅 `README.md` 完整文档
2. 参考 `game_refactored_test.go` 示例
3. 联系团队成员讨论

---

**祝迁移顺利！** 🎉
