# Data Availability Challenge

Data availability means the data is correctly stored on storage providers, and can be correctly downloaded by users.
The challenge module is designed and used to detect whether a data segment/piece is available in the
specified SP. For this kind of challenges, there will be three steps:

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

The data availability challenge mechanism workflow is as below:

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

## Messages

The following messages are introduced for data availability challenge. For the detailed definition, please refer
to [this](https://github.com/bnb-chain/greenfield/blob/develop/proto/greenfield/challenge/tx.proto).

### Submit Message

Anyone can submit this kind of messages to trigger data availability challenges, if he/she finds that the data is not
available or incorrect stored. The submitter will be called as challenger, and will be rewarded if the challenge
succeeds later.

### Attest Message

When there are more than 2/3 votes are collected, an attestation message will be submitted to slash the challenged
storage provider, and the voted validators, the attestation submitter, and the challenger (if there is) will be
rewarded accordingly.

### Heartbeat Message

Heartbeat message is submitted periodically to indicate the off-chain challenge detect module is running correctly.
Meanwhile, the income for securing stored objects will be transferred from payment account to distribution account,
and income can be withdrawn by validators and their delegators later.

## Events

The following events are introduced for data availability challenge. For the detailed definition, please refer
to [this](https://github.com/bnb-chain/greenfield/blob/develop/proto/greenfield/challenge/events.proto).

### Start Event

This kind of events indicates that a data availability challenge is triggered on-chain. The off-chain module should
monitor the events, asking the according storage prover for data, compute hashes and do the comparison, and submit
an attestation if needed.

### Complete Event

When an attestation is received and accepted, then this kind of events will be emitted. In the events, the slash
and rewards amounts are also recorded.

### Heartbeat Event

Heartbeat only includes the necessary information for liveness-check purpose. 