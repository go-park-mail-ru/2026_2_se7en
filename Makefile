COVERAGE_FILE ?= coverage.out

.PHONY: coverage swag

coverage:
	go test -coverpkg=./... -coverprofile=$(COVERAGE_FILE) . ./handlers
	go tool cover -func=$(COVERAGE_FILE) 

generate_swagger:
	swag init -g cmd/server/main.go --v3.1 -o ./docs
	