package keeper

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types2 "github.com/bnb-chain/greenfield/types"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	PublicReadBucketAllowedActions = map[permtypes.ActionType]bool{
		permtypes.ACTION_GET_OBJECT:     true,
		permtypes.ACTION_COPY_OBJECT:    true,
		permtypes.ACTION_EXECUTE_OBJECT: true,
		permtypes.ACTION_LIST_OBJECT:    true,
	}
	PublicReadObjectAllowedActions = map[permtypes.ActionType]bool{
		permtypes.ACTION_GET_OBJECT:     true,
		permtypes.ACTION_COPY_OBJECT:    true,
		permtypes.ACTION_EXECUTE_OBJECT: true,
	}
)

// VerifyBucketPermission Bucket permissions checks are divided into three steps:
// First, if the bucket is a public bucket and the action is a read-only action, it returns "allow".
// Second, if the operator is the owner of the bucket, it returns "allow", as the owner has the highest permission.
// Third, verify the policy corresponding to the bucket and the operator.
//  1. If the policy is evaluated as "allow", return "allow" to the user.
//  2. If it is evaluated as "deny" or "unspecified", return "deny".
func (k Keeper) VerifyBucketPermission(ctx sdk.Context, bucketInfo *types.BucketInfo, operator sdk.AccAddress,
	action permtypes.ActionType, options *permtypes.VerifyOptions) permtypes.Effect {
	// if bucket is public, anyone can read but can not write it.
	if bucketInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ && PublicReadBucketAllowedActions[action] {
		return permtypes.EFFECT_ALLOW
	}
	// if the operator is empty(may anonymous user), don't need check policy
	if operator.Empty() {
		return permtypes.EFFECT_DENY
	}
	// The owner has full permissions
	if operator.Equals(sdk.MustAccAddressFromHex(bucketInfo.Owner)) {
		return permtypes.EFFECT_ALLOW
	}
	// verify policy
	effect := k.VerifyPolicy(ctx, bucketInfo.Id, gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, options)
	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW
	}
	return permtypes.EFFECT_DENY
}

// VerifyObjectPermission Object permission checks are divided into four steps:
// First, if the object is a public object and the action is a read-only action, it returns "allow".
// Second, if the operator is the owner of the bucket, it returns "allow"
// Third, verify the policy corresponding to the bucket and the operator
//  1. If it is evaluated as "deny", return "deny"
//  2. If it is evaluated as "allow" or "unspecified", go ahead (Noted as EffectBucket)
//
// Four, verify the policy corresponding to the object and the operator
//  1. If it is evaluated as "deny", return "deny".
//  2. If it is evaluated as "allow", return "allow".
//  3. If it is evaluated as "unspecified", then if the EffectBucket is "allow", return allow
//  4. If it is evaluated as "unspecified", then if the EffectBucket is "unspecified", return deny
func (k Keeper) VerifyObjectPermission(ctx sdk.Context, bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo,
	operator sdk.AccAddress, action permtypes.ActionType) permtypes.Effect {
	// anyone can read but can not write it when the following case: 1) object is public 2) object is inherit, only when bucket is public
	visibility := false
	if objectInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ ||
		(objectInfo.Visibility == storagetypes.VISIBILITY_TYPE_INHERIT && bucketInfo.Visibility == storagetypes.VISIBILITY_TYPE_PUBLIC_READ) {
		visibility = true
	}
	if visibility && PublicReadObjectAllowedActions[action] {
		return permtypes.EFFECT_ALLOW
	}

	// if the operator is empty(may anonymous user), don't need check policy
	if operator.Empty() {
		return permtypes.EFFECT_DENY
	}
	// The owner has full permissions
	ownerAcc := sdk.MustAccAddressFromHex(objectInfo.Owner)
	if ownerAcc.Equals(operator) {
		return permtypes.EFFECT_ALLOW
	}

	// verify policy
	opts := &permtypes.VerifyOptions{
		Resource: types2.NewObjectGRN(objectInfo.BucketName, objectInfo.ObjectName).String(),
	}
	bucketEffect := k.VerifyPolicy(ctx, bucketInfo.Id, gnfdresource.RESOURCE_TYPE_BUCKET, operator, action, opts)
	if bucketEffect == permtypes.EFFECT_DENY {
		return permtypes.EFFECT_DENY
	}

	objectEffect := k.VerifyPolicy(ctx, objectInfo.Id, gnfdresource.RESOURCE_TYPE_OBJECT, operator, action,
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
	effect := k.VerifyPolicy(ctx, groupInfo.Id, gnfdresource.RESOURCE_TYPE_GROUP, operator, action, nil)
	if effect == permtypes.EFFECT_ALLOW {
		return permtypes.EFFECT_ALLOW
	}

	return permtypes.EFFECT_DENY
}

func (k Keeper) VerifyPolicy(ctx sdk.Context, resourceID math.Uint, resourceType gnfdresource.ResourceType,
	operator sdk.AccAddress, action permtypes.ActionType, opts *permtypes.VerifyOptions) permtypes.Effect {
	// verify policy which grant permission to account
	policy, found := k.permKeeper.GetPolicyForAccount(ctx, resourceID, resourceType, operator)
	if found {
		effect, newPolicy := policy.Eval(action, ctx.BlockTime(), opts)
		k.Logger(ctx).Info(fmt.Sprintf("CreateObject LimitSize update: %s, effect: %s, ctx.TxBytes : %d",
			newPolicy.String(), effect, ctx.TxSize()))
		if effect != permtypes.EFFECT_UNSPECIFIED {
			if effect == permtypes.EFFECT_ALLOW && action == permtypes.ACTION_CREATE_OBJECT && newPolicy != nil && ctx.TxBytes() != nil {
				_, err := k.permKeeper.PutPolicy(ctx, newPolicy)
				if err != nil {
					panic(fmt.Sprintf("Update policy error, %s", err))
				}
			}
			return effect
		}
	}

	// verify policy which grant permission to group
	policyGroup, found := k.permKeeper.GetPolicyGroupForResource(ctx, resourceID, resourceType)
	if found {
		allowed := false
		var allowedPolicy *permtypes.Policy
		for _, item := range policyGroup.Items {
			if !k.hasGroup(ctx, item.GroupId) {
				continue
			}
			// check the group has the right permission of this resource
			p := k.permKeeper.MustGetPolicyByID(ctx, item.PolicyId)
			effect, newPolicy := p.Eval(action, ctx.BlockTime(), opts)
			if effect != permtypes.EFFECT_UNSPECIFIED {
				// check the operator is the member of this group
				groupMember, memberFound := k.permKeeper.GetGroupMember(ctx, item.GroupId, operator)
				if memberFound && groupMember.ExpirationTime.After(ctx.BlockTime().UTC()) {
					if effect == permtypes.EFFECT_ALLOW {
						allowed = true
						allowedPolicy = newPolicy
					} else if effect == permtypes.EFFECT_DENY {
						return permtypes.EFFECT_DENY
					}
				}
			}
		}
		if allowed {
			if action == permtypes.ACTION_CREATE_OBJECT && allowedPolicy != nil && ctx.TxBytes() != nil {
				_, err := k.permKeeper.PutPolicy(ctx, allowedPolicy)
				if err != nil {
					panic(fmt.Sprintf("Update policy error, %s", err))
				}
			}
			return permtypes.EFFECT_ALLOW
		}
	}
	return permtypes.EFFECT_UNSPECIFIED
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
	if principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT {
		policy, found = k.permKeeper.GetPolicyForAccount(ctx, resID, grn.ResourceType(),
			principal.MustGetAccountAddress())
	} else if principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP {
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

		if policy.Principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP {
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
	k.normalizePrincipal(ctx, policy.Principal)
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

func (k Keeper) normalizePrincipal(ctx sdk.Context, principal *permtypes.Principal) {
	if principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP {
		if _, err := math.ParseUint(principal.Value); err == nil {
			return
		}
		var grn types2.GRN
		if err := grn.ParseFromString(principal.Value, false); err != nil {
			return
		}
		groupOwner, groupName, err := grn.GetGroupOwnerAndAccount()
		if err != nil {
			return
		}

		if groupInfo, found := k.GetGroupInfo(ctx, groupOwner, groupName); found {
			principal.Value = groupInfo.Id.String()
		}
	}
}

func (k Keeper) validatePrincipal(ctx sdk.Context, resOwner sdk.AccAddress, principal *permtypes.Principal) error {
	if principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_ACCOUNT {
		principalAccAddress, err := principal.GetAccountAddress()
		if err != nil {
			return err
		}
		if principalAccAddress.Equals(resOwner) {
			return gnfderrors.ErrInvalidPrincipal.Wrapf("principal account can not be the bucket owner")
		}
	} else if principal.Type == permtypes.PRINCIPAL_TYPE_GNFD_GROUP {
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
