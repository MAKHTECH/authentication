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

migrate:
	go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

generate:
	protoc --go_out=../protos/gen/go/sso/ --go_opt=paths=source_relative --go-grpc_out=../protos/gen/go/sso/ --go-grpc_opt=paths=source_relative ../protos/proto/sso/sso.proto

prometheus:
	prometheus --config.file=config/prometheus.yml --web.listen-address=":9091"