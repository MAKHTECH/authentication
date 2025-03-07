package grpc_auth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sso/sso/pkg/utils"
)

// LoggingInterceptor Унарный перехватчика
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	{
		//get header, convert to http.Header and insert header to context and
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Internal, "failed to get metadata from context")
		}
		ctx = context.WithValue(ctx, "fingerprint", utils.GetFingerprint(md))

		// Get Client IP and User-Agent
		clientIP, err := utils.GetGRPCClientIP(ctx, md)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to get client IP from metadata")
		}
		ctx = context.WithValue(ctx, "ip", clientIP)
		ctx = context.WithValue(ctx, "user-agent", md.Get("user-agent")[0])
	}

	resp, err = handler(ctx, req)
	return resp, err
}
