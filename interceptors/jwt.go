package interceptors

import (
	"context"
	"strings"

	"github.com/trancecho/mundo-points-system/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// JWTInterceptor 创建一个用于验证JWT的拦截器
func JWTInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 从元数据中获取token
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "metadata不存在")
		}

		// 获取Authorization header
		authorization := md.Get("authorization")
		if len(authorization) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "authorization header不存在")
		}

		// 提取token
		token := strings.TrimPrefix(authorization[0], "Bearer ")
		if token == authorization[0] {
			return nil, status.Errorf(codes.Unauthenticated, "token格式错误")
		}

		// 验证token
		claims, err := utils.ParseToken("mundo", token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token无效: %v", err)
		}

		// 将claims的具体字段添加到上下文中
		newCtx := context.WithValue(ctx, "claims", claims)
		newCtx = context.WithValue(newCtx, "user_id", claims.UserID)
		newCtx = context.WithValue(newCtx, "username", claims.Username)
		newCtx = context.WithValue(newCtx, "role", claims.Role)

		// 继续处理请求
		return handler(newCtx, req)
	}
}
