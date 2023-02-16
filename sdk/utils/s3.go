package utils

import (
	"encoding/hex"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"regexp"
	"strings"
	"unicode/utf8"
)

// EncodePath encode the strings from UTF-8 byte representations to HTML hex escape sequences
func EncodePath(pathName string) string {
	reservedNames := regexp.MustCompile("^[a-zA-Z0-9-_.~/]+$")
	// no need to encode
	if reservedNames.MatchString(pathName) {
		return pathName
	}
	var encodedPathname strings.Builder
	for _, s := range pathName {
		if 'A' <= s && s <= 'Z' || 'a' <= s && s <= 'z' || '0' <= s && s <= '9' { // §2.3 Unreserved characters (mark)
			encodedPathname.WriteRune(s)
			continue
		}
		switch s {
		case '-', '_', '.', '~', '/':
			encodedPathname.WriteRune(s)
			continue
		default:
			len := utf8.RuneLen(s)
			if len < 0 {
				// if utf8 cannot convert return the same string as is
				return pathName
			}
			u := make([]byte, len)
			utf8.EncodeRune(u, s)
			for _, r := range u {
				hex := hex.EncodeToString([]byte{r})
				encodedPathname.WriteString("%" + strings.ToUpper(hex))
			}
		}
	}
	return encodedPathname.String()
}

// IsValidBucketName judge if the bucketname is invalid
// The rule is based on https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html
func IsValidBucketName(bucketName string) error {
	return types.CheckValidBucketName(bucketName)
}

// IsValidObjectName judge if the objectname is invalid
func IsValidObjectName(objectName string) error {
	return types.CheckValidObjectName(objectName)
}
