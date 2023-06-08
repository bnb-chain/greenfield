package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreateGlobalVirtualGroup{}
	_ sdk.Msg = &MsgDeleteGlobalVirtualGroup{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgWithdraw{}
	_ sdk.Msg = &MsgSwapOut{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgCreateGlobalVirtualGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgCreateGlobalVirtualGroup) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.PrimarySpAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgCreateGlobalVirtualGroup) ValidateBasic() error {
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgDeleteGlobalVirtualGroup) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgDeleteGlobalVirtualGroup) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.PrimarySpAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgDeleteGlobalVirtualGroup) ValidateBasic() error {
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgDeposit) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgDeposit) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.FundingAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgDeposit) ValidateBasic() error {
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgWithdraw) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgWithdraw) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.FundingAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgWithdraw) ValidateBasic() error {
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgSwapOut) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgSwapOut) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.SpAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgSwapOut) ValidateBasic() error {
	return nil
}

// GetSignBytes implements the LegacyMsg interface.
func (m *MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromHexUnsafe(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	if err := m.Params.Validate(); err != nil {
		return err
	}

	return nil
}
