# 快速参考 - 优先修复清单

## 这周必须做的 (Critical Path)

### 第1天: 移除Panic调用 (2小时)

**文件1**: `/home/user/fish_server/internal/biz/game/fish_tide.go`
```go
// 改变这个
func (m *fishTideManager) StartTide(...) error {
    panic("not implemented")
}

// 改为
func (m *fishTideManager) StartTide(...) error {
    return fmt.Errorf("fish tide system not yet implemented")
}
```

**文件2**: `/home/user/fish_server/internal/data/postgres/fish_tide.go`
```go
// 改变所有panic调用
func (r *fishTideRepo) GetTideByID(...) (*game.FishTide, error) {
    panic("not implemented")  // <- 改为 return nil, ErrNotImplemented
}
```

### 第2天: 添加管理员认证 (3小时)

**文件**: `/home/user/fish_server/internal/app/admin/lobby_handlers.go`

```go
// 在 RegisterLobbyRoutes 之前添加
func AdminAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从header获取认证token
        token := c.GetHeader("Authorization")
        
        // 验证token有效性
        if !isValidAdminToken(token) {
            c.JSON(http.StatusUnauthorized, 
                gin.H{"error": "unauthorized admin access"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// 在路由注册时使用
func RegisterLobbyRoutes(r *gin.Engine, handler *AdminHandler) {
    adminGroup := r.Group("/api/v1/admin")
    adminGroup.Use(AdminAuthMiddleware())  // 添加此行
    {
        adminGroup.GET("/lobbies", handler.GetLobbies)
        // ... 其他路由
    }
}
```

### 第3天: 修复测试架构 (4小时)

**文件**: `/home/user/fish_server/internal/app/admin/handlers_test.go`

```go
// 当前问题: 无法注入Mock
type MockPlayerUsecase struct {
    mock.Mock
}

// 改为支持接口注入
type AdminHandlerTest struct {
    playerUC wallet.PlayerUsecase
    walletUC wallet.WalletUsecase
}

// 使用Wire进行依赖注入
func setupTest(t *testing.T) *AdminHandlerTest {
    // 为测试创建Mock
    mockPlayer := new(MockPlayerUsecase)
    mockWallet := new(MockWalletUsecase)
    
    return &AdminHandlerTest{
        playerUC: mockPlayer,
        walletUC: mockWallet,
    }
}
```

---

## 这个月需要完成的 (Priority 1)

### 1. 完成Fish Tide系统实现 (预计16小时)

**必要的文件**:
- `internal/biz/game/fish_tide.go` - 实现4个方法
- `internal/data/postgres/fish_tide.go` - 实现5个CRUD操作
- `internal/app/admin/fish_tide_handlers.go` - 实现6个HTTP处理器
- 创建数据库迁移文件 (CREATE TABLE fish_tide_config)

**核心逻辑框架**:
```go
// FishTideManager.StartTide() 应该做:
1. 从数据库加载魚潮配置 (GetTideByID)
2. 验证当前房间没有活跃魚潮
3. 创建定时器，在指定时间内触发生成
4. 向所有玩家广播魚潮开始事件
5. 设置自动停止定时器

// FishTideManager.ScheduleTides() 应该做:
1. 获取所有启用的魚潮配置 (GetActiveTides)
2. 根据触发规则创建定时器
3. 当触发条件满足时，调用 StartTide()
```

### 2. 完成OAuth登录系统 (预计12小时)

**必要的库**:
```bash
go get google.golang.org/oauth2
go get github.com/google.golang.org/oauth2/google
go get github.com/go-social-auth/...
```

**核心实现**:
```go
// internal/biz/account/oauth_service.go
type oAuthService struct {
    googleConfig    *oauth2.Config
    facebookConfig  *oauth2.Config
    qqConfig        *oauth2.Config
}

func (s *oAuthService) GetUserInfo(...) (*OAuthUserInfo, error) {
    switch provider {
    case "google":
        // 1. 使用code换取token
        token, err := s.googleConfig.Exchange(ctx, code)
        // 2. 使用token获取用户信息
        userInfo, err := getGoogleUserInfo(token)
        // 3. 返回标准化的OAuthUserInfo
        return convertToOAuthUserInfo(userInfo), nil
    // ... 其他平台
    }
}
```

### 3. 为关键模块添加单元测试 (预计20小时)

**优先级排序**:

1. **Account Module** (4小时):
   - `internal/biz/account/usecase_test.go`
   - 测试: Register, Login, GuestLogin, GetUserByID, UpdateUser

2. **Wallet Module** (3小时):
   - `internal/biz/wallet/usecase_test.go`
   - 测试: CreateWallet, Deposit, Withdraw, GetBalance

3. **Game Module** (8小时):
   - 扩展现有的 `game_test.go`
   - 测试: 鱼类生成、阵型管理、房间管理

4. **Data Layer** (5小时):
   - 补充缺失的数据层测试
   - 重点: player_repo, game_repo, formation_config_repo

---

## 代码片段 - 快速修复

### 修复1: 移除硬编码位置

**当前代码** (internal/app/game/message_handler.go:86):
```go
position := game.Position{X: 600, Y: 750} // 默认位置
if fireData.Position != nil {
    position = game.Position{
        X: fireData.Position.X,
        Y: fireData.Position.Y,
    }
}
```

**改进后**:
```go
// 从配置读取
position := getDefaultCannonPosition()  // X: 600, Y: 750
if fireData.Position != nil {
    position = game.Position{
        X: fireData.Position.X,
        Y: fireData.Position.Y,
    }
}
```

### 修复2: 提取魔术数字

**当前代码** (internal/app/game/hub.go:104-108):
```go
register:      make(chan *Client, 10),
unregister:    make(chan *Client, 10),
joinRoom:      make(chan *JoinRoomMessage, 10),
leaveRoom:     make(chan *LeaveRoomMessage, 10),
gameAction:    make(chan *GameActionMessage, 100),
broadcast:     make(chan *BroadcastMessage, 100),
```

**改进后**:
```go
const (
    ChannelBufferSmall  = 10
    ChannelBufferLarge  = 100
)

// 在Hub初始化中
register:      make(chan *Client, ChannelBufferSmall),
unregister:    make(chan *Client, ChannelBufferSmall),
joinRoom:      make(chan *JoinRoomMessage, ChannelBufferSmall),
leaveRoom:     make(chan *LeaveRoomMessage, ChannelBufferSmall),
gameAction:    make(chan *GameActionMessage, ChannelBufferLarge),
broadcast:     make(chan *BroadcastMessage, ChannelBufferLarge),
```

### 修复3: 添加交易缓存

**当前代码** (internal/data/wallet_repo.go:375):
```go
// TODO: [Cache] Caching transaction history...
func (r *walletRepo) FindTransactionsByWalletID(...) {
    // 直接查询数据库
}
```

**改进后**:
```go
const (
    TransactionCacheTTL = 2 * time.Minute
    TransactionCacheKeyFormat = "wallet:transactions:%d:page:%d"
)

func (r *walletRepo) FindTransactionsByWalletID(
    ctx context.Context, 
    walletID uint, 
    limit, offset int,
) ([]*wallet.Transaction, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf(TransactionCacheKeyFormat, walletID, offset/limit+1)
    if cached, err := r.data.redis.Get(ctx, cacheKey).Result(); err == nil {
        var transactions []*wallet.Transaction
        json.Unmarshal([]byte(cached), &transactions)
        return transactions, nil
    }
    
    // 2. 缓存未命中，查询数据库
    transactions, err := r.queryFromDB(ctx, walletID, limit, offset)
    if err != nil {
        return nil, err
    }
    
    // 3. 缓存结果
    if data, err := json.Marshal(transactions); err == nil {
        r.data.redis.Set(ctx, cacheKey, data, TransactionCacheTTL)
    }
    
    return transactions, nil
}
```

---

## 检查清单 - 完成验证

### 代码审查阶段:
- [ ] 所有panic()调用已移除或替换为错误返回
- [ ] 所有HTTP处理器都返回有效响应
- [ ] 管理员API都有认证保护
- [ ] 所有TODO都有对应的Issue/任务

### 测试阶段:
- [ ] 新增测试覆盖率达到>80%的关键模块
- [ ] 所有单元测试能正确运行 (`go test ./...`)
- [ ] 集成测试通过
- [ ] 没有跳过的测试 (t.Skip)

### 代码质量阶段:
- [ ] 运行linter通过: `make lint`
- [ ] 没有警告或错误
- [ ] 代码注释清晰
- [ ] 遵循项目编码规范

### 部署前:
- [ ] 新增/修改的API文档已更新
- [ ] 数据库迁移脚本已创建
- [ ] 环境变量配置已检查
- [ ] 没有硬编码的凭证或密钥

---

## 推荐的开发顺序

### Week 1: 修复Critical问题
1. 周一: 移除panic、添加认证 (4小时)
2. 周二: 重构测试架构 (4小时)
3. 周三-周五: 完成Fish Tide系统基础实现 (8小时)

### Week 2: 完成主要功能
1. 完成Fish Tide剩余逻辑 (8小时)
2. 完成OAuth登录系统 (12小时)

### Week 3-4: 测试和优化
1. 添加单元测试 (20小时)
2. 集成测试 (8小时)
3. 性能优化和缓存实现 (12小时)

---

## 推荐的Git提交顺序

```bash
# Commit 1: 移除panic
git commit -m "fix: replace panic calls with proper error handling"

# Commit 2: 添加认证中间件
git commit -m "feat: add admin authentication middleware"

# Commit 3: 重构测试架构
git commit -m "refactor: improve test mock injection architecture"

# Commit 4: 实现Fish Tide系统
git commit -m "feat: implement fish tide system (business logic)"

# Commit 5: 数据库实现
git commit -m "feat: implement fish tide PostgreSQL repository"

# Commit 6: HTTP API
git commit -m "feat: implement fish tide HTTP handlers"

# Commit 7: OAuth登录
git commit -m "feat: implement OAuth login system"

# Commit 8: 单元测试
git commit -m "test: add comprehensive unit tests for core modules"

# Commit 9: 缓存优化
git commit -m "perf: add transaction history caching with Redis"

# Commit 10: 架构改进
git commit -m "refactor: decouple Hub and RoomManager, improve concurrency"
```

