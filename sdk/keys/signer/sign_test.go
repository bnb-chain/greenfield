package keys

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestSigner(t *testing.T) {
	privKey, _, addr := testdata.KeyEthSecp256k1TestPubAddr()
	rawdata := []byte("this is a test stringToSign content")
	// generate signed string bytes
	stringToSign := crypto.Keccak256(rawdata)

	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	require.NoError(t, err)
	signer := NewMsgSigner(keyManager)

	signature, _, err := signer.Sign(stringToSign)
	require.NoError(t, err)
	fmt.Println("origin addr:", addr.String())

	// recover the sender addr
	recoverAcc, pk, err := RecoverAddr(stringToSign, signature)
	require.NoError(t, err)

	fmt.Println("recover sender addr:", recoverAcc.String())
	if !addr.Equals(recoverAcc) {
		t.Errorf("recover addr not same")
	}

	// verify the signature
	verifySucc := secp256k1.VerifySignature(pk.Bytes(), stringToSign, signature[:len(signature)-1])
	if !verifySucc {
		t.Errorf("verify fail")
	}
}
