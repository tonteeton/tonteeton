# TonTeeTon Contract Verifier

## Introduction

TonTeeTon Contract Verifier is a utility used to verify TonTeeTon contract attestation reports via the Azure Provisioning Certificate Caching Service (PCCS). The verifier does not require an SGX-enabled machine to run.


## Usage

**Run the command** with the `contractAddress` and the `expectedMeasurement`:

    docker run --rm ghcr.io/tonteeton/verifier <contractAddress> <expectedMeasurement>

Replace `<contractAddress>` with the address of the contract and `<expectedMeasurement>` with the expected enclave measurement value. For example:

    docker run --rm ghcr.io/tonteeton/verifier kQCgJ6O6lB1UtV1I86NvNknbLDWBT-05zCuikGkk3LFPg4Oy ef6d2adf7f08c3ea88305d7c9c73ad9837c60227db2378de8fc7d5e619637134

Upon successful verification, you should see the message:

    ✓ Contract attestation report is verified

## Development Build

1. **Clone the repository** and navigate to the verifier directory:

    ```sh
    git clone https://github.com/tonteeton/tonteeton.git
    cd tonteeton/verifier/
    ```

2. **Build the Docker image** using the following command:

    docker build -t t3-verifier .

3. **Run the command** with the `contractAddress` and the `expectedMeasurement`:

    `docker run --rm t3-verifier <contractAddress> <expectedMeasurement>`


## Reports

Reports for known contracts are available at [GitHub Actions](https://github.com/tonteeton/tonteeton/actions/workflows/verify-contract.yml).
