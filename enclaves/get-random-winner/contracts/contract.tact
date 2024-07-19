import "@stdlib/ownable";
import "@stdlib/deploy";
import "./enclaveProtocol.tact";
import "./oracleProtocol.tact";
import "./state.tact";

// OracleContract is a TonTeeTon app contract designed to run an event that randomly selects a winner.
// Contract source code is at https://github.com/tonteeton/tonteeton/tree/main/enclaves/get-random-winner/contracts
contract OracleContract with Deployable, Ownable {
    // Contract name and version, to include in state report.
    const contractName: String = "get-random-winner";
    const contractVersion: Int = 1 << 24 | 1 << 16 | 0;
    // Hardcoded event completion time.
    const eventCompletionTime: Int = 1721685600;
    // Hard limits for resources.
    const requestsMaxCount: Int = 128;
    const gasConsumption: Int = ton("0.03");

    owner: Address;
    newAddress: Address;
    prevAddress: Address?;

    enclavePublicKey: Int as uint256;
    enclaveMeasurment: Int as uint256;
    enclaveAttestation: String;

    randomHash: RandomHash;
    state: EventState;

    init(owner: Address, prevAddress: Address?, publicKey: Int, enclaveMeasurment: Int, enclaveAttestation: String) {
        self.owner = owner;
        self.enclavePublicKey = publicKey;
        self.enclaveMeasurment = enclaveMeasurment;
        self.enclaveAttestation  = enclaveAttestation;
        self.newAddress = myAddress();
        self.prevAddress = prevAddress;

        self.state = initialEventState(self.eventCompletionTime);

        self.randomHash = RandomHash{
            timestamp: 0,
            recipient: myAddress(),
            valueHash: beginCell().asSlice(),
        };

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

    // Returns the state of the Oracle contract.
    get fun eventState() : EventState {
        return self.getEventState();
    }

    // Handles an Update message from enclave: random value commit.
    receive(msg: UpdateCommit) {
        let payloadHash: Int = msg.payload.toCell().hash();
        require(checkSignature(payloadHash, msg.signature, self.enclavePublicKey), "Invalid signature");
        require(msg.payload.recipient == myAddress(), "Received an update intended for another contract.");
        self.state.requireNotOutdated(msg.payload.timestamp);

        let err : String = self.state.waitReveal();
        if (err == "") {
            self.randomHash = msg.payload;
            emit("reveal()".asComment());
        } else {
            self.reply(err.asComment());
        }
    }

    // Handles an Update message from enclave: random value reveal.
    receive(msg: UpdateReveal) {
        let payloadHash: Int = msg.payload.toCell().hash();
        require(checkSignature(payloadHash, msg.signature, self.enclavePublicKey), "Invalid signature");

        let randomValue: RandomValue = msg.payload;
        let revealed: RevealedValue =  RevealedValue{
            timestamp: self.randomHash.timestamp,
            recipient: myAddress(),
            nonce: randomValue.nonce,
            doraId: randomValue.doraId,
            name: randomValue.name,
        };

        let sb1: StringBuilder = beginString();
        sb1.append(revealed.toCell().hash().toString());
        emit(sb1.toString().asComment());

        require(revealed.toCell().hash() == self.randomHash.valueHash.loadUint(256), "Invalid hash for revealed value");

        let err : String = self.state.reveal(revealed.doraId, revealed.name);
        if (err == "") {
            let sb: StringBuilder = beginString();
            sb.append("#");
            sb.append(self.state.doraId.toString());
            emit(sb.toString().asComment());
        } else {
            self.reply(err.asComment());
        }
    }

    // Send prize to the winner address.
    receive (msg: SendPrize) {
        self.requireOwner();
        self.state.requireCompleted();
        send(SendParameters{
            to: msg.address,
            bounce: true,
            value: 0,
            mode: SendRemainingBalance
        });
    }

    receive (text: String) {
        if (text == "rock" || text == "Rock" || text == "ROCK") {
            self.rock();
        } else if (text == "roll" || text == "Roll" || text == "ROLL") {
            self.roll();
        }
    }

    fun rock() {
        let err : String = self.state.rock(context().value);
        if (err == "") {
            self.sendText(sender(), "🤘");
        } else {
            self.reply(err.asComment());
        }
    }

    fun roll() {
        let err : String = self.state.roll(context().value);
        if (err == "") {
            if self.state.changed == stateRoll {
                emit("random()".asComment());
            }
            self.sendText(sender(), "⚄");
        } else {
            self.reply(err.asComment());
        }
    }

    fun getState() : OracleState {
        return OracleState{
            name: self.contractName,
            version: self.contractVersion,
            owner: self.owner,
        };
    }

    fun getEventState() : EventState {
        let state : EventState = self.state;
        state.prize = myBalance() - ton("0.001");
        return state;
    }

    fun sendText(to: Address, text: String) {
        send(SendParameters{
            to: to,
            value: ton("0.01"),
            mode: SendIgnoreErrors,
            bounce: false,
            body: text.asComment(),
        });
    }

}
