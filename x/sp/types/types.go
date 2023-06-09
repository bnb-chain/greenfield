package types

import (
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	MaxMonikerLength  = 70
	MaxIdentityLength = 3000
	MaxWebsiteLength  = 140
	MaxDetailsLength  = 280
)

// NewStorageProvider constructs a new StorageProvider
func NewStorageProvider(
	operator sdk.AccAddress, fundingAddress sdk.AccAddress,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress, gcAddress sdk.AccAddress,
	totalDeposit math.Int, endpoint string,
	description Description,
	blsKey string) (StorageProvider, error) {

	blsKeyBytes, err := hex.DecodeString(blsKey)
	if err != nil {
		return StorageProvider{}, err
	}

	return StorageProvider{
		OperatorAddress: operator.String(),
		FundingAddress:  fundingAddress.String(),
		SealAddress:     sealAddress.String(),
		ApprovalAddress: approvalAddress.String(),
		GcAddress:       gcAddress.String(),
		TotalDeposit:    totalDeposit,
		Endpoint:        endpoint,
		Description:     description,
		BlsKey:          blsKeyBytes,
	}, nil
}

func (sp *StorageProvider) GetOperator() sdk.AccAddress {
	addr := sdk.MustAccAddressFromHex(sp.OperatorAddress)
	return addr
}

func (sp *StorageProvider) GetFundingAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr := sdk.MustAccAddressFromHex(sp.FundingAddress)
	return addr
}

func (sp *StorageProvider) GetSealAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr := sdk.MustAccAddressFromHex(sp.SealAddress)
	return addr
}

func (sp *StorageProvider) GetApprovalAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr := sdk.MustAccAddressFromHex(sp.ApprovalAddress)
	return addr
}

func (sp *StorageProvider) GetGcAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr := sdk.MustAccAddressFromHex(sp.GcAddress)
	return addr
}

func (sp *StorageProvider) IsInService() bool {
	return sp.GetStatus() == STATUS_IN_SERVICE
}

func (sp *StorageProvider) GetTotalDeposit() math.Int { return sp.TotalDeposit }

// constant used in flags to indicate that description field should not be updated
const DoNotModifyDesc = "[do-not-modify]"

func NewDescription(moniker, identity, website, details string) Description {
	return Description{
		Moniker:  moniker,
		Identity: identity,
		Website:  website,
		Details:  details,
	}
}

// EnsureLength ensures the length of a validator's description.
func (d *Description) EnsureLength() error {
	if len(d.Moniker) > MaxMonikerLength {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}

	if len(d.Identity) > MaxIdentityLength {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), MaxIdentityLength)
	}

	if len(d.Website) > MaxWebsiteLength {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), MaxWebsiteLength)
	}

	if len(d.Details) > MaxDetailsLength {
		return errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	return nil
}

func (d *Description) UpdateDescription(d2 *Description) (*Description, error) {
	if d2.Moniker == DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}

	if d2.Identity == DoNotModifyDesc {
		d2.Identity = d.Identity
	}

	if d2.Website == DoNotModifyDesc {
		d2.Website = d.Website
	}

	if d2.Details == DoNotModifyDesc {
		d2.Details = d.Details
	}

	if err := d2.EnsureLength(); err != nil {
		return d2, err
	}

	return d2, nil
}

func (s *SpStoragePrice) GetSpAccAddress() sdk.AccAddress {
	return sdk.MustAccAddressFromHex(s.SpAddress)
}
