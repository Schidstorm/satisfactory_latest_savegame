
GOLANGCI_LINT_VERSION ?= v2.12.1

.PHONY: test test-cover build build-debug debug fmt vet lint lint-install

check: test fmt vet lint

test:
	go test -v ./...

test-cover:
	go test -v -cover -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html

build:
	mkdir -p build && \
	CGO_ENABLED=1 GO111MODULE=on go build -o build/satisfactory_latest_savegame .

build-debug:
	mkdir -p build && \
	CGO_ENABLED=1 GO111MODULE=on go build -tags codes -gcflags="all=-N -l" -o build/satisfactory_latest_savegame .

debug: build-debug
	./build/satisfactory_latest_savegame

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run --timeout=5m ./...

lint-install:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

