package sp

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	spClient "github.com/bnb-chain/greenfield/sdk/client/sp"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/require"
)

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the COS client being tested.
	client *spClient.SPClient

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with  SPClient that is
// configured to talk to that test server.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()

	var err error

	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		log.Fatal("new key manager fail", err.Error())
	}
	client, err = spClient.NewSpClientWithKeyManager(server.URL[len("http://"):], &spClient.Option{}, keyManager)
	if err != nil {
		log.Fatal("create client  fail")
	}

}

func shutdown() {
	server.Close()
}

func startHandle(t *testing.T, r *http.Request) {
	t.Logf("start handle, Request method: %v, ", r.Method)
}

// testMethod judge if the method meeting expected
func testMethod(t *testing.T, r *http.Request, want string) {
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

// testHeader judge if the header meeting expected
func testHeader(t *testing.T, r *http.Request, header string, want string) {
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

func getUrl(r *http.Request) string {
	return r.URL.String()
}

// testHeader judge if the body meeting expected
func testBody(t *testing.T, r *http.Request, want string) {
	if r.Body == nil {
		t.Errorf("body empty")
		return
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Error reading request body: %v", err)
	}
	if got := string(b); got != want {
		t.Errorf("request Body is %s, want %s", got, want)
	}
}

// TestNewClient test new client and url function
func TestNewClient(t *testing.T) {
	mux_temp := http.NewServeMux()
	server_temp := httptest.NewServer(mux_temp)
	privKey, _, _ := testdata.KeyEthSecp256k1TestPubAddr()

	keyManager, err := keys.NewPrivateKeyManager(hex.EncodeToString(privKey.Bytes()))
	if err != nil {
		log.Fatal("new key manager fail")
	}
	c, err := spClient.NewSpClientWithKeyManager(server_temp.URL[7:], &spClient.Option{}, keyManager)
	if err != nil {
		log.Fatal("create client  fail")
	}

	if got, want := c.GetAgent(), spClient.UserAgent; got != want {
		t.Errorf("NewSpClient UserAgent is %v, want %v", got, want)
	}

	bucketName := "testBucket"
	objectName := "testObject"
	want := "http://" + server_temp.URL[7:] + "/testObject"
	got, _ := c.GenerateURL(bucketName, objectName, "", nil, false)
	fmt.Println("url2:", got)
	if got.String() != want {
		t.Errorf("URL is %v, want %v", got, want)
	}

}

// TestGetApproval test get approval request to preCreateBucket or preCreateObject
func TestGetApproval(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "test-bucket"
	ObjectName := "test-object"
	signature := "test-signature"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")
		url := getUrl(r)
		if strings.Contains(url, spClient.CreateObjectAction) {
			testHeader(t, r, spClient.HTTPHeaderResource, bucketName+"/"+ObjectName)
		} else if strings.Contains(url, spClient.CreateBucketAction) {
			testHeader(t, r, spClient.HTTPHeaderResource, bucketName)
		}

		w.Header().Set(spClient.HTTPHeaderPreSignature, signature)
		w.WriteHeader(200)
	})

	// test preCreateBucket
	gotSign, err := client.GetApproval(context.Background(), bucketName, "", spClient.NewAuthInfo(false, ""))
	require.NoError(t, err)

	if gotSign != signature {
		t.Errorf("get signature err")
	}

	//test preCreateObject
	gotSign, err = client.GetApproval(context.Background(), bucketName, ObjectName, spClient.NewAuthInfo(false, ""))

	require.NoError(t, err)

	if gotSign != signature {
		t.Errorf("get signature err")
	}

}
