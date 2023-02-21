# Consensus and Staking

BNB Greenfield blockchain uses Proof-of-Stake based on Tendermint-consensus for its own network security. Blocks are
created every 2 seconds on the Greenfield chain by a group of validators. BNB will be the staking token on this blockchain.

There will be a validators set in the genesis state, the staked BNB for these validators would be locked in the BSC side.
For the other validators, they should submit a create-validator proposal to become a validator, and if the validator
doesn't behave well, they can be impeached, once the impeach-validator proposal is passed, the validator would be jailed
forever.

It should be mentioned that the validators are separated from the storage provider. The validators are responsible for
generating the blocks, challenging the data availability, and cross-chain communication. The storage providers are
responsible for storing the data objects, there is no strong binding relationship between them.

## Create Validator

To become a validator, a create-validator proposal should be submitted and adopted by the majority of the current validators.
Meanwhile, in the early stage, only self delegation is allowed, and in the further open delegation can be supported.
Here are the steps for becoming a new validator:

- Self delegator of the new validator grants the delegate authorization to the gov module account.

- The new validator should initiate a create-validator proposal.

- Wait for the current validators to vote on this proposal.

- Once the proposal is passed, the new validator would be created automatically.

## Edit Validator

There are several fields that validators can edit after they have been created. Including description, commission rate,
min-self-delegation, relayer address, and relayer bls public key. All these fields can be edited by submitting an
edit-validator transaction.

## Impeach Validator

If a validator behaves badly, anyone can submit a proposal to impeach the validator, and if the proposal is passed, the
validator would be jailed forever. Here are the steps for impeaching a validator:

- Initiate an impeach-validator proposal.

- Wait for the current validators to vote on this proposal.

- Once the proposal is passed, the validator would be jailed forever automatically,
  that means it can’t become an active validator anymore.

## Staking Reward Distribution

There is no inflation of tokens on Greenfield, all the block rewards come from the transaction fee. Block rewards will be
distributed passively, delegators can submit withdrawal requests to get all the up-to-date block rewards.