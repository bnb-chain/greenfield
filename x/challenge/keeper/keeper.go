package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		tKey     storetypes.StoreKey

		bankKeeper         types.BankKeeper
		StorageKeeper      types.StorageKeeper
		SpKeeper           types.SpKeeper
		stakingKeeper      types.StakingKeeper
		paymentKeeper      types.PaymentKeeper
		VirtualGroupKeeper types.VirtualGroupKeeper

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey, tKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	storageKeeper types.StorageKeeper,
	spKeeper types.SpKeeper,
	stakingKeeper types.StakingKeeper,
	paymentKeeper types.PaymentKeeper,
	virtualGroupKeeper types.VirtualGroupKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		cdc:                cdc,
		storeKey:           storeKey,
		tKey:               tKey,
		bankKeeper:         bankKeeper,
		StorageKeeper:      storageKeeper,
		SpKeeper:           spKeeper,
		stakingKeeper:      stakingKeeper,
		paymentKeeper:      paymentKeeper,
		VirtualGroupKeeper: virtualGroupKeeper,
		authority:          authority,
	}
}

func (k Keeper) GetAuthority() string {
	return k.authority
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
		return false, errors.Wrapf(types.ErrNotChallenger, "sender (%s) is not an attestation submitter", submitter.String())
	}

	inturnBlsKey, _, err := k.getInturnSubmitter(ctx, k.GetParams(ctx).AttestationInturnInterval)
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
