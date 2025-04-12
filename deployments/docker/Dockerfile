# syntax=docker/dockerfile:1.4

# Этап сборки
FROM golang:1.24.1-alpine AS builder

# Устанавливаем зависимости для CGO (необходимы для SQLite)
RUN --mount=type=cache,target=/var/cache/apk \
    apk add --no-cache gcc musl-dev sqlite-dev

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

RUN apk add --no-cache sqlite-libs
WORKDIR /app
COPY --from=builder /build/sso /app/

# CMD из Dockerfile будет использован, так как команда из compose не перезаписывается
CMD ["/app/sso"]