import "@stdlib/stoppable";
import "@stdlib/ownable";
import "@stdlib/deploy";
import "./demo.tact";
import "./enclaveProtocol.tact";
import "./oracleProtocol.tact";
import "./movable.tact";

// OracleContract is a TonTeeTon app contract for providing TON price updates.
// Contract source code is at https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-simple-price/contracts
contract OracleContract with Deployable, Resumable, Movable, UsesTickers {
    // Contract name and version, to include in state report.
    const contractName: String = "get-simple-price";
    const contractVersion: Int = 1 << 24 | 1 << 16 | 0;

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

    // Returns the map of prices.
    get fun prices() : map<Int, PriceUpdate> {
        return self.prices;

    }

    // Returns the price update for a given ticker.
    get fun price(ticker: Int) : PriceUpdate? {
        return self.prices.get(ticker);

    }

    // Returns the address of deployed demo contract.
    get fun demoAddress() : Address {
        return self.demoAddress;
    }

    // Returns the public key (signature) of the enclave.
    get fun enclavePublicKey() : Int {
        return self.enclavePublicKey;
    }

    // Returns the measurement of the enclave.
    get fun enclaveMeasurment() : Int {
        return self.enclaveMeasurment;
    }

    // Returns the attestation report of the enclave.
    get fun enclaveAttestation() : String {
        return self.enclaveAttestation;
    }

    // Returns the balance of the contract.
    get fun balance() : Int {
        return myBalance();
    }

    // Returns the state of the Oracle contract.
    get fun state() : OracleState {
        return self.getState();
    }

    // Handles an OraclePriceRequest message.
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

    // Handles an Update message from enclave to update the prices.
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

    receive ("WithdrawBalance") {
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

    receive ("State"){
        self.reply(self.getState().toCell());
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

    fun getState() : OracleState {
        return OracleState{
            name: self.contractName,
            version: self.contractVersion,
            owner: self.owner,
            prevAddress: self.prevAddress,
            newAddress: self.newAddress,
            moved: self.moved,
            stopped: self.stopped,
        };
    }
}
