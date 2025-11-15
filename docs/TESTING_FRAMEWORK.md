# 测试框架重构 - 完整指南

## 📋 概述

本文档说明Fish Server项目的测试架构重构，引入统一的Mock框架和测试工具，提高测试代码的可维护性和可读性。

## 🎯 重构目标

### 问题

**旧测试架构存在的问题：**
1. ❌ Mock 实现分散在各个测试文件中，难以复用
2. ❌ 手写 Mock 无法验证方法调用
3. ❌ 测试设置代码冗长，重复度高
4. ❌ 测试数据内联创建，可读性差
5. ❌ 缺乏统一的测试模式和最佳实践

### 解决方案

**新测试架构提供：**
1. ✅ 统一的 Mock 包 (使用 testify/mock)
2. ✅ 可验证的 Mock 期望
3. ✅ 简化的测试环境设置 (一行代码)
4. ✅ 测试数据工厂 (Fixtures)
5. ✅ 完善的文档和示例

## 🏗️ 架构设计

### 目录结构

```
internal/testing/
├── mocks/                      # Mock 实现
│   ├── game_repo.go            # GameRepo Mock
│   ├── player_repo.go          # PlayerRepo Mock
│   ├── wallet_repo.go          # WalletRepo Mock
│   └── inventory_repo.go       # InventoryRepo Mock
│
├── testhelper/                 # 测试工具
│   ├── game_helper.go          # 测试环境设置
│   └── fixtures.go             # 测试数据工厂
│
├── examples/                   # 示例测试
│   └── game_test_example.go   # 使用示例
│
├── README.md                   # 使用文档
└── MIGRATION_GUIDE.md          # 迁移指南
```

### 核心组件

#### 1. Mock 包 (`internal/testing/mocks/`)

使用 `testify/mock` 实现的可验证 Mock：

```go
type GameRepo struct {
    mock.Mock
}

func (m *GameRepo) SaveRoom(ctx context.Context, room *game.Room) error {
    args := m.Called(ctx, room)
    return args.Error(0)
}
```

**特性：**
- ✅ 支持期望验证
- ✅ 灵活的返回值配置
- ✅ 调用次数控制 (.Once(), .Twice(), .Times(n))
- ✅ 参数匹配器 (mock.Anything, mock.AnythingOfType)

#### 2. 测试助手 (`internal/testing/testhelper/`)

##### GameTestEnv

完整的游戏测试环境：

```go
env := testhelper.NewGameTestEnv(t, nil)
defer env.AssertExpectations(t)

// 包含所有需要的组件：
// - env.GameRepo, env.PlayerRepo, env.WalletRepo, env.InventoryRepo
// - env.GameUsecase, env.RoomManager, env.RTPController
// - env.Ctx, env.Log
```

##### Test Fixtures

预定义的测试数据工厂：

```go
// 创建测试玩家
player := testhelper.NewTestPlayer(1)
richPlayer := testhelper.NewTestPlayerWithBalance(2, 500000)

// 创建测试鱼类型
fixtures := testhelper.NewFishTypeFixtures()
fish := testhelper.NewTestFish(1, fixtures.SmallFish)

// 创建测试库存
inventory := testhelper.NewTestInventory("novice", 10000, 8000) // RTP=80%
```

## 📚 使用指南

### 快速开始

```go
func TestMyFeature(t *testing.T) {
    // 1. 创建测试环境
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 2. 编写测试逻辑
    room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 1)
    assert.NoError(t, err)

    // 3. 验证结果
    assert.NotNil(t, room)
    assert.Equal(t, game.RoomTypeNovice, room.Type)
}
```

### 自定义 Mock 行为

```go
func TestCustomMock(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 设置 Mock 期望
    customPlayer := testhelper.NewTestPlayerWithBalance(123, 50000)
    env.PlayerRepo.On("GetPlayer", env.Ctx, int64(123)).
        Return(customPlayer, nil).Once()

    // 测试代码...
}
```

### 验证方法调用

```go
func TestMethodCalls(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)

    // 期望方法被调用恰好一次
    env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, int64(1), int64(90000)).
        Return(nil).Once()

    // 执行测试...

    // AssertExpectations 会验证是否被调用
}
```

## 🔄 迁移策略

### 迁移优先级

**高优先级（立即迁移）：**
- 核心业务逻辑测试
- 经常修改的测试
- 发现 Bug 需要修复的测试

**中优先级（逐步迁移）：**
- 稳定的功能测试
- 集成测试

**低优先级（可选迁移）：**
- 即将废弃的功能测试

### 迁移步骤

1. **添加导入**
   ```go
   import (
       "github.com/b7777777v/fish_server/internal/testing/testhelper"
       "github.com/stretchr/testify/mock"
   )
   ```

2. **替换测试设置**
   ```go
   // 旧代码
   env := setupTestEnvironment(t)

   // 新代码
   env := testhelper.NewGameTestEnv(t, nil)
   defer env.AssertExpectations(t)
   ```

3. **使用 Fixtures**
   ```go
   // 旧代码
   player := &game.Player{ID: 1, UserID: 1, Balance: 100000, ...}

   // 新代码
   player := testhelper.NewTestPlayer(1)
   ```

4. **添加 Mock 期望**
   ```go
   env.PlayerRepo.On("GetPlayer", env.Ctx, int64(1)).
       Return(testPlayer, nil)
   ```

详细迁移指南：`internal/testing/MIGRATION_GUIDE.md`

## 📖 文档资源

### 核心文档

| 文档 | 路径 | 内容 |
|------|------|------|
| 使用手册 | `internal/testing/README.md` | 完整使用文档 |
| 迁移指南 | `internal/testing/MIGRATION_GUIDE.md` | 从旧测试迁移 |
| 示例代码 | `internal/testing/examples/` | 实际使用示例 |
| 本文档 | `docs/TESTING_FRAMEWORK.md` | 架构概述 |

### 示例测试

参考文件：
- `internal/testing/examples/game_test_example.go` - 完整示例
- `internal/biz/game/game_test.go` - 原始测试（对比参考）

## 🛠️ 技术细节

### 使用的技术

- **testify/mock**: Mock 框架
- **testify/assert**: 断言库
- **依赖注入**: 通过构造函数注入依赖

### Mock 特性

| 特性 | 说明 | 示例 |
|------|------|------|
| 期望验证 | 验证方法是否被调用 | `env.AssertExpectations(t)` |
| 调用次数 | 控制方法调用次数 | `.Once()`, `.Twice()`, `.Times(3)` |
| 参数匹配 | 匹配方法参数 | `mock.Anything`, `mock.AnythingOfType` |
| 动态返回 | 函数计算返回值 | `Return(func(...) {...})` |
| 可选调用 | 方法可能被调用 | `.Maybe()` |

### 默认 Mock 行为

`NewGameTestEnv` 会自动设置默认 Mock 行为：

- **GameRepo**: 返回空数据或标准鱼类型
- **PlayerRepo**: 返回默认测试玩家
- **WalletRepo**: 返回默认钱包
- **InventoryRepo**: 返回空库存

可通过 `SkipDefaultMocks: true` 跳过默认设置。

## 📊 效果对比

### 测试代码量对比

**旧架构：**
```go
// ~50 行：Mock 定义
type MockGameRepo struct{}
func (m *MockGameRepo) SaveRoom(...) error { return nil }
// ... 更多方法

// ~40 行：测试设置
func setupTestEnvironment(t *testing.T) *testEnvironment {
    log := logger.New(...)
    gameRepo := &MockGameRepo{}
    // ... 更多设置
}

// ~20 行：测试代码
func TestFeature(t *testing.T) {
    env := setupTestEnvironment(t)
    // ...
}
```

**新架构：**
```go
// 0 行：Mock 已在 mocks/ 包中
// 0 行：测试设置已在 testhelper 中

// ~10 行：测试代码（更简洁！）
func TestFeature(t *testing.T) {
    env := testhelper.NewGameTestEnv(t, nil)
    defer env.AssertExpectations(t)
    // ...
}
```

**节省代码量：70-80%** 🎉

### 可维护性提升

| 指标 | 旧架构 | 新架构 | 改进 |
|------|--------|--------|------|
| Mock 复用 | ❌ 无法复用 | ✅ 完全复用 | ⭐⭐⭐ |
| 期望验证 | ❌ 不支持 | ✅ 完整支持 | ⭐⭐⭐ |
| 测试设置 | ~40 行 | 1 行 | ⭐⭐⭐ |
| 可读性 | 中等 | 高 | ⭐⭐ |
| 学习曲线 | 陡峭 | 平缓 | ⭐⭐ |

## ✅ 最佳实践

### 1. 始终验证 Mock 期望

```go
defer env.AssertExpectations(t) // ⭐ 重要！
```

### 2. 使用有意义的测试名称

```go
func TestRTPController_WhenRTPBelowTarget_ShouldForceWin(t *testing.T)
```

### 3. 使用子测试组织测试

```go
t.Run("create room", func(t *testing.T) { ... })
t.Run("join room", func(t *testing.T) { ... })
```

### 4. 使用 Fixtures 提高可读性

```go
player := testhelper.NewTestPlayer(1) // ✅ 清晰
```

### 5. 避免过度 Mock

只 Mock 必要的行为，使用 `.Maybe()` 标记可选调用。

## 🚀 下一步

### 对于开发者

1. 📖 阅读 `internal/testing/README.md`
2. 💻 查看 `internal/testing/examples/`
3. ✨ 新测试使用新框架
4. 🔄 逐步迁移旧测试

### 对于团队

1. **代码审查**：确保新测试使用新框架
2. **知识分享**：团队培训使用方法
3. **持续改进**：根据反馈优化框架

## 📞 支持

遇到问题？

1. 查阅 `internal/testing/README.md`
2. 参考 `internal/testing/examples/`
3. 联系团队成员

## 📝 更新日志

### v1.0.0 (2025-01-15)

**初始发布**
- ✨ 创建统一 Mock 包
- ✨ 实现测试助手工具
- ✨ 添加测试数据 Fixtures
- 📝 完善文档和示例
- 🎯 提供迁移指南

---

**Happy Testing! 🧪✨**
