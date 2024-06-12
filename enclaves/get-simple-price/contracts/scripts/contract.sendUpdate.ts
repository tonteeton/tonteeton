import * as fs from "fs";
import * as path from "path";
import {
    Address,
    beginCell,
    Builder,
    Cell,
    Contract,
    contractAddress,
    ContractProvider,
    Message,
    Sender,
    storeTransaction,
    toNano,
} from "@ton/core";
import { delay, getContractAddress, getContractInitParams, initContract, nextSeqno, newTonClient, newSender} from "./utils";

interface EnclaveResponse {
    signature: string;
    payload: string;
    hash: string;
}

function loadEnclaveResponse(): EnclaveResponse {
    return JSON.parse(fs.readFileSync(getEnclaveResponse(), "utf-8"));
}

function buildBody(response: EnclaveResponse) {
    let signature = Buffer.from(response.signature, "base64");
    let signatureCell = beginCell().storeBuffer(signature).endCell();
    let payloadCell = Cell.fromBase64(response.payload);
    return beginCell()
        .storeUint(0x9f89304e, 32)
        .storeRef(signatureCell)
        .storeBuilder(payloadCell.asBuilder())
        .endCell();
}

(async () => {
    const client = newTonClient();
    const senderCreated = await newSender(client);
    const wallet = senderCreated.wallet;
    const sender = senderCreated.sender;

    let seqno = await wallet.getSeqno();

    await sender.send({
        to: await getContractAddress(),
        value: toNano("0.05"),
        body: buildBody(loadEnclaveResponse()),
        bounce: false,
    });
    await nextSeqno(wallet);
})();


function getEnclaveResponse() {
    const args = process.argv.slice(2);
    if (args.length < 1) throw new Error("Path to enclave response is not specified.");
    return args[0];
}
