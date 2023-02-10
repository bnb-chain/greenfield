package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgTransferOut_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgTransferOut
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgTransferOut{
				From: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid to address",
			msg: MsgTransferOut{
				From: sample.AccAddress(),
				To:   "invalid address",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid amount",
			msg: MsgTransferOut{
				From:   sample.AccAddress(),
				To:     "0x0000000000000000000000000000000000001000",
				Amount: nil,
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid amount",
			msg: MsgTransferOut{
				From: sample.AccAddress(),
				To:   "0x0000000000000000000000000000000000001000",
				Amount: &sdk.Coin{
					Denom:  "%%%%%",
					Amount: sdk.NewInt(1),
				},
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid amount",
			msg: MsgTransferOut{
				From: sample.AccAddress(),
				To:   "0x0000000000000000000000000000000000001000",
				Amount: &sdk.Coin{
					Denom:  "coin",
					Amount: sdk.NewInt(-1),
				},
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid amount",
			msg: MsgTransferOut{
				From: sample.AccAddress(),
				To:   "0x0000000000000000000000000000000000001000",
				Amount: &sdk.Coin{
					Denom:  "coin",
					Amount: sdk.NewInt(0),
				},
			},
			err: sdkerrors.ErrInvalidCoins,
		},
		{
			name: "invalid amount",
			msg: MsgTransferOut{
				From: sample.AccAddress(),
				To:   "0x0000000000000000000000000000000000001000",
				Amount: &sdk.Coin{
					Denom:  "coin",
					Amount: sdk.NewInt(1),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
