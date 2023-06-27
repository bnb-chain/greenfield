package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	gnfdtypes "github.com/bnb-chain/greenfield/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

func (k Keeper) UnBindingObjectFromLVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32, payloadSize uint64) error {
	lvg, found := k.GetLVG(ctx, bucketID, lvgID)
	if !found {
		return vgtypes.ErrLVGNotExist
	}
	gvgsBindingOnBucket, found := k.GetGVGsBindingOnBucket(ctx, bucketID)
	if !found {
		panic(fmt.Sprintf("gvgs binding on bucket mapping not found, bucketID: %s", bucketID.String()))
	}
	gvgID := gvgsBindingOnBucket.GetGVGIDByLVGID(lvgID)
	gvg, found := k.GetGVG(ctx, gvgID)
	if !found {
		ctx.Logger().Error("GVG Not Exist, bucketID: %s, gvgID: %d, lvgID :%d", bucketID.String(), gvgID, lvgID)
		return vgtypes.ErrGVGNotExist
	}

	// TODO: if the store size is 0, remove it.
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
		store.Delete(vgtypes.GetLVGKey(bucketID, lvgID))
	}

	store.Delete(vgtypes.GetGVGsBindingOnBucketKey(bucketID))
	return nil
}

func (k Keeper) BindingEmptyObjectToGVG(ctx sdk.Context, bucketID math.Uint, primarySPID, familyID uint32) (*vgtypes.LocalVirtualGroup, error) {
	family, found := k.GetGVGFamily(ctx, primarySPID, familyID)
	if !found {
		return nil, vgtypes.ErrGVGFamilyNotExist
	}

	if len(family.GlobalVirtualGroupIds) == 0 {
		return nil, vgtypes.ErrGVGNotExist.Wrapf("The gvg family has no gvg")
	}

	// use the first gvg by default.
	gvgID := family.GlobalVirtualGroupIds[0]

	return k.BindingObjectToGVG(ctx, bucketID, primarySPID, familyID, gvgID, 0)
}

func (k Keeper) RebindingGVGsToBucket(ctx sdk.Context, bucketID math.Uint, dstSP *sptypes.StorageProvider, gvgMappings []*storagetypes.GVGMapping) error {
	gvgsBindingOnBucket, found := k.GetGVGsBindingOnBucket(ctx, bucketID)
	if !found {
		// empty bucket. do nothing
		return nil
	}
	var newGVGBindingOnBucket vgtypes.GlobalVirtualGroupsBindingOnBucket
	gvg2lvg := make(map[uint32]uint32)
	for i, gvgID := range gvgsBindingOnBucket.GlobalVirtualGroupIds {
		gvg2lvg[gvgID] = gvgsBindingOnBucket.LocalVirtualGroupIds[i]
	}

	// verify secondary signature
	var srcGVGs []*vgtypes.GlobalVirtualGroup
	var dstGVGs []*vgtypes.GlobalVirtualGroup
	for _, gvgMapping := range gvgMappings {
		dstGVG, found := k.GetGVG(ctx, gvgMapping.DstGlobalVirtualGroupId)
		if !found {
			return vgtypes.ErrGVGNotExist.Wrapf("dst gvg not found")
		}

		srcLVGID, exists := gvg2lvg[gvgMapping.SrcGlobalVirtualGroupId]
		if !exists {
			return vgtypes.ErrRebindingGVGsToBucketFailed.Wrapf("src global virtual group not found in gvg mappings, id: %d", gvgMapping.SrcGlobalVirtualGroupId)
		}

		lvg, found := k.GetLVG(ctx, bucketID, srcLVGID)
		if !found {
			return vgtypes.ErrGVGNotExist.Wrapf("lvg not found")
		}

		srcGVG, found := k.GetGVG(ctx, gvgMapping.SrcGlobalVirtualGroupId)
		if !found {
			return vgtypes.ErrGVGNotExist.Wrapf("src gvg not found, ID: %d", gvgMapping.SrcGlobalVirtualGroupId)
		}

		dstGVG.StoredSize += lvg.StoredSize
		srcGVG.StoredSize -= lvg.StoredSize
		// TODO(fynn): add deposit check
		newGVGBindingOnBucket.LocalVirtualGroupIds = append(newGVGBindingOnBucket.LocalVirtualGroupIds, lvg.Id)
		newGVGBindingOnBucket.GlobalVirtualGroupIds = append(newGVGBindingOnBucket.GlobalVirtualGroupIds, gvgMapping.DstGlobalVirtualGroupId)
		delete(gvg2lvg, gvgMapping.SrcGlobalVirtualGroupId)
		srcGVGs = append(srcGVGs, dstGVG)
		dstGVGs = append(dstGVGs, srcGVG)
	}

	if len(gvg2lvg) != 0 || len(gvgsBindingOnBucket.LocalVirtualGroupIds) != len(newGVGBindingOnBucket.LocalVirtualGroupIds) {
		return vgtypes.ErrRebindingGVGsToBucketFailed.Wrapf("LVG is not fully rebinding. please check new lvg to gvg mappings")
	}

	newGVGBindingOnBucket.BucketId = bucketID
	k.SetGVGsBindingOnBucket(ctx, &newGVGBindingOnBucket)
	for _, gvg := range srcGVGs {
		k.SetGVG(ctx, gvg)
	}
	for _, gvg := range dstGVGs {
		k.SetGVG(ctx, gvg)
	}
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
			return vgtypes.ErrInvalidBlsPubKey.Wrapf("BLS public key converts failed: %v", err)
		}
		secondarySpBlsPubKeys = append(secondarySpBlsPubKeys, spBlsPubKey)
	}
	return gnfdtypes.VerifyBlsAggSignature(secondarySpBlsPubKeys, signHash, signature)
}

func (k Keeper) SealObjectOnGVG(ctx sdk.Context, bucketInfo *storagetypes.BucketInfo, gvgID uint32, payloadSize uint64) (*vgtypes.LocalVirtualGroup, error) {
	gvg, err := k.virtualGroupKeeper.GetGlobalVirtualGroupIfAvailable(ctx, gvgID, payloadSize)
	if err != nil {
		return nil, err
	}

	internalBucketInfo := k.MustGetInternalBucketInfo(ctx, bucketInfo.Id)

	if len(internalBucketInfo.LocalVirtualGroups)

	var gvgsBindingOnBucket *vgtypes.GlobalVirtualGroupsBindingOnBucket
	var lvg *vgtypes.LocalVirtualGroup
	gvgsBindingOnBucket, found = k.GetGVGsBindingOnBucket(ctx, bucketID)
	var newLVG = false
	if !found {
		// Create a new key store the gvgs binding on bucket
		lvgID := k.GenNextLVGID(ctx)
		lvg = &vgtypes.LocalVirtualGroup{
			Id:                   lvgID,
			GlobalVirtualGroupId: gvgID,
			StoredSize:           payloadSize,
			BucketId:             bucketID,
		}
		newLVG = true
		gvgsBindingOnBucket = &vgtypes.GlobalVirtualGroupsBindingOnBucket{
			BucketId: bucketID,
		}
		gvgsBindingOnBucket.AppendGVGAndLVG(gvgID, lvgID)
	} else {
		lvgID := gvgsBindingOnBucket.GetLVGIDByGVGID(gvgID)
		if lvgID == 0 {
			if k.MaxLocalVirtualGroupNumPerBucket(ctx) < uint32(len(gvgsBindingOnBucket.LocalVirtualGroupIds)) {
				return nil, vgtypes.ErrLimitationExceed.Wrapf("The lvg number within the bucket exceeds the limit")
			}
			// not exist
			lvgID = k.GenNextLVGID(ctx)
			lvg = &vgtypes.LocalVirtualGroup{
				Id:                   lvgID,
				GlobalVirtualGroupId: gvgID,
				StoredSize:           payloadSize,
				BucketId:             bucketID,
			}
			newLVG = true
			gvgsBindingOnBucket.AppendGVGAndLVG(gvgID, lvgID)
		} else {
			lvg, found = k.GetLVG(ctx, bucketID, lvgID)
			if !found {
				return nil, vgtypes.ErrLVGNotExist
			}
			lvg.StoredSize += payloadSize
		}
	}

	gvg.StoredSize += payloadSize

	k.SetGVG(ctx, gvg)
	k.SetLVG(ctx, lvg)
	k.SetGVGsBindingOnBucket(ctx, gvgsBindingOnBucket)

	if newLVG {
		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventCreateLocalVirtualGroup{
			Id:                   lvg.Id,
			BucketId:             lvg.BucketId,
			GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
			StoredSize:           lvg.StoredSize,
		}); err != nil {
			return nil, err
		}
	} else {
		if err := ctx.EventManager().EmitTypedEvents(&vgtypes.EventUpdateLocalVirtualGroup{
			Id:                   lvg.Id,
			GlobalVirtualGroupId: lvg.GlobalVirtualGroupId,
			StoredSize:           lvg.StoredSize,
		}); err != nil {
			return nil, err
		}
	}
	return lvg, nil
}
