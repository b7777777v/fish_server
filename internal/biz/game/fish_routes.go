package game

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// ========================================
// 魚群路線管理系統
// ========================================

// initializeDefaultRoutes 初始化預設路線
func (fm *FishFormationManager) initializeDefaultRoutes() {
	// 直線路線 - 從左到右
	fm.routes["straight_left_right"] = &FishRoute{
		ID:       "straight_left_right",
		Name:     "左右直線",
		Points: []Position{
			{X: -100, Y: fm.roomConfig.RoomHeight / 2},
			{X: fm.roomConfig.RoomWidth + 100, Y: fm.roomConfig.RoomHeight / 2},
		},
		Duration:   time.Duration(10 * time.Second),
		Type:      RouteTypeStraight,
		Difficulty: 0.8,
		Looping:   false,
		Smooth:    false,
		CreatedAt: time.Now(),
	}

	// 直線路線 - 從右到左
	fm.routes["straight_right_left"] = &FishRoute{
		ID:       "straight_right_left",
		Name:     "右左直線",
		Points: []Position{
			{X: fm.roomConfig.RoomWidth + 100, Y: fm.roomConfig.RoomHeight / 2},
			{X: -100, Y: fm.roomConfig.RoomHeight / 2},
		},
		Duration:   time.Duration(10 * time.Second),
		Type:      RouteTypeStraight,
		Difficulty: 0.8,
		Looping:   false,
		Smooth:    false,
		CreatedAt: time.Now(),
	}

	// 對角線路線 - 左上到右下
	fm.routes["diagonal_top_left"] = &FishRoute{
		ID:       "diagonal_top_left",
		Name:     "對角線(左上右下)",
		Points: []Position{
			{X: -50, Y: -50},
			{X: fm.roomConfig.RoomWidth + 50, Y: fm.roomConfig.RoomHeight + 50},
		},
		Duration:   time.Duration(12 * time.Second),
		Type:      RouteTypeStraight,
		Difficulty: 0.9,
		Looping:   false,
		Smooth:    false,
		CreatedAt: time.Now(),
	}

	// 對角線路線 - 右上到左下
	fm.routes["diagonal_top_right"] = &FishRoute{
		ID:       "diagonal_top_right",
		Name:     "對角線(右上左下)",
		Points: []Position{
			{X: fm.roomConfig.RoomWidth + 50, Y: -50},
			{X: -50, Y: fm.roomConfig.RoomHeight + 50},
		},
		Duration:   time.Duration(12 * time.Second),
		Type:      RouteTypeStraight,
		Difficulty: 0.9,
		Looping:   false,
		Smooth:    false,
		CreatedAt: time.Now(),
	}

	// S型曲線路線
	fm.routes["s_curve"] = &FishRoute{
		ID:       "s_curve",
		Name:     "S型曲線",
		Points: []Position{
			{X: -100, Y: fm.roomConfig.RoomHeight * 0.2},
			{X: fm.roomConfig.RoomWidth * 0.3, Y: fm.roomConfig.RoomHeight * 0.8},
			{X: fm.roomConfig.RoomWidth * 0.7, Y: fm.roomConfig.RoomHeight * 0.2},
			{X: fm.roomConfig.RoomWidth + 100, Y: fm.roomConfig.RoomHeight * 0.8},
		},
		Duration:   time.Duration(15 * time.Second),
		Type:      RouteTypeCurved,
		Difficulty: 1.2,
		Looping:   false,
		Smooth:    true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// Z字型路線
	fm.routes["zigzag"] = &FishRoute{
		ID:       "zigzag",
		Name:     "Z字型",
		Points: []Position{
			{X: -50, Y: fm.roomConfig.RoomHeight * 0.1},
			{X: fm.roomConfig.RoomWidth * 0.8, Y: fm.roomConfig.RoomHeight * 0.1},
			{X: fm.roomConfig.RoomWidth * 0.2, Y: fm.roomConfig.RoomHeight * 0.5},
			{X: fm.roomConfig.RoomWidth * 0.8, Y: fm.roomConfig.RoomHeight * 0.5},
			{X: fm.roomConfig.RoomWidth * 0.2, Y: fm.roomConfig.RoomHeight * 0.9},
			{X: fm.roomConfig.RoomWidth + 50, Y: fm.roomConfig.RoomHeight * 0.9},
		},
		Duration:   time.Duration(18 * time.Second),
		Type:      RouteTypeZigzag,
		Difficulty: 1.4,
		Looping:   false,
		Smooth:    true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 圓形路線
	fm.routes["circle_clockwise"] = &FishRoute{
		ID:       "circle_clockwise",
		Name:     "順時針圓形",
		Points:   fm.generateCirclePoints(fm.roomConfig.RoomWidth/2, fm.roomConfig.RoomHeight/2, math.Min(fm.roomConfig.RoomWidth, fm.roomConfig.RoomHeight)*0.3, 16, false),
		Duration:  time.Duration(20 * time.Second),
		Type:     RouteTypeCircular,
		Difficulty: 1.0,
		Looping:  true,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 逆時針圓形路線
	fm.routes["circle_counterclockwise"] = &FishRoute{
		ID:       "circle_counterclockwise",
		Name:     "逆時針圓形",
		Points:   fm.generateCirclePoints(fm.roomConfig.RoomWidth/2, fm.roomConfig.RoomHeight/2, math.Min(fm.roomConfig.RoomWidth, fm.roomConfig.RoomHeight)*0.3, 16, true),
		Duration:  time.Duration(20 * time.Second),
		Type:     RouteTypeCircular,
		Difficulty: 1.0,
		Looping:  true,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 8字型路線
	fm.routes["figure_eight"] = &FishRoute{
		ID:       "figure_eight",
		Name:     "8字型",
		Points:   fm.generateFigureEightPoints(fm.roomConfig.RoomWidth/2, fm.roomConfig.RoomHeight/2, fm.roomConfig.RoomWidth*0.2, fm.roomConfig.RoomHeight*0.15),
		Duration:  time.Duration(25 * time.Second),
		Type:     RouteTypeCurved,
		Difficulty: 1.5,
		Looping:  true,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 螺旋路線 - 向內
	fm.routes["spiral_inward"] = &FishRoute{
		ID:       "spiral_inward",
		Name:     "向內螺旋",
		Points:   fm.generateSpiralPoints(fm.roomConfig.RoomWidth/2, fm.roomConfig.RoomHeight/2, math.Min(fm.roomConfig.RoomWidth, fm.roomConfig.RoomHeight)*0.4, 0, 24, true),
		Duration:  time.Duration(30 * time.Second),
		Type:     RouteTypeCurved,
		Difficulty: 1.6,
		Looping:  false,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 螺旋路線 - 向外
	fm.routes["spiral_outward"] = &FishRoute{
		ID:       "spiral_outward",
		Name:     "向外螺旋",
		Points:   fm.generateSpiralPoints(fm.roomConfig.RoomWidth/2, fm.roomConfig.RoomHeight/2, 20, math.Min(fm.roomConfig.RoomWidth, fm.roomConfig.RoomHeight)*0.4, 24, false),
		Duration:  time.Duration(30 * time.Second),
		Type:     RouteTypeCurved,
		Difficulty: 1.6,
		Looping:  false,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 波浪路線
	fm.routes["wave_horizontal"] = &FishRoute{
		ID:       "wave_horizontal",
		Name:     "水平波浪",
		Points:   fm.generateWavePoints(-100, fm.roomConfig.RoomWidth+100, fm.roomConfig.RoomHeight/2, 100, 3, 20),
		Duration:  time.Duration(16 * time.Second),
		Type:     RouteTypeCurved,
		Difficulty: 1.1,
		Looping:  false,
		Smooth:   true, // 啟用平滑插值
		CreatedAt: time.Now(),
	}

	// 三角形巡邏路線
	fm.routes["triangle_patrol"] = &FishRoute{
		ID:       "triangle_patrol",
		Name:     "三角巡邏",
		Points: []Position{
			{X: fm.roomConfig.RoomWidth * 0.2, Y: fm.roomConfig.RoomHeight * 0.2},
			{X: fm.roomConfig.RoomWidth * 0.8, Y: fm.roomConfig.RoomHeight * 0.2},
			{X: fm.roomConfig.RoomWidth * 0.5, Y: fm.roomConfig.RoomHeight * 0.8},
			{X: fm.roomConfig.RoomWidth * 0.2, Y: fm.roomConfig.RoomHeight * 0.2},
		},
		Duration:   time.Duration(22 * time.Second),
		Type:      RouteTypeCurved,
		Difficulty: 1.3,
		Looping:   true,
		CreatedAt: time.Now(),
	}

	// 隨機路線 (用於特殊事件)
	fm.routes["random_chaos"] = &FishRoute{
		ID:       "random_chaos",
		Name:     "隨機混沌",
		Points:   fm.generateRandomPoints(8),
		Duration:  time.Duration(20 * time.Second),
		Type:     RouteTypeRandom,
		Difficulty: 1.8,
		Looping:  false,
		CreatedAt: time.Now(),
	}

	fm.logger.Infof("Initialized %d default routes", len(fm.routes))
}

// generateCirclePoints 生成圓形路線點
func (fm *FishFormationManager) generateCirclePoints(centerX, centerY, radius float64, segments int, counterclockwise bool) []Position {
	points := make([]Position, segments)
	angleStep := 2 * math.Pi / float64(segments)
	
	for i := 0; i < segments; i++ {
		angle := float64(i) * angleStep
		if counterclockwise {
			angle = -angle
		}
		
		points[i] = Position{
			X: centerX + radius*math.Cos(angle),
			Y: centerY + radius*math.Sin(angle),
		}
	}
	
	return points
}

// generateFigureEightPoints 生成8字型路線點
func (fm *FishFormationManager) generateFigureEightPoints(centerX, centerY, radiusX, radiusY float64) []Position {
	points := make([]Position, 32)
	
	for i := 0; i < 32; i++ {
		t := float64(i) * 2 * math.Pi / 32
		
		// 8字型參數方程
		x := centerX + radiusX*math.Sin(t)
		y := centerY + radiusY*math.Sin(2*t)
		
		points[i] = Position{X: x, Y: y}
	}
	
	return points
}

// generateSpiralPoints 生成螺旋路線點
func (fm *FishFormationManager) generateSpiralPoints(centerX, centerY, startRadius, endRadius float64, segments int, inward bool) []Position {
	points := make([]Position, segments)
	
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		angle := t * 6 * math.Pi // 3圈
		
		var radius float64
		if inward {
			radius = startRadius + (endRadius-startRadius)*t
		} else {
			radius = startRadius + (endRadius-startRadius)*t
		}
		
		points[i] = Position{
			X: centerX + radius*math.Cos(angle),
			Y: centerY + radius*math.Sin(angle),
		}
	}
	
	return points
}

// generateWavePoints 生成波浪路線點
func (fm *FishFormationManager) generateWavePoints(startX, endX, centerY, amplitude float64, frequency float64, segments int) []Position {
	points := make([]Position, segments)
	
	for i := 0; i < segments; i++ {
		t := float64(i) / float64(segments-1)
		x := startX + (endX-startX)*t
		y := centerY + amplitude*math.Sin(frequency*2*math.Pi*t)
		
		points[i] = Position{X: x, Y: y}
	}
	
	return points
}

// generateRandomPoints 生成隨機路線點
func (fm *FishFormationManager) generateRandomPoints(count int) []Position {
	points := make([]Position, count)
	
	for i := 0; i < count; i++ {
		points[i] = Position{
			X: rand.Float64() * fm.roomConfig.RoomWidth,
			Y: rand.Float64() * fm.roomConfig.RoomHeight,
		}
	}
	
	return points
}

// CreateCustomRoute 創建自定義路線
func (fm *FishFormationManager) CreateCustomRoute(id, name string, points []Position, routeType FishRouteType, difficulty float64, looping bool) *FishRoute {
	if len(points) < 2 {
		fm.logger.Warn("Cannot create route with less than 2 points")
		return nil
	}
	
	route := &FishRoute{
		ID:         id,
		Name:       name,
		Points:     points,
		Duration:   fm.calculateRouteDuration(points, difficulty),
		Type:       routeType,
		Difficulty: difficulty,
		Looping:    looping,
		CreatedAt:  time.Now(),
	}
	
	fm.routes[id] = route
	fm.logger.Infof("Created custom route: %s with %d points", id, len(points))
	
	return route
}

// calculateRouteDuration 計算路線持續時間
func (fm *FishFormationManager) calculateRouteDuration(points []Position, difficulty float64) time.Duration {
	length := 0.0
	for i := 1; i < len(points); i++ {
		dx := points[i].X - points[i-1].X
		dy := points[i].Y - points[i-1].Y
		length += math.Sqrt(dx*dx + dy*dy)
	}
	
	// 基礎速度為每秒50像素，難度影響速度
	baseSpeed := 50.0 / difficulty
	duration := length / baseSpeed
	
	return time.Duration(duration * float64(time.Second))
}

// GetRoute 獲取路線
func (fm *FishFormationManager) GetRoute(routeID string) *FishRoute {
	return fm.routes[routeID]
}

// GetAllRoutes 獲取所有路線
func (fm *FishFormationManager) GetAllRoutes() []*FishRoute {
	routes := make([]*FishRoute, 0, len(fm.routes))
	for _, route := range fm.routes {
		routes = append(routes, route)
	}
	return routes
}

// GetRoutesByType 根據類型獲取路線
func (fm *FishFormationManager) GetRoutesByType(routeType FishRouteType) []*FishRoute {
	var routes []*FishRoute
	for _, route := range fm.routes {
		if route.Type == routeType {
			routes = append(routes, route)
		}
	}
	return routes
}

// GetRoutesByDifficulty 根據難度範圍獲取路線
func (fm *FishFormationManager) GetRoutesByDifficulty(minDifficulty, maxDifficulty float64) []*FishRoute {
	var routes []*FishRoute
	for _, route := range fm.routes {
		if route.Difficulty >= minDifficulty && route.Difficulty <= maxDifficulty {
			routes = append(routes, route)
		}
	}
	return routes
}

// RemoveRoute 移除路線
func (fm *FishFormationManager) RemoveRoute(routeID string) bool {
	if _, exists := fm.routes[routeID]; !exists {
		return false
	}
	
	// 檢查是否有陣型正在使用此路線
	for _, formation := range fm.formations {
		if formation.Route != nil && formation.Route.ID == routeID {
			fm.logger.Warnf("Cannot remove route %s: still in use by formation %s", routeID, formation.ID)
			return false
		}
	}
	
	delete(fm.routes, routeID)
	fm.logger.Infof("Removed route: %s", routeID)
	return true
}

// ModifyRoute 修改現有路線
func (fm *FishFormationManager) ModifyRoute(routeID string, points []Position, difficulty float64) bool {
	route := fm.routes[routeID]
	if route == nil {
		return false
	}
	
	if len(points) < 2 {
		fm.logger.Warn("Cannot modify route with less than 2 points")
		return false
	}
	
	route.Points = points
	route.Difficulty = difficulty
	route.Duration = fm.calculateRouteDuration(points, difficulty)
	
	fm.logger.Infof("Modified route: %s", routeID)
	return true
}

// GetRandomRoute 獲取隨機路線
func (fm *FishFormationManager) GetRandomRoute() *FishRoute {
	if len(fm.routes) == 0 {
		return nil
	}
	
	routes := fm.GetAllRoutes()
	randomIndex := rand.Intn(len(routes))
	return routes[randomIndex]
}

// GetRandomRouteByType 根據類型獲取隨機路線
func (fm *FishFormationManager) GetRandomRouteByType(routeType FishRouteType) *FishRoute {
	routes := fm.GetRoutesByType(routeType)
	if len(routes) == 0 {
		return nil
	}
	
	randomIndex := rand.Intn(len(routes))
	return routes[randomIndex]
}

// 路線驗證和優化
func (fm *FishFormationManager) ValidateRoute(route *FishRoute) []string {
	var issues []string
	
	// 檢查點數量
	if len(route.Points) < 2 {
		issues = append(issues, "路線至少需要2個點")
	}
	
	// 檢查點是否在房間範圍內
	for i, point := range route.Points {
		if point.X < -200 || point.X > fm.roomConfig.RoomWidth+200 ||
		   point.Y < -200 || point.Y > fm.roomConfig.RoomHeight+200 {
			issues = append(issues, fmt.Sprintf("點 %d 超出房間範圍", i))
		}
	}
	
	// 檢查路線長度
	length := fm.calculateRouteLength(route)
	if length < 100 {
		issues = append(issues, "路線太短")
	}
	
	// 檢查難度值
	if route.Difficulty < 0.1 || route.Difficulty > 3.0 {
		issues = append(issues, "難度值應在0.1-3.0之間")
	}
	
	return issues
}

// OptimizeRoute 優化路線（移除冗餘點，平滑路徑等）
func (fm *FishFormationManager) OptimizeRoute(route *FishRoute) *FishRoute {
	if len(route.Points) < 3 {
		return route
	}
	
	optimizedPoints := []Position{route.Points[0]}
	
	for i := 1; i < len(route.Points)-1; i++ {
		prev := route.Points[i-1]
		curr := route.Points[i]
		next := route.Points[i+1]
		
		// 計算角度變化
		angle1 := math.Atan2(curr.Y-prev.Y, curr.X-prev.X)
		angle2 := math.Atan2(next.Y-curr.Y, next.X-curr.X)
		angleDiff := math.Abs(angle1 - angle2)
		
		// 如果角度變化顯著，保留這個點
		if angleDiff > 0.1 {
			optimizedPoints = append(optimizedPoints, curr)
		}
	}
	
	optimizedPoints = append(optimizedPoints, route.Points[len(route.Points)-1])
	
	optimizedRoute := *route
	optimizedRoute.Points = optimizedPoints
	optimizedRoute.Duration = fm.calculateRouteDuration(optimizedPoints, route.Difficulty)
	
	return &optimizedRoute
}