-- ⚠️ MIGRATION 6 DOWN 已被禁用
-- users 表的刪除操作現在由 Migration 1 的 down 腳本處理
-- storage/migrations/000001_create_initial_tables.down.sql
--
-- 此文件是一個空操作（no-op）遷移回滾

-- 空操作
SELECT 1 AS migration_6_down_noop;
