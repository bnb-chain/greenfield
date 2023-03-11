# permission

The permission module implements basic permission control management for Greenfield Storage Network.

The data resources, including the objects, buckets, payment accounts, and groups, all have permissions related. These
permissions define whether each account can perform particular actions.

Group is a list of accounts that can be treated in the same way as a single account. Examples of permissions are:

* Put, List, Get, Delete, Copy, and Execute data objects;
* Create, Delete, and List buckets
* Create, Delete, ListMembers, Leave groups
* Create, Associate payment accounts
* Grant, Revoke the above permissions

These permissions are associated with the data resources and accounts/groups,
and the group definitions are stored on the Greenfield blockchain publicly. Now they are in plain text. Later a privacy
mode will be introduced based on Zero Knowledge Proof technology.

One thing that makes the permission operation more interesting is that they can be operated from BSC directly, either
through smart contracts or by an EOA.

The basic interface semantics of permission module are similar to those of S3.

## Concepts

### Terminology

- **Resource**: Buckets, objects, group are the Greenfield resources for which you can allow or deny permissions. In a
  policy, you use the Greenfield Resource Name (GRN)to identify the resource.
- **Action**: For each resource, Greenfield supports a set of operations. You should provide an action enum value to
  identify resource operations that you will allow (or deny).
- **Principal**: The account or group who is allowed access to the resource and action that you specify in `Policy`
- **Statement**: The detail specifics of policy which include `Effect`, `ActionList` and `Resouces`
- **Effect**: What the effect will be when the user requests the specific action—this can be either allow or deny.

### Resource

In greenfield, a resource is an entity that you can operate with. Buckets, objects and groups are the resources, and
both have associated subresources.

Bucket subresources include the following:

- **BucketInfo**: Uses can update some modifiable fields of a bucket. E.g, IsPublic, ReadQuota, paymentAccount and so
  on.
- **Policy**: Store access permissions information for the bucket
- **Objects**: Each object must be stored in some bucket
- **Objects ownership**: The bucket owner takes ownership of new objects in the bucket by default, regardless of who
  uploads them

Object subresources include the following:

- **ObjectInfo**:  Uses can update some modifiable fields of an object. E.g. IsPublic and so on.
- **Policy**: Store access permissions information for the object

Group subresources include the following:

- **GroupInfo**:  Uses can update some modifiable fields of a group. E.g. members, user-meta and so on
- **Policy**: Store access permissions information for the object
- **GroupMember**: Any account in Greenfield can become a member of the group, but a group can not become a member of
  another group.

### Ownership

The resources owner refers to the account that creates the resource. By default, only the resource owner has permission
to access its resources.

- The resource creator owns the resource.
- Each resource only has one owner and the ownership cannot be transferred once the resource is created.
- There are features that allow an address to "approve" another address to create and upload objects to be owned by the
  approver's address, as long as it is within limits.
- The owner or payment account of the owner pays for the resource.

### Definitions

* Ownership Permission: By default, the owner has all permissions to the resource.
* Public or Private Permission: By default, the resource is private, only the owner can access the resource. If a
  resource is public, anyone can access it.
* Shared Permission: These permissions are managed by the permission module. It usually manages who has permission for
  which resources.

There are two types of shared permissions: Single Account Permission and Group Permission, which are stored in different
formats in the blockchain state.

### Revoke

Users can actively one or more permissions. However, when a resource is deleted, its associated permissions should also
be deleted, perhaps not by the user taking the initiative to delete it, but by other clean-up mechanisms. Because if too
many accounts have permission to the deleting object, it takes too much time to process within one transaction handling.

### Example

Let's say Greenfield has two accounts, Bob(0x1110) and Alice(0x1111). A basic example scenario would be:

* Bob uploaded a picture named avatar.jpg into a bucket named "profile";
* Bob grants the GetObject action permission for the object avatar.jpg to Alice, it will store a key 0x11 | ResourceID(
profile_avatar.jpg) | Address(Alice) into a permission state tree.
* when Alice wants to read the avatar.jpg, the SP should check the Greenfield blockchain that whether the key Prefix(
ObjectPermission) | ResourceID(profile_avatar.jpg) | Address(Alice) existed in the permission state tree, and whether
the action list contains GetObject.

Let's move on to more complex scenarios:

* Bob created a bucket named "profile"
* Bob grants the PutObject action permission for the bucket "profile" to Alice, the key 0x10 | ResourceID(profile) |
Address(Alice) will be put into the permission state tree.
* When Alice wants to upload an object named avatar.jpg into the bucket profile, it creates a PutObject transaction and
will be executed on the chain.
* The Greenfield blockchain needs to confirm that Alice has permission to operate, so it checks whether the key 0x10 |
ResourceID(profile) | Address(Alice) existed in the permission state tree and got the permission information if it
existed.
* If the permission info shows that Alice has the PutObject action permission of the bucket profile, then she can do
PutObject operation.

Another more complex scenario that contains groups:

* Bob creates a group named "Games", and creates a bucket named "profile".
* Bob adds Alice to the Games group, the key 0x12 | ResourceID(Games) | Address(Alice) will be put into the permission
state tree
* Bob put a picture avatar.jpg to the bucket profile and grants the CopyObject action permission to the group Games
* When Alice wants to copy the object avatar.jpg . First, Greenfield blockchain checks whether she has permission via the
key 0x11 | ResourceID(avatar.jpg) | Address(Alice); if it is a miss, Greenfield will iterate all groups that the object
avatar.jpg associates and check whether Alice is a member of one of these groups by checking, e.g. if the key 0x21 |
ResourceID(group, e.g. Games) existed, then iterate the permissionInfo map, and determine if Alice is in a group which
has the permission to do CopyObject operation via the key 0x12| ResourceID(Games) | Address(Alice)

## State

The permission module keeps state of the following primary objects.

1. `Policy`: The owner account of the resource grant its specify permission to another account
2. `PolicyGroup`: A limited list of `Policy`, and each `Policy` items defines the owner account of the resource grant
   its specify permission to a group

These primary objects should be primarily stored and accessed by the `ID` which is an auto-incremented sequence. An
additional indices are maintained per primary objects in order to compatibility with the S3 object storage.

* BucketPolicyForAccount: `0x11 | BigEndian(BucketID) | AccAddress -> BigEndian(PolicyID)`
* ObjectPolicyForAccount: `0x12 | BigEndian(ObjectID) | AccAddress -> BigEndian(PolicyID)`
* GroupPolicyForAccount: `0x13 | BigEndian(GroupID) | AccAddress -> BigEndian(PolicyID)`
* BucketPolicyForGroup: `0x21 | BigEndian(BucketID) -> ProtoBuf(PolicyGroup)`
* ObjectPolicyForGroup: `0x22 | BigEndian(ObjectID) -> ProtoBuf(PolicyGroup)`
* PolicyByID: `0x31 | BigEndian(PolicyID) -> ProtoBuf(Policy)`

### Policy

```protobuf
message Policy {
  string id = 1 [
    (cosmos_proto.scalar) = "cosmos.Uint",
    (gogoproto.customtype) = "Uint",
    (gogoproto.nullable) = false
  ];
  permission.Principal principal = 2;
  resource.ResourceType resource_type = 3;
  string resource_id = 4 [
    (cosmos_proto.scalar) = "cosmos.Uint",
    (gogoproto.customtype) = "Uint",
    (gogoproto.nullable) = false
  ];
  repeated permission.Statement statements = 5;
  permission.Statement member_statement = 6;
}
```

### PolicyGroup

Each resource can only grant permissions to a limited number of groups and limited number is defined
by `MaximumGroupNum` in module params.

```protobuf
message PolicyGroup {
  message Item {
    string policy_id = 1 [
      (cosmos_proto.scalar) = "cosmos.Uint",
      (gogoproto.customtype) = "Uint",
      (gogoproto.nullable) = false
    ];
    string group_id = 2 [
      (cosmos_proto.scalar) = "cosmos.Uint",
      (gogoproto.customtype) = "Uint",
      (gogoproto.nullable) = false
    ];
  }
  repeated Item items = 1;
}
```

### params

```protobuf
// Params defines the parameters for the module.
message Params {
  option (gogoproto.goproto_stringer) = false;

  uint64 maximum_statements_num = 1;
  uint64 maximum_group_num = 2;
}
```

## Message

> Notice: Permission-related messages are defined in the storage module

### PutPolicy

Use to create a policy for the resource.

```protobuf
message MsgPutPolicy {
  option (cosmos.msg.v1.signer) = "operator";

  // operator defines the granter who grant the permission to another principal
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // Principal define the roles that can grant permissions. Currently, it can be account or group.
  permission.Principal principal = 2;

  // resource define a greenfield standard resource name that can be generated by GRN structure
  string resource = 3;

  // statements define a list of individual statement which describe the detail rules of policy
  repeated permission.Statement statements = 4;
}
```

### DeletePolicy

Use to delete the policy which associate with an account or a group and a resource.

```protobuf
message MsgDeletePolicy {
  option (cosmos.msg.v1.signer) = "operator";

  // operator defines the granter who grant the permission to another principal
  string operator = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // Principal define the roles that can grant permissions. Currently, it can be account or group.
  permission.Principal principal = 2;

  // resource define a greenfield standard resource name that can be generated by GRN structure
  string resource = 3;
}
```

## Events


