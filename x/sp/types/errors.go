package types

import (
	"cosmossdk.io/errors"
)

// x/sp module sentinel errors
var (
	ErrStorageProviderNotFound              = errors.Register(ModuleName, 1, "StorageProvider does not exist")
	ErrStorageProviderNotInService          = errors.Register(ModuleName, 2, "StorageProvider is not in service")
	ErrStorageProviderOwnerExists           = errors.Register(ModuleName, 3, "StorageProvider already exist for this operator address; must use new StorageProvider operator address")
	ErrInsufficientDepositAmount            = errors.Register(ModuleName, 4, "insufficient deposit amount")
	ErrDepositAccountNotAllowed             = errors.Register(ModuleName, 5, "the deposit address must be the fund address of sp.")
	ErrInvalidDenom                         = errors.Register(ModuleName, 6, "Invalid denom.")
	ErrStorageProviderFundingAddrExists     = errors.Register(ModuleName, 7, "StorageProvider already exist for this funding address; must use new StorageProvider funding address.")
	ErrStorageProviderSealAddrExists        = errors.Register(ModuleName, 8, "StorageProvider already exist for this seal address; must use new StorageProvider seal address.")
	ErrStorageProviderApprovalAddrExists    = errors.Register(ModuleName, 9, "StorageProvider already exist for this approval address; must use new StorageProvider approval address.")
	ErrStorageProviderGcAddrExists          = errors.Register(ModuleName, 10, "StorageProvider already exist for this gc address; must use new StorageProvider gc address.")
	ErrStorageProviderPriceExpired          = errors.Register(ModuleName, 11, "StorageProvider price expired")
	ErrStorageProviderNotChanged            = errors.Register(ModuleName, 12, "StorageProvider not changed")
	ErrStorageProviderExitFailed            = errors.Register(ModuleName, 13, "StorageProvider exit failed.")
	ErrStorageProviderInvalidBlsKey         = errors.Register(ModuleName, 14, "StorageProvider bls pubkey is invalid")
	ErrStorageProviderBlsKeyExists          = errors.Register(ModuleName, 15, "StorageProvider already exist for this bls pubkey; must use new bls pubkey")
	ErrStorageProviderStatusUpdateNotAllow  = errors.Register(ModuleName, 16, "StorageProvider status is not allow to change")
	ErrStorageProviderMaintenanceAddrExists = errors.Register(ModuleName, 17, "StorageProvider already exist for this maintenance address; must use new StorageProvider maintenance address.")
	ErrStorageProviderPriceUpdateNotAllow   = errors.Register(ModuleName, 18, "StorageProvider update price is disallowed")
	ErrStorageProviderWrongStatus           = errors.Register(ModuleName, 19, "StorageProvider is in wrong status")

	ErrSignerNotGovModule  = errors.Register(ModuleName, 40, "signer is not gov module account")
	ErrSignerEmpty         = errors.Register(ModuleName, 41, "signer is empty")
	ErrInvalidEndpointURL  = errors.Register(ModuleName, 42, "Invalid endpoint url")
	ErrSignerNotSPOperator = errors.Register(ModuleName, 43, "signer is not sp operator account")
)
