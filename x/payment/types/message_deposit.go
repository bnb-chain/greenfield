package types

import (
	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgDeposit = "deposit"

var _ sdk.Msg = &MsgDeposit{}

func NewMsgDeposit(creator string, to string, amount sdkmath.Int) *MsgDeposit {
	return &MsgDeposit{
		Creator: creator,
		To:      to,
		Amount:  amount,
	}
}

func (msg *MsgDeposit) Route() string {
	return RouterKey
}

func (msg *MsgDeposit) Type() string {
	return TypeMsgDeposit
}

func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgDeposit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	_, err = sdk.AccAddressFromHexUnsafe(msg.To)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid to address (%s)", err)
	}
	if msg.Amount.IsNil() || !msg.Amount.IsPositive() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "invalid amount (%s)", msg.Amount)
	}
	return nil
}
