-- TODO: 帳號模組 - 創建使用者表
-- 此表儲存所有使用者的帳號資訊，包括一般註冊、遊客和第三方登入

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE, -- 使用者名稱（一般註冊使用，遊客為 NULL）
    password_hash VARCHAR(255), -- 密碼雜湊（一般註冊使用，遊客和第三方登入為 NULL）
    nickname VARCHAR(100) NOT NULL, -- 暱稱
    avatar_url VARCHAR(255), -- 頭像 URL
    is_guest BOOLEAN NOT NULL DEFAULT FALSE, -- 是否為遊客
    third_party_provider VARCHAR(50), -- 第三方平台（google, facebook, qq 等）
    third_party_id VARCHAR(255), -- 第三方平台的使用者 ID
    coins BIGINT NOT NULL DEFAULT 1000, -- 金幣數量（初始贈送 1000）
    level INT NOT NULL DEFAULT 1, -- 等級
    exp BIGINT NOT NULL DEFAULT 0, -- 經驗值
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- 帳號是否啟用
    last_login_at TIMESTAMP WITH TIME ZONE, -- 最後登入時間
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引
CREATE INDEX idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX idx_users_third_party ON users(third_party_provider, third_party_id) WHERE third_party_provider IS NOT NULL;
CREATE INDEX idx_users_is_guest ON users(is_guest);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 創建約束：第三方登入的使用者必須有 provider 和 id
ALTER TABLE users ADD CONSTRAINT check_third_party
    CHECK (
        (third_party_provider IS NULL AND third_party_id IS NULL) OR
        (third_party_provider IS NOT NULL AND third_party_id IS NOT NULL)
    );

-- 創建約束：一般註冊的使用者必須有 username 和 password_hash
ALTER TABLE users ADD CONSTRAINT check_regular_user
    CHECK (
        is_guest = TRUE OR
        (username IS NOT NULL AND password_hash IS NOT NULL)
    );

-- 創建更新時間觸發器
CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_users_updated_at();
