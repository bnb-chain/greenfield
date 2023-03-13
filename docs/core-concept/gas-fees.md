# Gas and Fees

This document describes how Greenfield charge fee to different transaction types.

## Introduction to `Gas` and `Fees`

In the Cosmos SDK, `gas` is a special unit that is used to track the consumption of resources during execution. 

However, on application-specific blockchains like Greenfield, the primary factor determining the transaction fee 
is no longer the computational cost of storage, but rather the incentive mechanism of Greenfield. For example, 
creating and deleting a storage object consumes similar I/O and computational resources, but Greenfield 
incentives users to delete unused storage objects to free up more storage space, resulting in much cheaper 
transaction fees for the latter.

Therefore, we abandoned the gas meter design of cosmos-sdk and redesigned the gashub module to determine the gas 
consumption based on the type and content of the transaction, rather than the consumption of storage and computational resources.


## GasHub
All transaction types need to register their gas calculation logic to gashub. Currently, four types of calculation logic 
are supported:

```go
type MsgGasParams_FixedType struct {
	FixedType *MsgGasParams_FixedGasParams 
}
type MsgGasParams_GrantType struct {
	GrantType *MsgGasParams_DynamicGasParams 
}
type MsgGasParams_MultiSendType struct {
	MultiSendType *MsgGasParams_DynamicGasParams 
}
type MsgGasParams_GrantAllowanceType struct {
	GrantAllowanceType *MsgGasParams_DynamicGasParams 
}
```

### Block Gas Meter

`ctx.BlockGasMeter()` is the gas meter used to track gas consumption per block and make sure it does not go above a certain limit. 

However, Greenfield may charge a substantial fee for certain types of transactions, resulting in significant gas 
consumption. As a result, Greenfield does not impose any limitations on the gas usage of a block. Greenfield limits the 
block size to under 1MB in order to prevent excessively large blocks.

## Fee Table
TODO(coming soon)

