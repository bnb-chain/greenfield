# Using gnfd command to interact with bank module

## QuickStart

Start a local chain with 1 validator and 1 Storage provider
```shell
$ bash ./deployment/localup/localup.sh all 1 1
```

Export the accounts private key of sp

```shell
$ bash ./deployment/localup/localup.sh export_sp_privkey 1 1
```

## Create Storage Provider

### Prepare 4 account addresses in advance

Each storage provide will hold 4 account which for different uses.

* OperatorAddress: For edit the information of the StorageProvider
* FundingAddress: For deposit staking tokens and receive earnings
* SealAddress: For seal user's object
* ApprovalAddress: For approve user's requests.

### Authorize the Gov Module Account to debit tokens from the Funding account of SP

```shell
# Gov module account is 0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2 by default
./build/bin/gnfd tx sp grant 0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2 --spend-limit 1000000bnb --SPAddress 0x78FeF615b06251ecfA9Ba01B7DB2BFA892722dDC --from sp0_fund --home ./deployment/localup/.local/sp0 --keyring-backend test --node http://localhost:26750
```

* submit-proposal
```shell
./build/bin/gnfd tx gov submit-proposal ./deployment/localup/create_sp.json --from sp0 --keyring-backend test --home ./deployment/localup/.local/sp0  --node http://localhost:26750

# create_sp.json
./create_sp.json
{
  "messages":[
  {
    "@type":"/bnbchain.greenfield.sp.MsgCreateStorageProvider",
    "description":{
      "moniker": "sp0",
      "identity":"",
      "website":"",
      "security_contact":"",
      "details":""
    },
    "sp_address":"0x78FeF615b06251ecfA9Ba01B7DB2BFA892722dDC",
    "funding_address":"0x1d05CCD43A6c27fBCdfE6Ac727B0e9B889AAbC3B",
    "seal_address": "0x78FeF615b06251ecfA9Ba01B7DB2BFA892722dDC",
    "approval_address": "0x78FeF615b06251ecfA9Ba01B7DB2BFA892722dDC",
    "endpoint": "sp0.greenfield.io",
    "deposit":{
      "denom":"bnb",
      "amount":"10000"
    },
    "creator":"0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2"
  }
],
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "1bnb"
}
```

* deposit tokens to the proposal

```shell
./build/bin/gnfd tx gov deposit 1 10000bnb --from sp0 --keyring-backend test --home ./deployment/localup/.local/sp0  --node http://localhost:26750
```

* voted by validator 

```shell
./build/bin/gnfd tx gov deposit 1 10000bnb --from sp0 --keyring-backend test --home ./deployment/localup/.local/sp0  --node http://localhost:26750
```

* wait a VOTE_PERIOD(300s), and the check the status of sp

```shell
./build/bin/gnfd query sp storage-providers --node http://localhost:26750
```


## Deposit 

```shell
gnfd tx sp deposit [sp-address] [funding-address] [value] [flags]
```


## EditStorageProvider

```shell
gnfd tx sp edit-storage-provider [sp-address] [flags]
```