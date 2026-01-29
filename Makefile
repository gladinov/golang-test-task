# =========================
# Variables
# =========================

COMPOSE_FILE = deployments/docker-composes/docker-compose.yml
TEST_COMPOSE_FILE = deployments/docker-composes/docker-compose.test.yml

ENV_FILE = deployments/envs/prod.env
TEST_ENV_FILE = deployments/envs/test.env

GO_TEST_FLAGS = -count=1 -cover -covermode=atomic
COVER_DIR = build/coverage
UNIT_COVER_FILE = $(COVER_DIR)/unit.out
INTEG_COVER_FILE = $(COVER_DIR)/integration.out
COVER_FILE = $(COVER_DIR)/coverage.out

# =========================
# Phony targets
# =========================

.PHONY: up build-up down \
        up-test build-up-test down-test \
        test test-unit test-integration \
        wait-test-db \
        lint coverage coverage-all

# =========================
# Docker: production
# =========================

up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d

build-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v

# =========================
# Docker: test database
# =========================

up-test:
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) up -d

build-up-test:
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) up -d --build

down-test:
	docker compose -f $(TEST_COMPOSE_FILE) --env-file $(TEST_ENV_FILE) down -v

# =========================
# Helpers
# =========================

wait-test-db:
	@echo "Waiting for test database..."
	sleep 3

# ====================
# Unit tests 
# ====================
test-unit:
	@mkdir -p $(COVER_DIR)
	@echo "=== Running unit tests ==="
	go test -tags=unit ./... $(GO_TEST_FLAGS) -coverprofile=$(UNIT_COVER_FILE)

# ====================
# Integration tests (с тестовой БД)
# ====================
test-integration: build-up-test wait-test-db
	@mkdir -p $(COVER_DIR)
	@echo "=== Running integration tests ==="
	@status=0; \
	go test -tags=integration ./... $(GO_TEST_FLAGS) -coverprofile=$(INTEG_COVER_FILE) || status=$$?; \
	$(MAKE) down-test; \
	exit $$status

# ====================
# Combine coverage
# ====================
coverage: test-unit test-integration
	@mkdir -p $(COVER_DIR)
	@echo "=== Merging coverage files ==="
	@echo "mode: atomic" > $(COVER_FILE)
	@tail -n +2 $(UNIT_COVER_FILE) >> $(COVER_FILE) || true
	@tail -n +2 $(INTEG_COVER_FILE) >> $(COVER_FILE) || true
	@echo ""
	@echo "=== Coverage by package ==="
	go tool cover -func=$(COVER_FILE)
	@echo ""
	@echo "=== Total coverage ==="
	go tool cover -func=$(COVER_FILE) | grep total


