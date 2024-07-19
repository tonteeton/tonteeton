// TonTeeTon get-random-winner contract protocol.

// Message structure for the Oracle state.
message (0x79a88945) OracleState {
    name: String;                // Name of the oracle contract.
    version: Int as uint32;      // Version of the oracle contract, represented as an integer.
    owner: Address;              // Address of the oracle contract owner.
}

message (0x2b1aeeee) SendPrize {
    address: Address; // Address of the winner.
}