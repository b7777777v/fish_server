-- 創建用戶表（完整版本，包含遊客和第三方登入支持）
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(100) UNIQUE, -- 使用者名稱（一般註冊使用，遊客為 NULL）
    password_hash VARCHAR(255), -- 密碼雜湊（一般註冊使用，遊客和第三方登入為 NULL）
    email VARCHAR(100) UNIQUE, -- 電子郵件
    nickname VARCHAR(100) NOT NULL, -- 暱稱
    avatar_url VARCHAR(255), -- 頭像 URL
    is_guest BOOLEAN NOT NULL DEFAULT FALSE, -- 是否為遊客
    third_party_provider VARCHAR(50), -- 第三方平台（google, facebook, qq 等）
    third_party_id VARCHAR(255), -- 第三方平台的使用者 ID
    coins BIGINT NOT NULL DEFAULT 1000, -- 金幣數量（初始贈送 1000）
    level INT NOT NULL DEFAULT 1, -- 等級
    exp BIGINT NOT NULL DEFAULT 0, -- 經驗值
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- 帳號是否啟用
    status SMALLINT NOT NULL DEFAULT 1, -- 1: 正常, 0: 禁用
    last_login_at TIMESTAMP WITH TIME ZONE, -- 最後登入時間
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建錢包表
CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
    status SMALLINT NOT NULL DEFAULT 1, -- 1: 正常, 0: 凍結
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, currency)
);

-- 創建錢包交易記錄表
CREATE TABLE IF NOT EXISTS wallet_transactions (
    id SERIAL PRIMARY KEY,
    wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    amount DECIMAL(20, 2) NOT NULL, -- 正數表示收入，負數表示支出
    balance_before DECIMAL(20, 2) NOT NULL,
    balance_after DECIMAL(20, 2) NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'deposit', 'withdraw', 'game_win', 'game_lose', 'bonus', etc.
    status SMALLINT NOT NULL DEFAULT 1, -- 1: 成功, 0: 失敗, 2: 處理中
    reference_id VARCHAR(100), -- 外部參考ID，例如遊戲ID或支付系統交易ID
    description TEXT,
    metadata JSONB, -- 額外的交易相關數據
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建用戶表的約束
-- 約束：第三方登入的使用者必須有 provider 和 id
ALTER TABLE users ADD CONSTRAINT check_third_party
    CHECK (
        (third_party_provider IS NULL AND third_party_id IS NULL) OR
        (third_party_provider IS NOT NULL AND third_party_id IS NOT NULL)
    );

-- 約束：一般註冊的使用者必須有 username 和 password_hash
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

-- 創建索引（使用 IF NOT EXISTS 確保冪等性）
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_third_party ON users(third_party_provider, third_party_id) WHERE third_party_provider IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_is_guest ON users(is_guest);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_created_at ON wallet_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX IF NOT EXISTS idx_wallet_transactions_reference_id ON wallet_transactions(reference_id);