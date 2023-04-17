package keys

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	ctypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	ethHd "github.com/evmos/ethermint/crypto/hd"
)

const (
	defaultBIP39Passphrase = ""
	FullPath               = "m/44'/60'/0'/0/0"
)

type KeyManager interface {
	ctypes.PrivKey
	GetAddr() types.AccAddress
}

type keyManager struct {
	privKey  ctypes.PrivKey
	mnemonic string
	addr     types.AccAddress
}

// TODO: NewKeyStoreKeyManager to be implemented
func NewPrivateKeyManager(priKey string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromPrivateKey(priKey)
	return &k, err
}

func NewMnemonicKeyManager(mnemonic string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryFromMnemonic(mnemonic, FullPath)
	return &k, err
}

func NewBlsMnemonicKeyManager(mnemonic string) (KeyManager, error) {
	k := keyManager{}
	err := k.recoveryBlsFromMnemonic(mnemonic, FullPath)
	return &k, err
}

func (km *keyManager) recoveryFromPrivateKey(privateKey string) error {
	priBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return err
	}

	if len(priBytes) != 32 {
		return fmt.Errorf("Len of Keybytes is not equal to 32 ")
	}
	var keyBytesArray [32]byte
	copy(keyBytesArray[:], priBytes[:32])
	priKey := ethHd.EthSecp256k1.Generate()(keyBytesArray[:]).(*ethsecp256k1.PrivKey)
	km.privKey = priKey
	km.addr = types.AccAddress(km.privKey.PubKey().Address())
	return nil
}

func (km *keyManager) recoveryFromMnemonic(mnemonic, keyPath string) error {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		return fmt.Errorf("mnemonic length should either be 12 or 24")
	}
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, defaultBIP39Passphrase)
	if err != nil {
		return err
	}
	// create master key and derive first key:
	masterPriv, ch := hd.ComputeMastersFromSeed(seed)
	derivedPriv, err := hd.DerivePrivateKeyForPath(masterPriv, ch, keyPath)
	if err != nil {
		return err
	}
	priKey := ethHd.EthSecp256k1.Generate()(derivedPriv[:]).(*ethsecp256k1.PrivKey)
	km.privKey = priKey
	km.mnemonic = mnemonic
	km.addr = types.AccAddress(km.privKey.PubKey().Address())
	return nil
}

func (km *keyManager) recoveryBlsFromMnemonic(mnemonic, keyPath string) error {
	words := strings.Split(mnemonic, " ")
	if len(words) != 12 && len(words) != 24 {
		return fmt.Errorf("mnemonic length should either be 12 or 24")
	}

	derivedPriv, err := hd.EthBLS.Derive()(mnemonic, defaultBIP39Passphrase, keyPath)
	if err != nil {
		return err
	}

	priKey := hd.EthBLS.Generate()(derivedPriv)
	km.privKey = priKey
	km.mnemonic = mnemonic
	km.addr = types.AccAddress(km.privKey.PubKey().Address())
	return nil
}

func (km *keyManager) Bytes() []byte {
	panic("Not allow to get privKey bytes from KeyManager")
}

func (km *keyManager) Sign(msg []byte) ([]byte, error) {
	return km.privKey.Sign(msg)
}

func (km *keyManager) PubKey() ctypes.PubKey {
	return km.privKey.PubKey()
}

func (km *keyManager) Equals(key ctypes.LedgerPrivKey) bool {
	return km.privKey.Equals(key)
}

func (km *keyManager) Type() string {
	return km.privKey.Type()
}

func (km *keyManager) GetAddr() types.AccAddress {
	return km.addr
}

func (km *keyManager) String() string { return km.mnemonic }
func (km *keyManager) ProtoMessage()  {}
func (km *keyManager) Reset()         { *km = keyManager{} }
