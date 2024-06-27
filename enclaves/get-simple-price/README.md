# get-simple-price enclave

## Introduction

TonTeeTon enclave app to get the TON prices from CoinGecko,
providing TON contracts with up-to-date information on market prices.

## Contracts

- [./contracts](./contracts): TON contracts directory.

## Directories (Go packages)

- [appconf](./appconf): Application configuration management.
- [coingecko](./coingecko): A client for interacting with the CoinGecko API to fetch cryptocurrency price data.
- [coinconv](./coinconv): Conversion from CoinGecko format to enclave response format.
- [priceresp](./priceresp): Prepare price enclave TON-compatible response.

## Local build (build and check the enclave ID)

To build and check the enclave ID:

1. **Clone the repository** and navigate to the verifier directory:

    ```sh
    git clone https://github.com/tonteeton/tonteeton.git
    cd enclaves/get-simple-price/
    ```

2. **Build the Docker image** using the following command:

    ```sh
    make docker-build
    ```

3. **Run the command**

    ```sh
    docker run --rm t3-get-simple-price ego uniqueid ./enclave
    ```

## Reports

Reports for built enclave versions are available at [GitHub Actions](https://github.com/tonteeton/tonteeton/actions/workflows/get-simple-price.yml).
