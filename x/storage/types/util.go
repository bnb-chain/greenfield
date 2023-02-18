package types

import (
	"bytes"
	"crypto/sha256"
	"strings"
	"unicode/utf8"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

// CheckValidBucketName - checks if we have a valid input bucket name.
// This is a stricter version.
// - http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingBucket.html
func CheckValidBucketName(bucketName string) (err error) {
	if len(bucketName) == 0 || strings.TrimSpace(bucketName) == "" {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name cannot be empty")
	}
	if len(bucketName) < 3 {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name cannot be shorter than 3 characters")
	}
	if len(bucketName) > 63 {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name cannot be longer than 63 characters")
	}
	if ipAddress.MatchString(bucketName) {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name cannot be an ip address")
	}
	if strings.Contains(bucketName, "..") || strings.Contains(bucketName, ".-") || strings.Contains(bucketName, "-.") {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name contains invalid characters")
	}
	if !validBucketName.MatchString(bucketName) {
		return errors.Wrap(ErrInvalidBucketName, "Bucket name contains invalid characters")
	}

	return nil
}

const (
	// Bad path components to be rejected by the path validity handler.
	dotdotComponent = ".."
	dotComponent    = "."

	// SlashSeparator - slash separator.
	SlashSeparator = "/"
)

// / CheckValidObjectName - checks if we have a valid input object name.
//   - http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html
func CheckValidObjectName(objectName string) error {
	// check the length of objectname
	if len(objectName) == 0 || strings.TrimSpace(objectName) == "" {
		return errors.Wrap(ErrInvalidObjectName, "Object name cannot be empty")
	}
	if len(objectName) > 1024 {
		return errors.Wrap(ErrInvalidObjectName, "Object name cannot be longer than 1024 characters")
	}

	// check bad path component
	if hasBadPathComponent(objectName) {
		return errors.Wrap(ErrInvalidObjectName, "Object name with a bad path component are not supported")
	}
	// check UTF-8 strings
	if !utf8.ValidString(objectName) {
		return errors.Wrap(ErrInvalidObjectName, "Object name with non UTF-8 strings are not supported")
	}

	if strings.Contains(objectName, `//`) {
		return errors.Wrap(ErrInvalidObjectName, "Object name with a \"//\" prefix are not supported")
	}

	return nil
}

func CheckValidGroupName(groupName string) error {
	if !utf8.ValidString(groupName) {
		return errors.Wrap(ErrInvalidGroupName, "Group name with non UTF-8 strings are not supported")
	}
	return nil
}

// Check if the incoming path has bad path components,
// such as ".." and "."
func hasBadPathComponent(path string) bool {
	path = strings.TrimSpace(path)
	for _, p := range strings.Split(path, SlashSeparator) {
		switch strings.TrimSpace(p) {
		case dotdotComponent:
			return true
		case dotComponent:
			return true
		}
	}
	return false
}

// CheckValidExpectChecksums checks if the MSG have a valid SHA256 checksum.
func CheckValidExpectChecksums(expectChecksums [][]byte) error {
	// TODO(fynn): hard code here. will be replaced by module params.
	if len(expectChecksums) != 7 {
		return ErrInvalidChcecksum
	}
	for _, checksum := range expectChecksums {
		if len(checksum) != sha256.Size {
			return errors.Wrap(ErrInvalidChcecksum, "Invalid SHA256 checksum size.")
		}
	}
	return nil
}

func CheckValidContentType(contentType string) error {
	// TODO(fynn): check validity of the contentType
	return nil
}

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
