# Billing and Payment

> The associated white-paper section locates at <https://github.com/bnb-chain/greenfield-whitepaper/blob/main/part3.md#21-billing-and-payment>.

Greenfield will charge the users in two parts. Firstly, every
transaction will require gas fees to pay the Greenfield validator to
write the metadata on-chain. Secondly, the SPs charge the users for
their storage service. Such payment also happens on the Greenfield. This
document is about the latter: how such off-chain service fees are billed
and charged.

There are two kinds of fees for the off-chain service: object storage
fee and read fee:

1. Every object stored on Greenfield is charged a fee based on its
   size. The storage price is determined by the service providers.

2. There is a free quota for users' objects to be read based on their
   size, content types, and more. If exceeded, i.e. the object data
   has been downloaded too many times, SP will limit the bandwidth
   for more downloads. Users can raise their read quota to
   get more download quota. The read quota price is determined by the
   Primary Storage Provider users selected.

The fees are paid on Greenfield in the style of
"Stream" from users to the SPs at a constant rate. The fees are charged
every second as they are used.

## Concepts and Formulas

Streaming is a constant by-the-second movement of funds from a sending
account to a receiving account. Instead of sending payment transactions
now and then, Greenfield records the static balance, the latest update
timestamp, and flow rate in its payment module, and calculates the
dynamic balance with a formula using this data in the records. If the
net flow rate is not zero, the dynamic balance will change as time
elapses.

### Terminology

- **Payment Module:** This is a special module and ledger system
  managing the billing and payment on the Greenfield blockchain.
  Funds will be deposited or charged into it from users' balance or
  payment accounts in this Payment Module.

- **Stream Account**: The Payment Module has its own ledger for
  billing management. Users can deposit and withdraw funds into
  their corresponding "accounts" on the payment module ledger. These
  accounts are called "Stream Account", which will directly interact
  with the stream payment functions and bill the users for the
  storage and data service.

- **Payment Account**: Payment account has been discussed in other
  sections of Part 1 and Part 3 already. It is actually a type of
  Stream Account. Users can create different payment accounts and
  associate it with different buckets for different purposes.

- **Payment Stream:** Whenever the users add/delete/change a storage
  object or download a data package, they add or change one or more
  "payment streams" for that service provided by SPs.

- **Flow Rate**: The per-second rate at which a sender decreases its
  net outflow stream and increases a receiver's net inflow stream
  when creating or updating a payment stream.

- **Netflow Rate**: The per-second rate that an account's balance is
  changing. It is the sum of the account's inbound and outbound
  flow rates.

- **Sender**: The stream account that starts the payment stream by
  specifying a receiver and a flow rate at which its net flow
  decreases.

- **Receiver**: The stream account on the receiving end of one or more payment streams.

- **CRUD Timestamp**: The timestamp of when the user creates, updates,
  or deletes a payment stream.

- **Delta Balance**: The amount of the stream account's balance has
  changed since the latest CRUD timestamp due to the flow. It can be
  positive or negative.

- **Static Balance**: The balance of the stream account at the latest
  CRUD timestamp.

- **Dynamic Balance**: The actual balance of a stream account. It is
  the sum of the Static Balance and the Delta Balance.

When a user's stream account is initiated in the payment module by
deposit, several fields will be recorded for this stream account:

- CRUD Timestamp - will be the timestamp at the time.

- Static Balance - will be the deposit amount.

- Netflow Rate - will be 0.

- Buffer Balance - will be 0.

### Formula

*Static Balance = Initial Balance at latest CRUD timestamp*

*Delta Balance = Netflow Rate \* Seconds elapsed since latest CRUD timestamp*

*Current Balance = Static Balance + Delta Balance*

*Buffer Balance = - Netflow Rate \* pre-configed ReserveTime if Netflow Rate is negative*

<div align="center"><img src="https://github.com/bnb-chain/greenfield-whitepaper/blob/main/assets/21.1%20How%20a%20User%20Receives%20Inflow%20and%20Outflow%20of%20Funds.jpg"  height="60%" width="60%"></div>
<div align="center"><i>Figure 21.1 How a User Receives Inflow and Outflow of Funds</i></div>

Every time a user creates, updates, or deletes a payment stream, many
variables in the above formulas will be updated and the users' stream accounts in the
payment module will be settled.

- The net flow for the user's stream account in the payment module
  will be re-calculated to net all of its inbound and outbound flow
  rates by adding/removing the new payment stream against the
  current NetFlow Rate. E.g. when a user creates a new object whose
  price is \$0.4/month, its NetFlow Rate will add -\$0.4/month.

- If the NetFlow rate is negative, the associated amount of BNB will
  be reserved in a buffer balance. It is used to avoid the dynamic
  balance becoming negative. When the dynamic balance becomes under
  the threshold, the account will be forced settled.

- CRUD Timestamp will become the current timestamp.

- Static Balance will be re-calculated. The previous dynamic balance
  will be settled. The new static Balance will be the Current
  Balance plus the change of the Buffer Balance.

## Key Workflow

### Deposit and Withdrawal

All users(including SPs) can deposit and withdraw BNB in the payment
module. The StaticBalance in the StreamPayment data struct will be
"settled" first: the CRUDTimeStamp will be updated and StaticBalance
will be netted with DeltaBalance. Then the deposit and withdrawal number
will try to add/reduce the StaticBalance in the record. If the static
balance is less than the withdrawal amount, the withdrawal will fail.

Deposit/withdrawal via cross-chain will also be supported to enable
users to deposit/withdraw from BSC directly.

Specifically, the payment deposit can be triggered automatically during
the creation of objects or the addition of data packages. As long as
users have assets in their address accounts and payment accounts, the
payment module may directly charge the users by moving the funds from
address accounts.

### Payment Stream

Payment streams are flowing in one direction. Whenever the users deposit
from their address accounts into the stream accounts (including users'
default stream account and created payment accounts), the funds first go
from the users' address accounts to a system account maintained by the
Payment Module, although the fund size and other payment parameters will
be recorded on the users' stream account, i.e. the StreamPayment record,
in the Payment Module ledger. When the payment is settled, the funds
will go from the system account to SPs' address accounts according to
their in-flow calculation.

Every time users do the actions below, their StreamRecord will
be updated:

- Creating an object will create new streams to the SPs

- Deleting an object will delete associated streams to the SPs

- Adjusting the read quota will
  create/delete/update the associated streams

### Forced Settlement

If a user doesn't deposit for a long time, his previous deposit may be
all used up for the stored objects. Greenfield has a forced settlement
mechanism to ensure enough funds are secured for further service fees.

There are two configurations, ReserveTime and ForcedSettleTime. Let's say
7 days and 1 day. If a user wants to store an object at the price of
approximately $0.1 per month($0.00000004/second), he must reserve fees
for 7 days in buffer balance, which is `$0.00000004 * 7 * 86400 =
$0.024192`. If the user deposits $1 initially, the stream payment
record will be as below.

- CRUD Timestamp: 100

- Static Balance: $0.975808

- Netflow Rate: -$0.00000004/sec

- Buffer Balance: $0.024192

After 10000 seconds, the dynamic balance of the user is `0.975808 - 10000 * 0.00000004 = 0.975408`.

After 24395200 seconds(approximately 282 days), the dynamic balance of
the user will become negative. Users should have some alarms for such
events that remind them to supply more funds in time.

If no more funds are supplied and the dynamic balance plus buffer
balance is under the forced settlement threshold, the account will be
forcibly settled. All payment streams of the account will be closed and
the account will be marked as out of balance. The download speed for all
objects associated with the account or payment account will be
downgraded. The objects will be deleted by the SPs if no fund is
provided within the predefined threshold.

The forced settlement will also charge some fees which is another
incentive for users to replenish funds proactively.

Let's say the ForceSettlementTime is 1 day. After 24913601
seconds(approximately 288 days), the dynamic balance becomes `0.975808 -
24913601 *0.00000004 = -0.02073604`, plus the buffer balance is
$0.00345596. The forced settlement threshold is `86400* 0.00000004 =
0.003456`. The forced settlement will be triggered, and the record of the
user will become as below:

- CRUD Timestamp: 24913701

- Static Balance: \$0

- Netflow Rate: \$0/sec

- Buffer Balance: \$0

The validators will get the remaining $0.00345596 as a reward. The
account will be marked as "frozen" and his objects get downgraded
by SPs.

Every time a stream record is updated, Greenfield calculates the time
when the account will be out of balance. So Greenfield can keep an
on-chain list to trace the timestamps for the potential forced
settlement. The list will be checked by the end of every block
processing. When the block time passes the forced settlement timestamp,
the settlement of the associated stream accounts will be triggered
automatically.

### Payment Account

Payment account is a special "Stream Account" type in the Payment
Module. Users can create multiple payment accounts and have the
permission to link buckets to different payment accounts to pay for
storage and data package.

The payment accounts have the below traits:

- Every user can create multiple payment accounts. The payment
  accounts created by the user are recorded with a map on the
  Greenfield blockchain.

- The address format of the payment account is the same as normal
  accounts. It's derived by the hash of the user address and
  payment account index. The payment account only exists in the
  storage payment module. Users can deposit into, withdraw from and
  query the balance of payment accounts in the payment module, which
  means payment accounts cannot be used for transfer or staking.

- Users can only associate their buckets with their payment accounts
  to pay for storage and bandwidth. Users cannot associate their own
  buckets with others' payment accounts, and users cannot associate
  others' buckets with their own payment accounts.

- The owner of a payment account is the user who creates it. The owner
  can set it non-refundable. It's a one-way setting and can not be
  revoked. Thus, users can set some objects as "public goods" which
  can receive donations for storage and bandwidth while preserving
  the ownership.

### Account Freeze and Resume

If a payment account is out of balance, it will be settled and set a
flag as out of balance.

The NetflowRate will be set to 0, while the current settings of the
stream pay will be backed up in another data structure. The download
speed for all objects under this account will be downgraded.

If someone deposits into a frozen account and the static balance is
enough for reserved fees, the account will be resumed automatically. The
stream pay will be recovered from the backup.

During the OutOfBalance period, no objects can be associated with this
payment account again, which results in no object can be created under
the bucket associated with the account.

If the associated object is deleted, the backup stream pay settings will
be changed accordingly to ensure it reflects the latest once the account
is resumed.

### Storage Fee Price and Adjustment

The storage fee prices are determined by the SPs who supply the storage service.
The cost of the SPs are composed of 3 parts:
- The primary SP will store the whole object file.
- The secondary SPs will store part of the object file as a replica.
- The primary SP will supply all the download requests of the object.

There are 3 different on-chain prices.
- Primary SP Store Price
- Primary SP Read Price
- Secondary SP Store Price

Every SP can set their own store price and read price via on-chain transactions.
While the secondary SP store price is calculated by averaging all SPs' store price.

The unit of price is a decimal, which indicates wei BNB per byte per second.
E.g. the price is 0.027, means approximately $0.022 / GB / Month.
(`0.027 * (30 * 86400) * (1024 * 1024 * 1024) * 300 / 10 ** 18 ≈ 0.022`, assume the BNB price is 300 USD) 

The storage fees are calculated and charged in bucket level.
The store price and read price is up to the SP of bucket.
The secondary store price is stored in the chain state and the same for all buckets.
The total size of all objects and per secondary SP served size in a bucket will be recorded in the bucket metadata.
The charge size will be used instead of the real size, e.g. files under 1KB will be charged as 1KB to cover the cost. 
The payment bill will be calculated by the size statistics and prices, and it will be charged from
the stream account specified in the bucket to the SPs.

## State

The payment module keeps state of the following primary objects:
- The stream payment ledger.
- The payment accounts and total account created by users.
- An AutoSettleRecord list to keep track of the auto-settle timestamp of the stream accounts.

In addition, the payment module keeps the following indexes to manage the aforementioned state:

- StreamRecord Index. `address -> StreamRecord`
- PaymentAccount Index. `address -> PaymentAccount`
- PaymentAccountCount Index. `address -> PaymentAccountCount`
- AutoSettleRecord Index. `settle-timestamp | address -> 0`

```
message StreamRecord {
  string account = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  int64 crud_timestamp = 2;
  string netflow_rate = 3 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  string staticBalance = 4 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  string bufferBalance = 5 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  string lockBalance = 6 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  int32 status = 7;
  int64 settleTimestamp = 8;
  repeated OutFlow outFlows = 9 [(gogoproto.nullable) = false];
}

message PaymentAccount {
  string addr = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string owner = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  bool refundable = 3;
}

message PaymentAccountCount {
  string owner = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  uint64 count = 2;
}

message AutoSettleRecord {
  int64 timestamp = 1;
  string addr = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

## Parameters

The payment module contains the following parameters,
they can be updated with governance.

```
message Params {
  // Time duration which the buffer balance need to be reserved for NetOutFlow e.g. 6 month
  uint64 reserve_time = 1 [(gogoproto.moretags) = "yaml:\"reserve_time\""];
  // The maximum number of payment accounts that can be created by one user
  uint64 payment_account_count_limit = 2 [(gogoproto.moretags) = "yaml:\"payment_account_count_limit\""];
  // Time duration threshold of forced settlement.
  // If dynamic balance is less than NetOutFlowRate * forcedSettleTime, the account can be forced settled.
  uint64 forced_settle_time = 3 [(gogoproto.moretags) = "yaml:\"forced_settle_time\""];
  // the maximum number of accounts that will be forced settled in one block
  uint64 max_auto_force_settle_num = 4 [(gogoproto.moretags) = "yaml:\"max_auto_force_settle_num\""];
  // The denom of fee charged in payment module
  string fee_denom = 5 [(gogoproto.moretags) = "yaml:\"fee_denom\""];
}
```

|             Key             |  Type  |       Example       |
|:---------------------------:|:------:|:-------------------:|
|        reserve_time         | uint64 | 15552000 (180 days) |
| payment_account_count_limit | uint64 |         200         |
|     forced_settle_time      | uint64 |   604800 (7 days)   |
|  max_auto_force_settle_num  | uint64 |         100         |
|          fee_denom          | string |         BNB         |

## Keepers

The payment module keeper provides access to query the parameters,
payment account owner, storage price and several ways to update the ledger.

Currently, it's only used by the storage module.

```
type PaymentKeeper interface {
	GetParams(ctx sdk.Context) paymenttypes.Params
	IsPaymentAccountOwner(ctx sdk.Context, addr string, owner string) bool
	GetStoragePrice(ctx sdk.Context, params paymenttypes.StoragePriceParams) (price paymenttypes.StoragePrice, err error)
	ApplyFlowChanges(ctx sdk.Context, from string, flowChanges []paymenttypes.OutFlow) (err error)
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
}
```

## Messages

### MsgCreatePaymentAccount

Used to create new payment account for a user.

```
message MsgCreatePaymentAccount {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the address of the stream account that created the payment account
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgDeposit

Used to deposit to a stream account.

```
message MsgDeposit {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the message signer for MsgDeposit and the address of the account to deposit from
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // to is the address of the account to deposit to
  string to = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // amount is the amount to deposit
  string amount = 3 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
}
```

### MsgWithdraw

Used to withdraw from a stream account.

```
message MsgWithdraw {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the message signer for MsgWithdraw and the address of the receive account
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // from is the address of the account to withdraw from
  string from = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // amount is the amount to withdraw
  string amount = 3 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
}
```

### MsgDisableRefund

Used to make a stream account non-refundable.

```
message MsgDisableRefund {
  option (cosmos.msg.v1.signer) = "owner";

  // owner is the message signer for MsgDisableRefund and the address of the payment account owner
  string owner = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // addr is the address of the payment account to disable refund
  string addr = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```
