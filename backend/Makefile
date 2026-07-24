.PHONY: fmt vet lint test openapi check run

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -timeout 60s ./...

openapi:
	vacuum lint --details openapi/openapi.yaml

check: fmt vet lint test openapi

run:
	go run ./cmd/server
