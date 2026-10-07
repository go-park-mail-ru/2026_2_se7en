COVERAGE_FILE ?= coverage.out

.PHONY: coverage

coverage:
	go test -coverpkg=./... -coverprofile=$(COVERAGE_FILE) . ./handlers
	go tool cover -func=$(COVERAGE_FILE) 

