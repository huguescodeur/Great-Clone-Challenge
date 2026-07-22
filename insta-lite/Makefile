.PHONY: up down logs run run-api run-db build swag clean ps migrate migrate-stamp migrate-status migrate-docker migrate-docker-stamp test test-integration

## Docker
up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f api

ps:
	docker compose ps

run-db:
	docker compose up -d postgres redis

clean:
	docker compose down -v
	rm -rf bin/

## Local dev
run-api:
	@set -a; . ./.env; set +a; go run ./cmd/server

run: run-api

## Build
build:
	go build -o bin/server ./cmd/server

## Swagger
swag:
	swag init -g cmd/server/main.go -o docs

## Migrations
migrate:
	@set -a; . ./.env; set +a; go run ./cmd/migrate up

migrate-stamp:
	@set -a; . ./.env; set +a; go run ./cmd/migrate stamp

migrate-status:
	@set -a; . ./.env; set +a; go run ./cmd/migrate status

migrate-docker:
	@set -a; . ./.env; set +a; DATABASE_URL=postgres://postgres:$$DB_PASSWORD@localhost:5433/insta go run ./cmd/migrate up

migrate-docker-stamp:
	@set -a; . ./.env; set +a; DATABASE_URL=postgres://postgres:$$DB_PASSWORD@localhost:5433/insta go run ./cmd/migrate stamp

## Tests
test:
	go test ./...

test-integration:
	TEST_DATABASE_URL=postgres://postgres:${DB_PASSWORD}@localhost:5433/insta \
	TEST_REDIS_ADDR=localhost:6379 \
	go test -tags integration -count=1 ./...
