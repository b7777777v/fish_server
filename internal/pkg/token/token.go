// internal/pkg/token/token.go
package token

import (
	"context"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定義了我們想要在 JWT 中攜帶的自訂資料
type CustomClaims struct {
	UserID   int64  `json:"user_id"`
	IsGuest  bool   `json:"is_guest,omitempty"`  // 是否為遊客
	Nickname string `json:"nickname,omitempty"`  // 遊客昵稱（僅遊客使用）
	jwt.RegisteredClaims
}

// TokenCache 定義 token 快取的介面
type TokenCache interface {
	StoreToken(ctx context.Context, token string, userID int64) error
	ValidateToken(ctx context.Context, token string) (bool, error)
	RevokeToken(ctx context.Context, token string) error
	RefreshTokenExpiry(ctx context.Context, token string) error
}

// TokenHelper 是一個輔助工具，用於生成和解析 JWT
type TokenHelper struct {
	secret     []byte
	issuer     string
	expire     int64
	tokenCache TokenCache // Redis token cache（可選）
}

// NewTokenHelper 創建一個新的 TokenHelper
func NewTokenHelper(c *conf.JWT) *TokenHelper {
	return &TokenHelper{
		secret:     []byte(c.Secret),
		issuer:     c.Issuer,
		expire:     c.Expire,
		tokenCache: nil, // 預設不啟用 cache
	}
}

// NewTokenHelperWithCache 創建一個帶 Redis cache 的 TokenHelper
func NewTokenHelperWithCache(c *conf.JWT, cache TokenCache) *TokenHelper {
	return &TokenHelper{
		secret:     []byte(c.Secret),
		issuer:     c.Issuer,
		expire:     c.Expire,
		tokenCache: cache,
	}
}

// SetTokenCache 設置 token cache（用於依賴注入後設置）
func (h *TokenHelper) SetTokenCache(cache TokenCache) {
	h.tokenCache = cache
}

// GenerateToken 生成一個新的 JWT
// 已過時：請使用 GenerateTokenWithClaims
func (h *TokenHelper) GenerateToken(userID uint) (string, error) {
	return h.GenerateTokenWithClaims(int64(userID), false)
}

// GenerateTokenWithClaims 生成一個新的 JWT，支援自訂 claims
// 如果啟用了 Redis cache，會自動將 token 存儲到 Redis（10 分鐘過期）
func (h *TokenHelper) GenerateTokenWithClaims(userID int64, isGuest bool) (string, error) {
	claims := CustomClaims{
		UserID:  userID,
		IsGuest: isGuest,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(h.expire))),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		return "", err
	}

	// 如果啟用了 token cache，存儲到 Redis
	if h.tokenCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.tokenCache.StoreToken(ctx, tokenString, userID); err != nil {
			// Redis 存儲失敗不影響 token 生成，只記錄錯誤
			// 可以根據業務需求決定是否要返回錯誤
		}
	}

	return tokenString, nil
}

// GenerateGuestToken 生成遊客專用的 JWT token（不使用數據庫 user_id）
// 如果啟用了 Redis cache，會自動將 token 存儲到 Redis（10 分鐘過期）
func (h *TokenHelper) GenerateGuestToken(nickname string) (string, error) {
	claims := CustomClaims{
		UserID:   0, // 遊客使用虛擬 ID 0
		IsGuest:  true,
		Nickname: nickname,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    h.issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(h.expire))),
			NotBefore: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.secret)
	if err != nil {
		return "", err
	}

	// 如果啟用了 token cache，存儲到 Redis（遊客使用 userID=0）
	if h.tokenCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.tokenCache.StoreToken(ctx, tokenString, 0); err != nil {
			// Redis 存儲失敗不影響 token 生成
		}
	}

	return tokenString, nil
}

// ParseToken 解析並驗證一個 JWT
// 如果啟用了 Redis cache，會額外檢查 token 是否在 Redis 中存在（是否已被撤銷）
func (h *TokenHelper) ParseToken(tokenString string) (*CustomClaims, error) {
	// 1. 如果啟用了 token cache，先檢查 Redis
	if h.tokenCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		valid, err := h.tokenCache.ValidateToken(ctx, tokenString)
		if err != nil {
			// Redis 查詢失敗，繼續使用 JWT 驗證（容錯處理）
		} else if !valid {
			// Token 不在 Redis 中或已過期（已被撤銷或超過 10 分鐘）
			return nil, jwt.ErrTokenExpired
		}
	}

	// 2. 解析並驗證 JWT
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return h.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// RevokeToken 撤銷 token（從 Redis 中刪除）
// 用於用戶登出、強制下線等場景
func (h *TokenHelper) RevokeToken(tokenString string) error {
	if h.tokenCache == nil {
		return nil // 未啟用 cache，無需撤銷
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return h.tokenCache.RevokeToken(ctx, tokenString)
}
