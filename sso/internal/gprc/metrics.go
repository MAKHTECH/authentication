package gprc_metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"time"
)

// RequestDuration Гистограмма времени выполнения
var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "grpc_request_duration_seconds",
		Help:    "Время выполнения gRPC-запросов",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"},
)

// ErrorCounter Счетчик ошибок
var ErrorCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "grpc_errors_total",
		Help: "Количество ошибок gRPC",
	},
	[]string{"method", "error_type"},
)

////ActiveConnections Счетчик активных соединений
//var ActiveConnections = prometheus.NewGaugeVec(
//	prometheus.GaugeOpts{
//		Name: "grpc_active_connections",
//		Help: "Количество активных соединений gRPC",
//	},
//	[]string{"method"},
//)

func MetricsUnaryInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error,
) {
	start := time.Now()

	resp, err = handler(ctx, req)

	// записываем метрику выполнения handlers
	duration := time.Since(start).Seconds()
	RequestDuration.WithLabelValues(info.FullMethod).Observe(duration)

	// метрика для счетчика ошибок
	if err != nil {
		ErrorCounter.WithLabelValues(info.FullMethod, "internal").Inc()
	}

	return
}
