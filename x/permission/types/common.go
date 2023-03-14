package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewPrincipalWithAccount(addr sdk.AccAddress) *Principal {
	return &Principal{
		Type:  TYPE_GNFD_ACCOUNT,
		Value: addr.String(),
	}
}

func NewPrincipalWithGroup(groupID sdkmath.Uint) *Principal {
	return &Principal{
		Type:  TYPE_GNFD_GROUP,
		Value: groupID.String(),
	}
}

func (p *Principal) ValidateBasic() error {
	switch p.Type {
	case TYPE_GNFD_ACCOUNT:
		_, err := sdk.AccAddressFromHexUnsafe(p.Value)
		if err != nil {
			return ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err)
		}
	case TYPE_GNFD_GROUP:
		groupID, err := sdkmath.ParseUint(p.Value)
		if err != nil {
			return ErrInvalidPrincipal.Wrapf("Invalid groupID, principal: %s, err: %s", p.String(), err)
		}
		if groupID.Equal(sdkmath.ZeroUint()) {
			return ErrInvalidPrincipal.Wrapf("Zero groupID, principal %s", p.String())
		}
	default:
		return ErrInvalidPrincipal.Wrapf("Unknown principal type.")
	}
	return nil
}

func (p *Principal) GetAccountAddress() (sdk.AccAddress, error) {
	if p.Type != TYPE_GNFD_ACCOUNT {
		panic("principal type mismatch.")
	}

	accAddr, err := sdk.AccAddressFromHexUnsafe(p.Value)
	if err != nil {
		return nil, ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err)
	}
	return accAddr, nil
}

func (p *Principal) GetGroupID() (sdkmath.Uint, error) {
	if p.Type != TYPE_GNFD_GROUP {
		panic("principal type mismatch.")
	}

	groupID, err := sdkmath.ParseUint(p.Value)
	if err != nil {
		return sdkmath.ZeroUint(), ErrInvalidPrincipal.Wrapf("Invalid groupID, principal: %s, err: %s", p.String(), err)
	}
	return groupID, nil
}

func (p *Principal) MustGetAccountAddress() sdk.AccAddress {
	if p.Type != TYPE_GNFD_ACCOUNT {
		panic("principal type mismatch.")
	}

	accAddr, err := sdk.AccAddressFromHexUnsafe(p.Value)
	if err != nil {
		panic(ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err))
	}
	return accAddr
}

func (p *Principal) MustGetGroupID() sdkmath.Uint {
	if p.Type != TYPE_GNFD_GROUP {
		panic("principal type mismatch.")
	}

	groupID, err := sdkmath.ParseUint(p.Value)
	if err != nil {
		panic(ErrInvalidPrincipal.Wrapf("Invalid groupID, principal: %s, err: %s", p.String(), err))
	}
	return groupID
}
