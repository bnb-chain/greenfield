package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		spKeeper      types.SpKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	spKeeper types.SpKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		spKeeper:      spKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) QueryDynamicBalance(ctx sdk.Context, addr sdk.AccAddress) (amount sdkmath.Int, err error) {
	streamRecord, found := k.GetStreamRecord(ctx, addr)
	if !found {
		return sdkmath.ZeroInt(), nil
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(addr)
	err = k.UpdateStreamRecord(ctx, streamRecord, change, false)
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrapf(err, "update stream record failed")
	}
	return streamRecord.StaticBalance, nil
}

func (k Keeper) Withdraw(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amount sdkmath.Int) error {
	streamRecord, found := k.GetStreamRecord(ctx, fromAddr)
	if !found {
		return errors.Wrapf(types.ErrStreamRecordNotFound, "validator tax pool stream record not found")
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(fromAddr).WithStaticBalanceChange(amount.Neg())
	err := k.UpdateStreamRecord(ctx, streamRecord, change, false)
	if err != nil {
		return errors.Wrapf(err, "update stream record failed")
	}
	k.SetStreamRecord(ctx, streamRecord)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, amount)))
	if err != nil {
		return errors.Wrapf(err, "send coins from module to account failed")
	}
	return nil
}
