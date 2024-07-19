// Messages for updating oracle contract data.

struct RandomHash {
    timestamp: Int as uint32;
    recipient: Address;
    valueHash: Slice;
}

struct RandomValue {
    doraId: Int as uint64;
    name: String;
    revealTimestamp: Int as uint32;
    nonce: Int as uint64;
    txHash: Slice;
}

message(0xbb15fe7d) UpdateCommit {
    signature: Slice;
    payload: RandomHash;
}

message(0x6b91a49a) UpdateReveal {
    signature: Slice;
    payload: RandomValue;
}

struct RevealedValue {
    timestamp: Int as uint32;
    recipient: Address;
    nonce: Int as uint64;
    doraId: Int as uint64;
    name: String;
}
