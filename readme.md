# Greenfield
Official Golang implementation of the Greenfield Blockchain. It uses [tendermint](https://github.com/tendermint/tendermint/) f
or consensus and build on [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

BNB Greenfield targets to facilitate the decentralized data economy. It tries to achieve it by easing the process to store 
and manage data access and linking data ownership with the massive DeFi context of BSC.

It focuses on 3 parts that differ from existing centralized and decentralized storage systems:
- It enables Ethereum-compatible addresses to create and manage both data and token assets.
- It natively links data permissions and management logic onto BSC as exchangeable assets and smart contract programs with all other assets.
- It provides developers with similar API primitives and performance as popular existing Web2 cloud storage.

Enjoys the journey!


## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without 
notification and not ready for production use. The code and security audit have not been fully completed and not ready 
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Greenfield Core

The center of BNB Greenfield are two layers:
1. A new storage-oriented blockchain, and
2. network composed of "storage providers".

This repo is the official implementation of Greenfield blockchain. 

The BNB Greenfield blockchain maintains the ledger for the users and the storage metadata as the common blockchain state data. 
It has BNB, transferred from BNB Smart Chain, as its native token for gas and governance. BNB Greenfield blockchain also has 
its own staking logic for governance.

The Greenfield blockchain contains two categories of states "on-chain":

1. Accounts and their BNB balance ledger

2. The metadata of the object storage system and SPs, the metadata of the objects stored on this storage system, and the
   permission and billing information associated with this storage system.

Greenfield blockchain transactions can change the above states. These
states and the transactions comprise the major economic data of BNB
Greenfield.

When users want to create and use the data on Greenfield, they may
interact with the BNB Greenfield Core Infrastructure via BNB Greenfield
dApps (decentralized applications).

## Quick Started

*Note*: Requires [Go 1.18+](https://go.dev/dl/)

```
## Build from source
make build

## start a private network with 3 validators
$ bash ./deployment/localup/localup.sh all  3

## query the key of the first validator
$ ./build/bin/gnfd keys list --home   $(pwd)/deployment/localup/.local/validator0   --keyring-backend test

## get the balance of an account
$ addr=`./build/bin/gnfd keys list --home   $(pwd)/deployment/localup/.local/validator0   --keyring-backend test|grep address|awk '{print $3}'`
$ ./build/bin/gnfd q bank balances $addr  --node tcp://127.0.0.1:26750 

## send BNB
$ ./build/bin/gnfd tx bank send validator0 0x73a4Cf67b46D7E4efbb95Fc6F59D64129299c2E3 100000000000000000000BNB --from validator0 -y --node tcp://127.0.0.1:26750 --home $(pwd)/deployment/localup/.local/validator0 -keyring-backend test  --broadcast-mode block

## create a storage bucket
$ ./build/bin/gnfd  tx storage create-bucket bucketname 0x73a4Cf67b46D7E4efbb95Fc6F59D64129299c2E3  --from validator0  -y  --node tcp://127.0.0.1:26750 --home  $(pwd)/deployment/localup/.local/validator0 --keyring-backend test --broadcast-mode  block

## stop the private chain
$ bash ./deployment/localup/localup.sh stop 

## restart the private chain
bash ./deployment/localup/localup.sh start  3
```

More advanced script and command line usage, please refer to the [Tutorial](docs/cli/cli.md).

## Key Modules

- `x/bridge`: provide the cross chain token transfer function. BNB can freely flow between Greenfield and BSC network with native support.
- `x/payment`: handle the billing and payment of storage module. The fees are paid on Greenfield in the style of "Stream" 
from users to the Storage Providers(SPs) at a constant rate. The fees are charged every second as they are used.
- `x/sp`: manage the storage provider. 
- `x/storage`: user can manage its storage data through this module, like create/delete bucket, create/delete storage object.

And the following modules are in cosmos-sdk:

- `x/crosschain`: manage the cross chain packages, like store/query/update the cross chain package, channels, sequences.
- `x/gashub`: provide a governable and predictable fee charge mechanism.
- `x/oracle`: provide a secure runtime for cross chain packages.
- `x/staking`:  based on the Proof-of-Stake logic. The elected validators are responsible for the security of the Greenfield blockchain. 
They get involved in the governance and staking of the blockchain.

Refer to the [docs](docs/readme.md) to dive deep into these modules.

## Join Testnet && Mainnet (coming soon..)

## Related Projects
- [Greenfield-Contract](https://github.com/bnb-chain/greenfield-contracts): the cross chain contract for Greenfield that deployed on BSC network.
- [Greenfield-Tendermint](https://github.com/bnb-chain/greenfield-tendermint): the consensus layer of Greenfield blockchain.
- [Greenfield-Storage-Provider](https://github.com/bnb-chain/greenfield-storage-provider): the storage service infrastructures provided by either organizations or individuals.
- [Greenfield-Relayer](https://github.com/bnb-chain/greenfield-relayer): the service that relay cross chain package to both chains.
- [Greenfield-Cmd](https://github.com/bnb-chain/greenfield-cmd): the most powerful command line to interact with Greenfield system.
- [Awesome Cosmos](https://github.com/cosmos/awesome-cosmos): Collection of Cosmos related resources which also fits Greenfield.

## Contribution
Thank you for considering to help out with the source code! We welcome contributions from anyone on the internet, and are 
grateful for even the smallest of fixes!

If you'd like to contribute to Greenfield, please fork, fix, commit and send a pull request for the maintainers to review and merge into the main code base.
If you wish to submit more complex changes though, please check up with the core devs first through github issue(going to have a discord channel soon) 
to ensure those changes are in line with the general philosophy of the project and/or get some early feedback which can make both your efforts much 
lighter as well as our review and merge procedures quick and simple.

## Licence (pending)
