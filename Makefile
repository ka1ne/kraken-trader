GOTESTSUM_FORMAT ?= pkgname-and-test-fails

.PHONY: test test-integration test-all

test:
	go test -v -short -timeout 10s ./...

test-integration:
	go test -v -run "^TestIntegration" -timeout 30s ./...

test-all: test test-integration

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: install-tools
install-tools:
	go install gotest.tools/gotestsum@latest