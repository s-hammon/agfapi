PROJECT_NAME := agfapi
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

build: test clean
	@GOOS=${GOOS} GOARCH=${GOARCH} go build -o bin/${PROJECT_NAME} cmd/${PROJECT_NAME}/main.go

test:
	@go vet ./...
	@go test -cover ./...

clean:
	@rm -rf bin
	@go mod tidy

gosec:
	@gosec -terse -exclude=G104 ./...

lint:
	@golangci-lint run --disable=errcheck --timeout=2m

ready: test lint gosec

test-packages:
	go test -json $$(go list ./... | grep -v -e /bin -e /cmd -e /vendor -e /internal/api/models) |\
		tparse --follow -sort=elapsed -trimpath=auto -all

.PHONY: clean test build gosec lint ready
