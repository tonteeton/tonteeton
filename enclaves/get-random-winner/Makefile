SRC := $(wildcard *.go) $(wildcard */*.go)

all: format test audit enclave

.PHONY: test
test:
	 ego-go test -v ./... -coverprofile=coverage.out

.PHONY: test-integration
test-integration:
	 ego-go test -v ./... -coverprofile=coverage.out --tags=integration

.PHONY: format
format:
	ego-go fmt ./...
	ego-go mod tidy -v

.PHONY: audit
audit:
	ego-go mod verify
	ego-go vet ./...

enclave: $(SRC) go.mod go.sum buidls.json cacert.pem cacert.pem.sha256 enclave.json
	mkdir -p mount
	sha256sum -c cacert.pem.sha256
	ego-go build

.PHONY: docker-build
docker-build: private.pem
	docker buildx build --secret id=signingkey,src=$$PWD/private.pem -t t3-get-random-winner --no-cache .

.PHONY: docker-run
docker-run:
	docker run --rm -it \
	-v $$PWD/mount:/app/mount \
	--device /dev/sgx_enclave \
	--device /dev/sgx_provision \
	t3-get-random-winner \
	/bin/bash

private.pem:
	openssl genrsa -out private.pem -3 3072

cacert.pem: cacert.pem.sha256
	# https://curl.se/docs/caextract.html
	wget https://curl.se/ca/cacert.pem

cacert.pem.sha256:
	wget https://curl.se/ca/cacert.pem.sha256

all_buidls.json:
	wget -O $@ 'https://dorahacks.io/api/hackathon-buidls/361/?page=1&page_size=1000'

buidls.json: all_buidls.json
	jq -c '[.results[] | select(.demo_link != null) | {id, name}] | sort_by(.id)' $< >$@


.PHONY: clean
clean:
	go clean
	rm -f coverage.out
	rm -f cacert.pem.sha256 cacert.pem
	rm -f mount/*
	rm -f public.pem private.pem
