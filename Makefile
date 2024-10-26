GOMODCACHE ?= $(HOME)/go/pkg/mod
GOCACHE ?= $(HOME)/.cache/go-build

GO_DOCKER_VERSION := 1.23.2
GOLANGCI_LINT_VERSION := v1.61.0

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
	@echo "Starting linting."
	@$(DOCKER_RUN) golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run --out-format colored-line-number
	@echo "Linting finished."

.PHONY: test
test:
	@echo "Running unit tests with coverage report."
	@$(CGO_DOCKER_RUN) test ./pkg/... -coverprofile=.unit_tests_coverage.out -count=1 -race=1
	@$(GO_DOCKER_RUN) tool cover -html=.unit_tests_coverage.out -o .unit_tests_coverage.html
	@echo "Test suites completed."

.PHONY: open_coverage_report
open_coverage_report:
	@echo "Opening the unit tests coverage report in a web browser."
	@open .unit_tests_coverage.html
