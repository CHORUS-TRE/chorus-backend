TEST_CONFIG_FILE ?= ./../../../configs/ci_local/backend-config.yaml
DB_HOST          ?= 127.0.0.1
COVERAGE_DIR     ?= tests/coverage
REPORT           ?= acceptance

# Optional selectors: make test-unit PKG=workspace, make test-acceptance SUITE=user
PKG   ?=
SUITE ?=

UNIT_TARGET       = $(if $(PKG),./pkg/$(PKG)/service,./...)
ACCEPTANCE_TARGET = $(if $(SUITE),./tests/acceptance/$(SUITE),./tests/acceptance/...)

.PHONY: help deps deps-down deps-clean build run protos test-unit test-integration test-acceptance test-acceptance-coverage coverage-html clean

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

run: ## Run the backend with the dev config
	go run ./cmd/chorus/main.go start --config configs/config.yaml | go run ./cmd/logger/main.go

export-default-config: ## Print the code-level default configuration
	@go run ./cmd/chorus/main.go export-default-config

diff-config: ## Show drift between configs/config.yaml and the code-level defaults
	@go run ./cmd/chorus/main.go diff-config --config configs/config.yaml

jwks: ## Generate a JWKS for services.openid_connect_provider.jwks (add -public-key for the Keycloak one-liner too)
	@go run ./cmd/generate-jwks

trim-config: ## Remove fields from configs/config.yaml that are redundant with the code-level defaults (backs up to configs/config.yaml.bak first)
	@cp configs/config.yaml configs/config.yaml.bak
	@go run ./cmd/chorus/main.go trim-config --config configs/config.yaml > configs/config.yaml.tmp
	@mv configs/config.yaml.tmp configs/config.yaml
	@echo "Trimmed configs/config.yaml (previous version backed up to configs/config.yaml.bak)"

protos: ## Regenerate protobuf / gateway / openapi code
	./scripts/generate-protos.sh

test-unit: ## Run unit tests (PKG=<domain> for a single service package)
	@mkdir -p $(COVERAGE_DIR)
	go test -count=1 --tags unit $(UNIT_TARGET) -cover -coverprofile=$(COVERAGE_DIR)/unit.out

test-integration: ## Run integration tests (embedded postgres)
	@mkdir -p $(COVERAGE_DIR)
	go test -count=1 --tags integration -p 1 ./... -coverprofile=$(COVERAGE_DIR)/integration.out

test-acceptance: ## Run acceptance suites against a running backend (SUITE=<name> for one)
	TEST_CONFIG_FILE="$(TEST_CONFIG_FILE)" go test -count=1 -p 1 --tags acceptance $(ACCEPTANCE_TARGET) -args --ginkgo.junit-report=junit.xml

test-acceptance-coverage: ## Acceptance suites against an instrumented backend + coverage report
	DB_HOST="$(DB_HOST)" COVERAGE_DIR="$(COVERAGE_DIR)" ./scripts/run_acceptance_coverage.sh $(SUITE)

coverage-html: ## Open an HTML coverage report (REPORT=acceptance|unit|integration|all)
	go tool cover -html=$(COVERAGE_DIR)/$(REPORT).out

clean: ## Remove build and coverage artifacts
	rm -rf bin tests/coverage tests/acceptance/*/junit.xml
