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
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
		storageKeeper types.StorageKeeper
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

func (k *Keeper) SetStorageKeeper(storageKeeper types.StorageKeeper) {
	k.storageKeeper = storageKeeper
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

func (k Keeper) SetGVGAndEmitUpdateEvent(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error {
	k.SetGVG(ctx, gvg)
	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGlobalVirtualGroup{
		Id:             gvg.Id,
		PrimarySpId:    gvg.PrimarySpId,
		StoreSize:      gvg.StoredSize,
		TotalDeposit:   gvg.TotalDeposit,
		SecondarySpIds: gvg.SecondarySpIds,
	}); err != nil {
		return err
	}
	return nil
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

	gvgFamily, found := k.GetGVGFamily(ctx, gvg.FamilyId)
	if !found {
		panic("not found gvg family when delete gvg")
	}

	gvgFamily.MustRemoveGVG(gvg.Id)

	// update stat
	stat := k.MustGetGVGStatisticsWithinSP(ctx, gvgFamily.PrimarySpId)
	stat.PrimaryCount--
	k.SetGVGStatisticsWithSP(ctx, stat)

	for _, secondarySPID := range gvg.SecondarySpIds {
		gvgStatisticsWithinSP := k.MustGetGVGStatisticsWithinSP(ctx, secondarySPID)
		gvgStatisticsWithinSP.SecondaryCount--
		k.SetGVGStatisticsWithSP(ctx, gvgStatisticsWithinSP)
	}

	store.Delete(types.GetGVGKey(gvg.Id))
	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGlobalVirtualGroup{
		Id:          gvg.Id,
		PrimarySpId: gvg.PrimarySpId,
	}); err != nil {
		return err
	}

	if len(gvgFamily.GlobalVirtualGroupIds) == 0 &&
		k.paymentKeeper.IsEmptyNetFlow(ctx, sdk.MustAccAddressFromHex(gvgFamily.VirtualPaymentAddress)) &&
		!ctx.IsUpgraded(upgradetypes.Manchurian) {
		store.Delete(types.GetGVGFamilyKey(gvg.FamilyId))
		if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGlobalVirtualGroupFamily{
			Id:          gvgFamily.Id,
			PrimarySpId: gvgFamily.PrimarySpId,
		}); err != nil {
			return err
		}
	} else {
		if err := k.SetGVGFamilyAndEmitUpdateEvent(ctx, gvgFamily); err != nil {
			return err
		}
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

func (k Keeper) SetGVGFamilyAndEmitUpdateEvent(ctx sdk.Context, gvgFamily *types.GlobalVirtualGroupFamily) error {
	k.SetGVGFamily(ctx, gvgFamily)
	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGlobalVirtualGroupFamily{
		Id:                    gvgFamily.Id,
		PrimarySpId:           gvgFamily.PrimarySpId,
		GlobalVirtualGroupIds: gvgFamily.GlobalVirtualGroupIds,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) SetGVGFamily(ctx sdk.Context, gvgFamily *types.GlobalVirtualGroupFamily) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(gvgFamily)
	store.Set(types.GetGVGFamilyKey(gvgFamily.Id), bz)
}

func (k Keeper) GetGVGFamily(ctx sdk.Context, familyID uint32) (*types.GlobalVirtualGroupFamily, bool) {
	store := ctx.KVStore(k.storeKey)

	var gvgFamily types.GlobalVirtualGroupFamily
	bz := store.Get(types.GetGVGFamilyKey(familyID))
	if bz == nil {
		return nil, false
	}
	k.cdc.MustUnmarshal(bz, &gvgFamily)
	return &gvgFamily, true
}

func (k Keeper) GetAndCheckGVGFamilyAvailableForNewBucket(ctx sdk.Context, familyID uint32) (*types.GlobalVirtualGroupFamily, error) {
	gvgFamily, found := k.GetGVGFamily(ctx, familyID)
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

func (k Keeper) GetOrCreateEmptyGVGFamily(ctx sdk.Context, familyID uint32, primarySPID uint32) (*types.GlobalVirtualGroupFamily, error) {
	store := ctx.KVStore(k.storeKey)
	var gvgFamily types.GlobalVirtualGroupFamily
	// If familyID is not specified, a new family needs to be created
	if familyID == types.NoSpecifiedFamilyId {
		id := k.GenNextGVGFamilyID(ctx)
		gvgFamily = types.GlobalVirtualGroupFamily{
			Id:                    id,
			PrimarySpId:           primarySPID,
			VirtualPaymentAddress: k.DeriveVirtualPaymentAccount(types.GVGFamilyName, id).String(),
		}
		if ctx.IsUpgraded(upgradetypes.Serengeti) {
			gvgFamilyStatistics := k.GetOrCreateGVGFamilyStatisticsWithinSP(ctx, primarySPID)
			gvgFamilyStatistics.GlobalVirtualGroupFamilyIds = append(gvgFamilyStatistics.GlobalVirtualGroupFamilyIds, id)
			k.SetGVGFamilyStatisticsWithinSP(ctx, gvgFamilyStatistics)
		}
		return &gvgFamily, nil
	} else {
		bz := store.Get(types.GetGVGFamilyKey(familyID))
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
	stakingPrice := k.GVGStakingPerBytes(ctx)

	mustStakingTokens := stakingPrice.Mul(sdk.NewInt(int64(gvg.StoredSize)))

	return gvg.TotalDeposit.Sub(mustStakingTokens)
}

func (k Keeper) SwapAsPrimarySP(ctx sdk.Context, primarySP, successorSP *sptypes.StorageProvider, familyID uint32, swapIn bool) error {
	family, found := k.GetGVGFamily(ctx, familyID)
	if !found {
		return types.ErrGVGFamilyNotExist
	}
	if family.PrimarySpId != primarySP.Id {
		return types.ErrSwapOutFailed.Wrapf("the family(ID: %d) is not owned by primary sp(ID: %d)", family.Id, primarySP.Id)
	}

	srcStat := k.MustGetGVGStatisticsWithinSP(ctx, primarySP.Id)
	dstStat := k.GetOrCreateGVGStatisticsWithinSP(ctx, successorSP.Id)
	var srcVGFStat *types.GVGFamilyStatisticsWithinSP
	var dstVGFStat *types.GVGFamilyStatisticsWithinSP
	if ctx.IsUpgraded(upgradetypes.Serengeti) {
		srcVGFStat = k.MustGetGVGFamilyStatisticsWithinSP(ctx, primarySP.Id)
		dstVGFStat = k.GetOrCreateGVGFamilyStatisticsWithinSP(ctx, successorSP.Id)
	}

	var gvgs []*types.GlobalVirtualGroup
	for _, gvgID := range family.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return types.ErrGVGNotExist
		}
		if gvg.PrimarySpId != primarySP.Id {
			if swapIn {
				return types.ErrSwapInFailed.Wrapf(
					"the primary id (%d) in global virtual group does not match the target primary sp id (%d)", gvg.PrimarySpId, primarySP.Id)
			}
			return types.ErrSwapOutFailed.Wrapf(
				"the primary id (%d) in global virtual group is not match the primary sp id (%d)", gvg.PrimarySpId, primarySP.Id)
		}
		for _, secondarySPID := range gvg.SecondarySpIds {
			if secondarySPID == successorSP.Id {
				// the successor SP might have played as secondary SP already in this GVG
				if swapIn {
					dstStat.BreakRedundancyReqmtGvgCount++
					break
				}
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
		srcStat.PrimaryCount--
		dstStat.PrimaryCount++
	}

	family.PrimarySpId = successorSP.Id

	// settlement
	err := k.SettleAndDistributeGVGFamily(ctx, primarySP, family)
	if err != nil {
		if swapIn {
			return types.ErrSwapInFailed.Wrapf("fail to settle GVG family %d", familyID)
		}
		return types.ErrSwapOutFailed.Wrapf("fail to settle GVG family %d", familyID)
	}

	if err := k.SetGVGFamilyAndEmitUpdateEvent(ctx, family); err != nil {
		if swapIn {
			return types.ErrSwapInFailed.Wrapf("failed to set gvg family and emit update event, err: %s", err)
		}
		return types.ErrSwapOutFailed.Wrapf("failed to set gvg family and emit update event, err: %s", err)
	}

	for _, gvg := range gvgs {
		if err := k.SetGVGAndEmitUpdateEvent(ctx, gvg); err != nil {
			if swapIn {
				return types.ErrSwapInFailed.Wrapf("failed to set gvg and emit update event, err: %s", err)
			}
			return types.ErrSwapOutFailed.Wrapf("failed to set gvg and emit update event, err: %s", err)
		}
	}
	k.SetGVGStatisticsWithSP(ctx, srcStat)
	k.SetGVGStatisticsWithSP(ctx, dstStat)
	if ctx.IsUpgraded(upgradetypes.Serengeti) {
		k.DeleteSpecificGVGFamilyStatisticsFromSP(ctx, srcVGFStat.SpId, family.Id)
		dstVGFStat.GlobalVirtualGroupFamilyIds = append(dstVGFStat.GlobalVirtualGroupFamilyIds, family.Id)
		k.SetGVGFamilyStatisticsWithinSP(ctx, dstVGFStat)
	}

	return nil
}

func (k Keeper) SwapOutAsSecondarySP(ctx sdk.Context, secondarySP, successorSP *sptypes.StorageProvider, gvgID uint32) error {
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return types.ErrGVGNotExist
	}
	if gvg.PrimarySpId == successorSP.Id {
		return types.ErrSwapOutFailed.Wrapf("the successor sp(ID: %d) can not be the primary sp of gvg(%s).", successorSP.Id, gvg.String())
	}
	secondarySPFound := false
	secondarySPIndex := -1
	for i, spID := range gvg.SecondarySpIds {
		if spID == successorSP.Id {
			return types.ErrSwapOutFailed.Wrapf("the successor sp(ID: %d) can not be one of the secondary sp of gvg(%s).", successorSP.Id, gvg.String())
		}
		if spID == secondarySP.Id {
			secondarySPIndex = i
			secondarySPFound = true
		}
	}
	if !secondarySPFound {
		return types.ErrSwapOutFailed.Wrapf("The sp(ID: %d) that needs swap out is not one of the secondary sps of gvg gvg(%s).", secondarySP.Id, gvg.String())
	}
	// settlement
	err := k.SettleAndDistributeGVG(ctx, gvg)
	if err != nil {
		return types.ErrSwapOutFailed.Wrapf("fail to settle GVG %d", gvgID)
	}

	if secondarySPIndex == int(-1) {
		panic("secondary sp found but the index is not correct when swap out as secondary sp")
	}
	gvg.SecondarySpIds[secondarySPIndex] = successorSP.Id
	origin := k.MustGetGVGStatisticsWithinSP(ctx, secondarySP.Id)
	successor, found := k.GetGVGStatisticsWithinSP(ctx, successorSP.Id)
	if !found {
		successor = &types.GVGStatisticsWithinSP{StorageProviderId: successorSP.Id}
	}
	origin.SecondaryCount--
	successor.SecondaryCount++
	k.SetGVGStatisticsWithSP(ctx, origin)
	k.SetGVGStatisticsWithSP(ctx, successor)

	if err := k.SetGVGAndEmitUpdateEvent(ctx, gvg); err != nil {
		return err
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
	stat, found := k.GetGVGStatisticsWithinSP(ctx, spID)
	if found {
		if stat.PrimaryCount != 0 {
			return types.ErrSPCanNotExit.Wrapf("not swap out from all the family, stat: %s", stat.String())
		}

		if stat.SecondaryCount != 0 {
			return types.ErrSPCanNotExit.Wrapf("not swap out from all the gvgs, stat: %s", stat.String())
		}
	}
	return nil
}

func (k Keeper) GetOrCreateGVGFamilyStatisticsWithinSP(ctx sdk.Context, spID uint32) *types.GVGFamilyStatisticsWithinSP {
	store := ctx.KVStore(k.storeKey)

	ret := &types.GVGFamilyStatisticsWithinSP{
		SpId: spID,
	}
	bz := store.Get(types.GetGVGFamilyStatisticsWithinSPKey(spID))
	if bz == nil {
		return ret
	}

	k.cdc.MustUnmarshal(bz, ret)
	return ret
}

func (k Keeper) DeleteSpecificGVGFamilyStatisticsFromSP(ctx sdk.Context, spID uint32, familyID uint32) {
	gvgFamilyStatistics := k.MustGetGVGFamilyStatisticsWithinSP(ctx, spID)
	for i, id := range gvgFamilyStatistics.GlobalVirtualGroupFamilyIds {
		if id == familyID {
			gvgFamilyStatistics.GlobalVirtualGroupFamilyIds = append(gvgFamilyStatistics.GlobalVirtualGroupFamilyIds[:i], gvgFamilyStatistics.GlobalVirtualGroupFamilyIds[i+1:]...)
			k.SetGVGFamilyStatisticsWithinSP(ctx, gvgFamilyStatistics)
			break
		}
	}
}

func (k Keeper) GetGVGFamilyStatisticsWithinSP(ctx sdk.Context, spID uint32) (*types.GVGFamilyStatisticsWithinSP, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGFamilyStatisticsWithinSPKey(spID))
	if bz == nil {
		return nil, false
	}

	var ret types.GVGFamilyStatisticsWithinSP
	k.cdc.MustUnmarshal(bz, &ret)
	return &ret, true
}

func (k Keeper) MustGetGVGFamilyStatisticsWithinSP(ctx sdk.Context, spID uint32) *types.GVGFamilyStatisticsWithinSP {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGVGFamilyStatisticsWithinSPKey(spID))
	if bz == nil {
		panic("must get vgf within sp")
	}

	var ret types.GVGFamilyStatisticsWithinSP
	k.cdc.MustUnmarshal(bz, &ret)
	return &ret
}

func (k Keeper) SetGVGFamilyStatisticsWithinSP(ctx sdk.Context, vgfStatisticsWithinSP *types.GVGFamilyStatisticsWithinSP) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(vgfStatisticsWithinSP)

	store.Set(types.GetGVGFamilyStatisticsWithinSPKey(vgfStatisticsWithinSP.SpId), bz)
}

func (k Keeper) MigrateGlobalVirtualGroupFamiliesForSP(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	gvgFamilyStore := prefix.NewStore(store, types.GVGFamilyKey)

	iterator := gvgFamilyStore.Iterator(nil, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var gvgFamily types.GlobalVirtualGroupFamily
		k.cdc.MustUnmarshal(iterator.Value(), &gvgFamily)
		gvgFamilyStatistics := k.GetOrCreateGVGFamilyStatisticsWithinSP(ctx, gvgFamily.PrimarySpId)
		gvgFamilyStatistics.GlobalVirtualGroupFamilyIds = append(gvgFamilyStatistics.GlobalVirtualGroupFamilyIds, gvgFamily.Id)
		k.SetGVGFamilyStatisticsWithinSP(ctx, gvgFamilyStatistics)
	}
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
	total := gvg.TotalDeposit.Quo(k.GVGStakingPerBytes(ctx))
	if !total.IsUint64() {
		return math2.MaxUint64
	} else {
		return total.Uint64()
	}
}

func (k Keeper) GetGlobalVirtualFamilyTotalStakingAndStoredSize(ctx sdk.Context, gvgFamily *types.GlobalVirtualGroupFamily) (uint64, uint64, error) {
	var familyTotalStakingSize uint64
	var familyStoredSize uint64
	for _, gvgID := range gvgFamily.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			return 0, 0, types.ErrGVGNotExist
		}
		familyTotalStakingSize += k.GetTotalStakingStoreSize(ctx, gvg)
		familyStoredSize += gvg.GetStoredSize()
	}
	return familyTotalStakingSize, familyStoredSize, nil
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

func (k Keeper) SetSwapOutInfo(ctx sdk.Context, gvgFamilyID uint32, gvgIDs []uint32, spID uint32, successorSPID uint32) error {
	store := ctx.KVStore(k.storeKey)

	if gvgFamilyID != types.NoSpecifiedFamilyId {
		key := types.GetSwapOutFamilyKey(gvgFamilyID)
		found := store.Has(key)
		if found {
			return types.ErrSwapOutFailed.Wrapf("SwapOutInfo of this gvg family(ID: %d) already exist", gvgFamilyID)

		}
		swapOutInfo := &types.SwapOutInfo{
			SpId:          spID,
			SuccessorSpId: successorSPID,
		}

		bz := k.cdc.MustMarshal(swapOutInfo)
		store.Set(key, bz)
	} else {
		for _, gvgID := range gvgIDs {
			key := types.GetSwapOutGVGKey(gvgID)
			found := store.Has(key)
			if found {
				return types.ErrSwapOutFailed.Wrapf("SwapOutInfo of this gvg(ID: %d) already exist.", gvgID)
			}
			swapOutInfo := &types.SwapOutInfo{
				SpId:          spID,
				SuccessorSpId: successorSPID,
			}
			bz := k.cdc.MustMarshal(swapOutInfo)
			store.Set(key, bz)
		}
	}
	return nil
}

func (k Keeper) DeleteSwapOutInfo(ctx sdk.Context, gvgFamilyID uint32, gvgIDs []uint32, spID uint32) error {
	store := ctx.KVStore(k.storeKey)

	swapOutInfo := types.SwapOutInfo{}
	if gvgFamilyID != types.NoSpecifiedFamilyId {
		key := types.GetSwapOutFamilyKey(gvgFamilyID)
		bz := store.Get(key)
		k.cdc.MustUnmarshal(bz, &swapOutInfo)

		if swapOutInfo.SpId != spID {
			return sptypes.ErrStorageProviderNotFound.Wrapf("spID(%d) is different from the spID(%d) in swapOutInfo", spID, swapOutInfo.SpId)
		} else {
			store.Delete(key)
		}

	} else {
		for _, gvgID := range gvgIDs {
			key := types.GetSwapOutGVGKey(gvgID)
			bz := store.Get(key)
			k.cdc.MustUnmarshal(bz, &swapOutInfo)
			if swapOutInfo.SpId != spID {
				return sptypes.ErrStorageProviderNotFound.Wrapf("spID(%d) is different from the spID(%d) in swapOutInfo", spID, swapOutInfo.SpId)
			} else {
				store.Delete(key)
			}
		}
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCancelSwapOut{
		StorageProviderId:          swapOutInfo.SpId,
		GlobalVirtualGroupFamilyId: gvgFamilyID,
		GlobalVirtualGroupIds:      gvgIDs,
		SuccessorSpId:              swapOutInfo.SuccessorSpId,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CompleteSwapOut(ctx sdk.Context, gvgFamilyID uint32, gvgIDs []uint32, successorSP *sptypes.StorageProvider) error {
	store := ctx.KVStore(k.storeKey)

	swapOutInfo := types.SwapOutInfo{}
	if gvgFamilyID != types.NoSpecifiedFamilyId {
		key := types.GetSwapOutFamilyKey(gvgFamilyID)
		bz := store.Get(key)
		if bz == nil {
			return types.ErrSwapOutFailed.Wrapf("The swap info not found in blockchain.")
		}
		k.cdc.MustUnmarshal(bz, &swapOutInfo)

		if swapOutInfo.SuccessorSpId != successorSP.Id {
			return types.ErrSwapOutFailed.Wrapf("The successor sp(ID: %d) is mismatch with the specify successor sp (ID: %d)", successorSP.Id, swapOutInfo.SuccessorSpId)
		}

		sp, found := k.spKeeper.GetStorageProvider(ctx, swapOutInfo.SpId)
		if !found {
			return sptypes.ErrStorageProviderNotFound.Wrapf("The storage provider(ID: %d) not found when complete swap out.", swapOutInfo.SpId)
		}

		err := k.SwapAsPrimarySP(ctx, sp, successorSP, gvgFamilyID, false)
		if err != nil {
			return err
		}
		store.Delete(key)
	} else {
		for _, gvgID := range gvgIDs {
			key := types.GetSwapOutGVGKey(gvgID)
			bz := store.Get(key)
			if bz == nil {
				return types.ErrSwapOutFailed.Wrapf("The swap info not found in blockchain.")
			}
			k.cdc.MustUnmarshal(bz, &swapOutInfo)

			if swapOutInfo.SuccessorSpId != successorSP.Id {
				return types.ErrSwapOutFailed.Wrapf("The successor sp(ID: %d) is mismatch with the specify successor sp (ID: %d)", successorSP.Id, swapOutInfo.SuccessorSpId)
			}

			sp, found := k.spKeeper.GetStorageProvider(ctx, swapOutInfo.SpId)
			if !found {
				return sptypes.ErrStorageProviderNotFound.Wrapf("The storage provider(ID: %d) not found when complete swap out.", swapOutInfo.SpId)
			}

			err := k.SwapOutAsSecondarySP(ctx, sp, successorSP, gvgID)
			if err != nil {
				return err
			}
			store.Delete(key)
		}
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCompleteSwapOut{
		StorageProviderId:          successorSP.Id,
		SrcStorageProviderId:       swapOutInfo.SpId,
		GlobalVirtualGroupFamilyId: gvgFamilyID,
		GlobalVirtualGroupIds:      gvgIDs,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) SwapIn(ctx sdk.Context, gvgFamilyID uint32, gvgID uint32, successorSPID uint32, targetSP *sptypes.StorageProvider, expirationTime int64) error {
	// when swapIn a family as primary SP., the target sp needs to be exiting status.
	if gvgFamilyID != types.NoSpecifiedFamilyId {
		if targetSP.Status != sptypes.STATUS_GRACEFUL_EXITING && targetSP.Status != sptypes.STATUS_FORCED_EXITING {
			return sptypes.ErrStorageProviderWrongStatus.Wrapf("The target sp is not exiting, can not be swapped")
		}
		family, found := k.GetGVGFamily(ctx, gvgFamilyID)
		if !found {
			return types.ErrGVGFamilyNotExist
		}
		if family.PrimarySpId != targetSP.Id {
			return types.ErrSwapInFailed.Wrapf("the family(ID: %d) primary SP(ID: %d) does not match the target SP(ID: %d) which need to be swapped", family.Id, family.PrimarySpId, targetSP.Id)
		}
		return k.setSwapInInfo(ctx, types.GetSwapInFamilyKey(gvgFamilyID), successorSPID, targetSP.Id, expirationTime)
	}

	// swapIn GVG as secondary SP when there is secondary SP exiting
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return types.ErrGVGNotExist
	}
	if gvg.PrimarySpId == successorSPID {
		return types.ErrSwapInFailed.Wrapf("The SP(ID=%d) is already the primary SP of GVG(ID=%d)", successorSPID, gvgID)
	}
	exist := false
	for _, sspID := range gvg.GetSecondarySpIds() {
		if sspID == successorSPID {
			return types.ErrSwapInFailed.Wrapf("The sp(ID: %d) is already one of the secondary in this GVG(ID:%d)", successorSPID, gvgID)
		}
		if sspID == targetSP.Id {
			exist = true
		}
	}
	if !exist {
		return types.ErrSwapInFailed.Wrapf("The sp(ID: %d) that needs swap out is not one of the secondary sps of gvg gvg(%s).", targetSP.Id, gvg.String())
	}
	if targetSP.Status == sptypes.STATUS_GRACEFUL_EXITING || targetSP.Status == sptypes.STATUS_FORCED_EXITING {
		return k.setSwapInInfo(ctx, types.GetSwapInGVGKey(gvgID), successorSPID, targetSP.Id, expirationTime)
	}

	// swap into GVG that no SP exiting but not fulfil redundancy requirement. e.g. [1|2,3,4,5,6,1]
	breakRedundancyReqmt := false
	for _, sspID := range gvg.GetSecondarySpIds() {
		if sspID == gvg.PrimarySpId {
			breakRedundancyReqmt = true
			break
		}
	}
	if !breakRedundancyReqmt {
		return types.ErrSwapInFailed.Wrap("can not swap into GVG which all SP are unique")
	}
	return k.setSwapInInfo(ctx, types.GetSwapInGVGKey(gvgID), successorSPID, targetSP.Id, expirationTime)
}

func (k Keeper) setSwapInInfo(ctx sdk.Context, key []byte, successorSPID, targetSPID uint32, expirationTime int64) error {
	store := ctx.KVStore(k.storeKey)
	swapInInfo := &types.SwapInInfo{
		SuccessorSpId:  successorSPID,
		TargetSpId:     targetSPID,
		ExpirationTime: uint64(expirationTime),
	}
	bz := store.Get(key)
	if bz == nil {
		store.Set(key, k.cdc.MustMarshal(swapInInfo))
		return nil
	}
	curSwapInInfo := &types.SwapInInfo{}
	k.cdc.MustUnmarshal(bz, curSwapInInfo)
	if uint64(ctx.BlockTime().Unix()) < curSwapInInfo.ExpirationTime {
		return types.ErrSwapInFailed.Wrapf("already exist SP(ID=%d) reserved the swap, please re-check the GVG after timestamp %d.", curSwapInInfo.SuccessorSpId, curSwapInInfo.ExpirationTime)
	}
	// override the stale swapIn info of prev successor sp
	if curSwapInInfo.SuccessorSpId == successorSPID {
		return types.ErrSwapInFailed.Wrapf("already tried to swap in but expired")
	}
	store.Set(key, k.cdc.MustMarshal(swapInInfo))
	return nil
}

func (k Keeper) DeleteSwapInInfo(ctx sdk.Context, gvgFamilyID, gvgID uint32, successorSPID uint32) error {
	store := ctx.KVStore(k.storeKey)

	swapInInfo := types.SwapInInfo{}
	deleteSwapInfo := func(key []byte) error {
		bz := store.Get(key)
		if bz == nil {
			return types.ErrSwapInFailed.Wrapf("The swap info not found in blockchain.")
		}
		k.cdc.MustUnmarshal(bz, &swapInInfo)
		if swapInInfo.SuccessorSpId != successorSPID {
			return sptypes.ErrStorageProviderNotFound.Wrapf("spID(%d) is different from the spID(%d) in swapInInfo", successorSPID, swapInInfo.SuccessorSpId)
		}
		store.Delete(key)
		return nil
	}

	if gvgFamilyID != types.NoSpecifiedFamilyId {
		if err := deleteSwapInfo(types.GetSwapInFamilyKey(gvgFamilyID)); err != nil {
			return err
		}
	} else {
		if err := deleteSwapInfo(types.GetSwapInGVGKey(gvgID)); err != nil {
			return err
		}
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCancelSwapIn{
		StorageProviderId:          swapInInfo.SuccessorSpId,
		GlobalVirtualGroupFamilyId: gvgFamilyID,
		GlobalVirtualGroupId:       gvgID,
		TargetSpId:                 swapInInfo.TargetSpId,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CompleteSwapIn(ctx sdk.Context, gvgFamilyID uint32, gvgID uint32, successorSP *sptypes.StorageProvider) error {
	store := ctx.KVStore(k.storeKey)
	swapInInfo := types.SwapInInfo{}
	// swapIn family
	if gvgFamilyID != types.NoSpecifiedFamilyId {
		key := types.GetSwapInFamilyKey(gvgFamilyID)
		bz := store.Get(key)
		if bz == nil {
			return types.ErrSwapInFailed.Wrapf("The swap info not found in blockchain.")
		}
		k.cdc.MustUnmarshal(bz, &swapInInfo)
		if successorSP.Id != swapInInfo.SuccessorSpId {
			return types.ErrSwapInFailed.Wrapf("The SP(ID: %d) has not reserved the swap(swapInfo=%s)", successorSP.Id, swapInInfo.String())
		}
		targetPrimarySP, found := k.spKeeper.GetStorageProvider(ctx, swapInInfo.TargetSpId)
		if !found {
			return sptypes.ErrStorageProviderNotFound.Wrapf("The storage provider(ID: %d) not found when complete swap in.", swapInInfo.TargetSpId)
		}
		if err := k.SwapAsPrimarySP(ctx, targetPrimarySP, successorSP, gvgFamilyID, true); err != nil {
			return err
		}
		store.Delete(key)
	} else {
		key := types.GetSwapInGVGKey(gvgID)
		bz := store.Get(key)
		if bz == nil {
			return types.ErrSwapInFailed.Wrapf("The swap info not found in blockchain.")
		}
		k.cdc.MustUnmarshal(bz, &swapInInfo)
		if successorSP.Id != swapInInfo.SuccessorSpId {
			return types.ErrSwapInFailed.Wrapf("The sp(ID: %d) has not reserved the swap for secondary SP(ID: %d)", successorSP.Id, swapInInfo.TargetSpId)
		}
		targetSecondarySP, found := k.spKeeper.GetStorageProvider(ctx, swapInInfo.TargetSpId)
		if !found {
			return sptypes.ErrStorageProviderNotFound.Wrapf("The storage provider(ID: %d) not found when complete swap in.", swapInInfo.TargetSpId)
		}
		if err := k.completeSwapInGVG(ctx, successorSP.Id, targetSecondarySP.Id, gvgID); err != nil {
			return err
		}
		store.Delete(key)
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCompleteSwapIn{
		StorageProviderId:          successorSP.Id,
		TargetStorageProviderId:    swapInInfo.TargetSpId,
		GlobalVirtualGroupFamilyId: gvgFamilyID,
		GlobalVirtualGroupId:       gvgID,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) completeSwapInGVG(ctx sdk.Context, successorSPID, targetSecondarySPID uint32, gvgID uint32) error {
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		return types.ErrGVGNotExist
	}
	if successorSPID == gvg.PrimarySpId {
		return types.ErrSwapInFailed.Wrapf("the primary SP(ID: %d) can not swap into GVG's secondary (%s).", successorSPID, gvg.String())
	}

	// settlement
	err := k.SettleAndDistributeGVG(ctx, gvg)
	if err != nil {
		return types.ErrSwapInFailed.Wrapf("fail to settle GVG %d", gvgID)
	}
	secondarySPIndex := -1
	for i, sspID := range gvg.GetSecondarySpIds() {
		if sspID == targetSecondarySPID {
			secondarySPIndex = i
		}
	}
	if secondarySPIndex == -1 {
		panic("secondary sp found but the index is not correct when swap out as secondary sp")
	}
	gvg.SecondarySpIds[secondarySPIndex] = successorSPID
	origin := k.MustGetGVGStatisticsWithinSP(ctx, targetSecondarySPID)
	successor, found := k.GetGVGStatisticsWithinSP(ctx, successorSPID)
	if !found {
		successor = &types.GVGStatisticsWithinSP{StorageProviderId: successorSPID}
	}
	if targetSecondarySPID == gvg.PrimarySpId {
		origin.BreakRedundancyReqmtGvgCount--
	}
	origin.SecondaryCount--
	successor.SecondaryCount++
	k.SetGVGStatisticsWithSP(ctx, origin)
	k.SetGVGStatisticsWithSP(ctx, successor)
	return k.SetGVGAndEmitUpdateEvent(ctx, gvg)
}
