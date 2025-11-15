# 捕鱼游戏项目 - 详细问题清单

## 文件级别的问题汇总

### 🔴 高优先级 - 功能缺失

#### 1. 鱼潮系统 (Fish Tide System) - 3个关键文件

**文件1**: `/home/user/fish_server/internal/biz/game/fish_tide.go`
- 行数: 92
- 问题: 4个接口方法使用panic()实现
  - Line 66: `func (m *fishTideManager) StartTide()` → panic
  - Line 75: `func (m *fishTideManager) StopTide()` → panic
  - Line 82: `func (m *fishTideManager) GetActiveTide()` → panic
  - Line 91: `func (m *fishTideManager) ScheduleTides()` → panic
- TODO: 9处
- 优先级: 🔴 立即修复

**文件2**: `/home/user/fish_server/internal/data/postgres/fish_tide.go`
- 行数: 57
- 问题: 5个Repository方法使用panic()实现
  - Line 28: `func (r *fishTideRepo) GetTideByID()` → panic
  - Line 35: `func (r *fishTideRepo) GetActiveTides()` → panic
  - Line 42: `func (r *fishTideRepo) CreateTide()` → panic
  - Line 49: `func (r *fishTideRepo) UpdateTide()` → panic
  - Line 56: `func (r *fishTideRepo) DeleteTide()` → panic
- TODO: 5处
- 优先级: 🔴 立即修复

**文件3**: `/home/user/fish_server/internal/app/admin/fish_tide_handlers.go`
- 行数: 112
- 问题: 所有HTTP处理函数未实现（返回NotImplemented）
  - Line 55: `func handleGetFishTides()` → NotImplemented
  - Line 63: `func handleCreateFishTide()` → NotImplemented
  - Line 73: `func handleUpdateFishTide()` → NotImplemented
  - Line 85: `func handleDeleteFishTide()` → NotImplemented
  - Line 95: `func handleStartFishTide()` → NotImplemented
  - Line 106: `func handleStopFishTide()` → NotImplemented
- TODO: 9处
- 优先级: 🔴 立即修复
- 注意: 函数都是公开的但无法路由（RegisterFishTideRoutes未实现）

---

#### 2. OAuth登录系统 - 1个关键文件

**文件**: `/home/user/fish_server/internal/biz/account/oauth_service.go`
- 行数: 66
- 问题: GetUserInfo()方法返回错误而不是实现
  - Line 43-65: switch语句中所有case都返回"not implemented"
  - Google OAuth: Line 51-53
  - Facebook OAuth: Line 54-57
  - QQ OAuth: Line 58-61
- TODO: 9处
- 优先级: 🔴 立即修复
- 结构缺陷: oAuthService结构体无任何字段（Line 27-32）

---

#### 3. 大厅模块 - 4个相关文件

**文件1**: `/home/user/fish_server/internal/biz/lobby/repository.go`
- 行数: ~50
- 问题: 只有接口定义，无实现
- TODO: Line 7
- 状态: 需要实现

**文件2**: `/home/user/fish_server/internal/biz/lobby/usecase.go`
- 行数: 165
- 问题: 虽有TODO标记，实际已完整实现
- TODO: Line 8
- 注意: 可移除TODO注释

**文件3**: `/home/user/fish_server/internal/data/postgres/lobby.go`
- 行数: 92
- 问题: 虽有TODO标记，实际已完整实现
- TODO: Line 9
- 注意: 可移除TODO注释

**文件4**: `/home/user/fish_server/internal/data/redis/lobby.go`
- 行数: 80
- 问题: 虽有TODO标记，实际已完整实现
- TODO: Line 12
- 注意: 可移除TODO注释

---

### 🟡 中优先级 - 代码质量问题

#### 1. 账号数据库实现 - 标记错误

**文件**: `/home/user/fish_server/internal/data/postgres/account.go`
- 行数: 195
- 问题: 标记TODO但实际已完整实现
  - TODO: Line 11
  - CreateUser(): Line 27-55 ✅ 完整实现
  - GetUserByUsername(): Line 58-96 ✅ 完整实现
  - GetUserByID(): Line 99-135 ✅ 完整实现
  - GetUserByThirdParty(): Line 138-174 ✅ 完整实现
  - UpdateUser(): Line 177-194 ✅ 完整实现
- 建议: 移除TODO注释

---

#### 2. 测试架构缺陷

**文件1**: `/home/user/fish_server/internal/app/admin/handlers_test.go`
- 行数: ~150
- 问题: Mock注入无法工作
  - Line 109: `playerUC: nil, // TODO: 需要设置`
  - Line 110: `walletUC: nil, // TODO: 需要设计Mock注入`
- 影响: 无法进行完整的单元测试

**文件2**: `/home/user/fish_server/internal/app/admin/business_handlers_test.go`
- 行数: ~50
- 问题: 整个测试被跳过
  - Line 4: `// TODO: These tests require proper mock injection architecture`
  - Line 26: `t.Skip("TODO: Refactor test architecture...")`
- 影响: 业务处理器无测试覆盖

---

#### 3. 硬编码的值和魔术数字

**文件1**: `/home/user/fish_server/internal/app/game/message_handler.go`
- 行数: ~250
- 问题: 硬编码的子弹发射位置
  - Line 86: `position := game.Position{X: 600, Y: 750} // 默认位置`
- 建议: 改为配置参数或从请求获取

**文件2**: `/home/user/fish_server/internal/app/game/hub.go`
- 行数: ~500
- 问题: 硬编码的通道缓冲区大小
  - Line 104: `register: make(chan *Client, 10)`
  - Line 105: `unregister: make(chan *Client, 10)`
  - Line 106: `joinRoom: make(chan *JoinRoomMessage, 10)`
  - Line 107: `leaveRoom: make(chan *LeaveRoomMessage, 10)`
  - Line 108: `gameAction: make(chan *GameActionMessage, 100)`
  - Line 109: `broadcast: make(chan *BroadcastMessage, 100)`
- 建议: 提取为配置常量

---

#### 4. 性能优化TODO

**文件**: `/home/user/fish_server/internal/data/wallet_repo.go`
- 行数: 17764
- 问题: 缺少交易历史缓存
  - Line 375: `// TODO: [Cache] Caching transaction history...`
  - Line 380: `func (r *walletRepo) FindTransactionsByWalletID()` - 无缓存实现
- 影响: 高频访问时性能问题
- 建议: 实现Redis缓存，TTL 1-2分钟

---

#### 5. 缺少认证中间件

**文件**: `/home/user/fish_server/internal/app/admin/lobby_handlers.go`
- 行数: ~100
- 问题: 管理员API无认证保护
  - Line 41: `// TODO: 添加管理員認證中間件`
- 影响: 任何用户都能调用管理API
- 优先级: 🔴 安全问题

---

#### 6. 房间座位选择被注释掉

**文件**: `/home/user/fish_server/internal/app/game/room_manager.go`
- 行数: ~800
- 问题: 座位选择功能被禁用
  - Line 383: `// TODO: Uncomment after running 'make proto'...`
- 需要: 重新生成Protobuf代码

---

### 🟠 低优先级 - 架构改进

#### 1. 并发锁的优雅性问题

**文件**: `/home/user/fish_server/internal/app/game/hub.go`
- 问题位置: Line 298-303
- 代码:
  ```go
  h.mu.Unlock()
  h.broadcastToRoom(roomID, playerJoinMsg, client)
  h.mu.Lock() // 重新获取锁
  ```
- 问题: 显式的锁释放/重新获取容易出错
- 建议: 使用条件变量或通道重构

---

#### 2. Hub和RoomManager的紧耦合

**文件**: `/home/user/fish_server/internal/app/game/hub.go`
- 问题位置: Line 259-262
- 代码:
  ```go
  if h.roomManagers[roomID] == nil {
      roomManager := NewRoomManager(roomID, h.gameUsecase, h, h.logger)
      h.roomManagers[roomID] = roomManager
      go roomManager.Run()
  }
  ```
- 问题: 直接创建和管理RoomManager
- 建议: 使用工厂模式或依赖注入

---

## 按文件统计

### 完全未实现的文件 (5):
1. `internal/biz/game/fish_tide.go` - 92行 (panic)
2. `internal/data/postgres/fish_tide.go` - 57行 (panic)
3. `internal/app/admin/fish_tide_handlers.go` - 112行 (NotImplemented)
4. `internal/biz/account/oauth_service.go` - 66行 (ErrorReturn)
5. `internal/biz/lobby/repository.go` - ~50行 (Interface only)

### 部分实现但标记TODO的文件 (3):
1. `internal/data/postgres/account.go` - 195行 (已完整实现)
2. `internal/biz/lobby/usecase.go` - 165行 (已完整实现)
3. `internal/data/redis/lobby.go` - 80行 (已完整实现)

### 有严重设计问题的文件 (4):
1. `internal/app/admin/handlers_test.go` - Mock无法注入
2. `internal/app/admin/business_handlers_test.go` - 测试被跳过
3. `internal/app/game/message_handler.go` - 硬编码值
4. `internal/app/game/hub.go` - 锁的优雅性问题

---

## 按严重程度汇总

### 🔴 Critical (需立即修复):
- Fish Tide System: 3个文件，9处panic调用
- OAuth System: 1个文件，无法登录
- Admin Auth: 1个文件，安全漏洞
- **总计**: 5个文件，5处panic，1个安全漏洞

### 🟡 High (需尽快修复):
- Test Architecture: 2个文件，无法进行单元测试
- Hardcoded Values: 2个文件，多处魔术数字
- Cache Performance: 1个文件，性能问题
- Concurrency: 1个文件，潜在死锁风险
- **总计**: 6个文件

### 🟠 Medium (改进建议):
- Module Decoupling: 多处紧耦合
- Event Driven Architecture: 缺失
- Middleware: 日志、限流缺失
- **总计**: 多处架构改进机会

---

## 测试覆盖率详细分析

### 业务逻辑层 (internal/biz/)
**总文件**: 28个 (不含test.go)
**有测试的**: 1个 (game_test.go)
**无测试的**: 27个 (96.4%)

#### 无测试的关键模块:
1. **账号模块** (3文件):
   - account/usecase.go: Register, Login, GuestLogin, OAuthLogin
   - account/repository.go: 接口定义
   - account/oauth_service.go: GetUserInfo

2. **大厅模块** (2文件):
   - lobby/usecase.go: GetRoomList, GetPlayerStatus, etc.
   - lobby/repository.go: 接口定义

3. **玩家模块** (2文件):
   - player/usecase.go: 玩家相关业务
   - player/player.go: 玩家实体

4. **钱包模块** (2文件):
   - wallet/usecase.go: 充值、提现等
   - wallet/wallet.go: 钱包实体

5. **游戏模块** (主要，15+文件):
   - game/spawner.go: 鱼类生成
   - game/fish_formation.go: 鱼群阵型
   - game/fish_routes.go: 鱼类路线
   - game/room.go: 房间管理
   - game/entities.go: 实体定义
   - ... 其他10+文件无测试

### 数据层 (internal/data/)
**总文件**: 22个
**有测试的**: 4个 (18%)
**无测试的**: 18个 (82%)

#### 无测试的文件:
1. formation_config_repo.go
2. game_repo.go
3. inventory_repo.go
4. lobby_repos.go
5. player_repo.go
6. room_config_repo.go
7. postgres/account.go (已实现)
8. postgres/fish_tide.go (未实现)
9. postgres/lobby.go (已实现)
10. postgres/room_config.go
11. redis/lobby.go (已实现)
12. redis/token_cache.go
13. redis/room_config.go
14. ... 其他

