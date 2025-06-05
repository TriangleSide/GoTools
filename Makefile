GOMODCACHE ?= $(HOME)/go/pkg/mod
GOCACHE ?= $(HOME)/.cache/go-build

GO_DOCKER_VERSION := 1.23
GOLANGCI_LINT_VERSION := v1.64.8

DOCKER_RUN := docker run --rm -v $(PWD):/app -w /app --network host

GO_DOCKER_CACHES := -v $(GOMODCACHE):/go/pkg/mod -e GOMODCACHE=/go/pkg/mod -v $(GOCACHE):/root/.cache/go-build -e GOCACHE=/root/.cache/go-build
GO_DOCKER_RUN := $(DOCKER_RUN) $(GO_DOCKER_CACHES) -e CGO_ENABLED=0 golang:$(GO_DOCKER_VERSION) go
CGO_DOCKER_RUN := $(DOCKER_RUN) $(GO_DOCKER_CACHES) -e CGO_ENABLED=1 golang:$(GO_DOCKER_VERSION) go

.PHONY: clean
clean:
	@$(GO_DOCKER_RUN) clean -testcache
	@echo "Clean finished."

.PHONY: lint
lint:
	@echo "Starting linter."
	@$(DOCKER_RUN) golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run --out-format colored-line-number
	@echo "Linter finished."

.PHONY: test
test:
	@echo "Running unit tests with coverage report."
	@$(CGO_DOCKER_RUN) test ./pkg/... -coverprofile=test_coverage.out -count=1 -race=1
	@$(GO_DOCKER_RUN) tool cover -html=test_coverage.out -o test_coverage.html
	@echo "Test suites completed."
