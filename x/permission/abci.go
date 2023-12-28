package permission

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/permission/keeper"
	"fmt"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	if ctx.BlockHeight() == 3684712 {
		fmt.Println("debug")
	}
	k.RemoveExpiredPolicies(ctx)

}
