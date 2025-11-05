package game

import (
	"math"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// 魚群陣型功能使用示例
// ========================================

// FormationExamples 魚群陣型功能示例
type FormationExamples struct {
	roomManager *RoomManager
	logger      logger.Logger
}

// NewFormationExamples 創建示例實例
func NewFormationExamples(roomManager *RoomManager, logger logger.Logger) *FormationExamples {
	return &FormationExamples{
		roomManager: roomManager,
		logger:      logger,
	}
}

// DemoBasicFormations 演示基本陣型功能
func (fe *FormationExamples) DemoBasicFormations() {
	fe.logger.Info("=== 魚群陣型功能演示開始 ===")

	// 1. 創建房間
	room, err := fe.roomManager.CreateRoom(RoomTypeNovice, 4)
	if err != nil {
		fe.logger.Errorf("Failed to create room: %v", err)
		return
	}

	fe.logger.Infof("創建房間成功: %s", room.ID)

	// 2. 演示不同類型的陣型
	formationTypes := []FishFormationType{
		FormationTypeV,
		FormationTypeLine,
		FormationTypeCircle,
		FormationTypeTriangle,
		FormationTypeDiamond,
		FormationTypeWave,
	}

	for _, formationType := range formationTypes {
		fe.demoFormationType(room.ID, formationType)
		time.Sleep(2 * time.Second) // 等待2秒觀察效果
	}

	// 3. 演示特殊陣型
	fe.demoSpecialFormations(room.ID)

	// 4. 演示自定義路線
	fe.demoCustomRoutes(room.ID)

	// 5. 顯示統計信息
	fe.showFormationStatistics(room.ID)

	fe.logger.Info("=== 魚群陣型功能演示結束 ===")
}

// demoFormationType 演示特定陣型類型
func (fe *FormationExamples) demoFormationType(roomID string, formationType FishFormationType) {
	fe.logger.Infof("--- 演示陣型: %s ---", formationType)

	// 生成陣型
	formation, err := fe.roomManager.SpawnFormationInRoom(roomID, formationType, "straight_left_right")
	if err != nil {
		fe.logger.Errorf("Failed to spawn formation %s: %v", formationType, err)
		return
	}

	fe.logger.Infof("成功生成陣型: %s, 魚數量: %d", formation.Type, len(formation.Fishes))

	// 顯示陣型詳細信息
	fe.logFormationDetails(formation)
}

// demoSpecialFormations 演示特殊陣型
func (fe *FormationExamples) demoSpecialFormations(roomID string) {
	fe.logger.Info("--- 演示特殊陣型 ---")

	// Boss魚陣型
	bossFormation, err := fe.roomManager.SpawnSpecialFormationInRoom(
		roomID,
		FormationTypeV,
		"s_curve",
		[]int32{31, 32, 33}, // 龍王魚、金龍魚、海王魚
	)
	if err != nil {
		fe.logger.Errorf("Failed to spawn boss formation: %v", err)
	} else {
		fe.logger.Infof("成功生成Boss陣型: %d條Boss魚", len(bossFormation.Fishes))
	}

	// 小魚群陣型
	smallFishFormation, err := fe.roomManager.SpawnSpecialFormationInRoom(
		roomID,
		FormationTypeWave,
		"wave_horizontal",
		[]int32{1, 1, 1, 2, 2, 3, 3}, // 多條小魚
	)
	if err != nil {
		fe.logger.Errorf("Failed to spawn small fish formation: %v", err)
	} else {
		fe.logger.Infof("成功生成小魚群陣型: %d條小魚", len(smallFishFormation.Fishes))
	}
}

// demoCustomRoutes 演示自定義路線
func (fe *FormationExamples) demoCustomRoutes(roomID string) {
	fe.logger.Info("--- 演示自定義路線 ---")

	// 創建心形路線
	heartPoints := fe.generateHeartShapePoints()
	heartRoute, err := fe.roomManager.CreateCustomRoute(
		"heart_route",
		"心形路線",
		heartPoints,
		RouteTypeCurved,
		1.2,
		true,
	)
	if err != nil {
		fe.logger.Errorf("Failed to create heart route: %v", err)
		return
	}

	fe.logger.Infof("成功創建心形路線: %s", heartRoute.Name)

	// 使用心形路線生成陣型
	heartFormation, err := fe.roomManager.SpawnSpecialFormationInRoom(
		roomID,
		FormationTypeCircle,
		"heart_route",
		[]int32{1, 2, 3, 11, 12}, // 混合魚類
	)
	if err != nil {
		fe.logger.Errorf("Failed to spawn formation with heart route: %v", err)
	} else {
		fe.logger.Infof("成功在心形路線上生成陣型: %d條魚", len(heartFormation.Fishes))
	}

	// 創建星形路線
	starPoints := fe.generateStarShapePoints()
	starRoute, err := fe.roomManager.CreateCustomRoute(
		"star_route",
		"星形路線",
		starPoints,
		RouteTypeCurved,
		1.5,
		true,
	)
	if err != nil {
		fe.logger.Errorf("Failed to create star route: %v", err)
		return
	}

	fe.logger.Infof("成功創建星形路線: %s", starRoute.Name)
}

// generateHeartShapePoints 生成心形路線點
func (fe *FormationExamples) generateHeartShapePoints() []Position {
	points := make([]Position, 32)
	centerX := 600.0 // 房間中心X
	centerY := 400.0 // 房間中心Y
	scale := 100.0

	for i := 0; i < 32; i++ {
		t := float64(i) * 2 * 3.14159 / 32
		
		// 心形參數方程
		x := 16 * math.Pow(math.Sin(t), 3)
		y := 13*math.Cos(t) - 5*math.Cos(2*t) - 2*math.Cos(3*t) - math.Cos(4*t)
		
		points[i] = Position{
			X: centerX + x*scale/16,
			Y: centerY - y*scale/16, // Y軸反轉，因為屏幕坐標系
		}
	}

	return points
}

// generateStarShapePoints 生成星形路線點
func (fe *FormationExamples) generateStarShapePoints() []Position {
	points := make([]Position, 20)
	centerX := 600.0
	centerY := 400.0
	outerRadius := 150.0
	innerRadius := 75.0

	for i := 0; i < 20; i++ {
		angle := float64(i) * 2 * 3.14159 / 20
		var radius float64
		
		if i%2 == 0 {
			radius = outerRadius // 外圈點
		} else {
			radius = innerRadius // 內圈點
		}

		points[i] = Position{
			X: centerX + radius*math.Cos(angle),
			Y: centerY + radius*math.Sin(angle),
		}
	}

	return points
}

// logFormationDetails 記錄陣型詳細信息
func (fe *FormationExamples) logFormationDetails(formation *FishFormation) {
	fe.logger.Infof("陣型ID: %s", formation.ID)
	fe.logger.Infof("陣型類型: %s", formation.Type)
	fe.logger.Infof("路線: %s", formation.Route.Name)
	fe.logger.Infof("魚數量: %d", len(formation.Fishes))
	fe.logger.Infof("陣型大小: %.1fx%.1f", formation.Size.Width, formation.Size.Height)
	fe.logger.Infof("移動速度: %.1f", formation.Speed)
	fe.logger.Infof("狀態: %s", formation.Status)
	
	// 顯示魚的類型分布
	fishTypeCount := make(map[string]int)
	for _, fish := range formation.Fishes {
		fishTypeCount[fish.Type.Name]++
	}
	
	fe.logger.Info("魚類型分布:")
	for fishType, count := range fishTypeCount {
		fe.logger.Infof("  %s: %d條", fishType, count)
	}
}

// showFormationStatistics 顯示陣型統計信息
func (fe *FormationExamples) showFormationStatistics(roomID string) {
	fe.logger.Info("--- 陣型統計信息 ---")

	stats, err := fe.roomManager.GetFormationStatistics(roomID)
	if err != nil {
		fe.logger.Errorf("Failed to get formation statistics: %v", err)
		return
	}

	fe.logger.Infof("總陣型數量: %v", stats["total_formations"])
	fe.logger.Infof("陣型中總魚數: %v", stats["total_formation_fishes"])

	if formationsByType, ok := stats["formations_by_type"].(map[FishFormationType]int); ok {
		fe.logger.Info("按類型分布:")
		for formationType, count := range formationsByType {
			fe.logger.Infof("  %s: %d個", formationType, count)
		}
	}

	if formationsByStatus, ok := stats["formations_by_status"].(map[FormationStatus]int); ok {
		fe.logger.Info("按狀態分布:")
		for status, count := range formationsByStatus {
			fe.logger.Infof("  %s: %d個", status, count)
		}
	}
}

// DemoRouteManagement 演示路線管理功能
func (fe *FormationExamples) DemoRouteManagement() {
	fe.logger.Info("=== 路線管理功能演示 ===")

	// 獲取所有可用路線
	routes := fe.roomManager.GetAvailableRoutes()
	fe.logger.Infof("可用路線總數: %d", len(routes))

	// 按類型顯示路線
	routeTypes := []FishRouteType{
		RouteTypeStraight,
		RouteTypeCurved,
		RouteTypeZigzag,
		RouteTypeCircular,
		RouteTypeRandom,
	}

	for _, routeType := range routeTypes {
		typeRoutes := fe.roomManager.GetRoutesByType(routeType)
		fe.logger.Infof("%s 類型路線: %d條", routeType, len(typeRoutes))
		
		for _, route := range typeRoutes {
			fe.logger.Infof("  - %s (難度: %.1f, 循環: %v)", 
				route.Name, route.Difficulty, route.Looping)
		}
	}
}

// TestFormationPerformance 測試陣型性能
func (fe *FormationExamples) TestFormationPerformance() {
	fe.logger.Info("=== 陣型性能測試 ===")

	room, err := fe.roomManager.CreateRoom(RoomTypeAdvanced, 8)
	if err != nil {
		fe.logger.Errorf("Failed to create test room: %v", err)
		return
	}

	startTime := time.Now()
	formationCount := 0

	// 連續生成多個陣型測試性能
	for i := 0; i < 10; i++ {
		formation, err := fe.roomManager.SpawnFormationInRoom(room.ID, FormationTypeV, "straight_left_right")
		if err == nil && formation != nil {
			formationCount++
		}
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	fe.logger.Infof("性能測試結果: 生成 %d 個陣型, 耗時: %v", formationCount, duration)
	fe.logger.Infof("平均每個陣型生成時間: %v", duration/time.Duration(formationCount))

	// 測試陣型更新性能
	updateStartTime := time.Now()
	for i := 0; i < 100; i++ {
		// 模擬更新循環
		time.Sleep(10 * time.Millisecond)
	}
	updateDuration := time.Since(updateStartTime)
	fe.logger.Infof("100次更新循環耗時: %v", updateDuration)
}