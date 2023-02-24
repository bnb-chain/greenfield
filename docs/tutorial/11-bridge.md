# Using gnfd command to interact with token bridge module

## Abstract
The bridge module is responsible for handling the BNB transfers between greenfield and BSC.

Users can transfer BNB to BSC via gnfd command, and query the relayer fee for the cross-chain transfers.

## Quick Start

```
## Start a local cluster
$ bash ./deployment/localup/localup.sh all 3
$ alias gnfd="./build/bin/gnfd"
$ receiver=0x32Ff14Fa1547314b95991976DB432F9Aa648A423
## send 500BNB to the receiver (note the decimal of BNB is 18)
$ gnfd tx bridge transfer-out validator0 $receiver 500000000000000000000BNB --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
## query the relayer fees for crosschain transfers
$ gnfd q bridge params --node tcp://127.0.0.1:26750 
```

## Detailed CLI

A user can query and interact with the `bridge` module using the CLI.

### Query

The `query` commands allow users to query the params of the `bridge` module.

```sh
gnfd query bridge --help
```

#### params

The `params` command allows users to query the params of the `bridge` module.

```sh
gnfd query bridge params [flags]
```

Example:

```sh
gnfd query bridge params --node http://localhost:26750
```

Example Output:

```yml
params:
  transfer_out_ack_relayer_fee: "0"
  transfer_out_relayer_fee: "1"
```

### Transactions

The `tx` commands allow users to interact with the `bridge` module.

```sh
gnfd tx bridge --help
```

#### transfer-out

The `transfer-out` command allows users to send funds between accounts from greenfield to BSC.

```sh
gnfd tx bridge transfer-out [from_key_or_address] [to_address] [amount] [flags]
```

Example:

```sh
gnfd tx bridge transfer-out validator0 0x32Ff14Fa1547314b95991976DB432F9Aa648A423 500000000000000000000BNB --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
```