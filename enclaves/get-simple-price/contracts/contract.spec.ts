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
import { MoveTo, OracleContract, OraclePriceRequest, PriceUpdate, storePriceUpdate } from "./output/oracle_OracleContract";
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
    let deployer: SandboxContract<TreasuryContract>;
    let sender: Sender;
    let anonymous: SandboxContract<TreasuryContract>;
    let anonymousSender: Sender;
    let enclaveKeyPair: KeyPair;
    let validPayload: PriceUpdate;
    const now = Math.floor(Date.now() / 1000);
    const attestationReport = "test".repeat(257);

    beforeEach(async () => {
        enclaveKeyPair = keyPairFromSecretKey(
            Buffer.from(
                "c8c24d89465fde431e12443ed2be7bf8ebbc0c47ce2a6342fc108df5cd937cf7393c80a2c5e690f2f955ffbf4e0112ed8513324ce5c4543535ebfd4963704fd5",
                "hex"
            )
        );

        validPayload = {
            $$type: "PriceUpdate",
            lastUpdatedAt: BigInt(now),
            ticker: BigInt(0x72716023),
            usd: BigInt(345),
            usd24vol: BigInt(81968225604),
            usd24change: BigInt(1566),
            btc: BigInt(10967),
        };

        blockchain = await Blockchain.create();
        blockchain.now = now;
        blockchain.verbosity = {
            ...blockchain.verbosity,
            blockchainLogs: true,
            vmLogs: "vm_logs_full",
            debugLogs: true,
            print: false,
        };

        owner = await blockchain.treasury("owner");
        sender = owner.getSender();

        anonymous = await blockchain.treasury("anonymous");
        anonymousSender = anonymous.getSender();

        deployer = await blockchain.treasury("deployer");
        contract = blockchain.openContract(
            await OracleContract.fromInit(
                owner.address,
                null,
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

        expect((await contract.getDemoAddress()).toString()).toEqual(contract.address.toString());

        client = blockchain.openContract(await DemoContract.fromInit(contract.address));
        res = await client.send(deployer.getSender(), { value: toNano(5) }, "topup");
        expect((await client.getPriceOracleAddress()).toString()).toEqual(contract.address.toString());

        res = await contract.send(sender, { value: toNano("0.1") }, "DeployDemo");
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: true,
        });
        expect((await contract.getDemoAddress()).toString()).not.toEqual(contract.address.toString());
    });


    async function sendUpdate(payload: PriceUpdate, expectSuccess: boolean = true) {
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
            signature.toString("base64"),
            "\nlastUpdatedAt:",
            payload.lastUpdatedAt,
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
            success: expectSuccess,
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
        let payload = structuredClone(validPayload);
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

    it("should reject the outdated price update", async () => {
        let payload = structuredClone(validPayload);
        payload.lastUpdatedAt -= BigInt(10);
        await sendUpdate(payload);
        await sendUpdate(payload, false);
        payload.lastUpdatedAt += BigInt(10);
        await sendUpdate(payload);
    });

    it("should reject prices from the future", async () => {
        let payload = structuredClone(validPayload);
        payload.lastUpdatedAt += BigInt(600);
        await sendUpdate(payload, false);
    });

    it("should respond to price request with a price response", async () => {
        await sendUpdate(validPayload);

        let res = await client.send(sender, { value: toNano("0.03") }, "CallOracle");
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
        let res = await client.send(sender, { value: toNano("0.0123") }, "CallOracle");
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
        outdatedPrice.lastUpdatedAt -= BigInt(30 * 24 * 60 * 60);
        await sendUpdate(outdatedPrice);

        const initialOracleBalance = await contract.getBalance();

        let res = await client.send(sender, { value: toNano("0.02") }, "CallOracle");

        printTransactionFees(res.transactions);
        // The oracle balance remains unchanged after processing the request.
        expect(await contract.getBalance()).toBeGreaterThan(initialOracleBalance - BigInt(10));

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

    it("should send whole balance to owner", async () => {
        expect(await contract.getBalance()).toBeGreaterThan(0);
        let res = await contract.send(sender, { value: toNano("0.1") }, "withdrawBalance");
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: owner.address,
            success: true,
        });
        expect(await contract.getBalance()).toBe(BigInt(0));
    });

    it("owner is required to withdraw balance", async () => {
        const initialOracleBalance = await contract.getBalance();
        expect(initialOracleBalance).toBeGreaterThan(0);
        let res = await contract.send(anonymousSender, { value: toNano("0.1") }, "withdrawBalance");
        expect(res.transactions).toHaveTransaction({
            from: anonymous.address,
            to: contract.address,
            success: false,
        });
        expect(await contract.getBalance()).toBeGreaterThan(initialOracleBalance - BigInt(10));
    });

    it("oracle moving to the new address started", async () => {
        expect((await contract.getNewAddress()).toString()).toEqual(contract.address.toString());
        expect((await client.getPriceOracleAddress()).toString()).toEqual(contract.address.toString());

        let res = await contract.send(sender, { value: toNano("0.1") }, {
            $$type: "MoveTo",
            newAddress: anonymous.address,
            moveCompleted: false,
        });
        expect((await contract.getNewAddress()).toString()).toEqual(anonymous.address.toString());

        await sendUpdate(validPayload);
        res = await client.send(sender, { value: toNano("0.05") }, "CallOracle");
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            op: findOp(contract, "OraclePriceResponse"),
        });

        res = await contract.send(sender, { value: toNano("0.01") }, "NewAddress");
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: owner.address,
            success: true,
            op: findOp(contract, "OracleNewAddressResponse"),
        });

    });

    it("oracle not move without confirmation", async () => {
        expect((await contract.getNewAddress()).toString()).toEqual(contract.address.toString());
        expect((await client.getPriceOracleAddress()).toString()).toEqual(contract.address.toString());

        let res = await contract.send(sender, { value: toNano("0.2") }, {
            $$type: "MoveTo",
            newAddress: anonymous.address,
            moveCompleted: true,
        });
        expect((await contract.getNewAddress()).toString()).toEqual(anonymous.address.toString())

        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: anonymous.address,
            body: beginCell().storeUint(0, 32).storeStringTail("MoveConfirmation").endCell(),
            success: true,
        });

        await sendUpdate(validPayload, true); // Old contract is still functional
    });

    it("oracle moved to the new address", async () => {
        expect((await contract.getNewAddress()).toString()).toEqual(contract.address.toString());
        expect((await client.getPriceOracleAddress()).toString()).toEqual(contract.address.toString());

        let nextContract: SandboxContract<OracleContract>;
        nextContract = blockchain.openContract(
            await OracleContract.fromInit(
                owner.address,
                contract.address, // prevAddress
                BigInt("0x" + enclaveKeyPair.publicKey.toString("hex")),
                BigInt("0xef6d2adf7f08c3ea88305d7c9c73ad9837c60227db2378de8fc7d5e619637134"),
                attestationReport,
            )
        );
        let res = await nextContract.send(deployer.getSender(), { value: toNano(5) }, { $$type: "Deploy", queryId: 0n });
        expect(res.transactions).toHaveTransaction({
            from: deployer.address,
            to: nextContract.address,
            success: true,
            deploy: true,
        });
        expect((contract.address)).not.toEqual(nextContract.address);

        res = await contract.send(sender, { value: toNano("0.2") }, {
            $$type: "MoveTo",
            newAddress: nextContract.address,
            moveCompleted: true,
        });

        expect((await contract.getNewAddress()).toString()).toEqual(nextContract.address.toString());

        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: nextContract.address,
            body: beginCell().storeUint(0, 32).storeStringTail("MoveConfirmation").endCell(),
            success: true,
        });

        expect(res.transactions).toHaveTransaction({
            from: nextContract.address,
            to: contract.address,
            body: beginCell().storeUint(0, 32).storeStringTail("MoveCompleted").endCell(),
            success: true,
        });

        await sendUpdate(validPayload, false); // Old contract is disabled

        res = await client.send(sender, { value: toNano("0.02") }, "CallOracle");
        expect(res.transactions).not.toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            op: findOp(contract, "OraclePriceResponse"),
        });
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            op: findOp(contract, "OracleNewAddressResponse"),
        });

        expect((await client.getPriceOracleAddress()).toString()).toEqual(nextContract.address.toString());

    });


});
