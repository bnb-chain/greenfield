## Ecosystem Players
There are several player roles for the whole Greenfield ecosystem.

<div align="center"><img src="../asset/01-All%20Players%20of%20Greenfield.jpg"  height="80%" width="80%"></div>
<div align="center"><i>Figure All Players of Greenfield</i></div>

### Greenfield Validators

As a PoS blockchain, the Greenfield blockchain has its own validators.
These validators are elected based on the Proof-of-Stake logic.

These validators are responsible for the security of the Greenfield
blockchain. They get involved in the governance and staking of the
blockchain. They form a P2P network that is similar to other PoS
blockchain networks.

Meanwhile, they accept and process transactions to allow users to
operate their objects stored on Greenfield. They maintain the metadata
of Greenfield as the blockchain state, which is the control panel for
both Storage Providers (SPs) and users. These two parties use and update
these states to operate the object storage.

The network topology of Greenfield validators is similar to the existing
secure validator setup of PoS blockchains. "Sentry Nodes" are used to
defend against DDoS and provide a secure private network, as shown in
the below diagram.

<div align="center"><img src="../asset/02-Greenfield%20Blockchain%20Network.jpg"  height="80%" width="80%"></div>
<div align="center"><i>Figure Greenfield Blockchain Network</i></div>

### Greenfield Relayers
The Greenfield Relayer is a bidirectional relaying tool that facilitates communication between 
Greenfield and BSC. This standalone process can only be run by Greenfield validators. The relayer
independently monitors cross-chain events occurring on BSC and Greenfield, and persists them into
a database. Once a few blocks confirm the event and reach finality, the relayer will sign a message
using the BLS private key to confirm the event, and broadcast the signed event (known as "the vote")
through the P2P network on the Greenfield network. Once enough votes from the Greenfield relayer are
collected, the relayer will assemble a cross-chain package transaction and submit it to the BSC or
Greenfield network.

### Storage Providers (SPs)
Storage Providers are professional individuals and organizations who run
a series of services to provide data services based on the Greenfield
blockchain.

### Greenfield dApps
Greenfield dApps are applications that provide functions based on
Greenfield storage and its related economic traits to solve some
problems of their users.

## Participate in the Ecosystem
- [Become A Validator](../cli/validator-staking.md): validators secure the Greenfield by validating and relaying transactions,
  proposing, verifying and finalizing blocks.
- [Become A Storage Provider](../cli/storage-provider.md): SPs store the objects' real data, i.e. the payload data. and get paid
  by providing storage services.