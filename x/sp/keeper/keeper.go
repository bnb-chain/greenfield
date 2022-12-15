package keeper

import (
	"fmt"

	"github.com/bnb-chain/bfs/x/sp/types"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc         codec.BinaryCodec
		storeKey    storetypes.StoreKey
		memKey      storetypes.StoreKey
		paramstore  paramtypes.Subspace
		authKeeper  types.AccountKeeper
		bankKeeper  types.BankKeeper
		authzKeeper types.AuthzKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	azk types.AuthzKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:         cdc,
		storeKey:    key,
		authKeeper:  ak,
		bankKeeper:  bk,
		authzKeeper: azk,
		memKey:      memKey,
		paramstore:  ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
