package keeper

import (
	"fmt"

	errors "cosmossdk.io/errors"

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

func (k Keeper) QueryValidatorRewards(ctx sdk.Context) (amount sdkmath.Int, err error) {
	validatorTaxPoolStreamRecord, found := k.GetStreamRecord(ctx, types.ValidatorTaxPoolAddress)
	if !found {
		return sdkmath.ZeroInt(), nil
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(types.ValidatorTaxPoolAddress)
	err = k.UpdateStreamRecord(ctx, validatorTaxPoolStreamRecord, change, false)
	if err != nil {
		return sdkmath.ZeroInt(), errors.Wrapf(err, "update stream record failed")
	}
	return validatorTaxPoolStreamRecord.StaticBalance, nil
}

func (k Keeper) TransferValidatorRewards(ctx sdk.Context, toAddr sdk.AccAddress, amount sdkmath.Int) error {
	validatorTaxPoolStreamRecord, found := k.GetStreamRecord(ctx, types.ValidatorTaxPoolAddress)
	if !found {
		return errors.Wrapf(types.ErrStreamRecordNotFound, "validator tax pool stream record not found")
	}
	change := types.NewDefaultStreamRecordChangeWithAddr(types.ValidatorTaxPoolAddress).WithStaticBalanceChange(amount.Neg())
	err := k.UpdateStreamRecord(ctx, validatorTaxPoolStreamRecord, change, false)
	if err != nil {
		return errors.Wrapf(err, "update stream record failed")
	}
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, toAddr, sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, amount)))
	if err != nil {
		return errors.Wrapf(err, "send coins from module to account failed")
	}
	return nil
}
