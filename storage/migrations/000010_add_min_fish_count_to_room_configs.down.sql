-- 回滾：移除 min_fish_count 欄位

ALTER TABLE room_configs
    DROP COLUMN IF EXISTS min_fish_count;
