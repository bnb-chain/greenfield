# Accounts

This document describes the in-built account and public key system of the Greenfield

## Account Definition

In the Greenfield, an _account_ designates a pair of _public key_ `PubKey` and _private key_ `PrivKey`. 
The `PubKey` can be derived to generate various `Addresses`, which are used to identify users (among other parties) in 
the application.

## Signatures

The principal way of authenticating a user is done using [digital signatures](https://en.wikipedia.org/wiki/Digital_signature). 
Users sign transactions using their own private key. Signature verification is done with the associated public key. 
For on-chain signature verification purposes, we store the public key in an `Account` object (alongside other data required 
for a proper transaction validation).

In the node, all data is stored using Protocol Buffers serialization.

Greenfield only supports `secp256k1` key schemes for creating digital signatures:

|             | Address length in bytes | Public key length in bytes | Used for transaction authentication | Used for consensus (tendermint) |
|:-----------:|:-----------------------:|:--------------------------:|:-----------------------------------:|:-------------------------------:|
| `secp256k1` |           20            |             33             |                 yes                 |               no                |

## Addresses

`Addresses` and `PubKey`s are both public information that identifies actors in the application. `Account` is used to 
store authentication information. The basic account implementation is provided by a `BaseAccount` object.

Unlike Cosmos SDK who defines 3 types of addresses: `AccAddress`, 
`ValAddress` and `ConsAddress`,  Greenfield only use the `AccAddress`, and the address format follows [ERC-55](https://eips.ethereum.org/EIPS/eip-55).