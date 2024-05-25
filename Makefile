####################################################################################################
# Variables ########################################################################################
####################################################################################################

GOCMD=CGO_ENABLED=0 go
HELM_ENV?=local

####################################################################################################
# Lint #############################################################################################
####################################################################################################

.PHONY: lint_go
lint_go:
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run --out-format colored-line-number

.PHONY: lint_helm
lint_helm: helm_lint_charts

.PHONY: lint
lint: lint_go lint_helm

####################################################################################################
# Tests ############################################################################################
####################################################################################################

.PHONY: clean_unit_tests_cache
clean_unit_tests_cache:
	$(GOCMD) clean -testcache

.PHONY: unit_tests
unit_tests: clean_unit_tests_cache
	$(GOCMD) test ./pkg/... -coverprofile=.unit_tests_coverage.out

.PHONY: unit_tests_coverage
unit_tests_coverage:
	$(GOCMD) tool cover -html=.unit_tests_coverage.out

.PHONY: test
test: unit_tests

####################################################################################################
# Minikube #########################################################################################
####################################################################################################

.PHONY: minikube_delete_cluster
minikube_delete_cluster:
	$(GOCMD) run ./cmd/minikube/main.go delete

.PHONY: minikube_start_cluster
minikube_start_cluster:
	$(GOCMD) run ./cmd/minikube/main.go start

####################################################################################################
# Helm #############################################################################################
####################################################################################################

.PHONY: helm_lint_charts
helm_lint_charts:
	$(GOCMD) run ./cmd/helm/main.go lint

.PHONY: helm_install_charts
helm_install_charts:
	$(GOCMD) run ./cmd/helm/main.go install $(HELM_ENV)
