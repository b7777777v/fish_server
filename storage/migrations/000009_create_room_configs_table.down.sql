-- 删除房间配置表

-- 删除触发器
DROP TRIGGER IF EXISTS trigger_update_room_configs_updated_at ON room_configs;
DROP FUNCTION IF EXISTS update_room_configs_updated_at();

-- 删除表
DROP TABLE IF EXISTS room_configs;
