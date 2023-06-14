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
	ErrChargeFailed             = errors.Register(ModuleName, 1117, "charge failed error")
	ErrInvalidVisibility        = errors.Register(ModuleName, 1118, "Invalid type of visibility")

	ErrNoSuchPolicy          = errors.Register(ModuleName, 1120, "No such Policy")
	ErrInvalidParameter      = errors.Register(ModuleName, 1121, "Invalid parameter")
	ErrInvalidRedundancyType = errors.Register(ModuleName, 1122, "Invalid redundancy type")

	ErrInvalidCrossChainPackage = errors.Register(ModuleName, 3000, "invalid cross chain package")
	ErrAlreadyMirrored          = errors.Register(ModuleName, 3001, "resource is already mirrored")
	ErrInvalidOperationType     = errors.Register(ModuleName, 3002, "invalid operation type")
	ErrInvalidId                = errors.Register(ModuleName, 3003, "id is invalid")

	ErrInvalidObjectIds    = errors.Register(ModuleName, 3101, "object ids are invalid")
	ErrInvalidReason       = errors.Register(ModuleName, 3102, "reason is invalid")
	ErrNoMoreDiscontinue   = errors.Register(ModuleName, 3103, "no more discontinue requests")
	ErrBucketDiscontinued  = errors.Register(ModuleName, 3104, "the bucket is discontinued")
	ErrInvalidObjectStatus = errors.Register(ModuleName, 3105, "invalid object status")
	ErrInvalidBucketStatus = errors.Register(ModuleName, 3106, "invalid bucket status")

	ErrKeyNotExist     = errors.Register(ModuleName, 3201, "DeletePermissionKey not exist")
	ErrInvalidResource = errors.Register(ModuleName, 3202, "invalid resource type")

	ErrInvalidTaskId            = errors.Register(ModuleName, 3301, "invalid task id")
	ErrExecutionResultSubmitted = errors.Register(ModuleName, 3302, "execution result submitted")
)
