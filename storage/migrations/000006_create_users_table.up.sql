-- ⚠️ MIGRATION 6 已被合併到 MIGRATION 1
-- 此遷移文件保留僅用於維持遷移版本號的連續性
-- users 表的完整定義（包含遊客和第三方登入支持）現在位於:
-- storage/migrations/000001_create_initial_tables.up.sql
--
-- 原因：Migration 1 和 Migration 6 原本都創建 users 表，導致衝突
-- 解決方案：將完整的 users 表定義合併到 Migration 1
--
-- 此文件現在是一個空操作（no-op）遷移

-- 空操作：確保遷移系統可以順利執行
SELECT 1 AS migration_6_noop;
