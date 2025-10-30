// internal/biz/wallet/usecase.go
package wallet

import (
	"context"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// WalletUsecase 是錢包業務邏輯的用例
type WalletUsecase struct {
	repo   WalletRepo
	logger logger.Logger
}

// NewWalletUsecase 創建一個新的 WalletUsecase 實例
func NewWalletUsecase(repo WalletRepo, logger logger.Logger) *WalletUsecase {
	return &WalletUsecase{
		repo:   repo,
		logger: logger.With("module", "biz/wallet"),
	}
}

// GetWallet 獲取錢包信息
func (uc *WalletUsecase) GetWallet(ctx context.Context, id uint) (*Wallet, error) {
	return uc.repo.FindByID(ctx, id)
}

// GetWalletByUserID 根據用戶ID獲取錢包信息
func (uc *WalletUsecase) GetWalletByUserID(ctx context.Context, userID uint, currency string) (*Wallet, error) {
	return uc.repo.FindByUserID(ctx, userID, currency)
}

// CreateWallet 創建錢包
func (uc *WalletUsecase) CreateWallet(ctx context.Context, userID uint, currency string) (*Wallet, error) {
	// 檢查用戶是否已有該幣種的錢包
	existingWallet, err := uc.repo.FindByUserID(ctx, userID, currency)
	if err == nil && existingWallet != nil {
		// 用戶已有該幣種的錢包
		return existingWallet, nil
	}

	// 創建新錢包
	wallet := &Wallet{
		UserID:   userID,
		Balance:  0,
		Currency: currency,
		Status:   1, // 正常狀態
	}

	err = uc.repo.Create(ctx, wallet)
	if err != nil {
		uc.logger.Errorf("failed to create wallet: %v", err)
		return nil, err
	}

	return wallet, nil
}

// Deposit 存款
func (uc *WalletUsecase) Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	return uc.repo.Deposit(ctx, walletID, amount, txType, referenceID, description, metadata)
}

// Withdraw 提款
func (uc *WalletUsecase) Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	return uc.repo.Withdraw(ctx, walletID, amount, txType, referenceID, description, metadata)
}

// GetTransactions 獲取交易記錄
func (uc *WalletUsecase) GetTransactions(ctx context.Context, walletID uint, limit, offset int) ([]*Transaction, error) {
	return uc.repo.FindTransactionsByWalletID(ctx, walletID, limit, offset)
}

// FreezeWallet 凍結錢包
func (uc *WalletUsecase) FreezeWallet(ctx context.Context, walletID uint) error {
	wallet, err := uc.repo.FindByID(ctx, walletID)
	if err != nil {
		return err
	}

	wallet.Status = 0 // 凍結狀態
	return uc.repo.Update(ctx, wallet)
}

// UnfreezeWallet 解凍錢包
func (uc *WalletUsecase) UnfreezeWallet(ctx context.Context, walletID uint) error {
	wallet, err := uc.repo.FindByID(ctx, walletID)
	if err != nil {
		return err
	}

	wallet.Status = 1 // 正常狀態
	return uc.repo.Update(ctx, wallet)
}