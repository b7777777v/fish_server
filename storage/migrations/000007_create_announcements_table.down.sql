-- TODO: 大廳模組 - 刪除公告表

-- 刪除觸發器
DROP TRIGGER IF EXISTS trigger_update_announcements_updated_at ON announcements;
DROP FUNCTION IF EXISTS update_announcements_updated_at();

-- 刪除表
DROP TABLE IF EXISTS announcements;
