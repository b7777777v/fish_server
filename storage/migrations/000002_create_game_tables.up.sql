-- 創建遊戲房間表
CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(36) PRIMARY KEY, -- e.g., UUID
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'novice', 'intermediate', 'advanced', etc.
    status VARCHAR(20) NOT NULL DEFAULT 'waiting', -- 'waiting', 'playing', 'closed'
    max_players SMALLINT NOT NULL DEFAULT 4,
    config JSONB, -- 存儲房間配置，如最小/最大下注，魚生成率等
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建遊戲統計表
CREATE TABLE IF NOT EXISTS game_statistics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_shots BIGINT NOT NULL DEFAULT 0,
    total_hits BIGINT NOT NULL DEFAULT 0,
    total_rewards DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    total_costs DECIMAL(20, 2) NOT NULL DEFAULT 0.00,
    fish_killed BIGINT NOT NULL DEFAULT 0,
    play_time_seconds BIGINT NOT NULL DEFAULT 0,
    last_played_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id) -- 假設每個用戶只有一條總的統計記錄
);

-- 創建遊戲事件表
CREATE TABLE IF NOT EXISTS game_events (
    id BIGSERIAL PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    user_id INTEGER REFERENCES users(id), -- 有些事件可能與玩家無關
    event_type VARCHAR(50) NOT NULL, -- 'player_join', 'fish_spawn', 'bullet_fire', etc.
    data JSONB, -- 存儲事件的具體數據
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 為新表創建索引
CREATE INDEX idx_rooms_type ON rooms(type);
CREATE INDEX idx_game_statistics_user_id ON game_statistics(user_id);
CREATE INDEX idx_game_events_room_id ON game_events(room_id);
CREATE INDEX idx_game_events_user_id ON game_events(user_id);
CREATE INDEX idx_game_events_event_type ON game_events(event_type);
CREATE INDEX idx_game_events_timestamp ON game_events(timestamp);
