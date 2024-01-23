package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/types/s3util"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const TypeMsgCompleteMigrateBucket = "complete_migrate_bucket"

var _ sdk.Msg = &MsgCompleteMigrateBucket{}

func NewMsgCompleteMigrateBucket(operator sdk.AccAddress, bucketName string, globalVirtualGroupFamilyID uint32, gvgMappings []*GVGMapping) *MsgCompleteMigrateBucket {
	return &MsgCompleteMigrateBucket{
		Operator:                   operator.String(),
		BucketName:                 bucketName,
		GlobalVirtualGroupFamilyId: globalVirtualGroupFamilyID,
		GvgMappings:                gvgMappings,
	}
}

func (msg *MsgCompleteMigrateBucket) Route() string {
	return RouterKey
}

func (msg *MsgCompleteMigrateBucket) Type() string {
	return TypeMsgCompleteMigrateBucket
}

func (msg *MsgCompleteMigrateBucket) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCompleteMigrateBucket) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCompleteMigrateBucket) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if msg.GlobalVirtualGroupFamilyId == types.NoSpecifiedFamilyId {
		return gnfderrors.ErrInvalidMessage.Wrap("the global virtual group family id not specify.")
	}

	mappingMap := make(map[uint32]uint32)
	for _, gvgMapping := range msg.GvgMappings {
		if gvgMapping.SrcGlobalVirtualGroupId == 0 || gvgMapping.DstGlobalVirtualGroupId == 0 {
			return ErrInvalidGlobalVirtualGroup.Wrapf("the src gvg id cannot be 0")
		}
		if gvgMapping.SecondarySpBlsSignature == nil {
			return gnfderrors.ErrInvalidBlsSignature.Wrapf("empty signature in gvgMapping")

		}
		_, exist := mappingMap[gvgMapping.SrcGlobalVirtualGroupId]
		if exist {
			return ErrInvalidGlobalVirtualGroup.Wrapf("src gvg id duplicates")
		}
		mappingMap[gvgMapping.SrcGlobalVirtualGroupId] = gvgMapping.DstGlobalVirtualGroupId
	}

	return nil
}

func (msg *MsgCompleteMigrateBucket) ValidateRuntime(ctx sdk.Context) error {
	var err error
	if ctx.IsUpgraded(upgradetypes.Ural) {
		err = s3util.CheckValidBucketNameByCharacterLength(msg.BucketName)
	} else {
		err = s3util.CheckValidBucketName(msg.BucketName)
	}
	if err != nil {
		return err
	}

	return nil
}
