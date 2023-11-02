package types

import (
	"encoding/hex"

	"cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

const (
	TypeMsgCreateStorageProvider       = "create_storage_provider"
	TypeMsgEditStorageProvider         = "edit_storage_provider"
	TypeMsgDeposit                     = "deposit"
	TypeMsgUpdateSpStoragePrice        = "update_sp_storage_price"
	TypeMsgUpdateParams                = "update_params"
	TypeMsgUpdateStorageProviderStatus = "update_storage_provider_status"
)

var (
	_ sdk.Msg = &MsgCreateStorageProvider{}
	_ sdk.Msg = &MsgDeposit{}

	_ sdk.Msg = &MsgEditStorageProvider{}
	_ sdk.Msg = &MsgUpdateSpStoragePrice{}
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgUpdateStorageProviderStatus{}
)

// NewMsgCreateStorageProvider creates a new MsgCreateStorageProvider instance.
// creator is the module account of gov module
// SpAddress is the account address of storage provider
// fundAddress is another accoutn address of storage provider which used to deposit or rewarding
// blsKey is the public key of bls private key, which is used for sealing object and completing migration signature.
// blsProof is the signature signed via bls private key on bls public key bytes
func NewMsgCreateStorageProvider(
	creator sdk.AccAddress, spAddress sdk.AccAddress, fundingAddress sdk.AccAddress,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress, gcAddress sdk.AccAddress, maintenanceAddress sdk.AccAddress,
	description Description, endpoint string, deposit sdk.Coin, readPrice sdk.Dec, freeReadQuota uint64, storePrice sdk.Dec, blsKey, blsProof string) (*MsgCreateStorageProvider, error) {
	return &MsgCreateStorageProvider{
		Creator:            creator.String(),
		SpAddress:          spAddress.String(),
		FundingAddress:     fundingAddress.String(),
		SealAddress:        sealAddress.String(),
		ApprovalAddress:    approvalAddress.String(),
		GcAddress:          gcAddress.String(),
		MaintenanceAddress: maintenanceAddress.String(),
		Description:        description,
		Endpoint:           endpoint,
		Deposit:            deposit,
		ReadPrice:          readPrice,
		FreeReadQuota:      freeReadQuota,
		StorePrice:         storePrice,
		BlsKey:             blsKey,
		BlsProof:           blsProof,
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
	//MaintenanceAddress is validated in msg server
	if !msg.Deposit.IsValid() || !msg.Deposit.Amount.IsPositive() {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "invalid deposit amount")
	}
	if msg.Description == (Description{}) {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "empty description")
	}
	if err := validateBlsKeyAndProof(msg.BlsKey, msg.BlsProof); err != nil {
		return err
	}
	if err := ValidateEndpointURL(msg.Endpoint); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid endpoint (%s)", err)
	}
	if msg.ReadPrice.IsNil() || msg.ReadPrice.IsNegative() || msg.StorePrice.IsNil() || msg.StorePrice.IsNegative() {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "invalid price")
	}
	return nil
}

// NewMsgEditStorageProvider creates a new MsgEditStorageProvider instance
func NewMsgEditStorageProvider(spAddress sdk.AccAddress, endpoint string, description *Description,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress, gcAddress sdk.AccAddress, maintenanceAddress sdk.AccAddress, blsKey, blsProof string) *MsgEditStorageProvider {
	return &MsgEditStorageProvider{
		SpAddress:          spAddress.String(),
		Endpoint:           endpoint,
		Description:        description,
		SealAddress:        sealAddress.String(),
		ApprovalAddress:    approvalAddress.String(),
		GcAddress:          gcAddress.String(),
		MaintenanceAddress: maintenanceAddress.String(),
		BlsKey:             blsKey,
		BlsProof:           blsProof,
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
		err = ValidateEndpointURL(msg.Endpoint)
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
	if msg.MaintenanceAddress != "" {
		if _, err := sdk.AccAddressFromHexUnsafe(msg.MaintenanceAddress); err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid maintenance address (%s)", err)
		}
	}
	if msg.BlsKey != "" {
		if msg.BlsProof == "" {
			return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "bls proof is not provided")
		}
		if err := validateBlsKeyAndProof(msg.BlsKey, msg.BlsProof); err != nil {
			return err
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

	if _, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp address (%s)", err)
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

// GetSignBytes implements the LegacyMsg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
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

// NewMsgUpdateStorageProviderStatus creates a new MsgUpdateStorageProviderStatus instance
func NewMsgUpdateStorageProviderStatus(spAddress sdk.AccAddress, status Status, duration int64) *MsgUpdateStorageProviderStatus {
	return &MsgUpdateStorageProviderStatus{
		SpAddress: spAddress.String(),
		Status:    status,
		Duration:  duration,
	}
}

// Route implements the sdk.Msg interface.
func (msg *MsgUpdateStorageProviderStatus) Route() string {
	return RouterKey
}

// Type implements the sdk.Msg interface.
func (msg *MsgUpdateStorageProviderStatus) Type() string {
	return TypeMsgUpdateStorageProviderStatus
}

// GetSigners implements the sdk.Msg interface.
func (msg *MsgUpdateStorageProviderStatus) GetSigners() []sdk.AccAddress {
	operator, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{operator}
}

// GetSignBytes returns the message bytes to sign over.
func (msg *MsgUpdateStorageProviderStatus) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (msg *MsgUpdateStorageProviderStatus) ValidateBasic() error {
	if _, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}
	if msg.Status != STATUS_IN_SERVICE && msg.Status != STATUS_IN_MAINTENANCE {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "not allowed to update to status %s", msg.Status)
	}
	if msg.Status == STATUS_IN_MAINTENANCE && msg.Duration <= 0 {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "maintenance duration need to be set for %s", msg.Status)
	}
	return nil
}

func validateBlsKeyAndProof(blsKey, blsProof string) error {
	blsPk, err := hex.DecodeString(blsKey)
	if err != nil || len(blsPk) != sdk.BLSPubKeyLength {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "invalid bls pub key")
	}
	blsPubKey, err := bls.PublicKeyFromBytes(blsPk)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "invalid bls pub key")
	}
	bp, err := hex.DecodeString(blsProof)
	if err != nil || len(bp) != sdk.BLSSignatureLength {
		return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "invalid bls sig")
	}
	sig, err := bls.SignatureFromBytes(bp)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "invalid bls signature")
	}
	if !sig.Verify(blsPubKey, tmhash.Sum(blsPk)) {
		return sdkerrors.ErrorInvalidSigner.Wrapf("check bls proof failed.")
	}
	return nil
}
