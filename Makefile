.PHONY: sqlc migrate run-api run-worker seed test tidy
sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
migrate:
	go run ./cmd/migrate
run-api:
	go run ./cmd/api
run-worker:
	go run ./cmd/worker
seed:
	go run ./cmd/seed
test:
	go test ./...
tidy:
	go mod tidy
