package types

import (
	"cosmossdk.io/errors"
)

// x/sp module sentinel errors
var (
	ErrEmptyStorageProviderAddr    = errors.Register(ModuleName, 1, "empty StorageProvider address")
	ErrStorageProviderNotFound     = errors.Register(ModuleName, 2, "StorageProvider does not exist")
	ErrEmptyStorageProviderPubKey  = errors.Register(ModuleName, 3, "empty storage provider public key")
	ErrStorageProviderOwnerExists  = errors.Register(ModuleName, 4, "StorageProvider already exist for this operator address; must use new StorageProvider operator address")
	ErrStorageProviderPubKeyExists = errors.Register(ModuleName, 5, "StorageProvider already exist for this pubkey; must use new StorageProvider pubkey")
	ErrInsufficientDepositAmount   = errors.Register(ModuleName, 6, "insufficient deposit amount")
	ErrDepositAccountNotAllowed    = errors.Register(ModuleName, 7, "the deposit address must be the sp address or the fund address of sp.")
	ErrInvalidDepositDenom         = errors.Register(ModuleName, 8, "the deposit address must be the sp address or the fund address of sp.")

	ErrSignerNotGovModule = errors.Register(ModuleName, 40, "signer is not gov module account")
	ErrSignerEmpty        = errors.Register(ModuleName, 41, "signer is empty")
)
