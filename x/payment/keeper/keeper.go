package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		bankKeeper    types.BankKeeper
		accountKeeper types.AccountKeeper
		spKeeper      types.SpKeeper
		authority     string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
	spKeeper types.SpKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
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

func (k Keeper) GetAuthority() string {
	return k.authority
}
