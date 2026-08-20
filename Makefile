CONFIG_FILE           ?= configs/config.yaml
COVERAGE_DIR          ?= tests/coverage
REPORT                ?= acceptance

# Optional selectors: make test-unit PKG=workspace, make test-acceptance SUITE=user
PKG   ?=
SUITE ?=

UNIT_TARGET       = $(if $(PKG),./pkg/$(PKG)/service,./...)
ACCEPTANCE_TARGET = $(if $(SUITE),./tests/acceptance/$(SUITE),./tests/acceptance/...)

.PHONY: help deps deps-down deps-clean build run migrate protos test-unit test-integration test-acceptance test-acceptance-coverage coverage-html clean trim-config diff-config check-config

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-26s\033[0m %s\n", $$1, $$2}'

deps: ## Start local dependencies (postgres, minio)
	docker compose -f docker/compose.yml up -d

deps-down: ## Stop local dependencies
	docker compose -f docker/compose.yml down

deps-clean: ## Stop local dependencies and remove their volumes (postgres/minio data)
	docker compose -f docker/compose.yml down -v

build: ## Build the backend binary into bin/chorus
	go build -o bin/chorus ./cmd/chorus

migrate: ## Run pending migrations against the dev datastores (chorus, audit)
	go run ./cmd/chorus/main.go migrate chorus --config $(CONFIG_FILE) --set storage.migrations.chorus.username=admin --set storage.migrations.chorus.password=password
	go run ./cmd/chorus/main.go migrate audit --config $(CONFIG_FILE) --set storage.migrations.audit.username=admin --set storage.migrations.audit.password=password

run: migrate ## Run the backend with the dev config
	go run ./cmd/chorus/main.go start --config $(CONFIG_FILE) | go run ./cmd/logger/main.go

diff-config: ## Show drift between CONFIG_FILE and the code-level defaults, from live source
	@go run ./cmd/chorus/main.go diff-config --config $(CONFIG_FILE)

check-config: ## Validate CONFIG_FILE against the validation rules, from live source
	@go run ./cmd/chorus/main.go check-config --config $(CONFIG_FILE)

trim-config: ## Remove fields from CONFIG_FILE that are redundant with the code-level defaults (backs up to $(CONFIG_FILE).bak first)
	@cp $(CONFIG_FILE) $(CONFIG_FILE).bak
	@go run ./cmd/chorus/main.go trim-config --config $(CONFIG_FILE) > $(CONFIG_FILE).tmp
	@mv $(CONFIG_FILE).tmp $(CONFIG_FILE)
	@echo "Trimmed $(CONFIG_FILE) (previous version backed up to $(CONFIG_FILE).bak)"

protos: ## Regenerate protobuf / gateway / openapi code
	./scripts/generate-protos.sh

test-unit: ## Run unit tests (PKG=<domain> for a single service package)
	@mkdir -p $(COVERAGE_DIR)
	go test -count=1 --tags unit $(UNIT_TARGET) -cover -coverprofile=$(COVERAGE_DIR)/unit.out

test-integration: ## Run integration tests (embedded postgres)
	@mkdir -p $(COVERAGE_DIR)
	go test -count=1 --tags integration -p 1 ./... -coverprofile=$(COVERAGE_DIR)/integration.out

test-acceptance: ## Run acceptance test against a dedicated backend (SUITE=<suite> for a single suite)
	./scripts/run_acceptance_tests.sh --coverage $(SUITE)

coverage-html: ## Open an HTML coverage report (REPORT=acceptance|unit|integration|all)
	go tool cover -html=$(COVERAGE_DIR)/$(REPORT).out

clean: ## Remove build and coverage artifacts
	rm -rf bin tests/coverage tests/acceptance/*/junit.xml
