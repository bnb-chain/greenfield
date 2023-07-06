package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/types"
)

func NewPrincipalWithAccount(addr sdk.AccAddress) *Principal {
	return &Principal{
		Type:  PRINCIPAL_TYPE_GNFD_ACCOUNT,
		Value: addr.String(),
	}
}

func NewPrincipalWithGroupId(groupID sdkmath.Uint) *Principal {
	return &Principal{
		Type:  PRINCIPAL_TYPE_GNFD_GROUP,
		Value: groupID.String(),
	}
}

func NewPrincipalWithGroupInfo(groupOwner sdk.AccAddress, groupName string) *Principal {
	return &Principal{
		Type:  PRINCIPAL_TYPE_GNFD_GROUP,
		Value: types.NewGroupGRN(groupOwner, groupName).String(),
	}
}

func (p *Principal) ValidateBasic() error {
	switch p.Type {
	case PRINCIPAL_TYPE_UNSPECIFIED:
		return ErrInvalidPrincipal.Wrapf("Not allowed empty principal type.")
	case PRINCIPAL_TYPE_GNFD_ACCOUNT:
		_, err := sdk.AccAddressFromHexUnsafe(p.Value)
		if err != nil {
			return ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err)
		}
	case PRINCIPAL_TYPE_GNFD_GROUP:
		return nil
	default:
		return ErrInvalidPrincipal.Wrapf("Unknown principal type.")
	}
	return nil
}

func (p *Principal) GetAccountAddress() (sdk.AccAddress, error) {
	if p.Type != PRINCIPAL_TYPE_GNFD_ACCOUNT {
		panic("principal type mismatch.")
	}

	accAddr, err := sdk.AccAddressFromHexUnsafe(p.Value)
	if err != nil {
		return nil, ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err)
	}
	return accAddr, nil
}

func (p *Principal) GetGroupID() (sdkmath.Uint, error) {
	if p.Type != PRINCIPAL_TYPE_GNFD_GROUP {
		panic("principal type mismatch.")
	}

	groupID, err := sdkmath.ParseUint(p.Value)
	if err != nil {
		return sdkmath.ZeroUint(), ErrInvalidPrincipal.Wrapf("Invalid groupID, principal: %s, err: %s", p.String(), err)
	}
	return groupID, nil
}

func (p *Principal) MustGetAccountAddress() sdk.AccAddress {
	address, err := p.GetAccountAddress()
	if err != nil {
		panic(ErrInvalidPrincipal.Wrapf("Invalid account, principal: %s, err: %s", p.String(), err))
	}
	return address
}

func (p *Principal) MustGetGroupID() sdkmath.Uint {
	groupID, err := p.GetGroupID()
	if err != nil {
		panic(ErrInvalidPrincipal.Wrapf("Invalid groupID, principal: %s, err: %s", p.String(), err))
	}
	return groupID
}
