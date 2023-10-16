## Full Example

1. Prerequisites 
Please follow the instructions in [this page](https://hyperledger-fabric.readthedocs.io/en/latest/install.html) and install all the tool

2. Clone the repo
cd into the test-network directory inside fabric-samples

3. Bring up test network, added the -ca flag to use the CA
   ```
   ./network.sh up createChannel -c mychannel -ca 
   ```

4. Deploy the smart contract basic 
   ```
   ./network.sh deployCC -ccn basic -ccp ../asset-transfer-basic/chaincode-go/ -ccl go
   ```

5. Run application
cd into asset-transfer-basic/application-gateway-go
   ```
   go run .
   ```

6. Get Transaction Detail
cd back to the test-network directory
run 
   ```
   ./network.sh getBlock -num <newest|oldest|config|(number)>
   ```
see the block detail in block folder

7. Clean up 
   ```
   ./network.sh down
   ```

## Option

   ```
   # To use cryptogen to create organization artifacts, delete the -ca flag
   ./network.sh up createChannel -c mychannel

   # To use user instead of admin to invoke the chaincode 
   change the 'Admin' in certPath and keyPath of the file asset-transfer-basic/application-gatway-go/assetTransfer.go to 'User1'
   ```

   *Note - only the admin will be able to transferAsset, updateAsset or createAsset*

## File

### Chaincode implentment 
--> asset-transfer-basic/chaincode-go/smartcontract.go
### Application
--> asset-transfer-basic/application-gateway-go
### Channel Config
--> test-network/configtx/configtx.yaml
### Register Peer
--> test-network/organizations/fabric-ca/registerEnroll.sh
### Main Function 
--> test-network/network.sh 





