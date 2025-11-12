-- 創建陣型配置表
CREATE TABLE IF NOT EXISTS formation_configs (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(50) NOT NULL UNIQUE, -- 配置鍵（例如：'default', 'easy', 'hard'）
    config_data JSONB NOT NULL, -- 配置數據（JSON格式）
    description TEXT, -- 配置說明
    is_active BOOLEAN NOT NULL DEFAULT true, -- 是否啟用
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引（使用 IF NOT EXISTS 確保冪等性）
CREATE INDEX IF NOT EXISTS idx_formation_configs_key ON formation_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_formation_configs_active ON formation_configs(is_active);
CREATE INDEX IF NOT EXISTS idx_formation_configs_data ON formation_configs USING GIN (config_data);

-- 創建更新時間觸發器函數
CREATE OR REPLACE FUNCTION update_formation_configs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 創建觸發器
CREATE TRIGGER trigger_update_formation_configs_updated_at
    BEFORE UPDATE ON formation_configs
    FOR EACH ROW
    EXECUTE FUNCTION update_formation_configs_updated_at();

-- 插入默認配置
INSERT INTO formation_configs (config_key, config_data, description, is_active)
VALUES (
    'default',
    '{
        "enabled": true,
        "min_interval": 20000000000,
        "max_interval": 60000000000,
        "base_spawn_chance": 0.3,
        "formation_weights": {
            "v": 0.25,
            "line": 0.20,
            "circle": 0.15,
            "triangle": 0.15,
            "diamond": 0.10,
            "wave": 0.10,
            "spiral": 0.05
        },
        "min_fish_count": 5,
        "max_fish_count": 20,
        "fish_count_by_formation": {
            "v": {"min": 5, "max": 12},
            "line": {"min": 4, "max": 10},
            "circle": {"min": 6, "max": 15},
            "triangle": {"min": 6, "max": 14},
            "diamond": {"min": 5, "max": 11},
            "wave": {"min": 8, "max": 19},
            "spiral": {"min": 10, "max": 20}
        },
        "route_preferences": {
            "straight": 0.30,
            "curved": 0.35,
            "zigzag": 0.15,
            "circular": 0.15,
            "random": 0.05
        },
        "allow_random_route": true,
        "fish_size_preferences": {
            "small": 0.50,
            "medium": 0.35,
            "large": 0.12,
            "boss": 0.03
        },
        "uniform_type_chance": 0.7,
        "max_concurrent_formations": 3,
        "dynamic_difficulty": true,
        "special_event_multiplier": 1.0
    }'::jsonb,
    '默認陣型配置（普通難度）',
    true
)
ON CONFLICT (config_key) DO NOTHING;
