package permission

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/permission/keeper"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	k.RemoveExpiredPolicies(ctx)

}
