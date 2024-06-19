import { Address, toNano } from "@ton/core";
import { delay, openOracleContract, nextSeqno, newSender, newTonClient } from "./utils";

(async () => {
    const oracle = await openOracleContract();
    const client = newTonClient();
    const senderCreated = await newSender(client);
    const wallet = senderCreated.wallet;
    const sender = senderCreated.sender;

    const newAddress = getNewAddress();
    const moveCompleted = getMoveCompleted();
    console.log(`Move to new address: <${newAddress}>. Move completed: <${moveCompleted}>`);

    await oracle.send(sender, { value: toNano("0.3") }, {
        $$type: "MoveTo",
        newAddress: newAddress,
        moveCompleted: moveCompleted,
    });
    await nextSeqno(wallet);

})();


function getNewAddress() : Address {
    const args = process.argv.slice(2);
    if (args.length < 1) throw new Error("New address is not specified.");
    return Address.parse(args[0]);
}


function getMoveCompleted() : boolean {
    const args = process.argv.slice(2);
    if (args.length < 2) throw new Error("`moveCompleted` is not speciied.");
    const stringToBool = new Map([["1", true], ["0", false]]);
    const moveCompleted = stringToBool.get(args[1]);
    if (moveCompleted === undefined) {
        throw new Error("Invalid value for `moveCompleted`. Must be '1' or '0'.");
    }
    return moveCompleted;
}


