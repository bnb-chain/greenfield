package sp

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	common "github.com/bnb-chain/greenfield-common/go"
	sdktype "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/sdk/keys"
	signer "github.com/bnb-chain/greenfield/sdk/keys/signer"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/sdk/utils"
)

// SPClient is a client manages communication with the inscription API.
type SPClient struct {
	endpoint  *url.URL // Parsed endpoint url provided by the user.
	client    *http.Client
	userAgent string
	host      string

	conf       *SPClientConfig
	sender     sdktype.AccAddress // sender greenfield chain address
	keyManager keys.KeyManager
	signer     *signer.MsgSigner
}

// SPClientConfig is the config info of client
type SPClientConfig struct {
	Secure           bool // use https or not
	Transport        http.RoundTripper
	RetryOpt         RetryOptions
	UploadLimitSpeed uint64
}

type Option struct {
	secure bool
}

type RetryOptions struct {
	Count      int
	Interval   time.Duration
	StatusCode []int
}

// NewSpClient returns a new greenfield client
func NewSpClient(endpoint string, opt *Option) (*SPClient, error) {
	url, err := utils.GetEndpointURL(endpoint, opt.secure)
	if err != nil {
		return nil, toInvalidArgumentResp(err.Error())
	}

	httpClient := &http.Client{}
	c := &SPClient{
		client:    httpClient,
		userAgent: UserAgent,
		endpoint:  url,
		conf: &SPClientConfig{
			RetryOpt: RetryOptions{
				Count:    3,
				Interval: time.Duration(0),
			},
		},
	}

	return c, nil
}

// NewSpClientWithKeyManager returns a new greenfield client with keyManager in it
func NewSpClientWithKeyManager(endpoint string, opt *Option, keyManager keys.KeyManager) (*SPClient, error) {
	spClient, err := NewSpClient(endpoint, opt)
	if err != nil {
		return nil, err
	}

	if keyManager == nil {
		return nil, errors.New("keyManager can not be nil")
	}

	spClient.keyManager = keyManager
	if keyManager.GetPrivKey() == nil {
		return nil, errors.New("private key must be set")
	}

	signer := signer.NewMsgSigner(keyManager)
	spClient.signer = signer

	return spClient, nil
}

// GetKeyManager return the keyManager object
func (c *SPClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.KeyManagerNotInitError
	}
	return c.keyManager, nil
}

// GetMsgSigner return the signer
func (c *SPClient) GetMsgSigner() (*signer.MsgSigner, error) {
	if c.signer == nil {
		return nil, errors.New("signer is nil")
	}
	return c.signer, nil
}

// GetURL returns the URL of the S3 endpoint.
func (c *SPClient) GetURL() *url.URL {
	endpoint := *c.endpoint
	return &endpoint
}

// requestMeta - contain the metadata to construct the http request.
type requestMeta struct {
	bucketName       string
	objectName       string
	urlRelPath       string     // relative path of url
	urlValues        url.Values // url values to be added into url
	Range            string
	ApproveAction    string
	SignType         string
	contentType      string
	contentLength    int64
	contentMD5Base64 string // base64 encoded md5sum
	contentSHA256    string // hex encoded sha256sum
}

// sendOptions -  options to use to send the http message
type sendOptions struct {
	method           string      // request method
	body             interface{} // request body
	result           interface{} // unmarshal message of the resp.Body
	disableCloseBody bool        // indicate whether to disable automatic calls to resp.Body.Close()
	txnHash          string      // the transaction hash info
	isAdminApi       bool        // indicate if it is an admin api request
}

// SetHost set host name of request
func (c *SPClient) SetHost(hostName string) {
	c.host = hostName
}

// GetHost get host name of request
func (c *SPClient) GetHost() string {
	return c.host
}

// GetAccount get sender address info
func (c *SPClient) GetAccount() sdktype.AccAddress {
	return c.sender
}

// GetAgent get agent name
func (c *SPClient) GetAgent() string {
	return c.userAgent
}

// newRequest construct the http request, set url, body and headers
func (c *SPClient) newRequest(ctx context.Context,
	method string, meta requestMeta, body interface{}, txnHash string, isAdminAPi bool, authInfo AuthInfo) (req *http.Request, err error) {
	// construct the target url
	desURL, err := c.GenerateURL(meta.bucketName, meta.objectName, meta.urlRelPath, meta.urlValues, isAdminAPi)
	if err != nil {
		return nil, err
	}

	var reader io.Reader
	contentType := ""
	sha256Hex := ""
	if body != nil {
		// the body content is io.Reader type
		if ObjectReader, ok := body.(io.Reader); ok {
			reader = ObjectReader
			if meta.contentType == "" {
				contentType = ContentDefault
			}
		} else {
			// the body content is xml type
			content, err := xml.Marshal(body)
			if err != nil {
				return nil, err
			}
			contentType = contentTypeXML
			reader = bytes.NewReader(content)
			sha256Hex = utils.CalcSHA256Hex(content)
		}
	}

	// Initialize a new HTTP request for the method.
	req, err = http.NewRequestWithContext(ctx, method, desURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// need to turn the body into ReadCloser
	if body == nil {
		req.Body = nil
	} else {
		req.Body = io.NopCloser(reader)
	}

	// set content length
	req.ContentLength = meta.contentLength

	// set txn hash header
	if txnHash != "" {
		req.Header.Set(HTTPHeaderTransactionHash, txnHash)
	}

	// set content type header
	if meta.contentType != "" {
		req.Header.Set(HTTPHeaderContentType, meta.contentType)
	} else if contentType != "" {
		req.Header.Set(HTTPHeaderContentType, contentType)
	} else {
		req.Header.Set(HTTPHeaderContentType, ContentDefault)
	}

	// set md5 header
	if meta.contentMD5Base64 != "" {
		req.Header[HTTPHeaderContentMD5] = []string{meta.contentMD5Base64}
	}

	// set sha256 header
	if meta.contentSHA256 != "" {
		req.Header[HTTPHeaderContentSHA256] = []string{meta.contentSHA256}
	} else {
		req.Header[HTTPHeaderContentSHA256] = []string{sha256Hex}
	}

	if meta.Range != "" && method == http.MethodGet {
		req.Header.Set(HTTPHeaderRange, meta.Range)
	}

	if isAdminAPi {
		if meta.objectName == "" {
			req.Header.Set(HTTPHeaderResource, meta.bucketName)
		} else {
			req.Header.Set(HTTPHeaderResource, meta.bucketName+"/"+meta.objectName)
		}
	} else {
		// set request host
		if c.host != "" {
			req.Host = c.host
		} else if req.URL.Host != "" {
			req.Host = req.URL.Host
		}
	}

	// set date header
	stNow := time.Now().UTC()
	req.Header.Set(HTTPHeaderDate, stNow.Format(iso8601DateFormatSecond))

	// set user-agent
	req.Header.Set(HTTPHeaderUserAgent, c.userAgent)

	// sign the total http request info when auth type v1
	err = c.SignRequest(req, authInfo)
	if err != nil {
		return req, err
	}

	return
}

// doAPI call client.Do() to send request and read response from servers
func (c *SPClient) doAPI(ctx context.Context, req *http.Request, meta requestMeta, closeBody bool) (*http.Response, error) {
	var cancel context.CancelFunc
	if closeBody {
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		// If we got an error, and the context has been canceled,
		// the context's error is probably more useful.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if urlErr, ok := err.(*url.Error); ok {
			if strings.Contains(urlErr.Err.Error(), "EOF") {
				return nil, &url.Error{
					Op:  urlErr.Op,
					URL: urlErr.URL,
					Err: errors.New("Connection closed by foreign host " + urlErr.URL + ". Retry again."),
				}
			}
		}
		return nil, err
	}
	defer func() {
		if closeBody {
			utils.CloseResponse(resp)
		}
	}()

	// construct err responses and messages
	err = constructErrResponse(resp, meta.bucketName, meta.objectName)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// sendReq new restful request, send the message and handle the response
func (c *SPClient) sendReq(ctx context.Context, metadata requestMeta, opt *sendOptions, authInfo AuthInfo) (res *http.Response, err error) {
	req, err := c.newRequest(ctx, opt.method, metadata, opt.body, opt.txnHash, opt.isAdminApi, authInfo)
	if err != nil {
		log.Printf("new request error: %s , stop send request\n", err.Error())
		return nil, err
	}

	resp, err := c.doAPI(ctx, req, metadata, !opt.disableCloseBody)
	if err != nil {
		log.Printf("do api request fail: %s \n", err.Error())
		return nil, err
	}
	return resp, nil
}

// genURL construct the target request url based on the parameters
func (c *SPClient) GenerateURL(bucketName string, objectName string, relativePath string,
	queryValues url.Values, isAdminApi bool) (*url.URL, error) {
	host := c.endpoint.Host
	scheme := c.endpoint.Scheme

	// Strip port 80 and 443
	if h, p, err := net.SplitHostPort(host); err == nil {
		if scheme == "http" && p == "80" || scheme == "https" && p == "443" {
			host = h
			if ip := net.ParseIP(h); ip != nil && ip.To16() != nil {
				host = "[" + h + "]"
			}
		}
	}

	if bucketName == "" {
		err := errors.New("no bucketName in path")
		return nil, err
	}

	var urlStr string
	if isAdminApi {
		prefix := AdminURLPrefix + AdminURLVersion
		urlStr = scheme + "://" + host + prefix + "/"
	} else {
		// generate s3 virtual hosted style url
		if utils.IsDomainNameValid(host) {
			urlStr = scheme + "://" + bucketName + "." + host + "/"
		} else {
			urlStr = scheme + "://" + host + "/"
		}

		if objectName != "" {
			urlStr += utils.EncodePath(objectName)
		}
	}

	if relativePath != "" {
		urlStr += utils.EncodePath(relativePath)
	}

	if len(queryValues) > 0 {
		urlStrNew, err := utils.AddQueryValues(urlStr, queryValues)
		if err != nil {
			return nil, err
		}
		urlStr = urlStrNew
	}

	return url.Parse(urlStr)
}

// SignRequest sign the request and set authorization before send to server
func (c *SPClient) SignRequest(req *http.Request, info AuthInfo) error {
	var authStr []string
	if info.SignType == AuthV1 {
		signMsg := GetMsgToSign(req)

		if c.signer == nil {
			return errors.New("signer can not be nil with auth v1 type")
		}

		// sign the request header info, generate the signature
		signature, _, err := c.signer.Sign(signMsg)
		if err != nil {
			return err
		}

		authStr = []string{
			AuthV1 + " " + SignAlgorithm,
			" SignedMsg=" + hex.EncodeToString(signMsg),
			"Signature=" + hex.EncodeToString(signature),
		}

	} else if info.SignType == AuthV2 {
		if info.WalletSignStr == "" {
			return errors.New("wallet signature can not be empty with auth v2 type")
		}
		// wallet should use same sign algorithm
		authStr = []string{
			AuthV2 + " " + SignAlgorithm,
			" Signature=" + info.WalletSignStr,
		}
	} else {
		return errors.New("sign type error")
	}

	// set auth header
	req.Header.Set(HTTPHeaderAuthorization, strings.Join(authStr, ", "))

	return nil
}

// GetApproval return the signature info for the approval of preCreating resources
func (c *SPClient) GetApproval(ctx context.Context, bucketName, objectName string, authInfo AuthInfo) (string, error) {
	if err := utils.IsValidBucketName(bucketName); err != nil {
		return "", err
	}

	if objectName != "" {
		if err := utils.IsValidObjectName(objectName); err != nil {
			return "", err
		}
	}

	// set the action type
	urlVal := make(url.Values)
	if objectName != "" {
		urlVal["action"] = []string{CreateObjectAction}
	} else {
		urlVal["action"] = []string{CreateBucketAction}
	}

	reqMeta := requestMeta{
		bucketName:    bucketName,
		objectName:    objectName,
		urlValues:     urlVal,
		urlRelPath:    "get-approval",
		contentSHA256: EmptyStringSHA256,
	}

	sendOpt := sendOptions{
		method:     http.MethodGet,
		isAdminApi: true,
	}

	resp, err := c.sendReq(ctx, reqMeta, &sendOpt, authInfo)
	if err != nil {
		log.Printf("get approval rejected: %s \n", err.Error())
		return "", err
	}

	// fetch primary sp signature from sp response
	signature := resp.Header.Get(HTTPHeaderPreSignature)
	if signature == "" {
		return "", errors.New("fail to fetch pre createObject signature")
	}

	return signature, nil
}

// GetPieceHashRoots return primary pieces Hash and secondary piece Hash roots list and object size
// It is used for generate meta of object on the chain
func (c *SPClient) GetPieceHashRoots(reader io.Reader, segSize int64, ecShards int) (string, []string, int64, error) {
	pieceHashRoots, size, err := common.SplitAndComputerHash(reader, segSize, ecShards)
	if err != nil {
		log.Println("get hash roots fail", err.Error())
		return "", nil, 0, err
	}

	return pieceHashRoots[0], pieceHashRoots[1:], size, nil
}
