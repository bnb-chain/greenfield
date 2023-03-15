package cli

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagPublic               = "public"
	FlagPaymentAccount       = "payment-account"
	FlagPrimarySP            = "primary-sp"
	FlagExpectChecksums      = "expect-checksums"
	FlagRedundancyType       = "redundancy-type"
	FlagApproveSignature     = "approve-signature"
	FlagApproveTimeoutHeight = "approve-timeout-height"
)

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
