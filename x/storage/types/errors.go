package types

import (
	"cosmossdk.io/errors"
)

// x/storage module sentinel errors
var (
	ErrNoSuchBucket             = errors.Register(ModuleName, 1100, "No such bucket")
	ErrNoSuchObject             = errors.Register(ModuleName, 1101, "No such object")
	ErrNoSuchGroup              = errors.Register(ModuleName, 1102, "No such group")
	ErrNoSuchGroupMember        = errors.Register(ModuleName, 1103, "No such group member")
	ErrBucketAlreadyExists      = errors.Register(ModuleName, 1104, "Bucket already exists")
	ErrObjectAlreadyExists      = errors.Register(ModuleName, 1105, "Object already exists")
	ErrGroupAlreadyExists       = errors.Register(ModuleName, 1106, "Group already exists")
	ErrAccessDenied             = errors.Register(ModuleName, 1107, "Access denied")
	ErrObjectAlreadySealed      = errors.Register(ModuleName, 1108, "Object already sealed")
	ErrBucketNotEmpty           = errors.Register(ModuleName, 1109, "Bucket is not empty")
	ErrGroupMemberAlreadyExists = errors.Register(ModuleName, 1110, "Group member already exists")
	ErrNoSuchStorageProvider    = errors.Register(ModuleName, 1111, "No such storage provider")
	ErrObjectNotCreated         = errors.Register(ModuleName, 1112, "Object not created")
	ErrObjectNotSealed          = errors.Register(ModuleName, 1113, "Object not sealed")
	ErrSourceTypeMismatch       = errors.Register(ModuleName, 1114, "Object source type mismatch")
	ErrTooLargeObject           = errors.Register(ModuleName, 1115, "Object payload size is too large")
	ErrInvalidApproval          = errors.Register(ModuleName, 1116, "Invalid approval of sp")
	ErrBucketBillNotEmpty       = errors.Register(ModuleName, 1117, "bucket bill is not empty")

	ErrNoSuchPolicy          = errors.Register(ModuleName, 1120, "No such Policy")
	ErrInvalidParameter      = errors.Register(ModuleName, 1121, "Invalid parameter")
	ErrInvalidRedundancyType = errors.Register(ModuleName, 1122, "Invalid redundancy type")

	ErrInvalidCrossChainPackage = errors.Register(ModuleName, 3000, "invalid cross chain package")
	ErrAlreadyMirrored          = errors.Register(ModuleName, 3001, "resource is already mirrored")
	ErrInvalidOperationType     = errors.Register(ModuleName, 3002, "invalid operation type")
	ErrInvalidId                = errors.Register(ModuleName, 3003, "is is invalid")
)
