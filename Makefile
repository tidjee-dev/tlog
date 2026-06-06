MODULE    := github.com/tidjee-dev/tlog
BIN       := tlog
BUILD_DIR := ./dist

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.date=$(DATE)'

# ANSI styles
BOLD  := \033[1m
CYAN  := \033[36m
WHITE := \033[1;37m
DIM   := \033[2m
RESET := \033[0m

.DEFAULT_GOAL := help

.PHONY: \
	all \
	build \
	build-release \
	test \
	lint \
	bench \
	clean \
	tidy \
	release \
	tag \
	help

help: ## Show this help message
	@printf "\n$(WHITE)  tlog$(RESET) $(DIM)— structured terminal logger$(RESET)\n\n"
	@printf "  $(DIM)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@printf "  $(DIM)Targets:$(RESET)\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "    $(CYAN)%-16s$(RESET) $(DIM)%s$(RESET)\n", $$1, $$2}'
	@printf "\n"

all: lint test build ## Run lint, test, and build

build: ## Build debug binary into ./dist
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BIN) .

build-release: ## Build optimized release binary
	@printf "\n$(WHITE)Building release binary...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BIN) \
		.

test: ## Run tests with race detector
	go test -race -count=1 ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

bench: ## Run benchmarks
	go test -bench=. -benchmem -run='^$$' ./...

tidy: ## Tidy go.mod and go.sum
	go mod tidy

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR)

tag: ## Create and push a git tag (usage: make tag v=v1.0.0)
	@test -n "$(v)" || (echo "Usage: make tag v=v1.0.0" && exit 1)
	git tag $(v)
	git push origin $(v)

release: clean lint test build-release ## Full release pipeline
	@printf "\n$(WHITE)Release build completed$(RESET)\n\n"
	@printf "  Version: $(CYAN)$(VERSION)$(RESET)\n"
	@printf "  Commit:  $(CYAN)$(COMMIT)$(RESET)\n"
	@printf "  Binary:  $(CYAN)$(BUILD_DIR)/$(BIN)$(RESET)\n\n"
