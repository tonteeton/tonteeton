import "./oracleProtocol.tact";

contract DemoContract with UsesPriceOracle {

    priceOracleAddress : Address;

    init(priceOracleAddress: Address) {
        self.priceOracleAddress = priceOracleAddress;
    }

    get fun balance() : Int {
        return myBalance();
    }

    receive ("callOracle") {
        self.requestPriceFromOracle();
    }

    receive (text: String) {
    }

    override fun receivedOraclePriceResponse(msg: OraclePriceResponse) {
        require(sender() == self.priceOracleAddress, "Unknown oracle");
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


}
