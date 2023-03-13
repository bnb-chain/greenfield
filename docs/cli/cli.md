# Command-Line Interface

## Command-Line Interface


There is no set way to create a CLI, but Greenfield typically use the [Cobra Library](https://github.com/spf13/cobra). 
Building a CLI with Cobra entails defining commands, arguments, and flags. Commands understand the 
actions users wish to take, such as `tx` for creating a transaction and `query` for querying the application. 
Each command can also have nested subcommands, necessary for naming the specific transaction type. 
Users also supply **Arguments**, such as account numbers to send coins to, and flags to modify various 
aspects of the commands, such as gas prices or which node to broadcast to.

### Transaction Command
Here is an example of a command a user might enter to interact with `gnfd` in order to send some tokens:

```bash
$ gnfd tx bank send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000BNB --gas auto
```

The first four strings specify the command:

* The subcommand `tx`, which contains all commands that let users create transactions.
* The subcommand `bank` to indicate which module to route the command to `x/bank` module in this case.
* The type of transaction `send`.

The next two strings are arguments: the `from_address` the user wishes to send from, the `to_address` of the recipient, 
and the `amount` they want to send. Finally, the last few strings of the command are optional flags to indicate 
how much the user is willing to pay in fees.

### Query Commands

Queries are objects that allow users to retrieve information about the application's state. 

This `queryCommand` function adds all the queries available to end-users for the application. This typically includes:

* **QueryTx** and/or other transaction query commands from the `auth` module which allow the user to search for a transaction by inputting its hash, a list of tags, or a block height. These queries allow users to see if transactions have been included in a block.
* **Account command** from the `auth` module, which displays the state (e.g. account balance) of an account given an address.
* **Validator command** from the Cosmos SDK rpc client tools, which displays the validator set of a given height.
* **Block command** from the Cosmos SDK rpc client tools, which displays the block data for a given height.
* **All module query commands the application is dependent on,

Here is an example of a `queryCommand`:

```shell
## query the metadata of BNB
$ gnfd q bank  denom-metadata --node tcp://127.0.0.1:26750
```

## Environment variables

Each flag is bound to its respective named environment variable. Then name of the environment variable consist of two parts 
- capital case `basename` followed by flag name of the flag. `-` must be substituted with `_`. 
- For example flag `--home` for application with basename `GNFD` is bound to `GNFD_HOME`. It allows reducing 
the amount of flags typed for routine operations. For example instead of:

```sh
gnfd --home=./ --node=<node address> --chain-id="testchain-9000" --keyring-backend=test tx ... --from=<key name>
```

this will be more convenient:

```sh
# define env variables in .env, .envrc etc
GNFD_HOME=<path to home>
GNFD_NODE=<node address>
GNFD_CHAIN_ID="testchain-9000"
GNFD_KEYRING_BACKEND="test"

# and later just use
gnfd tx ... --from=<key name>
```
