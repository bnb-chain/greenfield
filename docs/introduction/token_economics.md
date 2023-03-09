# Token Economics

BNB remains the main utility token on Greenfield.

BNB can be transferred from BSC to Greenfield blockchain, and vice versa. It is used as:

- Staking token. It can be used to self-delegate and delegate as stake, which can earn gas rewards and may suffer slash for improper behaviors.
- Gas token. It can be used to pay the gas to submit transactions on the Greenfield blockchain, which includes Greenfield local transactions or 
  cross-chain transactions between Greenfield and BSC. This is charged at the time of transaction submission and dispatched to 
  Greenfield validators, and potentially Greenfield Storage Providers for some transactions. Fee distribution is done in-protocol and 
  a protocol specification is [described here](https://github.com/bnb-chain/greenfield-cosmos-sdk/blob/master/docs/spec/fee_distribution/f1_fee_distr.pdf).
- Storage service fee token. It can be used to pay fees for the object storage and download bandwidth data package. This is charged 
  as time goes on and dispatched to Greenfield Storage Providers.
- Governance token. BNB holders may govern the Greenfield by voting on proposals with their staked BNB (not day 1).

No initial donors, foundation, or company will get funds in the genesis setup.

## Genesis Setup
BNB is transferred from BSC to Greenfield as the first cross-chain action. The initial validator set and storage provider
of Greenfield at the genesis will first lock a certain amount of BNB into the "Greenfield Token Hub" contract on BSC. This contract 
is used as part of the native bridge for BNB transferring after the genesis. These initial locked BNB will be used as 
the self-stake of validators, the deposit of storage provider and early days gas fees.

The initial BNB allocation on greenfield is around 1M BNB. (TODO: not finalized)

## Circulation Model
There is no inflation of BNB in greenfield. Due to the dual chain structure, cross chain transfer is implemented to 
enable BNB flow between Greenfield and Smart Chain bi-directionally. The total circulation of BNB on Greenfield is volatile.

Greenfield use Lock/Unlock mechanism to ensure the total circulation of BNB on both chain is always less than the initial
total supply:
1. The transfer-out blockchain will lock the amount from source owner addresses into a module account or smart contract.
2. The transfer-in blockchain will unlock the amount from module account or contract and send it to target addresses.
3. Both networks will never mint BNB.

Refer to [cross chain model](../modules/cross-chain.md) to get more details about the mechanism.

## How to Participate in the Ecosystem
- [Become A Validator](../cli/validator-staking.md): validators secure the Greenfield by validating and relaying transactions, 
   proposing, verifying and finalizing blocks.
- [Become A Storage Provider](../cli/storage-provider.md): SPs store the objects' real data, i.e. the payload data. and get paid 
  by providing storage services.
- [Store/Manage Data](../cli/storage.md): store and manage your data in a decentralized way, control and own it all by yourself.
