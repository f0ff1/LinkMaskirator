.PHONY: lint test check clean help

# Variables
GOLANGCI_LINT = $(shell which golangci-lint)
GO_TEST = go test

# Colored output (works in Git Bash!)
RED = $(shell tput setaf 1)
GREEN = $(shell tput setaf 2)
YELLOW = $(shell tput setaf 3)
RESET = $(shell tput sgr0)

help:
	@echo "Available commands:"
	@echo "  make lint   - run golangci-lint"
	@echo "  make test   - run unit tests"
	@echo "  make check  - run linters and tests"
	@echo "  make clean  - clean cache"

lint:
	@echo "$(YELLOW)Running golangci-lint...$(RESET)"
	@if [ -z "$(GOLANGCI_LINT)" ]; then \
		echo "$(RED)golangci-lint not found. Installing...$(RESET)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...
	@echo "$(GREEN)[OK] Linter completed successfully$(RESET)"

test:
	@echo "$(YELLOW)Running unit tests...$(RESET)"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)[OK] Tests completed successfully$(RESET)"
	@go tool cover -func=coverage.out | grep total

check: lint test
	@echo "$(GREEN)[OK] All checks passed successfully$(RESET)"

clean:
	@echo "$(YELLOW)Cleaning temporary files...$(RESET)"
	go clean -testcache
	go clean -cache
	rm -f coverage.out
	@echo "$(GREEN)[OK] Clean completed successfully$(RESET)"