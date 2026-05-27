.PHONY: all backend-test backend-lint frontend-test frontend-lint \
        test lint build coverage coverage-check vet vulncheck docker-build \
        tidy ci

# =============================================================================
# 后端
# =============================================================================

GO := /opt/data/go/bin/go
GOCACHE := /tmp/gocache
GOPROXY := https://goproxy.cn,direct
GOFLAGS := GOCACHE=$(GOCACHE) GOPROXY=$(GOPROXY)

backend-tidy:
	cd backend && $(GOFLAGS) $(GO) mod tidy

backend-vet:
	cd backend && $(GOFLAGS) $(GO) vet ./...

backend-test:
	cd backend && $(GOFLAGS) $(GO) test -count=1 -short -race ./...

backend-test-full:
	cd backend && $(GOFLAGS) $(GO) test -count=1 -race ./...

backend-build:
	cd backend && $(GOFLAGS) $(GO) build ./...

backend-lint: backend-tidy backend-vet backend-build

# govulncheck (install: go install golang.org/x/vuln/cmd/govulncheck@latest)
backend-vulncheck:
	cd backend && which govulncheck >/dev/null 2>&1 && $(GOFLAGS) $(GO) run golang.org/x/vuln/cmd/govulncheck ./... || echo "govulncheck not installed, skipping"

# Coverage: 对纯逻辑包设阈值 (handler 跳过, 因为需 PG)
COVERAGE_PKGS := ./internal/app/service/... ./internal/app/task/... \
                 ./internal/app/middleware/... \
                 ./internal/infra/cache/... ./internal/infra/queue/...
COVERAGE_THRESHOLD := 30

coverage-report:
	cd backend && $(GOFLAGS) $(GO) test -count=1 -short -coverprofile=coverage.out \
		-covermode=atomic $(COVERAGE_PKGS) 2>/dev/null
	cd backend && $(GOFLAGS) $(GO) tool cover -func=coverage.out | tail -20

coverage-check:
	@cd backend && $(GOFLAGS) $(GO) test -count=1 -short -coverprofile=coverage.out \
		-covermode=atomic $(COVERAGE_PKGS) 2>/dev/null
	@cd backend && total=$$($(GOFLAGS) $(GO) tool cover -func=coverage.out | grep '^total' | awk '{print $$NF}' | tr -d '%'); \
		echo "Total coverage: $$total%"; \
		if [ "$$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc)" = "1" ]; then \
			echo "FAIL: coverage $$total% < $(COVERAGE_THRESHOLD)%"; \
			exit 1; \
		fi; \
		echo "PASS: coverage $$total% >= $(COVERAGE_THRESHOLD)%"

backend-ci: backend-tidy backend-vet backend-test coverage-check

# =============================================================================
# 前端
# =============================================================================

frontend-install:
	cd frontend && npm ci

frontend-lint:
	cd frontend && npx eslint src/

frontend-typecheck:
	cd frontend && npx tsc -b --noEmit

frontend-test:
	cd frontend && npx vitest run

frontend-build:
	cd frontend && npm run build

frontend-ci: frontend-install frontend-lint frontend-typecheck frontend-test frontend-build

# =============================================================================
# Docker
# =============================================================================

setup:
	bash scripts/setup.sh

docker-build:
	docker build -t personal-bookkeeping-backend:ci ./backend
	docker build -t personal-bookkeeping-frontend:ci ./frontend

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# =============================================================================
# 开发
# =============================================================================

test: backend-test frontend-test
lint: backend-lint frontend-lint
build: backend-build frontend-build
ci: backend-ci frontend-ci docker-build

all: lint test build

# =============================================================================
# 清理
# =============================================================================

clean:
	rm -f backend/coverage.out
	rm -rf frontend/dist
