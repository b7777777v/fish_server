// internal/app/game/server.go
package game

import (
	"context"

	"github.com/b7777777v/fish_server/internal/biz/player"
	pb "github.com/b7777777v/fish_server/pkg/pb/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GameServer 實現了 pb.GameServer 接口
type GameServer struct {
	pb.UnimplementedGameServer // 必須嵌入，以確保向前相容

	playerUsecase *player.PlayerUsecase
}

// NewGameServer 創建一個 GameServer
func NewGameServer(playerUsecase *player.PlayerUsecase) *GameServer {
	return &GameServer{
		playerUsecase: playerUsecase,
	}
}

// Login 處理 gRPC 的登入請求
func (s *GameServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := s.playerUsecase.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "login failed: %v", err)
	}

	return &pb.LoginResponse{Token: token}, nil
}

// GrpcGameApp 表示遊戲 gRPC 應用，管理 gRPC 伺服器
type GrpcGameApp struct {
	GrpcServer *grpc.Server
}

// NewGrpcGameApp 創建並註冊 gRPC 伺服器
func NewGrpcGameApp(gameServer *GameServer) *GrpcGameApp {
	grpcSrv := grpc.NewServer()
	pb.RegisterGameServer(grpcSrv, gameServer)

	return &GrpcGameApp{
		GrpcServer: grpcSrv,
	}
}
