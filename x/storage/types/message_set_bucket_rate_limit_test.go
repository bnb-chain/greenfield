package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgSetBucketFlowRateLimit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSetBucketFlowRateLimit
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgSetBucketFlowRateLimit{
				Operator:   "invalid_address",
				BucketName: testBucketName,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid address",
			msg: MsgSetBucketFlowRateLimit{
				Operator:       sample.RandAccAddressHex(),
				PaymentAddress: "invalid address",
				BucketName:     testBucketName,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid address",
			msg: MsgSetBucketFlowRateLimit{
				Operator:       sample.RandAccAddressHex(),
				PaymentAddress: sample.RandAccAddressHex(),
				BucketOwner:    "invalid address",
				BucketName:     testBucketName,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid bucket name",
			msg: MsgSetBucketFlowRateLimit{
				Operator:       sample.RandAccAddressHex(),
				PaymentAddress: sample.RandAccAddressHex(),
				BucketOwner:    sample.RandAccAddressHex(),
				BucketName:     string(testInvalidBucketNameWithLongLength[:]),
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid flow rate limit",
			msg: MsgSetBucketFlowRateLimit{
				Operator:       sample.RandAccAddressHex(),
				PaymentAddress: sample.RandAccAddressHex(),
				BucketOwner:    sample.RandAccAddressHex(),
				BucketName:     testBucketName,
				FlowRateLimit:  sdkmath.NewInt(-1),
			},
			err: sdkerrors.ErrInvalidRequest,
		}, {
			name: "valid case",
			msg: MsgSetBucketFlowRateLimit{
				Operator:       sample.RandAccAddressHex(),
				PaymentAddress: sample.RandAccAddressHex(),
				BucketOwner:    sample.RandAccAddressHex(),
				BucketName:     testBucketName,
				FlowRateLimit:  sdkmath.NewInt(1),
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
