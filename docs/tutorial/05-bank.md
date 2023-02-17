# Using gnfd command to interact with bank module

## Abstract
The bank module is responsible for handling BNB transfers between
accounts and module accounts.

In addition, the bank module tracks and provides query support for the total
supply of BNB in the application.

## Quick Start

```
## Start a local cluster
$ bash ./deployment/localup/localup.sh all 3
$ alias gnfd="./build/bin/gnfd"
$ receiver=0x32Ff14Fa1547314b95991976DB432F9Aa648A423
## query the balance of receiver
$ gnfd q bank balances $receiver --node tcp://127.0.0.1:26750 
## send 500BNB to the receiver (note the decimal of BNB is 18)
$ gnfd tx bank send validator0 $receiver 500000000000000000000BNB --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
## query the balance of receiver again
$ gnfd q bank balances $receiver --node tcp://127.0.0.1:26750 
## try send some token that does not exit, error is expected.
$ gnfd tx bank send validator0 $receiver 500000000000000000000ETH --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y
## try multi send, send each 500BNB to both receiver and receiver2
$ receiver2=0x6d6247501b822fd4eaa76fcb64baea360279497f
$ gnfd tx bank multi-send validator0 $receiver $receiver2 500000000000000000000BNB --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block  -y --gas 500000
## query the metadata of BNB
$ gnfd q bank  denom-metadata --node tcp://127.0.0.1:26750 
## query the total supply of BNB
$ gnfd q bank  total    --denom BNB   --node tcp://127.0.0.1:26750 
```

## Detailed CLI

A user can query and interact with the `bank` module using the CLI.

### Query

The `query` commands allow users to query `bank` state.

```sh
gnfd query bank --help
```

#### balances

The `balances` command allows users to query account balances by address.

```sh
gnfd query bank balances [address] [flags]
```

Example:

```sh
gnfd query bank balances 0xabcd..
```

Example Output:

```yml
balances:
- amount: "10000000000000000000000"
  denom: BNB
pagination:
  next_key: null
  total: "0"
```

#### denom-metadata

The `denom-metadata` command allows users to query metadata for coin denominations. A user can query metadata for a single denomination using the `--denom` flag or all denominations without it.

```sh
gnfd query bank denom-metadata [flags]
```

Example:

```sh
gnfd query bank denom-metadata --denom BNB
```

Example Output:

```yml
metadata:
  base: BNB
  denom_units:
    - aliases:
        - wei
      denom: BNB
      exponent: 0
  description: The native staking token of the Greenfield.
  display: BNB
  name: ""
  symbol: ""
  uri: ""
  uri_hash: ""
```

#### total

The `total` command allows users to query the total supply of coins. A user can query the total supply for a single coin using the `--denom` flag or all coins without it.

```sh
gnfd query bank total [flags]
```

Example:

```sh
gnfd query bank total --denom BNB
```

Example Output:

```yml
amount: "1000000000000000800000000000"
denom: BNB
```

### Transactions

The `tx` commands allow users to interact with the `bank` module.

```sh
gnfd tx bank --help
```

#### send

The `send` command allows users to send funds from one account to another.

```sh
gnfd tx bank send [from_key_or_address] [to_address] [amount] [flags]
```

Example:

```sh
gnfd tx bank send addr1.. addr2.. 100000000000000000000BNB
```