SRC := $(wildcard *.go) $(wildcard */*.go)

all: format test audit enclave

.PHONY: test
test:
	 go test -v ./... -coverprofile=coverage.out

.PHONY: test
test-integration:
	 go test -v ./... -coverprofile=coverage.out --tags=integration

.PHONY: tidy
format:
	go fmt ./...
	go mod tidy -v

.PHONY: audit
audit:
	go mod verify
	go vet ./...

enclave: $(SRC) go.mod go.sum
	go build

.PHONY: clean
clean:
	go clean
