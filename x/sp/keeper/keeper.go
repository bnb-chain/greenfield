package keeper

import (
	"fmt"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		authzKeeper   types.AuthzKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	azk types.AuthzKeeper,
	authority string,

) *Keeper {

	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		accountKeeper: ak,
		bankKeeper:    bk,
		authzKeeper:   azk,
		authority:     authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
