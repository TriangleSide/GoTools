####################################################################################################
# Variables ########################################################################################
####################################################################################################

CGO_ENABLED=0
GOCMD=CGO_ENABLED=$(CGO_ENABLED) go
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean

####################################################################################################
# Tests ############################################################################################
####################################################################################################

.PHONY: clean-unit-tests-cache
clean-unit-tests-cache:
	$(GOCLEAN) -testcache

.PHONY: unit-tests
unit-tests: clean-unit-tests-cache
	$(GOTEST) ./...

.PHONY: test
test: unit-tests

####################################################################################################
# Minikube #########################################################################################
####################################################################################################

.PHONY: minikube-check-version
minikube-check-version:
	@command ./scripts/minikube_check_version.zsh

.PHONY: minikube-delete-cluster
minikube-delete-cluster: minikube-check-version
	@command ./scripts/minikube_delete_cluster.zsh

.PHONY: minikube-start-cluster
minikube-start-cluster: minikube-check-version minikube-delete-cluster
	@command ./scripts/minikube_start_cluster.zsh

.PHONY: minikube-status
minikube-status: minikube-check-version
	@command ./scripts/minikube_status.zsh
