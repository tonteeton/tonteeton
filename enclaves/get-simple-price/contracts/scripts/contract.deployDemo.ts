import { Address, contractAddress, toNano } from "@ton/core";
import { delay, getDemoContractAddress, isTestnet, nextSeqno, newSender, newTonClient, openOracleContract } from "./utils";

(async () => {
    const oracle = await openOracleContract();
    const client = newTonClient();
    const senderCreated = await newSender(client);
    const wallet = senderCreated.wallet;
    const sender = senderCreated.sender;
    const address = await getDemoContractAddress();

    await oracle.send(sender, { value: toNano("0.3") }, "DeployDemo");
    await nextSeqno(wallet);
    console.log(`Demo contract address: ${address.toString({ testOnly: isTestnet() })}`);

})();

