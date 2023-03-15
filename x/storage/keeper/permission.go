package keeper

import (
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types2 "github.com/bnb-chain/greenfield/types"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	PublicBucketAllowedActions = map[permtypes.ActionType]bool{
		permtypes.ACTION_GET_OBJECT:     true,
		permtypes.ACTION_COPY_OBJECT:    true,
		permtypes.ACTION_EXECUTE_OBJECT: true,
		permtypes.ACTION_LIST_OBJECT:    true,
	}
	PublicObjectAllowedActions = map[permtypes.ActionType]bool{
		permtypes.ACTION_GET_OBJECT:     true,
		permtypes.ACTION_COPY_OBJECT:    true,
		permtypes.ACTION_EXECUTE_OBJECT: true,
	}
)

func (k Keeper) VerifyBucketPermission(ctx sdk.Context, bucketInfo *types.BucketInfo, operator sdk.AccAddress,
	action permtypes.ActionType, resource *string) permtypes.Effect {
	// if bucket is public, anyone can read but can not write it.
	if bucketInfo.IsPublic && PublicBucketAllowedActions[action] {
		return permtypes.EFFECT_ALLOW
	}
	// The owner has full permissions
	if strings.EqualFold(bucketInfo.Owner, operator.String()) {
		return permtypes.EFFECT_ALLOW
	}
	// verify policy
	effect := k.permKeeper.VerifyPolicy(ctx, bucketInfo.Id, gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, resource)
	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW
	}
	return permtypes.EFFECT_DENY
}

func (k Keeper) VerifyObjectPermission(ctx sdk.Context, bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo,
	operator sdk.AccAddress, action permtypes.ActionType) permtypes.Effect {
	// if object is public, anyone can read but can not write it.
	if objectInfo.IsPublic && PublicObjectAllowedActions[action] {
		return permtypes.EFFECT_ALLOW
	}
	// The owner has full permissions
	if strings.EqualFold(objectInfo.Owner, operator.String()) {
		return permtypes.EFFECT_ALLOW
	}

	// verify policy
	grn := types2.NewObjectGRN(objectInfo.BucketName, objectInfo.ObjectName)
	grnString := grn.String()
	bucketEffect := k.permKeeper.VerifyPolicy(ctx, bucketInfo.Id, gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, &grnString)
	if bucketEffect == permtypes.EFFECT_DENY {
		return permtypes.EFFECT_DENY
	}

	objectEffect := k.permKeeper.VerifyPolicy(ctx, objectInfo.Id, gnfdresource.RESOURCE_TYPE_OBJECT, operator, action,
		nil)
	if objectEffect == permtypes.EFFECT_DENY {
		return permtypes.EFFECT_DENY
	}

	if bucketEffect == permtypes.EFFECT_ALLOW || objectEffect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW
	}
	return permtypes.EFFECT_DENY
}

func (k Keeper) VerifyGroupPermission(ctx sdk.Context, groupInfo *types.GroupInfo, operator sdk.AccAddress,
	action permtypes.ActionType) permtypes.Effect {
	// The owner has full permissions
	if strings.EqualFold(groupInfo.Owner, operator.String()) {
		return permtypes.EFFECT_ALLOW
	}

	// verify policy
	effect := k.permKeeper.VerifyPolicy(ctx, groupInfo.Id, gnfdresource.RESOURCE_TYPE_GROUP, operator, action, nil)
	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW
	}

	return permtypes.EFFECT_DENY
}

func (k Keeper) GetPolicy(ctx sdk.Context, grn *types2.GRN, principal *permtypes.Principal) (*permtypes.Policy, error) {
	var resID math.Uint
	switch grn.ResourceType() {
	case gnfdresource.RESOURCE_TYPE_BUCKET:
		bucketName, grnErr := grn.GetBucketName()
		if grnErr != nil {
			return nil, grnErr
		}
		bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
		if !found {
			return nil, types.ErrNoSuchBucket.Wrapf("bucketName: %s", bucketName)
		}
		resID = bucketInfo.Id
	case gnfdresource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, grnErr := grn.GetBucketAndObjectName()
		if grnErr != nil {
			return nil, grnErr
		}
		objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
		if !found {
			return nil, types.ErrNoSuchObject.Wrapf("BucketName: %s, objectName: %s", bucketName, objectName)
		}

		resID = objectInfo.Id
	case gnfdresource.RESOURCE_TYPE_GROUP:
		groupOwner, groupName, grnErr := grn.GetGroupOwnerAndAccount()
		if grnErr != nil {
			return nil, grnErr
		}
		groupInfo, found := k.GetGroupInfo(ctx, groupOwner, groupName)
		if !found {
			return nil, types.ErrNoSuchBucket.Wrapf("groupOwner: %s, groupName: %s", groupOwner.String(), groupName)
		}
		resID = groupInfo.Id
	default:
		return nil, gnfderrors.ErrInvalidGRN.Wrap("Unknown resource type in greenfield resource name")
	}

	var policy *permtypes.Policy
	var found bool
	if principal.Type == permtypes.TYPE_GNFD_ACCOUNT {
		policy, found = k.permKeeper.GetPolicyForAccount(ctx, resID, grn.ResourceType(),
			principal.MustGetAccountAddress())
	} else if principal.Type == permtypes.TYPE_GNFD_GROUP {
		policy, found = k.permKeeper.GetPolicyForGroup(ctx, resID, grn.ResourceType(), principal.MustGetGroupID())
	} else {
		return nil, permtypes.ErrInvalidPrincipal
	}

	if !found {
		return nil, types.ErrNoSuchPolicy.Wrapf("GRN: %s, principal:%s", grn.String(), principal.String())
	}
	return policy, nil
}

func (k Keeper) PutPolicy(ctx sdk.Context, operator sdk.AccAddress, grn types2.GRN,
	policy *permtypes.Policy) (math.Uint, error) {

	var resOwner sdk.AccAddress
	var resID math.Uint
	switch grn.ResourceType() {
	case gnfdresource.RESOURCE_TYPE_BUCKET:
		bucketName, grnErr := grn.GetBucketName()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchBucket.Wrapf("bucketName: %s", bucketName)
		}
		resOwner = sdk.MustAccAddressFromHex(bucketInfo.Owner)
		resID = bucketInfo.Id
	case gnfdresource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, grnErr := grn.GetBucketAndObjectName()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchObject.Wrapf("BucketName: %s, objectName: %s", bucketName, objectName)
		}

		resOwner = sdk.MustAccAddressFromHex(objectInfo.Owner)
		resID = objectInfo.Id
	case gnfdresource.RESOURCE_TYPE_GROUP:
		groupOwner, groupName, grnErr := grn.GetGroupOwnerAndAccount()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		groupInfo, found := k.GetGroupInfo(ctx, groupOwner, groupName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchBucket.Wrapf("groupOwner: %s, groupName: %s", groupOwner.String(), groupName)
		}

		resOwner = sdk.MustAccAddressFromHex(groupInfo.Owner)
		resID = groupInfo.Id

		if policy.Principal.Type == permtypes.TYPE_GNFD_GROUP {
			return math.ZeroUint(), permtypes.ErrInvalidPrincipal.Wrapf("Not allow grant group's permission to another group")
		}
	default:
		return math.ZeroUint(), gnfderrors.ErrInvalidGRN.Wrap("Unknown resource type in greenfield resource name")
	}

	if !operator.Equals(resOwner) {
		return math.ZeroUint(), types.ErrAccessDenied.Wrapf(
			"Only resource owner can put bucket policy, operator (%s), owner(%s)",
			operator.String(), resOwner.String())
	}

	err := k.validatePrincipal(ctx, resOwner, policy.Principal)
	if err != nil {
		return math.ZeroUint(), err
	}
	policy.ResourceId = resID
	return k.permKeeper.PutPolicy(ctx, policy)
}

func (k Keeper) DeletePolicy(ctx sdk.Context, operator sdk.AccAddress, principal *permtypes.Principal,
	grn types2.GRN) (math.Uint,
	error) {
	var resOwner sdk.AccAddress
	var resID math.Uint

	switch grn.ResourceType() {
	case gnfdresource.RESOURCE_TYPE_BUCKET:
		bucketName, grnErr := grn.GetBucketName()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchBucket.Wrapf("bucketName: %s", bucketName)
		}
		resOwner = sdk.MustAccAddressFromHex(bucketInfo.Owner)
		resID = bucketInfo.Id
	case gnfdresource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, grnErr := grn.GetBucketAndObjectName()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchObject.Wrapf("BucketName: %s, objectName: %s", bucketName, objectName)
		}
		resOwner = sdk.MustAccAddressFromHex(objectInfo.Owner)
		resID = objectInfo.Id
	case gnfdresource.RESOURCE_TYPE_GROUP:
		groupOwner, groupName, grnErr := grn.GetGroupOwnerAndAccount()
		if grnErr != nil {
			return math.ZeroUint(), grnErr
		}
		groupInfo, found := k.GetGroupInfo(ctx, groupOwner, groupName)
		if !found {
			return math.ZeroUint(), types.ErrNoSuchBucket.Wrapf("groupOwner: %s, groupName: %s", groupOwner.String(), groupName)
		}
		resOwner = sdk.MustAccAddressFromHex(groupInfo.Owner)
		resID = groupInfo.Id
	default:
		return math.ZeroUint(), gnfderrors.ErrInvalidGRN.Wrap("Unknown resource type in greenfield resource name")
	}

	if !operator.Equals(resOwner) {
		return math.ZeroUint(), types.ErrAccessDenied.Wrapf(
			"Only resource owner can delete bucket policy, operator (%s), owner(%s)",
			operator.String(), resOwner.String())
	}
	return k.permKeeper.DeletePolicy(ctx, principal, grn.ResourceType(), resID)

}

func (k Keeper) validatePrincipal(ctx sdk.Context, resOwner sdk.AccAddress, principal *permtypes.Principal) error {
	if principal.Type == permtypes.TYPE_GNFD_ACCOUNT {
		principalAccAddress, err := principal.GetAccountAddress()
		if err != nil {
			return err
		}
		if principalAccAddress.Equals(resOwner) {
			return gnfderrors.ErrInvalidPrincipal.Wrapf("principal account can not be the bucket owner")
		}
	} else if principal.Type == permtypes.TYPE_GNFD_GROUP {
		groupID, err := math.ParseUint(principal.Value)
		if err != nil {
			return err
		}
		_, found := k.GetGroupInfoById(ctx, groupID)
		if !found {
			return types.ErrNoSuchGroup
		}
	} else {
		return permtypes.ErrInvalidPrincipal.Wrapf("Unknown principal type.")
	}
	return nil
}
