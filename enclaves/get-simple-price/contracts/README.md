# get-simple-price contract

## Motivation

To provide smart contracts with up-to-date TON market prices.


## Summary

Prices are obtained from [CoinGecko](https://docs.coingecko.com/reference/simple-price).

The enclave contract operates as follows:

- **Client Interaction**: Client contracts interact with the enclave using the [`OraclePriceRequest`](#oraclepricerequest) message to request specific asset prices (currently only TON is supported, ticker `0x72716023`).

- **Response Mechanism**:
  - **[`OraclePriceResponse`](#oraclepriceresponse)**: If the enclave has up-to-date data for the requested asset, it responds with the corresponding price details.
  - **[`OraclePriceScheduledResponse`](#oraclepricescheduledresponse)**: If the enclave doesn't have up-to-date data but acknowledges the request, it responds with details about the scheduled response. Clients can repeat the request later to obtain updated information.
  - **[`OracleNewAddressResponse`](#oraclenewaddressresponse)**: When the enclave updates and moves to a new contract address, it responds with the updated address details.

```{admonition} Diagram
:class: dropdown
![Diagram](diagrams/get_simple_price.svg)
```

## URLs

* Source code: https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price
  * Tact messages and trait to use: [oracleProtocol.tact](https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price/contracts/oracleProtocol.tact)
  - Main contract: [contract.tact](https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price/contracts/contract.tact)
  * Tact usage examples:
    * [demo.tact](https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price/contracts/demo.tact)
    * [TONtastic! contract](https://github.com/tonteeton/tontastic/blob/main/contracts/contract.tact)
  * Contract published containers: [container/get-simple-price-contracts](https://github.com/tonteeton/tonteeton/pkgs/container/get-simple-price-contracts)
  * Enclave published containers: [container/get-simple-price](https://github.com/tonteeton/tonteeton/pkgs/container/get-simple-price)
* Current address in TON mainnet: [EQCEcsyKQhGK2ofX4FIVpOTPd9DcC2HqzFUIpYrXo5kOyo61](https://tonviewer.com/EQCEcsyKQhGK2ofX4FIVpOTPd9DcC2HqzFUIpYrXo5kOyo61)
* Contract dashboard with stats is available as a Telegram mini-app: [@TonTeeTonBot/get_simple_price](https://t.me/TonTeeTonBot/get_simple_price).

## Contract client protocol messages

### OraclePriceRequest

This message structure is used for requesting price from the oracle.

| Field         | Type           | Description |
|---------------|----------------|-------------|
| `queryId`     | `Int as uint64` | Identifier for the request (chosen by the dApp developer). |
| `customPayload` | `Cell?`       | Optional payload to forward with the request, returned in the response. |
| `ticker`      | `Int as uint64` | Asset to request, value from the UsesTickers trait. |
| `minUpdatedAt`| `Int as uint64` | Minimum timestamp for the last update of price data. |


**TLB**

```text
oracle_price_request#c99be573 queryId:uint64 customPayload:Maybe ^cell
                              ticker:uint64 minUpdatedAt:uint64
= OraclePriceRequest
```

**Signature**

```text
OraclePriceRequest{queryId:uint64,customPayload:Maybe ^cell,
                   ticker:uint64,minUpdatedAt:uint64}
```

### OraclePriceResponse

This message structure is used for receiving price response from the oracle.

| Field            | Type           | Description |
|------------------|----------------|-------------|
| `queryId`        | `Int as uint64` | Identifier, as received from the price request. |
| `payload`        | `Cell?`        | Optional payload from the request. |
| `lastUpdatedAt`  | `Int as uint64` | Timestamp of the last update of price data. |
| `ticker`         | `Int as uint64` | Asset, value from the UsesTickers trait. |
| `usd`            | `Int as uint64` | Price of the asset in USD (2 decimal places precision, cents). |
| `usd24vol`       | `Int as uint64` | 24-hour volume in USD (cents). |
| `usd24change`    | `Int as int64` | 24-hour change relative to USD (2 decimal places precision, percent with sign). |
| `btc`            | `Int as uint64` | Price of the asset in BTC (8 decimal places precision, satoshi). |


**TLB**

```text
oracle_price_response#9735a9c2 queryId:uint64 payload:Maybe ^cell lastUpdatedAt:uint64
                               ticker:uint64 usd:uint64 usd24vol:uint64
                               usd24change:int64 btc:uint64
= OraclePriceResponse
```


**Signature**

```text
OraclePriceResponse{queryId:uint64,payload:Maybe ^cell,
                    lastUpdatedAt:uint64,ticker:uint64,
                    usd:uint64,usd24vol:uint64,
                    usd24change:int64,btc:uint64}
```

### OraclePriceScheduledResponse

This message structure is used for receiving scheduled response from the oracle.

| Field            | Type           | Description |
|------------------|----------------|-------------|
| `queryId`        | `Int as uint64` | Identifier matching the corresponding request. |
| `payload`        | `Cell?`        | Optional payload from the request. |
| `ticker`         | `Int as uint64` | Ticker symbol for the asset. |
| `lastUpdatedAt`  | `Int as uint64` | Timestamp of the last update of price data. |

**TLB**

```text
oracle_price_scheduled_response#00f8bc66 queryId:uint64 payload:Maybe ^cell
                                         ticker:uint64
                                         lastUpdatedAt:uint64
= OraclePriceScheduledResponse
```

**Signature**

```text
OraclePriceScheduledResponse{queryId:uint64,payload:Maybe ^cell,
                             ticker:uint64,lastUpdatedAt:uint64}
```

### OracleNewAddressResponse

This message structure is used for receiving new address response from the oracle when it moves to a new contract address.

| Field         | Type     | Description |
|---------------|----------|-------------|
| `newAddress`  | `Address`| New address of the oracle. |
| `queryId`     | `Int as uint64` | Identifier of the original request. |


**TLB**

```text
oracle_new_address_response#9899519b newAddress:address queryId:uint64
= OracleNewAddressResponse
```


**Signature**

```text
OracleNewAddressResponse{newAddress:address,queryId:uint64}
```

## Error Codes

| Code  | Description                                      |
|-------|--------------------------------------------------|
| 2     | Stack underflow                                  |
| 3     | Stack overflow                                   |
| 4     | Integer overflow                                 |
| 5     | Integer out of expected range                    |
| 6     | Invalid opcode                                   |
| 7     | Type check error                                 |
| 8     | Cell overflow                                    |
| 9     | Cell underflow                                   |
| 10    | Dictionary error                                 |
| 13    | Out of gas error                                 |
| 32    | Method ID not found                              |
| 34    | Action is invalid or not supported               |
| 37    | Not enough TON                                   |
| 38    | Not enough extra-currencies                      |
| 128   | Null reference exception                         |
| 129   | Invalid serialization prefix                     |
| 130   | Invalid incoming message                         |
| 131   | Constraints error                                |
| 132   | Access denied                                    |
| 133   | Contract stopped                                 |
| 134   | Invalid argument                                 |
| 135   | Code of a contract was not found                 |
| 136   | Invalid address                                  |
| 137   | Masterchain support is not enabled for this contract |
| 2138  | Price update is outdated                         |
| 9925  | Price update is from the future                  |
| 10613 | Already moved                                    |
| 17444 | Previous address required                        |
| 29990 | Address is not new                               |
| 39366 | Moved to new address                             |
| 39697 | Unknown ticker                                   |
| 40104 | Unknown caller                                   |
| 40368 | Contract stopped                                 |
| 45226 | Unknown oracle                                   |
| 48401 | Invalid signature                                |
| 53296 | Contract not stopped                             |
| 54415 | New address sender expected                      |
| 54462 | Unknown Demo                                     |
