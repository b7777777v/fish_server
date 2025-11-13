// internal/data/wallet_repo.go
package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/b7777777v/fish_server/internal/pkg/logger"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5"
)

// WalletPO 是錢包的持久化對象
type WalletPO struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	Status    int       `json:"status"` // 1: 正常, 0: 凍結
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TransactionPO 是交易記錄的持久化對象
type TransactionPO struct {
	ID            uint      `json:"id"`
	WalletID      uint      `json:"wallet_id"`
	Amount        float64   `json:"amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	Type          string    `json:"type"`
	Status        int       `json:"status"` // 1: 成功, 0: 失敗
	ReferenceID   string    `json:"reference_id"`
	Description   string    `json:"description"`
	Metadata      string    `json:"metadata"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type walletRepo struct {
	data   *Data
	logger logger.Logger
}

// NewWalletRepo 創建一個新的錢包儲存庫
func NewWalletRepo(data *Data, logger logger.Logger) wallet.WalletRepo {
	return &walletRepo{
		data:   data,
		logger: logger.With("module", "data/wallet_repo"),
	}
}

// po2do 將持久化對象轉換為領域對象
func (r *walletRepo) po2do(po *WalletPO) *wallet.Wallet {
	return &wallet.Wallet{
		ID:       po.ID,
		UserID:   po.UserID,
		Balance:  po.Balance,
		Currency: po.Currency,
		Status:   int8(po.Status),
	}
}

// do2po 將領域對象轉換為持久化對象
func (r *walletRepo) do2po(do *wallet.Wallet) *WalletPO {
	return &WalletPO{
		ID:        do.ID,
		UserID:    do.UserID,
		Balance:   do.Balance,
		Currency:  do.Currency,
		Status:    int(do.Status),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// txPo2do 將交易持久化對象轉換為領域對象
func (r *walletRepo) txPo2do(po *TransactionPO) *wallet.Transaction {
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(po.Metadata), &metadata); err != nil {
		// 如果解析失敗，使用空的 metadata
		metadata = make(map[string]interface{})
	}

	return &wallet.Transaction{
		ID:            po.ID,
		WalletID:      po.WalletID,
		Amount:        po.Amount,
		BalanceBefore: po.BalanceBefore,
		BalanceAfter:  po.BalanceAfter,
		Type:          po.Type,
		Status:        int8(po.Status),
		ReferenceID:   po.ReferenceID,
		Description:   po.Description,
		Metadata:      metadata,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

// txDo2po 將交易領域對象轉換為持久化對象
func (r *walletRepo) txDo2po(do *wallet.Transaction) (*TransactionPO, error) {
	metadataBytes, err := json.Marshal(do.Metadata)
	if err != nil {
		return nil, err
	}

	return &TransactionPO{
		ID:            do.ID,
		WalletID:      do.WalletID,
		Amount:        do.Amount,
		BalanceBefore: do.BalanceBefore,
		BalanceAfter:  do.BalanceAfter,
		Type:          do.Type,
		Status:        int(do.Status),
		ReferenceID:   do.ReferenceID,
		Description:   do.Description,
		Metadata:      string(metadataBytes),
		CreatedAt:     do.CreatedAt,
		UpdatedAt:     do.UpdatedAt,
	}, nil
}

// FindByID 根據ID查詢錢包
func (r *walletRepo) FindByID(ctx context.Context, id uint) (*wallet.Wallet, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("wallet:%d", id)
	walletJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		var w wallet.Wallet
		if err = json.Unmarshal([]byte(walletJSON), &w); err == nil {
			r.logger.Debugf("Cache hit for wallet: %d", id)
			return &w, nil
		}
		r.logger.Warnf("Failed to unmarshal wallet from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on FindByID: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for wallet: %d. Fetching from DB.", id)
	query := `SELECT id, user_id, balance, currency, status, created_at, updated_at FROM wallets WHERE id = $1`
	var po WalletPO
	err = r.data.db.QueryRow(ctx, query, id).Scan(
		&po.ID, &po.UserID, &po.Balance, &po.Currency, &po.Status, &po.CreatedAt, &po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("wallet not found")
		}
		r.logger.Errorf("failed to find wallet by id: %v", err)
		return nil, err
	}

	w := r.po2do(&po)

	// 3. 將數據寫入快取
	walletBytes, err := json.Marshal(w)
	if err != nil {
		r.logger.Warnf("Failed to marshal wallet for cache: %v", err)
	} else {
		if err = r.data.redis.Set(ctx, cacheKey, walletBytes, 5*time.Minute); err != nil {
			r.logger.Warnf("Failed to set wallet cache: %v", err)
		}
	}

	return w, nil
}

// FindByUserID 根據用戶ID查詢錢包
func (r *walletRepo) FindByUserID(ctx context.Context, userID uint, currency string) (*wallet.Wallet, error) {
	// 1. 從 Redis 讀取快取
	cacheKey := fmt.Sprintf("wallet:user_id:%d:currency:%s", userID, currency)
	walletJSON, err := r.data.redis.Get(ctx, cacheKey)
	if err == nil {
		var w wallet.Wallet
		if err = json.Unmarshal([]byte(walletJSON), &w); err == nil {
			r.logger.Debugf("Cache hit for wallet by user_id: %d", userID)
			return &w, nil
		}
		r.logger.Warnf("Failed to unmarshal wallet from cache: %v", err)
	}
	if err != redis.Nil {
		r.logger.Errorf("Redis error on FindByUserID: %v", err)
	}

	// 2. 快取未命中，從資料庫讀取
	r.logger.Debugf("Cache miss for wallet by user_id: %d. Fetching from DB.", userID)
	query := `SELECT id, user_id, balance, currency, status, created_at, updated_at FROM wallets WHERE user_id = $1 AND currency = $2`
	var po WalletPO
	err = r.data.db.QueryRow(ctx, query, userID, currency).Scan(
		&po.ID, &po.UserID, &po.Balance, &po.Currency, &po.Status, &po.CreatedAt, &po.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("wallet not found")
		}
		r.logger.Errorf("failed to find wallet by user_id and currency: %v", err)
		return nil, err
	}

	w := r.po2do(&po)

	// 3. 將數據寫入快取
	walletBytes, err := json.Marshal(w)
	if err != nil {
		r.logger.Warnf("Failed to marshal wallet for cache: %v", err)
	} else {
		if err = r.data.redis.Set(ctx, cacheKey, walletBytes, 5*time.Minute); err != nil {
			r.logger.Warnf("Failed to set wallet cache: %v", err)
		}
	}

	return w, nil
}

// FindAllByUserID 根據用戶ID查詢所有錢包
func (r *walletRepo) FindAllByUserID(ctx context.Context, userID uint) ([]*wallet.Wallet, error) {
	r.logger.Debugf("Fetching all wallets for user_id: %d from DB", userID)

	query := `SELECT id, user_id, balance, currency, status, created_at, updated_at FROM wallets WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.data.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Errorf("failed to query wallets by user_id: %v", err)
		return nil, err
	}
	defer rows.Close()

	var wallets []*wallet.Wallet
	for rows.Next() {
		var po WalletPO
		err := rows.Scan(&po.ID, &po.UserID, &po.Balance, &po.Currency, &po.Status, &po.CreatedAt, &po.UpdatedAt)
		if err != nil {
			r.logger.Errorf("failed to scan wallet row: %v", err)
			return nil, err
		}
		wallets = append(wallets, r.po2do(&po))
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("error iterating wallet rows: %v", err)
		return nil, err
	}

	return wallets, nil
}

// Create 創建錢包
func (r *walletRepo) Create(ctx context.Context, w *wallet.Wallet) error {
	// 準備SQL查詢
	query := `
		INSERT INTO wallets (user_id, balance, currency, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id
	`

	// 執行查詢
	po := r.do2po(w)
	err := r.data.db.QueryRow(
		ctx,
		query,
		po.UserID, po.Balance, po.Currency, po.Status, po.CreatedAt, po.UpdatedAt,
	).Scan(&w.ID)

	if err != nil {
		r.logger.Errorf("failed to create wallet: %v", err)
		return err
	}

	return nil
}

// Update 更新錢包
func (r *walletRepo) Update(ctx context.Context, w *wallet.Wallet) error {
	// 準備SQL查詢
	query := `
		UPDATE wallets 
		SET balance = $1, currency = $2, status = $3, updated_at = $4 
		WHERE id = $5
	`

	// 更新時間
	updatedAt := time.Now()

	// 執行查詢
	result, err := r.data.db.Exec(
		ctx,
		query,
		w.Balance, w.Currency, w.Status, updatedAt, w.ID,
	)

	if err != nil {
		r.logger.Errorf("failed to update wallet: %v", err)
		return err
	}

	// 檢查是否有記錄被更新
	if result.RowsAffected() == 0 {
		return fmt.Errorf("wallet with id %d not found", w.ID)
	}

	// 更新對象的 UpdatedAt 字段
	w.UpdatedAt = updatedAt

	// 操作成功後，使快取失效
	cacheKeyByID := fmt.Sprintf("wallet:%d", w.ID)
	if err := r.data.redis.Del(ctx, cacheKeyByID); err != nil {
		r.logger.Warnf("Failed to delete wallet cache by id: %v", err)
	}

	cacheKeyByUserID := fmt.Sprintf("wallet:user_id:%d:currency:%s", w.UserID, w.Currency)
	if err := r.data.redis.Del(ctx, cacheKeyByUserID); err != nil {
		r.logger.Warnf("Failed to delete wallet cache by user id: %v", err)
	}

	return nil
}

// CreateTransaction 創建交易記錄
func (r *walletRepo) CreateTransaction(ctx context.Context, tx *wallet.Transaction) error {
	// 將metadata轉換為JSON字符串
	metadataBytes, err := json.Marshal(tx.Metadata)
	if err != nil {
		r.logger.Errorf("failed to marshal metadata: %v", err)
		return err
	}

	// 準備SQL查詢
	query := `
		INSERT INTO wallet_transactions (
			wallet_id, amount, balance_before, balance_after, 
			type, status, reference_id, description, metadata, 
			created_at, updated_at
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id
	`

	// 設置創建和更新時間
	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	// 執行查詢
	err = r.data.db.QueryRow(
		ctx,
		query,
		tx.WalletID, tx.Amount, tx.BalanceBefore, tx.BalanceAfter,
		tx.Type, tx.Status, tx.ReferenceID, tx.Description, string(metadataBytes),
		tx.CreatedAt, tx.UpdatedAt,
	).Scan(&tx.ID)

	if err != nil {
		r.logger.Errorf("failed to create transaction: %v", err)
		return err
	}

	return nil
}

// FindTransactionsByWalletID 查詢錢包的交易記錄
// TODO: [Cache] Caching transaction history can improve performance for frequently accessed pages.
// However, this is more complex than caching a single entity.
// The cache key should include pagination details (e.g., `transactions:wallet_id:{wallet_id}:page:{page_num}`).
// CRITICAL: This cache MUST be invalidated every time a new transaction is created for this wallet (e.g., after Deposit or Withdraw).
// A short TTL (e.g., 1-2 minutes) might be a safer strategy here.
func (r *walletRepo) FindTransactionsByWalletID(ctx context.Context, walletID uint, limit, offset int) ([]*wallet.Transaction, error) {

	// 準備SQL查詢
	query := `
		SELECT 
			id, wallet_id, amount, balance_before, balance_after, 
			type, status, reference_id, description, metadata, 
			created_at, updated_at 
		FROM wallet_transactions 
		WHERE wallet_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	// 執行查詢
	rows, err := r.data.db.Query(ctx, query, walletID, limit, offset)
	if err != nil {
		r.logger.Errorf("failed to find transactions: %v", err)
		return nil, err
	}
	defer rows.Close()

	// 處理結果
	var transactions []*wallet.Transaction
	for rows.Next() {
		var po TransactionPO
		if err := rows.Scan(
			&po.ID, &po.WalletID, &po.Amount, &po.BalanceBefore, &po.BalanceAfter,
			&po.Type, &po.Status, &po.ReferenceID, &po.Description, &po.Metadata,
			&po.CreatedAt, &po.UpdatedAt,
		); err != nil {
			r.logger.Errorf("failed to scan transaction row: %v", err)
			return nil, err
		}

		transactions = append(transactions, r.txPo2do(&po))
	}

	if err := rows.Err(); err != nil {
		r.logger.Errorf("error iterating transaction rows: %v", err)
		return nil, err
	}

	return transactions, nil
}

// Deposit 存款操作
func (r *walletRepo) Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	// 開始事務
	tx, err := r.data.db.Begin(ctx)
	if err != nil {
		r.logger.Errorf("failed to begin transaction: %v", err)
		return err
	}
	defer tx.Rollback(ctx)

	// 檢查金額
	if amount <= 0 {
		return errors.New("deposit amount must be positive")
	}

	// 查詢錢包並鎖定
	var currentBalance float64
	query := `SELECT balance FROM wallets WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, query, walletID).Scan(&currentBalance)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("wallet not found")
		}
		r.logger.Errorf("failed to find wallet: %v", err)
		return err
	}

	// 計算新餘額
	newBalance := currentBalance + amount
	updateQuery := `UPDATE wallets SET balance = $1, updated_at = $2 WHERE id = $3`

	_, err = tx.Exec(ctx, updateQuery, newBalance, time.Now(), walletID)
	if err != nil {
		r.logger.Errorf("failed to update wallet balance: %v", err)
		return err
	}

	// 將metadata轉換為JSON字符串
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		r.logger.Errorf("failed to marshal metadata: %v", err)
		return err
	}

	// 創建交易記錄
	now := time.Now()
	txInsertQuery := `
		INSERT INTO wallet_transactions (
			wallet_id, amount, balance_before, balance_after, 
			type, status, reference_id, description, metadata, 
			created_at, updated_at
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(
		ctx,
		txInsertQuery,
		walletID, amount, currentBalance, newBalance,
		txType, 1, referenceID, description, metadataBytes,
		now, now,
	)

	if err != nil {
		r.logger.Errorf("failed to create transaction record: %v", err)
		return err
	}

	// 提交事務
	if err = tx.Commit(ctx); err != nil {
		r.logger.Errorf("failed to commit transaction: %v", err)
		return err
	}

	// TODO: [Cache] Implement cache invalidation after the transaction is successfully committed.
	// The keys to invalidate would be `wallet:<walletID>` and the key for the user ID, which needs to be fetched first or passed in.

	return nil
}

// Withdraw 提款操作
func (r *walletRepo) Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	// 開始事務
	tx, err := r.data.db.Begin(ctx)
	if err != nil {
		r.logger.Errorf("failed to begin transaction: %v", err)
		return err
	}

	defer tx.Rollback(ctx)

	// 查詢錢包並鎖定
	var w WalletPO
	query := `
		SELECT id, user_id, balance, status, created_at, updated_at 
		FROM wallets 
		WHERE id = $1 
		FOR UPDATE
	`

	err = tx.QueryRow(ctx, query, walletID).Scan(
		&w.ID, &w.UserID, &w.Balance, &w.Status, &w.CreatedAt, &w.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("wallet not found")
		}
		r.logger.Errorf("failed to find wallet: %v", err)
		return err
	}

	// 檢查錢包狀態
	if w.Status != 1 {
		return errors.New("wallet is frozen")
	}

	// 檢查金額
	if amount <= 0 {
		return errors.New("withdraw amount must be positive")
	}

	// 檢查餘額
	if w.Balance < amount {
		return errors.New("insufficient balance")
	}

	// 更新餘額
	oldBalance := w.Balance
	w.Balance -= amount
	w.UpdatedAt = time.Now()

	// 保存錢包
	updateQuery := `
		UPDATE wallets 
		SET balance = $1, updated_at = $2 
		WHERE id = $3
	`

	_, err = tx.Exec(ctx, updateQuery, w.Balance, w.UpdatedAt, w.ID)
	if err != nil {
		r.logger.Errorf("failed to update wallet balance: %v", err)
		return err
	}

	// 將metadata轉換為JSON字符串
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		r.logger.Errorf("failed to marshal metadata: %v", err)
		return err
	}

	// 創建交易記錄
	now := time.Now()
	txInsertQuery := `
		INSERT INTO wallet_transactions (
			wallet_id, amount, balance_before, balance_after, 
			type, status, reference_id, description, metadata, 
			created_at, updated_at
		) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = tx.Exec(
		ctx,
		txInsertQuery,
		walletID, -amount, oldBalance, w.Balance,
		txType, 1, referenceID, description, string(metadataBytes),
		now, now,
	)

	if err != nil {
		r.logger.Errorf("failed to create transaction record: %v", err)
		return err
	}

	// 提交事務
	if err = tx.Commit(ctx); err != nil {
		r.logger.Errorf("failed to commit transaction: %v", err)
		return err
	}

	// TODO: [Cache] Implement cache invalidation after the transaction is successfully committed.
	// The keys to invalidate would be `wallet:<walletID>` and the key for the user ID (`wallet:user_id:{user_id}:currency:{currency}`).

	return nil
}
