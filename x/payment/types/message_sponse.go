package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgSponse = "sponse"

var _ sdk.Msg = &MsgSponse{}

func NewMsgSponse(creator string, to string, rate int64) *MsgSponse {
	return &MsgSponse{
		Creator: creator,
		To:      to,
		Rate:    rate,
	}
}

func (msg *MsgSponse) Route() string {
	return RouterKey
}

func (msg *MsgSponse) Type() string {
	return TypeMsgSponse
}

func (msg *MsgSponse) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSponse) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSponse) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if msg.Rate <= 0 {
		return fmt.Errorf("rate must be positive")
	}
	if msg.Creator == msg.To {
		return fmt.Errorf("can not sponse to yourself")
	}
	return nil
}
