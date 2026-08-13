.PHONY: run build test vet migrate db-up db-down import-master import-boundaries import-kodepos import-roads import-all

# Run API server locally
run:
	go run ./cmd/api

build:
	go build -o bin/geomap-api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

# Database (Docker Compose)
db-up:
	docker compose up -d postgres redis

db-down:
	docker compose down

migrate:
	docker compose exec -T postgres psql -U postgres -d geomapping_id -f - < migrations/001_schema.sql

# Data import (requires cloned data in ./data, see README)
import-master:
	go run ./scripts/import_master

import-boundaries:
	go run ./scripts/import_boundaries

import-kodepos:
	go run ./scripts/import_kodepos

import-roads:
	bash ./scripts/import_roads/run.sh

import-all: import-master import-kodepos import-boundaries
