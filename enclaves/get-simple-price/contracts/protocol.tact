// Message structure for requesting price from the oracle.
message (0xc99be573) OraclePriceRequest {
    queryID: Int as uint64;       // Identifier for the request (arbitrary, chosen by the dApp developer).
    ticker: Int as uint64;        // Asset to request.
    minUpdatedAt: Int as uint64;  // Minimum timestamp for the last update of price data.
}

// Message structure for receiving price response from the oracle.
message (0x9735a9c2) OraclePriceResponse {
    queryID: Int as uint64;       // Identifier, as received from the price request.
    lastUpdatedAt: Int as uint64; // Timestamp of the last update of price data.
    ticker: Int as uint64;        // Asset.
    usd: Int as uint64;           // Price of the asset in USD (2 decimal places precision, cents).
    usd24vol: Int as uint64;      // 24-hour volume in USD (cents).
    usd24change: Int as int64;    // 24-hour change in USD (cents).
    btc: Int as uint64;           // Price of the asset in BTC (8 decimal places precision, satoshi).
}

// Define message structure for receiving scheduled response from the oracle.
message (0x00f8bc66) OraclePriceScheduledResponse {
    queryID: Int as uint64;       // Identifier matching the corresponding request.
    ticker: Int as uint64;        // Ticker symbol for the asset.
    lastUpdatedAt: Int as uint64; // Timestamp of the last update of price data.
}

// Trait for entities using ticker symbols.
trait UsesTickers {
    const TICKER_TON : Int = 0x72716023;
}

// Trait for entities interacting with the price oracle.
trait  UsesPriceOracle with UsesTickers {
    // Address of the price oracle should be definded to use this trait.
    priceOracleAddress: Address;

    // Function to request price data from the oracle.
    // This should be overridden by implementing contracts.
    virtual fun requestPriceFromOracle() {
        send(SendParameters{
            to: self.priceOracleAddress,
            body: OraclePriceRequest{
                queryID: 1,
                ticker: self.TICKER_TON,
                minUpdatedAt: now() - 86400, // Price must be no older than 24 hours.
            }.toCell(),
            value: ton("0.01"),
            mode: SendRemainingValue | SendIgnoreErrors,
            bounce: false
        });
    }

    // Handle response from the oracle containing price data.
    // This should be overridden by implementing contracts.
    virtual fun receivedOraclePriceResponse(msg: OraclePriceResponse) {
        throw(0xd21b4fb2);
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

    receive (msg: OraclePriceResponse) {
        self.receivedOraclePriceResponse(msg);
    }

    receive(msg: OraclePriceScheduledResponse) {
        self.receivedOraclePriceScheduledResponse(msg);
    }

    get fun priceOracleAddress() : Address {
        return self.priceOracleAddress;
    }

}
