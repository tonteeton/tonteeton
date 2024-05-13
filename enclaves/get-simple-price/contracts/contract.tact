import "@stdlib/stoppable";
import "@stdlib/ownable";
import "@stdlib/deploy";
import "./protocol.tact";
import "./enclaveProtocol.tact";

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

contract OracleContract with Deployable, Resumable, UsesTickers {

    owner: Address;
    demoAddress: Address;
    enclavePublicKey: Int as uint256;
    prices: map<Int, PriceUpdate>;
    stopped: Bool = false;


    init(publicKey: Int, owner: Address) {
        self.owner = owner;
        self.enclavePublicKey = publicKey;
        self.demoAddress = myAddress();
    }

    get fun prices() : map<Int, PriceUpdate> {
        return self.prices;

    }

    get fun price(ticker: Int) : PriceUpdate? {
        return self.prices.get(ticker);

    }

    get fun demoAddress() : Address {
        return self.demoAddress;
    }

    receive ("debug") {
      if (self.demoAddress == myAddress()) {
        self.deployDemo();
        emit("Demo contract deployed".asComment());
      } else {
        send(SendParameters{
            to: self.demoAddress,
            value: ton("0.01"),
            mode: SendIgnoreErrors,
            body: "callOracle".asComment()
        });
      }

  }

    receive ("DeployDemo") {
        self.deployDemo();
        emit("Demo contract deployed".asComment());
    }

    receive (msg: OraclePriceRequest) {
        emit("OraclePriceRequest".asComment());
        if (self.prices.get(msg.ticker) == null) {
            self.reply(OraclePriceScheduledResponse{
                queryID: msg.queryID,
                ticker: msg.ticker,
                lastUpdatedAt: 0
            }.toCell());
            emit("Unkwown ticker requested".asComment());
        } else {
            let price: PriceUpdate = self.prices.get(msg.ticker)!!;
            if (price.lastUpdatedAt >= msg.minUpdatedAt) {
                self.reply(OraclePriceResponse{
                    queryID: msg.queryID,
                    lastUpdatedAt: price.lastUpdatedAt,
                    ticker: msg.ticker,
                    usd: price.usd,
                    usd24vol: price.usd24vol,
                    usd24change: price.usd24change,
                    btc: price.btc
                }.toCell());
            } else {
                self.reply(OraclePriceScheduledResponse{
                    queryID: msg.queryID,
                    ticker: msg.ticker,
                    lastUpdatedAt: price.lastUpdatedAt
                }.toCell());
                emit("Price is outdated".asComment());
            }
        }
    }

    receive(msg: Update) {
        let payloadHash: Int = msg.payload.toCell().hash();
        require(checkSignature(payloadHash, msg.signature, self.enclavePublicKey), "Invalid signature");

        let price: PriceUpdate = msg.payload;
        self.prices.set(price.ticker, price);

        emit("Update".asComment());
    }

    receive (text: String) {
        emit("text received".asComment());
    }

    fun deployDemo() {
        require(self.demoAddress == myAddress(), "Demo contract is already deployed");
        let init: StateInit = initOf DemoContract(myAddress());
        self.demoAddress = contractAddress(init);
        send(SendParameters{
            to: self.demoAddress,
            value: ton("0.2"),
            mode: SendIgnoreErrors,
            code: init.code,
            data: init.data,
            body: "callOracle".asComment()
        });
    }

}
