# Greenfield blockchain events

There are two types of events doc:
1. Some old modules introduced in the cosmos-sdk don't emit typed events.
   Their events are listed in a Markdown document under the module's spec folder.
2. The new modules introduced in the cosmos-sdk or developed by the Greenfield
   team emit typed events. Their typed events are defined in a protobuf file.
   For these modules, we can refer to their protobuf file directly.

Following are the events that the Greenfield blockchain emits, grouped by modules:

* [Authz](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/proto/cosmos/authz/v1beta1/event.proto)

* [Bank](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/bank/spec/04_events.md)

* [Bridge](https://github.com/bnb-chain/greenfield/blob/master/proto/greenfield/bridge/event.proto)

* [Distribution](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/distribution/spec/06_events.md)

* [Feegrant](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/feegrant/spec/04_events.md)

* [Gov](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/gov/spec/04_events.md)

* [Oracle](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/proto/cosmos/oracle/v1/event.proto)

* [Slashing](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/slashing/spec/06_events.md)

* [SP](https://github.com/bnb-chain/greenfield/blob/develop/proto/greenfield/sp/events.proto)

* [Staking](https://github.com/bnb-chain/gnfd-cosmos-sdk/blob/master/x/staking/spec/07_events.md)

