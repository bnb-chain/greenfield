# Key Management

Greenfield use cosmos keyring module to help users to manage their private keys and interacte with node.

The `Keyring` interface defines the methods that a type needs to implement to be used as key storage backend. To create a new  instance, there’re two key options to determine, backend and supported algorithms.

## Backend

There’re 6 kinds of backends out-of-the-box.

1. os. The os backend uses the operating system's default credentials store to handle keys storage operations securely. It should be noted that the keyring may be kept unlocked for the whole duration of the user session.
2. file. This backend more closely resembles the previous keyring storage used prior to cosmos-sdk v0.38.1. It stores the keyring encrypted within the app's configuration directory. This keyring will request a password each time it is accessed, which may occur multiple times in a single command resulting in repeated password prompts.
3. kwallet. This backend uses KDE Wallet Manager as a credentials management application: https://github.com/KDE/kwallet.
4. pass. This backend uses the pass command line utility to store and retrieve keys: https://www.passwordstore.org/.
5. test. This backend stores keys insecurely to disk. It does not prompt for a password to be unlocked and it should be use only for testing purposes.
6. memory. Same instance as returned by NewInMemory. This backend uses a transient storage. Keys are discarded when the process terminates or the type instance is garbage collected.

## Algorithms

The cosmos-sdk could support as many sign algorithms as users want. But in greenfield’s context, we only support `eth_secp256k1` and `ed25519`.