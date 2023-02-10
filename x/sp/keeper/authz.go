package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

func (k Keeper) CheckDepositAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msg sdk.Msg) error {
	grant, found := k.authzKeeper.GetGrant(ctx, grantee, granter, sdk.MsgTypeURL(msg))
	if !found {
		return authz.ErrNoAuthorizationFound
	}

	if grant.Expiration != nil && grant.Expiration.Before(ctx.BlockTime()) {
		return authz.ErrAuthorizationExpired
	}

	authorization, err := grant.GetAuthorization()
	if err != nil {
		return err
	}

	resp, err := authorization.Accept(ctx, msg)
	if err != nil {
		return err
	}

	if resp.Delete {
		err = k.authzKeeper.DeleteGrant(ctx, grantee, granter, sdk.MsgTypeURL(msg))
	} else if resp.Updated != nil {
		err = k.authzKeeper.Update(ctx, grantee, granter, resp.Updated)
	}
	if err != nil {
		return err
	}

	if !resp.Accept {
		return sdkerrors.ErrUnauthorized
	}

	return nil
}
