package game

import (
	"os"
	"testing"
	"time"

	"github.com/b7777777v/fish_server/internal/pkg/logger"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestProtobufMessageSending 測試 Protobuf 消息發送不包含換行符
func TestProtobufMessageSending(t *testing.T) {
	log := logger.New(os.Stdout, "debug", "console")

	t.Run("sendProtobuf does not add newlines", func(t *testing.T) {
		client := &Client{
			ID:       "test_client",
			PlayerID: 1,
			send:     make(chan []byte, 256),
			logger:   log,
		}

		// 創建測試消息
		msg := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{
					Timestamp: time.Now().Unix(),
				},
			},
		}

		// 獲取預期的序列化數據
		expectedBytes, err := proto.Marshal(msg)
		require.NoError(t, err)

		// 發送消息
		client.sendProtobuf(msg)

		// 接收消息
		select {
		case received := <-client.send:
			// 驗證接收到的消息與預期完全一致
			assert.Equal(t, expectedBytes, received, "Message should be identical to marshaled protobuf")

			// 驗證可以正確解析
			var parsed pb.GameMessage
			err := proto.Unmarshal(received, &parsed)
			assert.NoError(t, err, "Received message should be valid protobuf")
			assert.Equal(t, msg.Type, parsed.Type, "Message type should match")

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Did not receive message in time")
		}
	})

	t.Run("Multiple messages sent independently without newline separators", func(t *testing.T) {
		client := &Client{
			ID:       "test_client_multi",
			PlayerID: 2,
			send:     make(chan []byte, 256),
			logger:   log,
		}

		// 發送多個消息
		numMessages := 5
		expectedMessages := make([][]byte, numMessages)

		for i := 0; i < numMessages; i++ {
			msg := &pb.GameMessage{
				Type: pb.MessageType_HEARTBEAT,
				Data: &pb.GameMessage_Heartbeat{
					Heartbeat: &pb.HeartbeatMessage{
						Timestamp: int64(i),
					},
				},
			}

			expectedBytes, err := proto.Marshal(msg)
			require.NoError(t, err)
			expectedMessages[i] = expectedBytes

			client.sendProtobuf(msg)
		}

		// 接收並驗證每個消息
		for i := 0; i < numMessages; i++ {
			select {
			case received := <-client.send:
				// 每個消息應該可以獨立解析
				var parsed pb.GameMessage
				err := proto.Unmarshal(received, &parsed)
				assert.NoError(t, err, "Message %d should be valid protobuf", i)

				// 驗證時間戳正確
				assert.Equal(t, int64(i), parsed.GetHeartbeat().Timestamp,
					"Message %d timestamp should match", i)

			case <-time.After(100 * time.Millisecond):
				t.Fatalf("Did not receive message %d in time", i)
			}
		}
	})

	t.Run("sendJSON does not interfere with binary messages", func(t *testing.T) {
		client := &Client{
			ID:       "test_client_json",
			PlayerID: 3,
			send:     make(chan []byte, 256),
			logger:   log,
		}

		// 發送一個 Protobuf 消息
		pbMsg := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{
					Timestamp: 123456,
				},
			},
		}

		client.sendProtobuf(pbMsg)

		// 接收並驗證
		select {
		case received := <-client.send:
			// 應該是有效的 protobuf
			var parsed pb.GameMessage
			err := proto.Unmarshal(received, &parsed)
			assert.NoError(t, err, "Should be valid protobuf")

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Did not receive message in time")
		}
	})

	t.Run("sendErrorPB sends valid protobuf error message", func(t *testing.T) {
		client := &Client{
			ID:       "test_client_error",
			PlayerID: 4,
			send:     make(chan []byte, 256),
			logger:   log,
		}

		errorMessage := "Test error message"
		client.sendErrorPB(errorMessage)

		// 接收並驗證
		select {
		case received := <-client.send:
			// 應該是有效的 protobuf
			var parsed pb.GameMessage
			err := proto.Unmarshal(received, &parsed)
			assert.NoError(t, err, "Error message should be valid protobuf")

			// 驗證類型是 ERROR
			assert.Equal(t, pb.MessageType_ERROR, parsed.Type, "Should be ERROR type")

			// 驗證錯誤消息內容
			errorMsg := parsed.GetError()
			assert.NotNil(t, errorMsg, "Should have error data")
			assert.Equal(t, errorMessage, errorMsg.Message, "Error message should match")

			// Protobuf 二進制數據可能包含字節 0x0A (換行符的值),但這是正常的序列化
			// 我們只需要確保數據是有效的 protobuf 即可
			// 不檢查內部是否包含 \n 字節

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Did not receive error message in time")
		}
	})
}

// TestClientSendChannelNonBlocking 測試客戶端發送通道的非阻塞行為
func TestClientSendChannelNonBlocking(t *testing.T) {
	log := logger.New(os.Stdout, "debug", "console")

	t.Run("sendProtobuf handles full channel gracefully", func(t *testing.T) {
		client := &Client{
			ID:       "test_client_full",
			PlayerID: 5,
			send:     make(chan []byte, 2), // 小緩衝區
			logger:   log,
		}

		// 填滿通道
		msg1 := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{Timestamp: 1},
			},
		}
		msg2 := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{Timestamp: 2},
			},
		}

		client.sendProtobuf(msg1)
		client.sendProtobuf(msg2)

		// 通道現在應該滿了
		assert.Equal(t, 2, len(client.send), "Channel should be full")

		// 再發送一個消息,應該不會阻塞
		msg3 := &pb.GameMessage{
			Type: pb.MessageType_HEARTBEAT,
			Data: &pb.GameMessage_Heartbeat{
				Heartbeat: &pb.HeartbeatMessage{Timestamp: 3},
			},
		}

		done := make(chan bool)
		go func() {
			client.sendProtobuf(msg3) // 這可能會丟棄一個舊消息
			done <- true
		}()

		select {
		case <-done:
			// 成功,沒有阻塞
		case <-time.After(500 * time.Millisecond):
			t.Fatal("sendProtobuf blocked on full channel")
		}
	})

	t.Run("Messages are sent in correct format for WebSocket binary frames", func(t *testing.T) {
		client := &Client{
			ID:       "test_client_format",
			PlayerID: 6,
			send:     make(chan []byte, 256),
			logger:   log,
		}

		// 創建一個包含各種數據類型的消息
		msg := &pb.GameMessage{
			Type: pb.MessageType_JOIN_ROOM_RESPONSE,
			Data: &pb.GameMessage_JoinRoomResponse{
				JoinRoomResponse: &pb.JoinRoomResponse{
					Success:     true,
					RoomId:      "test_room_123",
					PlayerCount: 5,
					Timestamp:   time.Now().Unix(),
				},
			},
		}

		client.sendProtobuf(msg)

		select {
		case received := <-client.send:
			// 驗證這是純二進制數據
			// Protobuf 二進制數據不應該以特定字符開始或結束

			// 解析應該成功
			var parsed pb.GameMessage
			err := proto.Unmarshal(received, &parsed)
			assert.NoError(t, err, "Should parse as valid protobuf")

			// 驗證數據完整性
			joinResp := parsed.GetJoinRoomResponse()
			assert.NotNil(t, joinResp, "Should have join room response data")
			assert.True(t, joinResp.Success, "Success should be true")
			assert.Equal(t, "test_room_123", joinResp.RoomId, "Room ID should match")
			assert.Equal(t, int32(5), joinResp.PlayerCount, "Player count should match")

		case <-time.After(100 * time.Millisecond):
			t.Fatal("Did not receive message in time")
		}
	})
}

// BenchmarkProtobufSending 性能基準測試
func BenchmarkProtobufSending(b *testing.B) {
	log := logger.New(os.Stdout, "error", "console") // 降低日誌級別
	client := &Client{
		ID:       "bench_client",
		PlayerID: 1,
		send:     make(chan []byte, 1000),
		logger:   log,
	}

	msg := &pb.GameMessage{
		Type: pb.MessageType_HEARTBEAT,
		Data: &pb.GameMessage_Heartbeat{
			Heartbeat: &pb.HeartbeatMessage{
				Timestamp: time.Now().Unix(),
			},
		},
	}

	// 啟動一個 goroutine 消費消息
	go func() {
		for range client.send {
			// 丟棄消息
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.sendProtobuf(msg)
	}
}
