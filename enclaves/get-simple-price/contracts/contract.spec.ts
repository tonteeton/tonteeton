import "@ton/test-utils";
import { getSecureRandomBytes, keyPairFromSeed, keyPairFromSecretKey, sign, signVerify, KeyPair } from "@ton/crypto";
import { Blockchain, SandboxContract, TreasuryContract, printTransactionFees, internal } from "@ton/sandbox";
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
import { OracleContract, OraclePriceRequest, PriceUpdate, storePriceUpdate } from "./output/oracle_OracleContract";
import { DemoContract } from "./output/oracle_DemoContract";

function findOp(contract: SandboxContract<OracleContract>, name: string) {
    let selected = contract.abi.types?.find((type) => type?.name === name);
    if (!selected?.header) {
        throw new Error(`Opcode '${name}' not found in ABI types.`);
    }
    return selected.header;
}


function errorCode(contract: SandboxContract<OracleContract>, message: string): number {
    const entries = Object.entries(contract.abi.errors || {});
    const foundEntry = entries.find(([errorCode, error]) => error.message === message);
    if (foundEntry) {
        return parseInt(foundEntry[0]);
    }
    throw new Error(`Error code for message "${message}" not found`);
}


describe("contract", () => {
    let blockchain: Blockchain;
    let contract: SandboxContract<OracleContract>;
    let client: SandboxContract<DemoContract>;
    let owner: SandboxContract<TreasuryContract>;
    let sender: Sender;
    let enclaveKeyPair: KeyPair;
    const now = Math.floor(Date.now() / 1000);
    const validPayload: PriceUpdate = {
        $$type: "PriceUpdate",
        lastUpdatedAt: BigInt(now),
        ticker: BigInt(0x72716023),
        usd: BigInt(345),
        usd24vol: BigInt(81968225604),
        usd24change: BigInt(1566),
        btc: BigInt(10967),
    };
    const attestationReport = "test".repeat(257);

    beforeEach(async () => {
        enclaveKeyPair = keyPairFromSecretKey(
            Buffer.from(
                "c8c24d89465fde431e12443ed2be7bf8ebbc0c47ce2a6342fc108df5cd937cf7393c80a2c5e690f2f955ffbf4e0112ed8513324ce5c4543535ebfd4963704fd5",
                "hex"
            )
        );
        blockchain = await Blockchain.create();
        blockchain.verbosity = {
            ...blockchain.verbosity,
            blockchainLogs: true,
            vmLogs: "vm_logs_full",
            debugLogs: true,
            print: false,
        };

        owner = await blockchain.treasury("owner");
        sender = owner.getSender();

        let deployer = await blockchain.treasury("deployer");
        contract = blockchain.openContract(
            await OracleContract.fromInit(
                owner.address,
                BigInt("0x" + enclaveKeyPair.publicKey.toString("hex")),
                BigInt("0xef6d2adf7f08c3ea88305d7c9c73ad9837c60227db2378de8fc7d5e619637134"),
                attestationReport,
            )
        );
        let res = await contract.send(deployer.getSender(), { value: toNano(5) }, { $$type: "Deploy", queryId: 0n });
        expect(res.transactions).toHaveTransaction({
            from: deployer.address,
            to: contract.address,
            success: true,
            deploy: true,
        });
        expect((await contract.getDemoAddress()) == contract.address);

        client = blockchain.openContract(await DemoContract.fromInit(contract.address));
        res = await client.send(deployer.getSender(), { value: toNano(5) }, "topup");
        expect((await client.getPriceOracleAddress()) == contract.address);

        res = await contract.send(sender, { value: toNano("0.023") }, "DeployDemo");
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: true,
        });
        expect((await contract.getDemoAddress()) != contract.address);
    });


    async function sendUpdate(payload: PriceUpdate) {
        let hash = beginCell().store(storePriceUpdate(payload)).endCell().hash();
        let signature = sign(hash, enclaveKeyPair.secretKey);

        console.log(
            "Private key:",
            enclaveKeyPair.secretKey.toString("base64"),
            "\nPublic key:",
            enclaveKeyPair.publicKey.toString("hex"),
            "\nPayload:",
            beginCell().store(storePriceUpdate(payload)).endCell().toBoc().toString("base64"),
            "\nHash:",
            hash.toString("base64"),
            "\nSignature:",
            signature.toString("base64")
        );

        let res = await contract.send(
            sender,
            { value: toNano("0.1") },
            {
                $$type: "Update",
                signature: beginCell().storeBuffer(signature).endCell(),
                payload: payload,
            }
        );
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: true,
            op: findOp(contract, "Update"),
        });
    }


    it("should deploy correctly", async () => {});

    it("should have enclave measurment", async () => {
        expect(await contract.getEnclaveMeasurment()).toBe(
            BigInt("108295653040254449567541984650930352199134953056630027524746292829709349908788")
        );
    });

    it("should have attestation report", async () => {
        expect(await contract.getEnclaveAttestation()).toBe(attestationReport);
    });

    it("should handle price update from enclave", async () => {
        await sendUpdate(validPayload);
    });

    it("should reject requests with wrong signature", async () => {
        let payload = validPayload;
        let hash = beginCell().store(storePriceUpdate(payload)).endCell().hash();
        payload.usd += BigInt(1);
        let signature = sign(hash, enclaveKeyPair.secretKey);

        let res = await contract.send(
            sender,
            { value: toNano("0.1") },
            {
                $$type: "Update",
                signature: beginCell().storeBuffer(signature).endCell(),
                payload: payload,
            }
        );
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: false,
            op: findOp(contract, "Update"),
            exitCode: 48401,
        });
    });


    it("should respond to price request with a price response", async () => {
        await sendUpdate(validPayload);

        let res = await client.send(sender, { value: toNano("0.02") }, "callOracle");
        expect(res.transactions).toHaveTransaction({
            from: client.address,
            to: contract.address,
            success: true,
            op: findOp(contract, "OraclePriceRequest"),
        });
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            op: findOp(contract, "OraclePriceResponse"),
        });

    });

    it("should throw Unknown Ticker", async () => {
        let unknownPrice : PriceUpdate = structuredClone(validPayload);
        unknownPrice.ticker = BigInt(12345678);
        let res = await client.send(sender, { value: toNano("0.0123") }, "callOracle");
        expect(res.transactions).toHaveTransaction({
            from: client.address,
            to: contract.address,
            success: false,
            op: findOp(contract, "OraclePriceRequest"),
            exitCode: errorCode(contract, `Unknown ticker`),
        });
    });

    it("should notify if available price is outdated", async () => {
        let outdatedPrice : PriceUpdate = structuredClone(validPayload);
        outdatedPrice.lastUpdatedAt = BigInt(now - 30 * 24 * 60 * 60);
        console.log(validPayload);
        console.log(outdatedPrice);
        await sendUpdate(outdatedPrice);

        const initialOracleBalance = await contract.getBalance();

        let res = await client.send(sender, { value: toNano("0.02") }, "callOracle");

        console.log('total fees = ', res.transactions[1].totalFees);

        printTransactionFees(res.transactions);


        // The oracle balance remains unchanged after processing the request.
        expect(await contract.getBalance()).toBe(initialOracleBalance);

        expect(res.transactions).toHaveTransaction({
            from: client.address,
            to: contract.address,
            success: true,
            op: findOp(contract, "OraclePriceRequest"),
        });
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            op: findOp(contract, "OraclePriceScheduledResponse"),
        });
    });
});
