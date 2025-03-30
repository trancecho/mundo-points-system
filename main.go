package main

import (
	"fmt"
	gw_sdk "github.com/trancecho/mundo-gateway-sdk"
	"github.com/trancecho/mundo-points-system/config"
	"github.com/trancecho/mundo-points-system/domain"
	"github.com/trancecho/mundo-points-system/interceptors"
	"github.com/trancecho/mundo-points-system/pkg/utils"
	"github.com/trancecho/mundo-points-system/po/repository"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"
	"github.com/trancecho/mundo-points-system/initialize"
	pb "github.com/trancecho/mundo-points-system/proto/point/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// 初始化配置
	config.InitConfig()
	// 顺序不能错
	utils.InitSecret()

	// 初始化数据库连接
	db := initialize.InitDB()

	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", viper.GetInt("grpc.port")))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	//创建实现
	userRepo := repository.NewUserRepository(db)
	pointRepo := repository.NewPointRepository(db)
	statRepo := repository.NewStatisticsRepository(db)
	// 创建带有JWT拦截器的gRPC服务器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.JWTInterceptor()),
	)
	pb.RegisterUserServiceServer(grpcServer, domain.NewUserService(userRepo, pointRepo, statRepo))

	// 注册反射服务
	reflection.Register(grpcServer)

	// 打印已注册的服务和方法
	serviceInfo := grpcServer.GetServiceInfo()
	log.Println("注册的gRPC服务列表:")
	for svc, info := range serviceInfo {
		log.Printf("服务: %s", svc)
		for _, method := range info.Methods {
			log.Printf("  - 方法: %s", method.Name)
		}
	}

	// 启动 gRPC 服务器
	go func() {
		log.Printf("Starting gRPC server on port %d", viper.GetInt("grpc.port"))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	gatewaySDK := gw_sdk.NewGatewaySDK(viper.GetString("service.name"), viper.GetString("gateway.mundo.myaddr"), "grpc", viper.GetString("gateway.mundo.url"))
	// 自动注册 gRPC 路由到网关
	if err = gatewaySDK.AutoRegisterGRPCRoutes(grpcServer, "internal_message"); err != nil {
		log.Printf("无法自动注册gRPC路由: %v", err)
	} else {
		log.Println("所有gRPC路由已成功自动注册")
	}

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// 关闭服务
	//cancel()
	grpcServer.GracefulStop()
	log.Println("Server shutdown gracefully")
}
