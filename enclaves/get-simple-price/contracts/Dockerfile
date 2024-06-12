FROM node:lts

WORKDIR /app

RUN git clone https://github.com/tact-lang/tact-template.git . && \
    git checkout 1cd1673fd5c02ec52e1d227fd1e170316cf94cf4

RUN yarn install
RUN yarn add @ton/ton @ton/sandbox @ton/test-utils eslint
RUN npm install ton-crypto

COPY . /app/sources
COPY ./tact.config.json /app/tact.config.json

RUN yarn build && yarn test

RUN chmod +x /app/sources/scripts/entrypoint.sh

ENTRYPOINT ["/app/sources/scripts/entrypoint.sh"]

LABEL org.opencontainers.image.title="get-simple-price-contracts"
LABEL org.opencontainers.image.description="TonTeeTon get-simple-price contracts"
LABEL org.opencontainers.image.url="https://github.com/tonteeton/tonteeton/pkgs/container/get-simple-price-contracts"
LABEL org.opencontainers.image.source="https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price/contracts"
