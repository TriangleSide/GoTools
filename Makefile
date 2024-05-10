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