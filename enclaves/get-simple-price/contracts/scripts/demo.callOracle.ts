import { toNano } from "@ton/core";
import { delay, openDemoContract, nextSeqno, newSender, newTonClient } from "./utils";

(async () => {
    const demo = await openDemoContract();
    const client = newTonClient();
    const senderCreated = await newSender(client);
    const wallet = senderCreated.wallet;
    const sender = senderCreated.sender;

    await demo.send(sender, { value: toNano("0.1") }, "CallOracle");
    await nextSeqno(wallet);
})();
