-- 創建魚類類型表
CREATE TABLE IF NOT EXISTS fish_types (
    id INT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    size VARCHAR(20) NOT NULL, -- 'small', 'medium', 'large', 'boss'
    base_health INT NOT NULL,
    base_value BIGINT NOT NULL,
    base_speed FLOAT NOT NULL,
    rarity FLOAT NOT NULL, -- 稀有度 (0.0 - 1.0)
    hit_rate FLOAT NOT NULL, -- 基礎命中率 (0.0 - 1.0)
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 創建索引（使用 IF NOT EXISTS 確保冪等性）
CREATE INDEX IF NOT EXISTS idx_fish_types_size ON fish_types(size);
CREATE INDEX IF NOT EXISTS idx_fish_types_rarity ON fish_types(rarity);
