package keys

import (
	"encoding/hex"
	"testing"

	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/stretchr/testify/assert"
)

func TestCreateKeyManagerFromPrivateKeyHex(t *testing.T) {
	planText := []byte("Test")
	priv := "ab463aca3d2965233da3d1d6108aa521274c5ddc2369ff72970a52a451863fbf"
	keyManager, err := NewPrivateKeyManager(priv)
	assert.NoError(t, err)
	sigs, err := keyManager.Sign(planText)
	assert.NoError(t, err)
	valid := keyManager.PubKey().VerifySignature(planText, sigs)
	assert.True(t, valid)
}

func TestCreateKeyManagerFromMnemonic(t *testing.T) {
	mnemonic := "dragon shy author wave swamp avoid lens hen please series heavy squeeze alley castle crazy action peasant green vague camp mirror amount person legal"
	keyManager, err := NewMnemonicKeyManager(mnemonic)
	assert.NoError(t, err)
	address := keyManager.GetAddr().String()
	assert.Equal(t, "0x535E34B319B3575108Deaf2f4FEeeC73AEbE3eF9", address)
}

func TestCreateBlsKeyManagerFromPrivateKeyHex(t *testing.T) {
	blsPrivKey, _ := bls.RandKey()
	blsPubKey := hex.EncodeToString(blsPrivKey.PublicKey().Marshal())
	km, err := NewBlsPrivateKeyManager(hex.EncodeToString(blsPrivKey.Marshal()))
	assert.NoError(t, err)
	assert.Equal(t, blsPubKey, hex.EncodeToString(km.PubKey().Bytes()))
}
