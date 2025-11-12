package game

import (
	"math"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
)

// ========================================
// 魚群陣型和路線系統
// ========================================

// FishFormationType 魚群陣型類型
type FishFormationType string

const (
	FormationTypeV        FishFormationType = "v_shape"        // V字型
	FormationTypeLine     FishFormationType = "line"          // 直線型
	FormationTypeCircle   FishFormationType = "circle"        // 圓形
	FormationTypeTriangle FishFormationType = "triangle"      // 三角形
	FormationTypeDiamond  FishFormationType = "diamond"       // 菱形
	FormationTypeWave     FishFormationType = "wave"          // 波浪型
	FormationTypeSpiral   FishFormationType = "spiral"        // 螺旋型
)

// FishRoute 魚群路線
type FishRoute struct {
	ID          string              `json:"id"`          // 路線ID
	Name        string              `json:"name"`        // 路線名稱
	Points      []Position          `json:"points"`      // 路線關鍵點
	Duration    time.Duration       `json:"duration"`    // 路線總時長
	Type        FishRouteType       `json:"type"`        // 路線類型
	Difficulty  float64             `json:"difficulty"`  // 難度係數 (0.5-2.0)
	Looping     bool                `json:"looping"`     // 是否循環
	Smooth      bool                `json:"smooth"`      // 是否使用平滑插值（Catmull-Rom樣條）
	CreatedAt   time.Time          `json:"created_at"`
}

// FishRouteType 魚群路線類型
type FishRouteType string

const (
	RouteTypeStraight  FishRouteType = "straight"   // 直線路線
	RouteTypeCurved    FishRouteType = "curved"     // 曲線路線
	RouteTypeZigzag    FishRouteType = "zigzag"     // Z字型路線
	RouteTypeCircular  FishRouteType = "circular"   // 圓形路線
	RouteTypeRandom    FishRouteType = "random"     // 隨機路線
)

// FishFormation 魚群陣型
type FishFormation struct {
	ID           string              `json:"id"`            // 陣型ID
	Type         FishFormationType   `json:"type"`          // 陣型類型
	LeaderFish   *Fish              `json:"leader_fish"`   // 領頭魚
	Fishes       []*Fish            `json:"fishes"`        // 陣型中的魚群
	Route        *FishRoute         `json:"route"`         // 移動路線
	Position     Position           `json:"position"`      // 陣型中心位置
	Direction    float64            `json:"direction"`     // 移動方向
	Speed        float64            `json:"speed"`         // 移動速度
	Size         FormationSize      `json:"size"`          // 陣型大小
	Status       FormationStatus    `json:"status"`        // 陣型狀態
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Progress     float64           `json:"progress"`      // 路線進度 (0.0-1.0)
	LoopCount    int               `json:"loop_count"`    // 循環次數計數
	Config       FormationConfig   `json:"config"`        // 陣型配置
}

// FormationSize 陣型大小
type FormationSize struct {
	Width  float64 `json:"width"`  // 陣型寬度
	Height float64 `json:"height"` // 陣型高度
	Depth  float64 `json:"depth"`  // 陣型深度
}

// FormationStatus 陣型狀態
type FormationStatus string

const (
	FormationStatusForming   FormationStatus = "forming"    // 組建中
	FormationStatusMoving    FormationStatus = "moving"     // 移動中
	FormationStatusScattered FormationStatus = "scattered"  // 散開
	FormationStatusComplete  FormationStatus = "complete"   // 完成
)

// FormationConfig 陣型配置
type FormationConfig struct {
	Spacing          float64 `json:"spacing"`           // 魚之間的間距
	Cohesion         float64 `json:"cohesion"`          // 聚合力 (0.0-1.0)
	Alignment        float64 `json:"alignment"`         // 對齊力 (0.0-1.0)
	Separation       float64 `json:"separation"`        // 分離力 (0.0-1.0)
	FollowLeader     bool    `json:"follow_leader"`     // 是否跟隨領頭魚
	MaintainSpeed    bool    `json:"maintain_speed"`    // 是否保持統一速度
	AllowBreakaway   bool    `json:"allow_breakaway"`   // 是否允許脫離陣型
	MinFishes        int     `json:"min_fishes"`        // 最少魚數量
	MaxFishes        int     `json:"max_fishes"`        // 最多魚數量
	ReformThreshold  float64 `json:"reform_threshold"`  // 重組閾值
}

// FishFormationManager 魚群陣型管理器
type FishFormationManager struct {
	formations   map[string]*FishFormation `json:"formations"`
	routes       map[string]*FishRoute     `json:"routes"`
	logger       logger.Logger
	roomConfig   RoomConfig
}

// NewFishFormationManager 創建魚群陣型管理器
func NewFishFormationManager(logger logger.Logger, roomConfig RoomConfig) *FishFormationManager {
	manager := &FishFormationManager{
		formations: make(map[string]*FishFormation),
		routes:     make(map[string]*FishRoute),
		logger:     logger.With("component", "formation_manager"),
		roomConfig: roomConfig,
	}
	
	// 初始化預設路線
	manager.initializeDefaultRoutes()
	
	return manager
}

// CreateFormation 創建魚群陣型
func (fm *FishFormationManager) CreateFormation(formationType FishFormationType, fishes []*Fish, routeID string) *FishFormation {
	if len(fishes) == 0 {
		fm.logger.Warn("Cannot create formation with empty fish list")
		return nil
	}
	
	route := fm.routes[routeID]
	if route == nil {
		fm.logger.Warnf("Route not found: %s", routeID)
		return nil
	}
	
	formation := &FishFormation{
		ID:         generateFormationID(),
		Type:       formationType,
		LeaderFish: fishes[0], // 第一條魚作為領頭魚
		Fishes:     fishes,
		Route:      route,
		Position:   calculateCenterPosition(fishes),
		Direction:  fishes[0].Direction,
		Speed:      calculateAverageSpeed(fishes),
		Size:       fm.calculateFormationSize(formationType, len(fishes)),
		Status:     FormationStatusForming,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Progress:   0.0,
		Config:     fm.getDefaultFormationConfig(formationType),
	}
	
	// 設置魚的陣型位置
	fm.arrangeFormation(formation)
	
	fm.formations[formation.ID] = formation
	fm.logger.Infof("Created formation: id=%s, type=%s, fish_count=%d", formation.ID, formationType, len(fishes))
	
	return formation
}

// UpdateFormations 更新所有陣型
func (fm *FishFormationManager) UpdateFormations(deltaTime float64) {
	for _, formation := range fm.formations {
		fm.updateFormation(formation, deltaTime)
	}
}

// updateFormation 更新單個陣型
func (fm *FishFormationManager) updateFormation(formation *FishFormation, deltaTime float64) {
	if formation.Status != FormationStatusMoving {
		return
	}

	// 保存旧位置用于计算方向
	oldPosition := formation.Position

	// 更新路線進度
	distancePerSecond := formation.Speed
	progressIncrement := (distancePerSecond * deltaTime) / fm.calculateRouteLength(formation.Route)
	formation.Progress += progressIncrement

	// 檢查是否完成路線
	if formation.Progress >= 1.0 {
		if formation.Route.Looping {
			formation.Progress = 0.0
			formation.LoopCount++

			// 限制循環次數，防止阵型永远存在
			// 循環 2-3 次後標記為完成
			maxLoops := 3
			if formation.LoopCount >= maxLoops {
				fm.logger.Infof("Formation %s completed %d loops, marking as complete", formation.ID, formation.LoopCount)
				formation.Status = FormationStatusComplete
				return
			}
		} else {
			formation.Status = FormationStatusComplete
			return
		}
	}

	// 計算當前位置
	newPosition := fm.interpolateRoutePosition(formation.Route, formation.Progress)
	formation.Position = newPosition

	// 計算移動方向（基於位置變化）
	dx := newPosition.X - oldPosition.X
	dy := newPosition.Y - oldPosition.Y
	if dx != 0 || dy != 0 {
		// 使用 atan2 計算方向（弧度）
		formation.Direction = math.Atan2(dy, dx)
	}

	// 更新陣型中所有魚的位置
	fm.updateFishPositions(formation)

	formation.UpdatedAt = time.Now()
}

// arrangeFormation 排列陣型
func (fm *FishFormationManager) arrangeFormation(formation *FishFormation) {
	switch formation.Type {
	case FormationTypeV:
		fm.arrangeVFormation(formation)
	case FormationTypeLine:
		fm.arrangeLineFormation(formation)
	case FormationTypeCircle:
		fm.arrangeCircleFormation(formation)
	case FormationTypeTriangle:
		fm.arrangeTriangleFormation(formation)
	case FormationTypeDiamond:
		fm.arrangeDiamondFormation(formation)
	case FormationTypeWave:
		fm.arrangeWaveFormation(formation)
	case FormationTypeSpiral:
		fm.arrangeSpiralFormation(formation)
	}
}

// arrangeVFormation 排列V字陣型
func (fm *FishFormationManager) arrangeVFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	angle := math.Pi / 6 // 30度角
	
	for i, fish := range formation.Fishes {
		if i == 0 {
			// 領頭魚在頂點
			fish.Position = formation.Position
		} else {
			// 其他魚排列在V字兩側
			side := ((i - 1) % 2) * 2 - 1 // -1 或 1
			row := (i - 1) / 2 + 1
			
			offsetX := float64(side) * float64(row) * spacing * math.Cos(angle)
			offsetY := -float64(row) * spacing * math.Sin(angle)
			
			fish.Position = Position{
				X: formation.Position.X + offsetX,
				Y: formation.Position.Y + offsetY,
			}
		}
	}
}

// arrangeLineFormation 排列直線陣型
func (fm *FishFormationManager) arrangeLineFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	
	for i, fish := range formation.Fishes {
		offsetX := float64(i) * spacing
		fish.Position = Position{
			X: formation.Position.X + offsetX,
			Y: formation.Position.Y,
		}
	}
}

// arrangeCircleFormation 排列圓形陣型
func (fm *FishFormationManager) arrangeCircleFormation(formation *FishFormation) {
	radius := formation.Size.Width / 2
	angleStep := 2 * math.Pi / float64(len(formation.Fishes))
	
	for i, fish := range formation.Fishes {
		angle := float64(i) * angleStep
		offsetX := radius * math.Cos(angle)
		offsetY := radius * math.Sin(angle)
		
		fish.Position = Position{
			X: formation.Position.X + offsetX,
			Y: formation.Position.Y + offsetY,
		}
	}
}

// arrangeTriangleFormation 排列三角陣型
func (fm *FishFormationManager) arrangeTriangleFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	currentRow := 0
	currentPos := 0
	fishInRow := 1
	
	for _, fish := range formation.Fishes {
		if currentPos >= fishInRow {
			currentRow++
			currentPos = 0
			fishInRow++
		}
		
		// 計算在該行的位置
		rowWidth := float64(fishInRow-1) * spacing
		startX := formation.Position.X - rowWidth/2
		
		fish.Position = Position{
			X: startX + float64(currentPos)*spacing,
			Y: formation.Position.Y + float64(currentRow)*spacing,
		}
		
		currentPos++
	}
}

// arrangeDiamondFormation 排列菱形陣型
func (fm *FishFormationManager) arrangeDiamondFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	fishCount := len(formation.Fishes)
	halfCount := fishCount / 2
	
	for i, fish := range formation.Fishes {
		var row, posInRow int
		
		if i <= halfCount {
			// 上半部分
			row = i
			posInRow = 0
		} else {
			// 下半部分
			row = fishCount - i - 1
			posInRow = 0
		}
		
		offsetX := float64(posInRow) * spacing
		offsetY := float64(row) * spacing
		
		fish.Position = Position{
			X: formation.Position.X + offsetX,
			Y: formation.Position.Y + offsetY,
		}
	}
}

// arrangeWaveFormation 排列波浪陣型
func (fm *FishFormationManager) arrangeWaveFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	amplitude := formation.Size.Height / 4
	frequency := 2.0
	
	for i, fish := range formation.Fishes {
		offsetX := float64(i) * spacing
		offsetY := amplitude * math.Sin(frequency*offsetX/100.0)
		
		fish.Position = Position{
			X: formation.Position.X + offsetX,
			Y: formation.Position.Y + offsetY,
		}
	}
}

// arrangeSpiralFormation 排列螺旋陣型
func (fm *FishFormationManager) arrangeSpiralFormation(formation *FishFormation) {
	spacing := formation.Config.Spacing
	spiralFactor := 5.0
	
	for i, fish := range formation.Fishes {
		angle := float64(i) * 0.5
		radius := float64(i) * spacing / spiralFactor
		
		offsetX := radius * math.Cos(angle)
		offsetY := radius * math.Sin(angle)
		
		fish.Position = Position{
			X: formation.Position.X + offsetX,
			Y: formation.Position.Y + offsetY,
		}
	}
}

// updateFishPositions 更新陣型中魚的位置
func (fm *FishFormationManager) updateFishPositions(formation *FishFormation) {
	// 重新排列陣型（考慮新的中心位置）
	fm.arrangeFormation(formation)
	
	// 更新每條魚的方向和速度
	for _, fish := range formation.Fishes {
		fish.Direction = formation.Direction
		if formation.Config.MaintainSpeed {
			fish.Speed = formation.Speed
		}
	}
}

// StartFormation 啟動陣型移動
func (fm *FishFormationManager) StartFormation(formationID string) bool {
	formation := fm.formations[formationID]
	if formation == nil {
		fm.logger.Warnf("Formation not found: %s", formationID)
		return false
	}
	
	formation.Status = FormationStatusMoving
	formation.Progress = 0.0
	fm.logger.Infof("Started formation: %s", formationID)
	return true
}

// StopFormation 停止陣型移動
func (fm *FishFormationManager) StopFormation(formationID string) bool {
	formation := fm.formations[formationID]
	if formation == nil {
		return false
	}
	
	formation.Status = FormationStatusScattered
	fm.logger.Infof("Stopped formation: %s", formationID)
	return true
}

// GetFormation 獲取陣型
func (fm *FishFormationManager) GetFormation(formationID string) *FishFormation {
	return fm.formations[formationID]
}

// GetAllFormations 獲取所有陣型
func (fm *FishFormationManager) GetAllFormations() []*FishFormation {
	formations := make([]*FishFormation, 0, len(fm.formations))
	for _, formation := range fm.formations {
		formations = append(formations, formation)
	}
	return formations
}

// RemoveFormation 移除陣型
func (fm *FishFormationManager) RemoveFormation(formationID string) bool {
	if _, exists := fm.formations[formationID]; !exists {
		return false
	}
	
	delete(fm.formations, formationID)
	fm.logger.Infof("Removed formation: %s", formationID)
	return true
}

// 工具函數
func generateFormationID() string {
	return "formation_" + time.Now().Format("20060102150405") + "_" + string(rune(time.Now().UnixNano()%1000))
}

func calculateCenterPosition(fishes []*Fish) Position {
	if len(fishes) == 0 {
		return Position{X: 0, Y: 0}
	}
	
	var totalX, totalY float64
	for _, fish := range fishes {
		totalX += fish.Position.X
		totalY += fish.Position.Y
	}
	
	return Position{
		X: totalX / float64(len(fishes)),
		Y: totalY / float64(len(fishes)),
	}
}

func calculateAverageSpeed(fishes []*Fish) float64 {
	if len(fishes) == 0 {
		return 0
	}
	
	var totalSpeed float64
	for _, fish := range fishes {
		totalSpeed += fish.Speed
	}
	
	return totalSpeed / float64(len(fishes))
}

func (fm *FishFormationManager) calculateFormationSize(formationType FishFormationType, fishCount int) FormationSize {
	baseSize := 100.0 + float64(fishCount)*20.0
	
	switch formationType {
	case FormationTypeV:
		return FormationSize{Width: baseSize * 1.5, Height: baseSize, Depth: baseSize * 0.5}
	case FormationTypeLine:
		return FormationSize{Width: baseSize * 2, Height: 50, Depth: 50}
	case FormationTypeCircle:
		radius := baseSize / 2
		return FormationSize{Width: radius * 2, Height: radius * 2, Depth: radius}
	case FormationTypeTriangle:
		return FormationSize{Width: baseSize, Height: baseSize * 0.8, Depth: baseSize * 0.6}
	case FormationTypeDiamond:
		return FormationSize{Width: baseSize * 0.8, Height: baseSize * 1.2, Depth: baseSize * 0.7}
	case FormationTypeWave:
		return FormationSize{Width: baseSize * 2, Height: baseSize * 0.5, Depth: 50}
	case FormationTypeSpiral:
		return FormationSize{Width: baseSize, Height: baseSize, Depth: baseSize}
	default:
		return FormationSize{Width: baseSize, Height: baseSize, Depth: baseSize}
	}
}

func (fm *FishFormationManager) getDefaultFormationConfig(formationType FishFormationType) FormationConfig {
	baseConfig := FormationConfig{
		Spacing:          50.0,
		Cohesion:         0.7,
		Alignment:        0.8,
		Separation:       0.6,
		FollowLeader:     true,
		MaintainSpeed:    true,
		AllowBreakaway:   false,
		MinFishes:        3,
		MaxFishes:        20,
		ReformThreshold:  100.0,
	}
	
	switch formationType {
	case FormationTypeV:
		baseConfig.Spacing = 60.0
		baseConfig.Cohesion = 0.8
	case FormationTypeLine:
		baseConfig.Spacing = 40.0
		baseConfig.Alignment = 0.9
	case FormationTypeCircle:
		baseConfig.Separation = 0.8
		baseConfig.Cohesion = 0.9
	case FormationTypeWave:
		baseConfig.Spacing = 45.0
		baseConfig.AllowBreakaway = true
	case FormationTypeSpiral:
		baseConfig.Spacing = 35.0
		baseConfig.FollowLeader = false
	}
	
	return baseConfig
}

func (fm *FishFormationManager) calculateRouteLength(route *FishRoute) float64 {
	if len(route.Points) < 2 {
		return 0
	}
	
	var totalLength float64
	for i := 1; i < len(route.Points); i++ {
		dx := route.Points[i].X - route.Points[i-1].X
		dy := route.Points[i].Y - route.Points[i-1].Y
		totalLength += math.Sqrt(dx*dx + dy*dy)
	}
	
	return totalLength
}

// catmullRomSpline 計算 Catmull-Rom 樣條插值
// t: 插值參數 (0-1)
// p0, p1, p2, p3: 四個控制點
func catmullRomSpline(t float64, p0, p1, p2, p3 Position) Position {
	t2 := t * t
	t3 := t2 * t

	// Catmull-Rom 基函數
	x := 0.5 * ((2 * p1.X) +
		(-p0.X + p2.X) * t +
		(2*p0.X - 5*p1.X + 4*p2.X - p3.X) * t2 +
		(-p0.X + 3*p1.X - 3*p2.X + p3.X) * t3)

	y := 0.5 * ((2 * p1.Y) +
		(-p0.Y + p2.Y) * t +
		(2*p0.Y - 5*p1.Y + 4*p2.Y - p3.Y) * t2 +
		(-p0.Y + 3*p1.Y - 3*p2.Y + p3.Y) * t3)

	return Position{X: x, Y: y}
}

func (fm *FishFormationManager) interpolateRoutePosition(route *FishRoute, progress float64) Position {
	if len(route.Points) == 0 {
		return Position{X: 0, Y: 0}
	}

	if len(route.Points) == 1 {
		return route.Points[0]
	}

	// 限制進度範圍
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	// 如果啟用平滑插值且點數足夠，使用 Catmull-Rom 樣條
	if route.Smooth && len(route.Points) >= 3 {
		return fm.interpolateSmoothRoute(route, progress)
	}

	// 否則使用線性插值
	totalLength := fm.calculateRouteLength(route)
	targetDistance := progress * totalLength

	currentDistance := 0.0
	for i := 1; i < len(route.Points); i++ {
		segmentStart := route.Points[i-1]
		segmentEnd := route.Points[i]

		dx := segmentEnd.X - segmentStart.X
		dy := segmentEnd.Y - segmentStart.Y
		segmentLength := math.Sqrt(dx*dx + dy*dy)

		if currentDistance+segmentLength >= targetDistance {
			// 在這個線段內
			segmentProgress := (targetDistance - currentDistance) / segmentLength
			return Position{
				X: segmentStart.X + dx*segmentProgress,
				Y: segmentStart.Y + dy*segmentProgress,
			}
		}

		currentDistance += segmentLength
	}

	// 如果超出範圍，返回最後一個點
	return route.Points[len(route.Points)-1]
}

// interpolateSmoothRoute 使用 Catmull-Rom 樣條進行平滑插值
func (fm *FishFormationManager) interpolateSmoothRoute(route *FishRoute, progress float64) Position {
	points := route.Points
	n := len(points)

	if n < 3 {
		return fm.interpolateLinearRoute(route, progress)
	}

	// 計算應該在哪個線段上
	segmentProgress := progress * float64(n-1)
	segmentIndex := int(segmentProgress)

	if segmentIndex >= n-1 {
		return points[n-1]
	}

	t := segmentProgress - float64(segmentIndex)

	// 獲取四個控制點
	var p0, p1, p2, p3 Position

	// p1 和 p2 是當前線段的兩個端點
	p1 = points[segmentIndex]
	p2 = points[segmentIndex+1]

	// p0 是前一個點（如果存在）
	if segmentIndex > 0 {
		p0 = points[segmentIndex-1]
	} else {
		// 外推第一個點
		p0 = Position{
			X: 2*p1.X - p2.X,
			Y: 2*p1.Y - p2.Y,
		}
	}

	// p3 是後一個點（如果存在）
	if segmentIndex+2 < n {
		p3 = points[segmentIndex+2]
	} else {
		// 外推最後一個點
		p3 = Position{
			X: 2*p2.X - p1.X,
			Y: 2*p2.Y - p1.Y,
		}
	}

	return catmullRomSpline(t, p0, p1, p2, p3)
}

// interpolateLinearRoute 使用線性插值（原始方法）
func (fm *FishFormationManager) interpolateLinearRoute(route *FishRoute, progress float64) Position {
	totalLength := fm.calculateRouteLength(route)
	targetDistance := progress * totalLength

	currentDistance := 0.0
	for i := 1; i < len(route.Points); i++ {
		segmentStart := route.Points[i-1]
		segmentEnd := route.Points[i]

		dx := segmentEnd.X - segmentStart.X
		dy := segmentEnd.Y - segmentStart.Y
		segmentLength := math.Sqrt(dx*dx + dy*dy)

		if currentDistance+segmentLength >= targetDistance {
			segmentProgress := (targetDistance - currentDistance) / segmentLength
			return Position{
				X: segmentStart.X + dx*segmentProgress,
				Y: segmentStart.Y + dy*segmentProgress,
			}
		}

		currentDistance += segmentLength
	}

	return route.Points[len(route.Points)-1]
}