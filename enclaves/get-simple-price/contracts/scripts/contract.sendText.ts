import { toNano } from "@ton/core";
import { delay, getDemoContractAddress, isTestnet, nextSeqno, newSender, newTonClient, openOracleContract } from "./utils";

(async () => {
    const oracle = await openOracleContract();
    const client = newTonClient();
    const senderCreated = await newSender(client);
    const wallet = senderCreated.wallet;
    const sender = senderCreated.sender;

    const textMessage = getTextMessage()
    await oracle.send(sender, { value: toNano("0.1") }, textMessage);
    await nextSeqno(wallet);
})();

function getTextMessage() {
    const args = process.argv.slice(2);
    if (args.length < 1) throw new Error("Text to send should be specified, like: 'DeployDemo'.");
    const name = args[0];
    return name.charAt(0).toUpperCase() + name.slice(1);
}
