# Governance

The Greenfield BlockChain utilizes on-chain governance, which is achieved by steps listed below:

- Proposal submission: Proposal is submitted to the blockchain with a deposit.
- Vote: Once deposit reaches threshold `MinDeposit`, proposal is confirmed and vote opens. Bonded BNB holders can vote on the proposal.
- Execution: After voting period, the votes are tallied and if proposal `passed`, the messages in the proposal will be executed

There are various types of proposals. Include but not limited to:
- Proposals for creating and editing validators, staking rewards distribution, details as described in [staking_model](./08-consensus_and_staking.md)
- Proposals for creating and removing storage provider which specified in [storage_provider_model](./04-storage_provider_management.md)
- Parameters change proposal for `Greenfield` modules
- Parameters change and upgrade proposals for `BSC` smart contracts 


## Governance Parameters
Several of the numbers involved in governance are parameters and can thus be changed by passing a parameter change proposal.
- Minimum deposit: 1000000000000000000 BNB
- Maximum deposit period: 300s
- Voting period: 300s
- Quorum: 33.40% of participating voting power
- Pass threshold: 50% of participating voting power
- Veto threshold: 33.40% of participating voting power


## Submit proposal:

In `Greenfield`, any account can submit proposals by sending `MsgSubmitProposal` transaction

## Deposit:

Proposals must be submitted with a deposit in `BNB` defined by the `MinDeposit` param, the deposit is required as spam 
protection. Any BNB holder can contribute to this deposit to support proposal, the submitter is not mandatory to provide 
the deposit itself, thought it is usually filled by proposal maker. To transit a newly created proposal into active status, 
`MinDeposit` need to be met, otherwise proposal will stay inactive. Once deposit end time comes, the proposal will be 
destroyed and all deposits will be burned. For 
proposals which deposits reaches minimum threshold, status turn into active and `voting period` starts.

## Voting period:

All bonded BNB holders get the right to vote on proposals with one of following options:

- Yes: Approval of the proposal in its current form.
- No: Disapproval of the proposal in its current form.
- NoWithVeto: Which indicates a proposal either (1) spam (2) infringes on minority interests (3) violates rules of engagement
- Abstain: The voter wishes to contribute to quorum without voting for or against a proposal.

Voters may change their vote at any time before voting period ends, be aware of that voting power is measured in terms 
of bonded BNB, and delegators inherit the vote of validators they are delegated to. If delegator votes after its validator, 
it will override its validator vote by its own.

## Tally

Following requirements need to be met for a proposal to be accepted by the end of voting period:

- Quorum: A minimal of 33.40% of total bonded BNB(voting power) need to be participated in the vote
- Threshold: `Yes` votes should be more than 50% excluding `Abstain` votes.
- Veto: `NoWithVeto` votes is less than 33.40%, including Abstain votes.

If a proposal is approved or rejected without `NoWithVeto`, deposit will be refunded to depositor. In the case where
`NoWithVeto` exceed 33.40% , deposits will be burned and proposal information will be removed from state.

Validators not in the active set can cast a vote, but their voting power (including their delegators) 
will not count toward the vote if they are not in the active set when the voting period ends. That means that if BNB 
is delegated to a validator that is jailed, tombstone, or outside the active set at the time that the voting period 
ends, that BNB's stake-weight will not count in the vote.


