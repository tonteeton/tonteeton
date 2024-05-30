SRC := $(wildcard *.go) $(wildcard */*.go)

all: format test audit enclave

.PHONY: test
test:
	 ego-go test -v ./... -coverprofile=coverage.out

.PHONY: test
test-integration:
	 ego-go test -v ./... -coverprofile=coverage.out --tags=integration

.PHONY: tidy
format:
	ego-go fmt ./...
	ego-go mod tidy -v

.PHONY: audit
audit:
	ego-go mod verify
	ego-go vet ./...

enclave: proto $(SRC) go.mod go.sum cacert.pem cacert.pem.sha256 enclave.json
	sha256sum -c cacert.pem.sha256
	ego-go build

proto: ereport/private_keys_report.pb.go

ereport/private_keys_report.pb.go: ereport/private_keys_report.proto
	protoc --go_out=paths=source_relative:./ -I. ereport/private_keys_report.proto

cacert.pem: cacert.pem.sha256
	# https://curl.se/docs/caextract.html
	wget https://curl.se/ca/cacert.pem

cacert.pem.sha256:
	wget https://curl.se/ca/cacert.pem.sha256

.PHONY: clean
clean:
	go clean
	rm -f coverage.out
	rm -f cacert.pem.sha256 cacert.pem
	rm -f mount/*
	rm -f public.pem private.pem
