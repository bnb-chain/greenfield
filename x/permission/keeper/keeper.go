package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		accountKeeper types.AccountKeeper

		// policy sequence
		policySeq      sequence.Sequence[math.Uint]
		groupMemberSeq sequence.Sequence[math.Uint]

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,

	accountKeeper types.AccountKeeper,
	authority string,
) *Keeper {

	k := &Keeper{
		cdc:      cdc,
		storeKey: storeKey,

		accountKeeper: accountKeeper,
		authority:     authority,
	}
	k.policySeq = sequence.NewSequence[math.Uint](types.PolicySequencePrefix)
	k.groupMemberSeq = sequence.NewSequence[math.Uint](types.GroupMemberSequencePrefix)
	return k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetGroupMemberKey(groupID, member)
	if store.Has(memberKey) {
		return math.ZeroUint(), storagetypes.ErrGroupMemberAlreadyExists
	}
	groupMember := types.GroupMember{
		GroupId: groupID,
		Member:  member.String(),
	}
	id := k.groupMemberSeq.NextVal(store)
	store.Set(memberKey, id.Bytes())
	store.Set(types.GetGroupMemberByIDKey(id), k.cdc.MustMarshal(&groupMember))
	return id, nil
}

func (k Keeper) AddGroupMemberWithExpiration(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, expiration time.Time) error {
	id, err := k.AddGroupMember(ctx, groupID, member)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	// We can simply override the whole value here, because the expiration time is the only field in the value.
	// If there are more fields in the future, we should use a more sophisticated way to update the value.
	memberExtra := types.GroupMemberExtra{
		ExpirationTime: expiration,
	}
	store.Set(types.GetGroupMemberExtraKey(groupID, member), id.Bytes())
	store.Set(types.GetGroupMemberExtraByIDKey(id), k.cdc.MustMarshal(&memberExtra))
	return nil
}

func (k Keeper) UpdateGroupMemberExpiration(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, memberID math.Uint, expiration time.Time) {
	store := ctx.KVStore(k.storeKey)
	// We can simply override the whole value here, because the expiration time is the only field in the value.
	// If there are more fields in the future, we should use a more sophisticated way to update the value.
	memberExtra := types.GroupMemberExtra{
		ExpirationTime: expiration,
	}
	store.Set(types.GetGroupMemberExtraKey(groupID, member), memberID.Bytes())
	store.Set(types.GetGroupMemberExtraByIDKey(memberID), k.cdc.MustMarshal(&memberExtra))
}

func (k Keeper) RemoveGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetGroupMemberKey(groupID, member)
	bz := store.Get(memberKey)
	if bz == nil {
		return storagetypes.ErrNoSuchGroupMember
	}
	store.Delete(memberKey)
	store.Delete(types.GetGroupMemberByIDKey(k.groupMemberSeq.DecodeSequence(bz)))
	return nil
}

func (k Keeper) RemoveGroupMemberExtra(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetGroupMemberExtraKey(groupID, member)
	bz := store.Get(memberKey)
	if bz == nil {
		return storagetypes.ErrNoSuchGroupMember
	}
	store.Delete(memberKey)
	store.Delete(types.GetGroupMemberExtraByIDKey(k.groupMemberSeq.DecodeSequence(bz)))
	return nil
}

func (k Keeper) GetGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (*types.GroupMember, bool) {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetGroupMemberKey(groupID, member)
	bz := store.Get(memberKey)
	if bz == nil {
		return nil, false
	}

	return k.GetGroupMemberByID(ctx, k.groupMemberSeq.DecodeSequence(bz))
}

func (k Keeper) GetGroupMemberByID(ctx sdk.Context, groupMemberID math.Uint) (*types.GroupMember, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetGroupMemberByIDKey(groupMemberID))
	if bz == nil {
		return nil, false
	}
	var groupMember types.GroupMember
	k.cdc.MustUnmarshal(bz, &groupMember)
	groupMember.Id = groupMemberID
	return &groupMember, true
}

func (k Keeper) GetGroupMemberExtra(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (*types.GroupMemberExtra, bool) {
	store := ctx.KVStore(k.storeKey)
	memberKey := types.GetGroupMemberExtraKey(groupID, member)
	bz := store.Get(memberKey)
	if bz == nil {
		return nil, false
	}

	return k.GetGroupMemberExtraByID(ctx, k.groupMemberSeq.DecodeSequence(bz))
}

func (k Keeper) GetGroupMemberExtraByID(ctx sdk.Context, groupMemberID math.Uint) (*types.GroupMemberExtra, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetGroupMemberExtraByIDKey(groupMemberID))
	if bz == nil {
		return nil, false
	}
	var groupMember types.GroupMemberExtra
	k.cdc.MustUnmarshal(bz, &groupMember)
	return &groupMember, true
}

func (k Keeper) updatePolicy(ctx sdk.Context, policy *types.Policy, newPolicy *types.Policy) *types.Policy {
	store := ctx.KVStore(k.storeKey)
	policy.Statements = newPolicy.Statements
	policy.ExpirationTime = newPolicy.ExpirationTime
	store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
	return policy
}

func (k Keeper) PutPolicy(ctx sdk.Context, policy *types.Policy) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	var newPolicy *types.Policy
	if policy.Principal.Type == types.PRINCIPAL_TYPE_GNFD_ACCOUNT {
		policyKey := types.GetPolicyForAccountKey(policy.ResourceId, policy.ResourceType,
			policy.Principal.MustGetAccountAddress())
		bz := store.Get(policyKey)
		if bz != nil {
			id := k.policySeq.DecodeSequence(bz)
			// override write
			newPolicy = k.updatePolicy(ctx, k.MustGetPolicyByID(ctx, id), policy)
		} else {
			policy.Id = k.policySeq.NextVal(store)
			store.Set(policyKey, k.policySeq.EncodeSequence(policy.Id))
			bz := k.cdc.MustMarshal(policy)
			store.Set(types.GetPolicyByIDKey(policy.Id), bz)
			newPolicy = policy
		}
	} else if policy.Principal.Type == types.PRINCIPAL_TYPE_GNFD_GROUP {
		policyGroupKey := types.GetPolicyForGroupKey(policy.ResourceId, policy.ResourceType)
		bz := store.Get(policyGroupKey)
		if bz != nil {
			policyGroup := types.PolicyGroup{}
			k.cdc.MustUnmarshal(bz, &policyGroup)
			if (uint64)(len(policyGroup.Items)) >= k.MaximumPolicyGroupSize(ctx) {
				return math.ZeroUint(), types.ErrLimitExceeded.Wrapf("group number limit to %d, actual %d",
					k.MaximumPolicyGroupSize(ctx),
					len(policyGroup.Items))
			}
			isFound := false
			for i := 0; i < len(policyGroup.Items); i++ {
				if policyGroup.Items[i].GroupId.Equal(policy.Principal.MustGetGroupID()) {
					// override write
					newPolicy = k.updatePolicy(ctx, k.MustGetPolicyByID(ctx, policyGroup.Items[i].PolicyId), policy)
					isFound = true
				}
			}
			if !isFound {
				policy.Id = k.policySeq.NextVal(store)
				policyGroup.Items = append(policyGroup.Items, &types.PolicyGroup_Item{PolicyId: policy.Id,
					GroupId: policy.Principal.MustGetGroupID()})
				store.Set(policyGroupKey, k.cdc.MustMarshal(&policyGroup))
				store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
				newPolicy = policy
			}
		} else {
			policy.Id = k.policySeq.NextVal(store)
			policyGroup := types.PolicyGroup{}
			policyGroup.Items = append(policyGroup.Items, &types.PolicyGroup_Item{PolicyId: policy.Id,
				GroupId: policy.Principal.MustGetGroupID()})
			store.Set(policyGroupKey, k.cdc.MustMarshal(&policyGroup))
			store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
			newPolicy = policy
		}
	} else {
		return math.ZeroUint(), types.ErrInvalidPrincipal.Wrap("Unknown principal type.")
	}

	// emit PutPolicy Event
	if err := ctx.EventManager().EmitTypedEvents(&types.EventPutPolicy{
		PolicyId:       newPolicy.Id,
		Principal:      newPolicy.Principal,
		ResourceType:   newPolicy.ResourceType,
		ResourceId:     newPolicy.ResourceId,
		Statements:     newPolicy.Statements,
		ExpirationTime: newPolicy.ExpirationTime,
	}); err != nil {
		return math.ZeroUint(), err
	}
	return policy.Id, nil
}

func (k Keeper) GetPolicyByID(ctx sdk.Context, policyID math.Uint) (*types.Policy, bool) {
	store := ctx.KVStore(k.storeKey)

	policy := types.Policy{}
	bz := store.Get(types.GetPolicyByIDKey(policyID))
	if bz == nil {
		return &policy, false
	}

	k.cdc.MustUnmarshal(bz, &policy)
	return &policy, true
}

func (k Keeper) MustGetPolicyByID(ctx sdk.Context, policyID math.Uint) *types.Policy {
	policy, found := k.GetPolicyByID(ctx, policyID)
	if !found {
		panic("Must Get policy id but not found ")
	}
	return policy
}

func (k Keeper) GetPolicyForAccount(ctx sdk.Context, resourceID math.Uint,
	resourceType resource.ResourceType, addr sdk.AccAddress) (policy *types.Policy,
	isFound bool) {
	store := ctx.KVStore(k.storeKey)
	policyKey := types.GetPolicyForAccountKey(resourceID, resourceType, addr)

	bz := store.Get(policyKey)
	if bz == nil {
		return policy, false
	}

	return k.GetPolicyByID(ctx, k.policySeq.DecodeSequence(bz))
}

func (k Keeper) GetPolicyForGroup(ctx sdk.Context, resourceID math.Uint,
	resourceType resource.ResourceType, groupID math.Uint) (policy *types.Policy,
	isFound bool) {
	store := ctx.KVStore(k.storeKey)
	policyGroupKey := types.GetPolicyForGroupKey(resourceID, resourceType)
	k.Logger(ctx).Info(fmt.Sprintf("GetPolicy, resourceID: %s, groupID: %s", resourceID.String(), groupID.String()))

	bz := store.Get(policyGroupKey)
	if bz == nil {
		return policy, false
	}

	var policyGroup types.PolicyGroup
	k.cdc.MustUnmarshal(bz, &policyGroup)
	for _, item := range policyGroup.Items {
		k.Logger(ctx).Info(fmt.Sprintf("GetPolicy, policyID: %s, groupID: %s", item.PolicyId.String(), item.GroupId.String()))
		if item.GroupId.Equal(groupID) {
			return k.MustGetPolicyByID(ctx, item.PolicyId), true
		}
	}
	return nil, false
}

func (k Keeper) VerifyPolicy(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType,
	operator sdk.AccAddress, action types.ActionType, opts *types.VerifyOptions) types.Effect {
	// verify policy which grant permission to account
	policy, found := k.GetPolicyForAccount(ctx, resourceID, resourceType, operator)
	if found {
		effect, newPolicy := policy.Eval(action, ctx.BlockTime(), opts)
		k.Logger(ctx).Info(fmt.Sprintf("CreateObject LimitSize update: %s, effect: %s, ctx.TxBytes : %d",
			newPolicy.String(), effect, ctx.TxSize()))
		if effect != types.EFFECT_UNSPECIFIED {
			if effect == types.EFFECT_ALLOW && action == types.ACTION_CREATE_OBJECT && newPolicy != nil && ctx.TxBytes() != nil {
				_, err := k.PutPolicy(ctx, newPolicy)
				if err != nil {
					panic(fmt.Sprintf("Update policy error, %s", err))
				}
			}
			return effect
		}
	}

	// verify policy which grant permission to group
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPolicyForGroupKey(resourceID, resourceType))
	if bz != nil {
		policyGroup := types.PolicyGroup{}
		k.cdc.MustUnmarshal(bz, &policyGroup)
		allowed := false
		var (
			newPolicy *types.Policy
			effect    types.Effect
		)
		for _, item := range policyGroup.Items {
			// check the group has the right permission of this resource
			p := k.MustGetPolicyByID(ctx, item.PolicyId)
			effect, newPolicy = p.Eval(action, ctx.BlockTime(), opts)
			if effect != types.EFFECT_UNSPECIFIED {
				// check the operator is the member of this group
				_, memberFound := k.GetGroupMember(ctx, item.GroupId, operator)
				if memberFound {
					if effect == types.EFFECT_ALLOW {
						allowed = true
					} else if effect == types.EFFECT_DENY {
						return types.EFFECT_DENY
					}
				}
			}
		}
		if allowed {
			if action == types.ACTION_CREATE_OBJECT && newPolicy != nil && ctx.TxBytes() != nil {
				if effect == types.EFFECT_ALLOW && action == types.ACTION_CREATE_OBJECT && newPolicy != nil && ctx.TxBytes() != nil {
					_, err := k.PutPolicy(ctx, newPolicy)
					if err != nil {
						panic(fmt.Sprintf("Update policy error, %s", err))
					}
				}
			}
			return types.EFFECT_ALLOW
		}
	}

	return types.EFFECT_UNSPECIFIED
}

func (k Keeper) DeletePolicy(ctx sdk.Context, principal *types.Principal, resourceType resource.ResourceType,
	resourceID math.Uint) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	var policyID math.Uint
	if principal.Type == types.PRINCIPAL_TYPE_GNFD_ACCOUNT {
		accAddr := sdk.MustAccAddressFromHex(principal.Value)
		policyKey := types.GetPolicyForAccountKey(resourceID, resourceType, accAddr)
		bz := store.Get(policyKey)
		policyID = k.policySeq.DecodeSequence(bz)
		if bz != nil {
			store.Delete(policyKey)
			store.Delete(types.GetPolicyByIDKey(policyID))
		}
	} else if principal.Type == types.PRINCIPAL_TYPE_GNFD_GROUP {
		groupID, err := principal.GetGroupID()
		if err != nil {
			return math.ZeroUint(), err
		}
		bz := store.Get(types.GetPolicyForGroupKey(resourceID, resourceType))
		if bz != nil {
			policyGroup := types.PolicyGroup{}
			k.cdc.MustUnmarshal(bz, &policyGroup)

			for i := 0; i < len(policyGroup.Items); i++ {
				if policyGroup.Items[i].GroupId.Equal(groupID) {
					// delete this item
					policyID = policyGroup.Items[i].PolicyId
					policyGroup.Items = append(policyGroup.Items[:i], policyGroup.Items[i+1:]...)

					// delete the concrete policy
					store.Delete(types.GetPolicyByIDKey(policyID))
				}
			}
			// delete the key if value is empty
			if len(policyGroup.Items) == 0 {
				store.Delete(types.GetPolicyForGroupKey(resourceID, resourceType))
			}
		}
	} else {
		return math.ZeroUint(), types.ErrInvalidPrincipal.Wrap("Unknown principal type.")
	}
	// emit DeletePolicy Event
	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeletePolicy{
		PolicyId: policyID,
	}); err != nil {
		return math.ZeroUint(), err
	}
	return policyID, nil
}

// ForceDeleteAccountPolicyForResource deletes all individual accounts policy enforced on resources, if
func (k Keeper) ForceDeleteAccountPolicyForResource(ctx sdk.Context, maxDelete, deletedTotal uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool) {
	if resourceType == resource.RESOURCE_TYPE_UNSPECIFIED {
		return deletedTotal, true
	}
	store := ctx.KVStore(k.storeKey)
	resourceAccountsPolicyStore := prefix.NewStore(store, types.PolicyForAccountPrefix(resourceID, resourceType))
	iterator := resourceAccountsPolicyStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// delete mapping policyId -> policy
		policyID := k.policySeq.DecodeSequence(iterator.Value())
		store.Delete(types.GetPolicyByIDKey(policyID))
		// delete mapping policyKey -> policyId
		resourceAccountsPolicyStore.Delete(iterator.Key())

		// emit DeletePolicy Event
		_ = ctx.EventManager().EmitTypedEvents(&types.EventDeletePolicy{
			PolicyId: policyID,
		})
		deletedTotal++
		// if exceeding the limit, pause the GC and mark the current resource's deletion is not complete yet
		if deletedTotal > maxDelete {
			return deletedTotal, false
		}

	}
	return deletedTotal, true
}

// ForceDeleteGroupPolicyForResource deletes group policy enforced on resource
func (k Keeper) ForceDeleteGroupPolicyForResource(ctx sdk.Context, maxDelete, deletedTotal uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool) {
	if resourceType == resource.RESOURCE_TYPE_UNSPECIFIED || resourceType == resource.RESOURCE_TYPE_GROUP {
		return deletedTotal, true
	}
	policyForGroupKey := types.GetPolicyForGroupKey(resourceID, resourceType)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(policyForGroupKey)
	if bz != nil {
		policyGroup := types.PolicyGroup{}
		k.cdc.MustUnmarshal(bz, &policyGroup)
		for i := 0; i < len(policyGroup.Items); i++ {
			// delete concrete policy by policyId
			policyId := policyGroup.Items[i].PolicyId
			store.Delete(types.GetPolicyByIDKey(policyId))

			_ = ctx.EventManager().EmitTypedEvents(&types.EventDeletePolicy{
				PolicyId: policyId,
			})
			deletedTotal++
			if deletedTotal > maxDelete {
				return deletedTotal, false
			}
		}
		store.Delete(policyForGroupKey)
	}
	return deletedTotal, true
}

// ForceDeleteGroupMembers deletes group members when user deletes a group
func (k Keeper) ForceDeleteGroupMembers(ctx sdk.Context, maxDelete, deletedTotal uint64, groupId math.Uint) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	groupMembersPrefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.GroupMembersPrefix(groupId))
	iter := groupMembersPrefixStore.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		memberID := k.groupMemberSeq.DecodeSequence(iter.Value())
		// delete GroupMemberByIDPrefix_id -> groupMember
		store.Delete(types.GetGroupMemberByIDKey(memberID))
		// delete GroupMemberPrefix_groupId_memberAddr -> memberSequence(id)
		groupMembersPrefixStore.Delete(iter.Key())
		deletedTotal++
		if deletedTotal > maxDelete {
			return deletedTotal, false
		}
	}
	return deletedTotal, true
}

func (k Keeper) ExistAccountPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool {
	if resourceType == resource.RESOURCE_TYPE_UNSPECIFIED {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	resourceAccountsPolicyStore := prefix.NewStore(store, types.PolicyForAccountPrefix(resourceID, resourceType))
	iterator := resourceAccountsPolicyStore.Iterator(nil, nil)
	defer iterator.Close()
	return iterator.Valid()
}

func (k Keeper) ExistGroupPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool {
	if resourceType == resource.RESOURCE_TYPE_UNSPECIFIED || resourceType == resource.RESOURCE_TYPE_GROUP {
		return false
	}
	policyForGroupKey := types.GetPolicyForGroupKey(resourceID, resourceType)
	store := ctx.KVStore(k.storeKey)
	return store.Has(policyForGroupKey) && store.Get(policyForGroupKey) != nil
}

func (k Keeper) ExistGroupMemberForGroup(ctx sdk.Context, groupId math.Uint) bool {
	groupMembersPrefixStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.GroupMembersPrefix(groupId))
	iter := groupMembersPrefixStore.Iterator(nil, nil)
	defer iter.Close()
	return iter.Valid()
}
