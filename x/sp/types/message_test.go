package types

import (
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

var (
	coinPos  = sdk.NewInt64Coin(DefaultDepositDenom, 100000)
	coinZero = sdk.NewInt64Coin(DefaultDepositDenom, 0)
)

func TestMsgCreateStorageProvider_ValidateBasic(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	spAddr := sdk.AccAddress(pk1.Address())
	blsPubKey, blsProof := sample.RandBlsPubKeyAndBlsProof()

	tests := []struct {
		name, moniker, identity, website, details                                                       string
		creator, spAddress, fundingAddress, sealAddress, approvalAddress, gcAddress, maintenanceAddress sdk.AccAddress
		blsKey, blsProof                                                                                string
		deposit                                                                                         sdk.Coin
		err                                                                                             error
	}{
		{"basic", "a", "b", "c", "d", spAddr, spAddr, spAddr, spAddr, spAddr, spAddr, spAddr, blsPubKey, blsProof, coinPos, nil},
		{"basic_empty", "a", "b", "c", "d", sdk.AccAddress{}, spAddr, spAddr, spAddr, spAddr, spAddr, spAddr, blsPubKey, blsProof, coinPos, sdkerrors.ErrInvalidAddress},
		{"zero deposit", "a", "b", "c", "d", spAddr, spAddr, spAddr, spAddr, spAddr, spAddr, spAddr, blsPubKey, blsProof, coinZero, sdkerrors.ErrInvalidCoins},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgCreateStorageProvider{
				Creator:            tt.creator.String(),
				Description:        NewDescription(tt.moniker, tt.identity, tt.website, tt.details),
				SpAddress:          tt.spAddress.String(),
				FundingAddress:     tt.fundingAddress.String(),
				SealAddress:        tt.sealAddress.String(),
				ApprovalAddress:    tt.approvalAddress.String(),
				GcAddress:          tt.gcAddress.String(),
				MaintenanceAddress: tt.maintenanceAddress.String(),
				BlsKey:             tt.blsKey,
				BlsProof:           tt.blsProof,
				Endpoint:           "http://127.0.0.1:9033",
				StorePrice:         sdk.ZeroDec(),
				ReadPrice:          sdk.ZeroDec(),
				Deposit:            tt.deposit,
			}
			err := msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgEditStorageProvider_ValidateBasic(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	spAddr := sdk.AccAddress(pk1.Address())
	blsPubKey, blsProof := sample.RandBlsPubKeyAndBlsProof()

	tests := []struct {
		name, moniker, identity, website, details string
		spAddress                                 sdk.AccAddress
		blsKey, blsProof                          string
		err                                       error
	}{
		{"basic", "a1", "b1", "c1", "d1", spAddr, blsPubKey, blsProof, nil},
		{"empty", "", "", "", "", spAddr, blsPubKey, blsProof, sdkerrors.ErrInvalidRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := NewDescription(tt.moniker, tt.identity, tt.website, tt.details)
			msg := MsgEditStorageProvider{
				SpAddress:   tt.spAddress.String(),
				Endpoint:    "http://127.0.0.1:9033",
				Description: &desc,
			}
			err := msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgDeposit_ValidateBasic(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()
	pk2 := ed25519.GenPrivKey().PubKey()
	fundAddr := sdk.AccAddress(pk1.Address())
	spAddr := sdk.AccAddress(pk2.Address())
	tests := []struct {
		name                   string
		fundAddress, spAddress sdk.AccAddress
		deposit                sdk.Coin
		err                    error
	}{
		{"basic", fundAddr, spAddr, coinPos, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgDeposit{Creator: tt.fundAddress.String(), SpAddress: spAddr.String(), Deposit: tt.deposit}
			err := msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
