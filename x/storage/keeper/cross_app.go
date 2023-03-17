package keeper

import (
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func RegisterCrossApps(keeper Keeper) {
	bucketApp := NewBucketApp(keeper)
	err := keeper.crossChainKeeper.RegisterChannel(types.BucketChannel, types.BucketChannelId, bucketApp)
	if err != nil {
		panic(err)
	}

	objectApp := NewObjectApp(keeper)
	err = keeper.crossChainKeeper.RegisterChannel(types.ObjectChannel, types.ObjectChannelId, objectApp)
	if err != nil {
		panic(err)
	}

	groupApp := NewGroupApp(keeper)
	err = keeper.crossChainKeeper.RegisterChannel(types.GroupChannel, types.GroupChannelId, groupApp)
	if err != nil {
		panic(err)
	}
}
