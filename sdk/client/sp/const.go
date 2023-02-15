package sp

import "runtime"

const (
	HTTPHeaderContentLength   = "Content-Length"
	HTTPHeaderContentMD5      = "Content-MD5"
	HTTPHeaderContentType     = "Content-Type"
	HTTPHeaderTransactionHash = "X-Gnfd-Txn-Hash"
	HTTPHeaderResource        = "X-Gnfd-Resource"
	HTTPHeaderPreSignature    = "X-Gnfd-Pre-Signature"
	HTTPHeaderDate            = "X-Gnfd-Date"
	HTTPHeaderEtag            = "ETag"
	HTTPHeaderRange           = "Range"
	HTTPHeaderUserAgent       = "User-Agent"
	HTTPHeaderContentSHA256   = "X-Gnfd-Content-Sha256"

	// EmptyStringSHA256 is the hex encoded sha256 value of an empty string
	EmptyStringSHA256       = `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
	iso8601DateFormatSecond = "2006-01-02T15:04:05Z"

	AdminURLPrefix  = "/greenfield/admin"
	AdminURLVersion = "/v1"

	CreateObjectAction = "CreateObject"
	CreateBucketAction = "CreateBucket"
	SegmentSize        = 16 * 1024 * 1024
	EncodeShards       = 6

	libName        = "Greenfield-go-sdk"
	Version        = "v0.0.1"
	UserAgent      = "Greenfield (" + runtime.GOOS + "; " + runtime.GOARCH + ") " + libName + "/" + Version
	contentTypeXML = "application/xml"
	ContentDefault = "application/octet-stream"
)
