# Using gnfd command to manage keys

## Abstract

The keyring module could help managing keys. It holds the private/public 
keypairs used to interact with a node

## Quick Start

```
## Start a local cluster
$ bash ./deployment/localup/localup.sh all 3
$ alias gnfd="./build/bin/gnfd"
$ moniker=temp
## create a new key
$ gnfd keys add $moniker --keyring-backend test --keyring-dir ./deployment/localup/.local/validator0
## show info of the key
$ gnfd keys show $moniker --keyring-backend test --keyring-dir ./deployment/localup/.local/validator0
## list all existed keys
$ gnfd keys list --keyring-backend test --keyring-dir ./deployment/localup/.local/validator0
## delete the key
$ gnfd keys delete $moniker --keyring-backend test --keyring-dir ./deployment/localup/.local/validator0
```

## Detailed CLI

A user can interact with the `keyring` module using the CLI.

#### add

The `add` command allows users to add a new key with specified name, encrypt it, and save to <name> file.

```sh
gnfd keys add <name> [flags]
```

Example:

```sh
gnfd keys add temp
```

Example Output:

```yml
- address: 0x48D54D...48b6d2
  name: temp
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AqAzVo.../rZgIu"}'
  pubkey_hex: 02a03356...eb66022e
  type: local

**Important** write this mnemonic phrase in a safe place.
It is the only way to recover your account if you ever forget your password.

position number holiday ... cause bounce mercy
```

#### delete

The `delete` command allows users to delete given keys.

```sh
gnfd keys delete <name>... [flags]
```

Example:

```sh
gnfd keys delete temp
```

Example Output:

```yml
Key reference will be deleted. Continue? [y/N]: y
Key deleted forever (uh oh!)
```

#### export

The `export` command allows users to export private keys.

```sh
gnfd keys export <name> [flags]
```

Example:

```sh
gnfd keys export temp
```

Example Output:

```yml
Enter passphrase to encrypt the exported key:

-----BEGIN TENDERMINT PRIVATE KEY-----
type: eth_secp256k1
kdf: bcrypt
salt: D4AEA1...36B597

3tvq8w39N...pTwLRdgVGI=
=Wfrn
-----END TENDERMINT PRIVATE KEY-----
```

#### import

The `import` command allows users to import private keys.

```sh
gnfd keys import <name> <keyfile> [flags]
```

Example:

```sh
gnfd keys import temp temp.info
```

#### list

The `list` command allows users to list all existed keys.

```sh
gnfd keys list [flags]
```

Example:

```sh
gnfd keys list
```

Example Output:

```yml
- address: 0xA627c3...B61dF7
  name: temp
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AhJl...0t8hJC"}'
  pubkey_hex: 021265c6...2df21242
  type: local
```

#### migrate

The `migrate` command can migrate keys from amino to proto serialization format.

```sh
gnfd keys migrate [flags]
```

Example:

```sh
gnfd keys migrate
```

#### mnemonic

The `mnemonic` command can compute the bip39 mnemonic for some input entropy.

```sh
gnfd keys mnemonic [flags]
```

Example:

```sh
gnfd keys mnemonic
```

Example Output:

```yml
bridge bleak zebra ... gym quick antique
```

#### rename

The `rename` command allows users to rename an existed key.

```sh
gnfd keys rename <old_name> <new_name> [flags]
```

Example:

```sh
gnfd keys rename temp new_name
```

Example Output:

```yml
Key reference will be renamed from temp to new_name. Continue? [y/N]: y
Key was successfully renamed from temp to new_name
```

#### show

The `show` command allows users to retrieve key information by name or address.

```sh
gnfd keys show [name_or_address [name_or_address...]] [flags]
```

Example:

```sh
gnfd keys show temp
```

Example Output:

```yml
- address: 0xA627c3...B61dF7
  name: temp
  pubkey: '{"@type":"/ethermint.crypto.v1.ethsecp256k1.PubKey","key":"AhJl...0t8hJC"}'
  pubkey_hex: 021265c6...2df21242
  type: local
```

