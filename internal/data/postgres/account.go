package postgres

import (
	"context"
	"database/sql"

	"github.com/b7777777v/fish_server/internal/biz/account"
)

// TODO: 實現帳號資料庫訪問層
// 此檔案實現 AccountRepo 介面，提供與 PostgreSQL 資料庫的互動功能

// accountRepo 實現 account.AccountRepo 介面
type accountRepo struct {
	db *Client
}

// NewAccountRepo 建立新的 AccountRepo 實例
func NewAccountRepo(db *Client) account.AccountRepo {
	return &accountRepo{
		db: db,
	}
}

// CreateUser 建立新使用者
func (r *accountRepo) CreateUser(ctx context.Context, user *account.User, passwordHash string) (*account.User, error) {
	query := `
		INSERT INTO users (username, password_hash, nickname, avatar_url, is_guest, third_party_provider, third_party_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	var id int64
	var createdAt, updatedAt string

	err := r.db.QueryRow(
		ctx,
		query,
		sql.NullString{String: user.Username, Valid: user.Username != ""},
		sql.NullString{String: passwordHash, Valid: passwordHash != ""},
		user.Nickname,
		sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		user.IsGuest,
		sql.NullString{String: user.ThirdPartyProvider, Valid: user.ThirdPartyProvider != ""},
		sql.NullString{String: user.ThirdPartyID, Valid: user.ThirdPartyID != ""},
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

// GetUserByUsername 根據使用者名稱獲取使用者
func (r *accountRepo) GetUserByUsername(ctx context.Context, username string) (*account.User, string, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_url, is_guest,
		       third_party_provider, third_party_id
		FROM users
		WHERE username = $1 AND is_active = true
	`

	var user account.User
	var passwordHash sql.NullString
	var usernameCol, avatarURL, thirdPartyProvider, thirdPartyID sql.NullString

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&usernameCol,
		&passwordHash,
		&user.Nickname,
		&avatarURL,
		&user.IsGuest,
		&thirdPartyProvider,
		&thirdPartyID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", nil // 使用者不存在
		}
		return nil, "", err
	}

	// 處理可為 NULL 的欄位
	user.Username = usernameCol.String
	user.AvatarURL = avatarURL.String
	user.ThirdPartyProvider = thirdPartyProvider.String
	user.ThirdPartyID = thirdPartyID.String

	return &user, passwordHash.String, nil
}

// GetUserByID 根據 ID 獲取使用者
func (r *accountRepo) GetUserByID(ctx context.Context, userID int64) (*account.User, error) {
	query := `
		SELECT id, username, nickname, avatar_url, is_guest,
		       third_party_provider, third_party_id
		FROM users
		WHERE id = $1 AND is_active = true
	`

	var user account.User
	var username, avatarURL, thirdPartyProvider, thirdPartyID sql.NullString

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&username,
		&user.Nickname,
		&avatarURL,
		&user.IsGuest,
		&thirdPartyProvider,
		&thirdPartyID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 使用者不存在
		}
		return nil, err
	}

	// 處理可為 NULL 的欄位
	user.Username = username.String
	user.AvatarURL = avatarURL.String
	user.ThirdPartyProvider = thirdPartyProvider.String
	user.ThirdPartyID = thirdPartyID.String

	return &user, nil
}

// GetUserByThirdParty 根據第三方平台 ID 獲取使用者
func (r *accountRepo) GetUserByThirdParty(ctx context.Context, provider, thirdPartyID string) (*account.User, error) {
	query := `
		SELECT id, username, nickname, avatar_url, is_guest,
		       third_party_provider, third_party_id
		FROM users
		WHERE third_party_provider = $1 AND third_party_id = $2 AND is_active = true
	`

	var user account.User
	var username, avatarURL, thirdPartyProviderCol, thirdPartyIDCol sql.NullString

	err := r.db.QueryRow(ctx, query, provider, thirdPartyID).Scan(
		&user.ID,
		&username,
		&user.Nickname,
		&avatarURL,
		&user.IsGuest,
		&thirdPartyProviderCol,
		&thirdPartyIDCol,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 使用者不存在
		}
		return nil, err
	}

	// 處理可為 NULL 的欄位
	user.Username = username.String
	user.AvatarURL = avatarURL.String
	user.ThirdPartyProvider = thirdPartyProviderCol.String
	user.ThirdPartyID = thirdPartyIDCol.String

	return &user, nil
}

// UpdateUser 更新使用者資料
func (r *accountRepo) UpdateUser(ctx context.Context, user *account.User) error {
	query := `
		UPDATE users
		SET nickname = $1, avatar_url = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.Exec(
		ctx,
		query,
		user.Nickname,
		sql.NullString{String: user.AvatarURL, Valid: user.AvatarURL != ""},
		user.ID,
	)

	return err
}
