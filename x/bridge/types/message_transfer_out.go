package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgTransferOut = "transfer_out"

var _ sdk.Msg = &MsgTransferOut{}

func NewMsgTransferOut(from string, to string, amount *sdk.Coin) *MsgTransferOut {
	return &MsgTransferOut{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func (msg *MsgTransferOut) Route() string {
	return RouterKey
}

func (msg *MsgTransferOut) Type() string {
	return TypeMsgTransferOut
}

func (msg *MsgTransferOut) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.From)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgTransferOut) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgTransferOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.From)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid from address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.To)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid to address (%s)", err)
	}

	if msg.Amount == nil {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "amount should not be nil")
	}

	if !msg.Amount.IsValid() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "%v%s", msg.Amount.Amount, msg.Amount.Denom)
	}

	if !msg.Amount.IsPositive() {
		return errors.Wrapf(sdkerrors.ErrInvalidCoins, "amount should be positive")
	}

	return nil
}
