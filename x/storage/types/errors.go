package types

import (
	"cosmossdk.io/errors"
)

// x/storage module sentinel errors
var (
	ErrNoSuchBucket             = errors.Register(ModuleName, 1100, "No such bucket")
	ErrNoSuchObject             = errors.Register(ModuleName, 1101, "No such object")
	ErrNoSuchGroup              = errors.Register(ModuleName, 1102, "No such group")
	ErrBucketAlreadyExists      = errors.Register(ModuleName, 1103, "Bucket already exists")
	ErrObjectAlreadyExists      = errors.Register(ModuleName, 1104, "Object already exists")
	ErrGroupAlreadyExists       = errors.Register(ModuleName, 1105, "Group already exists")
	ErrAccessDenied             = errors.Register(ModuleName, 1106, "Access denied")
	ErrObjectAlreadySealed      = errors.Register(ModuleName, 1107, "Object already sealed")
	ErrBucketNotEmpty           = errors.Register(ModuleName, 1108, "Bucket is not empty")
	ErrGroupMemberAlreadyExists = errors.Register(ModuleName, 1110, "Group member already exists")
	ErrNoSuchStorageProvider    = errors.Register(ModuleName, 1112, "No such storage provider")
	ErrObjectNotInit            = errors.Register(ModuleName, 1114, "Not a INIT object")
	ErrObjectNotInService       = errors.Register(ModuleName, 1115, "Object not in service")
	ErrSourceTypeMismatch       = errors.Register(ModuleName, 1116, "Object source type mismatch")
	ErrTooLargeObject           = errors.Register(ModuleName, 1117, "Object payload size is too large")
	ErrInvalidApproval          = errors.Register(ModuleName, 1118, "Invalid approval of sp")
	ErrBucketBillNotEmpty       = errors.Register(ModuleName, 1119, "bucket bill is not empty")

	ErrNoSuchPolicy     = errors.Register(ModuleName, 1120, "No such Policy")
	ErrInvalidParameter = errors.Register(ModuleName, 1121, "Invalid parameter")
)
