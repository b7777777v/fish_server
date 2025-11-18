-- 刪除觸發器
DROP TRIGGER IF EXISTS trigger_update_game_records_updated_at ON game_records;

-- 刪除函數
DROP FUNCTION IF EXISTS update_game_records_updated_at();

-- 刪除索引（表刪除時會自動刪除，但明確列出更清晰）
DROP INDEX IF EXISTS idx_game_records_user_id;
DROP INDEX IF EXISTS idx_game_records_room_id;
DROP INDEX IF EXISTS idx_game_records_session_id;
DROP INDEX IF EXISTS idx_game_records_start_time;
DROP INDEX IF EXISTS idx_game_records_status;
DROP INDEX IF EXISTS idx_game_records_user_status;

-- 刪除表
DROP TABLE IF EXISTS game_records;
