// internal/data/redis/token_cache.go
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

const (
	// TokenCacheDuration token 在 Redis 中的快取時間（10 分鐘）
	TokenCacheDuration = 10 * time.Minute

	// tokenKeyPrefix token 在 Redis 中的 key 前綴
	tokenKeyPrefix = "token:"
)

// TokenCache 提供 token 的 Redis 快取功能
type TokenCache struct {
	client *Client
	logger logger.Logger
}

// NewTokenCache 創建一個新的 TokenCache
func NewTokenCache(client *Client, logger logger.Logger) *TokenCache {
	return &TokenCache{
		client: client,
		logger: logger.With("module", "data/redis/token_cache"),
	}
}

// StoreToken 將 token 存儲到 Redis，設置 10 分鐘過期時間
// key 格式：token:{token_string}
// value 格式：user_id:{user_id}
func (tc *TokenCache) StoreToken(ctx context.Context, token string, userID int64) error {
	key := tokenKeyPrefix + token
	value := fmt.Sprintf("user_id:%d", userID)

	err := tc.client.Set(ctx, key, value, TokenCacheDuration)
	if err != nil {
		tc.logger.Errorf("failed to store token in redis: %v", err)
		return fmt.Errorf("failed to store token: %w", err)
	}

	tc.logger.Debugf("stored token for user %d, expires in %v", userID, TokenCacheDuration)
	return nil
}

// ValidateToken 驗證 token 是否存在於 Redis 中（是否有效）
// 返回 true 表示 token 有效且未過期
func (tc *TokenCache) ValidateToken(ctx context.Context, token string) (bool, error) {
	key := tokenKeyPrefix + token

	exists, err := tc.client.Exists(ctx, key)
	if err != nil {
		tc.logger.Errorf("failed to check token existence: %v", err)
		return false, fmt.Errorf("failed to validate token: %w", err)
	}

	return exists, nil
}

// RevokeToken 撤銷 token（從 Redis 中刪除）
// 用於登出、強制下線等場景
func (tc *TokenCache) RevokeToken(ctx context.Context, token string) error {
	key := tokenKeyPrefix + token

	err := tc.client.Del(ctx, key)
	if err != nil {
		tc.logger.Errorf("failed to revoke token: %v", err)
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	tc.logger.Infof("revoked token: %s", token)
	return nil
}

// RefreshTokenExpiry 刷新 token 的過期時間（重新設置為 10 分鐘）
// 用於保持活躍用戶的 session
func (tc *TokenCache) RefreshTokenExpiry(ctx context.Context, token string) error {
	key := tokenKeyPrefix + token

	err := tc.client.Expire(ctx, key, TokenCacheDuration)
	if err != nil {
		tc.logger.Errorf("failed to refresh token expiry: %v", err)
		return fmt.Errorf("failed to refresh token expiry: %w", err)
	}

	tc.logger.Debugf("refreshed token expiry: %s", token)
	return nil
}
