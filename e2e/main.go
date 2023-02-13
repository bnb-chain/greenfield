package main

import (
	"bufio"
	"context"
	"fmt"
	client "github.com/bnb-chain/greenfield/sdk/client/rpc"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdkCli "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/crypto"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

func main() {
	log.Println("---- e2e test start ----")
	ctx := sdkCli.Context{}
	ctx = ctx.WithCodec(types.Cdc())
	// Parse validator0 mnemonic and init key manager
	val0Mnemonic := ParseValidatorMnemonic(0)
	log.Printf("validator0 mnemonic: %s\n", val0Mnemonic)
	val0Km, err := keys.NewMnemonicKeyManager(val0Mnemonic)
	if err != nil {
		panic(err)
	}
	val0Addr := val0Km.GetAddr()
	// query test
	c := client.NewGreenfieldClient("localhost:9090", "greenfield_9000-121")
	res, err := c.BankQueryClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{Address: val0Addr.String()})
	log.Printf("res: %+v, err: %v\n", res, err)
	// send tx test
	c.SetKeyManager(val0Km)
	to := GenRandomAddr()
	log.Printf("to address: %s\n", to.String())
	balanceBeforeTransfer, err := c.BankQueryClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{Address: to.String()})
	log.Printf("balance before transfer: %+v, err: %v\n", balanceBeforeTransfer, err)
	AssertNoError(err)
	AssertEqual(len(balanceBeforeTransfer.Balances), 0)
	transfer := banktypes.NewMsgSend(val0Addr, to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 1)))
	amount := sdk.NewInt(1)
	txOpt := &types.TxOption{
		Async:     false,
		GasLimit:  1000000,
		Memo:      "",
		FeeAmount: sdk.Coins{{"bnb", amount}},
	}
	txRes, err := SendTxAndWaitForCommit(&c, transfer, txOpt)
	log.Printf("tx: %v, err: %v", txRes, err)
	AssertNoError(err)
	balanceAfterTransfer, err := c.BankQueryClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{Address: to.String()})
	log.Printf("balance after transfer: %+v, err: %v\n", balanceAfterTransfer, err)
	AssertNoError(err)
	AssertEqual(len(balanceAfterTransfer.Balances), 1)
	Assert(balanceAfterTransfer.Balances[0].Amount.Equal(amount), fmt.Sprintf("amount not equal, expect: %s, actual: %s", amount.String(), balanceAfterTransfer.Balances[0].Amount.String()))
	log.Println("---- e2e test end ----")
}

func SendTxAndWaitForCommit(c *client.GreenfieldClient, msg sdk.Msg, txOpt *types.TxOption) (txRes *tx.GetTxResponse, err error) {
	response, err := c.BroadcastTx([]sdk.Msg{msg}, txOpt)
	if err != nil {
		return nil, err
	}
	log.Printf("response: %s", response)
	retry := 10
	for {
		getTxRequest := &tx.GetTxRequest{
			Hash: response.TxResponse.TxHash,
		}
		txRes, err = c.GetTx(context.Background(), getTxRequest)
		//log.Printf("tx: %s, err: %v", txRes, err)
		if err == nil || !strings.Contains(err.Error(), "tx not found") {
			return
		}
		retry--
		if retry < 0 {
			return txRes, fmt.Errorf("reach max retry")
		}
		time.Sleep(time.Second)
	}
}

func GenRandomAddr() sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%d", rand.Int()))))
}

// ParseValidatorMnemonic read a file and return the last non-empty line
func ParseValidatorMnemonic(i int) string {
	file, err := os.Open(fmt.Sprintf("deployment/localup/.local/validator%d/info", i))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var line string
	for scanner.Scan() {
		if scanner.Text() != "" {
			line = scanner.Text()
		}
	}
	return line
}

func AssertNoError(err error) {
	if err != nil {
		panic(err)
	}
}

func AssertEqual(a, b interface{}) {
	if a != b {
		panic(fmt.Sprintf("%v != %v", a, b))
	}
}

func Assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}
