# Data Availability Challenge

Data availability means the data is correctly stored on storage providers, and can be correctly downloaded by users.
The challenge module is designed and used to detect whether a data segment/piece is correctly stored on a
specified storage provider. For this kind of challenges, there will be three steps:

1. Each validator asks the challenged SP for this data piece and the local manifest of the object, if the validator
   can't get the expected piece, the piece should be regarded as unavailable.
2. Compute the hash of the local manifest, and compare with the related global manifest that recorded in the metadata of
   the object, once they are different, the piece should be regarded as unavailable.
3. Compute the hash of the challenged data piece, compare it with the related data that recorded in the local manifest,
   once they are different, the piece should be regarded as unavailable.

The validator collects challenge signatures, if there are more than 2/3 validators voted the same result for the current
challenge, the validator will aggregate these signatures and assemble a data challenge attestation, submit this
attestation on-chain through a transaction, the first one submits the attestation can get reward, the later transaction
which submits the attestation won't pass the verification. Only the validators whose votes wrapped into the attestation
will be rewarded, so it may be that some validators voted, but their votes were not assembled, they won't get reward
for this data availability challenge.

## Workflow

The data availability challenge workflow is as below:

1. Anyone can submit a transaction to challenge data availability, the challenge information would be recorded on-chain
   temporarily, and also would be written into the typed event after the transaction has been executed.
2. By default, at the end block phase of each block, we will use a random algorithm to generate some data availability
   challenge typed events. All challenge information will be persisted in storage until it has been confirmed or
   expired.
3. The off-chain data availability detect module keeps tracking the on-chain challenge information, and then initiates
   an
   off-chain data availability detect.
4. The validator uses its BLS private key to sign a data challenge vote according to the result, the vote data should be
   the same for all validators to sign, it should include block header, data challenge information, and the challenge
   result.
5. The validator keeps collecting data challenge votes, aggregates the signatures, assembles data challenge attestation.
6. The validator submits the attestation when there are more than 2/3 validators that have reached an agreement.
7. The data challenge attestation transaction will be executed, verify the attestation, clean the data challenge
   storage,
   slash the malicious SP, distribute the reward, and set a cooling-off period for successful data challenges to avoid
   attacking.
8. The cooling-off period is set for the validator to regain, recover, or shift this piece of data, once the cooling off
   period time expires, this data availability can be challenged again, if this piece of data is still unavailable, the
   validator would be slashed again.
<div align="center"><img src="https://raw.githubusercontent.com/bnb-chain/greenfield-whitepaper/main/assets/19.2%20Data%20Availability%20Challenge.jpg"  height="80%" width="80%"></div>
<div align="center"><i>Data Availability Challenge Workflow</i></div>

## Create Challenge

There are two ways to trigger challenges.

### Submitted Challenges

Anyone can send `MsgSubmit` messages to trigger data availability challenges, if he/she finds that the data is not
available or incorrect stored. When submitting the challenge, user can choose the segment/piece of an object to
challenge or let the blockchain randomly selects a segment/piece to challenge.
The submitter will be called as challenger, and will be rewarded if the challenge
succeeds later.

### Random Challenges

In each block, challenges will be automatically created, to challenge different objects which are stored on different 
storage providers. The count of random challenges in each block is governed, and can be changed by submitting proposals.
To support randomness, a *RANDAO* mechanism is introduced in Greenfield blockchain. For more information about *RANDAO*,
please refer to the following section.

## Attest Challenge

Each validator will listen to the events of challenge creations, and vote the challenge by using its own BLS key.
When there are more than 2/3 votes are collected, an attestation message `MsgAttest` will be submitted to slash the 
challenged storage provider. And the voted validators, the attestation submitter, and the challenger (if there is) will 
be rewarded accordingly.


## Challenge Heartbeat

To indicate the off-chain challenge detect module is running correctly, validators have to vote and submit 
`MsgHeartbeat` messages periodically to the blockchain. During processing this kind of messages, the income for securing 
stored objects will be transferred from payment account to distribution account,
and income can be withdrawn by validators and their delegators later.

## Challenge Events

The following events are introduced for data availability challenge. For the detailed definition, please refer
to [this](https://github.com/bnb-chain/greenfield/blob/master/proto/greenfield/challenge/events.proto).

### Start Event

This kind of events indicates that a data availability challenge is triggered on-chain. The off-chain module should
monitor the events, asking the according storage prover for data, compute hashes and do the comparison, and submit
an attestation if needed.

### Complete Event

When an attestation is received and accepted, then this kind of events will be emitted. In the events, the slash
and rewards amounts are also recorded.

### Heartbeat Event

Heartbeat only includes the necessary information for liveness-check purpose. 

## RANDAO

To support random challenges, a RANDAO mechanism is introduced like the following.
Overall, the idea is very similar to the RANDAO in Ethereum beacon chain, you can refer to
[here](https://eth2book.info/altair/part2/building_blocks/randomness) for more information.

When proposing a new block, the proposer, i.e. a validator, needs to sign the current block number to get 
a `randao reveal`, and mixes the reveal into randao result `randao mix` by using `xor` operation. 
The other validators will verify the `randao reveal` and `randao mix` by following steps: 
1. The signature is verified using the proposer's public key. It means that the proposer has almost no choice 
about what it contributes to the RANDAO. It either contributes the correct signature over the block number, 
or it gives up the right for proposing the current block. If the validator does propose the current block, 
it still cannot predict the reveal from other validators, and even be slashed for stopping proposing blocks.
2. The `randao mix` is correctly updated by using `xor` operation.


The implementation is conducted in Tendermint layer - a new field called `randao_mix` is added into block header.
Greenfield blockchain then uses the field as a seed to randomly pick objects and storage providers to challenge 
in each block.