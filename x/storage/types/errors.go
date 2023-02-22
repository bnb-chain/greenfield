package types

import (
	"cosmossdk.io/errors"
)

// x/storage module sentinel errors
var (
	ErrNoSuchBucket                = errors.Register(ModuleName, 1100, "no such bucket")
	ErrNoSuchObject                = errors.Register(ModuleName, 1101, "no such object")
	ErrNoSuchGroup                 = errors.Register(ModuleName, 1102, "no such group")
	ErrBucketAlreadyExists         = errors.Register(ModuleName, 1103, "bucket already exists")
	ErrObjectAlreadyExists         = errors.Register(ModuleName, 1104, "object already exists")
	ErrGroupAlreadyExists          = errors.Register(ModuleName, 1105, "group already exists")
	ErrAccessDenied                = errors.Register(ModuleName, 1106, "access denied")
	ErrObjectAlreadySealed         = errors.Register(ModuleName, 1107, "object already sealed")
	ErrBucketNotEmpty              = errors.Register(ModuleName, 1108, "bucket is not empty")
	ErrNoSuchGroupMember           = errors.Register(ModuleName, 1109, "no such group member")
	ErrGroupMemberAlreadyExists    = errors.Register(ModuleName, 1110, "group member already exists")
	ErrGroupMemberNotExists        = errors.Register(ModuleName, 1111, "group member already exists")
	ErrNoSuchStorageProvider       = errors.Register(ModuleName, 1112, "no such storage provider")
	ErrSequenceUniqueConstraint    = errors.Register(ModuleName, 1113, "sequence already initialized")
	ErrStorageProviderNotInService = errors.Register(ModuleName, 1114, "storage provider not in service")
	ErrObjectNotInit               = errors.Register(ModuleName, 1115, "not a INIT object")
	ErrObjectNotInService          = errors.Register(ModuleName, 1116, "object not in service")
	ErrSourceTypeMismatch          = errors.Register(ModuleName, 1117, "object source type mismatch")
	ErrTooLargeObject              = errors.Register(ModuleName, 1118, "object payload size is too large")
	ErrInvalidApproval             = errors.Register(ModuleName, 1119, "Invalid approval of sp")

	ErrInvalidBucketName  = errors.Register(ModuleName, 2000, "invalid bucket name")
	ErrInvalidObjectName  = errors.Register(ModuleName, 2001, "invalid object name")
	ErrInvalidGroupName   = errors.Register(ModuleName, 2002, "invalid group name")
	ErrInvalidChcecksum   = errors.Register(ModuleName, 2003, "invalid checksum")
	ErrInvalidContentType = errors.Register(ModuleName, 2004, "invalid content type")
	ErrInvalidSPSignature = errors.Register(ModuleName, 2005, "invalid sp signature")
	ErrInvalidSPAddress   = errors.Register(ModuleName, 2006, "invalid sp address")
)
