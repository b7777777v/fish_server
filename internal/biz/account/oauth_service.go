package account

import (
	"context"
	"fmt"
)

// TODO: 實現 OAuth 第三方登入服務
// 此檔案提供與第三方平台（Google, Facebook, QQ 等）的 OAuth 整合

// OAuthService 定義 OAuth 服務介面
type OAuthService interface {
	// GetUserInfo 使用 authorization code 換取使用者資訊
	GetUserInfo(ctx context.Context, provider, code string) (*OAuthUserInfo, error)
}

// OAuthUserInfo OAuth 獲取的使用者資訊
type OAuthUserInfo struct {
	Provider      string // 平台名稱（google, facebook, qq）
	ThirdPartyID  string // 第三方平台的使用者 ID
	Email         string // 電子郵件（如果有）
	Nickname      string // 暱稱
	AvatarURL     string // 頭像 URL
}

// oAuthService 實現 OAuthService 介面
type oAuthService struct {
	// TODO: 添加第三方平台的 OAuth 配置
	// - Client ID
	// - Client Secret
	// - Redirect URI
}

// NewOAuthService 建立新的 OAuthService 實例
func NewOAuthService( /* TODO: 添加配置參數 */ ) OAuthService {
	return &oAuthService{
		// TODO: 初始化配置
	}
}

// GetUserInfo 使用 authorization code 換取使用者資訊
func (s *oAuthService) GetUserInfo(ctx context.Context, provider, code string) (*OAuthUserInfo, error) {
	// TODO: 實現 OAuth 流程
	// 1. 根據 provider 選擇對應的第三方平台
	// 2. 使用 code 換取 access_token
	// 3. 使用 access_token 獲取使用者資訊
	// 4. 將第三方使用者資訊轉換為 OAuthUserInfo

	switch provider {
	case "google":
		// TODO: 實現 Google OAuth
		// 參考：https://developers.google.com/identity/protocols/oauth2
		return nil, fmt.Errorf("google oauth is not implemented yet")
	case "facebook":
		// TODO: 實現 Facebook OAuth
		// 參考：https://developers.facebook.com/docs/facebook-login/manually-build-a-login-flow
		return nil, fmt.Errorf("facebook oauth is not implemented yet")
	case "qq":
		// TODO: 實現 QQ OAuth
		// 參考：https://wiki.connect.qq.com/oauth2-0%e7%ae%80%e4%bb%8b
		return nil, fmt.Errorf("qq oauth is not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported oauth provider: %s", provider)
	}
}
