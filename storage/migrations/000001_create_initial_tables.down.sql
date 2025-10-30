-- 按照依賴關係的相反順序刪除表

-- 首先刪除錢包交易記錄表
DROP TABLE IF EXISTS wallet_transactions;

-- 然後刪除錢包表
DROP TABLE IF EXISTS wallets;

-- 最後刪除用戶表
DROP TABLE IF EXISTS users;