package utils

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// AddAuthTokenToContext добавляет accessToken в метаданные gRPC контекста
// Используется ключ с суффиксом "-bin" — gRPC автоматически кодирует значение в Base64 при отправке
// и декодирует при получении. Поэтому мы передаём сырой токен без ручного кодирования.
// Это решает проблему с недопустимыми символами в метаданных gRPC
func AddAuthTokenToContext(ctx context.Context, accessToken string) context.Context {
	// gRPC автоматически кодирует бинарные метаданные (-bin) в Base64
	// Не нужно кодировать вручную!
	md := metadata.Pairs("authorization-bin", accessToken)

	// Объединяем с существующими метаданными, если они есть
	if existingMd, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existingMd, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// AddAuthTokenToContextWithUserAgent добавляет токен и User-Agent в метаданные
func AddAuthTokenToContextWithUserAgent(ctx context.Context, accessToken string, userAgent string) context.Context {
	// gRPC автоматически кодирует бинарные метаданные (-bin) в Base64
	md := metadata.Pairs(
		"authorization-bin", accessToken,
		"user-agent", userAgent,
	)

	if existingMd, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existingMd, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}
