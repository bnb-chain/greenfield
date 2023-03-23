# Storage Management Module

## Concepts

### Bucket

Bucket is the unit to group storage "objects". BucketName has to be globally unique. Every user account can create a
bucket. The account will become the "owner" of the bucket.

Each bucket should be associated with its own Primary SP, and the payment accounts for Read and Store. The owner's
address will be the default payment account.

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

### Group

A Group is a collection of accounts with the same permissions. The group name is not allowed to be duplicated under the
same user. However, a group can not create or own any resource. A group can not be a member of another group either.

A resource can only have a limited number of groups associated with it for permissions. This ensures that the on-chain
permission check can be finished within a constant time.

## State

The storage module keeps state of the following primary objects:

* BucketInfo
* ObjectInfo
* GroupInfo

These primary objects should be primarily stored and accessed by the `ID` which is a auto-incremented sequence. An
additional indices are maintained per primary objects in order to compatibility with the S3 object storage.

* BucketInfo: `0x11 | hash(bucketName) -> BigEndian(bucketId)`
* ObjectInfo: `0x12 | hash(bucketName)_hash(objectName) -> BigEndian(objectId)`
* GroupInfo: `0x13 | OwnerAddr_hash(groupName) -> BigEndian(groupId)`

* BucketInfoById: `0x21 | BigEndian(bucketId) -> ProtoBuf(BucketInfo)`
* ObjectInfoById: `0x22 | BigEndian(objectId) -> ProtoBuf(ObjectInfo)`
* GroupInfoById: `0x23 | BigEndian(groupId) -> ProtoBuf(GroupInfo)`

### Params

The storage module contains the following parameters,
they can be updated with governance.

| name                    | default value | meaning                                                                                                                                                                              |
|-------------------------|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| MaxSegmentSize          | 16M           | The maximum size of the segment. The payload data of an object will split into several segment. Only the size of the last segment can be less than MaxSegmentSize, others is equals. |
| RedundantDataChunkNum   | 4             | The number of the data chunks in Erasure-Code algorithm.                                                                                                                             |
| RedundantParityChunkNum | 2             | The number of the parity chunks in Erasure-Code algorithm.                                                                                                                           |
| MaxPayloadSize          | 2G            | The maximum size of the payload data that allowed in greenfield storage network.                                                                                                     |


## Messages

### MsgCreateBucket

Used to create a bucket, a bucket is used to contain storage objects.

```protobuf
message MsgCreateBucket {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the account address of bucket creator, it is also the bucket owner.
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is a globally unique name of bucket
  string bucket_name = 2;
  // visibility means the bucket is private or public. if private, only bucket owner or grantee can read it,
  // otherwise every greenfield user can read it.
  VisibilityType visibility = 3;
  // payment_address is an account address specified by bucket owner to pay the read fee. Default: creator
  string payment_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_address is the address of primary sp.
  string primary_sp_address = 6 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // primary_sp_approval is the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // read_quota
  ReadQuota read_quota = 8;
}
```

### MsgDeleteBucket

Used to delete bucket. It is important to note that you cannot delete a non-empty bucket.

```protobuf
message MsgDeleteBucket {
  option (cosmos.msg.v1.signer) = "operator";

  // creator is the account address of the grantee who has the DeleteBucket permission of the bucket to be deleted.
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // bucket_name is the name of the bucket to be deleted.
  string bucket_name = 2;
}
```

### MsgUpdateBucketInfo
Used to update bucket info.

```protobuf
message MsgUpdateBucketInfo {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the account address of the operator
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // bucket_name is the name of bucket which you'll update
  string bucket_name = 2;

  // read_quota is the traffic quota that you read from primary sp
  ReadQuota read_quota = 3;

  // payment_address is the account address of the payment account
  string payment_address = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

### MsgCreateObject

Used to create an initial object under a bucket.

```protobuf
message MsgCreateObject {
  option (cosmos.msg.v1.signer) = "creator";

  // creator is the account address of object uploader
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name is the name of object
  string object_name = 3;
  // payload_size is size of the object's payload
  uint64 payload_size = 4;
  // visibility means the bucket is private or public. if private, only bucket owner or grantee can access it,
  // otherwise every greenfield user can access it.
  VisibilityType visibility = 5;
  // content_type is a standard MIME type describing the format of the object.
  string content_type = 6;
  // primary_sp_approval is the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval primary_sp_approval = 7;
  // expect_checksums is a list of hashes which was generate by redundancy algorithm.
  repeated bytes expect_checksums = 8;
  // redundancy_type can be ec or replica
  RedundancyType redundancy_type = 9;
  // expect_secondarySPs is a list of StorageProvider address, which is optional
  repeated string expect_secondary_sp_addresses = 10 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}

```
### MsgDeleteObject

Used to delete object that is no longer useful, the deleted storage object is not recoverable.

```protobuf
message MsgDeleteObject {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the account address of the operator who has the DeleteObject permission of the object to be deleted.
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket where the object which to be deleted is stored.
  string bucket_name = 2;
  // object_name is the name of the object which to be deleted.
  string object_name = 3;
}

```
### MsgSealObject

Storage provider seal an object once the underlying files are well saved by both primary and second SPs.

```protobuf
message MsgSealObject {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the account address of primary SP
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name is the name of object to be sealed.
  string object_name = 3;
  // secondary_sp_addresses is a list of storage provider which store the redundant data.
  repeated string secondary_sp_addresses = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // secondary_sp_signatures is the signature of the secondary sp that can
  // acknowledge that the payload data has received and stored.
  repeated bytes secondary_sp_signatures = 5;
}

```
### MsgCopyObject

Used to copy an exact same object to another user.

```protobuf
message MsgCopyObject {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the account address of the operator who has the CopyObject permission of the object to be deleted.
  string operator = 1;
  // src_bucket_name is the name of the bucket where the object to be copied is located
  string src_bucket_name = 2;
  // dst_bucket_name is the name of the bucket where the object is copied to.
  string dst_bucket_name = 3;
  // src_object_name is the name of the object which to be copied
  string src_object_name = 4;
  // dst_object_name is the name of the object which is copied to
  string dst_object_name = 5;
  // primary_sp_approval is the approval info of the primary SP which indicates that primary sp confirm the user's request.
  Approval dst_primary_sp_approval = 6;
}
```
### MsgRejectSealObject

A storage provider may reject to seal an object if it refuses to, or it can not because of unexpect error.

```protobuf
message MsgRejectSealObject {
  option (cosmos.msg.v1.signer) = "operator";
  // operator is the account address of the object owner
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket where the object is stored.
  string bucket_name = 2;
  // object_name is the name of unsealed object to be reject.
  string object_name = 3;
}
```
### MsgCancelCreateObject

User are able to cancel an initial object before it is sealed.

```protobuf
message MsgCancelCreateObject {
  option (cosmos.msg.v1.signer) = "operator";
  // operator is the account address of the operator
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // bucket_name is the name of the bucket
  string bucket_name = 2;
  // object_name is the name of the object
  string object_name = 3;
}
```
### MsgCreateGroup

Used to create group.

```protobuf
message MsgCreateGroup {
  option (cosmos.msg.v1.signer) = "creator";

  // owner is the account address of group owner who create the group
  string creator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_name is the name of the group. it's not globally unique.
  string group_name = 2;
  // member_request is a list of member which to be add or remove
  repeated string members = 3 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```
### MsgDeleteGroup

Used to delete group that is no longer used. Please note that the underlying members are not deleted yet.

```protobuf
message MsgDeleteGroup {
  option (cosmos.msg.v1.signer) = "operator";

  // operator is the account address of the operator who has the DeleteGroup permission of the group to be deleted.
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_name is the name of the group which to be deleted
  string group_name = 2;
}

```
### MsgLeaveGroup

A group member can choose to leave a group. 

```protobuf
message MsgLeaveGroup {
  option (cosmos.msg.v1.signer) = "member";

  // member is the account address of the member who want to leave the group
  string member = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_owner is the owner of the group you want to leave
  string group_owner = 2 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  // group_name is the name of the group you want to leave
  string group_name = 3;
}
```
