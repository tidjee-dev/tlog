MODULE    := github.com/tidjee-dev/tlog
BIN       := tlog
BUILD_DIR := ./bin

# -----------------------
# Versioning
# -----------------------

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X 'main.version=$(VERSION)' \
	-X 'main.commit=$(COMMIT)' \
	-X 'main.date=$(DATE)'

SVU ?= svu

# -----------------------
# UI
# -----------------------

UNDERLINE := \033[4m
CYAN  := \033[36m
PURPLE:= \033[35m
WHITE := \033[1;37m
DIM   := \033[2m
GREEN := \033[32m
RESET := \033[0m

# -----------------------
# Defaults
# -----------------------

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
	help \
	tag \
	version \
	prerelease \
	release \
	release-major \
	release-minor \
	release-patch

# -----------------------
# Help
# -----------------------

help: ## Show this help message
	@printf "\n$(WHITE)tlog$(RESET) $(DIM)— structured terminal logger$(RESET)\n\n"
	@printf "  $(DIM)Usage:$(RESET) make $(CYAN)<target>$(RESET)\n\n"

	@awk 'BEGIN {FS=":.*##"} \
	/^###/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)} \
	/^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

	@printf "\n"

# -----------------------
### 🧱 Core
# -----------------------

all: lint test build ## Run lint, test, build

build: ## Build debug binary
	@mkdir -p $(BUILD_DIR)
	@printf "$(DIM)→ building debug binary...$(RESET)\n"
	go build -o $(BUILD_DIR)/$(BIN) .

build-release: ## Build optimized binary
	@mkdir -p $(BUILD_DIR)
	@printf "$(DIM)→ building release binary...$(RESET)\n"
	@CGO_ENABLED=0 go build \
		-trimpath \
		-ldflags="$(LDFLAGS)" \
		-o $(BUILD_DIR)/$(BIN) \
		.

# -----------------------
### 🧪 Code Quality
# -----------------------

test: ## Run tests
	@printf "$(DIM)→ running tests...$(RESET)\n"
	go test -race -count=1 ./...

lint: ## Run golangci-lint
	@printf "$(DIM)→ linting...$(RESET)\n"
	golangci-lint run ./...

check: lint test ## Run lint and tests

bench: ## Run benchmarks
	go test -bench=. -benchmem -run='^$$' ./...

tidy: ## Tidy go modules
	go mod tidy

# -----------------------
### 🧹 Maintenance
# -----------------------

clean: ## Remove build artifacts
	@printf "$(DIM)→ cleaning dist...$(RESET)\n"
	rm -rf $(BUILD_DIR)


# -----------------------
### 🛡️  Safety
# -----------------------

prerelease: ## Ensure clean working tree
	@git diff --quiet || (echo "Working tree is dirty" && exit 1)
	@git diff --cached --quiet || (echo "Staged changes present" && exit 1)

# -----------------------
### 🚀 Release
# -----------------------

version: ## Show next version (current | major | minor | patch)
	@printf "\n$(WHITE)Versions of $(PURPLE)$(UNDERLINE)$(MODULE)$(RESET)\n\n"
	@printf "Current version: $(CYAN)$(shell svu current)$(RESET)\n\n"
	@printf "Next versions:\n"
	@printf "  Major: $(CYAN)$(shell svu major)$(RESET)\n"
	@printf "  Minor: $(CYAN)$(shell svu minor)$(RESET)\n"
	@printf "  Patch: $(CYAN)$(shell svu patch)$(RESET)\n"

# Centralized tagging
define do_tag
	@test -n "$(1)" || (echo "Missing version" && exit 1)
	@printf "$(GREEN)→ releasing $(1)...$(RESET)\n"
	@git tag -a $(1) -m "release $(1)"
	@git push origin $(1)
endef

tag: prerelease ## Create and push git tag (usage: make tag v=v1.2.3)
	$(call do_tag,$(v))

define release_summary
	@printf "\n$(GREEN)✓ release completed$(RESET)\n\n"
	@printf "  Version: $(CYAN)$(1)$(RESET)\n"
	@printf "  Commit:  $(CYAN)$(COMMIT)$(RESET)\n"
	@printf "  Binary:  $(CYAN)$(BUILD_DIR)/$(BIN)$(RESET)\n\n"
endef

release: prerelease clean lint test build-release ## Full release build
	$(call release_summary,$(VERSION))

release-major: prerelease clean lint test ## Major bump release
	$(eval V := $(shell svu major))
	$(call do_tag,$(V))
	@$(MAKE) build-release VERSION=$(V)
	$(call release_summary,$(V))

release-minor: prerelease clean lint test ## Minor bump release
	$(eval V := $(shell svu minor))
	$(call do_tag,$(V))
	@$(MAKE) build-release VERSION=$(V)
	$(call release_summary,$(V))

release-patch: prerelease clean lint test ## Patch bump release
	$(eval V := $(shell svu patch))
	$(call do_tag,$(V))
	@$(MAKE) build-release VERSION=$(V)
	$(call release_summary,$(V))
