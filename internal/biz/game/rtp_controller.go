package game

import (
	"math/rand"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// RTPController manages the game's Return-To-Player mechanism.
type RTPController struct {
	inventoryManager *InventoryManager
	logger           logger.Logger
}

// NewRTPController creates a new RTP controller.
func NewRTPController(im *InventoryManager, logger logger.Logger) *RTPController {
	return &RTPController{
		inventoryManager: im,
		logger:           logger.With("component", "rtp_controller"),
	}
}

// ApproveKill decides if a potential reward should be granted based on RTP.
func (rc *RTPController) ApproveKill(roomType RoomType, targetRTP float64, potentialReward int64) bool {
	inv := rc.inventoryManager.GetInventory(roomType)

	// If there's not enough data, always approve the kill.
	// The RTP will be volatile at the beginning anyway.
	if inv.TotalIn < 100000 { // 1000元
		return true
	}

	currentRTP := inv.CurrentRTP

	// If current RTP is significantly higher than the target, start denying wins.
	if currentRTP > targetRTP*1.05 {
		// Deny a certain percentage of wins to lower the RTP.
		// The higher the surplus, the higher the chance of denial.
		denialChance := (currentRTP - targetRTP) / currentRTP
		if rand.Float64() < denialChance {
			rc.logger.Debugf("RTP is high (%.2f%% vs target %.2f%%). Denying kill.", currentRTP*100, targetRTP*100)
			return false
		}
	}

	// If current RTP is lower than the target, always approve the win to help it catch up.
	if currentRTP < targetRTP {
		rc.logger.Debugf("RTP is low (%.2f%% vs target %.2f%%). Approving kill.", currentRTP*100, targetRTP*100)
		return true
	}

	// In the normal range, approve the kill.
	return true
}
