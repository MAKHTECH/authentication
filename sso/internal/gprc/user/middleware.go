package gprc_user

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sso/sso/internal/config"
	user_jwt "sso/sso/internal/lib/jwt"
	"strings"
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

		// Get AccessToken
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to get metadata from context")
		}

		authorization := md.Get("authorization")
		if len(authorization) <= 0 {
			return nil, status.Error(codes.PermissionDenied, "authorization token not found")
		}
		accessToken := authorization[0]

		data, err := user_jwt.ParseToken(accessToken, true, cfg.Secret)
		if err != nil {
			fmt.Println(err)
			return nil, status.Error(codes.PermissionDenied, "invalid access token")
		}

		// write data to context ==> *models.AccessTokenData
		ctx = context.WithValue(ctx, "data", data)

		resp, err = handler(ctx, req)
		return resp, err
	}
}
