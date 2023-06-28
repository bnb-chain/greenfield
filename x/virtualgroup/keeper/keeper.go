package keeper

import (
	"encoding/binary"
	"fmt"
	math2 "math"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/internal/sequence"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
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
		paymentKeeper types.PaymentKeeper
		// sequence
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
	paymentKeeper types.PaymentKeeper,
) *Keeper {

	k := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		tStoreKey:     tStoreKey,
		authority:     authority,
		spKeeper:      spKeeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
		paymentKeeper: paymentKeeper,
	}

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

func (k Keeper) DeleteGVG(ctx sdk.Context, primarySp *sptypes.StorageProvider, gvgID uint32) error {

	store := ctx.KVStore(k.storeKey)

	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return types.ErrGVGNotExist
	}

	if gvg.StoredSize != 0 {
		return types.ErrGVGNotEmpty
	}

	if !k.paymentKeeper.IsEmptyNetFlow(ctx, sdk.MustAccAddressFromHex(gvg.VirtualPaymentAddress)) {
		return types.ErrGVGNotEmpty.Wrap("The virtual payment account still not empty")
	}

	if !gvg.TotalDeposit.IsZero() {
		// send back the deposit
		coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForGVG(ctx), gvg.TotalDeposit))
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.MustAccAddressFromHex(primarySp.FundingAddress), coins)
		if err != nil {
			return err
		}
	}

	gvgFamily, found := k.GetGVGFamily(ctx, primarySp.Id, gvg.FamilyId)
	if !found {
		panic("not found gvg family when delete gvg")
	}

	gvgFamily.MustRemoveGVG(gvg.Id)

	for _, secondarySPID := range gvg.SecondarySpIds {
		gvgStatisticsWithinSP := k.MustGetGVGStatisticsWithinSP(ctx, secondarySPID)
		gvgStatisticsWithinSP.SecondaryCount--
		k.SetGVGStatisticsWithSP(ctx, gvgStatisticsWithinSP)
	}

	store.Delete(types.GetGVGKey(gvg.Id))

	if len(gvgFamily.GlobalVirtualGroupIds) == 0 {
		store.Delete(types.GetGVGFamilyKey(gvg.PrimarySpId, gvgFamily.Id))
	} else {
		k.SetGVGFamily(ctx, gvg.PrimarySpId, gvgFamily)
	}
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

func (k Keeper) GetAndCheckGVGFamilyAvailableForNewBucket(ctx sdk.Context, spID, familyID uint32) (*types.GlobalVirtualGroupFamily, error) {
	gvgFamily, found := k.GetGVGFamily(ctx, spID, familyID)
	if !found {
		return nil, types.ErrGVGFamilyNotExist
	}

	// check the maximum store size for a family
	// If yes, no more buckets will be served
	storeSize := k.GetStoreSizeOfFamily(ctx, gvgFamily)
	if storeSize >= k.MaxStoreSizePerFamily(ctx) {
		return nil, types.ErrLimitationExceed.Wrapf("The storage size within the family exceeds the limit and can't serve more buckets.. Current: %d, now: %d", k.MaxStoreSizePerFamily(ctx), storeSize)
	}
	return gvgFamily, nil
}

func (k Keeper) GetOrCreateEmptyGVGFamily(ctx sdk.Context, familyID uint32, spID uint32) (*types.GlobalVirtualGroupFamily, error) {
	store := ctx.KVStore(k.storeKey)
	var gvgFamily types.GlobalVirtualGroupFamily
	// If familyID is not specified, a new family needs to be created
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

func (k Keeper) GetAvailableStakingTokens(ctx sdk.Context, gvg *types.GlobalVirtualGroup) math.Int {
	stakingPrice := k.GVGStakingPrice(ctx)

	mustStakingTokens := stakingPrice.Mul(sdk.NewInt(int64(gvg.StoredSize)))

	return gvg.TotalDeposit.Sub(mustStakingTokens)
}

func (k Keeper) SwapOutAsPrimarySP(ctx sdk.Context, primarySP, successorSP *sptypes.StorageProvider, familyID uint32) error {
	store := ctx.KVStore(k.storeKey)

	family, found := k.GetGVGFamily(ctx, primarySP.Id, familyID)
	if !found {
		return types.ErrGVGFamilyNotExist
	}
	var gvgs []*types.GlobalVirtualGroup
	for _, gvgID := range family.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return types.ErrGVGNotExist
		}
		if gvg.PrimarySpId != primarySP.Id {
			return types.ErrSwapOutFailed.Wrapf(
				"the primary id (%d) in global virtual group is not match the primary sp id (%d)", gvg.PrimarySpId, primarySP.Id)
		}
		for _, spID := range gvg.SecondarySpIds {
			if spID == successorSP.Id {
				return types.ErrSwapOutFailed.Wrapf("the successor primary sp(ID: %d) can not be the secondary sp of gvg(%s).", successorSP.Id, gvg.String())
			}
		}

		// swap deposit
		if !gvg.TotalDeposit.IsZero() {
			coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForGVG(ctx), gvg.TotalDeposit))
			err := k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromHex(successorSP.FundingAddress), sdk.MustAccAddressFromHex(primarySP.FundingAddress), coins)
			if err != nil {
				return err
			}
		}

		gvg.PrimarySpId = successorSP.Id
		gvgs = append(gvgs, gvg)
	}

	// settlement
	err := k.SettleAndDistributeGVGFamily(ctx, primarySP.Id, family)
	if err != nil {
		return types.ErrSwapOutFailed.Wrapf("fail to settle GVG family %d", familyID)
	}

	k.SetGVGFamily(ctx, successorSP.Id, family)
	for _, gvg := range gvgs {
		k.SetGVG(ctx, gvg)
	}
	store.Delete(types.GetGVGFamilyKey(primarySP.Id, familyID))
	return nil
}

func (k Keeper) SwapOutAsSecondarySP(ctx sdk.Context, secondarySPID, successorSPID uint32, gvgIDs []uint32) error {
	for _, gvgID := range gvgIDs {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return types.ErrGVGNotExist
		}
		if gvg.PrimarySpId == successorSPID {
			return types.ErrSwapOutFailed.Wrapf("the successor sp(ID: %d) can not be the primary sp of gvg(%s).", successorSPID, gvg.String())
		}
		secondarySPFound := false
		secondarySPIndex := int(-1)
		for i, spID := range gvg.SecondarySpIds {
			if spID == successorSPID {
				return types.ErrSwapOutFailed.Wrapf("the successor sp(ID: %d) can not be one of the secondary sp of gvg(%s).", successorSPID, gvg.String())
			}
			if spID == secondarySPID {
				secondarySPIndex = i
				secondarySPFound = true
			}
		}
		if !secondarySPFound {
			return types.ErrSwapOutFailed.Wrapf("The sp(ID: %d) that needs swap out is not one of the secondary sps of gvg gvg(%s).", secondarySPID, gvg.String())
		}
		// settlement
		err := k.SettleAndDistributeGVG(ctx, gvg)
		if err != nil {
			return types.ErrSwapOutFailed.Wrapf("fail to settle GVG %d", gvgID)
		}

		if secondarySPIndex == int(-1) {
			panic("secondary sp found but the index is not correct when swap out as secondary sp")
		}
		gvg.SecondarySpIds[secondarySPIndex] = successorSPID
		origin := k.MustGetGVGStatisticsWithinSP(ctx, secondarySPID)
		successor := k.MustGetGVGStatisticsWithinSP(ctx, successorSPID)
		origin.SecondaryCount--
		successor.SecondaryCount++
		k.SetGVGStatisticsWithSP(ctx, origin)
		k.SetGVGStatisticsWithSP(ctx, successor)
		k.SetGVG(ctx, gvg)
	}
	return nil
}

func (k Keeper) GetOrCreateGVGStatisticsWithinSP(ctx sdk.Context, spID uint32) *types.GVGStatisticsWithinSP {
	store := ctx.KVStore(k.storeKey)

	ret := &types.GVGStatisticsWithinSP{
		StorageProviderId: spID,
		SecondaryCount:    0,
	}
	bz := store.Get(types.GetGVGStatisticsWithinSPKey(spID))
	if bz == nil {
		return ret
	}

	k.cdc.MustUnmarshal(bz, ret)
	return ret
}

func (k Keeper) GetGVGStatisticsWithinSP(ctx sdk.Context, spID uint32) (*types.GVGStatisticsWithinSP, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGStatisticsWithinSPKey(spID))
	if bz == nil {
		return nil, false
	}

	var ret types.GVGStatisticsWithinSP
	k.cdc.MustUnmarshal(bz, &ret)
	return &ret, true
}

func (k Keeper) MustGetGVGStatisticsWithinSP(ctx sdk.Context, spID uint32) *types.GVGStatisticsWithinSP {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGStatisticsWithinSPKey(spID))
	if bz == nil {
		panic("must get gvg statistics within sp")
	}

	var ret types.GVGStatisticsWithinSP
	k.cdc.MustUnmarshal(bz, &ret)
	return &ret
}

func (k Keeper) SetGVGStatisticsWithSP(ctx sdk.Context, gvgsStatisticsWithinSP *types.GVGStatisticsWithinSP) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(gvgsStatisticsWithinSP)

	store.Set(types.GetGVGStatisticsWithinSPKey(gvgsStatisticsWithinSP.StorageProviderId), bz)
}

func (k Keeper) BatchSetGVGStatisticsWithinSP(ctx sdk.Context, gvgsStatisticsWithinSP []*types.GVGStatisticsWithinSP) {
	for _, g := range gvgsStatisticsWithinSP {
		k.SetGVGStatisticsWithSP(ctx, g)
	}
}

func (k Keeper) StorageProviderExitable(ctx sdk.Context, spID uint32) error {
	store := ctx.KVStore(k.storeKey)

	prefixStore := prefix.NewStore(store, types.GetGVGFamilyPrefixKey(spID))
	iter := prefixStore.Iterator(nil, nil)
	if iter.Valid() {
		var family types.GlobalVirtualGroupFamily
		k.cdc.MustUnmarshal(iter.Value(), &family)
		return types.ErrSPCanNotExit.Wrapf("not swap out from all the family, key: %s", family.String())
	}

	gvgStat, found := k.GetGVGStatisticsWithinSP(ctx, spID)
	if found && gvgStat.SecondaryCount != 0 {
		return types.ErrSPCanNotExit.Wrapf("not swap out from all the gvgs, remain: %d", gvgStat.SecondaryCount)
	}
	return nil
}

// GetStoreSizeOfFamily Rather than calculating the stored size of a Global Virtual Group Family (GVGF) in real-time,
// it is preferable to calculate it once during the creation of a Global Virtual Group (GVG). This approach is favored
// because GVG creation is infrequent and occurs with low frequency.
func (k Keeper) GetStoreSizeOfFamily(ctx sdk.Context, gvgFamily *types.GlobalVirtualGroupFamily) uint64 {
	var totalStoreSize uint64
	for _, gvgID := range gvgFamily.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			panic("gvg not found when get store size of family")
		}
		totalStoreSize += gvg.StoredSize
	}
	return totalStoreSize
}

func (k Keeper) GetTotalStakingStoreSize(ctx sdk.Context, gvg *types.GlobalVirtualGroup) uint64 {
	total := gvg.TotalDeposit.Quo(k.GVGStakingPrice(ctx))
	if !total.IsUint64() {
		return math2.MaxUint64
	} else {
		return total.Uint64()
	}
}

func (k Keeper) GetGlobalVirtualGroupIfAvailable(ctx sdk.Context, gvgID uint32, expectedStoreSize uint64) (*types.GlobalVirtualGroup, error) {
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	// check staking
	if gvg.StoredSize+expectedStoreSize > k.GetTotalStakingStoreSize(ctx, gvg) {
		return nil, types.ErrInsufficientStaking.Wrapf("gvg state: %s", gvg.String())
	}
	return gvg, nil
}
