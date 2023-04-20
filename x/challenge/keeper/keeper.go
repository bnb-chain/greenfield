package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		tKey       storetypes.StoreKey
		paramstore paramtypes.Subspace

		bankKeeper    types.BankKeeper
		StorageKeeper types.StorageKeeper
		SpKeeper      types.SpKeeper
		stakingKeeper types.StakingKeeper
		paymentKeeper types.PaymentKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey, memKey, tKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	storageKeeper types.StorageKeeper,
	spKeeper types.SpKeeper,
	stakingKeeper types.StakingKeeper,
	paymentKeeper types.PaymentKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		tKey:          tKey,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		StorageKeeper: storageKeeper,
		SpKeeper:      spKeeper,
		stakingKeeper: stakingKeeper,
		paymentKeeper: paymentKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// isInturnAttestation returns whether the attestation submitted in turn
func (k Keeper) isInturnAttestation(ctx sdk.Context, submitter sdk.AccAddress, validators []stakingtypes.Validator) (bool, error) {
	var validatorIndex int64 = -1
	var vldr stakingtypes.Validator
	for index, validator := range validators {
		if validator.ChallengerAddress == submitter.String() {
			validatorIndex = int64(index)
			vldr = validator
			break
		}
	}

	if validatorIndex < 0 {
		return false, errors.Wrapf(types.ErrNotChallenger, "sender (%s) is not a attestation submitter", submitter.String())
	}

	inturnBlsKey, _, err := k.getInturnSubmitter(ctx, k.AttestationInturnInterval(ctx))
	if err != nil {
		return false, err
	}

	return bytes.Equal(inturnBlsKey, vldr.BlsKey), nil
}

func (k Keeper) getInturnSubmitter(ctx sdk.Context, interval uint64) ([]byte, *types.SubmitInterval, error) {
	historicalInfo, ok := k.stakingKeeper.GetHistoricalInfo(ctx, ctx.BlockHeight())
	if !ok {
		return nil, nil, errors.Wrap(types.ErrInvalidVoteValidatorSet, "fail to get validators")
	}
	validators := historicalInfo.Valset

	validatorsSize := len(validators)

	// totalIntervals is sum of intervals from all challengers
	totalIntervals := interval * uint64(validatorsSize)

	curTimeStamp := uint64(ctx.BlockTime().Unix())

	remainder := curTimeStamp % totalIntervals
	inTurnIndex := remainder / interval

	start := curTimeStamp - (remainder - inTurnIndex*interval)
	end := start + interval

	inturnChallenger := validators[inTurnIndex]

	return inturnChallenger.BlsKey, &types.SubmitInterval{
		Start: start,
		End:   end,
	}, nil
}
