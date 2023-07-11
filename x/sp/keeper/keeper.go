package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		accountKeeper types.AccountKeeper
		bankKeeper    types.BankKeeper
		authzKeeper   types.AuthzKeeper

		spSequence sequence.Sequence[uint32]
		authority  string
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

	k := &Keeper{
		cdc:           cdc,
		storeKey:      key,
		accountKeeper: ak,
		bankKeeper:    bk,
		authzKeeper:   azk,
		authority:     authority,
	}

	k.spSequence = sequence.NewSequence[uint32](types.StorageProviderSequenceKey)
	return k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetNextSpID(ctx sdk.Context) uint32 {
	store := ctx.KVStore(k.storeKey)

	seq := k.spSequence.NextVal(store)
	return seq
}
