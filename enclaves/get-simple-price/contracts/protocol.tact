// Message structure for requesting price from the oracle.
message (0xc99be573) OraclePriceRequest {
    requestID: Int as uint32;    // Identifier for the request (arbitrary, chosen by the dApp developer).
    ticker: Int as uint32;        // Asset to request.
    minUpdatedAt: Int as uint32;  // Minimum timestamp for the last update of price data.
}

// Message structure for receiving price response from the oracle.
message (0x9735a9c2) OraclePriceResponse {
    requestID: Int as uint32;     // Identifier, as received from the price request.
    ticker: Int as uint32;        // Asset.
    usd: Int as uint32;           // Price of the asset in USD.
    lastUpdatedAt: Int as uint32; // Timestamp of the last update of price data.
}

// Define message structure for receiving scheduled response from the oracle.
message (0x00f8bc66) OraclePriceScheduledResponse {
    requestID: Int as uint32;     // Identifier matching the corresponding request.
    ticker: Int as uint32;         // Ticker symbol for the asset.
    lastUpdatedAt: Int as uint32; // Timestamp of the last update of price data.
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
    fun requestPriceFromOracle() {
        send(SendParameters{
            to: self.priceOracleAddress,
            body: OraclePriceRequest{
                requestID: 1,
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
    receive(msg: OraclePriceResponse) {
        let sb: StringBuilder = beginString();
        sb.append("TON price received: ");
        sb.append(msg.usd.toString());
        self.reply(sb.toString().asComment());
    }

    // Handle scheduled response from the oracle indicating outdated price data.
    // This should be overridden by implementing contracts.
    receive(msg: OraclePriceScheduledResponse) {
        let sb: StringBuilder = beginString();
        sb.append("TON price is outdated");
        self.reply(sb.toString().asComment());
    }
}