package sp

import (
	"context"
	"net/http"
	"testing"

	spClient "github.com/bnb-chain/greenfield/sdk/client/sp"
)

// TestCreateBucket test creating a new bucket
func TestCreateBucket(t *testing.T) {
	setup()
	defer shutdown()

	bucketName := "testbucket"
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		startHandle(t, r)
		testMethod(t, r, "GET")
		testHeader(t, r, spClient.HTTPHeaderContentSHA256, spClient.EmptyStringSHA256)
		w.WriteHeader(200)
	})

	client.CreateBucket(context.Background(), bucketName, spClient.NewAuthInfo(false, ""))

}
