package keeper

import (
	"encoding/binary"
	"fmt"
	math2 "math"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bnb-chain/greenfield/internal/sequence"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

	if k.paymentKeeper.IsEmptyNetFlow(ctx, sdk.MustAccAddressFromHex(gvgFamily.VirtualPaymentAddress)) {
		store.Delete(types.GetGVGFamilyKey(gvg.FamilyId))

		if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGlobalVirtualGroupFamily{
			Id:          gvgFamily.Id,
			PrimarySpId: gvgFamily.PrimarySpId,
		}); err != nil {
			return err
		}
	} else {
		k.SetGVGFamily(ctx, gvgFamily)
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

func (k Keeper) SwapOutAsPrimarySP(ctx sdk.Context, primarySP, successorSP *sptypes.StorageProvider, familyID uint32) error {

	family, found := k.GetGVGFamily(ctx, familyID)
	if !found {
		return types.ErrGVGFamilyNotExist
	}
	if family.PrimarySpId != primarySP.Id {
		return types.ErrSwapOutFailed.Wrapf("the family(ID: %d) is not owned by primary sp(ID: %d)", family.Id, primarySP.Id)
	}

	srcStat := k.MustGetGVGStatisticsWithinSP(ctx, primarySP.Id)
	dstStat := k.GetOrCreateGVGStatisticsWithinSP(ctx, successorSP.Id)

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
		srcStat.PrimaryCount--
		dstStat.PrimaryCount++
	}

	family.PrimarySpId = successorSP.Id

	// settlement
	err := k.SettleAndDistributeGVGFamily(ctx, primarySP, family)
	if err != nil {
		return types.ErrSwapOutFailed.Wrapf("fail to settle GVG family %d", familyID)
	}

	k.SetGVGFamily(ctx, family)
	for _, gvg := range gvgs {
		k.SetGVG(ctx, gvg)
	}
	k.SetGVGStatisticsWithSP(ctx, srcStat)
	k.SetGVGStatisticsWithSP(ctx, dstStat)
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
	k.SetGVG(ctx, gvg)
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

		err := k.SwapOutAsPrimarySP(ctx, sp, successorSP, gvgFamilyID)
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
