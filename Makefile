####################################################################################################
# Variables ########################################################################################
####################################################################################################

GOCMD=CGO_ENABLED=0 go

####################################################################################################
# Lint #############################################################################################
####################################################################################################

.PHONY: lint_go
lint_go:
	@docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run --out-format colored-line-number

.PHONY: lint
lint: lint_go

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

.PHONY: minikube_start_cluster
minikube_start_cluster: minikube_check_version minikube_delete_cluster
	@./scripts/minikube_start_cluster.zsh

.PHONY: minikube_status
minikube_status: minikube_check_version
	@./scripts/minikube_status.zsh
