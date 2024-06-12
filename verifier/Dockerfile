# Base image: https://github.com/edgelesssys/ego/pkgs/container/ego-dev/225924119?tag=v1.5.3
ARG EGO_VERSION=sha256:c2dc19831d230f26cdc8760fa08dea6eeea54f1c283b1029e2c0e3a0c465ac7e
FROM ghcr.io/edgelesssys/ego-dev@${EGO_VERSION}

WORKDIR /app

COPY . .

RUN make main

COPY sgx_default_qcnl.conf /etc/sgx_default_qcnl.conf

ENTRYPOINT ["./main"]

LABEL org.opencontainers.image.title="t3-verifier"
LABEL org.opencontainers.image.description="TonTeeTon contracts Verifier"
LABEL org.opencontainers.image.url="https://github.com/tonteeton/tonteeton/pkgs/container/verifier"
LABEL org.opencontainers.image.source="https://github.com/tonteeton/tonteeton/tree/main/verifier"
