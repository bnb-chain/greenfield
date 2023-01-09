## Setup Local Chain

1. Build binary
```bash
make build
```

2. Creates all the configuration files
```bash
# The argument <moniker> is the custom username of your node, it should be human-readable.
./build/bin/bfsd init <moniker> --chain-id inscription_9000-01 --staking-bond-denom bnb
```

All these configuration files are in ~/.bfsd by default, but you can overwrite the location of this folder by passing the --home flag.

The ~/.bfsd folder has the following structure:
```
.                                   # ~/.bfsd
  |- data                           # Contains the databases used by the node.
  |- config/
      |- app.toml                   # Application-related configuration file.
      |- config.toml                # Tendermint-related configuration file.
      |- genesis.json               # The genesis file.
      |- node_key.json              # Private key to use for node authentication in the p2p protocol.
      |- priv_validator_key.json    # Private key to use as a validator in the consensus protocol.

```

3. Adding keys to the keyring
```bash
# new key
./build/bin/bfsd keys add validator --keyring-backend test
./build/bin/bfsd keys add relayer --keyring-backend test --algo eth_bls

# list accounts
./build/bin/bfsd keys list --keyring-backend test
```

The keyring supports multiple storage backends, some of which may not be available on all operating systems.
See more details: https://docs.cosmos.network/v0.46/run-node/keyring.html#available-backends-for-the-keyring


4. Adding genesis accounts
Before starting the chain, you need to populate the state with at least one account.
```bash
VALIDATOR=$(./build/bin/bfsd keys show validator -a --keyring-backend test)
RELAYER=$(./build/bin/bfsd keys show relayer -a --keyring-backend test)
RELAYER_BLS=$(./build/bin/bfsd keys show relayer -a --keyring-backend test --output json | jq -r .pubkey_hex)"
./build/bin/bfsd add-genesis-account $VALIDATOR 100000000000bnb
```

5. Create validator in genesis state
```bash
# create a gentx.
./build/bin/bfsd gentx validator 1000000000bnb $VALIDATOR $RELAYER $RELAYER_BLS --keyring-backend=test --chain-id=inscription_9000-121 \
    --moniker="validator" \
    --commission-max-change-rate=0.01 \
    --commission-max-rate=1.0 \
    --commission-rate=0.07

# Add the gentx to the genesis file.
./build/bin/bfsd collect-gentxs
```

6. Run local node
```bash
./build/bin/bfsd start
```

## Quickly Setup a Local Cluster Network
1. Start
```bash
SIZE=3 # The number of nodes in the cluster.
bash ./deployment/localup/localup.sh all ${SIZE}
```

2. Stop
```bash
bash ./deployment/localup/localup.sh stop
```

3. Send Tx
```bash
./build/bin/bfsd tx bank send validator0 0x32Ff14Fa1547314b95991976DB432F9Aa648A423 500bnb --home ./deployment/localup/.local/validator0 --keyring-backend test --node http://localhost:26750 -b block
```

4. Restart the chain without state initialization
```bash
bash ./deployment/localup/localup.sh stop
bash ./deployment/localup/localup.sh start ${SIZE}
```

## Quickly Setup Fullnode
1. Start
```bash
SIZE=3 # The number of nodes in the cluster.
bash ./deployment/localup/localup_fullnode.sh all ${SIZE}
```

2. Stop
```bash
bash ./deployment/localup/localup_fullnode.sh stop
```

3. Restart the fullnodes without initialization
```bash
bash ./deployment/localup/localup_fullnode.sh stop
bash ./deployment/localup/localup_fullnode.sh start ${SIZE}
```