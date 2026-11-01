PKG ?= std
ADDR ?= 127.0.0.1:8080
N ?= 10

.DEFAULT_GOAL := help

.PHONY: help build install serve search test race eval fmt vet check clean

help: ## Show available targets
	@grep -E '^[a-z]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  %-10s %s\n", $$1, $$2}'

build: ## Build bin/godex
	go build -o bin/godex ./cmd/godex

install: ## Install godex into GOBIN
	go install ./cmd/godex

serve: ## Start the web page (PKG, ADDR)
	go run ./cmd/godex serve -p '$(PKG)' -a '$(ADDR)'

search: ## Search from the terminal (Q, PKG, N)
	@test -n '$(Q)' || { echo 'usage: make search Q="close idle connections"' >&2; exit 2; }
	go run ./cmd/godex search -p '$(PKG)' -n '$(N)' -- '$(Q)'

test: ## Run tests
	go test ./...

race: ## Run tests with the race detector
	go test -race ./...

eval: ## Show search quality per query
	go test -count=1 -v -run TestRelevance ./internal/eval/

clean: ## Remove build output
	rm -rf bin
