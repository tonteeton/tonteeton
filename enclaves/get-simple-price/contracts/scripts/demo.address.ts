import { getDemoContractAddress, isTestnet } from "./utils";

(async () => {
       let address = await getDemoContractAddress();
       console.log(address.toString({ testOnly: isTestnet() }));
})();
