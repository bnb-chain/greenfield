package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var EmptyURL = url.URL{}

func IsIPValid(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsDomainNameValid CheckdDomainName validates if input string is a valid domain name.
func IsDomainNameValid(hostName string) bool {
	// See RFC 1035, RFC 3696.
	hostName = strings.TrimSpace(hostName)
	if len(hostName) == 0 || len(hostName) > 255 {
		return false
	}
	if hostName[len(hostName)-1:] == "-" || hostName[:1] == "-" {
		return false
	}
	if hostName[len(hostName)-1:] == "_" || hostName[:1] == "_" {
		return false
	}
	if hostName[:1] == "." {
		return false
	}

	if strings.ContainsAny(hostName, "`~!@#$%^&*()+={}[]|\\\"';:><?/") {
		return false
	}
	return true
}

// GetEndpointURL - construct a new endpoint.
func GetEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	// If secure is false, use 'http' scheme.
	scheme := "https"
	if !secure {
		scheme = "http"
	}

	// Construct a secured endpoint URL.
	endpointURLStr := scheme + "://" + endpoint
	endpointURL, err := url.Parse(endpointURLStr)
	if err != nil {
		return nil, err
	}
	// check endpoint if it is valid
	if err := checkEndpointUrl(*endpointURL); err != nil {
		return nil, err
	}
	return endpointURL, nil
}

// checkEndpointUrl verify if endpoint url is valid, and return error
func checkEndpointUrl(endpointURL url.URL) error {
	if endpointURL == EmptyURL {
		return errors.New("Endpoint url is empty.")
	}

	if endpointURL.Path != "/" && endpointURL.Path != "" {
		return errors.New("Endpoint paths invalid")
	}

	host := endpointURL.Hostname()
	if !IsIPValid(host) && !IsDomainNameValid(host) {
		msg := endpointURL.Host + " does not meet ip address or domain name standards"
		return errors.New(msg)
	}

	return nil
}

// CalcSHA256Hex compute checksum of sha256 hash and encode it to hex
func CalcSHA256Hex(buf []byte) (hexStr string) {
	sum := CalcSHA256(buf)
	hexStr = hex.EncodeToString(sum)
	return
}

// CalcSHA256 compute checksum of sha256 from byte array
func CalcSHA256(buf []byte) []byte {
	h := sha256.New()
	h.Write(buf)
	sum := h.Sum(nil)
	return sum[:]
}

func DecodeURIComponent(s string) (string, error) {
	decodeStr, err := url.QueryUnescape(s)
	if err != nil {
		return s, err
	}
	return decodeStr, err
}

// AddQueryValues adds queryValue to url
func AddQueryValues(s string, qs url.Values) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	q := u.RawQuery
	rq := qs.Encode()
	if q != "" {
		if rq != "" {
			u.RawQuery = fmt.Sprintf("%s&%s", q, qs.Encode())
		}
	} else {
		u.RawQuery = rq
	}
	return u.String(), nil
}

// CloseResponse close the response body
func CloseResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}
}

// GetContentLength return the size of reader
func GetContentLength(reader io.Reader) (int64, error) {
	var contentLength int64
	var err error
	switch v := reader.(type) {
	case *bytes.Buffer:
		contentLength = int64(v.Len())
	case *bytes.Reader:
		contentLength = int64(v.Len())
	case *strings.Reader:
		contentLength = int64(v.Len())
	case *os.File:
		fInfo, fError := v.Stat()
		if fError != nil {
			err = fmt.Errorf("can't get reader content length,%s", fError.Error())
		} else {
			contentLength = fInfo.Size()
		}
	default:
		err = fmt.Errorf("can't get reader content length,unkown reader type")
	}
	return contentLength, err
}
