import { openOracleContract } from "./utils";

(async () => {
    const oracle = await openOracleContract();
    const methodName = getMethodName()
    const method = (oracle as any)[methodName];
    if (typeof method !== 'function') {
        throw new Error(`Method <${methodName}> does not exist on oracle.`);
    }
    console.log(await method());
})();


function getMethodName() {
    const args = process.argv.slice(2);
    if (args.length < 1) throw new Error("Get method name should be specified, like: 'Balance'.");
    const name = args[0];
    return "get" + name.charAt(0).toUpperCase() + name.slice(1);
}
