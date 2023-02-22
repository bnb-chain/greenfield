# Key Management

The [greenfield-cosmos-sdk](https://github.com/bnb-chain/greenfield-cosmos-sdk) provides a flexible and secure way to manage private keys for blockchain applications. The `Keyring` interface is used to define the methods that a type needs to implement to be used as a key storage backend. This document provides an overview of the different backend options available and the supported sign algorithms.

And to interact with BSC(BNB-smart-chain) more convinent and user-friendly, we add [EIP-712](https://eips.ethereum.org/EIPS/eip-712) support. Any Etherrum wallet could connect to a greenfield node and sign a EIP-712 transaction directly.

## EIP-712 Support

The greenfield chain supports and only supports EIP-712 structured transaction. To achieve this, the following changes have been made.

1. An Ethereum-compatible RPC backend. Be noted that we only support necessacry methods to connect a wallet(`eth_chainId`, `eth_networkId`, `eth_blockNumber`, `eth_getBlockByNumber` and `eth_getBalance`). Other RPC methods are not implemented.
2. Same signing algorithm(`eth_scep256k1`) as Ethereum.

For developers, they can refer to [greenfield-go-sdk](https://github.com/bnb-chain/greenfield-go-sdk) and [greenfield-js-sdk](https://github.com/bnb-chain/greenfield-js-sdk) for more infos.

## Keyring Interface

The `Keyring` interface is the primary interface for key management in the greenfield-cosmos-sdk. It defines the methods that a type needs to implement to be used as a key storage backend. These methods include:

-   `Get`: retrieves a key by name.
-   `List`: lists all keys stored in the keyring.
-   `Delete`: deletes a key by name.
-   `Sign`: signs a message using a key.

By implementing these methods, you can create a custom key storage backend that meets the specific needs of your application.

## Backend Options

The greenfield-cosmos-sdk provides several backend options for key storage. Each backend has its own strengths and weaknesses, and the choice of backend will depend on your specific use case. Here are the available options:

1. **os**: The os backend uses the operating system's default credentials store to handle key storage operations securely. The keyring may be kept unlocked for the whole duration of the user session.

2. **file**: This backend more closely resembles the previous keyring storage used prior to cosmos-sdk v0.38.1. It stores the keyring encrypted within the app's configuration directory. This keyring will request a password each time it is accessed, which may occur multiple times in a single command resulting in repeated password prompts.

3. **kwallet**: This backend uses the KDE Wallet Manager as a credentials management application.

4. **pass**: This backend uses the `pass` command line utility to store and retrieve keys.

5. **test**: This backend stores keys insecurely to disk. It does not prompt for a password to be unlocked and should only be used for testing purposes.

6. **memory**: This backend uses a transient storage. Keys are discarded when the process terminates or the type instance is garbage collected.

## Supported Sign Algorithms

The greenfield-cosmos-sdk supports as many sign algorithms as users want, but in Greenfield's context, we only support `eth_secp256k1` and `ed25519`. These algorithms were chosen for their security and compatibility with the Ethereum and Tendermint ecosystems.

## Conclusion

In conclusion, key management is a critical aspect of building secure blockchain applications. By understanding the available backend options and supported sign algorithms in the greenfield-cosmos-sdk, you can build applications that are both secure and user-friendly. The `Keyring` interface provides a flexible way to manage private keys, and by implementing the required methods, you can create a custom key storage backend that meets the specific needs of your application.
