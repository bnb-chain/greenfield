# Governance

The Greenfield BlockChain utilizes on-chain governance, it is achieved by steps listed below:

- Proposal submission: Proposal is submitted to the blockchain with a deposit.
- Vote: Once deposit reaches threshold `MinDeposit`, proposal is confirmed and vote opens. Bonded BNB holders can vote on the proposal.
- Execution: After a period of time, the votes are tallied and depending on the result, the messages in the proposal will be executed

There are various types of proposals. Include but not limited to:
- create and edit validators, staking rewards distribution, details as described in [staking_model](./08-consensus_and_staking.md)
- create and remove storage provider which specified in [storage_provider_model](./04-storage_provider_management.md)
- `Greenfield` modules parameters governance
- `BSC` smart contract parameters and version upgrades governance

## Submit proposal:

In `Greenfield`, any account can submit proposal by sending a `MsgSubmitProposal` transaction

## Deposit:

Proposals must be submitted with a deposit in `BNB` defined by the `MinDeposit` param, the deposit is required as spam 
protection. Any BNB holder can contribute to this deposit to support proposal, the submitter does not need to provide 
the deposit itself. To transit a newly created proposal into active status, `MinDeposit` need to be met, otherwise it 
will stay inactive, once deposit end time comes, the proposal will be destroyed and all deposits will be burned. For 
proposals which deposit reaches minimum threshold, their status turn into active and `voting period` starts.

## Vote:

All bonded BNB holders get the right to vote on proposals with one of below options:

- Yes
- No
- NoWithVeto 
- Abstain

Be aware of that voting power is measured in terms of bonded BNB, and delegators inherit the vote of validators they are 
delegated to. If delegator votes after its validator, it will override its validator vote with its own.

## Tally

Following requirements need to be met for a proposal to be accepted:

- Quorum: More than 1/3 of total bonded BNB need to be participated in the vote
- Threshold: `Yes` votes should be more than to 1/2 excluding `Abstain` votes.
- Veto: `NoWithVeto` votes is less than 1/3, including Abstain votes.

If a proposal is approved or rejected without `NoWithVeto`, deposit will be refunded to depositor. In the case where
`NoWithVeto` exceed 1/3, deposits will be burned and proposal information will be removed.


## Proposal implementation

A `passed` proposal will be implemented 

