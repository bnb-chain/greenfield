package utils

import (
	"encoding/hex"
	"fmt"
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
	nameLen := len(bucketName)
	if nameLen < 3 || nameLen > 63 {
		return fmt.Errorf("bucket name %s len is between [3-63],now is %d", bucketName, nameLen)
	}

	ipAddress := regexp.MustCompile(`^(\d+\.){3}\d+$`)

	if ipAddress.MatchString(bucketName) {
		return fmt.Errorf("The bucket name %s cannot be formatted as an IP address", bucketName)
	}

	if strings.Contains(bucketName, "..") || strings.Contains(bucketName, ".-") || strings.Contains(bucketName, "-.") {
		return fmt.Errorf("Bucket name %s  contains invalid characters", bucketName)
	}
	for _, v := range bucketName {
		if !(('a' <= v && v <= 'z') || ('0' <= v && v <= '9') || v == '-' || v == '.') {
			return fmt.Errorf("bucket name %s can only include lowercase letters, numbers, - and .", bucketName)
		}
	}
	if bucketName[0] == '-' || bucketName[nameLen-1] == '-' || bucketName[0] == '.' || bucketName[nameLen-1] == '.' {
		return fmt.Errorf("bucket name %s must start and end with a lowercase letter or number", bucketName)
	}
	return nil
}

// IsValidObjectName judge if the objectname is invalid
func IsValidObjectName(objectName string) error {
	if len(objectName) == 0 {
		return fmt.Errorf("object name is empty")
	}
	if len(objectName) > 1024 {
		return fmt.Errorf("object length can not be longer than 1024")
	}

	if !utf8.ValidString(objectName) {
		return fmt.Errorf("object name invalid, with non UTF-8 strings")
	}
	return nil
}
