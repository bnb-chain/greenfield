package types

import (
	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgWithdraw = "withdraw"

var _ sdk.Msg = &MsgWithdraw{}

func NewMsgWithdraw(creator string, from string, amount sdkmath.Int) *MsgWithdraw {
	return &MsgWithdraw{
		Creator: creator,
		From:    from,
		Amount:  amount,
	}
}

func (msg *MsgWithdraw) Route() string {
	return RouterKey
}

func (msg *MsgWithdraw) Type() string {
	return TypeMsgWithdraw
}

func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgWithdraw) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWithdraw) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.From != "" {
		_, err = sdk.AccAddressFromHexUnsafe(msg.From)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
		}
	}

	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount (%s)", msg.Amount)
	}
	return nil
}
