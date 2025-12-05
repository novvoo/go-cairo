# Makefile for go-cairo

.PHONY: all test cover clean

all: test

test:
	@echo "Running tests..."
	@go test ./...

cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out

clean:
	@echo "Cleaning up..."
	@rm -f coverage.out
	@rm -f go-cairo-updated.zip
