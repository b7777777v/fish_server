// internal/biz/wallet/wallet.go
package wallet

import (
	"context"
	"time"
)

// Wallet 是錢包的領域模型
type Wallet struct {
	ID        uint
	UserID    uint
	Balance   float64
	Currency  string
	Status    int8 // 1: 正常, 0: 凍結
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Transaction 是錢包交易的領域模型
type Transaction struct {
	ID            uint
	WalletID      uint
	Amount        float64 // 正數表示收入，負數表示支出
	BalanceBefore float64
	BalanceAfter  float64
	Type          string // 'deposit', 'withdraw', 'game_win', 'game_lose', 'bonus', etc.
	Status        int8   // 1: 成功, 0: 失敗, 2: 處理中
	ReferenceID   string // 外部參考ID，例如遊戲ID或支付系統交易ID
	Description   string
	Metadata      map[string]interface{} // 額外的交易相關數據
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// WalletRepo 定義了錢包數據倉庫的接口
type WalletRepo interface {
	// 查詢錢包
	FindByID(ctx context.Context, id uint) (*Wallet, error)
	FindByUserID(ctx context.Context, userID uint, currency string) (*Wallet, error)

	// 創建錢包
	Create(ctx context.Context, w *Wallet) error

	// 更新錢包
	Update(ctx context.Context, w *Wallet) error

	// 交易相關
	CreateTransaction(ctx context.Context, tx *Transaction) error
	FindTransactionsByWalletID(ctx context.Context, walletID uint, limit, offset int) ([]*Transaction, error)

	// 餘額操作
	Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error
	Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error
}
