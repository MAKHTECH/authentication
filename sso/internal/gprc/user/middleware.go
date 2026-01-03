package gprc_user

import (
	"context"
	"fmt"
	"sso/sso/internal/config"
	user_jwt "sso/sso/internal/lib/jwt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UserInterceptor add user information to context
func UserInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		skipTokenMethods := []string{
			"/Login", "/Register",
			"/Logout", "/RefreshToken",
		}

		for _, skipMethod := range skipTokenMethods {
			if strings.Contains(info.FullMethod, skipMethod) {
				resp, err = handler(ctx, req)
				return resp, err
			}
		}

		// Get AccessToken from binary metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to get metadata from context")
		}

		// Получаем токен из метаданных
		// Приоритет: authorization (обычный, для Postman/grpcurl) -> authorization-bin (бинарный, для Go клиентов)
		var accessToken string

		// 1. Пробуем получить из authorization (удобно для тестирования в Postman)
		authorization := md.Get("authorization")
		if len(authorization) > 0 && authorization[0] != "" {
			accessToken = authorization[0]
		}

		// 2. Fallback на authorization-bin (для Go клиентов, gRPC автоматически декодирует из Base64)
		if accessToken == "" {
			authBinary := md.Get("authorization-bin")
			if len(authBinary) > 0 && authBinary[0] != "" {
				accessToken = authBinary[0]
			}
		}

		// 3. Если токен не найден
		if accessToken == "" {
			return nil, status.Error(codes.PermissionDenied, "authorization token not found")
		}

		data, err := user_jwt.ParseToken(accessToken, true, cfg.PrivateKey)
		if err != nil {
			fmt.Printf("Failed to parse token: %v\nToken length: %d\nToken preview: %.50s...\n", err, len(accessToken), accessToken)
			return nil, status.Error(codes.PermissionDenied, "invalid access token")
		}

		// write data to context ==> *models.AccessTokenData
		ctx = context.WithValue(ctx, "data", data)

		resp, err = handler(ctx, req)
		return resp, err
	}
}
