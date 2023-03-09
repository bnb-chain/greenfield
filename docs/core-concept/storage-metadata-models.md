# Storage MetaData Models

The Greenfield is designed to be as compatible as possible with popular Web2 and Web3 standards.
It provides developers with similar API primitives and models as the most popular Web2 cloud storage: AWS s3.

## Abstract
Below are the basic data models for Greenfield storage:

- bucket
- object
- group
- permission

These metadata are stored as blockchain state into the persistent storage of the Greenfield blockchain.

## Concepts

### Bucket

Bucket is the unit to group storage "objects". BucketName has to be globally unique. Every user account can create several
buckets. The account will become the "owner" of the bucket.

Each bucket should be associated with its own Primary SP, and the payment accounts for Read and Store. The owner's
address will be the default payment account.

The prototype definition of a bucket:

```protobuf
message BucketInfo {
  // owner is the account address of bucket creator, it is also the bucket owner.
  string owner;
  // bucket_name is a globally unique name of bucket
  string bucket_name;
  // is_public define the highest permissions for bucket. When the bucket is public, everyone can get the object in it.
  bool is_public;
  // id is the unique identification for bucket.
  string id;
  // source_type define the source of the bucket
  SourceType source_type;
  // create_at define the block number when the bucket created.
  int64 create_at;
  // payment_address is the address of the payment account
  string payment_address;
  // primary_sp_address is the address of the primary sp. Objects belongs to this bucket will never
  // leave this SP, unless you explicitly shift them to another SP.
  string primary_sp_address;
  // read_quota defines the traffic quota for read
  ReadQuota read_quota;
  int64 payment_price_time;
  // payment_out_flows, for billing;
  repeated payment.OutFlowInUSD payment_out_flows;
}
```

### Object

Object is the basic unit to store data on Greenfield. The metadata for the object will be stored on the Greenfield
blockchain:

- name and ID
- owner
- bucket that hosts it
- size and timestamps
- content type
- checkSums for the storage pieces
- storage status
- associated SP information

Object metadata is stored with the bucket name as the prefix of the key. It is possible to iterate through all
objects under the same bucket, but it may be a heavy-lifting job for a large bucket with lots of objects.

The prototype definition of an object:

```protobuf

message ObjectInfo {
  string owner;
  // bucket_name is the name of the bucket
  string bucket_name;
  // object_name is the name of object
  string object_name;
  // id is the unique identifier of object
  string id;
  // payloadSize is the total size of the object payload
  uint64 payload_size;
  // is_public define the highest permissions for object. When the object is public, everyone can access it.
  bool is_public;
  // content_type define the format of the object which should be a standard MIME type.
  string content_type;
  // create_at define the block number when the object created
  int64 create_at;
  // object_status define the upload status of the object.
  ObjectStatus object_status;
  // redundancy_type define the type of the redundancy which can be multi-replication or EC.
  RedundancyType redundancy_type;
  // source_type define the source of the object.
  SourceType source_type;
  // checksums define the root hash of the pieces which stored in a SP.
  repeated bytes checksums;
  // secondary_sp_addresses define the addresses of secondary_sps
  repeated string secondary_sp_addresses ;
  // lockedBalance
  string lockedBalance;
}
```

### Group

A Group is a collection of accounts with the same permissions. The group name is not allowed to be duplicated under the
same user. However, a group can not create or own any resource. A group can not be a member of another group either.

A resource can only have a limited number of groups associated with it for permissions. This ensures that the on-chain
permission check can be finished within a constant time.

The prototype definition of a group:

```protobuf
message GroupInfo {
  // owner is the owner of the group. It can not changed once created.
  string owner ;
  // group_name is the name of group which is unique under an account.
  string group_name;
  // source_type
  SourceType source_type;
  // id is the unique identifier of group
  string id;
}
```
