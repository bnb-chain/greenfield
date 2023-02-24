# Overview

BNB Greenfield blockchain uses Proof-of-Stake based on Tendermint-consensus for its own network security. Blocks are 
created every 2 seconds on the Greenfield chain by a group of validators.

BNB will be the gas and governance token on this blockchain. There is a native cross-chain bridge between the Greenfield
blockchain and BSC. The initial BNB will be locked on BSC and re-minted on Greenfield. BNB and data operation primitives 
can flow between Greenfield and BSC.

Total circulation of BNB will stay unchanged as it is now but flow among BNB Beacon Chain, BSC, and Greenfield.

The validator election and governance are based on a proposal-vote mechanism, which is revised based on Cosmos SDK's 
governance module: anyone can create and propose to become a validator, and the election into the active set will be 
based on the stake ranking (initially new validators may request the existing validator set's votes to be qualified for 
election). As validators will host all the critical metadata and respond to all data operation transactions, they should 
run professionally in terms of performance and stability.

To facilitate cross-chain operation and convenient asset management, the address format of the Greenfield blockchain 
will be fully compatible with BSC (and Ethereum). It also accepts EIP712 transaction signing and verification. These 
enable the existing wallet infrastructure to interact with Greenfield at the beginning naturally.

## Ecosystem Players
There are several player roles for the whole Greenfield ecosystem.

<div align="center"><img src="../asset/01-All%20Players%20of%20Greenfield.jpg"  height="80%" width="80%"></div>
<div align="center"><i>Figure All Players of Greenfield</i></div>

### Greenfield Validators

As a PoS blockchain, the Greenfield blockchain has its own validators.
These validators are elected based on the Proof-of-Stake logic.

These validators are responsible for the security of the Greenfield
blockchain. They get involved in the governance and staking of the
blockchain. They form a P2P network that is similar to other PoS
blockchain networks.

Meanwhile, they accept and process transactions to allow users to
operate their objects stored on Greenfield. They maintain the metadata
of Greenfield as the blockchain state, which is the control panel for
both Storage Providers (SPs) and users. These two parties use and update
these states to operate the object storage.

Greenfield validators also have the obligation to run the relayer system
for cross-chain communication with BSC.

The network topology of Greenfield validators is similar to the existing
secure validator setup of PoS blockchains. "Sentry Nodes" are used to
defend against DDoS and provide a secure private network, as shown in
the below diagram.

<div align="center"><img src="../asset/02-Greenfield%20Blockchain%20Network.jpg"  height="80%" width="80%"></div>
<div align="center"><i>Figure Greenfield Blockchain Network</i></div>

### Storage Providers (SPs)

Storage Providers are professional individuals and organizations who run
a series of services to provide data services based on the Greenfield
blockchain.

### Greenfield dApps

Greenfield dApps are applications that provide functions based on
Greenfield storage and its related economic traits to solve some
problems of their users.

## Greenfield Blockchain Data Storage
All Greenfield validators have such active data in full (at least the latest state). Anyone can join the blockchain as 
full nodes to synchronize these data for free.

These data are on-chain and can be only changed through transactions onto the Greenfield blockchain. It has several types 
as described below.

### Accounts and Balance
Each user has their "Owner Address" as the identifier for their owner account to "own" the data resources. There is 
another "payment account" type dedicated to billing and payment purposes and owned by owner addresses.

Both owner accounts and payment accounts can hold the BNB balance on Greenfield. Users can deposit BNB from BSC, accept
transfers from other users, and spend them on transaction gas and storage usage.

### Validator and SP Metadata
These are the basic information about the Greenfield validators and Greenfield SPs. SPs may have more information, as 
it has to publish their service information for users' data operations. There should be a reputation mechanism for SPs 
as well.

### Storage Metadata
The "storage metadata" includes size, ownership, checksum hashes, and distribution location among SPs. Similar to AWS S3, 
the basic unit of the storage is an "object", which can be a piece of binary data, text files, photos, videos, or any 
other format. Users can create their objects under their "bucket". A bucket is globally unique. The object can be referred 
to via the bucket name and the object ID. It can also be located by the bucket name, the prefix tag, and the object ID 
via off-chain facilitations.

### Permission Metadata
Data resources on Greenfield, such as the data objects and the buckets, all have access control, such as which address 
can create, read, list, or even execute the resources, and which address can grant/revoke these permissions.

Two other data resources also have access control. One is "Group". A group represents a group of user addresses that have
the same permissions to the same resources. It can be used in the same way as an address in the access control. Meanwhile, 
it requires permission too to change the group. The other is "payment account". They are created by the owner accounts.

Here the access control is enforced by the SPs off-chain. People can test and challenge the SPs if they mess up the 
control. Slash and reward will happen to keep the SPs sticking to the principles.

### Billing Metadata
Users have to pay fees to store data objects on Greenfield. While each object enjoys a free quota to download by users 
who are permitted to, the excessive download will require extra data packages to be paid for the bandwidth. Besides 
the owner address, users can derive multiple "Payment Addresses" to pay these fees. Objects are stored under buckets, 
while each bucket can be associated with these payment addresses, and the system will charge these accounts for storing
and/or downloading. Many buckets can share the same payment address. Such association information is also stored on 
chains with consensus as well.