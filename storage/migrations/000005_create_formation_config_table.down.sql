-- 刪除觸發器
DROP TRIGGER IF EXISTS trigger_update_formation_configs_updated_at ON formation_configs;

-- 刪除觸發器函數
DROP FUNCTION IF EXISTS update_formation_configs_updated_at();

-- 刪除表
DROP TABLE IF EXISTS formation_configs;
