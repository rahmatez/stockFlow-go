.PHONY: dev up down migrate run build test tidy seed

dev: up
	@sleep 2
	@cp -n .env.example .env 2>/dev/null || true
	go run ./cmd/api

up:
	docker compose up -d

down:
	docker compose down

migrate:
	go run ./cmd/api &
	@sleep 3
	@kill $$! 2>/dev/null || true

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

seed:
	@echo "Use POST /api/v1/auth/register to create tenant and user"
