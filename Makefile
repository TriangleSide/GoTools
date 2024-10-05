####################################################################################################
# Docker ###########################################################################################
####################################################################################################

DOCKER_CLIENT_VERSION := $(shell docker version --format '{{.Client.Version}}')
DOCKER_RUN := docker run --rm -v $(PWD):/app -w /app  --network host

####################################################################################################
# GoLang ###########################################################################################
####################################################################################################

GOMODCACHE ?= $(HOME)/go/pkg/mod
GOCACHE ?= $(HOME)/.cache/go-build

GO_DOCKER_VERSION := 1.23.1
GO_DOCKER_CACHES := -v $(GOMODCACHE):/go/pkg/mod -e GOMODCACHE=/go/pkg/mod -v $(GOCACHE):/root/.cache/go-build -e GOCACHE=/root/.cache/go-build
GO_DOCKER_RUN := $(DOCKER_RUN) $(GO_DOCKER_CACHES) -e CGO_ENABLED=0 golang:$(GO_DOCKER_VERSION) go
GO_CGO_DOCKER_RUN := $(DOCKER_RUN) $(GO_DOCKER_CACHES) -e CGO_ENABLED=1 golang:$(GO_DOCKER_VERSION) go

####################################################################################################
# Clean ############################################################################################
####################################################################################################

.PHONY: clean
clean: test_clean
	@echo "Clean finished."

####################################################################################################
# Lint #############################################################################################
####################################################################################################

GOLANGCI_LINT_VERSION := v1.59

.PHONY: lint_go
lint_go:
	@echo "Linting go files."
	@$(DOCKER_RUN) golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run --out-format colored-line-number

.PHONY: lint
lint: lint_go
	@echo "Linting finished."

####################################################################################################
# Tests ############################################################################################
####################################################################################################

.PHONY: test_clean
test_clean:
	@echo "Cleaning test cache."
	@$(GO_DOCKER_RUN) clean -testcache

.PHONY: test_unit
test_unit:
	@echo "Running unit tests with coverage report."
	@$(GO_CGO_DOCKER_RUN) test ./pkg/... -coverprofile=.unit_tests_coverage.out -count=1 -race=1
	@$(GO_DOCKER_RUN) tool cover -html=.unit_tests_coverage.out -o .unit_tests_coverage.html

.PHONY: test_unit_coverage_report
test_unit_coverage_report:
	@echo "Opening the unit tests coverage report in a web browser."
	@open .unit_tests_coverage.html

.PHONY: test
test: test_unit
	@echo "Test suites completed."
