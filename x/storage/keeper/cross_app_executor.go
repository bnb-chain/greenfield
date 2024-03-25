package keeper

import (
	"encoding/hex"
	"fmt"
	"runtime/debug"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
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

func (app *ExecutorApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) (result sdk.ExecuteResult) {
	// This app will not have any ack/failAck package, so we should not panic.
	defer func() {
		if r := recover(); r != nil {
			log := fmt.Sprintf("recovered: %v\nstack:\n%v", r, string(debug.Stack()))
			app.sKeeper.Logger(ctx).Error("execute executor syn package panic", "error", log)
			result.Err = fmt.Errorf("execute executor syn package panic: %v", r)
		}
	}()

	pack, err := DeserializeSynPackage(payload)
	if err != nil {
		app.sKeeper.Logger(ctx).Error("deserialize executor syn package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		return sdk.ExecuteResult{
			Err: fmt.Errorf("deserialize executor syn package error: %v", err),
		}
	}

	for i, msgBz := range pack {
		msg, err := DeserializeExecutorMsg(msgBz)
		if err != nil {
			app.sKeeper.Logger(ctx).Error("deserialize executor msg error", "msg bytes", hex.EncodeToString(msgBz), "error", err.Error())
			return sdk.ExecuteResult{
				Err: fmt.Errorf("deserialize executor msg error: %v", err),
			}
		}

		err = app.msgHandler(ctx, msg)
		if err != nil {
			app.sKeeper.Logger(ctx).Error("execute executor msg error", "index", i, "data", msgBz, "error", err.Error())
			return sdk.ExecuteResult{
				Err: fmt.Errorf("execute executor msg error: %v", err),
			}
		}
	}

	return result
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
	MsgTypeCreatePaymentAccount MsgType = 1
	MsgTypeDeposit              MsgType = 2
	MsgTypeDisableRefund        MsgType = 3
	MsgWithdraw                 MsgType = 4
	MsgMigrateBucket            MsgType = 5
	MsgCancelMigrateBucket      MsgType = 6
	MsgUpdateBucketInfo         MsgType = 7
	MsgToggleSPAsDelegatedAgent MsgType = 8
	MsgSetBucketFlowRateLimit   MsgType = 9
	MsgCopyObject               MsgType = 10
	MsgUpdateObjectInfo         MsgType = 11
	MsgUpdateGroupExtra         MsgType = 12
	MsgSetTag                   MsgType = 13
)

func (app *ExecutorApp) msgHandler(ctx sdk.Context, msg ExecutorMsg) error {
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
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.pMsgServer.CreatePaymentAccount(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgTypeDeposit:
		var gnfdMsg paymentmoduletypes.MsgDeposit
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.pMsgServer.Deposit(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgTypeDisableRefund:
		var gnfdMsg paymentmoduletypes.MsgDisableRefund
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.pMsgServer.DisableRefund(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgWithdraw:
		var gnfdMsg paymentmoduletypes.MsgWithdraw
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.pMsgServer.Withdraw(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgMigrateBucket:
		var gnfdMsg types.MsgMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.MigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCancelMigrateBucket:
		var gnfdMsg types.MsgCancelMigrateBucket
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.CancelMigrateBucket(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateBucketInfo:
		var gnfdMsg types.MsgUpdateBucketInfo
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.UpdateBucketInfo(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgToggleSPAsDelegatedAgent:
		var gnfdMsg types.MsgToggleSPAsDelegatedAgent
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.ToggleSPAsDelegatedAgent(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgSetBucketFlowRateLimit:
		var gnfdMsg types.MsgSetBucketFlowRateLimit
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.SetBucketFlowRateLimit(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgCopyObject:
		var gnfdMsg types.MsgCopyObject
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.CopyObject(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateObjectInfo:
		var gnfdMsg types.MsgUpdateObjectInfo
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.UpdateObjectInfo(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgUpdateGroupExtra:
		var gnfdMsg types.MsgUpdateGroupExtra
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.UpdateGroupExtra(sdk.WrapSDKContext(ctx), &gnfdMsg)
	case MsgSetTag:
		var gnfdMsg types.MsgSetTag
		err = gnfdMsg.Unmarshal(msg.Data)
		if err != nil {
			return err
		}
		if err = checkSigner(msgSender, &gnfdMsg); err != nil {
			return err
		}
		_, err = app.sMsgServer.SetTag(sdk.WrapSDKContext(ctx), &gnfdMsg)
	default:
		err = fmt.Errorf("invalid msg type")
	}
	return err
}

type ExecutorSynPackage [][]byte

type ExecutorMsg struct {
	Sender common.Address
	Type   uint8
	Data   []byte
}

var (
	executorSynPackageTypeDef = `[{"type": "bytes[]"}]`

	executorMsgTypeDef = `[{"type": "address"}, {"type": "uint8"}, {"type": "bytes"}]`
)

func DeserializeSynPackage(payload []byte) (ExecutorSynPackage, error) {
	unpacked, err := abiDecode(executorSynPackageTypeDef, payload)
	if err != nil {
		return ExecutorSynPackage{}, err
	}

	unpackedStruct := abi.ConvertType(unpacked[0], ExecutorSynPackage{})
	pkgStruct, ok := unpackedStruct.(ExecutorSynPackage)
	if !ok {
		return ExecutorSynPackage{}, err
	}
	return pkgStruct, nil
}

func DeserializeExecutorMsg(msgBz []byte) (ExecutorMsg, error) {
	unpacked, err := abiDecode(executorMsgTypeDef, msgBz)
	if err != nil {
		return ExecutorMsg{}, err
	}

	var executorMsg ExecutorMsg
	executorMsg.Sender = abi.ConvertType(unpacked[0], common.Address{}).(common.Address)
	executorMsg.Type = abi.ConvertType(unpacked[1], uint8(0)).(uint8)
	executorMsg.Data = abi.ConvertType(unpacked[2], []byte{}).([]byte)
	return executorMsg, nil
}

func abiDecode(typeDef string, encodedBz []byte) ([]interface{}, error) {
	outDef := fmt.Sprintf(`[{ "name" : "method", "type": "function", "outputs": %s}]`, typeDef)
	outAbi, err := abi.JSON(strings.NewReader(outDef))
	if err != nil {
		return nil, err
	}
	return outAbi.Unpack("method", encodedBz)
}

func checkSigner(msgSender sdk.AccAddress, msg sdk.Msg) error {
	if len(msg.GetSigners()) != 1 {
		return fmt.Errorf("invalid signers number")
	}
	if !msg.GetSigners()[0].Equals(msgSender) {
		return fmt.Errorf("invalid msg sender")
	}
	return nil
}
