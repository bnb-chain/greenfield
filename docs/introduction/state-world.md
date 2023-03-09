# State World of Greenfield Blockchain
All Greenfield validators have such active data in full (at least the latest state). Anyone can join the blockchain as
full nodes to synchronize these data for free.

These data are on-chain and can be only changed through transactions onto the Greenfield blockchain. It has several types
as described below.

## Accounts and Balance
Each user has their "Owner Address" as the identifier for their owner account to "own" the data resources. There is
another "payment account" type dedicated to billing and payment purposes and owned by owner addresses.

Both owner accounts and payment accounts can hold the BNB balance on Greenfield. Users can deposit BNB from BSC, accept
transfers from other users, and spend them on transaction gas and storage usage.

## Validator and SP Metadata
These are the basic information about the Greenfield validators and Greenfield SPs. SPs may have more information, as
it has to publish their service information for users' data operations. There should be a reputation mechanism for SPs
as well.

## Storage Metadata
The "storage metadata" includes size, ownership, checksum hashes, and distribution location among SPs. Similar to AWS S3,
the basic unit of the storage is an "object", which can be a piece of binary data, text files, photos, videos, or any
other format. Users can create their objects under their "bucket". A bucket is globally unique. The object can be referred
to via the bucket name and the object ID. It can also be located by the bucket name, the prefix tag, and the object ID
via off-chain facilitations.

## Permission Metadata
Data resources on Greenfield, such as the data objects and the buckets, all have access control, such as which address
can create, read, list, or even execute the resources, and which address can grant/revoke these permissions.

Two other data resources also have access control. One is "Group". A group represents a group of user addresses that have
the same permissions to the same resources. It can be used in the same way as an address in the access control. Meanwhile,
it requires permission too to change the group. The other is "payment account". They are created by the owner accounts.

Here the access control is enforced by the SPs off-chain. People can test and challenge the SPs if they mess up the
control. Slash and reward will happen to keep the SPs sticking to the principles.

## Billing Metadata
Users have to pay fees to store data objects on Greenfield. While each object enjoys a free quota to download by users
who are permitted to, the excessive download will require extra data packages to be paid for the bandwidth. Besides
the owner address, users can derive multiple "Payment Addresses" to pay these fees. Objects are stored under buckets,
while each bucket can be associated with these payment addresses, and the system will charge these accounts for storing
and/or downloading. Many buckets can share the same payment address. Such association information is also stored on
chains with consensus as well.