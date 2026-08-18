.PHONY: run test race vet build web-install web-test web-build migrate seed compose-up compose-down

run:
	go run ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build ./...

web-install:
	cd web && npm install

web-test:
	cd web && npm test

web-build:
	cd web && npm run build

migrate:
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f migrations/001_init.sql -f migrations/002_seed.sql -f migrations/003_indexes.sql

seed:
	STORE_MODE=postgres SEED_DEMO=true go run ./cmd/server

compose-up:
	docker compose up --build

compose-down:
	docker compose down
