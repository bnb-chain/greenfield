package keeper_test

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func TestDecodeMsg(t *testing.T) {
	operator := sample.RandAccAddress()
	bucketName := string(sample.RandStr(10))
	msg := types.NewMsgMigrateBucket(operator, bucketName, rand.Uint32())
	msgBz, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	var gnfdMsg types.MsgMigrateBucket
	err = gnfdMsg.Unmarshal(msgBz)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, msg.Operator, gnfdMsg.Operator)
	assert.Equal(t, msg.BucketName, gnfdMsg.BucketName)
	assert.Equal(t, msg.DstPrimarySpId, gnfdMsg.DstPrimarySpId)
}

func TestDecodeSynPackage(t *testing.T) {
	payloadStr := "00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000c0000000000000000000000000bfd66d9636253f11ae43f3428e8df73b5ad6950f00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000060000000000000000000000000000000000000000000000000000000000000002c0a2a3078663339466436653531616164383846364634636536614238383237323739636666466239323236360000000000000000000000000000000000000000"
	payloadBz, _ := hex.DecodeString(payloadStr)
	pack, err := keeper.DeserializeSynPackage(payloadBz)
	if err != nil {
		t.Fatal(err)
	}

	msg, err := keeper.DeserializeExecutorMsg(pack[0])
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(msg)
}
