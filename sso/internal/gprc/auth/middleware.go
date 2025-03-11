package grpc_auth

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"sso/sso/internal/lib/ratelimiter"
	"sso/sso/pkg/utils"
	"strings"
)

func checkMethods(methods []string, infoMethod string) bool {
	for _, method := range methods {
		if strings.Contains(infoMethod, method) {
			return true
		}
	}
	return false
}

func AuthInterceptor(limiter *ratelimiter.RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		rateCheckMethods := []string{
			"/Login",
		}

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

		if checkMethods(rateCheckMethods, info.FullMethod) {
			if limiter.IsBlocked(ctx, clientIP) {
				return nil, status.Error(codes.ResourceExhausted, "Too many requests from your IP")
			}

			// rate limit
			limiter.UberLimiter.Take()

			if err := limiter.CheckAndIncrementAttempts(ctx, clientIP); err != nil {
				return nil, status.Error(codes.ResourceExhausted, "Too many requests from your IP")
			}
		}

		resp, err = handler(ctx, req)
		if err == nil {
			if checkMethods(rateCheckMethods, info.FullMethod) {
				err := limiter.ResetAttempts(ctx, clientIP)
				if err != nil {
					return nil, status.Error(codes.Internal, "internal error")
				}
			}
		}

		return resp, err
	}
}
