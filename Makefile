COVERAGE_FILE ?= coverage.out
SWAG ?= $(shell go env GOPATH)/bin/swag

.PHONY: coverage generate_swagger

coverage:
	go test -coverpkg=./... -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

generate_swagger:
	$(SWAG) init -g cmd/server/main.go --v3.1 -o ./docs