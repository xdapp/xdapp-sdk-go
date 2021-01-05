package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/xdapp/xdapp-sdk-go/example/grpc/pb"
)

const (
	port = ":7777"
)

type UserService struct {}

func (userService *UserService) HelloWorld(ctx context.Context, req *pb.TextCheckReq) (*pb.TextCheckResp, error) {
	log.Printf("HelloWorld 执行中")

	return &pb.TextCheckResp{Status:1, Message: "success", Resp: "hello", AdminId: 1, AdminName: "施利鸣"}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// 创建 RPC 服务容器
	grpcServer := grpc.NewServer()

	pb.RegisterTextCheckServer(grpcServer, &UserService{})

	reflection.Register(grpcServer)

	log.Printf("##### server 已开启 ####")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}