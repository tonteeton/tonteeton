import "@stdlib/stoppable";
import "@stdlib/ownable";
import "@stdlib/deploy";
import "./demo.tact";
import "./enclaveProtocol.tact";
import "./oracleProtocol.tact";
import "./movable.tact";

contract OracleContract with Deployable, Resumable, Movable, UsesTickers {

    owner: Address;
    demoAddress: Address;
    newAddress: Address;
    prevAddress: Address?;

    stopped: Bool;
    moved: Bool;

    prices: map<Int, PriceUpdate>;

    enclavePublicKey: Int as uint256;
    enclaveMeasurment: Int as uint256;
    enclaveAttestation: String;

    init(owner: Address, prevAddress: Address?, publicKey: Int, enclaveMeasurment: Int, enclaveAttestation: String) {
        self.owner = owner;
        self.enclavePublicKey = publicKey;
        self.enclaveMeasurment = enclaveMeasurment;
        self.enclaveAttestation  = enclaveAttestation;
        self.demoAddress = myAddress();
        self.newAddress = myAddress();
        self.prevAddress = prevAddress;
        self.stopped = false;
        self.moved = false;
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

    get fun enclavePublicKey() : Int {
        return self.enclavePublicKey;
    }

    get fun enclaveMeasurment() : Int {
        return self.enclaveMeasurment;
    }

    get fun enclaveAttestation() : String {
        return self.enclaveAttestation;
    }

    get fun balance() : Int {
        return myBalance();
    }

    receive (msg: OraclePriceRequest) {
        self.requireNotStopped();
        if (self.moved) {
            self.replyNewAddress(msg.queryId);
            return;
        }
        require(self.prices.get(msg.ticker) != null, "Unknown ticker");

        let price: PriceUpdate = self.prices.get(msg.ticker)!!;
        if (price.lastUpdatedAt >= msg.minUpdatedAt) {
            self.reply(OraclePriceResponse{
                queryId: msg.queryId,
                payload: msg.customPayload,
                lastUpdatedAt: price.lastUpdatedAt,
                ticker: msg.ticker,
                usd: price.usd,
                usd24vol: price.usd24vol,
                usd24change: price.usd24change,
                btc: price.btc
            }.toCell());
        } else {
            // Notify the caller that the up-to-date price is not known to the oracle.
            // The enclave app should discover this transaction and update the oracle with the price accordingly.
            self.reply(OraclePriceScheduledResponse{
                queryId: msg.queryId,
                payload: msg.customPayload,
                ticker: msg.ticker,
                lastUpdatedAt: price.lastUpdatedAt,
            }.toCell());
        }
    }

    receive(msg: Update) {
        self.requireNotStopped();
        self.requireNotMoved();

        let payloadHash: Int = msg.payload.toCell().hash();
        require(checkSignature(payloadHash, msg.signature, self.enclavePublicKey), "Invalid signature");

        let price: PriceUpdate = msg.payload;

        require(price.lastUpdatedAt <= now(), "Price update is from the future");
        if (self.prices.get(price.ticker) != null) {
            let prevPrice: PriceUpdate = self.prices.get(price.ticker)!!;
            require(price.lastUpdatedAt > prevPrice.lastUpdatedAt, "Price update is outdated");
        }

        self.prices.set(price.ticker, price);

        self.reply(price.lastUpdatedAt.toString().asComment());
    }

    receive ("DeployDemo") {
        self.requireOwner();
        self.deployDemo();
        let sb: StringBuilder = beginString();
        sb.append("Demo: ");
        sb.append(self.demoAddress.toString());
        emit(sb.toString().asComment());
    }

    receive ("withdrawBalance") {
        self.requireOwner();
        send(SendParameters{
            to: self.owner,
            bounce: true,
            value: 0,
            mode: SendRemainingBalance + SendIgnoreErrors
        });
    }

    receive ("DemoReady") {
        require(sender() == self.demoAddress, "Unknown Demo");
        emit("Demo is ready".asComment());
    }

    receive (text: String) {

    }

    fun deployDemo() {
        let init: StateInit = initOf DemoContract(myAddress());
        self.demoAddress = contractAddress(init);
        send(SendParameters{
            to: self.demoAddress,
            value: ton("0.2"),
            mode: SendIgnoreErrors,
            body: "ReportOracle".asComment(),
            code: init.code,
            data: init.data,
        });
    }

}
