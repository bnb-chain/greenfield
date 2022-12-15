package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	DestChainId sdk.ChainID = 1 // TODO:  remove this to config
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace

		DestChainId sdk.ChainID

		bankKeeper       types.BankKeeper
		stakingKeeper    types.StakingKeeper
		crossChainKeeper types.CrossChainKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

	bankKeeper types.BankKeeper,
	stakingKeepr types.StakingKeeper,
	crossChainKeeper types.CrossChainKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,

		DestChainId: DestChainId,

		bankKeeper:       bankKeeper,
		stakingKeeper:    stakingKeepr,
		crossChainKeeper: crossChainKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetRefundTransferInPayload(transferInClaim *types.TransferInSynPackage, refundReason types.RefundReason) ([]byte, error) {
	refundPackage := &types.TransferInRefundPackage{
		ContractAddr:    transferInClaim.ContractAddress,
		RefundAddresses: transferInClaim.RefundAddresses,
		RefundAmounts:   transferInClaim.Amounts,
		RefundReason:    refundReason,
	}

	encodedBytes, err := rlp.EncodeToBytes(refundPackage)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidPackage, fmt.Sprintf("encode refund package error"))
	}
	return encodedBytes, nil
}
