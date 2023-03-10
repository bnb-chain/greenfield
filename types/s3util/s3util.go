package s3util

import (
	"crypto/sha256"
	"regexp"
	"strings"
	"unicode/utf8"

	"cosmossdk.io/errors"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

var (
	validBucketName = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)
	ipAddress       = regexp.MustCompile(`^(\d+\.){3}\d+$`)
)

// CheckValidBucketName - checks if we have a valid input bucket name.
// This is a stricter version.
// - http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingBucket.html
func CheckValidBucketName(bucketName string) (err error) {
	if len(bucketName) == 0 || strings.TrimSpace(bucketName) == "" {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name cannot be empty")
	}
	if len(bucketName) < 3 {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name cannot be shorter than 3 characters")
	}
	if len(bucketName) > 63 {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name cannot be longer than 63 characters")
	}
	if ipAddress.MatchString(bucketName) {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name cannot be an ip address")
	}
	if strings.Contains(bucketName, "..") || strings.Contains(bucketName, ".-") || strings.Contains(bucketName, "-.") {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name contains invalid characters")
	}
	if !validBucketName.MatchString(bucketName) {
		return errors.Wrap(gnfderrors.ErrInvalidBucketName, "Bucket name contains invalid characters")
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

// CheckValidObjectName checks if we have a valid input object name.
//
//	http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingMetadata.html
func CheckValidObjectName(objectName string) error {
	// check the length of objectname
	if len(objectName) == 0 || strings.TrimSpace(objectName) == "" {
		return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name cannot be empty")
	}
	if len(objectName) > 1024 {
		return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name cannot be longer than 1024 characters")
	}

	// check bad path component
	if hasBadPathComponent(objectName) {
		return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name with a bad path component are not supported")
	}
	// check UTF-8 strings
	if !utf8.ValidString(objectName) {
		return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name with non UTF-8 strings are not supported")
	}

	if strings.Contains(objectName, `//`) {
		return errors.Wrap(gnfderrors.ErrInvalidObjectName, "Object name with a \"//\" prefix are not supported")
	}

	return nil
}

func CheckValidGroupName(groupName string) error {
	if len(groupName) == 0 || strings.TrimSpace(groupName) == "" {
		return errors.Wrap(gnfderrors.ErrInvalidGroupName, "Group name cannot be empty")
	}
	if len(groupName) < 3 {
		return errors.Wrap(gnfderrors.ErrInvalidGroupName, "Group name cannot be shorter than 3 characters")
	}
	if len(groupName) > 63 {
		return errors.Wrap(gnfderrors.ErrInvalidGroupName, "Group name cannot be longer than 63 characters")
	}
	if !utf8.ValidString(groupName) {
		return errors.Wrap(gnfderrors.ErrInvalidGroupName, "Group name with non UTF-8 strings are not supported")
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
	if len(expectChecksums) != 7 {
		return gnfderrors.ErrInvalidChecksum
	}
	for _, checksum := range expectChecksums {
		if len(checksum) != sha256.Size {
			return errors.Wrap(gnfderrors.ErrInvalidChecksum, "Invalid SHA256 checksum size.")
		}
	}
	return nil
}

func CheckValidContentType(contentType string) error {
	// TODO(fynn): check validity of the contentType
	return nil
}
