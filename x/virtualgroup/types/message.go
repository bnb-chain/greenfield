package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

const (
	TypeMsgCreateGlobalVirtualGroup = "create_global_virtual_group"
	TypeMsgDeleteGlobalVirtualGroup = "delete_global_virtual_group"
	TypeMsgDeposit                  = "deposit"
	TypeMsgWithdraw                 = "withdraw"
	TypeMsgSwapOut                  = "swap_out"
	TypeMsgUpdateParams             = "update_params"
	TypeMsgSettle                   = "settle"
)

var (
	_ sdk.Msg = &MsgCreateGlobalVirtualGroup{}
	_ sdk.Msg = &MsgDeleteGlobalVirtualGroup{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgSwapOut{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgSettle{}
)

func NewMsgCreateGlobalVirtualGroup(primarySpAddress sdk.AccAddress, globalVirtualFamilyId uint32, secondarySpIds []uint32, deposit sdk.Coin) *MsgCreateGlobalVirtualGroup {
	return &MsgCreateGlobalVirtualGroup{
		StorageProvider: primarySpAddress.String(),
		FamilyId:        globalVirtualFamilyId,
		SecondarySpIds:  secondarySpIds,
		Deposit:         deposit,
	}
}

func (msg *MsgCreateGlobalVirtualGroup) Route() string {
	return RouterKey
}

func (msg *MsgCreateGlobalVirtualGroup) Type() string {
	return TypeMsgCreateGlobalVirtualGroup
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgCreateGlobalVirtualGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgCreateGlobalVirtualGroup) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgCreateGlobalVirtualGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
	}

	if !msg.Deposit.IsValid() || !msg.Deposit.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deposit amount")
	}

	return nil
}

func NewMsgDeleteGlobalVirtualGroup(primarySpAddress sdk.AccAddress, globalVirtualGroupID uint32) *MsgDeleteGlobalVirtualGroup {
	return &MsgDeleteGlobalVirtualGroup{
		StorageProvider:      primarySpAddress.String(),
		GlobalVirtualGroupId: globalVirtualGroupID,
	}
}

func (msg *MsgDeleteGlobalVirtualGroup) Route() string {
	return RouterKey
}

func (msg *MsgDeleteGlobalVirtualGroup) Type() string {
	return TypeMsgDeleteGlobalVirtualGroup
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgDeleteGlobalVirtualGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgDeleteGlobalVirtualGroup) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgDeleteGlobalVirtualGroup) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
	}

	return nil
}

func NewMsgDeposit(fundingAddress sdk.AccAddress, globalVirtualGroupID uint32, deposit sdk.Coin) *MsgDeposit {
	return &MsgDeposit{
		StorageProvider:      fundingAddress.String(),
		GlobalVirtualGroupId: globalVirtualGroupID,
		Deposit:              deposit,
	}
}

func (msg *MsgDeposit) Route() string {
	return RouterKey
}

func (msg *MsgDeposit) Type() string {
	return TypeMsgDeposit
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgDeposit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
	}

	if !msg.Deposit.IsValid() || !msg.Deposit.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deposit amount")
	}

	return nil
}

func NewMsgWithdraw(fundingAddress sdk.AccAddress, globalVirtualGroupID uint32, withdraw sdk.Coin) *MsgWithdraw {
	return &MsgWithdraw{
		StorageProvider:      fundingAddress.String(),
		GlobalVirtualGroupId: globalVirtualGroupID,
		Withdraw:             withdraw,
	}
}

func (msg *MsgWithdraw) Route() string {
	return RouterKey
}

func (msg *MsgWithdraw) Type() string {
	return TypeMsgWithdraw
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgWithdraw) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgWithdraw) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgWithdraw) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
	}

	if !msg.Withdraw.IsValid() || !msg.Withdraw.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid or non-positive withdraw amount")
	}
	return nil
}

func NewMsgSwapOut(operatorAddress sdk.AccAddress, globalVirtualGroupFamilyID uint32, globalVirtualGroupIDs []uint32, successorSPID uint32) *MsgSwapOut {
	return &MsgSwapOut{
		StorageProvider:            operatorAddress.String(),
		GlobalVirtualGroupFamilyId: globalVirtualGroupFamilyID,
		GlobalVirtualGroupIds:      globalVirtualGroupIDs,
		SuccessorSpId:              successorSPID,
	}
}

func (msg *MsgSwapOut) Route() string {
	return RouterKey
}

func (msg *MsgSwapOut) Type() string {
	return TypeMsgSwapOut
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgSwapOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg *MsgSwapOut) GetApprovalBytes() []byte {
	fakeMsg := proto.Clone(msg).(*MsgSwapOut)
	fakeMsg.SuccessorSpApproval.Sig = nil
	return fakeMsg.GetSignBytes()
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgSwapOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgSwapOut) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
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

	if msg.SuccessorSpId == 0 {
		return gnfderrors.ErrInvalidMessage.Wrap("The successor sp id is not specified.")
	}

	if msg.SuccessorSpApproval == nil {
		return gnfderrors.ErrInvalidMessage.Wrap("The successor sp approval is not specified.")
	}

	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := msg.Params.Validate(); err != nil {
		return err
	}

	return nil
}

func NewMsgSettle(fundingAddress sdk.AccAddress, globalVirtualGroupFamilyID uint32, globalVirtualGroupIDs []uint32) *MsgSettle {
	return &MsgSettle{
		StorageProvider:            fundingAddress.String(),
		GlobalVirtualGroupFamilyId: globalVirtualGroupFamilyID,
		GlobalVirtualGroupIds:      globalVirtualGroupIDs,
	}
}

func (msg *MsgSettle) Route() string {
	return RouterKey
}

func (msg *MsgSettle) Type() string {
	return TypeMsgSettle
}

// GetSignBytes implements the LegacyMsg interface.
func (msg *MsgSettle) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgSettle) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgSettle) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.StorageProvider)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid storage provider address (%s)", err)
	}

	if msg.GlobalVirtualGroupFamilyId == NoSpecifiedFamilyId {
		if len(msg.GlobalVirtualGroupIds) == 0 || len(msg.GlobalVirtualGroupIds) > 10 {
			return ErrInvalidGVGCount
		}
	}

	return nil
}
