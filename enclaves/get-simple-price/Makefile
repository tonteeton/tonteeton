
.PHONY: test
test:
	 go test ./... -coverprofile=coverage.out

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v


.PHONY: audit
audit:
	go mod verify
	go vet ./...

.PHONY: clean
clean:
	go clean
