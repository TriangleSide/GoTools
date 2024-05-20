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
	$(GOCMD) test ./pkg/...

.PHONY: test
test: unit_tests

####################################################################################################
# Minikube #########################################################################################
####################################################################################################

.PHONY: minikube_check_version
minikube_check_version:
	@./scripts/minikube_check_version.zsh

.PHONY: minikube_delete_cluster
minikube_delete_cluster: minikube_check_version
	@./scripts/minikube_delete_cluster.zsh

.PHONY: minikube_create_cluster
minikube_create_cluster: minikube_check_version minikube_delete_cluster
	@./scripts/minikube_create_cluster.zsh

####################################################################################################
# Helm #############################################################################################
####################################################################################################

.PHONY: helm_check_version
helm_check_version:
	@./scripts/helm_check_version.zsh

.PHONY: helm_lint_charts
helm_lint_charts: helm_check_version
	@./scripts/helm_lint_charts.zsh

.PHONY: helm_install_charts
helm_install_charts: helm_check_version helm_lint_charts
	@./scripts/helm_install_charts.zsh $(HELM_ENV)
