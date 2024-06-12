import "./oracleProtocol.tact";

contract DemoContract with UsesPriceOracle {

    priceOracleAddress : Address;

    init(priceOracleAddress: Address) {
        self.priceOracleAddress = priceOracleAddress;
    }

    get fun balance() : Int {
        return myBalance();
    }

    receive ("CallOracle") {
        self.requestPriceFromOracle();
    }

    receive (text: String) {
    }

    override fun receivedOraclePriceResponse(msg: OraclePriceResponse) {
        let sb: StringBuilder = beginString();
        sb.append("TON to USD: ");
        sb.append(msg.usd.toString());
        sb.append("\nTo BTC: ");
        sb.append(msg.btc.toString());
        emit(sb.toString().asComment());
    }

    override fun receivedOraclePriceScheduledResponse(msg: OraclePriceScheduledResponse) {
        emit("Oracle price is outdated".asComment());
    }

    override fun receivedOracleNewAddressResponse(msg: OracleNewAddressResponse) {
        emit("Oracle updated.".asComment());
        self.priceOracleAddress = msg.newAddress;
        if (msg.queryID > 0) {
            // Resend request to new oracle address
            self.requestPriceFromOracle();
        }
    }

}
