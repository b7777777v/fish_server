-- 添加 min_fish_count 欄位到 room_configs 表
-- 此欄位用於確保房間內魚數量不會低於最小值

ALTER TABLE room_configs
    ADD COLUMN IF NOT EXISTS min_fish_count INT NOT NULL DEFAULT 10;

COMMENT ON COLUMN room_configs.min_fish_count IS '最小魚數量（低於此值將強制補充魚群）';

-- 更新現有房間配置的 min_fish_count 值（設置為 max_fish_count 的 50%）
UPDATE room_configs SET min_fish_count = 10 WHERE room_type = 'novice';
UPDATE room_configs SET min_fish_count = 12 WHERE room_type = 'intermediate';
UPDATE room_configs SET min_fish_count = 15 WHERE room_type = 'advanced';
UPDATE room_configs SET min_fish_count = 18 WHERE room_type = 'vip';
