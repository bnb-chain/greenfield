package main

//
//import (
//	"bytes"
//	"context"
//	"encoding/base64"
//	"encoding/json"
//	"io"
//	"net/http"
//	"net/http/httputil"
//	"net/url"
//	"strings"
//
//	types2 "github.com/cosmos/cosmos-sdk/types"
//	"github.com/cosmos/cosmos-sdk/types/tx"
//	"github.com/cosmos/cosmos-sdk/types/tx/signing"
//	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
//	"github.com/cosmos/gogoproto/proto"
//	"github.com/labstack/echo"
//
//	"github.com/bnb-chain/greenfield/sdk/client"
//	"github.com/bnb-chain/greenfield/sdk/keys"
//	"github.com/bnb-chain/greenfield/sdk/types"
//)
//
//var (
//	privateKey = "bc6e02007a6d1ba1a0b1e3c85bbeb3570959683a60b4233a7fa6d99c4911168d"
//)
//
//type JsonRpcRequest struct {
//	Jsonrpc string `json:"jsonrpc"`
//	ID      int    `json:"id"`
//	Method  string `json:"method"`
//	Params  struct {
//		Tx string `json:"tx"`
//	} `json:"params"`
//}
//
//func IsMsgTypeSupported(msg types2.Msg) bool {
//	var supportedMessages = map[string]bool{
//		"greenfield.storage.MsgCreateObject":       true,
//		"greenfield.storage.MsgCreateGroup":        true,
//		"greenfield.storage.MsgDeleteGroup":        true,
//		"greenfield.storage.MsgLeaveGroup":         true,
//		"greenfield.storage.MsgDeleteObject":       true,
//		"greenfield.storage.MsgCancelCreateObject": true,
//		"greenfield.storage.MsgCopyObject":         true,
//		"greenfield.storage.MsgUpdateGroupMember":  true,
//		"greenfield.storage.MsgUpdateGroupExtra":   true,
//		"greenfield.storage.MsgRenewGroupMember":   true,
//	}
//
//	msgName := proto.MessageName(msg)
//	return supportedMessages[msgName]
//}
//
//func SignTransaction(cli *client.GreenfieldClient, txStr string) (string, error) {
//	txBytes, err := base64.StdEncoding.DecodeString(txStr)
//
//	txConfig := authtx.NewTxConfig(cli.GetCodec(), []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
//	originalTx, err := txConfig.TxDecoder()(txBytes)
//	if err != nil {
//		return "", err
//	}
//
//	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
//	txOpt := &types.TxOption{
//		Mode: &mode,
//		Memo: "",
//	}
//	originalTxBuilder, err := txConfig.WrapTxBuilder(originalTx)
//	if err != nil {
//		return "", err
//	}
//
//	// check fee payer address against key manager address
//	km, err := cli.GetKeyManager()
//	if err != nil {
//		return "", err
//	}
//
//	// if fee payer is not set or not equal to the key manager address, return raw tx
//	if originalTxBuilder.GetTx().FeePayer().Empty() || !originalTxBuilder.GetTx().FeePayer().Equals(km.GetAddr()) {
//		return txStr, nil
//	}
//
//	if originalTxBuilder.GetTx().FeePayer().Equals(km.GetAddr()) {
//		if len(originalTxBuilder.GetTx().GetSigners()) == 1 && originalTxBuilder.GetTx().GetSigners()[0].Equals(km.GetAddr()) {
//			return txStr, nil
//		}
//	}
//
//	// only support one message at first
//	if len(originalTxBuilder.GetTx().GetMsgs()) != 1 {
//		return txStr, nil
//	}
//
//	if !IsMsgTypeSupported(originalTxBuilder.GetTx().GetMsgs()[0]) {
//		return txStr, nil
//	}
//
//	// todo: check the msg sender address against the subscribers
//
//	// set tx opt with original tx opt
//	txOpt.NoSimulate = true
//	txOpt.Memo = originalTxBuilder.GetTx().GetMemo()
//	txOpt.GasLimit = originalTxBuilder.GetTx().GetGas()
//	txOpt.FeeAmount = originalTxBuilder.GetTx().GetFee()
//	txOpt.FeePayer = originalTxBuilder.GetTx().FeePayer()
//
//	newSignedTxBytes, err := cli.SignTx(context.Background(), originalTxBuilder.GetTx().GetMsgs(), txOpt)
//	if err != nil {
//		return "", err
//	}
//
//	newSignedTx, err := txConfig.TxDecoder()(newSignedTxBytes)
//	if err != nil {
//		return "", err
//	}
//	newSignedTxBuilder, err := txConfig.WrapTxBuilder(newSignedTx)
//	if err != nil {
//		return "", err
//	}
//
//	signatures, err := originalTxBuilder.GetTx().GetSignaturesV2()
//	if err != nil {
//		return "", err
//	}
//	feePayerSignatures, err := newSignedTxBuilder.GetTx().GetSignaturesV2()
//	if err != nil {
//		return "", err
//	}
//
//	signatures = append(signatures, feePayerSignatures...)
//	err = originalTxBuilder.SetSignatures(signatures...)
//	if err != nil {
//		return "", err
//	}
//
//	newSignedBytes, err := txConfig.TxEncoder()(originalTxBuilder.GetTx())
//	if err != nil {
//		return "", err
//	}
//	return base64.StdEncoding.EncodeToString(newSignedBytes), nil
//}
//
//func initGreenFieldClient(privateKey string) (*client.GreenfieldClient, error) {
//	km, err := keys.NewPrivateKeyManager(privateKey)
//	if err != nil {
//		return nil, err
//	}
//	cli, err := client.NewGreenfieldClient("http://localhost:26750", "greenfield_9000-121")
//	if err != nil {
//		return nil, err
//	}
//	cli.SetKeyManager(km)
//	return cli, nil
//}
//
//func main() {
//	e := echo.New()
//
//	// create the reverse proxy
//	url, err := url.Parse("http://localhost:26750")
//	if err != nil {
//		panic(err)
//	}
//	proxy := httputil.NewSingleHostReverseProxy(url)
//
//	reverseProxyRoutePrefix := "/gnfd"
//	routerGroup := e.Group(reverseProxyRoutePrefix)
//	routerGroup.Use(func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
//		return func(ctx echo.Context) error {
//			req := ctx.Request()
//			res := ctx.Response().Writer
//
//			req.Host = url.Host
//			req.URL.Host = url.Host
//			req.URL.Scheme = url.Scheme
//
//			if req.Method == http.MethodPost {
//				body, err := io.ReadAll(req.Body)
//				if err != nil {
//					return err
//				}
//				req.Body.Close()
//
//				rpcRequest := JsonRpcRequest{}
//				err = json.Unmarshal(body, &rpcRequest)
//				if err != nil {
//					return err
//				}
//
//				if strings.Contains(rpcRequest.Method, "broadcast") {
//					cli, err := initGreenFieldClient(privateKey)
//					if err != nil {
//						return err
//					}
//
//					signBytes, err := SignTransaction(cli, rpcRequest.Params.Tx)
//					if err != nil {
//						return err
//					}
//					rpcRequest.Params.Tx = signBytes
//
//					newRpcRequest, err := json.Marshal(&rpcRequest)
//					if err != nil {
//						return err
//					}
//					body = newRpcRequest
//				}
//
//				req.ContentLength = int64(len(body))
//				req.Body = io.NopCloser(bytes.NewBuffer(body))
//			}
//
//			path := req.URL.Path
//			req.URL.Path = strings.TrimLeft(path, reverseProxyRoutePrefix)
//
//			proxy.ServeHTTP(res, req)
//			return nil
//		}
//	})
//
//	e.Start(":8080")
//}
