package app

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

const reconStoreKey = "reconciliation"

// unbalancedBlockHeightKey for saving unbalanced block height for reconciliation
var unbalancedBlockHeightKey = []byte("0x01")

var (
	SupplyKey          = banktypes.SupplyKey
	DenomAddressPrefix = banktypes.DenomAddressPrefix
	BalancesPrefix     = banktypes.BalancesPrefix

	StreamRecordKeyPrefix = paymenttypes.StreamRecordKeyPrefix
)

// reconcile will do reconciliation for accounts balances.
func (app *App) reconcile(ctx sdk.Context, bankIavl *iavl.Store, paymentIavl *iavl.Store) {
	if ctx.BlockHeight() <= 2 {
		return
	}

	height, exists := app.getUnbalancedBlockHeight(ctx)
	if exists {
		panic(fmt.Sprintf("unbalanced state at block height %d, please use hardfork to bypass it", height))
	}

	bankBalanced := app.reconBankChanges(ctx, bankIavl)
	bankIavl.ResetDiff()

	paymentBalanced := app.reconPaymentChanges(ctx, paymentIavl)
	paymentIavl.ResetDiff()

	if !bankBalanced || !paymentBalanced {
		ctx.Logger().Error("reconciliation unbalanced", "bank", bankBalanced, "payment", paymentBalanced,
			"height", ctx.BlockHeight())
		app.saveUnbalancedBlockHeight(ctx)
	}
}

// reconBankChanges will reconcile bank balance changes
func (app *App) reconBankChanges(ctx sdk.Context, bankIavl *iavl.Store) bool {
	supplyPre := sdk.Coins{}
	balancePre := sdk.Coins{}
	supplyCurrent := sdk.Coins{}
	balanceCurrent := sdk.Coins{}

	diff := bankIavl.GetDiff()
	version := ctx.BlockHeight() - 2
	ctx.Logger().Debug("reconciliation bank changes", "height", ctx.BlockHeight(), "version", version)
	for k := range diff {
		kBz := []byte(k)
		denom := ""
		isSupply := false
		if bytes.HasPrefix(kBz, SupplyKey) {
			isSupply = true
			denom = parseDenomFromSupplyKey(kBz)
			amount := math.ZeroInt()
			if vBz := bankIavl.Get(kBz); vBz != nil {
				amount = parseAmountFromValue(vBz)
			}
			supplyCurrent = supplyCurrent.Add(sdk.NewCoin(denom, amount))
		} else if bytes.HasPrefix(kBz, BalancesPrefix) {
			denom = parseDenomFromBalanceKey(kBz)
			amount := math.ZeroInt()
			if vBz := bankIavl.Get(kBz); vBz != nil {
				amount = parseAmountFromValue(vBz)
			}
			balanceCurrent = balanceCurrent.Add(sdk.NewCoin(denom, amount))
		} else {
			continue
		}

		preStore, err := bankIavl.GetImmutable(version)
		if err != nil {
			panic(fmt.Sprintf("fail to find store at version %d", version))
		}
		vBz := preStore.Get(kBz)
		if vBz != nil {
			coin := sdk.NewCoin(denom, parseAmountFromValue(vBz))
			if isSupply {
				supplyPre = supplyPre.Add(coin)
			} else {
				balancePre = balancePre.Add(coin)
			}
		}
	}

	supplyChanges, _ := supplyCurrent.SafeSub(supplyPre...)
	balanceChanges, _ := balanceCurrent.SafeSub(balancePre...)

	ctx.Logger().Debug("reconciliation change details", "supplyCurrent", supplyCurrent, "supplyPre", supplyPre,
		"balanceCurrent", balanceCurrent, "balancePre", balancePre,
		"supplyChanges", supplyChanges, "balanceChanges", balanceChanges, "height", ctx.BlockHeight(), "version", version)
	return supplyChanges.IsEqual(balanceChanges)
}

// reconPaymentChanges will reconcile payment flow rate changes
func (app *App) reconPaymentChanges(ctx sdk.Context, paymentIavl *iavl.Store) bool {
	flowCurrent := sdk.ZeroInt()
	flowPre := sdk.ZeroInt()

	diff := paymentIavl.GetDiff()
	version := ctx.BlockHeight() - 2
	ctx.Logger().Debug("reconciliation payment changes", "height", ctx.BlockHeight(), "version", version)
	for k := range diff {
		kBz := []byte(k)
		if bytes.HasPrefix(kBz, StreamRecordKeyPrefix) {
			if vBz := paymentIavl.Get(kBz); vBz != nil {
				var sr paymenttypes.StreamRecord
				err := app.cdc.Unmarshal(vBz, &sr)
				if err != nil {
					ctx.Logger().Error("fail to unmarshal stream record", "err", err.Error())
				} else {
					flowCurrent = flowCurrent.Add(sr.NetflowRate)

					j, _ := json.Marshal(sr)
					ctx.Logger().Debug("stream_record_current", "stream record", j, "addr", parseAddressFromStreamRecordKey(kBz))
				}
			}

			preStore, err := paymentIavl.GetImmutable(version)
			if err != nil {
				panic(fmt.Sprintf("fail to find store at version %d", version))
			}
			vBz := preStore.Get(kBz)
			if vBz != nil {
				var sr paymenttypes.StreamRecord
				err = app.cdc.Unmarshal(vBz, &sr)
				if err != nil {
					ctx.Logger().Error("fail to unmarshal stream record", "err", err.Error())
				} else {
					flowPre = flowPre.Add(sr.NetflowRate)
					j, _ := json.Marshal(sr)
					ctx.Logger().Debug("stream_record_previous", "stream record", j, "addr", parseAddressFromStreamRecordKey(kBz))
				}
			}
		}
	}

	ctx.Logger().Debug("reconciliation payment details", "flowCurrent", flowCurrent.String(), "flowPre", flowPre.String(),
		"height", ctx.BlockHeight(), "version", version)
	return flowCurrent.Equal(flowPre)
}

func (app *App) saveUnbalancedBlockHeight(ctx sdk.Context) {
	reconStore := app.CommitMultiStore().GetCommitStore(sdk.NewKVStoreKey(reconStoreKey)).(*iavl.Store)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz[:], uint64(ctx.BlockHeight()))
	reconStore.Set(unbalancedBlockHeightKey, bz)
}

func (app *App) getUnbalancedBlockHeight(ctx sdk.Context) (uint64, bool) {
	reconStore := app.CommitMultiStore().GetCommitStore(sdk.NewKVStoreKey(reconStoreKey)).(*iavl.Store)
	bz := reconStore.Get(unbalancedBlockHeightKey)
	if bz == nil {
		return 0, false
	}
	return binary.BigEndian.Uint64(bz), true
}

// parseDenomFromBalanceKey parse denom from bank balance key.
// Key format is: BalancesPrefix + Length Prefixed Address + DenomAddressPrefix + Denom + 0x00
func parseDenomFromBalanceKey(key []byte) string {
	l := len(key)
	start := len(BalancesPrefix) + 1 + 20 + len(DenomAddressPrefix) - 1
	return string(key[start:l])
}

// parseAddressFromBalanceKey parse address from bank balance key.
// Key format is: BalancesPrefix + Length Prefixed Address + DenomAddressPrefix + Denom + 0x00
func parseAddressFromBalanceKey(key []byte) string {
	start := len(BalancesPrefix) + 1 // prefix-length
	end := start + 20
	return sdk.AccAddress(key[start:end]).String()
}

// parseDenomFromSupplyKey parse address from bank supply key.
// Key format is: SupplyKey + Denom
func parseDenomFromSupplyKey(key []byte) string {
	start := len(SupplyKey)
	return string(key[start:])
}

func parseAmountFromValue(value []byte) math.Int {
	var amount math.Int
	err := amount.Unmarshal(value)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal amount value %v", err))
	}
	return amount
}

func parseAddressFromStreamRecordKey(key []byte) string {
	start := len(StreamRecordKeyPrefix)
	return sdk.AccAddress(key[start:]).String()
}
