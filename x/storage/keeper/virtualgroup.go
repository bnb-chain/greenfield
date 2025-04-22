package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"

	gnfdtypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (k Keeper) DeleteObjectFromVirtualGroup(ctx sdk.Context, bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo) error {

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)

	lvg := internalBucketInfo.MustGetLVG(objectInfo.LocalVirtualGroupId)

	gvg, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
	if !found {
		ctx.Logger().Error("GVG Not Exist, bucketID: %s, gvgID: %d, lvgID :%d", bucketInfo.Id.String(), lvg.GlobalVirtualGroupId, lvg.Id)
		return vgtypes.ErrGVGNotExist
	}

	lvg.StoredSize -= objectInfo.PayloadSize
	gvg.StoredSize -= objectInfo.PayloadSize

	// delete lvg when total charge size is 0
	if lvg.TotalChargeSize == 0 {
		if lvg.StoredSize != 0 {
			panic("The store size is non-zero when total charge size is zero.")
		}
		internalBucketInfo.DeleteLVG(lvg.Id)
		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventDeleteLocalVirtualGroup{
			Id:       lvg.Id,
			BucketId: bucketInfo.Id,
		}); err != nil {
			return err
		}
	}

	if err := k.virtualGroupKeeper.SetGVGAndEmitUpdateEvent(ctx, gvg); err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventUpdateLocalVirtualGroup{
		Id:                   lvg.Id,
		BucketId:             bucketInfo.Id,
		GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
		StoredSize:           lvg.StoredSize,
	}); err != nil {
		return err
	}

	k.SetInternalBucketInfo(ctx, bucketInfo.Id, internalBucketInfo)
	return nil
}

func (k Keeper) RebindingVirtualGroup(ctx sdk.Context, bucketInfo *types.BucketInfo, internalBucketInfo *types.InternalBucketInfo, gvgMappings []*types.GVGMapping) error {
	gvgsMap := make(map[uint32]uint32)
	for _, mapping := range gvgMappings {
		gvgsMap[mapping.SrcGlobalVirtualGroupId] = mapping.DstGlobalVirtualGroupId
	}

	for _, lvg := range internalBucketInfo.LocalVirtualGroups {
		dstGVGID, exist := gvgsMap[lvg.GlobalVirtualGroupId]
		if !exist {
			return types.ErrVirtualGroupOperateFailed.Wrapf("global virtual group not found in GVGMapping of  message. ID: %d", lvg.GlobalVirtualGroupId)
		}

		dstGVG, found := k.virtualGroupKeeper.GetGVG(ctx, dstGVGID)
		if !found {
			return types.ErrVirtualGroupOperateFailed.Wrapf("dst global virtual group not found in blockchain state. ID: %d", dstGVGID)
		}

		srcGVG, found := k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)
		if !found {
			return types.ErrVirtualGroupOperateFailed.Wrapf("src global virtual group not found in blockchain state. ID: %d", lvg.GlobalVirtualGroupId)
		}

		err := k.virtualGroupKeeper.SettleAndDistributeGVG(ctx, srcGVG)
		if err != nil {
			return types.ErrVirtualGroupOperateFailed.Wrapf("fail to settle gvg. ID: %d", srcGVG.Id)
		}

		srcGVG.StoredSize -= lvg.StoredSize
		dstGVG.StoredSize += lvg.StoredSize

		lvg.GlobalVirtualGroupId = dstGVGID

		if err := k.virtualGroupKeeper.SetGVGAndEmitUpdateEvent(ctx, srcGVG); err != nil {
			return types.ErrVirtualGroupOperateFailed.Wrapf("fail to set src gvg. ID: %d, err: %s", srcGVG.Id, err)
		}
		if err := k.virtualGroupKeeper.SetGVGAndEmitUpdateEvent(ctx, dstGVG); err != nil {
			return types.ErrVirtualGroupOperateFailed.Wrapf("fail to set dst gvg. ID: %d, err: %s", dstGVG.Id, err)
		}

		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventUpdateLocalVirtualGroup{
			Id:                   lvg.Id,
			BucketId:             bucketInfo.Id,
			GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
			StoredSize:           lvg.StoredSize,
		}); err != nil {
			return err
		}
	}

	err := k.ChargeBucketReadStoreFee(ctx, bucketInfo, internalBucketInfo)
	if err != nil {
		return types.ErrMigrationBucketFailed.Wrapf("charge bucket failed, err: %s", err)
	}

	k.SetInternalBucketInfo(ctx, bucketInfo.Id, internalBucketInfo)

	return nil
}

func (k Keeper) VerifyGVGSecondarySPsBlsSignature(ctx sdk.Context, gvg *vgtypes.GlobalVirtualGroup, signHash [32]byte, signature []byte) error {
	secondarySpBlsPubKeys := make([]bls.PublicKey, 0, len(gvg.SecondarySpIds))
	for _, spId := range gvg.GetSecondarySpIds() {
		secondarySp, found := k.spKeeper.GetStorageProvider(ctx, spId)
		if !found {
			panic("should not happen")
		}
		spBlsPubKey, err := bls.PublicKeyFromBytes(secondarySp.BlsKey)
		if err != nil {
			return types.ErrInvalidBlsPubKey.Wrapf("BLS public key converts failed: %v", err)
		}
		secondarySpBlsPubKeys = append(secondarySpBlsPubKeys, spBlsPubKey)
	}
	return gnfdtypes.VerifyBlsAggSignature(secondarySpBlsPubKeys, signHash, signature)
}

func (k Keeper) SealObjectOnVirtualGroup(ctx sdk.Context, bucketInfo *types.BucketInfo, gvgID uint32, objectInfo *types.ObjectInfo) (*types.LocalVirtualGroup, error) {
	gvg, err := k.virtualGroupKeeper.GetGlobalVirtualGroupIfAvailable(ctx, gvgID, objectInfo.PayloadSize)
	if err != nil {
		return nil, err
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)

	lvg, found := internalBucketInfo.GetLVGByGVGID(gvgID)
	if !found {
		if len(internalBucketInfo.LocalVirtualGroups) > int(k.MaxLocalVirtualGroupNumPerBucket(ctx)) {
			return nil, types.ErrVirtualGroupOperateFailed.Wrapf("The local virtual groups binding on bucket are exceed limitation. Num: %d, Allows: %d", len(internalBucketInfo.LocalVirtualGroups), k.MaxLocalVirtualGroupNumPerBucket(ctx))
		}
		// create a new lvg and add to the internalBucketInfo
		lvg = &types.LocalVirtualGroup{
			Id:                   internalBucketInfo.GetMaxLVGID() + 1,
			GlobalVirtualGroupId: gvg.Id,
			StoredSize:           0,
		}
		internalBucketInfo.AppendLVG(lvg)
	}

	lvg.StoredSize += objectInfo.PayloadSize
	gvg.StoredSize += objectInfo.PayloadSize
	objectInfo.LocalVirtualGroupId = lvg.Id

	if objectInfo.PayloadSize == 0 {
		// unlock and charge store fee
		err = k.ChargeObjectStoreFee(ctx, gvg.PrimarySpId, bucketInfo, internalBucketInfo, objectInfo)
		if err != nil {
			return nil, err
		}
	} else {
		// unlock and charge store fee
		err = k.UnlockAndChargeObjectStoreFee(ctx, gvg.PrimarySpId, bucketInfo, internalBucketInfo, objectInfo)
		if err != nil {
			return nil, err
		}
	}

	if err := k.virtualGroupKeeper.SetGVGAndEmitUpdateEvent(ctx, gvg); err != nil {
		return nil, err
	}
	k.SetInternalBucketInfo(ctx, bucketInfo.Id, internalBucketInfo)

	if !found {
		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventCreateLocalVirtualGroup{
			Id:                   lvg.Id,
			BucketId:             bucketInfo.Id,
			GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
			StoredSize:           lvg.StoredSize,
		}); err != nil {
			return nil, err
		}
	} else {
		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventUpdateLocalVirtualGroup{
			Id:                   lvg.Id,
			BucketId:             bucketInfo.Id,
			GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
			StoredSize:           lvg.StoredSize,
		}); err != nil {
			return nil, err
		}
	}

	return lvg, nil
}

func (k Keeper) SealEmptyObjectOnVirtualGroup(ctx sdk.Context, bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo) (*types.LocalVirtualGroup, error) {
	family, found := k.virtualGroupKeeper.GetGVGFamily(ctx, bucketInfo.GlobalVirtualGroupFamilyId)
	if !found {
		return nil, vgtypes.ErrGVGFamilyNotExist
	}

	if len(family.GlobalVirtualGroupIds) == 0 {
		return nil, vgtypes.ErrGVGNotExist.Wrapf("The gvg family has no gvg")
	}

	// use the first gvg by default.
	gvgID := family.GlobalVirtualGroupIds[0]

	return k.SealObjectOnVirtualGroup(ctx, bucketInfo, gvgID, objectInfo)
}

func (k Keeper) GetObjectGVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32) (*vgtypes.GlobalVirtualGroup, bool) {
	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketID)

	lvg, found := internalBucketInfo.GetLVG(lvgID)
	if !found {
		return nil, false
	}

	return k.virtualGroupKeeper.GetGVG(ctx, lvg.GlobalVirtualGroupId)

}
