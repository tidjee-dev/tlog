MODULE    := github.com/tidjee-dev/tlog
BIN       := tlog
BUILD_DIR := ./dist

# ANSI styles
BOLD  := \033[1m
CYAN  := \033[36m
WHITE := \033[1;37m
DIM   := \033[2m
RESET := \033[0m

.DEFAULT_GOAL := help

.PHONY: all build test lint bench clean tidy help

help: ## Show this help message
	@printf "\n$(WHITE)  tlog$(RESET) $(DIM)— structured terminal logger$(RESET)\n\n"
	@printf "  $(DIM)Usage:$(RESET)  make $(CYAN)<target>$(RESET)\n\n"
	@printf "  $(DIM)Targets:$(RESET)\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "    $(CYAN)%-10s$(RESET) $(DIM)%s$(RESET)\n", $$1, $$2}'
	@printf "\n"

all: lint test build ## Run lint, test, and build

build: ## Build the binary into ./dist
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BIN) ./cmd/...

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
