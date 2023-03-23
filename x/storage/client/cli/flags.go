package cli

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	flag "github.com/spf13/pflag"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

const (
	FlagVisibility           = "visibility"
	FlagPaymentAccount       = "payment-account"
	FlagPrimarySP            = "primary-sp"
	FlagExpectChecksums      = "expect-checksums"
	FlagRedundancyType       = "redundancy-type"
	FlagApproveSignature     = "approve-signature"
	FlagApproveTimeoutHeight = "approve-timeout-height"
)

func GetVisibilityType(str string) (storagetypes.VisibilityType, error) {
	v, ok := storagetypes.VisibilityType_value[str]
	if !ok {
		return storagetypes.VISIBILITY_TYPE_PRIVATE, gnfderrors.ErrInvalidVisibilityType
	}
	visibility := storagetypes.VisibilityType(v)

	return visibility, nil
}

// GetPrimarySPField returns a from account address, account name and keyring type, given either an address or key name.
func GetPrimarySPField(kr keyring.Keyring, primarySP string) (sdk.AccAddress, string, keyring.KeyType, error) {
	if primarySP == "" {
		return nil, "", 0, nil
	}

	addr, err := sdk.AccAddressFromHexUnsafe(primarySP)

	var k *keyring.Record
	if err == nil {
		k, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", 0, err
		}
	} else {
		k, err = kr.Key(primarySP)
		if err != nil {
			return nil, "", 0, err
		}
	}

	addr, err = k.GetAddress()
	if err != nil {
		return nil, "", 0, err
	}

	return addr, k.Name, k.GetType(), nil
}

// GetPaymentAccountField returns a from account address, account name and keyring type, given either an address or key name.
func GetPaymentAccountField(kr keyring.Keyring, paymentAcc string) (sdk.AccAddress, string, keyring.KeyType, error) {
	if paymentAcc == "" {
		return nil, "", 0, nil
	}

	addr, err := sdk.AccAddressFromHexUnsafe(paymentAcc)

	var k *keyring.Record
	if err == nil {
		k, err = kr.KeyByAddress(addr)
		if err != nil {
			return nil, "", 0, err
		}
	} else {
		k, err = kr.Key(paymentAcc)
		if err != nil {
			return nil, "", 0, err
		}
	}

	addr, err = k.GetAddress()
	if err != nil {
		return nil, "", 0, err
	}

	return addr, k.Name, k.GetType(), nil
}

// FlagSetVisibility Returns the flagset for set visibility related operations.
func FlagSetVisibility() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String(FlagVisibility, "VISIBILITY_TYPE_PRIVATE", "If private(by default), only owner and grantee can access it. Otherwise,"+
		"every one has permission to access it. Select visibility's type (VISIBILITY_TYPE_PRIVATE|VISIBILITY_TYPE_PUBLIC_READ|VISIBILITY_TYPE_DEFAULT)")
	return fs
}
