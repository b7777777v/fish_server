// internal/pkg/token/token.go
package token

import (
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

// TokenHelper 是一個輔助工具，用於生成和解析 JWT
type TokenHelper struct {
	secret []byte
	issuer string
	expire int64
}

// NewTokenHelper 創建一個新的 TokenHelper
func NewTokenHelper(c *conf.JWT) *TokenHelper {
	return &TokenHelper{
		secret: []byte(c.Secret),
		issuer: c.Issuer,
		expire: c.Expire,
	}
}

// GenerateToken 生成一個新的 JWT
// 已過時：請使用 GenerateTokenWithClaims
func (h *TokenHelper) GenerateToken(userID uint) (string, error) {
	return h.GenerateTokenWithClaims(int64(userID), false)
}

// GenerateTokenWithClaims 生成一個新的 JWT，支援自訂 claims
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
	return token.SignedString(h.secret)
}

// GenerateGuestToken 生成遊客專用的 JWT token（不使用數據庫 user_id）
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
	return token.SignedString(h.secret)
}

// ParseToken 解析並驗證一個 JWT
func (h *TokenHelper) ParseToken(tokenString string) (*CustomClaims, error) {
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
