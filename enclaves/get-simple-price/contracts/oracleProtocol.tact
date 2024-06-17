// Message structure for requesting price from the oracle.
message (0xc99be573) OraclePriceRequest {
    queryId: Int as uint64;       // Identifier for the request (arbitrary, greater then 0, chosen by the dApp developer).
    customPayload: Cell?;         // Optional payload to forward with the request, returned in the response.
    ticker: Int as uint64;        // Asset to request.
    minUpdatedAt: Int as uint64;  // Minimum timestamp for the last update of price data.
}

// Message structure for receiving price response from the oracle.
message (0x9735a9c2) OraclePriceResponse {
    queryId: Int as uint64;       // Identifier, as received from the price request.
    payload: Cell?;               // Optional payload from the request.
    lastUpdatedAt: Int as uint64; // Timestamp of the last update of price data.
    ticker: Int as uint64;        // Asset.
    usd: Int as uint64;           // Price of the asset in USD (2 decimal places precision, cents).
    usd24vol: Int as uint64;      // 24-hour volume in USD (cents).
    usd24change: Int as int64;    // 24-hour change in USD (cents).
    btc: Int as uint64;           // Price of the asset in BTC (8 decimal places precision, satoshi).
}

// Message structure for receiving scheduled response from the oracle.
message (0x00f8bc66) OraclePriceScheduledResponse {
    queryId: Int as uint64;       // Identifier matching the corresponding request.
    payload: Cell?;               // Optional payload from the request.
    ticker: Int as uint64;        // Ticker symbol for the asset.
    lastUpdatedAt: Int as uint64; // Timestamp of the last update of price data.
}

// Message structure for receiving new address response from the oracle.
// Sended when oracle moves to a new contract address.
message (0x9899519b) OracleNewAddressResponse {
    newAddress: Address;          // New address of the oracle.
    queryId: Int as uint64;       // Identifier of the original request.
}

// Trait for entities using ticker symbols.
trait UsesTickers {
    const TICKER_TON : Int = 0x72716023;
}

// Trait for entities interacting with the price oracle.
trait  UsesPriceOracle with UsesTickers {
    // Address of the price oracle should be definded to use this trait.
    priceOracleAddress: Address;

    receive (msg: OraclePriceResponse) {
        self.requireOracle();
        self.receivedOraclePriceResponse(msg);
    }

    receive(msg: OraclePriceScheduledResponse) {
        self.requireOracle();
        self.receivedOraclePriceScheduledResponse(msg);
    }

    receive(msg: OracleNewAddressResponse) {
        self.requireOracle();
        self.receivedOracleNewAddressResponse(msg);
    }

    get fun priceOracleAddress() : Address {
        return self.priceOracleAddress;
    }

    // Ensure the sender is the expected oracle contract.
    fun requireOracle() {
        require(sender() == self.priceOracleAddress, "Unknown oracle");
    }

    // Function to request price data from the oracle.
    // This should be overridden by implementing contracts.
    virtual fun requestPriceFromOracle() {
        send(SendParameters{
            to: self.priceOracleAddress,
            body: OraclePriceRequest{
                queryId: 1,
                customPayload: null,
                ticker: self.TICKER_TON,
                minUpdatedAt: now() - 86400, // Price must be no older than 24 hours.
            }.toCell(),
            value: 0,
            mode: SendRemainingValue,
            bounce: false
        });
    }

    // Handle response from the oracle containing price data.
    // This should be overridden by implementing contracts.
    virtual fun receivedOraclePriceResponse(msg: OraclePriceResponse) {
        let sb: StringBuilder = beginString();
        sb.append("TON price received: ");
        sb.append(msg.usd.toString());
        emit(sb.toString().asComment());
    }

    // Handle scheduled response from the oracle indicating outdated price data.
    // This should be overridden by implementing contracts.
    virtual fun receivedOraclePriceScheduledResponse(msg: OraclePriceScheduledResponse) {
        emit("TON price is outdated".asComment());
    }

    // Handle response from the oracle indicating a new address.
    // This should be overridden by implementing contracts.
    // The contract should decide whether it is safe to trust the new address.
    virtual fun receivedOracleNewAddressResponse(msg: OracleNewAddressResponse) {
        let sb: StringBuilder = beginString();
        sb.append("Oracle contract moved to: ");
        sb.append(msg.newAddress.toString());
        emit(sb.toString().asComment());
    }

}
