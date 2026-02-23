.PHONY: test lint fix fmt vet

test:
	go test ./...

lint: vet
	golangci-lint run ./...

fix:
	go fix ./...

fmt:
	gofmt -w .

vet:
	go vet ./...
