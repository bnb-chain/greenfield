# Storage Provider Management

## Abstract

SPs play a different role from Greenfield validators, although the same organizations or individuals can run both SPs
and validators if they follow all the rules and procedures to get elected.

The SP module is responsible for managing and maintaining storage providers in the BNB Greenfield decentralized
storage network.

- Metadata: The basic information about the SP, including address, tokens, status, and so on.
- Deposit: Each SP who wants to join the storage network should stake tokens to ensure that it can provide storage
  services normally
- Slash: The data stored on the SP will be challenged from time to time. When a challenge is successful, the SP will be
  slashed, and deducted some tokens.
- Reputation: We will introduce a reputation system to evaluate SP's service quality. Users can choose one SP to store
  data based on it's reputation score.
- Exit: A SP can leave voluntarily by following some specific rules and get back the staked tokens. At the same time,
  Greenfield can force it to exit when it has insufficient staked tokens or its reputation score is too low to meet
  basic requirements as one SP.

## Key Workflow

### Join the network

SPs have to register themselves first by depositing on the Greenfield blockchain as their "Service Stake". Greenfield
validators will go through a dedicated governance procedure to vote for the SPs of their election. SPs are encouraged to
advertise their information and prove to the community their capability, as SPs have to provide a professional storage
system with high-quality SLA.

It will take several transactions to join the greenfield storage network for storage provider.

1. The funding account of sp should grant the governance module account to deduct tokens for staking.
2. The SP submit a `CreateStorageProvider` proposal to governance module
3. Deposit enough tokens for this proposal
4. The validators should vote for this proposal. Pass or reject.
5. When enough validators vote the proposal, the storage provider will be automatically created on chain.

### Leave the network

When the SPs join and leave the network, they have to follow a series of actions to ensure data redundancy for the
users; otherwise, their "Service Stake" will be fined. This is achieved through the data availability challenge and
validator governance votes.

### Reputation

We'll introduce a reputation system for storage provider to evaluate the quality of service of SP.

## State

### StorageProvider

Storage Provider can be one of several statuses:

* `STATUS_IN_SERVICE`: The sp is in service. it can serve user's Create/Upload/Download request.
* `STATUS_IN_JAILED`: The sp has been slashed many times, and its deposit tokens is insufficient.
* `STATUS_GRACEFUL_EXITING`: The SP is exiting gracefully. All the object stored in it will be shifted to another sp.
* `STATUS_OUT_OF_SERVICE`: The SP is out of service. it can be a short-lived service unavailable. Users are unable
  to store or get payload data on it.

Storage providers metadata should be primarily stored and accessed by the `OperatorAddr`, an EIP712 account address
for the operator of the storage provider. Three additional indices are maintained per storage provider metadata in
order to fulfill required lookups for SealObject/Deposit/Slash/GetApproval.

* StorageProvider: `0x21 | OperatorAddr -> ProtocolBuffer(StorageProvider)`
* StorageProviderByFundingAddress: `0x22 | FundingAddress -> OperatorAddr`
* StorageProviderBySealAddress: `0x23 | SealAddress -> OperatorAddr`
* StorageProviderByApprovalAddress: `0x24 | ApprovalAddress -> OperatorAddr`

Each storage provider's state is stored in a `StorageProvider` struct.

```protobuf
message StorageProvider {
  option (gogoproto.equal) = false;
  option (gogoproto.goproto_stringer) = false;

  // operator_address defines the address of the sp's operator; It also is the unqiue index key of sp.
  string operator_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // fund_address define the account address of the storage provider for deposit, remuneration.
  string funding_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // seal_address define the account address of the storage provider for sealObject
  string seal_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // approval_address define the account address of the storage provider for ack CreateBuclet/Object.
  string approval_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // total_deposit define the deposit token
  string total_deposit = 5 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
  // status is the status of sp, which can be (in_service/read_only_service/graceful_exiting/out_of_service)
  Status status = 6;
  // endpoint is the service address of the storage provider
  string endpoint = 7;
  // description defines the description terms for the validator.
  Description description = 8 [(gogoproto.nullable) = false];
}
```

### Params

Params is a module-wide configuration structure that stores system parameters
and defines overall functioning of the sp module.

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.equal) = true;
  option (gogoproto.goproto_stringer) = false;

  // deposit_denom defines the staking coin denomination.
  string deposit_denom = 1;
  // min_deposit_amount defines the minimum deposit amount for storage providers.
  string min_deposit = 2 [
    (cosmos_proto.scalar) = "cosmos.Int",
    (gogoproto.customtype) = "github.com/cosmos/cosmos-sdk/types.Int",
    (gogoproto.nullable) = false
  ];
}
```

### Deposit Pool

Sp module use its module account to manage all the staking tokens which deposit by storage providers.

## Message

### MsgCreateStorageProvider

A storage provider is created using the `MsgCreateProvider` messages.

```protobuf
message MsgCreateStorageProvider {
  option (cosmos.msg.v1.signer) = "creator";

  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  Description description = 2 [(gogoproto.nullable) = false];
  string sp_address = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string funding_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string seal_address = 5 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string approval_address = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string endpoint = 7;
  cosmos.base.v1beta1.Coin deposit = 8 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* another storage provider with this operator/seal/funding/approval address is already registered.
* the initial deposit tokens are of a denom not specified as the deposit denom of sp module
* the deposit tokens is insufficient

### MsgEditStorageProvider

The metadata of a storage provider can be edited by using `MsgEditStorageProvider` messages.

```protobuf
// MsgEditStorageProvider defines a SDK message for editing an existed SP.
message MsgEditStorageProvider {
  option (cosmos.msg.v1.signer) = "sp_address";

  string sp_address = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  string endpoint = 2;
  Description description = 3 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* the storage provider is not existed
* the description fields are too large

### MsgDeposit

When the deposit tokens of a storage provider are insufficient, it can use `MsgDeposit` messages to resupply the
deposit tokens.

```protobuf
// MsgDeposit defines a SDK message to deposit token for SP.
message MsgDeposit {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the msg signer, it should be sp address or sp's fund address
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // sp_address is the operator address of sp
  string sp_address = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // deposit is a mount of token which used to deposit for SP
  cosmos.base.v1beta1.Coin deposit = 3 [(gogoproto.nullable) = false];
}
```

This message is expected to fail if:

* the storage provider is not existed
* the deposit tokens are of a denom not specified as the deposit denom of sp module 

