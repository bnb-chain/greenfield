package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

const TypeMsgCancelSwapOut = "cancel_swap_out"

var _ sdk.Msg = &MsgCancelSwapOut{}

func NewMsgCancelSwapOut(storageProvider sdk.AccAddress, globalVirtualGroupFamilyID uint32, globalVirtualGroupIDs []uint32) *MsgCancelSwapOut {
	return &MsgCancelSwapOut{
		StorageProvider:            storageProvider.String(),
		GlobalVirtualGroupFamilyId: globalVirtualGroupFamilyID,
		GlobalVirtualGroupIds:      globalVirtualGroupIDs,
	}
}

func (msg *MsgCancelSwapOut) Route() string {
	return RouterKey
}

func (msg *MsgCancelSwapOut) Type() string {
	return TypeMsgCancelSwapOut
}

func (msg *MsgCancelSwapOut) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCancelSwapOut) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCancelSwapOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid creator address (%s)", err)
	}
	if msg.GlobalVirtualGroupFamilyId == NoSpecifiedFamilyId {
		if len(msg.GlobalVirtualGroupIds) == 0 {
			return gnfderrors.ErrInvalidMessage.Wrap("The gvgs are not allowed to be empty when familyID is not specified.")
		}
	} else {
		if len(msg.GlobalVirtualGroupIds) > 0 {
			return gnfderrors.ErrInvalidMessage.Wrap("The gvgs are not allowed to be non-empty when familyID is specified.")
		}
	}
	return nil
}
