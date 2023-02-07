package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/storage module sentinel errors
var (
	ErrNoSuchBucket                = sdkerrors.Register(ModuleName, 1100, "no such bucket")
	ErrNoSuchObject                = sdkerrors.Register(ModuleName, 1101, "no such object")
	ErrNoSuchGroup                 = sdkerrors.Register(ModuleName, 1102, "no such group")
	ErrBucketAlreadyExists         = sdkerrors.Register(ModuleName, 1103, "bucket already exists")
	ErrObjectAlreadyExists         = sdkerrors.Register(ModuleName, 1104, "object already exists")
	ErrGroupAlreadyExists          = sdkerrors.Register(ModuleName, 1105, "group already exists")
	ErrAccessDenied                = sdkerrors.Register(ModuleName, 1106, "access denied")
	ErrSPAddressMismatch           = sdkerrors.Register(ModuleName, 1107, "sp address mismatch")
	ErrObjectAlreadySealed         = sdkerrors.Register(ModuleName, 1108, "object already sealed")
	ErrBucketNotEmpty              = sdkerrors.Register(ModuleName, 1109, "bucket is not empty")
	ErrNoSuchGroupMember           = sdkerrors.Register(ModuleName, 1110, "no such group member")
	ErrGroupMemberAlreadyExists    = sdkerrors.Register(ModuleName, 1111, "group member already exists")
	ErrNoSuchStorageProvider       = sdkerrors.Register(ModuleName, 1112, "no such storage provider")
	ErrSequenceUniqueConstraint    = sdkerrors.Register(ModuleName, 1113, "sequence already initialized")
	ErrStorageProviderNotInService = sdkerrors.Register(ModuleName, 1114, "storage provider not in serive")

	ErrInvalidBucketName  = sdkerrors.Register(ModuleName, 2000, "invalid bucket name")
	ErrInvalidObjectName  = sdkerrors.Register(ModuleName, 2001, "invalid object name")
	ErrInvalidGroupName   = sdkerrors.Register(ModuleName, 2002, "invalid group name")
	ErrInvalidChcecksum   = sdkerrors.Register(ModuleName, 2003, "invalid checksum")
	ErrInvalidContentType = sdkerrors.Register(ModuleName, 2004, "invalid content type")
	ErrInvalidSPSignature = sdkerrors.Register(ModuleName, 2005, "invalid sp signature")
	ErrInvalidSPAddress   = sdkerrors.Register(ModuleName, 2006, "invalid sp address")
)
