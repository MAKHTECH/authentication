package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sso/sso/internal/config"
	gprc_metrics "sso/sso/internal/gprc"
	grpc_auth "sso/sso/internal/gprc/auth"
	grpc_transactions "sso/sso/internal/gprc/transactions"
	gprc_user "sso/sso/internal/gprc/user"
	"sso/sso/internal/lib/kafka"
	"sso/sso/internal/lib/ratelimiter"
	"sso/sso/internal/services/auth"
	"sso/sso/internal/services/transactions"
	"sso/sso/internal/services/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	producer   *kafka.Producer
	port       int
}

// Добавление несколько interceptors
func chainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var chainHandler grpc.UnaryHandler = handler
		for i := len(interceptors) - 1; i >= 0; i-- {
			innerHandler := chainHandler
			currentInterceptor := interceptors[i]
			chainHandler = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return currentInterceptor(currentCtx, currentReq, info, innerHandler)
			}
		}
		return chainHandler(ctx, req)
	}
}

func New(
	log *slog.Logger, cfg *config.Config,
	producer *kafka.Producer,
	limiter *ratelimiter.RateLimiter,
	auth auth.Auth,
	user user.UserManagement,
	transactions transactions.TransactionsManagement,
) *App {

	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(chainUnaryInterceptors(
		gprc_user.UserInterceptor(cfg),
		grpc_auth.AuthInterceptor(limiter),
		gprc_metrics.MetricsUnaryInterceptor(producer),
	)))

	// Включаем серверную рефлексию (полезно для отладки)
	reflection.Register(gRPCServer)

	grpc_auth.Register(gRPCServer, auth)
	gprc_user.Register(gRPCServer, user)
	grpc_transactions.Register(gRPCServer, user, transactions)

	return &App{
		log:        log,
		port:       cfg.GRPC.Port,
		gRPCServer: gRPCServer,
	}
}

func (a *App) MustRun() {
	if err := a.run(); err != nil {
		panic(err)
	}
}

func (a *App) run() error {
	const op = "grpcapp.Run"

	a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	if err = a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(
		slog.String("op", op),
	).Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}
