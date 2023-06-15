package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		tStoreKey storetypes.StoreKey
		authority string

		// Keepers
		spKeeper      types.SpKeeper
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		// sequence
		lvgSequence       sequence.Sequence[uint32]
		gvgSequence       sequence.Sequence[uint32]
		gvgFamilySequence sequence.Sequence[uint32]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	tStoreKey storetypes.StoreKey,
	authority string,
	spKeeper types.SpKeeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) *Keeper {

	k := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		tStoreKey:     tStoreKey,
		authority:     authority,
		spKeeper:      spKeeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}

	k.lvgSequence = sequence.NewSequence[uint32](types.LVGSequencePrefix)
	k.gvgSequence = sequence.NewSequence[uint32](types.GVGSequencePrefix)
	k.gvgFamilySequence = sequence.NewSequence[uint32](types.GVGFamilySequencePrefix)

	return &k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GenNextLVGID(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.storeKey)

	seq := k.lvgSequence.NextVal(store)
	return seq
}

func (k Keeper) GenNextGVGID(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.storeKey)

	seq := k.gvgSequence.NextVal(store)
	return seq
}

func (k Keeper) GenNextGVGFamilyID(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.storeKey)

	seq := k.gvgFamilySequence.NextVal(store)
	return seq
}

func (k Keeper) SetGVG(ctx sdk.Context, gvg *types.GlobalVirtualGroup) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(gvg)
	store.Set(types.GetGVGKey(gvg.Id), bz)
}

func (k Keeper) DeleteGVG(ctx sdk.Context, primarySpID, gvgID uint32) error {
	store := ctx.KVStore(k.storeKey)

	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return types.ErrGVGNotExist
	}

	if gvg.StoredSize != 0 {
		return types.ErrGVGNotEmpty
	}

	gvgFamily, found := k.GetGVGFamily(ctx, primarySpID, gvg.FamilyId)
	if !found {
		panic("not found gvg family when delete gvg")
	}

	err := gvgFamily.RemoveGVG(gvg.Id)
	if err == types.ErrGVGNotExist {
		panic("gvg not found in gvg family when delete gvg")
	}

	store.Delete(types.GetGVGKey(gvg.Id))
	k.SetGVGFamily(ctx, gvg.PrimarySpId, gvgFamily)
	return nil
}

func (k Keeper) GetGVG(ctx sdk.Context, gvgID uint32) (*types.GlobalVirtualGroup, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGKey(gvgID))
	if bz == nil {
		return nil, false
	}

	var gvg types.GlobalVirtualGroup
	k.cdc.MustUnmarshal(bz, &gvg)
	return &gvg, true
}

func (k Keeper) SetLVG(ctx sdk.Context, lvg *types.LocalVirtualGroup) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(lvg)
	store.Set(types.GetLVGKey(lvg.BucketId, lvg.Id), bz)
}

func (k Keeper) GetLVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32) (*types.LocalVirtualGroup, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetLVGKey(bucketID, lvgID))
	if bz == nil {
		return nil, false
	}
	var lvg types.LocalVirtualGroup
	k.cdc.MustUnmarshal(bz, &lvg)
	return &lvg, true
}

func (k Keeper) GetGVGsBindingOnBucket(ctx sdk.Context, bucketID math.Uint) (*types.GlobalVirtualGroupsBindingOnBucket, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGsBindingOnBucketKey(bucketID))
	if bz == nil {
		return nil, false
	}

	var gvgsBindingOnBucket types.GlobalVirtualGroupsBindingOnBucket
	k.cdc.MustUnmarshal(bz, &gvgsBindingOnBucket)
	return &gvgsBindingOnBucket, true
}

func (k Keeper) SetGVGsBindingOnBucket(ctx sdk.Context, gvgsBindingOnBucket *types.GlobalVirtualGroupsBindingOnBucket) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(gvgsBindingOnBucket)
	store.Set(types.GetGVGsBindingOnBucketKey(gvgsBindingOnBucket.BucketId), bz)
}

func (k Keeper) SetGVGFamily(ctx sdk.Context, primarySpID uint32, gvgFamily *types.GlobalVirtualGroupFamily) {

	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(gvgFamily)
	store.Set(types.GetGVGFamilyKey(primarySpID, gvgFamily.Id), bz)
}

func (k Keeper) GetGVGFamily(ctx sdk.Context, spID, familyID uint32) (*types.GlobalVirtualGroupFamily, bool) {
	store := ctx.KVStore(k.storeKey)

	var gvgFamily types.GlobalVirtualGroupFamily
	bz := store.Get(types.GetGVGFamilyKey(spID, familyID))
	if bz == nil {
		return nil, false
	}
	k.cdc.MustUnmarshal(bz, &gvgFamily)
	return &gvgFamily, true
}

func (k Keeper) GenerateOrSetLVGForBucket(ctx sdk.Context, bucketID math.Uint, gvgID uint32) {
}

func (k Keeper) GetOrCreateEmptyGVGFamily(ctx sdk.Context, familyID uint32, spID uint32) (*types.GlobalVirtualGroupFamily, error) {
	store := ctx.KVStore(k.storeKey)
	var gvgFamily types.GlobalVirtualGroupFamily
	if familyID == types.NoSpecifiedFamilyId {
		id := k.GenNextGVGFamilyID(ctx)
		gvgFamily = types.GlobalVirtualGroupFamily{
			Id:                    id,
			VirtualPaymentAddress: k.DeriveVirtualPaymentAccount(types.GVGFamilyName, id).String(),
		}
		return &gvgFamily, nil
	} else {
		bz := store.Get(types.GetGVGFamilyKey(spID, familyID))
		if bz == nil {
			return nil, types.ErrGVGFamilyNotExist
		}
		k.cdc.MustUnmarshal(bz, &gvgFamily)
		return &gvgFamily, nil
	}
}

func (k Keeper) DeriveVirtualPaymentAccount(groupType string, id uint32) sdk.AccAddress {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, id)

	return address.Module(types.ModuleName, append([]byte(groupType), b...))
}

func (k Keeper) GetAvailableStakingTokens(ctx sdk.Context, gvg *types.GlobalVirtualGroup) sdk.Dec {
	stakingPrice := k.GVGStakingPrice(ctx)

	mustStakingTokens := stakingPrice.MulInt64(int64(gvg.StoredSize))

	return gvg.TotalDeposit.Sub(mustStakingTokens)
}

func (k Keeper) BindingObjectToGVG(ctx sdk.Context, bucketID math.Uint, primarySPID, familyID, gvgID uint32, payloadSize uint64) (*types.LocalVirtualGroup, error) {
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	gvgFamily, found := k.GetGVGFamily(ctx, primarySPID, familyID)
	if !found {
		return nil, errors.Wrapf(types.ErrGVGFamilyNotExist, "familyID: %d, primarySPID: %d", familyID, primarySPID)
	}

	if !gvgFamily.Contains(gvg.Id) {
		return nil, types.ErrGVGNotExistInFamily
	}

	var gvgsBindingOnBucket *types.GlobalVirtualGroupsBindingOnBucket
	var lvg *types.LocalVirtualGroup
	gvgsBindingOnBucket, found = k.GetGVGsBindingOnBucket(ctx, bucketID)
	if !found {
		// Create a new key store the gvgs binding on bucket
		lvgID := k.GenNextLVGID(ctx)
		lvg = &types.LocalVirtualGroup{
			Id:                    lvgID,
			GlobalVirtualGroupId:  gvgID,
			VirtualPaymentAddress: k.DeriveVirtualPaymentAccount(types.LVGName, lvgID).String(),
			StoredSize:            payloadSize,
			BucketId:              bucketID,
		}
		gvgsBindingOnBucket = &types.GlobalVirtualGroupsBindingOnBucket{
			BucketId: bucketID,
		}
		gvgsBindingOnBucket.AppendGVGAndLVG(gvgID, lvgID)
	} else {
		lvgID := gvgsBindingOnBucket.GetLVGIDByGVGID(gvgID)
		if lvgID == 0 {
			// not exist
			lvgID = k.GenNextLVGID(ctx)
			lvg = &types.LocalVirtualGroup{
				Id:                    lvgID,
				GlobalVirtualGroupId:  gvgID,
				VirtualPaymentAddress: k.DeriveVirtualPaymentAccount(types.LVGName, lvgID).String(),
				StoredSize:            payloadSize,
				BucketId:              bucketID,
			}
			gvgsBindingOnBucket.AppendGVGAndLVG(gvgID, lvgID)
		} else {
			lvg, found = k.GetLVG(ctx, bucketID, lvgID)
			if !found {
				return nil, types.ErrLVGNotExist
			}
			lvg.StoredSize += payloadSize
		}
	}

	gvg.StoredSize += payloadSize

	k.SetGVG(ctx, gvg)
	k.SetLVG(ctx, lvg)
	k.SetGVGsBindingOnBucket(ctx, gvgsBindingOnBucket)
	return lvg, nil
}

func (k Keeper) UnBindingObjectFromLVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32, payloadSize uint64) error {
	lvg, found := k.GetLVG(ctx, bucketID, lvgID)
	if !found {
		return types.ErrLVGNotExist
	}
	gvgsBindingOnBucket, found := k.GetGVGsBindingOnBucket(ctx, bucketID)
	if !found {
		panic(fmt.Sprintf("gvgs binding on bucket mapping not found, bucketID: %s", bucketID.String()))
	}
	gvgID := gvgsBindingOnBucket.GetGVGIDByLVGID(lvgID)
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		ctx.Logger().Error("GVG Not Exist, bucketID: %s, gvgID: %d, lvgID :%d", bucketID.String(), gvgID, lvgID)
		return types.ErrGVGNotExist
	}

	lvg.StoredSize -= payloadSize
	gvg.StoredSize -= payloadSize

	k.SetLVG(ctx, lvg)
	k.SetGVG(ctx, gvg)
	return nil
}

func (k Keeper) UnBindingBucketFromGVG(ctx sdk.Context, bucketID math.Uint) error {
	store := ctx.KVStore(k.storeKey)

	gvgsBindingOnBucket, found := k.GetGVGsBindingOnBucket(ctx, bucketID)
	if !found {
		return nil
	}

	for _, lvgID := range gvgsBindingOnBucket.LocalVirtualGroupIds {
		store.Delete(types.GetLVGKey(bucketID, lvgID))
	}

	store.Delete(types.GetGVGsBindingOnBucketKey(bucketID))
	return nil
}

func (k Keeper) BindingEmptyObjectToGVG(ctx sdk.Context, bucketID math.Uint, primarySPID, familyID uint32) (*types.LocalVirtualGroup, error) {
	family, found := k.GetGVGFamily(ctx, primarySPID, familyID)
	if !found {
		return nil, types.ErrGVGFamilyNotExist
	}

	if len(family.GlobalVirtualGroupIds) == 0 {
		return nil, types.ErrGVGNotExist
	}

	gvgID := family.GlobalVirtualGroupIds[0]

	return k.BindingObjectToGVG(ctx, bucketID, primarySPID, familyID, gvgID, 0)
}

func (k Keeper) SwapOutAsPrimarySP(ctx sdk.Context, primarySPID, familyID, successorSPID uint32) error {
	store := ctx.KVStore(k.storeKey)

	family, found := k.GetGVGFamily(ctx, primarySPID, familyID)
	if !found {
		return types.ErrGVGFamilyNotExist
	}
	var gvgs []*types.GlobalVirtualGroup
	for _, gvgID := range family.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return types.ErrGVGNotExist
		}
		if gvg.PrimarySpId != primarySPID {
			return types.ErrSwapoutFailed.Wrapf(
				"the primary id (%d) in global virtual group is not match the primary sp id (%d)", gvg.PrimarySpId, primarySPID)
		}
		for _, spID := range gvg.SecondarySpIds {
			if spID == successorSPID {
				return types.ErrSwapoutFailed.Wrapf("the successor primary sp(ID: %d) can not be the secondary sp of gvg(%s).", successorSPID, gvg.String())
			}
		}
		gvg.PrimarySpId = successorSPID
		gvgs = append(gvgs, gvg)
	}
	k.SetGVGFamily(ctx, successorSPID, family)
	for _, gvg := range gvgs {
		k.SetGVG(ctx, gvg)
	}
	store.Delete(types.GetGVGFamilyKey(primarySPID, familyID))
	return nil
}

func (k Keeper) SwapOutAsSecondarySP(ctx sdk.Context, secondarySPID, successorSPID uint32, gvgIDs []uint32) error {
	var gvgs []*types.GlobalVirtualGroup
	for _, gvgID := range gvgIDs {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return types.ErrGVGNotExist
		}
		if gvg.PrimarySpId == successorSPID {
			return types.ErrSwapoutFailed.Wrapf("the successor primary sp(ID: %d) can not be the primary sp of gvg(%s).", successorSPID, gvg.String())
		}
		for i, spID := range gvg.SecondarySpIds {
			if spID == secondarySPID {
				gvg.SecondarySpIds = append(gvg.SecondarySpIds[:i], gvg.SecondarySpIds[i+1:]...)
				gvgs = append(gvgs, gvg)
			}
		}
	}
	for _, gvg := range gvgs {
		k.SetGVG(ctx, gvg)
	}
	return nil
}
