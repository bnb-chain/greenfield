# Greenfield blockchain events

## Table of Content

- [1 event types](#event-types)
- [2 Staking](#staking)

Greenfield blockchain events follow the cosmos events format. Refer to
https://docs.cosmos.network/v0.46/core/events.html for details.

## event types

Greenfield blockchain includes the following event types:

| Type                         | Module   | Description                                         |
|------------------------------|----------|-----------------------------------------------------|
| tx                           | -        | Event for each transaction has been executed        |
| message                      | -        | Event for each message has been executed            |
| complete_unbonding           | staking  | Undelegation has completed its unbonding period     |
| complete_redelegation        | staking  | Redelegation has completed its unbonding period     |
| create_validator             | staking  | A new validator has been created                    |
| edit_validator               | staking  | Validator's fields has been updated                 |
| delegate                     | staking  | A new delegation from the delegator                 |
| unbond                       | staking  | A new undelegation from the delegator               |
| cancel_unbonding_delegation  | staking  | Cancel undelegation within the unbonding period     |
| redelegate                   | staking  | Delegator redelegates from one validator to another |

## Staking

The staking module emits the following events:

### EndBlocker

| Type                  | Attribute Key         | Attribute Value           | Value type |
| --------------------- | --------------------- | ------------------------- |------------|
| complete_unbonding    | amount                | {totalUnbondingAmount}    | Coins      |
| complete_unbonding    | validator             | {validatorAddress}        | string     |
| complete_unbonding    | delegator             | {delegatorAddress}        | string     |
| complete_redelegation | amount                | {totalRedelegationAmount} | Coins      |
| complete_redelegation | source_validator      | {srcValidatorAddress}     | string     |
| complete_redelegation | destination_validator | {dstValidatorAddress}     | string     |
| complete_redelegation | delegator             | {delegatorAddress}        | string     |

### Msg's

#### MsgCreateValidator

| Type             | Attribute Key   | Attribute Value    | Value type |
| ---------------- |-----------------|--------------------|------------|
| create_validator | validator       | {validatorAddress} | string     |
| create_validator | sel_del_address | {selfDelAddress}   | string     |
| create_validator | relayer_address | {relayerAddress}   | string     |
| create_validator | relayer_bls_key | {relayerBlsKey}    | string     |
| create_validator | amount          | {delegationAmount} | Coin       |
| message          | module          | staking            | string     |
| message          | sender          | {fromAddress}      | string     |

#### MsgEditValidator

| Type           | Attribute Key       | Attribute Value     | Value type |
| -------------- |---------------------|---------------------|------------|
| edit_validator | commission_rate     | {commissionRate}    | Commission |
| edit_validator | min_self_delegation | {minSelfDelegation} | BigInt     |
| edit_validator | relayer_address     | {relayerAddress}    | string     |
| edit_validator | relayer_bls_key     | {relayerBlsKey}     | string     |
| message        | module              | staking             | string     |
| message        | sender              | {senderAddress}     | string     |

#### MsgDelegate

| Type     | Attribute Key | Attribute Value    | Value type |
| -------- | ------------- | ------------------ |------------|
| delegate | validator     | {validatorAddress} | string     |
| delegate | amount        | {delegationAmount} | Coins      |
| delegate | new_shares    | {newShares}        | BigInt     |
| message  | module        | staking            | string     |
| message  | sender        | {senderAddress}    | string     |

#### MsgUndelegate

| Type    | Attribute Key       | Attribute Value    | Value type |
| ------- | ------------------- | ------------------ |------------|
| unbond  | validator           | {validatorAddress} | string     |
| unbond  | amount              | {unbondAmount}     | Coins      |
| unbond  | completion_time [0] | {completionTime}   | Time       |
| message | module              | staking            | string     |
| message | sender              | {senderAddress}    | string     |

* [0] Time is formatted in the RFC3339 standard

#### MsgCancelUnbondingDelegation

| Type                          | Attribute Key       | Attribute Value                     |  Value type  |
| ----------------------------- | ------------------  | ------------------------------------| -------------|
| cancel_unbonding_delegation   | validator           | {validatorAddress}                  | string       |
| cancel_unbonding_delegation   | delegator           | {delegatorAddress}                  | string       |
| cancel_unbonding_delegation   | amount              | {cancelUnbondingDelegationAmount}   | Coin         |
| cancel_unbonding_delegation   | creation_height     | {unbondingCreationHeight}           | int64        |

#### MsgBeginRedelegate

| Type       | Attribute Key         | Attribute Value       | Value type |
| ---------- | --------------------- | --------------------- |------------|
| redelegate | source_validator      | {srcValidatorAddress} | string     |
| redelegate | destination_validator | {dstValidatorAddress} | string     |
| redelegate | amount                | {unbondAmount}        | Coin       |
| redelegate | completion_time [0]   | {completionTime}      | Time       |
| message    | module                | staking               | string     |
| message    | sender                | {senderAddress}       | string     |

* [0] Time is formatted in the RFC3339 standard