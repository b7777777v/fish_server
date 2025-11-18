package game_test

import (
	"errors"
	"testing"

	"github.com/b7777777v/fish_server/internal/biz/game"
	"github.com/b7777777v/fish_server/internal/testing/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestFireBullet_WalletWithdrawFailure 測試開火時錢包扣款失敗的回滾
func TestFireBullet_WalletWithdrawFailure(t *testing.T) {
	// 使用測試輔助工具創建環境（使用默認mocks）
	env := testhelper.NewGameTestEnv(t, nil)

	// 創建測試房間
	room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 4)
	assert.NoError(t, err)

	// 創建測試玩家
	playerID := int64(1)
	testPlayer := testhelper.NewTestPlayer(playerID)
	testPlayer.WalletID = 1

	// Mock 玩家倉庫返回
	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
	env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil)

	// 玩家加入房間
	err = env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)
	assert.NoError(t, err)

	// 獲取房間中的玩家以記錄初始餘額
	roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
	initialBalance := roomState.Players[playerID].Balance

	// Mock 錢包扣款失敗（清除默認mock）
	env.WalletRepo.ExpectedCalls = nil
	env.WalletRepo.On("Withdraw",
		env.Ctx,
		uint(1),
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(errors.New("wallet operation failed"))

	// 測試開火（預期錢包扣款失敗並回滾）
	bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 0.0, 10, game.Position{X: 600, Y: 750})

	// 驗證結果
	assert.Error(t, err, "應該返回錢包操作失敗的錯誤")
	assert.Nil(t, bullet, "開火應該失敗")
	assert.Contains(t, err.Error(), "wallet operation failed", "錯誤訊息應該包含錢包操作失敗")

	// 驗證玩家餘額已回滾
	roomState, _ = env.GameUsecase.GetRoomState(env.Ctx, room.ID)
	assert.Equal(t, initialBalance, roomState.Players[playerID].Balance, "玩家餘額應該回滾到初始值")
}

// TestFireBullet_WalletWithdrawSuccess 測試開火時錢包扣款成功
func TestFireBullet_WalletWithdrawSuccess(t *testing.T) {
	// 使用測試輔助工具創建環境
	env := testhelper.NewGameTestEnv(t, nil)

	// 創建測試房間
	room, err := env.GameUsecase.CreateRoom(env.Ctx, game.RoomTypeNovice, 4)
	assert.NoError(t, err)

	// 創建測試玩家
	playerID := int64(1)
	testPlayer := testhelper.NewTestPlayer(playerID)
	testPlayer.WalletID = 1

	// Mock 玩家倉庫
	env.PlayerRepo.On("GetPlayer", env.Ctx, playerID).Return(testPlayer, nil)
	env.PlayerRepo.On("UpdatePlayerStatus", env.Ctx, playerID, game.PlayerStatusPlaying).Return(nil)
	env.PlayerRepo.On("UpdatePlayerBalance", env.Ctx, playerID, mock.Anything).Return(nil)

	// 玩家加入房間
	err = env.GameUsecase.JoinRoom(env.Ctx, room.ID, playerID)
	assert.NoError(t, err)

	// 獲取初始餘額
	roomState, _ := env.GameUsecase.GetRoomState(env.Ctx, room.ID)
	initialBalance := roomState.Players[playerID].Balance

	// 測試開火（預期成功）
	bullet, err := env.GameUsecase.FireBullet(env.Ctx, room.ID, playerID, 0.0, 10, game.Position{X: 600, Y: 750})

	// 驗證結果
	assert.NoError(t, err, "開火應該成功")
	assert.NotNil(t, bullet, "應該返回子彈對象")

	// 驗證玩家餘額已扣除
	roomState, _ = env.GameUsecase.GetRoomState(env.Ctx, room.ID)
	assert.Less(t, roomState.Players[playerID].Balance, initialBalance, "玩家餘額應該減少")
	assert.Equal(t, initialBalance-bullet.Cost, roomState.Players[playerID].Balance, "餘額應該減少子彈成本")
}
