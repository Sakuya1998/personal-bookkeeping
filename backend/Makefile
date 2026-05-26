GO       := /opt/data/go/bin/go
GOCACHE  := /tmp/gocache
GOPROXY  := https://goproxy.cn,direct

BIN      := server
CMD      := ./cmd/server

# Docker
IMAGE    := personal-bookkeeping-backend
COMPOSE  := docker compose
COMPOSE_FILE := ../docker-compose.yml

.PHONY: build run run/hot test vet tidy clean
.PHONY: logs logs/tail logs/errors logs/truncate
.PHONY: docker/build docker/run docker/up docker/down docker/logs docker/ps docker/rebuild
.PHONY: migration/new migration/up migration/down migration/version
.PHONY: swagger swagger/serve swagger/clean

# ── Go ──────────────────────────────────────────────────────

build:
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) build -o $(BIN) $(CMD)

run: build
	./$(BIN)

run/hot: build
	./$(BIN) 2>&1 | jq -Rr 'fromjson? // .'

test:
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test ./... -v -count=1

test/short:
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test ./... -short -count=1

test/db:
	@echo "Creating test database if not exists..."
	-docker exec bookkeeping-db psql -U bookkeeper -d postgres -c "CREATE DATABASE bookkeeping_test;" 2>/dev/null || true

test/run: test/db
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) test ./internal/app/handler/ -v -count=1 -timeout 60s

vet:
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) vet ./...

tidy:
	GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) $(GO) mod tidy

clean:
	rm -f $(BIN)
	rm -rf logs/
	rm -rf /tmp/gocache

# ── Logs ────────────────────────────────────────────────────

logs:
	@ls -la logs/ 2>/dev/null || echo "logs/ not found"

logs/tail:
	tail -f logs/app.log 2>/dev/null || echo "logs/app.log not found"

logs/errors:
	tail -f logs/error.log 2>/dev/null || echo "logs/error.log not found"

logs/truncate:
	: > logs/app.log 2>/dev/null || true
	: > logs/warn.log 2>/dev/null || true
	: > logs/error.log 2>/dev/null || true

# ── Migrations ──────────────────────────────────────────────

MIGRATE_DSN ?= "postgres://bookkeeper:***@localhost:5432/bookkeeping?sslmode=disable"
MIGRATE_DIR := internal/infra/migrations

migration/new:
	@read -p "Migration name: " name; \
	files=$$(ls $(MIGRATE_DIR)/ 2>/dev/null | grep -E '^[0-9]+_' | tail -1); \
	if [ -z "$$files" ]; then \
		seq=000001; \
	else \
		last=$$(echo "$$files" | grep -oE '^[0-9]+' | tail -1); \
		seq=$$(printf "%06d" $$((10#$$last + 1))); \
	fi; \
	touch $(MIGRATE_DIR)/$${seq}_$${name}.up.sql; \
	touch $(MIGRATE_DIR)/$${seq}_$${name}.down.sql; \
	echo "created $(MIGRATE_DIR)/$${seq}_$${name}.{up,down}.sql"

migration/up:
	migrate -path $(MIGRATE_DIR) -database $(MIGRATE_DSN) up

migration/down:
	migrate -path $(MIGRATE_DIR) -database $(MIGRATE_DSN) down

migration/version:
	migrate -path $(MIGRATE_DIR) -database $(MIGRATE_DSN) version

# ── Swagger ────────────────────────────────────────────────

SWAG := /opt/data/home/go/bin/swag

swagger: swagger/clean
	GOROOT=/opt/data/go PATH="/opt/data/go/bin:$(SWAG_DIR):$$PATH" GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY) GO111MODULE=on $(SWAG) init -g cmd/server/main.go --output internal/infra/swagger --parseInternal

swagger/serve:
	@echo "Open http://localhost:8000/swagger/index.html"
	@echo ""

swagger/clean:
	rm -rf internal/infra/swagger

# ── Docker ──────────────────────────────────────────────────

docker/build:
	docker build -t $(IMAGE) .

docker/run:
	docker run --rm -p 8000:8000 --name $(IMAGE) $(IMAGE)

docker/up:
	cd .. && $(COMPOSE) -f docker-compose.yml up -d

docker/down:
	cd .. && $(COMPOSE) -f docker-compose.yml down

docker/logs:
	cd .. && $(COMPOSE) -f docker-compose.yml logs -f

docker/ps:
	cd .. && $(COMPOSE) -f docker-compose.yml ps

docker/rebuild:
	cd .. && $(COMPOSE) -f docker-compose.yml up -d --build

.DEFAULT_GOAL := build
