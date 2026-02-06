.PHONY: all build test test-coverage lint fmt clean run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
BINARY_NAME=ownu

all: lint test build

build:
	CGO_ENABLED=0 $(GOBUILD) -o $(BINARY_NAME) ./cmd/server

run:
	$(GOCMD) run ./cmd/server

test:
	$(GOTEST) -v -race ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short:
	$(GOTEST) -short ./...

lint:
	$(GOVET) ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

fmt:
	$(GOFMT) ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Docker targets
docker-build:
	docker build -t ownu-backend .

docker-test:
	docker run --rm ownu-backend go test -v ./...
