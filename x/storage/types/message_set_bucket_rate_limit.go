package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/bnb-chain/greenfield/types/s3util"
)

const TypeMsgSetBucketFlowRateLimit = "cancel_migrate_bucket"

var _ sdk.Msg = &MsgSetBucketFlowRateLimit{}

func NewMsgSetBucketFlowRateLimit(operator, bucketOwner, paymentAccount sdk.AccAddress, bucketName string, rateLimit sdkmath.Int) *MsgSetBucketFlowRateLimit {
	return &MsgSetBucketFlowRateLimit{
		Operator:       operator.String(),
		PaymentAddress: paymentAccount.String(),
		BucketName:     bucketName,
		BucketOwner:    bucketOwner.String(),
		FlowRateLimit:  rateLimit,
	}
}

func (msg *MsgSetBucketFlowRateLimit) Route() string {
	return RouterKey
}

func (msg *MsgSetBucketFlowRateLimit) Type() string {
	return TypeMsgSetBucketFlowRateLimit
}

func (msg *MsgSetBucketFlowRateLimit) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSetBucketFlowRateLimit) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSetBucketFlowRateLimit) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid operator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid payment address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.BucketOwner)
	if err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid bucket owner address (%s)", err)
	}

	err = s3util.CheckValidBucketName(msg.BucketName)
	if err != nil {
		return err
	}

	if msg.FlowRateLimit.IsNegative() {
		return sdkerrors.ErrInvalidRequest.Wrap("flow rate limit cannot be negative")
	}

	return nil
}
