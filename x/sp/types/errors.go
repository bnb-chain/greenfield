package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/sp module sentinel errors
var (
	ErrEmptyStorageProviderAddr    = sdkerrors.Register(ModuleName, 1, "empty StorageProvider address")
	ErrStorageProviderNotFound     = sdkerrors.Register(ModuleName, 2, "StorageProvider does not exist")
	ErrEmptyStorageProviderPubKey  = sdkerrors.Register(ModuleName, 3, "empty storage provider public key")
	ErrStorageProviderOwnerExists  = sdkerrors.Register(ModuleName, 4, "StorageProvider already exist for this operator address; must use new StorageProvider operator address")
	ErrStorageProviderPubKeyExists = sdkerrors.Register(ModuleName, 5, "StorageProvider already exist for this pubkey; must use new StorageProvider pubkey")
	ErrInsufficientDepositAmount   = sdkerrors.Register(ModuleName, 6, "insufficient deposit amount")
	ErrDepositAccountNotAllowed    = sdkerrors.Register(ModuleName, 7, "the deposit address must be the sp address or the fund address of sp.")
	ErrInvalidDepositDenom         = sdkerrors.Register(ModuleName, 8, "the deposit address must be the sp address or the fund address of sp.")

	ErrSignerNotGovModule = sdkerrors.Register(ModuleName, 40, "signer is not gov module account")
	ErrSignerEmpty        = sdkerrors.Register(ModuleName, 41, "signer is empty")
)
