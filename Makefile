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

.PHONY: clean test build
