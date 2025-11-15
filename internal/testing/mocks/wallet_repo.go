package mocks

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/wallet"
	"github.com/stretchr/testify/mock"
)

// WalletRepo is a mock implementation of wallet.WalletRepo interface
type WalletRepo struct {
	mock.Mock
}

// FindByID mocks the FindByID method
func (m *WalletRepo) FindByID(ctx context.Context, id uint) (*wallet.Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

// FindByUserID mocks the FindByUserID method
func (m *WalletRepo) FindByUserID(ctx context.Context, userID uint, currency string) (*wallet.Wallet, error) {
	args := m.Called(ctx, userID, currency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

// FindAllByUserID mocks the FindAllByUserID method
func (m *WalletRepo) FindAllByUserID(ctx context.Context, userID uint) ([]*wallet.Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*wallet.Wallet), args.Error(1)
}

// Create mocks the Create method
func (m *WalletRepo) Create(ctx context.Context, w *wallet.Wallet) error {
	args := m.Called(ctx, w)
	return args.Error(0)
}

// Update mocks the Update method
func (m *WalletRepo) Update(ctx context.Context, w *wallet.Wallet) error {
	args := m.Called(ctx, w)
	return args.Error(0)
}

// Deposit mocks the Deposit method
func (m *WalletRepo) Deposit(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	args := m.Called(ctx, walletID, amount, txType, referenceID, description, metadata)
	return args.Error(0)
}

// Withdraw mocks the Withdraw method
func (m *WalletRepo) Withdraw(ctx context.Context, walletID uint, amount float64, txType, referenceID, description string, metadata map[string]interface{}) error {
	args := m.Called(ctx, walletID, amount, txType, referenceID, description, metadata)
	return args.Error(0)
}

// CreateTransaction mocks the CreateTransaction method
func (m *WalletRepo) CreateTransaction(ctx context.Context, tx *wallet.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

// FindTransactionsByWalletID mocks the FindTransactionsByWalletID method
func (m *WalletRepo) FindTransactionsByWalletID(ctx context.Context, walletID uint, limit, offset int) ([]*wallet.Transaction, error) {
	args := m.Called(ctx, walletID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*wallet.Transaction), args.Error(1)
}
