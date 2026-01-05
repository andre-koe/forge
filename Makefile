# ============================================================================
# FORGE - Production Makefile
# ============================================================================

# Variablen
BINARY_NAME := forge
MODULE := github.com/andre-koe/forge
MAIN_PATH := ./cmd/forge
BUILD_DIR := bin
COVERAGE_DIR := coverage

# Build Info
VERSION ?= v0.0.0-dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILDDATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go-Einstellungen
GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w -X $(MODULE)/pkg/version.Version=$(VERSION) -X $(MODULE)/pkg/version.BuildDate=$(BUILD_TIME) -X $(MODULE)/pkg/version.Commit=$(GIT_COMMIT)

# Versionierung
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Betriebssystem-Erkennung
GOOS ?= $(shell $(GO) env GOOS)
GOARCH ?= $(shell $(GO) env GOARCH)

# Farben für Output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m

.PHONY: all build build-all clean test test-coverage lint fmt vet check deps tidy run install help

# Default-Target
all: check build

# ============================================================================
# BUILD
# ============================================================================

## build: Baut die Anwendung für das aktuelle OS
build:
	@echo "$(GREEN)▸ Building $(BINARY_NAME) $(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "$(GREEN)✓ Built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-linux: Baut für Linux
build-linux:
	@echo "$(GREEN)▸ Building for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

## build-darwin: Baut für macOS
build-darwin:
	@echo "$(GREEN)▸ Building for macOS...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

## build-windows: Baut für Windows
build-windows:
	@echo "$(GREEN)▸ Building for Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

## build-all: Baut für alle Plattformen
build-all: build-linux build-darwin build-windows
	@echo "$(GREEN)✓ All builds completed$(NC)"

# ============================================================================
# DEVELOPMENT
# ============================================================================

## run: Startet die Anwendung
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

## install: Installiert die Anwendung in $GOPATH/bin
install:
	@echo "$(GREEN)▸ Installing $(BINARY_NAME)...$(NC)"
	$(GO) install $(GOFLAGS) -ldflags "$(LDFLAGS)" $(MAIN_PATH)
	@echo "$(GREEN)✓ Installed to $(shell $(GO) env GOPATH)/bin/$(BINARY_NAME)$(NC)"

## dev: Startet im Entwicklungsmodus mit Live-Reload (benötigt air)
dev:
	@command -v air > /dev/null || (echo "$(YELLOW)Installing air...$(NC)" && go install github.com/air-verse/air@latest)
	air

# ============================================================================
# QUALITÄTSSICHERUNG
# ============================================================================

## test: Führt alle Tests aus
test:
	@echo "$(GREEN)▸ Running tests...$(NC)"
	$(GO) test -race -v ./...

## test-short: Führt schnelle Tests aus
test-short:
	@echo "$(GREEN)▸ Running short tests...$(NC)"
	$(GO) test -short -v ./...

## test-coverage: Führt Tests mit Coverage-Report aus
test-coverage:
	@echo "$(GREEN)▸ Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)✓ Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

## bench: Führt Benchmarks aus
bench:
	@echo "$(GREEN)▸ Running benchmarks...$(NC)"
	$(GO) test -bench=. -benchmem ./...

## lint: Führt golangci-lint aus
lint:
	@echo "$(GREEN)▸ Running linter...$(NC)"
	@command -v golangci-lint > /dev/null || (echo "$(YELLOW)Installing golangci-lint...$(NC)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

## fmt: Formatiert den Code
fmt:
	@echo "$(GREEN)▸ Formatting code...$(NC)"
	$(GO) fmt ./...
	@command -v goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

## vet: Führt go vet aus
vet:
	@echo "$(GREEN)▸ Running go vet...$(NC)"
	$(GO) vet ./...

## vuln: Prüft auf Sicherheitslücken
vuln:
	@echo "$(GREEN)▸ Checking for vulnerabilities...$(NC)"
	@command -v govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

## check: Führt alle Qualitätsprüfungen aus
check: fmt vet lint

# ============================================================================
# DEPENDENCIES
# ============================================================================

## deps: Installiert alle Abhängigkeiten
deps:
	@echo "$(GREEN)▸ Downloading dependencies...$(NC)"
	$(GO) mod download

## tidy: Bereinigt go.mod und go.sum
tidy:
	@echo "$(GREEN)▸ Tidying modules...$(NC)"
	$(GO) mod tidy

## update: Aktualisiert alle Abhängigkeiten
update:
	@echo "$(GREEN)▸ Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy

## verify: Verifiziert Abhängigkeiten
verify:
	@echo "$(GREEN)▸ Verifying dependencies...$(NC)"
	$(GO) mod verify

# ============================================================================
# DOCKER
# ============================================================================

## docker-build: Baut das Docker-Image
docker-build:
	@echo "$(GREEN)▸ Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

## docker-run: Startet den Docker-Container
docker-run:
	docker run --rm -it $(BINARY_NAME):latest

# ============================================================================
# CLEANUP
# ============================================================================

## clean: Löscht Build-Artefakte
clean:
	@echo "$(GREEN)▸ Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f coverage.out
	$(GO) clean -cache -testcache
	@echo "$(GREEN)✓ Cleaned$(NC)"

# ============================================================================
# RELEASE
# ============================================================================

## release-check: Prüft ob Release möglich ist
release-check:
	@echo "$(GREEN)▸ Checking release requirements...$(NC)"
	@test -n "$(shell git status --porcelain)" && echo "$(RED)✗ Working directory not clean$(NC)" && exit 1 || true
	@echo "$(GREEN)✓ Ready for release$(NC)"

## changelog: Generiert Changelog (benötigt git-chglog)
changelog:
	@command -v git-chglog > /dev/null || (echo "$(YELLOW)Installing git-chglog...$(NC)" && go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest)
	git-chglog -o CHANGELOG.md

# ============================================================================
# HELP
# ============================================================================

## help: Zeigt diese Hilfe an
help:
	@echo "$(GREEN)Forge Makefile$(NC)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
