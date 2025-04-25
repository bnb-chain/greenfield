package types

import (
	"bytes"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

func VerifySignature(sigAccAddress sdk.AccAddress, sigHash []byte, sig []byte) error {
	if len(sig) != ethcrypto.SignatureLength {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "signature length (actual: %d) doesn't match typical [R||S||V] signature 65 bytes", len(sig))
	}
	if sig[ethcrypto.RecoveryIDOffset] == 27 || sig[ethcrypto.RecoveryIDOffset] == 28 {
		sig[ethcrypto.RecoveryIDOffset] -= 27
	}
	pubKeyBytes, err := secp256k1.RecoverPubkey(sigHash, sig)
	if err != nil {
		return errors.Wrap(err, "failed to recover delegated fee payer from sig")
	}

	ecPubKey, err := ethcrypto.UnmarshalPubkey(pubKeyBytes)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal recovered fee payer pubkey")
	}

	pubKeyAddr := ethcrypto.PubkeyToAddress(*ecPubKey)
	if !bytes.Equal(pubKeyAddr.Bytes(), sigAccAddress.Bytes()) {
		return errors.Wrapf(sdkerrors.ErrInvalidPubKey, "pubkey %s is different from approval pubkey %s", pubKeyAddr, sigAccAddress)
	}

	recoveredSignerAcc := sdk.AccAddress(pubKeyAddr.Bytes())

	if !recoveredSignerAcc.Equals(sigAccAddress) {
		return errors.Wrapf(sdkerrors.ErrorInvalidSigner, "failed to verify delegated fee payer %s signature", recoveredSignerAcc)
	}

	// VerifySignature of ethsecp256k1 accepts 64 byte signature [R||S]
	// WARNING! Under NO CIRCUMSTANCES try to use pubKey.VerifySignature there
	if !secp256k1.VerifySignature(pubKeyBytes, sigHash, sig[:len(sig)-1]) {
		return errors.Wrap(sdkerrors.ErrorInvalidSigner, "unable to verify signer signature of EIP712 typed data")
	}

	return nil
}

func VerifyBlsSignature(blsPubKey bls.PublicKey, sigHash [32]byte, blsSig []byte) error {
	sig, err := bls.SignatureFromBytes(blsSig)
	if err != nil {
		return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "BLS signature conversion failed: %v", err)
	}
	if !sig.Verify(blsPubKey, sigHash[:]) {
		return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "signature verification failed")
	}
	return nil
}

func VerifyBlsAggSignature(blsPubKeys []bls.PublicKey, sigHash [32]byte, blsAggSig []byte) error {
	aggSig, err := bls.SignatureFromBytes(blsAggSig)
	if err != nil {
		return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "BLS signature conversion failed: %v", err)
	}
	if !aggSig.FastAggregateVerify(blsPubKeys[:], sigHash) {
		return errors.Wrapf(gnfderrors.ErrInvalidBlsSignature, "aggregated signature verification failed")
	}
	return nil
}
