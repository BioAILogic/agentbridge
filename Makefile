.PHONY: build run sqlc migrate-local

build:
	go build -o bin/synbridge ./cmd/synbridge

run: build
	./bin/synbridge

sqlc:
	sqlc generate

migrate-local:
	psql $$DATABASE_URL -f internal/db/schema.sql
