import "@ton/test-utils";
import { getSecureRandomBytes, keyPairFromSeed, keyPairFromSecretKey, sign, signVerify, KeyPair } from "@ton/crypto";
import { Blockchain, SandboxContract, TreasuryContract, prettyLogTransactions, printTransactionFees, internal } from "@ton/sandbox";
import {
    Address,
    beginCell,
    Builder,
    Cell,
    Slice,
    Contract,
    contractAddress,
    ContractProvider,
    Message,
    Sender,
    storeTransaction,
    toNano,
} from "@ton/core";
import { OracleContract, UpdateCommit, RandomHash, RandomValue, RevealedValue, storeUpdateCommit, storeRevealedValue, storeRandomHash, storeRandomValue } from "./output/oracle_OracleContract";


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
    //let client: SandboxContract<DemoContract>;
    let owner: SandboxContract<TreasuryContract>;
    let deployer: SandboxContract<TreasuryContract>;
    let sender: Sender;
    let client: SandboxContract<TreasuryContract>;
    let clientSender: Sender;
    let enclaveKeyPair: KeyPair;
    let validRandomHashPayload: RandomHash;
    let validRandomValuePayload: RandomValue;

    const now = Math.floor(Date.now() / 1000);
    const attestationReport = "test".repeat(257);
    const validNonce = BigInt(0xabcd);

    const stateInit = 0;
    const stateRock = 1;
    const stateRoll = 2;
    const stateWaitReveal = 3;
    const stateCompleted = 15;

    beforeEach(async () => {
        enclaveKeyPair = keyPairFromSecretKey(
            Buffer.from(
                "c8c24d89465fde431e12443ed2be7bf8ebbc0c47ce2a6342fc108df5cd937cf7393c80a2c5e690f2f955ffbf4e0112ed8513324ce5c4543535ebfd4963704fd5",
                "hex"
            )
        );

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

        client = await blockchain.treasury("client");
        clientSender = client.getSender();

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

        validRandomValuePayload = {
            $$type: "RandomValue",
            doraId: BigInt(0xaaaaaa),
            name: "Test project",
            revealTimestamp: BigInt(now),
            nonce: validNonce,
            txHash: beginCell().endCell(),
        };

        let revealed: RevealedValue = {
            $$type: "RevealedValue",
            timestamp: BigInt(now),
            recipient: contract.address,
            nonce: validNonce,
            doraId: BigInt(0xaaaaaa),
            name: "Test project",
        }

        let hash = beginCell().store(storeRevealedValue(revealed)).endCell().hash();
        validRandomHashPayload = {
            $$type: "RandomHash",
            timestamp: BigInt(now),
            recipient: contract.address,
            valueHash: beginCell().storeBuffer(hash).endCell(),
        };

    });


    async function sendUpdate(payload: RandomHash, expectExitCode: number = 0) {
        let hash = beginCell().store(storeRandomHash(payload)).endCell().hash();
        let signature = sign(hash, enclaveKeyPair.secretKey);

        console.log(
            "Private key:",
            enclaveKeyPair.secretKey.toString("base64"),
            "\nPublic key:",
            enclaveKeyPair.publicKey.toString("hex"),
            "\nPayload:",
            beginCell().store(storeRandomHash(payload)).endCell().toBoc().toString("base64"),
            "\nHash:",
            hash.toString("base64"),
            "\nSignature:",
            signature.toString("base64"),
        );

        let res = await contract.send(
            sender,
            { value: toNano("0.5") },
            {
                $$type: "UpdateCommit",
                signature: beginCell().storeBuffer(signature).endCell(),
                payload: payload,
            }
        );
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: expectExitCode ? false: true,
            exitCode: expectExitCode,
            op: findOp(contract, "Update"),
        });
        return res;
    }

    async function sendUpdateCommit(payload: RandomHash, expectExitCode: number = 0) {
        let hash = beginCell().store(storeRandomHash(payload)).endCell().hash();
        let signature = sign(hash, enclaveKeyPair.secretKey);

        console.log(
            "Private key:",
            enclaveKeyPair.secretKey.toString("base64"),
            "\nPublic key:",
            enclaveKeyPair.publicKey.toString("hex"),
            "\nPayload:",
            beginCell().store(storeRandomHash(payload)).endCell().toBoc().toString("base64"),
            "\nHash:",
            hash.toString("base64"),
            "\nSignature:",
            signature.toString("base64"),
        );

        let res = await contract.send(
            sender,
            { value: toNano("0.5") },
            {
                $$type: "UpdateCommit",
                signature: beginCell().storeBuffer(signature).endCell(),
                payload: payload,
            }
        );
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: expectExitCode ? false: true,
            exitCode: expectExitCode,
            op: findOp(contract, "UpdateCommit"),
        });
        return res;
    }

    async function sendUpdateReveal(payload: RandomValue, expectExitCode: number = 0) {
        let hash = beginCell().store(storeRandomValue(payload)).endCell().hash();
        let signature = sign(hash, enclaveKeyPair.secretKey);

        let res = await contract.send(
            sender,
            { value: toNano("0.5") },
            {
                $$type: "UpdateReveal",
                signature: beginCell().storeBuffer(signature).endCell(),
                payload: payload,
            }
        );
        expect(res.transactions).toHaveTransaction({
            from: owner.address,
            to: contract.address,
            success: expectExitCode ? false: true,
            exitCode: expectExitCode,
            op: findOp(contract, "UpdateReveal"),
        });
        return res;
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

    it("should have name and version", async () => {
        const state = await contract.getState();
        expect(state.name).toEqual("get-random-winner");
        expect(state.version).toBeGreaterThanOrEqual(16842752);
    });

    it("should have correctinitial state", async () => {
        let estate = await contract.getEventState();
        expect(estate.stakeOnRock).toEqual(BigInt(0));
        expect(estate.stakeOnRoll).toEqual(BigInt(0));
        expect(estate.state).toEqual(BigInt(stateInit));
    });


    it("should call random on valid roll", async () => {
        let res = await contract.send(clientSender, { value: toNano("0.5") }, "roll");
        //prettyLogTransactions(res.transactions);
        expect(res.transactions).toHaveTransaction({
            from: contract.address,
            to: client.address,
            success: true,
            body: beginCell().storeUint(0,32).storeStringTail("⚄").endCell()
        });

        let estate = await contract.getEventState();
        expect(estate.stakeOnRock).toEqual(toNano("0"));
        expect(estate.stakeOnRoll).toEqual(toNano("0.5"));
        expect(estate.state).toEqual(BigInt(stateRoll));
        expect(estate.changed).toEqual(BigInt(stateRoll));

        let externals = res.externals;
        expect(externals.length).toBe(1)
        expect(externals[0].body).toEqualCell(beginCell().storeUint(0,32).storeStringTail("random()").endCell())
    });

    it("should handle random commit", async () => {
        let res = await contract.send(clientSender, { value: toNano("0.5") }, "roll");
        let estate = await contract.getEventState();
        expect(estate.changed).toEqual(BigInt(stateRoll));
    });

    it("should handle random reveal", async() => {
        let res = await contract.send(clientSender, { value: toNano("0.5") }, "roll");
        let estate = await contract.getEventState();
        expect(estate.changed).toEqual(BigInt(stateRoll));

        await sendUpdateCommit(validRandomHashPayload);
        await sendUpdateReveal(validRandomValuePayload);
        estate = await contract.getEventState();
        expect(estate.changed).toEqual(BigInt(stateRock));
    });

});
