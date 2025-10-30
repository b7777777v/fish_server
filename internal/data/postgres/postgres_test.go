// internal/data/postgres/postgres_test.go
package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 測試環境配置
var (
	testClient *Client
	testLogger logger.Logger
)

// 設置測試環境
func setupTestDB(t *testing.T) {
	// 使用環境變量或默認值設置測試數據庫連接
	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		// Try different common configurations
		testDSNs := []string{
			"host=localhost user=user password=password dbname=fish_db port=5432 sslmode=disable TimeZone=Asia/Shanghai", // docker-compose default
			"host=localhost user=postgres password=postgres dbname=fish_test port=5432 sslmode=disable TimeZone=Asia/Shanghai", // postgres default
			"host=localhost user=postgres password= dbname=fish_test port=5432 sslmode=disable TimeZone=Asia/Shanghai", // no password
		}
		
		// Try to connect with each DSN
		for _, testDSN := range testDSNs {
			dbConfig := &conf.Database{
				Driver: "postgres",
				Source: testDSN,
			}
			client, err := NewClientFromDatabase(dbConfig, logger.New(os.Stdout, "error", "console"))
			if err == nil {
				client.Close()
				dsn = testDSN
				break
			}
		}
		
		if dsn == "" {
			t.Skip("Skipping test: no accessible PostgreSQL database found. Please start PostgreSQL or set TEST_POSTGRES_DSN environment variable.")
		}
	}

	// 創建日誌記錄器
	testLogger = logger.New(os.Stdout, "info", "console")

	// 創建客戶端
	dbConfig := &conf.Database{
		Driver: "postgres",
		Source: dsn,
	}
	var err error
	testClient, err = NewClientFromDatabase(dbConfig, testLogger)
	require.NoError(t, err)
	require.NotNil(t, testClient)

	// 創建測試表
	createTestTables(t)

	// 清理測試數據
	cleanTestData(t)
}

// 創建測試表
func createTestTables(t *testing.T) {
	ctx := context.Background()
	
	// 創建用戶表
	_, err := testClient.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL UNIQUE,
			password_hash VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE,
			nickname VARCHAR(50),
			avatar_url VARCHAR(255),
			status SMALLINT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// 創建錢包表
	_, err = testClient.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wallets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id),
			balance DECIMAL(20,2) NOT NULL DEFAULT 0.00,
			currency VARCHAR(10) NOT NULL DEFAULT 'CNY',
			status SMALLINT NOT NULL DEFAULT 1,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	// 創建交易表
	_, err = testClient.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS wallet_transactions (
			id SERIAL PRIMARY KEY,
			wallet_id INTEGER NOT NULL REFERENCES wallets(id),
			amount DECIMAL(20,2) NOT NULL,
			balance_before DECIMAL(20,2) NOT NULL,
			balance_after DECIMAL(20,2) NOT NULL,
			type VARCHAR(20) NOT NULL,
			status SMALLINT NOT NULL DEFAULT 1,
			reference_id VARCHAR(100),
			description TEXT,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)
}

// 清理測試數據
func cleanTestData(t *testing.T) {
	ctx := context.Background()
	
	// 按照依賴關係順序刪除數據
	_, err := testClient.Exec(ctx, "DELETE FROM wallet_transactions")
	require.NoError(t, err)

	_, err = testClient.Exec(ctx, "DELETE FROM wallets")
	require.NoError(t, err)

	_, err = testClient.Exec(ctx, "DELETE FROM users")
	require.NoError(t, err)
}

// 測試結束後清理資源
func teardownTestDB(t *testing.T) {
	if testClient != nil {
		err := testClient.Close()
		require.NoError(t, err)
	}
}

// TestMain 設置和清理測試環境
func TestMain(m *testing.M) {
	// 運行測試
	code := m.Run()

	// 退出
	os.Exit(code)
}

// TestNewClient 測試創建新的PostgreSQL客戶端
func TestNewClient(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	assert.NotNil(t, testClient)
	assert.NotNil(t, testClient.Pool)
}

// TestUserCRUD 測試用戶表的CRUD操作
func TestUserCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	// 創建用戶
	var userID int
	err := testClient.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, email, nickname, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, "testuser", "hashedpassword", "test@example.com", "Test User", 1).Scan(&userID)
	assert.NoError(t, err)
	assert.NotZero(t, userID)

	// 讀取
	var username, email, nickname string
	err = testClient.QueryRow(ctx, `
		SELECT username, email, nickname FROM users WHERE id = $1
	`, userID).Scan(&username, &email, &nickname)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", username)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "Test User", nickname)

	// 更新
	newNickname := "Updated User"
	_, err = testClient.Exec(ctx, `
		UPDATE users SET nickname = $1 WHERE id = $2
	`, newNickname, userID)
	assert.NoError(t, err)

	// 驗證更新
	err = testClient.QueryRow(ctx, `
		SELECT nickname FROM users WHERE id = $1
	`, userID).Scan(&nickname)
	assert.NoError(t, err)
	assert.Equal(t, newNickname, nickname)

	// 刪除
	_, err = testClient.Exec(ctx, `
		DELETE FROM users WHERE id = $1
	`, userID)
	assert.NoError(t, err)

	// 驗證刪除
	var count int
	err = testClient.QueryRow(ctx, `
		SELECT COUNT(*) FROM users WHERE id = $1
	`, userID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestWalletCRUD 測試錢包表的CRUD操作
func TestWalletCRUD(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	// 創建用戶
	var userID int
	err := testClient.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, email, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "walletuser", "hashedpassword", "wallet@example.com", 1).Scan(&userID)
	assert.NoError(t, err)

	// 創建錢包
	var walletID int
	err = testClient.QueryRow(ctx, `
		INSERT INTO wallets (user_id, balance, currency, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, 100.00, "CNY", 1).Scan(&walletID)
	assert.NoError(t, err)
	assert.NotZero(t, walletID)

	// 讀取
	var userIDFromDB int
	var balance float64
	var currency string
	err = testClient.QueryRow(ctx, `
		SELECT user_id, balance, currency FROM wallets WHERE id = $1
	`, walletID).Scan(&userIDFromDB, &balance, &currency)
	assert.NoError(t, err)
	assert.Equal(t, userID, userIDFromDB)
	assert.Equal(t, 100.00, balance)
	assert.Equal(t, "CNY", currency)

	// 更新
	newBalance := 200.00
	_, err = testClient.Exec(ctx, `
		UPDATE wallets SET balance = $1 WHERE id = $2
	`, newBalance, walletID)
	assert.NoError(t, err)

	// 驗證更新
	err = testClient.QueryRow(ctx, `
		SELECT balance FROM wallets WHERE id = $1
	`, walletID).Scan(&balance)
	assert.NoError(t, err)
	assert.Equal(t, newBalance, balance)

	// 刪除
	_, err = testClient.Exec(ctx, `
		DELETE FROM wallets WHERE id = $1
	`, walletID)
	assert.NoError(t, err)

	// 驗證刪除
	var count int
	err = testClient.QueryRow(ctx, `
		SELECT COUNT(*) FROM wallets WHERE id = $1
	`, walletID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestTransactionAndConcurrency 測試事務和並發操作
func TestTransactionAndConcurrency(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	// 創建用戶
	var userID int
	err := testClient.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, email, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "txuser", "hashedpassword", "tx@example.com", 1).Scan(&userID)
	assert.NoError(t, err)

	// 創建錢包
	var walletID int
	err = testClient.QueryRow(ctx, `
		INSERT INTO wallets (user_id, balance, currency, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, 1000.00, "CNY", 1).Scan(&walletID)
	assert.NoError(t, err)

	// 測試事務 - 成功場景
	tx, err := testClient.Begin(ctx)
	assert.NoError(t, err)

	// 更新錢包餘額
	_, err = tx.Exec(ctx, `
		UPDATE wallets SET balance = balance + $1 WHERE id = $2
	`, 100.0, walletID)
	assert.NoError(t, err)

	// 創建交易記錄
	_, err = tx.Exec(ctx, `
		INSERT INTO wallet_transactions (wallet_id, amount, balance_before, balance_after, type, status, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, walletID, 100.0, 1000.0, 1100.0, "deposit", 1, "Test deposit")
	assert.NoError(t, err)

	// 提交事務
	err = tx.Commit(ctx)
	assert.NoError(t, err)

	// 驗證餘額更新
	var updatedBalance float64
	err = testClient.QueryRow(ctx, `
		SELECT balance FROM wallets WHERE id = $1
	`, walletID).Scan(&updatedBalance)
	assert.NoError(t, err)
	assert.Equal(t, 1100.00, updatedBalance)

	// 驗證交易記錄
	var txAmount float64
	err = testClient.QueryRow(ctx, `
		SELECT amount FROM wallet_transactions WHERE wallet_id = $1 AND type = $2
	`, walletID, "deposit").Scan(&txAmount)
	assert.NoError(t, err)
	assert.Equal(t, 100.0, txAmount)

	// 測試事務 - 失敗場景
	tx, err = testClient.Begin(ctx)
	assert.NoError(t, err)

	// 更新錢包餘額
	_, err = tx.Exec(ctx, `
		UPDATE wallets SET balance = balance + $1 WHERE id = $2
	`, 200.0, walletID)
	assert.NoError(t, err)

	// 回滾事務
	err = tx.Rollback(ctx)
	assert.NoError(t, err)

	// 驗證餘額未變化（事務回滾）
	var balanceAfterRollback float64
	err = testClient.QueryRow(ctx, `
		SELECT balance FROM wallets WHERE id = $1
	`, walletID).Scan(&balanceAfterRollback)
	assert.NoError(t, err)
	assert.Equal(t, 1100.00, balanceAfterRollback) // 仍然是1100，而不是1300
}

// TestDatabaseConnection 測試數據庫連接池設置
func TestDatabaseConnection(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// 驗證連接池設置
	assert.Equal(t, int32(100), testClient.Pool.Config().MaxConns)
	assert.Equal(t, int32(10), testClient.Pool.Config().MinConns)

	// 測試連接是否正常工作
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := testClient.Pool.Ping(ctx)
	assert.NoError(t, err)
}

// TestConcurrentWalletOperations 測試並發錢包操作
func TestConcurrentWalletOperations(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	ctx := context.Background()

	// 創建用戶
	var userID int
	err := testClient.QueryRow(ctx, `
		INSERT INTO users (username, password_hash, email, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, "concurrentuser", "hashedpassword", "concurrent@example.com", 1).Scan(&userID)
	assert.NoError(t, err)

	// 創建錢包
	var walletID int
	err = testClient.QueryRow(ctx, `
		INSERT INTO wallets (user_id, balance, currency, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, userID, 1000.00, "CNY", 1).Scan(&walletID)
	assert.NoError(t, err)

	// 執行並發操作的函數
	updateWallet := func(amount float64, done chan bool) {
		ctx := context.Background()
		
		// 開始事務
		tx, err := testClient.Begin(ctx)
		if err != nil {
			t.Logf("Error starting transaction: %v", err)
			done <- false
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback(ctx)
			}
		}()

		// 使用FOR UPDATE鎖定行
		var currentBalance float64
		err = tx.QueryRow(ctx, `
			SELECT balance FROM wallets WHERE id = $1 FOR UPDATE
		`, walletID).Scan(&currentBalance)
		if err != nil {
			t.Logf("Error selecting wallet for update: %v", err)
			done <- false
			return
		}

		// 更新餘額
		newBalance := currentBalance + amount
		_, err = tx.Exec(ctx, `
			UPDATE wallets SET balance = $1 WHERE id = $2
		`, newBalance, walletID)
		if err != nil {
			t.Logf("Error updating wallet balance: %v", err)
			done <- false
			return
		}

		// 創建交易記錄
		_, err = tx.Exec(ctx, `
			INSERT INTO wallet_transactions (wallet_id, amount, balance_before, balance_after, type, status, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, walletID, amount, currentBalance, newBalance, "update", 1, "Concurrent update")
		if err != nil {
			t.Logf("Error creating transaction record: %v", err)
			done <- false
			return
		}

		// 提交事務
		err = tx.Commit(ctx)
		if err != nil {
			t.Logf("Error committing transaction: %v", err)
			done <- false
			return
		}

		done <- true
	}

	// 並發執行10個更新操作
	done := make(chan bool, 10)
	for i := 0; i < 5; i++ {
		go updateWallet(10.0, done)  // 5個存款操作
		go updateWallet(-5.0, done)  // 5個取款操作
	}

	// 等待所有操作完成
	successCount := 0
	for i := 0; i < 10; i++ {
		if <-done {
			successCount++
		}
	}
	assert.Equal(t, 10, successCount, "All concurrent operations should succeed")

	// 驗證最終餘額
	var finalBalance float64
	err = testClient.QueryRow(ctx, `
		SELECT balance FROM wallets WHERE id = $1
	`, walletID).Scan(&finalBalance)
	assert.NoError(t, err)
	assert.Equal(t, 1025.00, finalBalance) // 1000 + 5*10 - 5*5 = 1025

	// 驗證交易記錄數量
	var count int
	err = testClient.QueryRow(ctx, `
		SELECT COUNT(*) FROM wallet_transactions WHERE wallet_id = $1
	`, walletID).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 10, count)
}