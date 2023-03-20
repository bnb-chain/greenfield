# Using gnfd command to interact with staking module

The staking module is responsible for handling validator and delegator related operations.

## Quick Start

```
## Start a local cluster
$ bash ./deployment/localup/localup.sh all 3
$ alias gnfd="./build/bin/gnfd"
$ val=0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
## query the the self delegation of a validator
$ gnfd q staking delegation $val $val --node tcp://127.0.0.1:26750
## query the validator info
$ gnfd q staking validator $val --node tcp://127.0.0.1:26750
## try self delegate 100BNB for validator0
$ gnfd tx staking delegate $val 100000000000000000000BNB --from validator0 --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
## try unbond 100BNB from validator0
$ gnfd tx staking unbond $val 100000000000000000000BNB --from validator0 --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
## query the unbonding-delegation
$ gnfd q staking unbonding-delegation $val $val --node tcp://127.0.0.1:26750
```

## Query

The CLI `query` commands allow users to query `staking` state.

```sh
gnfd query staking --help
```

### delegation

The `delegation` command allows users to query a delegation based on address and validator address.

```sh
gnfd query staking delegation [delegator-addr] [validator-addr] [flags]
```

Example:

```sh
gnfd query staking delegation 0x9fB29.. 0x91D7d..
```

Example Output:

```yml
balance:
  amount: "10000000000000000000000000"
  denom: BNB
delegation:
  delegator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
  shares: "10000000000000000000000000.000000000000000000"
  validator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
```

### historical-info

The `historical-info` command allows users to query historical info at given height.

```sh
gnfd query staking historical-info [height] [flags]
```

Example:

```sh
gnfd query staking historical-info 1
```

Example Output:

```yml
header:
  app_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  chain_id: greenfield_9000-121
  consensus_hash: BICRvH3cKD93v7+R1zxE2ljD34qcvIZ0Bdi389qtoi8=
  data_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  evidence_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  height: "1"
  last_block_id:
    hash: null
    part_set_header:
      hash: null
      total: 0
  last_commit_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  last_results_hash: 47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
  next_validators_hash: 66w0qjSBjRkSTh3S5gPNaCRR/E+RsbFhMNLHTNighoo=
  proposer_address: TY90UCbGrdAHcfcwgh2TSHKisL4=
  time: "2023-02-20T15:24:16.056235Z"
  validators_hash: 66w0qjSBjRkSTh3S5gPNaCRR/E+RsbFhMNLHTNighoo=
  version:
    app: "0"
    block: "11"
valset:
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: R/yEnaoqxfWcqns01j8w/U9PQ5DK2Vl6rkoJwh5CFn4=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator1
      identity: ""
      moniker: validator1
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0x3D1c0E4aEdA87e0779E895307Eb3dFF6432a3cDa
    relayer_address: 0xebd48ceed75B8e8174F6143394BD74fd2cf102d4
    challenger_address: 0xebd48ceed75B8e8174F6143394BD74fd2cf102d4
    bls_key: lBsWg+/1/3Q/mJFE97PDBkPSNF631enVSUin/u+DEeUmttGbArKHYyXdtvP20iOw
    self_del_address: 0x3D1c0E4aEdA87e0779E895307Eb3dFF6432a3cDa
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: gUtwmz/uuTGhoRhsZbwycyw9nfHyVpIdPMaNE4GMQ3Q=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator2
      identity: ""
      moniker: validator2
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0xe7374890fc6c8811Ab6eac3A30D9e5391C1077D5
    relayer_address: 0x3D4da133bDb0443917Cf29b06E3Bd18a9Cd5dc73
    challenger_address: 0x3D4da133bDb0443917Cf29b06E3Bd18a9Cd5dc73
    bls_key: t+wJJfvKxjgKDam3L+TgEWIaDpTZOOASwI0qC0lAgtQY4ow9RxJKjROFjxD9x9iR
    self_del_address: 0xe7374890fc6c8811Ab6eac3A30D9e5391C1077D5
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: 4hGnslsUEWptUbeeFXx/44eUodKoFv61IFimxPB7qbc=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator0
      identity: ""
      moniker: validator0
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
    relayer_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
    challenger_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
    bls_key: qt3VpcjybvsWdR8smO0eCoaD+4IkAczpl+mcBJOKnXQro2g3TUXWCZPU7VvI01KJ
    self_del_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
```

### params

The `params` command allows users to query the current staking parameters information.

```sh
gnfd query staking params [flags]
```

Example:

```sh
gnfd query staking params
```

Example Output:

```yml
bond_denom: BNB
historical_entries: 10000
max_entries: 7
max_validators: 100
min_commission_rate: "0.000000000000000000"
min_self_delegation: "1"
unbonding_time: 1814400s
```

### pool

The `pool` command allows users to query the current staking pool values.

```sh
gnfd query staking pool [flags]
```

Example:

```sh
gnfd query staking pool
```

Example Output:

```yml
bonded_tokens: "30000000000000000000000000"
not_bonded_tokens: "0"
```

### unbonding-delegation

The `unbonding-delegation` command allows users to query an unbonding-delegation record based on delegator and validator address.

```sh
gnfd query staking unbonding-delegation [delegator-addr] [validator-addr] [flags]
```

Example:

```sh
gnfd query staking unbonding-delegation 0x9fB29.. 0x91D7d..
```

Example Output:

```yml
delegator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
entries:
  - balance: "1000000000000000000"
    completion_time: "2023-03-14T01:46:43.524923Z"
    creation_height: "6736"
    initial_balance: "1000000000000000000"
validator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
```

### validator

The `validator` command allows users to query a validator.

```sh
gnfd query staking validator [validator-addr] [flags]
```

Example:

```sh
gnfd query staking validator 0x91D7d..
```

Example Output:

```yml
commission:
  commission_rates:
    max_change_rate: "0.010000000000000000"
    max_rate: "1.000000000000000000"
    rate: "0.070000000000000000"
  update_time: "2023-02-20T15:24:16.056235Z"
consensus_pubkey:
  '@type': /cosmos.crypto.ed25519.PubKey
  key: 4hGnslsUEWptUbeeFXx/44eUodKoFv61IFimxPB7qbc=
delegator_shares: "10000000000000000000000000.000000000000000000"
description:
  details: validator0
  identity: ""
  moniker: validator0
  security_contact: ""
  website: http://website
jailed: false
min_self_delegation: "1"
operator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
relayer_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
challenger_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
bls_key: qt3VpcjybvsWdR8smO0eCoaD+4IkAczpl+mcBJOKnXQro2g3TUXWCZPU7VvI01KJ
self_del_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
status: BOND_STATUS_BONDED
tokens: "10000000000000000000000000"
unbonding_height: "0"
unbonding_time: "1970-01-01T00:00:00Z"
```

### validators

The `validators` command allows users to query for all validators.

```sh
gnfd query staking validators [flags]
```

Example:

```sh
gnfd query staking validators
```

Example Output:

```yml
pagination:
  next_key: null
  total: "0"
validators:
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: R/yEnaoqxfWcqns01j8w/U9PQ5DK2Vl6rkoJwh5CFn4=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator1
      identity: ""
      moniker: validator1
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0x3D1c0E4aEdA87e0779E895307Eb3dFF6432a3cDa
    relayer_address: 0xebd48ceed75B8e8174F6143394BD74fd2cf102d4
    challenger_address: 0xebd48ceed75B8e8174F6143394BD74fd2cf102d4
    bls_key: lBsWg+/1/3Q/mJFE97PDBkPSNF631enVSUin/u+DEeUmttGbArKHYyXdtvP20iOw
    self_del_address: 0x3D1c0E4aEdA87e0779E895307Eb3dFF6432a3cDa
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: 4hGnslsUEWptUbeeFXx/44eUodKoFv61IFimxPB7qbc=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator0
      identity: ""
      moniker: validator0
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
    relayer_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
    challenger_address: 0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321
    bls_key: qt3VpcjybvsWdR8smO0eCoaD+4IkAczpl+mcBJOKnXQro2g3TUXWCZPU7VvI01KJ
    self_del_address: 0xCd6D1332a09c29A8a5Fe5Ea4b485F63881f26999
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
  - commission:
      commission_rates:
        max_change_rate: "0.010000000000000000"
        max_rate: "1.000000000000000000"
        rate: "0.070000000000000000"
      update_time: "2023-02-20T15:24:16.056235Z"
    consensus_pubkey:
      '@type': /cosmos.crypto.ed25519.PubKey
      key: gUtwmz/uuTGhoRhsZbwycyw9nfHyVpIdPMaNE4GMQ3Q=
    delegator_shares: "10000000000000000000000000.000000000000000000"
    description:
      details: validator2
      identity: ""
      moniker: validator2
      security_contact: ""
      website: http://website
    jailed: false
    min_self_delegation: "1"
    operator_address: 0xe7374890fc6c8811Ab6eac3A30D9e5391C1077D5
    relayer_address: 0x3D4da133bDb0443917Cf29b06E3Bd18a9Cd5dc73
    challenger_address: 0x3D4da133bDb0443917Cf29b06E3Bd18a9Cd5dc73
    bls_key: t+wJJfvKxjgKDam3L+TgEWIaDpTZOOASwI0qC0lAgtQY4ow9RxJKjROFjxD9x9iR
    self_del_address: 0xe7374890fc6c8811Ab6eac3A30D9e5391C1077D5
    status: BOND_STATUS_BONDED
    tokens: "10000000000000000000000000"
    unbonding_height: "0"
    unbonding_time: "1970-01-01T00:00:00Z"
```

## Transactions

The CLI `tx` commands allow users to send `staking` related transactions.

### delegate

The `delegate` command allows users to delegate liquid tokens to a validator, in the early stage, only self-delegate is
supported.

```sh
gnfd tx staking delegate [validator-addr] [amount] [flags]
```

Example:

```sh
gnfd tx staking delegate 0x91D7d.. 1000000000000000000000BNB --from mykey
```

### edit-validator

The `edit-validator` command allows the user to edit an existing validator account.

```sh
gnfd tx staking edit-validator [flags]
```

Example:

```sh
gnfd tx staking edit-validator --addr-relayer 0x91D7d.. --from mykey
```

### unbond

The `unbond` command allows the user to unbond shares from a validator.

```sh
gnfd tx staking unbond [validator-addr] [amount] [flags]
```

Example:

```sh
gnfd tx staking unbond 0x91D7d.. 100000000000000000000BNB --from mykey
```

## Steps for creating a new validator

To become a validator, a create-validator proposal should be submitted and adopted by the majority of the current validators.

### Grant delegate authorization

The self delegator account of the new validator should grant the delegate authorization to the gov module account, make
sure the spend limit should be no less than the value for creating validator, and the allowed validator should be the
operator address of the new validator. The gov module account of Greenfield is "0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2".

```sh
gnfd tx authz grant <grantee> delegate --sepend-limit <value> --allowed-validators <valAddr> [flags]
```

Example:

```sh
gnfd tx authz grant 0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2 delegate --sepend-limit 10000000000000000000000BNB --allowed-validators myvaladdr --from mykey
```

### Submit create-validator proposal

Use `gov submit-proposal` command to submit a create-validator proposal.

```sh
gnfd tx gov submit-proposal create-validator.json [flags]
```

Example:

The content of create-validator.json:

```json
{
  "messages":[
  {
    "@type":"/cosmos.staking.v1beta1.MsgCreateValidator",
    "description":{
      "moniker":"validator",
      "identity":"",
      "website":"",
      "security_contact":"",
      "details":""
    },
    "commission":{
      "rate":"0.100000000000000000",
      "max_rate":"0.200000000000000000",
      "max_change_rate":"0.010000000000000000"
    },
    "min_self_delegation":"1000000000000000000000",
    "delegator_address":"0x6D967dc83b625603c963713eABd5B43A281E595e",
    "validator_address":"0x6D967dc83b625603c963713eABd5B43A281E595e",
    "pubkey":{
      "@type":"/cosmos.crypto.ed25519.PubKey",
      "key":"POIf1u/xC0RoHhD5c5qWszVLnjuhSVlgyrhoIriSjf0="
    },
    "value":{
      "denom":"BNB",
      "amount":"10000000000000000000000"
    },
    "from":"0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2",
    "relayer_address":"0xBefD69cb71403DF7BA20310FB216e1Fb7bAC6321",
    "relayer_blskey":"ac1e598ae0ccbeeaafa31bc6faefa85c2ae3138699cac79169cd718f1a38445201454ec092a86f200e08a15266bdc600"
  }
],
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "1000000000000000000BNB"
}
```

```sh
gnfd tx gov submit-proposal create-validator.json --from mykey --gas auto
```

### Vote on the proposal

The current validators can use `gov vote` command to vote on the proposal.

```sh
gnfd tx gov vote [proposal-id] [option] [flags]
```

Example:

```sh
gnfd tx gov vote 1 yes --from mykey
```

### Tally and execute automatically

Once the proposal's voting period is over, the votes for the create-validator proposal would be tallied. If the proposal
is passed, the create-validator message in the proposal would be executed automatically.

## Steps for impeaching a malicious validator.

If a validator doesn't behave well, anyone can submit an impeach-validator proposal, if the proposal adopted by the
majority of the current validators, the malicious validator would be jailed forever.

### Submit an impeach-validator proposal

Use `gov submit-proposal` command to submit an impeach-validator proposal.

```sh
gnfd tx gov submit-proposal impeach-validator.json [flags]
```

Example:

The content of impeach-validator.json:

```json
{
  "messages":[
    {
      "@type":"/cosmos.slashing.v1beta1.MsgImpeach",
      "from":"0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2",
      "validator_address":"0x6D967dc83b625603c963713eABd5B43A281E595e"
    }
  ],
  "metadata": "4pIMOgIGx1vZGU=",
  "deposit": "1000000000000000000BNB"
}
```

```sh
gnfd tx gov submit-proposal impeach-validator.json --from mykey --gas auto
```

### Vote on the proposal

The current validators can use `gov vote` command to vote on the proposal.

```sh
gnfd tx gov vote [proposal-id] [option] [flags]
```

Example:

```sh
gnfd tx gov vote 2 yes --from mykey
```

### Tally and execute automatically

Once the proposal's voting period is over, the votes for the impeach-validator proposal would be tallied. If the proposal
is passed, the impeach-validator message in the proposal would be executed automatically.