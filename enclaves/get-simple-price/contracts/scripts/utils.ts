import { Address, Contract, contractAddress, toNano } from "@ton/core";
import { TonClient, WalletContractV3R2 } from "@ton/ton";
import { mnemonicNew, mnemonicToPrivateKey } from "@ton/crypto";
import { OracleContract } from "../output/oracle_OracleContract";
import { DemoContract } from "../output/oracle_DemoContract";

export async function getContractInitParams(): Promise<[Address, bigint, bigint, string]> {
    const requiredEnvVars = [
        "TON_CONTRACT_OWNER",
        "ENCLAVE_PUBLIC_KEY",
        "ENCLAVE_MEASUREMENT",
        "ENCLAVE_ATTESTATION"
    ];
    for (const envVar of requiredEnvVars) {
        if (!process.env[envVar]) {
            throw new Error(`${envVar} environment variable is not set.`);
        }
    }

    const TON_CONTRACT_OWNER = Address.parse(process.env.TON_CONTRACT_OWNER!);
    const ENCLAVE_MEASUREMENT = BigInt("0x" + process.env.ENCLAVE_MEASUREMENT!);
    const ENCLAVE_ATTESTATION = process.env.ENCLAVE_ATTESTATION!;
    const ENCLAVE_PUBLIC_KEY = BigInt(
       "0x" + Buffer.from(process.env.ENCLAVE_PUBLIC_KEY!, "base64").toString("hex"),
    );

    return [
        TON_CONTRACT_OWNER,
        ENCLAVE_PUBLIC_KEY,
        ENCLAVE_MEASUREMENT,
        ENCLAVE_ATTESTATION,
    ];
}

export async function initContract() {
    return await OracleContract.init(...await getContractInitParams());
}

export async function initDemoContract() {
    return await DemoContract.init(await getContractAddress());
}

export async function getContractAddress() {
    return contractAddress(0, await initContract());
}

export async function getDemoContractAddress() {
    return contractAddress(0, (await initDemoContract()));
}

export function newTonClient() {
    const TON_TONCENTER_KEY = process.env.TON_TONCENTER_KEY;
    return new TonClient({
        endpoint: isTestnet() ? `https://testnet.toncenter.com/api/v2/jsonRPC` : `https://toncenter.com/api/v2/jsonRPC`,
        apiKey: TON_TONCENTER_KEY,
    })
}

export function isTestnet(): boolean {
    return parseBoolEnv("TON_TESTNET");
}

export async function newSender(client: TonClient) {
    const TON_PRIVATE_MNEMONIC = process.env.TON_PRIVATE_MNEMONIC;
    if (!TON_PRIVATE_MNEMONIC) throw new Error("TON_PRIVATE_MNEMONIC is not set.");
    const key = await mnemonicToPrivateKey(TON_PRIVATE_MNEMONIC.split(" "));
    const wallet = client.open(WalletContractV3R2.create({ publicKey: key.publicKey, workchain: 0 }))
    const sender = wallet.sender(key.secretKey);
    return {
        wallet: wallet,
        sender: sender,
    }
}

export async function openOracleContract(client?: TonClient) {
    return openContract(OracleContract, await getContractAddress(), client)
}

export async function openDemoContract(client?: TonClient) {
    return openContract(DemoContract, await getDemoContractAddress(), client)
}

async function openContract(contractType: any, address: Address, client?: TonClient) {
    if (client === undefined) {
        client = newTonClient();
    }
    let contract = await contractType.fromAddress(address);
    return await client.open(contract);
}

export async function nextSeqno(wallet: any) {
    let seqno = await wallet.getSeqno();
    let newSeqno = seqno;
    while (newSeqno <= seqno) {
        newSeqno = await wallet.getSeqno();
        await delay(500);
    }
}

export function delay(ms: number) {
    return new Promise((resolve) => {
        setTimeout(resolve, ms);
    });
}

export function parseBoolEnv(envVar: string): boolean {
    if (!process.env[envVar]) {
            throw new Error(`${envVar} environment variable is not set.`);
    }
    const value = process.env[envVar]!;
    const stringToBool = new Map([["1", true], ["0", false]]);
    const parsed = stringToBool.get(value);
    if (parsed === undefined) {
        throw new Error(`Invalid value for ${envVar}. Must be '1' or '0'.`);
    }
    return parsed;
}
