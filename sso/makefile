run:
	go run cmd/sso/app.go

# Запуск с hot reload через Air (локально)
dev:
	air -c ../.air.toml

# Docker: сборка и запуск dev окружения с hot reload (watch mode)
docker-dev:
	docker compose -f ../deployments/docker/docker-compose.dev.yml watch

# Docker: запуск без watch (обычный up)
docker-dev-up:
	docker compose -f ../deployments/docker/docker-compose.dev.yml up --build

# Docker: пересборка без кеша
docker-dev-rebuild:
	docker compose -f ../deployments/docker/docker-compose.dev.yml build --no-cache
	docker compose -f ../deployments/docker/docker-compose.dev.yml watch

# Docker: остановка dev окружения
docker-dev-down:
	docker compose -f ../deployments/docker/docker-compose.dev.yml down

database:
	docker compose -f ../deployments/docker/docker-compose.dev.yml exec -it postgres psql -U sso -d sso

# Миграции через мигратор для PostgreSQL
migrate:
	go run ./cmd/migrator --migrations-path=./migrations

# Подключение к PostgreSQL через Docker контейнер
psql:
	docker compose -f ../deployments/docker/docker-compose.dev.yml exec postgres psql -U sso -d sso

generate:
	protoc -I ../protos/proto --go_out=../protos/gen/go/ --go_opt=paths=source_relative --go-grpc_out=../protos/gen/go/ --go-grpc_opt=paths=source_relative ../protos/proto/sso/*.proto

generate-jwt:
	go run cmd/genjwt/generate.go

prometheus:
	prometheus --config.file=config/prometheus.yml --web.listen-address=":9091"