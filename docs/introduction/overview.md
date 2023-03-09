# Overview

## What is the Greenfield Blockchain

There are four important roles in the Greenfield ecosystem:

1. **Greenfield Blockchain**. It is the core of Greenfield and is built on the Cosmos/Tendermint framework. The Greenfield 
  blockchain contains two categories of states "on-chain":
   - Accounts and their BNB balance ledger.
   - The metadata of the object storage system and SPs, the metadata of the objects stored on this storage system, and the 
      permission and billing information associated with this storage system.
   
   Greenfield blockchain transactions can change the above states. These states and the transactions comprise the major 
   economic data of BNB Greenfield.

2. **Storage Provider**. Storage Providers (SP) are storage service infrastructures that organizations or individuals provide 
  and the corresponding roles they play. They use Greenfield as the ledger and the single source of truth. Each SP can and 
  will respond to users' requests to write (upload) and read (download) data, and serve as the gatekeeper for user rights and 
  authentications.

3. **Cross-chain Facilities**. Greenfield and BSC network natively support cross-chain communication, which means that BNB 
  can freely circulate between the two chains, and data assets can also be mapped from Greenfield to BSC. The cross-chain facilities
  bring programmability to Greenfield.

4. **Dapps**. When users want to create and use the data on Greenfield, they may interact with the BNB Greenfield Core 
   Infrastructure via BNB Greenfield dApps (decentralized applications). Greenfield provides a friendly smart contract 
   library on the cross-chain facility, which can be easily integrated into dapps. 

Technically, BNB Greenfield blockchain uses Proof-of-Stake based on Tendermint-consensus for its own network security. 
The validator election and governance are based on a proposal-vote mechanism, which is revised based on Cosmos SDK's
governance module. Blocks are created every 2 seconds on the Greenfield chain by a group of validators.

BNB will be the gas and governance token on this blockchain. The initial BNB will be locked on BSC and re-minted on Greenfield.
BNB and data operation primitives can flow between Greenfield and BSC through cross-chain communication. Total circulation of 
BNB will stay unchanged as it is now but flow among BNB Beacon Chain, BSC, and Greenfield.


## Why Greenfield Blockchain

BNB Greenfield Blockchain is an infrastructure and ecosystem targeting to facilitate the decentralized data economy. 
It tries to achieve it by easing the process to store and manage data access and linking data ownership with the massive DeFi context of BSC.

The core of the greenfield blockchain revolves around data and includes several aspects:

1. **Decentralized storage of data**. Anyone or organization can register as a storage provider on the greenfield blockchain. 
As a ledger, the greenfield blockchain records the distribution of data stored on multiple SPs and coordinates data backup and recovery.
2. **Data ownership**. Greenfield Blockchain can provide users with powerful data ownership management functions. 
Users can grant access, modify, create, delete, and even execute permissions to their data to other individuals or groups.
On the greenfield blockchain, users truly achieve complete control over their own data.
3. **Data assetization**. All data and access permissions on Greenfield can be mapped into NFT assets on the BSC network. 
Users' operations on these NFT assets, such as minting or burning, will ultimately be converted into changes in their data 
permissions on the greenfield blockchain through cross-chain technology. These NFT assets will follow the ERC721 and ERC1155 standards, 
maximizing the reuse of existing NFT browsers and markets on BSC.
4. **Programmability of data ownership**.
5. **Flow of data value**. Due to the assetization of data on BSC, the flow of data value is greatly promoted through various dapps.

## Getting Started with the Greenfield Blockchain
- Quick start with Greenfield Blockchain. (TODO)
- Learn more about the ecosystem players of Greenfield. (TODO)
- Learn more about the module design of Greenfield Blockchain. (TODO)
- Building dapp from scratch with Greenfield. (TODO)
- Becoming a player of Greenfield ecosystem. (TODO)