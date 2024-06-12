SRC := $(wildcard *.go) $(wildcard */*.go)

all: format main

.PHONY: tidy
format:
	go fmt ./...
	go mod tidy -v


main:	main.go
	go build


.PHONY: clean
clean:
	go clean
