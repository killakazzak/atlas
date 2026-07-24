.PHONY: fmt vet lint test check run

fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

check: fmt vet lint test

run:
	go run ./cmd/server
