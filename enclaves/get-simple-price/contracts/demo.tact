import "./oracleProtocol.tact";

contract DemoContract with UsesPriceOracle {

    priceOracleAddress : Address;

    init(priceOracleAddress: Address) {
        self.priceOracleAddress = priceOracleAddress;
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

    receive ("callOracle") {
        emit("callOracle".asComment());
        self.requestPriceFromOracle();
    }

    receive (text: String) {
    }

}
