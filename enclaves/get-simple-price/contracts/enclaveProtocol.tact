// Messages for updating oracle contract data.

struct PriceUpdate {
    lastUpdatedAt: Int as uint64;
    ticker: Int as uint64;
    usd: Int as uint64;
    usd24vol: Int as uint64;
    usd24change: Int as int64;
    btc: Int as uint64;
}

message(0x9f89304e) Update {
    signature: Slice;
    payload: PriceUpdate;
}
