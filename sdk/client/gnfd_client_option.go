package client

import (
	"github.com/bnb-chain/greenfield/sdk/keys"
	"google.golang.org/grpc"
)

// GreenfieldClientOption configures how we set up the greenfield client.
type GreenfieldClientOption interface {
	Apply(*GreenfieldClient)
}

// GreenfieldClientOptionFunc defines an applied function for setting the greenfield client.
type GreenfieldClientOptionFunc func(*GreenfieldClient)

// Apply set up the option field to the client instance.
func (f GreenfieldClientOptionFunc) Apply(client *GreenfieldClient) {
	f(client)
}

// WithKeyManager returns a GreenfieldClientOption which configures a client key manager option.
func WithKeyManager(km keys.KeyManager) GreenfieldClientOption {
	return GreenfieldClientOptionFunc(func(client *GreenfieldClient) {
		client.keyManager = km
	})
}

// WithGrpcDialOption returns a GreenfieldClientOption which configures a grpc client connection options.
func WithGrpcDialOption(opts ...grpc.DialOption) GreenfieldClientOption {
	return GreenfieldClientOptionFunc(func(client *GreenfieldClient) {
		client.grpcDialOption = opts
	})
}
