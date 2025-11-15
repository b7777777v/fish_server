# Testing Framework Documentation

本目录包含统一的测试工具和 Mock 实现，用于简化和标准化项目中的单元测试。

## 📁 目录结构

```
internal/testing/
├── mocks/              # Mock implementations using testify/mock
│   ├── game_repo.go
│   ├── player_repo.go
│   ├── wallet_repo.go
│   └── inventory_repo.go
├── testhelper/         # Test helper functions and utilities
│   ├── game_helper.go  # Game test environment setup
│   └── fixtures.go     # Test data fixtures
└── README.md           # This file
```

## 🎯 核心组件

### 1. Mock 包 (`mocks/`)

使用 `testify/mock` 实现的 Repository Mock，提供：
- **可验证的期望**：验证方法是否被正确调用
- **灵活的返回值**：支持动态返回值和错误注入
- **调用次数控制**：精确控制方法调用次数

#### 可用的 Mock

- `mocks.GameRepo` - 游戏仓储 Mock
- `mocks.PlayerRepo` - 玩家仓储 Mock
- `mocks.WalletRepo` - 钱包仓储 Mock
- `mocks.InventoryRepo` - 库存仓储 Mock

### 2. Test Helper 包 (`testhelper/`)

提供简化测试设置的工具函数。

#### GameTestEnv

完整的游戏测试环境，包含所有必要的 Mock 和业务逻辑组件。

```go
type GameTestEnv struct {
    Ctx context.Context
    Log logger.Logger

    // Mocked Repositories
    GameRepo      *mocks.GameRepo
    PlayerRepo    *mocks.PlayerRepo
    WalletRepo    *mocks.WalletRepo
    InventoryRepo *mocks.InventoryRepo

    // Business Logic Components
    WalletUsecase    *wallet.WalletUsecase
    Spawner          *game.FishSpawner
    MathModel        *game.MathModel
    InventoryManager *game.InventoryManager
    RTPController    *game.RTPController
    RoomManager      *game.RoomManager
    GameUsecase      *game.GameUsecase

    // Test Configuration
    RoomConfig game.RoomConfig
}
```

#### 测试数据工厂 (Fixtures)

预定义的测试数据构造函数：

- `NewTestPlayer(playerID)` - 创建测试玩家
- `NewTestPlayerWithBalance(playerID, balance)` - 创建带余额的测试玩家
- `NewTestWallet(walletID, userID)` - 创建测试钱包
- `NewTestFish(fishID, fishType)` - 创建测试鱼
- `NewTestBullet(bulletID, playerID, power, cost)` - 创建测试子弹
- `NewTestInventory(inventoryID, totalIn, totalOut)` - 创建测试库存
- `NewFishTypeFixtures()` - 创建标准鱼类型配置

## 🚀 快速开始

### 基础用法

```go
func TestMyFeature(t *testing.T) {
    // 1. 创建测试环境（自动设置所有 Mock 和依赖）
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t) // 验证所有 Mock 期望

    // 2. 使用测试环境进行测试
    room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
    assert.NoError(t, err)

    // 3. Mock 会自动处理默认行为
}
```

### 自定义 Mock 行为

```go
func TestWithCustomMock(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 设置自定义 Mock 期望
    customPlayer := testhelper.NewTestPlayerWithBalance(123, 50000)
    env.PlayerRepo.On("GetPlayer", env.Ctx, int64(123)).
        Return(customPlayer, nil).Once()

    // 测试代码...
    player, err := env.PlayerRepo.GetPlayer(env.Ctx, 123)
    assert.NoError(t, err)
    assert.Equal(t, int64(50000), player.Balance)
}
```

### 跳过默认 Mock

```go
func TestCustomSetup(t *testing.T) {
    // 跳过默认 Mock 设置，完全自定义
    env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
        SkipDefaultMocks: true,
    })
    defer env.AssertExpectations(t)

    // 完全自定义所有 Mock 行为
    env.GameRepo.On("GetAllFishTypes", mock.Anything).
        Return([]*game.FishType{ /* custom fish types */ }, nil)

    // 测试代码...
}
```

### 自定义房间配置

```go
func TestWithCustomConfig(t *testing.T) {
    customConfig := game.RoomConfig{
        MaxPlayers:   8,  // 自定义最大玩家数
        MinBet:       10,
        MaxBet:       500,
        MinFishCount: 20,
        MaxFishCount: 40,
        // ... 其他配置
    }

    env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
        RoomConfig: &customConfig,
    })
    defer env.AssertExpectations(t)

    // 测试使用自定义配置...
}
```

## 📚 进阶用法

### 验证方法调用次数

```go
func TestMethodCallCounts(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 期望方法被调用恰好一次
    env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, int64(1), int64(90000)).
        Return(nil).Once()

    // 期望方法被调用两次
    env.InventoryRepo.On("SaveInventory", env.Ctx, mock.Anything).
        Return(nil).Twice()

    // 期望方法被调用指定次数
    env.GameRepo.On("SaveGameEvent", env.Ctx, mock.Anything).
        Return(nil).Times(5)

    // 期望方法可能被调用（可选）
    env.GameRepo.On("GetRoom", env.Ctx, mock.Anything).
        Return(nil, nil).Maybe()
}
```

### 参数匹配器

```go
func TestArgumentMatchers(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 匹配任何 context
    env.PlayerRepo.On("GetPlayer", mock.Anything, int64(123)).Return(nil, nil)

    // 匹配任何类型的参数
    env.GameRepo.On("SaveRoom", mock.Anything, mock.AnythingOfType("*game.Room")).
        Return(nil)

    // 使用自定义匹配函数
    env.InventoryRepo.On("SaveInventory", env.Ctx, mock.MatchedBy(func(inv *game.Inventory) bool {
        return inv.CurrentRTP > 0.9 // 只匹配 RTP > 90% 的库存
    })).Return(nil)
}
```

### 动态返回值

```go
func TestDynamicReturnValues(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 使用函数动态计算返回值
    env.PlayerRepo.On("GetPlayer", mock.Anything, mock.Anything).
        Return(func(ctx context.Context, playerID int64) *game.Player {
            return &game.Player{
                ID:       playerID,
                Nickname: fmt.Sprintf("Player_%d", playerID),
                Balance:  playerID * 1000, // 动态余额
            }
        }, nil)
}
```

### 错误注入

```go
func TestErrorHandling(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, &testhelper.GameTestEnvOptions{
        SkipDefaultMocks: true,
    })
    defer env.AssertExpectations(t)

    // 模拟数据库错误
    env.PlayerRepo.On("GetPlayer", env.Ctx, int64(999)).
        Return(nil, errors.New("player not found"))

    // 测试错误处理
    player, err := env.PlayerRepo.GetPlayer(env.Ctx, 999)
    assert.Error(t, err)
    assert.Nil(t, player)
    assert.Contains(t, err.Error(), "player not found")
}
```

## 🎨 测试数据 Fixtures 使用

### 鱼类型 Fixtures

```go
func TestWithFishTypes(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 获取标准鱼类型
    fixtures := testhelper.NewFishTypeFixtures()

    // 使用预定义的鱼类型
    smallFish := testhelper.NewTestFish(1, fixtures.SmallFish)
    bossFish := testhelper.NewTestFish(2, fixtures.BossFish)

    // 或获取所有鱼类型
    allFishTypes := fixtures.AllFishTypes()
    env.GameRepo.On("GetAllFishTypes", env.Ctx).
        Return(allFishTypes, nil)
}
```

### 库存 Fixtures

```go
func TestInventoryScenarios(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 低 RTP 场景 (80%)
    lowRTPInv := testhelper.NewTestInventory("novice", 10000, 8000)

    // 高 RTP 场景 (110%)
    highRTPInv := testhelper.NewTestInventory("advanced", 200000, 220000)

    // 零库存场景
    emptyInv := testhelper.NewTestInventory("vip", 0, 0)

    env.InventoryRepo.On("GetInventory", env.Ctx, "novice").
        Return(lowRTPInv, nil)
}
```

## ✅ 最佳实践

### 1. 始终验证 Mock 期望

```go
func TestExample(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t) // ⭐ 重要：确保所有期望都被验证

    // 测试代码...
}
```

### 2. 使用有意义的测试名称

```go
func TestRTPController_WhenRTPBelowTarget_ShouldForceWin(t *testing.T) {
    // 清晰的测试名称说明了：
    // - 测试什么：RTPController
    // - 场景：When RTP Below Target
    // - 预期：Should Force Win
}
```

### 3. 使用子测试组织测试用例

```go
func TestGameFlow(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    t.Run("create room", func(t *testing.T) {
        // 测试创建房间
    })

    t.Run("join room", func(t *testing.T) {
        // 测试加入房间
    })

    t.Run("fire bullet", func(t *testing.T) {
        // 测试发射子弹
    })
}
```

### 4. 避免过度 Mock

```go
// ❌ 不好：Mock 太多细节
env.GameRepo.On("SaveRoom", env.Ctx, mock.MatchedBy(func(r *game.Room) bool {
    return r.ID == "room-1" &&
           r.Type == game.RoomTypeNovice &&
           r.MaxPlayers == 4 &&
           r.Status == game.RoomStatusWaiting
})).Return(nil)

// ✅ 好：只 Mock 必要的行为
env.GameRepo.On("SaveRoom", env.Ctx, mock.AnythingOfType("*game.Room")).
    Return(nil)
```

### 5. 使用 Fixtures 提高可读性

```go
// ❌ 不好：内联创建测试数据
player := &game.Player{
    ID: 1, UserID: 1, Nickname: "test",
    Balance: 100000, WalletID: 1, Status: game.PlayerStatusIdle,
}

// ✅ 好：使用 Fixture
player := testhelper.NewTestPlayerWithBalance(1, 100000)
```

## 🔧 故障排查

### Mock 期望未满足

```
Error: mock: Unexpected Method Call
```

**原因**：Mock 方法被调用但没有设置期望。

**解决方案**：
1. 添加 Mock 期望：`env.PlayerRepo.On("GetPlayer", ...).Return(...)`
2. 或使用默认 Mock（不设置 `SkipDefaultMocks: true`）

### Mock 期望未被调用

```
Error: FAIL: 0 out of 1 expectation(s) were met.
```

**原因**：设置了 Mock 期望但代码没有调用该方法。

**解决方案**：
1. 检查测试逻辑是否正确
2. 使用 `.Maybe()` 标记可选调用
3. 移除不必要的期望设置

## 📖 参考示例

完整的测试示例请参考：
- `internal/biz/game/game_refactored_test.go` - 重构后的游戏测试示例
- `internal/biz/game/game_test.go` - 原始测试（对比参考）

## 🤝 贡献指南

添加新的 Mock 时：
1. 在 `internal/testing/mocks/` 创建新文件
2. 使用 `testify/mock` 实现接口
3. 在 `testhelper/game_helper.go` 的 `setupDefaultMocks` 中添加默认行为
4. 更新此 README 文档

## 📝 许可证

与主项目相同。
