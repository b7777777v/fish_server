// internal/data/wallet_repo_test.go
package data

import (
	"context"
	"os"
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/conf"
	"github.com/b7777777v/fish_server/internal/data/postgres"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 測試環境設置
func setupWalletRepoTest(t *testing.T) (*Data, wallet.WalletRepo, func()) {
	// 創建日誌記錄器
	log := logger.New(os.Stdout, "info", "console")

	// 連接測試數據庫
	dsns := []string{
		"host=localhost user=user password=password dbname=fish_db port=5432 sslmode=disable TimeZone=Asia/Shanghai", // docker-compose default
		"host=localhost user=postgres password=postgres dbname=fish_test port=5432 sslmode=disable TimeZone=Asia/Shanghai", // postgres default
		"host=localhost user=postgres password= dbname=fish_test port=5432 sslmode=disable TimeZone=Asia/Shanghai", // no password
	}
	
	var pgClient *postgres.Client
	var err error
	for _, dsn := range dsns {
		dbConfig := &conf.Database{
			Driver: "postgres",
			Source: dsn,
		}
		pgClient, err = postgres.NewClientFromDatabase(dbConfig, log)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		t.Skipf("Skipping test: no accessible PostgreSQL database found. Error: %v", err)
	}

	// 創建數據層
	data := &Data{db: pgClient}

	// 創建錢包存儲庫
	repo := NewWalletRepo(data, log)
	require.NotNil(t, repo)

	// 清理函數
	cleanup := func() {
		// 清理測試數據
		ctx := context.Background()
		_, err := data.db.Exec(ctx, "DELETE FROM wallet_transactions")
		require.NoError(t, err)
		_, err = data.db.Exec(ctx, "DELETE FROM wallets")
		require.NoError(t, err)
		_, err = data.db.Exec(ctx, "DELETE FROM users")
		require.NoError(t, err)

		// 關閉數據庫連接
		err = pgClient.Close()
		require.NoError(t, err)
	}

	// 預先清理數據
	cleanup()

	// 創建測試用戶
	ctx := context.Background()
	_, err = data.db.Exec(ctx, "INSERT INTO users (id, username, password_hash, email, status, created_at, updated_at) VALUES (1, 'testuser', 'hash', 'test@example.com', 1, NOW(), NOW())")
	require.NoError(t, err)

	return data, repo, cleanup
}

// TestCreateWallet 測試創建錢包
func TestCreateWallet(t *testing.T) {
	_, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()
	w := &wallet.Wallet{
		UserID:   1,
		Balance:  100.0,
		Currency: "CNY",
		Status:   1,
	}

	// 創建錢包
	err := repo.Create(ctx, w)
	assert.NoError(t, err)
	assert.NotZero(t, w.ID)
	assert.Equal(t, uint(1), w.UserID)
	assert.Equal(t, 100.0, w.Balance)
	assert.Equal(t, "USD", w.Currency)
	assert.Equal(t, 1, w.Status)
}

// TestFindByID 測試通過ID查找錢包
func TestFindByID(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (100, 1, 200.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 查找錢包
	w, err := repo.FindByID(ctx, 100)
	assert.NoError(t, err)
	assert.NotNil(t, w)
	assert.Equal(t, uint(100), w.ID)
	assert.Equal(t, uint(1), w.UserID)
	assert.Equal(t, 200.0, w.Balance)
	assert.Equal(t, "CNY", w.Currency)
	assert.Equal(t, int8(1), w.Status)
}

// TestFindByUserID 測試通過用戶ID查找錢包
func TestFindByUserID(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (101, 1, 300.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 查找錢包
	foundWallet, err := repo.FindByUserID(ctx, 1, "USD")
	assert.NoError(t, err)
	assert.NotNil(t, foundWallet)
	assert.Equal(t, uint(1), foundWallet.UserID)
}

// TestUpdate 測試更新錢包
func TestUpdate(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (102, 1, 400.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 查找錢包
	w, err := repo.FindByID(ctx, 102)
	assert.NoError(t, err)

	// 更新錢包
	w.Balance = 500.0
	w.Status = 0
	err = repo.Update(ctx, w)
	assert.NoError(t, err)
	assert.Equal(t, 500.0, w.Balance)
	assert.Equal(t, 0, w.Status)

	// 再次查找確認更新
	checkWallet, err := repo.FindByID(ctx, 102)
	assert.NoError(t, err)
	assert.Equal(t, 500.0, checkWallet.Balance)
	assert.Equal(t, 0, checkWallet.Status)
}

// TestDeposit 測試存款
func TestDeposit(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (103, 1, 100.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 存款
	metadata := make(map[string]interface{})
	metadata["note"] = "test deposit"
	err = repo.Deposit(ctx, 103, 50.0, "deposit", "test-ref", "Test deposit", metadata)
	assert.NoError(t, err)

	// 檢查錢包餘額
	w, err := repo.FindByID(ctx, 103)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, w.Balance)
}

// TestWithdraw 測試取款
func TestWithdraw(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (104, 1, 200.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 取款
	metadata2 := make(map[string]interface{})
	metadata2["note"] = "test withdraw"
	err = repo.Withdraw(ctx, 104, 50.0, "withdraw", "test-ref", "Test withdraw", metadata2)
	assert.NoError(t, err)

	// 檢查錢包餘額
	w, err := repo.FindByID(ctx, 104)
	assert.NoError(t, err)
	assert.Equal(t, 150.0, w.Balance) // 200 - 50 = 150
}

// TestWithdrawInsufficientFunds 測試餘額不足的取款
func TestWithdrawInsufficientFunds(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (105, 1, 50.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 嘗試取款超過餘額
	metadata3 := make(map[string]interface{})
	err = repo.Withdraw(ctx, 105, 100.0, "withdraw", "test-ref", "Test withdraw", metadata3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient")

	// 檢查錢包餘額未變
	w, err := repo.FindByID(ctx, 105)
	assert.NoError(t, err)
	assert.Equal(t, 50.0, w.Balance)
}

// TestCreateTransaction 測試創建交易記錄
func TestCreateTransaction(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (106, 1, 300.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 創建交易記錄
	tx := &wallet.Transaction{
		WalletID:      106,
		Amount:        100.0,
		BalanceBefore: 300.0,
		BalanceAfter:  400.0,
		Type:          "deposit",
		Status:        1,
		ReferenceID:   "test-ref",
		Description:   "Test transaction",
		Metadata:      map[string]interface{}{"test": "data"},
	}

	err = repo.CreateTransaction(ctx, tx)
	assert.NoError(t, err)
	assert.NotZero(t, tx.ID)
	assert.Equal(t, uint(106), tx.WalletID)
	assert.Equal(t, 75.0, tx.Amount)
	assert.Equal(t, "manual", tx.Type)
	assert.Equal(t, 1, tx.Status)
}

// TestFindTransactionsByWalletID 測試查找錢包交易記錄
func TestFindTransactionsByWalletID(t *testing.T) {
	data, repo, cleanup := setupWalletRepoTest(t)
	defer cleanup()

	ctx := context.Background()

	// 創建測試錢包
	_, err := data.db.Exec(ctx, "INSERT INTO wallets (id, user_id, balance, currency, status, created_at, updated_at) VALUES (107, 1, 500.0, 'CNY', 1, NOW(), NOW())")
	require.NoError(t, err)

	// 創建測試交易記錄
	_, err = data.db.Exec(ctx, `
		INSERT INTO wallet_transactions 
		(wallet_id, amount, balance_before, balance_after, type, status, reference_id, description, metadata, created_at, updated_at) 
		VALUES 
		(107, 100.0, 400.0, 500.0, 'deposit', 1, 'ref1', 'Test 1', '{}', NOW(), NOW()),
		(107, 50.0, 500.0, 450.0, 'withdraw', 1, 'ref2', 'Test 2', '{}', NOW(), NOW())
	`)
	require.NoError(t, err)

	// 查找交易記錄
	txs, err := repo.FindTransactionsByWalletID(ctx, 107, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, txs, 2)
	assert.Equal(t, uint(107), txs[0].WalletID)
}