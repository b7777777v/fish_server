-- TODO: 帳號模組 - 刪除使用者表

-- 刪除觸發器
DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_users_updated_at();

-- 刪除表
DROP TABLE IF EXISTS users;
