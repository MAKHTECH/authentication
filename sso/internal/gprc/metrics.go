package gprc_metrics

import (
	"context"
	"google.golang.org/grpc"
	"sso/sso/internal/domain/models"
	"sso/sso/internal/lib/kafka"
	"time"
)

// MetricsUnaryInterceptor интерцептор
func MetricsUnaryInterceptor(producer *kafka.Producer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error,
	) {
		start := time.Now()

		resp, err = handler(ctx, req)

		// todo panic: send on closed channel
		// записываем метрику выполнения handlers
		duration := time.Since(start).Seconds()
		producer.SendMetric(map[string]interface{}{
			"method":   info.FullMethod,
			"duration": duration,
		}, models.HandlerDurationTopic)

		// метрика для счетчика ошибок
		if err != nil {
			producer.SendMetric(map[string]interface{}{
				"method":     info.FullMethod,
				"error_type": "internal",
			}, models.HandlerDurationTopic)
		}

		return
	}
}
