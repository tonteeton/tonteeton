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
        sb.append("New price received: ");
        sb.append(msg.usd.toString());
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
    lastPrice: OraclePriceResponse;
    stopped: Bool;


    init(publicKey: Int, owner: Address) {
        self.owner = owner;
        self.enclavePublicKey = publicKey;
        self.lastPrice = OraclePriceResponse{
            queryID: 0,
            ticker: self.TICKER_TON,
            usd: 0,
            lastUpdatedAt: now()
        };
        self.demoAddress = myAddress();
        self.stopped = false;
    }

  get fun price() : OraclePriceResponse {
      return self.lastPrice;

  }

 get fun demoAddress() : Address {
     return self.demoAddress;
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
        let lastUpdatedAt : Int = now() - 3600;
        if (lastUpdatedAt >= msg.minUpdatedAt) {
            self.reply(OraclePriceResponse{
                queryID: msg.queryID,
                ticker: msg.ticker,
                usd: 5704518,
                lastUpdatedAt: lastUpdatedAt
            }.toCell());
        } else {
            self.reply(OraclePriceScheduledResponse{
                queryID: msg.queryID,
                ticker: msg.ticker,
                lastUpdatedAt: lastUpdatedAt
            }.toCell());
        }
        emit("price response sended".asComment());
    }

    receive(msg: Update) {
        let payloadHash: Int = msg.payload.toCell().hash();
        require(checkSignature(payloadHash, msg.signature, self.enclavePublicKey), "Invalid signature");
        emit("Update".asComment());
    }

    receive (text: String) {
        emit("text received".asComment());
    }

}
