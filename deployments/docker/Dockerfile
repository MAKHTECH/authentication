# syntax=docker/dockerfile:1.4

# Этап сборки
FROM golang:1.25-alpine AS builder

# Устанавливаем gcc и musl-dev для CGO (необходимо для lib/pq)
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Копируем файлы модулей для лучшего кеширования
COPY go.mod go.sum ./

# Загрузка зависимостей с использованием кеша
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download -x

# Копируем исходный код
COPY . .

# Сборка приложения с использованием кеша для артефактов сборки
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o sso ./sso/cmd/sso

# Финальный этап
FROM alpine:latest

WORKDIR /app

# Создаем директорию для конфигурации
RUN mkdir -p /app/config

COPY --from=builder /build/sso /app/
COPY --from=builder /build/sso/config/local.json /app/config/local.json

# Устанавливаем переменные окружения
ENV CONFIG_PATH=/app/config/local.json

# CMD из Dockerfile будет использован, так как команда из compose не перезаписывается
CMD ["/app/sso"]