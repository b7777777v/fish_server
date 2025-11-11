-- 创建房间配置表
-- 存储不同类型房间的配置信息

CREATE TABLE IF NOT EXISTS room_configs (
    id BIGSERIAL PRIMARY KEY,
    room_type VARCHAR(50) NOT NULL UNIQUE, -- 房间类型（novice, intermediate, advanced, vip）
    room_name VARCHAR(100) NOT NULL, -- 房间名称
    max_players INT NOT NULL DEFAULT 4, -- 最大玩家数（默认4人）
    min_bet BIGINT NOT NULL DEFAULT 10, -- 最小下注（单位：分）
    max_bet BIGINT NOT NULL DEFAULT 1000, -- 最大下注（单位：分）
    entry_fee BIGINT NOT NULL DEFAULT 0, -- 进入房间所需金币（单位：分）
    bullet_cost_multiplier DECIMAL(10, 2) NOT NULL DEFAULT 1.0, -- 子弹成本倍数
    fish_spawn_rate DECIMAL(10, 2) NOT NULL DEFAULT 1.0, -- 鱼类生成率
    max_fish_count INT NOT NULL DEFAULT 50, -- 最大鱼数量
    room_width DECIMAL(10, 2) NOT NULL DEFAULT 1920.0, -- 房间宽度
    room_height DECIMAL(10, 2) NOT NULL DEFAULT 1080.0, -- 房间高度
    target_rtp DECIMAL(5, 4) NOT NULL DEFAULT 0.96, -- 目标RTP (0.96 = 96%)
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- 是否启用
    description TEXT, -- 房间描述
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_room_configs_room_type ON room_configs(room_type);
CREATE INDEX idx_room_configs_is_active ON room_configs(is_active);

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_room_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_room_configs_updated_at
    BEFORE UPDATE ON room_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_room_configs_updated_at();

-- 插入默认房间配置
INSERT INTO room_configs (room_type, room_name, max_players, min_bet, max_bet, entry_fee, target_rtp, description) VALUES
('novice', '新手房', 4, 10, 100, 0, 0.98, '适合新手玩家，入场免费，低倍率'),
('intermediate', '中级房', 4, 50, 500, 1000, 0.96, '适合有经验的玩家，需要1000金币入场'),
('advanced', '高级房', 4, 100, 2000, 5000, 0.95, '高倍率房间，需要5000金币入场'),
('vip', 'VIP房', 4, 500, 10000, 10000, 0.94, 'VIP专属房间，高风险高回报');
