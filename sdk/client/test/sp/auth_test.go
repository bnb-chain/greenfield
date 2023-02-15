package sp

import (
	"encoding/hex"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	spClient "github.com/bnb-chain/greenfield/sdk/client/sp"
	"github.com/bnb-chain/greenfield/sdk/keys"
	signer "github.com/bnb-chain/greenfield/sdk/keys/signer"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/stretchr/testify/require"
)

func TestRequestSignV1(t *testing.T) {
	// client actions: new request and sign the request
	urlmap := url.Values{}
	urlmap.Add("greenfield", "chain")
	parms := io.NopCloser(strings.NewReader(urlmap.Encode()))
	req, err := http.NewRequest("POST", "gnfd.nodereal.com", parms)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "testBucket.gnfd.nodereal.com"
	req.Header.Set("X-Gnfd-Date", "11:10")

	privKey, _, addr := testdata.KeyEthSecp256k1TestPubAddr()

	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	require.NoError(t, err)

	client, err = spClient.NewSpClientWithKeyManager("gnfd.nodereal.com", &spClient.Option{}, keyManager)
	require.NoError(t, err)
	err = client.SignRequest(req, spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	// server actions
	// (1) get the header, verify header and check data
	authHeader := req.Header.Get(spClient.HTTPHeaderAuthorization)
	if authHeader == "" {
		t.Errorf("authorization header should not be empty")
	}

	if !strings.Contains(authHeader, spClient.AuthV1) {
		t.Errorf("auth type error")
	}

	// get stringTosign
	signStrIndex := strings.Index(authHeader, " SignedMsg=")
	index := len(" SignedMsg=") + signStrIndex

	// get Siganture
	signatureIndex := strings.Index(authHeader, "Signature=")
	signStr := authHeader[index : signatureIndex-2]

	signature := authHeader[len("Signature=")+signatureIndex:]
	sigBytes, err := hex.DecodeString(signature)
	require.NoError(t, err)

	// (2) server get sender addr
	signMsg := spClient.GetMsgToSign(req)
	if hex.EncodeToString(signMsg) != signStr {
		t.Errorf("string to sign not same")
	}

	recoverAddr, pk, err := signer.RecoverAddr(signMsg, sigBytes)

	require.NoError(t, err)

	if !addr.Equals(recoverAddr) {
		t.Errorf("recover addr not same")
	}

	// (3) server verify the signature
	verifySucc := secp256k1.VerifySignature(pk.Bytes(), signMsg, sigBytes[:len(sigBytes)-1])
	if !verifySucc {
		t.Errorf("verify fail")
	}
}
