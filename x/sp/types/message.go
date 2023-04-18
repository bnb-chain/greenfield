package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgCreateStorageProvider = "create_storage_provider"
	TypeMsgEditStorageProvider   = "edit_storage_provider"
	TypeMsgDeposit               = "deposit"
	TypeMsgUpdateSpStoragePrice  = "update_sp_storage_price"
)

var (
	_ sdk.Msg = &MsgCreateStorageProvider{}
	_ sdk.Msg = &MsgEditStorageProvider{}
	_ sdk.Msg = &MsgDeposit{}
	_ sdk.Msg = &MsgUpdateSpStoragePrice{}
)

// NewMsgCreateStorageProvider creates a new MsgCreateStorageProvider instance.
// creator is the module account of gov module
// SpAddress is the account address of storage provider
// fundAddress is another accoutn address of storage provider which used to deposit or rewarding
func NewMsgCreateStorageProvider(
	creator sdk.AccAddress, SpAddress sdk.AccAddress, fundingAddress sdk.AccAddress,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress, gcAddress sdk.AccAddress,
	description Description, endpoint string, deposit sdk.Coin, readPrice sdk.Dec, freeReadQuota uint64, storePrice sdk.Dec) (*MsgCreateStorageProvider, error) {
	return &MsgCreateStorageProvider{
		Creator:         creator.String(),
		SpAddress:       SpAddress.String(),
		FundingAddress:  fundingAddress.String(),
		SealAddress:     sealAddress.String(),
		ApprovalAddress: approvalAddress.String(),
		GcAddress:       gcAddress.String(),
		Description:     description,
		Endpoint:        endpoint,
		Deposit:         deposit,
		ReadPrice:       readPrice,
		FreeReadQuota:   freeReadQuota,
		StorePrice:      storePrice,
	}, nil
}

// Route implements the sdk.Msg interface.
func (msg *MsgCreateStorageProvider) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgCreateStorageProvider) Type() string {
	return TypeMsgCreateStorageProvider
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgCreateStorageProvider) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgCreateStorageProvider) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgCreateStorageProvider) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.FundingAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid fund address (%s)", err)
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.SealAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seal address (%s)", err)
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.ApprovalAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid approval address (%s)", err)
	}
	if _, err := sdk.AccAddressFromHexUnsafe(msg.GcAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid gc address (%s)", err)
	}
	if !msg.Deposit.IsValid() || !msg.Deposit.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deposit amount")
	}

	if msg.Description == (Description{}) {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	err := IsValidEndpointURL(msg.Endpoint)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid endpoint (%s)", err)
	}
	if msg.ReadPrice.IsNegative() || msg.StorePrice.IsNegative() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid price")
	}
	return nil
}

// NewMsgEditStorageProvider creates a new MsgEditStorageProvider instance
func NewMsgEditStorageProvider(spAddress sdk.AccAddress, endpoint string, description *Description,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress, gcAddress sdk.AccAddress) *MsgEditStorageProvider {
	return &MsgEditStorageProvider{
		SpAddress:       spAddress.String(),
		Endpoint:        endpoint,
		Description:     description,
		SealAddress:     sealAddress.String(),
		ApprovalAddress: approvalAddress.String(),
		GcAddress:       gcAddress.String(),
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgEditStorageProvider) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgEditStorageProvider) Type() string {
	return TypeMsgEditStorageProvider
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgEditStorageProvider) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgEditStorageProvider) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgEditStorageProvider) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.Description != nil && *msg.Description == (Description{}) {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}

	if len(msg.Endpoint) != 0 {
		err = IsValidEndpointURL(msg.Endpoint)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid endpoint (%s)", err)
		}
	}

	if msg.SealAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.SealAddress); err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid seal address (%s)", err)
		}
	}

	if msg.ApprovalAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.ApprovalAddress); err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid approval address (%s)", err)
		}
	}

	if msg.GcAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.GcAddress); err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid gc address (%s)", err)
		}
	}
	return nil
}

// NewMsgDeposit creates a new MsgDeposit instance
func NewMsgDeposit(fundAddress sdk.AccAddress, spAddress sdk.AccAddress, deposit sdk.Coin) *MsgDeposit {
	return &MsgDeposit{
		Creator:   fundAddress.String(),
		SpAddress: spAddress.String(),
		Deposit:   deposit,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgDeposit) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgDeposit) Type() string {
	return TypeMsgDeposit
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgDeposit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgDeposit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgDeposit) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !msg.Deposit.IsValid() || !msg.Deposit.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid deposit amount")
	}

	return nil
}

func (msg *MsgUpdateSpStoragePrice) Route() string {
	return RouterKey
}

func (msg *MsgUpdateSpStoragePrice) Type() string {
	return TypeMsgUpdateSpStoragePrice
}

func (msg *MsgUpdateSpStoragePrice) GetSigners() []sdk.AccAddress {
	spAddr, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{spAddr}
}

func (msg *MsgUpdateSpStoragePrice) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateSpStoragePrice) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s)", err)
	}
	if msg.ReadPrice.IsNil() || msg.ReadPrice.IsNegative() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid read price (%s)", msg.ReadPrice)
	}
	if msg.StorePrice.IsNil() || msg.StorePrice.IsNegative() {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid store price (%s)", msg.StorePrice)
	}
	return nil
}
