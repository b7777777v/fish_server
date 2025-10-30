-- 創建用戶表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    nickname VARCHAR(50),
    avatar_url VARCHAR(255),
    status SMALLINT NOT NULL DEFAULT 1, -- 1: 正常, 0: 禁用
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

-- 創建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions(wallet_id);
CREATE INDEX idx_wallet_transactions_created_at ON wallet_transactions(created_at);
CREATE INDEX idx_wallet_transactions_type ON wallet_transactions(type);
CREATE INDEX idx_wallet_transactions_reference_id ON wallet_transactions(reference_id);