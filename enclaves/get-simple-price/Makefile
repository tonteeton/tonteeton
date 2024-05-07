SRC := $(wildcard *.go)

all: format test audit enclave

.PHONY: test
test:
	 go test ./... -coverprofile=coverage.out

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
