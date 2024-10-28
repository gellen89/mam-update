.PHONY: help default \
        build build/linux \
				lint \
				vet \
        clean
default: help

help:
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_\/]+:.*##/ { printf "\033[36m%-30s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

vet: ## Runs go vet
	go vet ./...

build: vet ## Runs go vet and build the binary
	go build -ldflags "-s -w" -o bin/mam-update *.go

build/linux: vet ## Runs go vet and builds the binary specifically for Linux
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/mam-update *.go

clean: ## Cleans the bin directory
	rm -rf bin

lint: ## Runs the Go Linter
	golangci-lint run
