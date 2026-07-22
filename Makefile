.PHONY: run build test tidy migrate-up migrate-down docker-up docker-down fmt vet sqlc

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

sqlc:
	sqlc generate

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/marketplace?sslmode=disable" down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-reset:
	docker compose down -v && docker compose up -d --build