import { openDemoContract } from "./utils";

(async () => {
    const demo = await openDemoContract();
    const methodName = getMethodName()
    const method = (demo as any)[methodName];
    if (typeof method !== 'function') {
        throw new Error(`Method <${methodName}> does not exist on Demo contract.`);
    }
    console.log(await method());
})();


function getMethodName() {
    const args = process.argv.slice(2);
    if (args.length < 1) throw new Error("Get method name should be specified, like: 'Balance'.");
    const name = args[0];
    return "get" + name.charAt(0).toUpperCase() + name.slice(1);
}
