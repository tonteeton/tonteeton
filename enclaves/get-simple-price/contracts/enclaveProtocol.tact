// Messages for updating oracle contract data.

struct PriceUpdate {
    usd: Int as uint64;
    lastUpdatedAt: Int as uint64;
}

message(0x9f89304e) Update {
    signature: Slice;
    payload: PriceUpdate;
}
