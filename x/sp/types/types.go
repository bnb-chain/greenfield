package types

import (
	"bytes"
	"sort"
	"strings"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"sigs.k8s.io/yaml"
)

const (
	MaxMonikerLength  = 70
	MaxIdentityLength = 3000
	MaxWebsiteLength  = 140
	MaxDetailsLength  = 280
)

var (
	// SecondarySpStorePriceRatio shows the ratio of the store price of the secondary sp to the primary sp, the default value is 80%
	SecondarySpStorePriceRatio = sdk.NewDecFromIntWithPrec(sdk.NewInt(8), 1)
)

// NewStorageProvider constructs a new StorageProvider
func NewStorageProvider(
	operator sdk.AccAddress, fundingAddress sdk.AccAddress,
	sealAddress sdk.AccAddress, approvalAddress sdk.AccAddress,
	totalDeposit math.Int, endpoint string,
	description Description) (StorageProvider, error) {
	return StorageProvider{
		OperatorAddress: operator.String(),
		FundingAddress:  fundingAddress.String(),
		SealAddress:     sealAddress.String(),
		ApprovalAddress: approvalAddress.String(),
		TotalDeposit:    totalDeposit,
		Endpoint:        endpoint,
		Description:     description,
	}, nil
}

func (sp StorageProvider) GetOperator() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromHexUnsafe(sp.OperatorAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (sp StorageProvider) GetFundingAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromHexUnsafe(sp.FundingAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (sp StorageProvider) GetSealAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromHexUnsafe(sp.SealAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (sp StorageProvider) GetApprovalAccAddress() sdk.AccAddress {
	if sp.OperatorAddress == "" {
		return sdk.AccAddress{}
	}
	addr, err := sdk.AccAddressFromHexUnsafe(sp.ApprovalAddress)
	if err != nil {
		panic(err)
	}
	return addr
}

func (sp StorageProvider) IsInService() bool {
	return sp.GetStatus() == STATUS_IN_SERVICE
}

func (sp StorageProvider) GetTotalDeposit() math.Int { return sp.TotalDeposit }

// String implements the Stringer interface for a Validator object.
func (sp StorageProvider) String() string {
	bz, err := codec.ProtoMarshalJSON(&sp, nil)
	if err != nil {
		panic(err)
	}

	out, err := yaml.JSONToYAML(bz)
	if err != nil {
		panic(err)
	}

	return string(out)
}

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
func (d Description) EnsureLength() (Description, error) {
	if len(d.Moniker) > MaxMonikerLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), MaxMonikerLength)
	}

	if len(d.Identity) > MaxIdentityLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), MaxIdentityLength)
	}

	if len(d.Website) > MaxWebsiteLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), MaxWebsiteLength)
	}

	if len(d.Details) > MaxDetailsLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), MaxDetailsLength)
	}

	return d, nil
}

func (d Description) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}

func (d Description) UpdateDescription(d2 Description) (Description, error) {
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

	return NewDescription(
		d2.Moniker,
		d2.Identity,
		d2.Website,
		d2.Details,
	).EnsureLength()
}

// Validators is a collection of Validator
type StorageProviders []StorageProvider

func (v StorageProviders) String() (out string) {
	for _, val := range v {
		out += val.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// ToSDKValidators -  convenience function convert []Validator to []sdk.ValidatorI
func (v StorageProviders) ToSDKValidators() (sps []StorageProvider) {
	for _, val := range v {
		sps = append(sps, val)
	}

	return sps
}

// Sort Validators sorts validator array in ascending operator address order
func (v StorageProviders) Sort() {
	sort.Sort(v)
}

// Implements sort interface
func (v StorageProviders) Len() int {
	return len(v)
}

// Implements sort interface
func (v StorageProviders) Less(i, j int) bool {
	return bytes.Compare(v[i].GetOperator().Bytes(), v[j].GetOperator().Bytes()) == -1
}

// Implements sort interface
func (v StorageProviders) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}
