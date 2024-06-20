# Base image: https://github.com/edgelesssys/ego/pkgs/container/ego-dev/225924119?tag=v1.5.3
ARG EGO_VERSION=sha256:c2dc19831d230f26cdc8760fa08dea6eeea54f1c283b1029e2c0e3a0c465ac7e
FROM ghcr.io/edgelesssys/ego-dev@${EGO_VERSION} AS build

WORKDIR /app
ENV GOPATH=/app/go

COPY . .

# Obtain and verify the integrity of the CA certificate
RUN set -eux; \
    make clean cacert.pem; \
    echo "1794c1d4f7055b7d02c2170337b61b48a2ef6c90d77e95444fd2596f4cac609f  cacert.pem" | sha256sum -c -

# Build the enclave binary
RUN ego-go build

# Sign the enclave binary
RUN --mount=type=secret,id=signingkey,dst=private.pem,required=false ego sign enclave

# Build a single-executable bundle with the current EGo runtime
RUN ego bundle enclave

# Create clear environment for app deployment
FROM ghcr.io/edgelesssys/ego-dev@${EGO_VERSION} as deploy

WORKDIR /app

# Copy the built files from the build stage
COPY --from=build /app/enclave /app/enclave
COPY --from=build /app/enclave-bundle /app/enclave-bundle

COPY sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf
RUN mkdir mount

CMD ["ego", "run", "enclave", "get-price"]

LABEL org.opencontainers.image.title="get-simple-price"
LABEL org.opencontainers.image.description="TonTeeTon enclave app to get the TON prices from CoinGecko"
LABEL org.opencontainers.image.url="https://github.com/tonteeton/tonteeton/pkgs/container/get-simple-price"
LABEL org.opencontainers.image.source="https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price"
