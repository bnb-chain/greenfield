# Using gnfd command to interact with storage module

## HeadBucket

```shell
gnfd query storage head-bucket [bucket-name] [flags]
```

## HeadObject

```shell
gnfd query storage head-object [bucket-name] [object-name] [flags]
```

## Mirror Bucket

The `mirror-bucket` allows users to mirror a bucket to BSC by bucket id.

```shell
gnfd tx storage mirror-bucket [bucket-id] [flags]
```

## Mirror Object

The `mirror-object` allows users to mirror an object to BSC by object id.

```shell
gnfd tx storage mirror-object [object-id] [flags]
```

## Mirror Group

The `mirror-group` allows users to mirror a group to BSC by group id.

```shell
gnfd tx storage mirror-group [group-id] [flags]
```


# Others operations
Interacting with the storage module involves a lot of interface interactions with the Storage Provider in order to 
complete tasks such as obtaining authentication information and sending request data. As a result, a single gnfd client
cannot complete the entire process, such as obtaining an approval signature from the SP. Therefore, we 
recommend using more powerful [greenfield commands](https://github.com/bnb-chain/greenfield-cmd) to complete such transactions and queries.