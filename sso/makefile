run:
	go run cmd/sso/app.go

migrate:
	go run ./cmd/migrator --storage-path=./storage/sso.db --migrations-path=./migrations

generate:
	protoc --go_out=../protos/gen/go/sso/ --go_opt=paths=source_relative --go-grpc_out=../protos/gen/go/sso/ --go-grpc_opt=paths=source_relative ../protos/proto/sso/sso.proto

prometheus:
	prometheus --config.file=config/prometheus.yml --web.listen-address=":9091"