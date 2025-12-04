GO_CONTAINER_VERSION := $(shell grep '^go ' go.mod | awk '{print $$2}')
GOLANGCI_LINT_VERSION := v2.6.1

CONTAINER_RUN := podman run --rm -v $(PWD):/app -w /app --network host

GO_CONTAINER_RUN := $(CONTAINER_RUN) -e CGO_ENABLED=0 golang:$(GO_CONTAINER_VERSION) go
CGO_CONTAINER_RUN := $(CONTAINER_RUN) -e CGO_ENABLED=1 golang:$(GO_CONTAINER_VERSION) go

.PHONY: clean
clean:
	@$(GO_CONTAINER_RUN) clean -testcache
	@echo "Clean finished."

.PHONY: lint
lint:
	@echo "Starting linter."
	@$(CONTAINER_RUN) golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run
	@echo "Linter finished."

.PHONY: test
test:
	@echo "Running unit tests with coverage report."
	@$(CGO_CONTAINER_RUN) test ./pkg/... -coverprofile=test_coverage.out -count=1 -race=1
	@$(GO_CONTAINER_RUN) tool cover -html=test_coverage.out -o test_coverage.html
	@echo "Test suites completed."
