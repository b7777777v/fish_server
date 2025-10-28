// internal/pkg/token/token.go
package token

import (
	"time"

	"github.com/b7777777v/fish_server/internal/conf"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims 定義了我們想要在 JWT 中攜帶的自訂資料
type CustomClaims struct {
	UserID uint `json:"user_id"`
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
func (h *TokenHelper) GenerateToken(userID uint) (string, error) {
	claims := CustomClaims{
		UserID: userID,
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
