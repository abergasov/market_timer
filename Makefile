PROJECT_NAME:=market-timer
FILE_HASH := $(shell git rev-parse HEAD)
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

install-lint: ## Installs golangci-lint tool which a go linter
ifndef GOLANGCI_LINT
	${info golangci-lint not found, installing golangci-lint@latest}
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif

gogen: ## generate code
	${info generate code...}
	go generate ./internal...

test: ## Runs tests
	${info Running tests...}
	go test -v -race ./... -cover -coverprofile cover.out
	go tool cover -func cover.out | grep total

bench: ## Runs benchmarks
	${info Running benchmarks...}
	go test -bench=. -benchmem ./... -run=^#

lint: install-lint ## Runs linters
	@echo "-- linter running"
	golangci-lint run -c .golangci.yaml ./internal...
	golangci-lint run -c .golangci.yaml ./cmd...

stop: ## Stops the local environment
	${info Stopping containers...}
	docker container ls -q --filter name=${PROJECT_NAME} ; true
	${info Dropping containers...}
	docker rm -f -v $(shell docker container ls -q --filter name=${PROJECT_NAME}) ; true

build: ## Build binary
	${info Building binary...}
	go build -ldflags="-X 'main.dbPath=/var/lib/marketimer/storage.db'" -o ./bin/marketimer ./cmd


run: ## Runs binary local with environment in docker
	${info Run app containered}
	GIT_HASH=${FILE_HASH} docker compose -p ${PROJECT_NAME} up --build -d

logs: ## Shows logs
	${info Showing logs...}
	docker logs -f marketimer


.PHONY: help install-lint test gogen lint stop dev_up build run logs
.DEFAULT_GOAL := help