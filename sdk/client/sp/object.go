package sp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield/sdk/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// PutObjectMeta  represents meta which is used to construct PutObjectMsg
type PutObjectMeta struct {
	PaymentAccount sdk.AccAddress
	PrimarySp      string
	IsPublic       bool
	ObjectSize     int64
	ContentType    string
}

// ObjectMeta represents meta which may needed when upload payload
type ObjectMeta struct {
	ObjectSize  int64
	ContentType string
}

// UploadResult contains information about the object which has been uploaded
type UploadResult struct {
	BucketName string
	ObjectName string
	ETag       string // Hex encoded unique entity tag of the object.
}

func (t *UploadResult) String() string {
	return fmt.Sprintf("upload finish, bucket name  %s, objectname %s, etag %s", t.BucketName, t.ObjectName, t.ETag)
}

// CreateObject get approval of creating object and send txn to greenfield chain
func (c *SPClient) CreateObject(ctx context.Context, bucketName, objectName string,
	meta PutObjectMeta, reader io.Reader, authInfo AuthInfo) (string, error) {
	// get approval of creating bucket from sp
	signature, err := c.GetApproval(ctx, bucketName, objectName, authInfo)
	if err != nil {
		return "", err
	}
	log.Println("get approve from sp finish,signature is: ", signature)

	// get hash and objectSize from reader
	_, _, _, err = c.GetPieceHashRoots(reader, SegmentSize, EncodeShards)
	if err != nil {
		return "", err
	}

	// TODO(leo) call chain sdk to send a createObject txn to greenfield, return txnHash
	// return txnHash, err

	return "", err
}

// PutObject supports the second stage of uploading the object to bucket.
func (c *SPClient) PutObject(ctx context.Context, bucketName, objectName, txnHash string,
	reader io.Reader, meta ObjectMeta, authInfo AuthInfo) (res UploadResult, err error) {
	if txnHash == "" {
		return UploadResult{}, errors.New("txn hash empty")
	}

	if meta.ObjectSize <= 0 {
		return UploadResult{}, errors.New("object size not set")
	}

	if meta.ContentType == "" {
		return UploadResult{}, errors.New("content type not set")
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: EmptyStringSHA256,
		contentLength: meta.ObjectSize,
		contentType:   meta.ContentType,
	}

	sendOpt := sendOptions{
		method:  http.MethodPut,
		body:    reader,
		txnHash: txnHash,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Printf("upload payload the object failed: %s \n", err.Error())
		return UploadResult{}, err
	}

	etagValue := resp.Header.Get(HTTPHeaderEtag)

	return UploadResult{
		BucketName: bucketName,
		ObjectName: objectName,
		ETag:       etagValue,
	}, nil
}

// FPutObject support upload object from local file
func (c *SPClient) FPutObject(ctx context.Context, bucketName, objectName,
	filePath, txnHash, contentType string, authInfo AuthInfo) (res UploadResult, err error) {
	fReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return UploadResult{}, err
	}
	defer fReader.Close()

	// Save the file stat.
	stat, err := fReader.Stat()
	if err != nil {
		return UploadResult{}, err
	}

	meta := ObjectMeta{
		ObjectSize: stat.Size(),
	}

	if contentType == "" {
		meta.ContentType = "application/octet-stream"
	} else {
		meta.ContentType = contentType
	}

	return c.PutObject(ctx, bucketName, objectName, txnHash, fReader, meta, authInfo)
}

// ObjectInfo contain the meta of downloaded objects
type ObjectInfo struct {
	ObjectName  string
	Etag        string
	ContentType string
	Size        int64
}

// GetObjectOptions contains the options of getObject
type GetObjectOptions struct {
	ResponseContentType string `url:"response-content-type,omitempty" header:"-"`
	Range               string `url:"-" header:"Range,omitempty"`
}

// GetObject download s3 object payload and return the related object info
func (c *SPClient) GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions, authInfo AuthInfo) (io.ReadCloser, ObjectInfo, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return nil, ObjectInfo{}, err
	}
	if err := utils.IsValidObjectName(objectName); err != nil {
		return nil, ObjectInfo{}, err
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		contentSHA256: EmptyStringSHA256,
	}

	//  use for override certain response header values
	//  https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
	if opts.ResponseContentType != "" {
		urlVal := make(url.Values)
		urlVal["response-content-type"] = []string{opts.ResponseContentType}
		reqMeta.urlValues = urlVal
	}

	if opts.Range != "" {
		reqMeta.Range = opts.Range
	}

	sendOpt := sendOptions{
		method:           http.MethodGet,
		disableCloseBody: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Printf("get Object %s fail: %s \n", objectName, err.Error())
		return nil, ObjectInfo{}, err
	}

	ObjInfo, err := getObjInfo(bucketName, objectName, resp.Header)
	if err != nil {
		log.Printf("get ObjectInfo %s fail: %s \n", objectName, err.Error())
		utils.CloseResponse(resp)
		return nil, ObjectInfo{}, err
	}

	return resp.Body, ObjInfo, nil

}

// FGetObject download s3 object payload adn write the object content into local file specified by filePath
func (c *SPClient) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts GetObjectOptions, authinfo AuthInfo) error {
	// Verify if destination already exists.
	st, err := os.Stat(filePath)
	if err == nil {
		// If the destination exists and is a directory.
		if st.IsDir() {
			return errors.New("fileName is a directory.")
		}
	}

	// If file exist, open it in append mode
	fd, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}

	body, _, err := c.GetObject(ctx, bucketName, objectName, opts, authinfo)
	if err != nil {
		log.Printf("download object:%s fail %s \n", objectName, err.Error())
		return err
	}
	defer body.Close()

	_, err = io.Copy(fd, body)
	fd.Close()
	if err != nil {
		return err
	}

	return nil
}

// getObjInfo generate objectInfo base on the response http header content
func getObjInfo(bucketName string, objectName string, h http.Header) (ObjectInfo, error) {
	var etagVal string
	etag := h.Get("Etag")
	if etag != "" {
		etagVal = strings.TrimSuffix(strings.TrimPrefix(etag, "\""), "\"")
	}

	// Parse content length is exists
	var size int64 = -1
	var err error
	contentLength := h.Get(HTTPHeaderContentLength)
	if contentLength != "" {
		size, err = strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return ObjectInfo{}, ErrResponse{
				Code:       "InternalError",
				Message:    fmt.Sprintf("Content-Length parse error %v", err),
				BucketName: bucketName,
				ObjectName: objectName,
				RequestID:  h.Get("x-gnfd-request-id"),
			}
		}
	}

	// fetch content type
	contentType := strings.TrimSpace(h.Get("Content-Type"))
	if contentType == "" {
		contentType = ContentDefault
	}

	return ObjectInfo{
		ObjectName:  objectName,
		Etag:        etagVal,
		ContentType: contentType,
		Size:        size,
	}, nil

}
