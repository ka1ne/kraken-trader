.PHONY: test test-integration test-all

test:
	go test -v -short ./...

test-integration:
	go test -v -run "^TestIntegration" ./...

test-all: test test-integration

.PHONY: test-coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out 