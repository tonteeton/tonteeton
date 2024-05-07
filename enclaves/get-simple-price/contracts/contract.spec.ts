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

describe("contract", () => {
    let blockchain: Blockchain;
    let contract: SandboxContract<OracleContract>;
    let client: SandboxContract<DemoContract>;
    let owner: SandboxContract<TreasuryContract>;
    let sender: Sender;
    let enclaveKeyPair: KeyPair;
    const validPayload: PriceUpdate = {
        $$type: "PriceUpdate",
        usd: BigInt(345),
        lastUpdatedAt: BigInt(1715092161),
    };

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
            await OracleContract.fromInit(BigInt("0x" + enclaveKeyPair.publicKey.toString("hex")), owner.address)
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

    it("should deploy correctly", async () => {});

    it("should respond to price request with a price response", async () => {
        let res = await client.send(sender, { value: toNano("0.0123") }, "callOracle");
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

    it("should handle price update from enclave", async () => {
        let payload = validPayload;
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
});
