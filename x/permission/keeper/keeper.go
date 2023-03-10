package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramstore    paramtypes.Subspace
		accountKeeper types.AccountKeeper

		// policy sequence
		policySeq sequence.U256
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	accountKeeper types.AccountKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		accountKeeper: accountKeeper,
	}
	k.policySeq = sequence.NewSequence256(types.PolicySequencePrefix)
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error {
	store := ctx.KVStore(k.storeKey)
	policy, found := k.getPolicyToAccount(ctx, groupID, resource.RESOURCE_TYPE_GROUP, member)
	if !found {
		policy = types.NewDefaultPolicyForGroupMember(groupID, member)
		policy.Id = k.policySeq.NextVal(store)
	} else {
		if policy.MemberStatement != nil {
			return storagetypes.ErrGroupMemberAlreadyExists
		}
		policy.MemberStatement = types.NewMemberStatement()
	}

	k.setPolicyToAccount(ctx, groupID, resource.RESOURCE_TYPE_GROUP, member, policy)
	return nil
}

func (k Keeper) RemoveGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)

	policy, found := k.getPolicyToAccount(ctx, groupID, resource.RESOURCE_TYPE_GROUP, member)
	if found {
		if policy.Statements == nil {
			store.Delete(types.GetPolicyByIDKey(policy.Id))
			store.Delete(types.GetPolicyToAccountKey(groupID, resource.RESOURCE_TYPE_GROUP, member))
		} else {
			policy.MemberStatement = nil
			store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
		}
	}
}

func (k Keeper) updatePolicy(ctx sdk.Context, policy *types.Policy, newPolicy *types.Policy) {
	store := ctx.KVStore(k.storeKey)
	policy.Statements = newPolicy.Statements
	store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
}

func (k Keeper) PutPolicy(ctx sdk.Context, policy *types.Policy) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	if policy.Principal.Type == types.TYPE_GNFD_ACCOUNT {
		policyKey := types.GetPolicyToAccountKey(policy.ResourceId, policy.ResourceType,
			policy.Principal.MustGetAccountAddress())
		bz := store.Get(policyKey)
		if bz != nil {
			id := sequence.DecodeSequence(bz)
			k.updatePolicy(ctx, k.MustGetPolicyByID(ctx, id), policy)
		} else {
			policy.Id = k.policySeq.NextVal(store)
			store.Set(policyKey, sequence.EncodeSequence(policy.Id))
			bz := k.cdc.MustMarshal(policy)
			store.Set(types.GetPolicyByIDKey(policy.Id), bz)
		}
	} else if policy.Principal.Type == types.TYPE_GNFD_GROUP {
		policyGroupKey := types.GetPolicyToGroupKey(policy.ResourceId, policy.ResourceType)
		bz := store.Get(policyGroupKey)
		if bz != nil {
			policyGroup := types.PolicyGroup{}
			k.cdc.MustUnmarshal(bz, &policyGroup)
			if (uint64)(len(policyGroup.Items)) >= k.MaximumGroupNum(ctx) {
				return math.ZeroUint(), types.ErrLimitExceeded.Wrapf("group number limit to %d, actual %d",
					k.MaximumGroupNum(ctx),
					len(policyGroup.Items))
			}
			isFound := false
			for i := 0; i < len(policyGroup.Items); i++ {
				if policyGroup.Items[i].GroupId.Equal(policy.Principal.MustGetGroupID()) {
					// override write
					k.updatePolicy(ctx, k.MustGetPolicyByID(ctx, policyGroup.Items[i].PolicyId), policy)
					isFound = true
				}
			}
			if !isFound {
				policy.Id = k.policySeq.NextVal(store)
				policyGroup.Items = append(policyGroup.Items, &types.PolicyGroup_Item{PolicyId: policy.Id,
					GroupId: policy.Principal.MustGetGroupID()})
				store.Set(policyGroupKey, k.cdc.MustMarshal(&policyGroup))
				store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
			}
		} else {
			policy.Id = k.policySeq.NextVal(store)
			policyGroup := types.PolicyGroup{}
			policyGroup.Items = append(policyGroup.Items, &types.PolicyGroup_Item{PolicyId: policy.Id,
				GroupId: policy.Principal.MustGetGroupID()})
			store.Set(policyGroupKey, k.cdc.MustMarshal(&policyGroup))
			store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
		}
	} else {
		return math.ZeroUint(), types.ErrInvalidPrincipal.Wrap("Unknown principal type.")
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

func (k Keeper) getPolicyToAccount(ctx sdk.Context, resourceID math.Uint,
	resourceType resource.ResourceType, addr sdk.AccAddress) (policy *types.Policy,
	isFound bool) {
	store := ctx.KVStore(k.storeKey)
	policyKey := types.GetPolicyToAccountKey(resourceID, resourceType, addr)

	bz := store.Get(policyKey)
	if bz == nil {
		return policy, false
	}

	return k.GetPolicyByID(ctx, sequence.DecodeSequence(bz))
}

func (k Keeper) setPolicyToAccount(ctx sdk.Context, resourceID math.Uint,
	resourceType resource.ResourceType, addr sdk.AccAddress, policy *types.Policy) {
	store := ctx.KVStore(k.storeKey)
	policyKey := types.GetPolicyToAccountKey(resourceID, resourceType, addr)

	store.Set(policyKey, sequence.EncodeSequence(policy.Id))
	store.Set(types.GetPolicyByIDKey(policy.Id), k.cdc.MustMarshal(policy))
}

func (k Keeper) VerifyPolicy(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType,
	operator sdk.AccAddress, action types.ActionType, resource *string) types.Effect {
	// verify policy which grant permission to account
	policy, found := k.getPolicyToAccount(ctx, resourceID, resourceType, operator)
	if found {
		effect := policy.Eval(action, resource)
		if effect != types.EFFECT_PASS {
			return effect
		}
	}

	// verify policy which grant permission to group
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPolicyToGroupKey(resourceID, resourceType))
	policyGroup := types.PolicyGroup{}
	k.cdc.MustUnmarshal(bz, &policyGroup)

	allowed := false
	for _, item := range policyGroup.Items {
		// check the group has the right permission of this resource
		p := k.MustGetPolicyByID(ctx, item.PolicyId)
		effect := p.Eval(action, resource)
		if effect != types.EFFECT_PASS {
			// check the operator is the member of this group
			groupPolicy, found := k.getPolicyToAccount(ctx, item.GroupId, resourceType, operator)
			if found {
				memberEffect := groupPolicy.Eval(types.ACTION_GROUP_MEMBER, nil)
				if memberEffect != types.EFFECT_PASS {
					if effect == types.EFFECT_ALLOW && memberEffect == types.EFFECT_ALLOW {
						allowed = true
					} else if effect == types.EFFECT_DENY && memberEffect == types.EFFECT_ALLOW {
						return types.EFFECT_DENY
					}
				}

			}
		}
	}
	if allowed {
		return types.EFFECT_ALLOW
	}
	return types.EFFECT_PASS
}

func (k Keeper) DeletePolicy(ctx sdk.Context, principal *types.Principal, resourceType resource.ResourceType,
	resourceID math.Uint) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	var policyID math.Uint
	if principal.Type == types.TYPE_GNFD_ACCOUNT {
		accAddr := sdk.MustAccAddressFromHex(principal.Value)
		policyKey := types.GetPolicyToAccountKey(resourceID, resourceType, accAddr)
		bz := store.Get(policyKey)
		policyID = sequence.DecodeSequence(bz)
		if bz != nil {
			store.Delete(policyKey)
			store.Delete(types.GetPolicyByIDKey(policyID))
		}
	} else if principal.Type == types.TYPE_GNFD_GROUP {
		groupID, err := principal.GetGroupID()
		if err != nil {
			return math.ZeroUint(), err
		}
		bz := store.Get(types.GetPolicyToGroupKey(resourceID, resourceType))
		if bz != nil {
			policyGroup := types.PolicyGroup{}
			k.cdc.MustUnmarshal(bz, &policyGroup)
			for i := 0; i < len(policyGroup.Items); i++ {
				if policyGroup.Items[i].GroupId.Equal(groupID) {
					// delete this item
					policyID = policyGroup.Items[i].PolicyId
					policyGroup.Items = append(policyGroup.Items[:i], policyGroup.Items[i+1:]...)
				}
			}
		}
	} else {
		return math.ZeroUint(), types.ErrInvalidPrincipal.Wrap("Unknown principal type.")
	}
	return policyID, nil
}
