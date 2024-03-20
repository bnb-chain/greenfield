package keeper

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var _ sdk.CrossChainApplication = &ExecutorApp{}

type ExecutorApp struct {
	sKeeper    types.StorageKeeper
	sMsgServer types.StorageMsgServer
	pMsgServer types.PaymentMsgServer
}

func NewExecutorApp(storageKeeper types.StorageKeeper, storageMsgServer types.StorageMsgServer, paymentMsgServer types.PaymentMsgServer) *ExecutorApp {
	return &ExecutorApp{
		sKeeper:    storageKeeper,
		sMsgServer: storageMsgServer,
		pMsgServer: paymentMsgServer,
	}
}

func (app *ExecutorApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := deserializeSynPackage(payload)
	if err != nil {
		app.sKeeper.Logger(ctx).Error("deserialize executor syn package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize executor syn package error")
	}

	for i, msgBz := range pack.Msgs {
		msg, err := deserializeExecutorMsg(msgBz)
		if err != nil {
			app.sKeeper.Logger(ctx).Error("deserialize executor msg error", "msg bytes", hex.EncodeToString(msgBz), "error", err.Error())
			panic("deserialize executor msg error")
		}

		err = app.msgHandler(ctx, msg)
		if err != nil {
			app.sKeeper.Logger(ctx).Debug("executor msg error", "index", i, "data", msgBz, "error", err.Error())
		}
	}

	return sdk.ExecuteResult{}
}

func (app *ExecutorApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.sKeeper.Logger(ctx).Error("received execute ack package ")
	return sdk.ExecuteResult{}
}

func (app *ExecutorApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.sKeeper.Logger(ctx).Error("received execute fail ack package ")
	return sdk.ExecuteResult{}
}

type MsgType uint8

const (
	MsgTypeCreatePaymentAccount  MsgType = 1
	MsgTypeDeposit               MsgType = 2
	MsgTypeDisableRefund         MsgType = 3
	MsgWithdraw                  MsgType = 4
	MsgMigrateBucket             MsgType = 5
	MsgCancelMigrateBucket       MsgType = 6
	MsgCompleteMigrateBucket     MsgType = 7
	MsgRejectMigrateBucket       MsgType = 8
	MsgUpdateBucketInfo          MsgType = 9
	MsgToggleSPAsDelegatedAgent  MsgType = 10
	MsgDiscontinueBucket         MsgType = 11
	MsgSetBucketFlowRateLimit    MsgType = 12
	MsgCopyObject                MsgType = 13
	MsgDiscontinueObject         MsgType = 14
	MsgUpdateObjectInfo          MsgType = 15
	MsgLeaveGroup                MsgType = 16
	MsgUpdateGroupExtra          MsgType = 17
	MsgSetTag                    MsgType = 18
	MsgCancelUpdateObjectContent MsgType = 19
)

func (app *ExecutorApp) msgHandler(ctx sdk.Context, msg ExecutorMsgStruct) error {
	msgSender, err := sdk.AccAddressFromHexUnsafe(msg.Sender.String())
	if err != nil {
		return err
	}

	switch MsgType(msg.Type) {
	case MsgTypeCreatePaymentAccount:
		var gnfdMsg paymentmoduletypes.MsgCreatePaymentAccount
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.pMsgServer.CreatePaymentAccount(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgTypeDeposit:
		var gnfdMsg paymentmoduletypes.MsgDeposit
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.pMsgServer.Deposit(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgTypeDisableRefund:
		var gnfdMsg paymentmoduletypes.MsgDisableRefund
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.pMsgServer.DisableRefund(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgWithdraw:
		var gnfdMsg paymentmoduletypes.MsgWithdraw
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.pMsgServer.Withdraw(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgMigrateBucket:
		var gnfdMsg types.MsgMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.MigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCancelMigrateBucket:
		var gnfdMsg types.MsgCancelMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.CancelMigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCompleteMigrateBucket:
		var gnfdMsg types.MsgCompleteMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.CompleteMigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgRejectMigrateBucket:
		var gnfdMsg types.MsgRejectMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.RejectMigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateBucketInfo:
		var gnfdMsg types.MsgUpdateBucketInfo
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.UpdateBucketInfo(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgToggleSPAsDelegatedAgent:
		var gnfdMsg types.MsgToggleSPAsDelegatedAgent
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.ToggleSPAsDelegatedAgent(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgDiscontinueBucket:
		var gnfdMsg types.MsgDiscontinueBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.DiscontinueBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgSetBucketFlowRateLimit:
		var gnfdMsg types.MsgSetBucketFlowRateLimit
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.SetBucketFlowRateLimit(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCopyObject:
		var gnfdMsg types.MsgCopyObject
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.CopyObject(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgDiscontinueObject:
		var gnfdMsg types.MsgDiscontinueObject
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.DiscontinueObject(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateObjectInfo:
		var gnfdMsg types.MsgUpdateObjectInfo
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.UpdateObjectInfo(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgLeaveGroup:
		var gnfdMsg types.MsgLeaveGroup
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.LeaveGroup(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateGroupExtra:
		var gnfdMsg types.MsgUpdateGroupExtra
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.UpdateGroupExtra(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgSetTag:
		var gnfdMsg types.MsgSetTag
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.SetTag(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCancelUpdateObjectContent:
		var gnfdMsg types.MsgCancelUpdateObjectContent
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if !gnfdMsg.GetSigners()[0].Equals(msgSender) {
			return fmt.Errorf("invalid msg sender")
		}
		_, err = app.sMsgServer.CancelUpdateObjectContent(sdk.WrapSDKContext(ctx), &gnfdMsg)
	default:
		err = fmt.Errorf("invalid msg type")
	}
	return err
}

type ExecutorSynPackageStruct struct {
	Msgs [][]byte
}

type ExecutorMsgStruct struct {
	Sender common.Address
	Type   uint8
	Data   []byte
}

var (
	executorSynPackageStructType, _ = abi.NewType("bytes[]", "", nil)

	executorSynPackageArgs = abi.Arguments{
		{Type: executorSynPackageStructType},
	}

	executorMsgStructType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Sender", Type: "address"},
		{Name: "Type", Type: "uint8"},
		{Name: "Data", Type: "bytes"},
	})

	executorMsgArgs = abi.Arguments{
		{Type: executorMsgStructType},
	}
)

func deserializeSynPackage(payload []byte) (ExecutorSynPackageStruct, error) {
	unpacked, err := executorSynPackageArgs.Unpack(payload)
	if err != nil {
		return ExecutorSynPackageStruct{}, err
	}
	unpackedStruct := abi.ConvertType(unpacked[0], ExecutorSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(ExecutorSynPackageStruct)
	if !ok {
		return ExecutorSynPackageStruct{}, err
	}
	return pkgStruct, nil
}

func deserializeExecutorMsg(msgBz []byte) (ExecutorMsgStruct, error) {
	unpacked, err := executorMsgArgs.Unpack(msgBz)
	if err != nil {
		return ExecutorMsgStruct{}, err
	}
	unpackedStruct := abi.ConvertType(unpacked[0], ExecutorMsgStruct{})
	pkgStruct, ok := unpackedStruct.(ExecutorMsgStruct)
	if !ok {
		return ExecutorMsgStruct{}, err
	}
	return pkgStruct, nil
}
