package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey

		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		spKeeper      types.SpKeeper
		authority     string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	spKeeper types.SpKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		bankKeeper:    bankKeeper,
		accountKeeper: accountKeeper,
		spKeeper:      spKeeper,
		authority:     authority,
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
	err = k.UpdateStreamRecord(ctx, streamRecord, change)
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrapf(err, "update stream record failed")
	}
	return streamRecord.StaticBalance, nil
}

func (k Keeper) Withdraw(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amount sdkmath.Int) error {
	forced, _ := ctx.Value(types.ForceUpdateStreamRecordKey).(bool) // force update in end block

	streamRecord, found := k.GetStreamRecord(ctx, fromAddr)
	if !found {
		return errors.Wrapf(types.ErrStreamRecordNotFound, "stream record not found %s", fromAddr.String())
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(fromAddr).WithStaticBalanceChange(amount.Neg())
	err := k.UpdateStreamRecord(ctx, streamRecord, change)
	if err != nil {
		return errors.Wrapf(err, "update stream record failed %s", fromAddr.String())
	}
	k.SetStreamRecord(ctx, streamRecord)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, amount)))
	if !forced && err != nil {
		return errors.Wrapf(err, "send coins from module to account failed %s", toAddr.String())
	}
	return nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}
