package grpcapp

import (
	"context"
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"net/http"
	"sso/sso/internal/config"
	gprc_metrics "sso/sso/internal/gprc"
	grpc_auth "sso/sso/internal/gprc/auth"
	gprc_user "sso/sso/internal/gprc/user"
	"sso/sso/internal/lib/logger/sl"
	"sso/sso/internal/storage/ratelimiter"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	registry   *prometheus.Registry
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
	auth grpc_auth.Auth, user gprc_user.UserManagement,
	limiter *ratelimiter.RateLimiter,
) *App {

	// Создаём реестр метрик
	grpcMetrics := grpc_prometheus.NewServerMetrics()

	reg := prometheus.NewRegistry()
	reg.MustRegister(grpcMetrics)

	// Дополнительные пользовательские метрики (например, счетчик ошибок)
	reg.MustRegister(
		gprc_metrics.ErrorCounter,
		gprc_metrics.RequestDuration,
	)

	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(chainUnaryInterceptors(
		gprc_user.UserInterceptor(cfg),
		grpc_auth.AuthInterceptor(limiter),
		gprc_metrics.MetricsUnaryInterceptor,
	)),
		grpc.ChainUnaryInterceptor(
			grpcMetrics.UnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			grpcMetrics.StreamServerInterceptor(),
		),
	)

	// Включаем серверную рефлексию (полезно для отладки)
	reflection.Register(gRPCServer)

	// Включаем сбор метрик для gRPC
	grpcMetrics.InitializeMetrics(gRPCServer)

	grpc_auth.Register(gRPCServer, auth)
	gprc_user.Register(gRPCServer, user)

	return &App{
		log:        log,
		port:       cfg.GRPC.Port,
		gRPCServer: gRPCServer,
		registry:   reg,
	}
}

func (a *App) MustRun() {
	go func() {
		// Поднимаем HTTP сервер для экспорта метрик
		http.Handle("/metrics", promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{}))
		a.log.Debug("Prometheus метрики доступны на :9090/metrics")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			a.log.Error("Ошибка запуска HTTP сервера:", sl.Err(err))
		}
	}()

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

	if err := a.gRPCServer.Serve(l); err != nil {
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
