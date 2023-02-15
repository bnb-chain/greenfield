package sp

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/bnb-chain/greenfield/sdk/utils"
	"github.com/rs/zerolog/log"
)

/* **** SAMPLE ERROR RESPONSE ****
<?xml version="1.0" encoding="UTF-8"?>
<Error>
   <Code>AccessDenied</Code>
   <Message>Access Denied</Message>
   <RequestId>xxx</RequestId>
</Error>
*/

// ErrResponse define the information of the error response
type ErrResponse struct {
	XMLName    xml.Name       `xml:"Error" json:"-"`
	Response   *http.Response `xml:"-"`
	Code       string
	StatusCode int
	Message    string
	Resource   string
	RequestID  string `xml:"RequestId"`
	Server     string
	BucketName string
	ObjectName string
}

// Error returns the error msg
func (r ErrResponse) Error() string {
	decodeURL := ""
	method := ""
	if r.Response != nil {
		decodeURL, _ = utils.DecodeURIComponent(r.Response.Request.URL.String())
		method = r.Response.Request.Method
	}

	return fmt.Sprintf("%v %v: %d %v  (Message: %v)",
		method, decodeURL,
		r.StatusCode, r.Code, r.Message)
}

// constructErrResponse  check the response is an error response
func constructErrResponse(r *http.Response, bucketName, objectName string) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	if r == nil {
		msg := "Response is empty. "
		return toInvalidArgumentResp(msg)
	}

	errorResp := ErrResponse{
		StatusCode: r.StatusCode,
		Server:     r.Header.Get("Server"),
	}

	if errorResp.RequestID == "" {
		errorResp.RequestID = r.Header.Get("X-Gnfd-Request-Id")
	}
	// read error body of response
	var data []byte
	var readErr error
	if r.Body != nil {
		data, readErr = io.ReadAll(r.Body)
		if readErr != nil {
			errorResp = ErrResponse{
				StatusCode: r.StatusCode,
				Code:       r.Status,
				Message:    readErr.Error(),
				BucketName: bucketName,
			}
		}
	}

	var decodeErr error
	// decode the xml error body if exists
	if readErr == nil && data != nil {
		decodeErr = xml.Unmarshal(data, &errorResp)
		if decodeErr != nil {
			log.Error().Err(decodeErr).Msg("unmarshal xml body fail ")
		}
	}

	if decodeErr != nil || data == nil || r.Body == nil {
		errBody := bytes.TrimSpace(data)

		switch r.StatusCode {
		case http.StatusNotFound:
			if objectName == "" {
				errorResp = ErrResponse{
					StatusCode: r.StatusCode,
					Code:       "NoSuchBucket",
					Message:    "The specified bucket does not exist.",
					BucketName: bucketName,
				}
			} else {
				errorResp = ErrResponse{
					StatusCode: r.StatusCode,
					Code:       "NoSuchObject",
					Message:    "The specified object does not exist.",
					BucketName: bucketName,
					ObjectName: objectName,
				}
			}
		case http.StatusForbidden:
			errorResp = ErrResponse{
				StatusCode: r.StatusCode,
				Code:       "AccessDenied",
				Message:    "no permission to access the resource",
				BucketName: bucketName,
				ObjectName: objectName,
			}
		default:
			msg := "unknown error"
			if len(errBody) > 0 {
				msg = string(errBody)
			}
			errorResp = ErrResponse{
				StatusCode: r.StatusCode,
				Code:       r.Status,
				Message:    msg,
				BucketName: bucketName,
			}
		}
	}

	return errorResp
}

// toInvalidArgumentResp return invalid  argument response.
func toInvalidArgumentResp(message string) error {
	return ErrResponse{
		StatusCode: http.StatusBadRequest,
		Code:       "InvalidArgument",
		Message:    message,
		RequestID:  "greenfield",
	}
}
