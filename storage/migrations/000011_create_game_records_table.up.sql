-- 創建遊戲記錄表（完整的遊戲會話記錄）
CREATE TABLE IF NOT EXISTS game_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id VARCHAR(36) NOT NULL,
    session_id VARCHAR(100), -- 遊戲會話ID，可用於關聯多個記錄

    -- 時間相關
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE, -- NULL 表示遊戲進行中
    duration_seconds INT, -- 遊戲時長（秒）

    -- 財務統計
    total_bets DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- 總投注（所有子彈費用）
    total_wins DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- 總獎勵（所有捕獲獎勵）
    net_profit DECIMAL(20, 2) NOT NULL DEFAULT 0.00, -- 淨盈虧（total_wins - total_bets）

    -- 遊戲統計
    bullets_fired BIGINT NOT NULL DEFAULT 0, -- 發射子彈數量
    bullets_hit BIGINT NOT NULL DEFAULT 0, -- 命中子彈數量
    fish_caught BIGINT NOT NULL DEFAULT 0, -- 捕獲魚數量
    hit_rate DECIMAL(5, 2), -- 命中率（百分比）

    -- 獎勵統計
    max_single_win DECIMAL(20, 2) DEFAULT 0.00, -- 最大單次獎勵
    bonus_count INT DEFAULT 0, -- 獎金次數（暴擊、特殊魚等）

    -- 狀態
    status VARCHAR(20) NOT NULL DEFAULT 'playing', -- 'playing', 'finished', 'abandoned'

    -- 額外數據（JSONB 格式，靈活擴展）
    metadata JSONB, -- 例如：魚類型分佈、使用的砲台等級等

    -- 時間戳
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引以提升查詢性能
CREATE INDEX IF NOT EXISTS idx_game_records_user_id ON game_records(user_id);
CREATE INDEX IF NOT EXISTS idx_game_records_room_id ON game_records(room_id);
CREATE INDEX IF NOT EXISTS idx_game_records_session_id ON game_records(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_game_records_start_time ON game_records(start_time);
CREATE INDEX IF NOT EXISTS idx_game_records_status ON game_records(status);
CREATE INDEX IF NOT EXISTS idx_game_records_user_status ON game_records(user_id, status);

-- 創建更新時間觸發器
CREATE OR REPLACE FUNCTION update_game_records_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_game_records_updated_at ON game_records;
CREATE TRIGGER trigger_update_game_records_updated_at
    BEFORE UPDATE ON game_records
    FOR EACH ROW
    EXECUTE FUNCTION update_game_records_updated_at();

-- 添加註釋
COMMENT ON TABLE game_records IS '遊戲會話記錄表，記錄玩家每次遊戲的完整統計信息';
COMMENT ON COLUMN game_records.session_id IS '遊戲會話ID，可用於關聯同一次登入的多個遊戲局';
COMMENT ON COLUMN game_records.net_profit IS '淨盈虧 = total_wins - total_bets，正數表示盈利，負數表示虧損';
COMMENT ON COLUMN game_records.hit_rate IS '命中率 = (bullets_hit / bullets_fired) * 100';
COMMENT ON COLUMN game_records.metadata IS '額外的遊戲數據，例如：{"fish_types": {"small": 10, "medium": 5}, "cannon_levels": [1, 2, 3]}';
