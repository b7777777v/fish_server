// cmd/test-player/main.go
// 测试玩家创建和游戏流程验证工具

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	pb "fish_server/api/proto/v1"
)

const (
	defaultAdminURL = "http://localhost:6060"
	defaultGameURL  = "ws://localhost:9090"
)

var (
	adminURL   = flag.String("admin", defaultAdminURL, "Admin server URL")
	gameURL    = flag.String("game", defaultGameURL, "Game server WebSocket URL")
	username   = flag.String("username", "", "Username for test player (required)")
	password   = flag.String("password", "test123456", "Password for test player")
	createOnly = flag.Bool("create-only", false, "Only create player without testing game flow")
	verbose    = flag.Bool("verbose", false, "Enable verbose logging")
)

type TestPlayer struct {
	Username string
	Password string
	Token    string
	UserID   int64
	Nickname string
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
	IsGuest   bool   `json:"is_guest"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func main() {
	flag.Parse()

	if *username == "" {
		fmt.Println("错误: 必须提供用户名")
		fmt.Println("使用方式: go run main.go -username <用户名> [-password <密码>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	player := &TestPlayer{
		Username: *username,
		Password: *password,
	}

	fmt.Println("🐟 鱼游戏测试工具")
	fmt.Println("==================")
	fmt.Printf("Admin Server: %s\n", *adminURL)
	fmt.Printf("Game Server:  %s\n", *gameURL)
	fmt.Printf("测试用户:     %s\n", player.Username)
	fmt.Println()

	// 步骤1: 创建/注册玩家
	if err := registerPlayer(player); err != nil {
		log.Printf("❌ 注册失败（可能已存在）: %v", err)
		log.Println("尝试直接登入...")
	} else {
		fmt.Printf("✅ 玩家注册成功: %s\n", player.Username)
	}

	// 步骤2: 登入
	if err := loginPlayer(player); err != nil {
		log.Fatalf("❌ 登入失败: %v", err)
	}
	fmt.Printf("✅ 登入成功\n")
	fmt.Printf("   Token: %s...\n", player.Token[:50])
	fmt.Printf("   用户ID: %d\n", player.UserID)
	fmt.Printf("   昵称: %s\n", player.Nickname)
	fmt.Println()

	// 步骤3: 获取玩家信息
	if err := getPlayerProfile(player); err != nil {
		log.Printf("⚠️  获取玩家资料失败: %v", err)
	} else {
		fmt.Println("✅ 玩家资料验证成功")
	}
	fmt.Println()

	if *createOnly {
		fmt.Println("✅ 测试玩家创建完成（仅创建模式）")
		return
	}

	// 步骤4: 连接到游戏服务器
	fmt.Println("📡 连接到游戏服务器...")
	if err := testGameFlow(player); err != nil {
		log.Fatalf("❌ 游戏流程测试失败: %v", err)
	}

	fmt.Println()
	fmt.Println("🎉 所有测试通过！")
}

// registerPlayer 注册新玩家
func registerPlayer(player *TestPlayer) error {
	reqBody := RegisterRequest{
		Username: player.Username,
		Password: player.Password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/register", *adminURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("注册失败 [%d]: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("注册失败 [%d]: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	player.Token = authResp.Token
	player.UserID = authResp.User.ID
	player.Nickname = authResp.User.Nickname

	return nil
}

// loginPlayer 登入玩家
func loginPlayer(player *TestPlayer) error {
	reqBody := LoginRequest{
		Username: player.Username,
		Password: player.Password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/auth/login", *adminURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("登入失败 [%d]: %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("登入失败 [%d]: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	player.Token = authResp.Token
	player.UserID = authResp.User.ID
	player.Nickname = authResp.User.Nickname

	return nil
}

// getPlayerProfile 获取玩家资料
func getPlayerProfile(player *TestPlayer) error {
	url := fmt.Sprintf("%s/api/v1/user/profile", *adminURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+player.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("获取资料失败 [%d]: %s", resp.StatusCode, string(body))
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if *verbose {
		fmt.Printf("   ID: %d\n", user.ID)
		fmt.Printf("   用户名: %s\n", user.Username)
		fmt.Printf("   昵称: %s\n", user.Nickname)
		fmt.Printf("   头像: %s\n", user.AvatarURL)
		fmt.Printf("   游客: %v\n", user.IsGuest)
	}

	return nil
}

// testGameFlow 测试游戏流程
func testGameFlow(player *TestPlayer) error {
	// 构建WebSocket URL，包含token作为查询参数
	u, err := url.Parse(*gameURL)
	if err != nil {
		return fmt.Errorf("解析游戏服务器URL失败: %w", err)
	}

	// 添加token到查询参数
	q := u.Query()
	q.Set("token", player.Token)
	u.RawQuery = q.Encode()

	// 连接WebSocket
	fmt.Printf("连接到: %s\n", u.String())
	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("WebSocket连接失败: %w", err)
	}
	defer ws.Close()

	fmt.Println("✅ WebSocket连接成功")

	// 创建context用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理中断信号
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 消息接收通道
	done := make(chan struct{})
	messages := make(chan *pb.GameMessage, 10)

	// 启动消息接收goroutine
	go func() {
		defer close(done)
		for {
			_, data, err := ws.ReadMessage()
			if err != nil {
				if *verbose {
					log.Printf("读取消息错误: %v", err)
				}
				return
			}

			var msg pb.GameMessage
			if err := proto.Unmarshal(data, &msg); err != nil {
				log.Printf("解析消息失败: %v", err)
				continue
			}

			messages <- &msg
		}
	}()

	// 测试流程
	testSteps := []struct {
		name string
		fn   func() error
	}{
		{"等待欢迎消息", func() error { return waitForWelcome(messages) }},
		{"获取房间列表", func() error { return getRoomList(ws, messages) }},
		{"发送心跳", func() error { return sendHeartbeat(ws, messages) }},
		{"获取玩家信息", func() error { return getPlayerInfo(ws, messages) }},
	}

	for _, step := range testSteps {
		fmt.Printf("📋 %s...\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s失败: %w", step.name, err)
		}
		fmt.Printf("   ✅ %s成功\n", step.name)
		time.Sleep(500 * time.Millisecond)
	}

	// 等待一段时间或中断
	select {
	case <-done:
		fmt.Println("连接已关闭")
	case <-interrupt:
		fmt.Println("\n收到中断信号，正在关闭...")
		cancel()
	case <-time.After(2 * time.Second):
		fmt.Println("测试完成")
	}

	return nil
}

// waitForWelcome 等待欢迎消息
func waitForWelcome(messages <-chan *pb.GameMessage) error {
	select {
	case msg := <-messages:
		if msg.Type == pb.MessageType_WELCOME {
			if *verbose {
				fmt.Printf("   收到欢迎消息: %s\n", msg.GetWelcome().GetMessage())
			}
			return nil
		}
		return fmt.Errorf("期望WELCOME消息，收到: %s", msg.Type.String())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("等待欢迎消息超时")
	}
}

// getRoomList 获取房间列表
func getRoomList(ws *websocket.Conn, messages <-chan *pb.GameMessage) error {
	msg := &pb.GameMessage{
		Type:        pb.MessageType_GET_ROOM_LIST,
		GetRoomList: &pb.GetRoomListRequest{},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 等待响应
	select {
	case resp := <-messages:
		if resp.Type == pb.MessageType_GET_ROOM_LIST_RESPONSE {
			rooms := resp.GetGetRoomListResponse().GetRooms()
			if *verbose {
				fmt.Printf("   房间数量: %d\n", len(rooms))
				for i, room := range rooms {
					fmt.Printf("   房间%d: ID=%d, 玩家=%d/%d, 状态=%s\n",
						i+1, room.GetId(), room.GetCurrentPlayers(), room.GetMaxPlayers(), room.GetStatus())
				}
			}
			return nil
		}
		return fmt.Errorf("期望GET_ROOM_LIST_RESPONSE，收到: %s", resp.Type.String())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("等待房间列表响应超时")
	}
}

// sendHeartbeat 发送心跳
func sendHeartbeat(ws *websocket.Conn, messages <-chan *pb.GameMessage) error {
	msg := &pb.GameMessage{
		Type:      pb.MessageType_HEARTBEAT,
		Heartbeat: &pb.HeartbeatRequest{},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 等待响应
	select {
	case resp := <-messages:
		if resp.Type == pb.MessageType_HEARTBEAT_RESPONSE {
			if *verbose {
				fmt.Printf("   服务器时间: %d\n", resp.GetHeartbeatResponse().GetServerTime())
			}
			return nil
		}
		return fmt.Errorf("期望HEARTBEAT_RESPONSE，收到: %s", resp.Type.String())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("等待心跳响应超时")
	}
}

// getPlayerInfo 获取玩家信息
func getPlayerInfo(ws *websocket.Conn, messages <-chan *pb.GameMessage) error {
	msg := &pb.GameMessage{
		Type:          pb.MessageType_GET_PLAYER_INFO,
		GetPlayerInfo: &pb.GetPlayerInfoRequest{},
	}

	data, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	if err := ws.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return fmt.Errorf("发送消息失败: %w", err)
	}

	// 等待响应
	select {
	case resp := <-messages:
		if resp.Type == pb.MessageType_GET_PLAYER_INFO_RESPONSE {
			playerResp := resp.GetGetPlayerInfoResponse()
			if *verbose {
				fmt.Printf("   玩家ID: %d\n", playerResp.GetPlayerId())
				fmt.Printf("   用户名: %s\n", playerResp.GetUsername())
				fmt.Printf("   余额: %d\n", playerResp.GetBalance())
			}
			return nil
		}
		return fmt.Errorf("期望GET_PLAYER_INFO_RESPONSE，收到: %s", resp.Type.String())
	case <-time.After(5 * time.Second):
		return fmt.Errorf("等待玩家信息响应超时")
	}
}
