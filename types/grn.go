package types

import (
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/types/s3util"
)

const (
	BucketTypeAbbr = "b"
	ObjectTypeAbbr = "o"
	GroupTypeAbbr  = "g"
)

var (
	validGRNRegex       = regexp.MustCompile("^grn:([b|o|g]):([^:]*):([^:]*)$")
	validGRNRegexNoWild = regexp.MustCompile("^grn:([b|o|g]):([^:]*):([^:*]*)$")
)

// GRN define a standard ResourceName format, full name: GreenFieldResourceName
// valid format:
//
//	bucket: "grn:b::bucketName"
//	object: "grn:o::bucketName/objectName"
//	group: "grn:g:ownerAddress/groupName"
//
// Notice: all the name support wildcards
type GRN struct {
	resType    resource.ResourceType
	groupOwner sdk.AccAddress
	name       string // can be bucketName, bucketName/objectName, groupName
}

// NewBucketGRN use to generate a bucket resource with specify bucketName
// The bucketName support wildcards. E.g. samplebucket, sample*bucket, sample?bucket
func NewBucketGRN(bucketName string) *GRN {
	return &GRN{resType: resource.RESOURCE_TYPE_BUCKET, name: bucketName}
}

func NewObjectGRN(bucketName, objectName string) *GRN {
	name := strings.Join([]string{bucketName, objectName}, "/")
	return &GRN{resType: resource.RESOURCE_TYPE_OBJECT, name: name}
}

func NewGroupGRN(owner sdk.AccAddress, groupName string) *GRN {
	return &GRN{resType: resource.RESOURCE_TYPE_GROUP, groupOwner: owner, name: groupName}
}

func (r *GRN) String() string {
	var res string
	switch r.resType {
	case resource.RESOURCE_TYPE_BUCKET:
		res = strings.Join([]string{"grn", BucketTypeAbbr, "", r.name}, ":")
	case resource.RESOURCE_TYPE_OBJECT:
		res = strings.Join([]string{"grn", ObjectTypeAbbr, "", r.name}, ":")
	case resource.RESOURCE_TYPE_GROUP:
		res = strings.Join([]string{"grn", GroupTypeAbbr, r.groupOwner.String(), r.name}, ":")
	default:
		return ""
	}
	return strings.TrimSuffix(res, ":")
}

func (r *GRN) GetBucketName() (string, error) {
	switch r.resType {
	case resource.RESOURCE_TYPE_BUCKET:
		return r.name, nil
	case resource.RESOURCE_TYPE_OBJECT:
		bucketName, _, err := r.parseBucketAndObjectName(r.name)
		if err != nil {
			return "", err
		}
		return bucketName, nil
	default:
		return "", gnfderrors.ErrGRNTypeMismatch.Wrap("Can not GetBucketName from a non bucket or object resource type")
	}
}

func (r *GRN) MustGetBucketName() string {
	bucketName, err := r.GetBucketName()
	if err != nil {
		panic(err)
	}
	return bucketName
}

func (r *GRN) GetBucketAndObjectName() (string, string, error) {
	switch r.resType {
	case resource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, err := r.parseBucketAndObjectName(r.name)
		if err != nil {
			return "", "", err
		}
		return bucketName, objectName, nil
	default:
		return "", "", gnfderrors.ErrGRNTypeMismatch.Wrap(
			"Can not GetBucketAndObjectName from a non-object resource type")
	}
}
func (r *GRN) MustGetBucketAndObjectName() (string, string) {
	bucketName, objectName, err := r.GetBucketAndObjectName()
	if err != nil {
		panic(err)
	}
	return bucketName, objectName
}

func (r *GRN) GetGroupOwnerAndAccount() (sdk.AccAddress, string, error) {
	switch r.resType {
	case resource.RESOURCE_TYPE_GROUP:
		return r.groupOwner, r.name, nil
	default:
		return sdk.AccAddress{}, "", gnfderrors.ErrGRNTypeMismatch.Wrap(
			"Can not GetGroupOwnerAndAccount from a non-group resource type")
	}
}

func (r *GRN) MustGetGroupOwnerAndAccount() (sdk.AccAddress, string) {
	account, groupName, err := r.GetGroupOwnerAndAccount()
	if err != nil {
		panic(err)
	}
	return account, groupName
}

func (r *GRN) ResourceType() resource.ResourceType {
	return r.resType
}

func (r *GRN) ParseFromString(res string, wildcards bool) error {
	var result [][]string
	if wildcards {
		result = validGRNRegex.FindAllStringSubmatch(res, 1)
	} else {
		result = validGRNRegexNoWild.FindAllStringSubmatch(res, 1)
	}
	if result == nil || len(result) != 1 || len(result[0]) != 4 {
		return gnfderrors.ErrInvalidGRN.Wrapf("regex match error. Res: %s.", res)
	}

	var err error
	abbr := result[0][1]
	acc := result[0][2]
	name := result[0][3]
	if abbr == BucketTypeAbbr {
		if acc != "" {
			return gnfderrors.ErrInvalidGRN.Wrapf("Not allowed acc non-empty in bucket resource name")
		}
		r.resType = resource.RESOURCE_TYPE_BUCKET
		r.name = name
		if strings.Contains(name, "/") {
			return gnfderrors.ErrInvalidGRN.Wrapf("Not allowed '/' in bucket resource name")
		}
		if !wildcards {
			err = s3util.CheckValidBucketName(r.name)
			if err != nil {
				return gnfderrors.ErrInvalidGRN.Wrapf("invalid bucketName: %s, err: %s", r.name, err)
			}
		}
	} else if abbr == ObjectTypeAbbr {
		if acc != "" {
			return gnfderrors.ErrInvalidGRN.Wrapf("Not allowed acc non-empty in bucket resource name")
		}
		r.resType = resource.RESOURCE_TYPE_OBJECT
		r.name = name
		_, _, err = r.parseBucketAndObjectName(r.name)
		if err != nil {
			return gnfderrors.ErrInvalidGRN.Wrapf("invalid name, err : %s", err)
		}
	} else if abbr == GroupTypeAbbr {
		r.resType = resource.RESOURCE_TYPE_GROUP
		r.groupOwner, err = sdk.AccAddressFromHexUnsafe(acc)
		if err != nil {
			return gnfderrors.ErrInvalidGRN.Wrapf("invalid group owner account, err : %s", err)
		}
		if !wildcards {
			err = s3util.CheckValidGroupName(name)
			if err != nil {
				return gnfderrors.ErrInvalidGRN.Wrapf("invalid group name, err : %s", err)
			}
		}
		r.name = name
	}

	return nil
}

func (r *GRN) parseBucketAndObjectName(name string) (string, string, error) {
	nameArr := strings.Split(name, "/")
	if len(nameArr) != 2 {
		return "", "", gnfderrors.ErrInvalidGRN.Wrapf("expect bucketName/object, actual %s", r.name)
	}
	err := s3util.CheckValidBucketName(nameArr[0])
	if err != nil {
		return "", "", gnfderrors.ErrInvalidGRN.Wrapf("invalid bucketName, err: %s", err)
	}
	err = s3util.CheckValidObjectName(nameArr[1])
	if err != nil {
		return "", "", gnfderrors.ErrInvalidGRN.Wrapf("invalid objectName, err: %s", err)
	}
	return nameArr[0], nameArr[1], nil
}
