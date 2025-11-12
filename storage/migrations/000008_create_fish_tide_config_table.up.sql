-- TODO: 魚潮系統 - 創建魚潮配置表
-- 此表儲存魚潮事件的配置資訊

CREATE TABLE IF NOT EXISTS fish_tide_config (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL, -- 魚潮名稱
    description TEXT, -- 魚潮描述
    fish_type_id INT NOT NULL, -- 魚種 ID（關聯到 fish_types 表）
    fish_count INT NOT NULL, -- 魚的數量
    duration_seconds INT NOT NULL, -- 持續時間（秒）
    spawn_interval_ms INT NOT NULL, -- 生成間隔（毫秒）
    speed_multiplier FLOAT NOT NULL DEFAULT 1.0, -- 速度倍率
    trigger_rule VARCHAR(50) NOT NULL, -- 觸發規則：'fixed_time'（固定時間）, 'random'（隨機）, 'manual'（手動）
    trigger_config JSONB, -- 觸發配置（JSON 格式，如固定時間的 cron 表達式）
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- 是否啟用
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引（使用 IF NOT EXISTS 確保冪等性）
CREATE INDEX IF NOT EXISTS idx_fish_tide_config_is_active ON fish_tide_config(is_active);
CREATE INDEX IF NOT EXISTS idx_fish_tide_config_fish_type ON fish_tide_config(fish_type_id);
CREATE INDEX IF NOT EXISTS idx_fish_tide_config_trigger_rule ON fish_tide_config(trigger_rule);

-- 添加外鍵約束（關聯到 fish_types 表）
ALTER TABLE fish_tide_config
    ADD CONSTRAINT fk_fish_tide_fish_type
    FOREIGN KEY (fish_type_id)
    REFERENCES fish_types(id)
    ON DELETE CASCADE;

-- 創建更新時間觸發器
CREATE OR REPLACE FUNCTION update_fish_tide_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_fish_tide_config_updated_at
    BEFORE UPDATE ON fish_tide_config
    FOR EACH ROW
    EXECUTE FUNCTION update_fish_tide_config_updated_at();

-- 插入示例魚潮配置
-- 假設已有魚種 ID 5 為高價值魚種
INSERT INTO fish_tide_config (name, description, fish_type_id, fish_count, duration_seconds, spawn_interval_ms, speed_multiplier, trigger_rule, trigger_config, is_active) VALUES
('黃金魚潮', '大量黃金魚快速游過螢幕，持續 30 秒', 5, 100, 30, 300, 1.5, 'random', '{"min_interval_minutes": 30, "max_interval_minutes": 60}', TRUE),
('午間魚潮', '每天中午 12 點觸發的特殊魚潮', 5, 150, 60, 200, 2.0, 'fixed_time', '{"cron": "0 12 * * *"}', TRUE);
