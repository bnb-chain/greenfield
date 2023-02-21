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

In BNB Greenfield blockchain, rewards gained from transaction fees are paid to validators. The fee distribution
module fairly distributes the rewards to the validators' constituent delegators.

Rewards are calculated per period. The period is updated each time a validator's delegation changes, for example, when
the validator receives a new delegation. The rewards for a single validator can then be calculated by taking the total
rewards for the period before the delegation started, minus the current total rewards.

The commission to the validator is paid when the validator is removed or when the validator requests a withdrawal.
The commission is calculated and incremented at every `BeginBlock` operation to update accumulated fee amounts.

The rewards to a delegator are distributed when the delegation is changed or removed, or a withdrawal is requested.
Before rewards are distributed, all slashes to the validator that occurred during the current delegation are applied.