package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey

		bankKeeper       types.BankKeeper
		stakingKeeper    types.StakingKeeper
		crossChainKeeper types.CrossChainKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	stakingKeepr types.StakingKeeper,
	crossChainKeeper types.CrossChainKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		bankKeeper:       bankKeeper,
		stakingKeeper:    stakingKeepr,
		crossChainKeeper: crossChainKeeper,
		authority:        authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetRefundTransferInPayload(transferInClaim *types.TransferInSynPackage, refundReason uint32) ([]byte, error) {
	refundPackage := &types.TransferInRefundPackage{
		RefundAddress: transferInClaim.RefundAddress,
		RefundAmount:  transferInClaim.Amount,
		RefundReason:  refundReason,
	}

	encodedBytes, err := refundPackage.Serialize()
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidPackage, "encode refund package error")
	}
	return encodedBytes, nil
}
