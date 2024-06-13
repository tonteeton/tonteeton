import { getContractAddress, isTestnet } from "./utils";

(async () => {
       let address = await getContractAddress();
       console.log(address.toString({ testOnly: isTestnet() }));
})();
