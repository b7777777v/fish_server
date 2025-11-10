-- TODO: 魚潮系統 - 刪除魚潮配置表

-- 刪除觸發器
DROP TRIGGER IF EXISTS trigger_update_fish_tide_config_updated_at ON fish_tide_config;
DROP FUNCTION IF EXISTS update_fish_tide_config_updated_at();

-- 刪除表
DROP TABLE IF EXISTS fish_tide_config;
