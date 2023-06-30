package app

import (
	"bytes"
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/iavl"
)

const globalAccountNumber = "globalAccountNumber"

//// unbalancedBlockHeightKey for saving unbalanced block height for reconciliation
//var unbalancedBlockHeightKey = []byte("0x01")

// reconBalance will do reconciliation for accounts balances.
func (app *App) reconBalance(ctx sdk.Context, bankIavl *iavl.Store) {
	//height, exists := app.getUnbalancedBlockHeight(ctx)
	//if exists {
	//	panic(fmt.Sprintf("unbalanced state at block height %d, please use hardfork to bypass it", height))
	//}

	if ctx.BlockHeight() < 3 {
		return
	}
	balanced := app.getAccountChanges(ctx, bankIavl)
	bankIavl.ResetDiff()

	if !balanced {
		panic("not balanced")
	}
}

var (
	SupplyKey          = []byte{0x00}
	DenomAddressPrefix = []byte{0x03}
	BalancesPrefix     = []byte{0x02}
)

func (app *App) getAccountChanges(ctx sdk.Context, bankIavl *iavl.Store) bool {
	supplyPre := sdk.Coins{}
	balancePre := sdk.Coins{}
	supplyCurrent := sdk.Coins{}
	balanceCurrent := sdk.Coins{}

	diff := bankIavl.GetDiff()
	version := ctx.BlockHeight() - 2
	fmt.Printf("reconciliation at: %d, version: %d \n", ctx.BlockHeight(), version)
	for k := range diff {
		kBz := []byte(k)
		denom := ""
		isSupply := false
		if bytes.HasPrefix([]byte(k), SupplyKey) {
			isSupply = true
			denom = parseDenomFromSupplyKey(kBz)
			amount := parseAmountFromValue(bankIavl.Get(kBz))
			supplyCurrent = supplyCurrent.Add(sdk.NewCoin(denom, amount))
		} else if bytes.HasPrefix([]byte(k), BalancesPrefix) {
			denom = parseDenomFromBalanceKey(kBz)
			amount := parseAmountFromValue(bankIavl.Get(kBz))
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

	fmt.Println("supplyCurrent", supplyCurrent)
	fmt.Println("supplyPre", supplyPre)
	fmt.Println("balanceCurrent", balanceCurrent)
	fmt.Println("balancePre", balancePre)
	fmt.Println("supplyChanges", supplyChanges)
	fmt.Println("balanceChanges", balanceChanges)
	return supplyChanges.IsEqual(balanceChanges)
}

//func (app *App) saveUnbalancedBlockHeight(ctx sdk.Context) {
//	reconStore := app.GetCommitMultiStore().GetCommitStore(common.ReconStoreKey).(*store.IavlStore)
//	bz := make([]byte, 8)
//	binary.BigEndian.PutUint64(bz[:], uint64(ctx.BlockHeight()))
//	reconStore.Set(unbalancedBlockHeightKey, bz)
//}
//
//func (app *App) getUnbalancedBlockHeight(ctx sdk.Context) (uint64, bool) {
//	reconStore := app.GetCommitMultiStore().GetCommitStore(common.ReconStoreKey).(*store.IavlStore)
//
//	bz := reconStore.Get(unbalancedBlockHeightKey)
//	if bz == nil {
//		return 0, false
//	}
//	return binary.BigEndian.Uint64(bz), true
//}

func parseDenomFromBalanceKey(key []byte) string {
	l := len(key)
	start := len(BalancesPrefix) + 20 + len(DenomAddressPrefix)
	return string(key[start:l])
}

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
