package keys

import (
	"fmt"

	"github.com/bnb-chain/greenfield/sdk/keys"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

// MsgSigner defines a type that for signing msg in the way that is same with MsgEthereumTx
type MsgSigner struct {
	// privKey cryptotypes.PrivKey
	keyManager keys.KeyManager
}

func NewMsgSigner(key keys.KeyManager) *MsgSigner {
	return &MsgSigner{
		keyManager: key,
	}
}

// Sign signs the message using the underlying private key
func (m MsgSigner) Sign(msg []byte) ([]byte, cryptotypes.PubKey, error) {
	privKey := m.keyManager.GetPrivKey()
	if privKey.Type() != ethsecp256k1.KeyType {
		return nil, nil, fmt.Errorf(
			"invalid private key type, expected %s, got %s", ethsecp256k1.KeyType, privKey.Type(),
		)
	}

	sig, err := privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, privKey.PubKey(), nil
}

// RecoverAddr recover the sender address from msg and signature
func RecoverAddr(msg []byte, sig []byte) (sdk.AccAddress, ethsecp256k1.PubKey, error) {
	pubKeyByte, err := secp256k1.RecoverPubkey(msg, sig)
	if err != nil {
		return nil, ethsecp256k1.PubKey{}, err
	}
	pubKey, _ := ethcrypto.UnmarshalPubkey(pubKeyByte)
	pk := ethsecp256k1.PubKey{
		Key: ethcrypto.CompressPubkey(pubKey),
	}

	recoverAcc := sdk.AccAddress(pk.Address().Bytes())

	return recoverAcc, pk, nil
}
